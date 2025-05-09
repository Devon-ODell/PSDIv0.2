package internal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Devon-ODell/PSDI/internal/config"
	"github.com/Devon-ODell/PSDI/internal/model"
)

// SyncEvent represents an event in the sync queue
type SyncEvent struct {
	ID           int64
	EventType    string
	Payload      model.PaycorWebhookPayload
	Status       string // Pending, Processing, Success, Failed
	CreatedAt    time.Time
	UpdatedAt    time.Time
	RetryCount   int
	ErrorDetail  string
	JiraObjectID string // Populated after successful sync to Jira
}

// PostgresClient handles database operations
type PostgresClient struct {
	db *sql.DB
}

// NewPostgresClient creates a new database client
func NewPostgresClient(config config.DatabaseConfig) (*PostgresClient, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Name, config.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &PostgresClient{db: db}, nil
}

// Close closes the database connection
func (p *PostgresClient) Close() error {
	return p.db.Close()
}

// Ping tests database connection
func (p *PostgresClient) Ping() error {
	return p.db.Ping()
}

// InsertEvent inserts a new event into the sync queue
func (p *PostgresClient) InsertEvent(event *SyncEvent) error {
	// Convert payload to JSON
	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	query := `
		INSERT INTO sync_queue (
			event_type, payload, status, created_at, updated_at, retry_count, error_detail
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING id
	`

	err = p.db.QueryRow(
		query,
		event.EventType,
		payloadJSON,
		event.Status,
		event.CreatedAt,
		event.UpdatedAt,
		event.RetryCount,
		event.ErrorDetail,
	).Scan(&event.ID)

	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	return nil
}

// GetPendingEvents retrieves pending events for processing
func (p *PostgresClient) GetPendingEvents(limit int) ([]SyncEvent, error) {
	query := `
		SELECT 
			id, event_type, payload, status, created_at, updated_at, 
			retry_count, error_detail, jira_object_id
		FROM sync_queue
		WHERE status = 'Pending'
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := p.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending events: %w", err)
	}
	defer rows.Close()

	var events []SyncEvent
	for rows.Next() {
		var event SyncEvent
		var payloadJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&payloadJSON,
			&event.Status,
			&event.CreatedAt,
			&event.UpdatedAt,
			&event.RetryCount,
			&event.ErrorDetail,
			&event.JiraObjectID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}

		// Unmarshal payload
		if err := json.Unmarshal(payloadJSON, &event.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return events, nil
}

// UpdateEventStatus updates the status of an event
func (p *PostgresClient) UpdateEventStatus(id int64, status string, errorDetail string, jiraObjectID string) error {
	query := `
		UPDATE sync_queue
		SET 
			status = $1,
			updated_at = $2,
			error_detail = $3,
			jira_object_id = $4
		WHERE id = $5
	`

	_, err := p.db.Exec(
		query,
		status,
		time.Now(),
		errorDetail,
		jiraObjectID,
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to update event status: %w", err)
	}

	return nil
}

// IncrementRetryCount increments the retry count for an event
func (p *PostgresClient) IncrementRetryCount(id int64) error {
	query := `
		UPDATE sync_queue
		SET 
			retry_count = retry_count + 1,
			updated_at = $1
		WHERE id = $2
	`

	_, err := p.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to increment retry count: %w", err)
	}

	return nil
}

// ProcessPendingEvents processes pending events with a provided handler function
func (p *PostgresClient) ProcessPendingEvents(handler func(SyncEvent) error) error {
	// Get pending events
	events, err := p.GetPendingEvents(10)
	if err != nil {
		return err
	}

	for _, event := range events {
		// Mark as processing
		if err := p.UpdateEventStatus(event.ID, "Processing", "", ""); err != nil {
			return err
		}

		// Process event
		err := handler(event)

		if err != nil {
			// Mark as failed
			errorDetail := err.Error()
			if len(errorDetail) > 500 {
				// Truncate long error messages
				errorDetail = errorDetail[:500]
			}

			if err := p.UpdateEventStatus(event.ID, "Failed", errorDetail, ""); err != nil {
				return err
			}

			// Increment retry count
			if err := p.IncrementRetryCount(event.ID); err != nil {
				return err
			}
		} else {
			// Mark as success
			if err := p.UpdateEventStatus(event.ID, "Success", "", event.JiraObjectID); err != nil {
				return err
			}
		}
	}

	return nil
}
