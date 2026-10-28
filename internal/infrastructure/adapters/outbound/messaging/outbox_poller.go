package messaging

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	outboxExchange     = "consultation.events"
	outboxRoutingKey   = "post-consultation-summary-generated"
	outboxBatchSize    = 10
	outboxMaxAttempts  = 5
	outboxPollInterval = 5 * time.Second
	returnWaitTimeout  = 500 * time.Millisecond
)

// OutboxPoller is a background worker that reads outbox_event rows and
// publishes them to RabbitMQ with `mandatory` set. It is resilient to
// RabbitMQ being unavailable at startup or during runtime: the loop
// reconnects with exponential backoff and the API keeps serving traffic.
type OutboxPoller struct {
	pool     *pgxpool.Pool
	rabbitURL string

	conn    *amqp.Connection
	ch      *amqp.Channel
	returns chan amqp.Return

	interval time.Duration
	stopCh   chan struct{}
}

// NewOutboxPoller builds the poller without connecting to RabbitMQ.
// The connection is established lazily in Start(), so an unavailable
// broker does not block the API startup.
func NewOutboxPoller(pool *pgxpool.Pool, rabbitURL string) *OutboxPoller {
	return &OutboxPoller{
		pool:      pool,
		rabbitURL: rabbitURL,
		interval:  outboxPollInterval,
		stopCh:    make(chan struct{}),
	}
}

// Start launches the polling loop in a goroutine. It does not fail if
// RabbitMQ is down; the loop will reconnect with backoff.
func (p *OutboxPoller) Start() {
	log.Println("OutboxPoller started. Polling every", p.interval)
	go p.loop()
}

// Stop closes the stop channel and releases the AMQP resources.
func (p *OutboxPoller) Stop() {
	close(p.stopCh)
	p.closeAMQP()
}

func (p *OutboxPoller) loop() {
	backoff := time.Second
	maxBackoff := 60 * time.Second
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	// Try to connect once at startup; if it fails, the ticker will retry.
	if err := p.connect(); err != nil {
		log.Printf("WARN: OutboxPoller could not connect to RabbitMQ at startup (will retry with backoff): %v", err)
	}

	for {
		select {
		case <-ticker.C:
			if p.ch == nil {
				// Still no connection; try to establish one with backoff.
				if err := p.connect(); err != nil {
					log.Printf("WARN: OutboxPoller still cannot reach RabbitMQ (retry in %s): %v", backoff, err)
					time.Sleep(backoff)
					backoff *= 2
					if backoff > maxBackoff {
						backoff = maxBackoff
					}
					continue
				}
				backoff = time.Second
				log.Println("OutboxPoller reconnected to RabbitMQ")
			}

			if err := p.processBatch(context.Background()); err != nil {
				log.Printf("ERROR: OutboxPoller processBatch failed: %v", err)
			}

		case <-p.stopCh:
			log.Println("OutboxPoller stopped.")
			return
		}
	}
}

// connect establishes the AMQP connection, channel and exchange declaration.
// It stores the channel and returns a NotifyReturn channel for mandatory returns.
func (p *OutboxPoller) connect() error {
	conn, err := amqp.Dial(p.rabbitURL)
	if err != nil {
		return fmt.Errorf("dial rabbitmq: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("open channel: %w", err)
	}
	if err := ch.ExchangeDeclare(
		outboxExchange,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("declare exchange: %w", err)
	}
	returns := ch.NotifyReturn(make(chan amqp.Return, 1))

	p.conn = conn
	p.ch = ch
	p.returns = returns
	return nil
}

func (p *OutboxPoller) closeAMQP() {
	if p.ch != nil {
		_ = p.ch.Close()
		p.ch = nil
	}
	if p.conn != nil {
		_ = p.conn.Close()
		p.conn = nil
	}
	p.returns = nil
}

// processBatch reads pending events, publishes them with mandatory=true,
// and updates status/attempts/last_error in the same transaction.
func (p *OutboxPoller) processBatch(ctx context.Context) error {
	const qSelect = `
		SELECT id, event_id, event_type, payload, correlation_id
		FROM outbox_event
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`

	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, qSelect, outboxBatchSize)
	if err != nil {
		return fmt.Errorf("query pending events: %w", err)
	}
	defer rows.Close()

	type pending struct {
		id            string
		eventID       string
		eventType     string
		payload       []byte
		correlationID string
	}

	var batch []pending
	for rows.Next() {
		var ev pending
		if err := rows.Scan(&ev.id, &ev.eventID, &ev.eventType, &ev.payload, &ev.correlationID); err != nil {
			return fmt.Errorf("scan pending event: %w", err)
		}
		batch = append(batch, ev)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate pending events: %w", err)
	}

	if len(batch) == 0 {
		return nil
	}

	var (
		publishedIDs []string
		failedIDs    []string
		failedErrs   []string
	)

	for _, ev := range batch {
		err := p.ch.PublishWithContext(ctx,
			outboxExchange,
			outboxRoutingKey,
			true, // mandatory: bounce back if no queue is bound
			false,
			amqp.Publishing{
				ContentType:   "application/json",
				DeliveryMode:  amqp.Persistent,
				MessageId:     ev.eventID,
				CorrelationId: ev.correlationID,
				Type:          ev.eventType,
				Timestamp:     time.Now().UTC(),
				Body:          ev.payload,
			})

		if err != nil {
			log.Printf("WARN: OutboxPoller failed to publish event_id=%s (will retry): %v", ev.eventID, err)
			failedIDs = append(failedIDs, ev.id)
			failedErrs = append(failedErrs, err.Error())
			continue
		}

		// Wait briefly for a mandatory return: if the message was not routed
		// to any queue, RabbitMQ will return it here. Mark it as failed for this
		// attempt and retry in the next cycle.
		select {
		case r := <-p.returns:
			log.Printf("WARN: OutboxPoller event_id=%s returned by RabbitMQ (no queue bound): %s", ev.eventID, r.ReplyText)
			failedIDs = append(failedIDs, ev.id)
			failedErrs = append(failedErrs, "returned: "+r.ReplyText)
			continue
		case <-time.After(returnWaitTimeout):
			// Routed successfully.
		}

		log.Printf("[OUTBOX] Event published: id=%s event_id=%s type=%s", ev.id, ev.eventID, ev.eventType)
		publishedIDs = append(publishedIDs, ev.id)
	}

	if len(publishedIDs) > 0 {
		const qMarkPublished = `
			UPDATE outbox_event
			SET status = 'PUBLISHED',
			    published_at = now(),
			    updated_at = now(),
			    attempts = 0,
			    last_error = NULL
			WHERE id = ANY($1)
		`
		if _, err := tx.Exec(ctx, qMarkPublished, publishedIDs); err != nil {
			return fmt.Errorf("mark published: %w", err)
		}
	}

	if len(failedIDs) > 0 {
		const qMarkFailed = `
			UPDATE outbox_event
			SET attempts = attempts + 1,
			    last_error = $1,
			    updated_at = now(),
			    status = CASE WHEN attempts + 1 >= $2 THEN 'FAILED' ELSE 'PENDING' END
			WHERE id = ANY($3)
		`
		// Use the first error as the batch reason; individual events will
		// be retried and their attempts will be updated.
		reason := "publish failed"
		if len(failedErrs) > 0 {
			reason = failedErrs[0]
		}
		if _, err := tx.Exec(ctx, qMarkFailed, reason, outboxMaxAttempts, failedIDs); err != nil {
			return fmt.Errorf("mark failed: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit outbox tx: %w", err)
	}
	return nil
}
