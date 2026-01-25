package api

import (
	"encoding/json"
	"log"
	"net/http"

	"viacortex/internal/db"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) getRateLimit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domainID := chi.URLParam(r, "id")

	var rl db.RateLimit
	err := h.db.QueryRow(ctx, `
		SELECT id, domain_id, requests_per_second, burst_size, per_ip, enabled, created_at, updated_at
		FROM rate_limits
		WHERE domain_id = $1
	`, domainID).Scan(
		&rl.ID, &rl.DomainID, &rl.RequestsPerSecond, &rl.BurstSize,
		&rl.PerIP, &rl.Enabled, &rl.CreatedAt, &rl.UpdatedAt,
	)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rl)
}

func (h *Handlers) setRateLimit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domainID := chi.URLParam(r, "id")

	var rl db.RateLimit
	if err := json.NewDecoder(r.Body).Decode(&rl); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.db.QueryRow(ctx, `
		INSERT INTO rate_limits (domain_id, requests_per_second, burst_size, per_ip, enabled)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (domain_id) DO UPDATE SET
			requests_per_second = EXCLUDED.requests_per_second,
			burst_size = EXCLUDED.burst_size,
			per_ip = EXCLUDED.per_ip,
			enabled = EXCLUDED.enabled,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, created_at, updated_at
	`, domainID, rl.RequestsPerSecond, rl.BurstSize, rl.PerIP, true).Scan(
		&rl.ID, &rl.CreatedAt, &rl.UpdatedAt,
	)

	if err != nil {
		log.Printf("Error setting rate limit: %v", err)
		http.Error(w, "Failed to set rate limit", http.StatusInternalServerError)
		return
	}

	rl.Enabled = true

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rl)
}

func (h *Handlers) deleteRateLimit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domainID := chi.URLParam(r, "id")

	result, err := h.db.Exec(ctx, "DELETE FROM rate_limits WHERE domain_id = $1", domainID)
	if err != nil {
		log.Printf("Error deleting rate limit: %v", err)
		http.Error(w, "Failed to delete rate limit", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Rate limit not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Rate limit deleted successfully"})
}
