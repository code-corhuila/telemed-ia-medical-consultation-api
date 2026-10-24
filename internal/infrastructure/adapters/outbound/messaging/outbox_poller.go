package messaging

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// OutboxPoller es un worker en background que lee la tabla outbox_event
// y publica los mensajes pendientes.
type OutboxPoller struct {
	pool     *pgxpool.Pool
	interval time.Duration
	stopCh   chan struct{}
}

func NewOutboxPoller(pool *pgxpool.Pool) *OutboxPoller {
	return &OutboxPoller{
		pool:     pool,
		interval: 5 * time.Second,
		stopCh:   make(chan struct{}),
	}
}

// Start inicia la goroutine que hace polling.
func (p *OutboxPoller) Start() {
	log.Println("OutboxPoller started. Polling every 5 seconds...")
	go p.loop()
}

// Stop detiene el worker de forma segura.
func (p *OutboxPoller) Stop() {
	close(p.stopCh)
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
	// Usamos FOR UPDATE SKIP LOCKED para evitar que multiples replicas
	// procesen el mismo evento.
	const qSelect = `
		SELECT id, event_type, payload, correlation_id
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

		// TODO: Aquí se conectará con RabbitMQ en el siguiente paso
		log.Printf("[OUTBOX] Event pending to publish: %s (Type: %s)", id, eventType)
		
		eventIDs = append(eventIDs, id)
	}

	if len(eventIDs) == 0 {
		return nil
	}

	// Marcar como publicados (temporalmente para que no se dupliquen en la prueba)
	// En el siguiente paso, si falla RabbitMQ, haremos rollback y se quedarán en PENDING
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
