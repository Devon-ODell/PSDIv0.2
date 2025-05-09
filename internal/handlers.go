package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/Devon-ODell/PSDI/internal/config"
	"github.com/Devon-ODell/PSDI/internal/db"
	"github.com/Devon-ODell/PSDI/internal/jira"
	"github.com//internal/model"
)

// Handlers holds the dependencies for API handlers
type Handlers struct {
	DB         *db.PostgresClient
	JiraClient *jira.Client
	Config     *config.Config
}

// NewHandlers creates a new instance of Handlers
func NewHandlers(db *db.PostgresClient, jiraClient *jira.Client, cfg *config.Config) *Handlers {
	return &Handlers{
		DB:         db,
		JiraClient: jiraClient,
		Config:     cfg,
	}
}

// Router sets up the HTTP routes
func (h *Handlers) Router() *mux.Router {
	r := mux.NewRouter()

	// Webhook endpoint for Paycor events
	r.HandleFunc("/webhooks/paycor", h.handlePaycorWebhook).Methods("POST")

	// Health check endpoint
	r.HandleFunc("/health", h.handleHealthCheck).Methods("GET")

	// Apply middleware
	r := h.applyMiddleware(r)

	return r
}

// Middleware for request logging, auth, etc.
func (h *Handlers) applyMiddleware(r *mux.Router) *mux.Router {
	// Logging middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			log.Printf("[%s] %s %s %v", r.Method, r.RequestURI, r.RemoteAddr, time.Since(start))
		})
	})

	// Authentication middleware for webhook endpoints
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Implement webhook authentication
			// Example: Verify signature or API key in header
			next.ServeHTTP(w, r)
		})
	})

	return r
}

// handlePaycorWebhook processes incoming webhook events from Paycor
func (h *Handlers) handlePaycorWebhook(w http.ResponseWriter, r *http.Request) {
	// Parse webhook payload
	var payload model.PaycorWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		log.Printf("Error parsing webhook payload: %v", err)
		return
	}

	// Validate webhook payload
	if payload.EventType == "" {
		http.Error(w, "Missing event type", http.StatusBadRequest)
		return
	}

	// Log the event
	log.Printf("Received webhook event: %s", payload.EventType)

	// Store event in database for processing
	event := &db.SyncEvent{
		EventType:   payload.EventType,
		Payload:     payload,
		Status:      "Pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		RetryCount:  0,
		ErrorDetail: "",
	}

	if err := h.DB.InsertEvent(event); err != nil {
		http.Error(w, "Failed to process event", http.StatusInternalServerError)
		log.Printf("Error inserting event: %v", err)
		return
	}

	// Return success immediately to webhook source
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "accepted",
		"message": "Event queued for processing",
	})
}

// handleHealthCheck provides a simple health check endpoint
func (h *Handlers) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if err := h.DB.Ping(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Database connection failed",
		})
		return
	}

	// Check Jira connection
	if err := h.JiraClient.CheckConnection(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Jira connection failed",
		})
		return
	}

	// All checks passed
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"version": h.Config.Version,
	})
}
