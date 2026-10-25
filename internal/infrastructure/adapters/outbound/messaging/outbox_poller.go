package messaging

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

// OutboxPoller es un worker en background que lee la tabla outbox_event
// y publica los mensajes pendientes a RabbitMQ.
type OutboxPoller struct {
	pool      *pgxpool.Pool
	amqpConn  *amqp.Connection
	amqpChan  *amqp.Channel
	returns   chan amqp.Return
	interval  time.Duration
	stopCh    chan struct{}
}

func NewOutboxPoller(pool *pgxpool.Pool, rabbitURL string) (*OutboxPoller, error) {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declarar la cola (durable)
	// Declarar el Topic Exchange común
	err = ch.ExchangeDeclare(
		"consultation.events", // name
		"topic",               // type
		true,                  // durable
		false,                 // auto-deleted
		false,                 // internal
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	returns := ch.NotifyReturn(make(chan amqp.Return, 1))

	return &OutboxPoller{
		pool:     pool,
		amqpConn: conn,
		amqpChan: ch,
		returns:  returns,
		interval: 5 * time.Second,
		stopCh:   make(chan struct{}),
	}, nil
}

func (p *OutboxPoller) Start() {
	log.Println("OutboxPoller started. Polling every 5 seconds...")
	go p.loop()
}

func (p *OutboxPoller) Stop() {
	close(p.stopCh)
	p.amqpChan.Close()
	p.amqpConn.Close()
}

func (p *OutboxPoller) loop() {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := p.processBatch(context.Background()); err != nil {
				log.Printf("ERROR: OutboxPoller processBatch failed: %v", err)
			}
		case <-p.stopCh:
			log.Println("OutboxPoller stopped.")
			return
		}
	}
}

func (p *OutboxPoller) processBatch(ctx context.Context) error {
	const qSelect = `
		SELECT id, event_type, payload, COALESCE(correlation_id, '')
		FROM outbox_event
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT 10
		FOR UPDATE SKIP LOCKED
	`
	
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start tx: %w", err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, qSelect)
	if err != nil {
		return fmt.Errorf("failed to query pending events: %w", err)
	}
	defer rows.Close()

	var eventIDs []string

	for rows.Next() {
		var id, eventType, correlationID string
		var payload []byte
		if err := rows.Scan(&id, &eventType, &payload, &correlationID); err != nil {
			return err
		}

		// Publicar a RabbitMQ
		err := p.amqpChan.PublishWithContext(ctx,
			"consultation.events", // exchange
			"post-consultation-summary-generated", // routing key
			true,  // mandatory: true -> Si no hay cola, rebota
			false, // immediate
			amqp.Publishing{
				ContentType:  "application/json",
				DeliveryMode: amqp.Persistent,
				MessageId:    id,
				Type:         eventType,
				Body:         payload,
			})
		
		if err != nil {
			log.Printf("ERROR: Failed to publish event %s to RabbitMQ: %v", id, err)
			continue 
		}

		// Esperar un momento para ver si el mensaje rebota (si no hay cola vinculada)
		select {
		case r := <-p.returns:
			log.Printf("WARN: Event %s returned by RabbitMQ (no queue bound): %s", id, r.ReplyText)
			continue // No lo marcamos como PUBLISHED, se reintentará
		case <-time.After(500 * time.Millisecond):
			// El mensaje fue enrutado con éxito a alguna cola
		}

		log.Printf("[OUTBOX] Event published to RabbitMQ: %s (Type: %s)", id, eventType)
		eventIDs = append(eventIDs, id)
	}

	if len(eventIDs) == 0 {
		return nil
	}

	const qUpdate = `UPDATE outbox_event SET status = 'PUBLISHED', published_at = now() WHERE id = ANY($1)`
	_, err = tx.Exec(ctx, qUpdate, eventIDs)
	if err != nil {
		return fmt.Errorf("failed to update event status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit outbox tx: %w", err)
	}

	return nil
}
