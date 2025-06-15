package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"tennis-booker/internal/database"
	"tennis-booker/internal/secrets"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	secretsManager *secrets.SecretsManager
	db             database.Database
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(secretsManager *secrets.SecretsManager, db database.Database) *HealthHandler {
	return &HealthHandler{
		secretsManager: secretsManager,
		db:             db,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Services  struct {
		Database bool `json:"database"`
		Vault    bool `json:"vault"`
	} `json:"services"`
}

// SystemHealthResponse represents detailed system health
type SystemHealthResponse struct {
	Status      string                 `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	Version     string                 `json:"version"`
	Environment string                 `json:"environment"`
	Services    map[string]interface{} `json:"services"`
	Uptime      string                 `json:"uptime"`
}

var startTime = time.Now()

// Health handles basic health check
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0", // TODO: Get from build info
		Services: struct {
			Database bool `json:"database"`
			Vault    bool `json:"vault"`
		}{
			Database: h.checkDatabase(),
			Vault:    h.checkVault(),
		},
	}

	// If any service is down, mark as unhealthy
	if !response.Services.Database || !response.Services.Vault {
		response.Status = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SystemHealth handles detailed system health check
func (h *HealthHandler) SystemHealth(w http.ResponseWriter, r *http.Request) {
	services := make(map[string]interface{})
	
	// Check database
	dbStatus := h.checkDatabase()
	services["database"] = map[string]interface{}{
		"status":  dbStatus,
		"type":    "mongodb",
		"message": func() string {
			if dbStatus {
				return "Connected and responsive"
			}
			return "Connection failed"
		}(),
	}

	// Check Vault
	vaultStatus := h.checkVault()
	services["vault"] = map[string]interface{}{
		"status":  vaultStatus,
		"type":    "hashicorp-vault",
		"message": func() string {
			if vaultStatus {
				return "Connected and accessible"
			}
			return "Connection failed"
		}(),
	}

	// Calculate uptime
	uptime := time.Since(startTime)

	response := SystemHealthResponse{
		Status:      "healthy",
		Timestamp:   time.Now(),
		Version:     "1.0.0", // TODO: Get from build info
		Environment: "development", // TODO: Get from config
		Services:    services,
		Uptime:      uptime.String(),
	}

	// Check overall health
	allHealthy := true
	for _, service := range services {
		if serviceMap, ok := service.(map[string]interface{}); ok {
			if status, exists := serviceMap["status"].(bool); exists && !status {
				allHealthy = false
				break
			}
		}
	}

	if !allHealthy {
		response.Status = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// checkDatabase verifies database connectivity
func (h *HealthHandler) checkDatabase() bool {
	if h.db == nil {
		return false
	}
	
	// Try to ping the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	return h.db.Ping(ctx) == nil
}

// checkVault verifies Vault connectivity
func (h *HealthHandler) checkVault() bool {
	if h.secretsManager == nil {
		return false
	}
	
	// Try to get a secret to verify connectivity
	_, err := h.secretsManager.GetSecret("secret/data/tennisapp/prod/jwt", "secret")
	return err == nil
} 