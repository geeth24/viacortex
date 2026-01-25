package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"viacortex/internal/db"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) getDomains(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rows, err := h.db.Query(ctx, `
		SELECT id, name, target_url, ssl_enabled, force_https, http2_enabled,
			hsts_enabled, hsts_max_age, websocket_enabled, custom_headers,
			basic_auth_enabled, basic_auth_username, enabled, created_at, updated_at
		FROM domains
		ORDER BY name
	`)
	if err != nil {
		log.Printf("Error fetching domains: %v", err)
		http.Error(w, "Failed to fetch domains", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var domains []db.Domain
	for rows.Next() {
		var d db.Domain
		var username *string
		err := rows.Scan(
			&d.ID, &d.Name, &d.TargetURL, &d.SSLEnabled, &d.ForceHTTPS,
			&d.HTTP2Enabled, &d.HSTSEnabled, &d.HSTSMaxAge, &d.WebsocketEnabled,
			&d.CustomHeaders, &d.BasicAuthEnabled, &username, &d.Enabled,
			&d.CreatedAt, &d.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning domain: %v", err)
			continue
		}
		if username != nil {
			d.BasicAuthUsername = *username
		}
		domains = append(domains, d)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domains)
}

func (h *Handlers) createDomain(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var d db.Domain

	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if d.CustomHeaders == nil {
		d.CustomHeaders = json.RawMessage("{}")
	}

	err := h.db.QueryRow(ctx, `
		INSERT INTO domains (
			name, target_url, ssl_enabled, force_https, http2_enabled,
			hsts_enabled, hsts_max_age, websocket_enabled, custom_headers,
			basic_auth_enabled, basic_auth_username, basic_auth_password, enabled
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`, d.Name, d.TargetURL, d.SSLEnabled, d.ForceHTTPS, d.HTTP2Enabled,
		d.HSTSEnabled, d.HSTSMaxAge, d.WebsocketEnabled, d.CustomHeaders,
		d.BasicAuthEnabled, d.BasicAuthUsername, d.BasicAuthPassword, true,
	).Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)

	if err != nil {
		log.Printf("Error creating domain: %v", err)
		http.Error(w, "Failed to create domain", http.StatusInternalServerError)
		return
	}

	d.Enabled = true
	d.BasicAuthPassword = ""

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(d)
}

func (h *Handlers) getDomain(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domainID := chi.URLParam(r, "id")

	var d db.Domain
	var username *string
	err := h.db.QueryRow(ctx, `
		SELECT id, name, target_url, ssl_enabled, force_https, http2_enabled,
			hsts_enabled, hsts_max_age, websocket_enabled, custom_headers,
			basic_auth_enabled, basic_auth_username, enabled, created_at, updated_at
		FROM domains WHERE id = $1
	`, domainID).Scan(
		&d.ID, &d.Name, &d.TargetURL, &d.SSLEnabled, &d.ForceHTTPS,
		&d.HTTP2Enabled, &d.HSTSEnabled, &d.HSTSMaxAge, &d.WebsocketEnabled,
		&d.CustomHeaders, &d.BasicAuthEnabled, &username, &d.Enabled,
		&d.CreatedAt, &d.UpdatedAt,
	)

	if err != nil {
		http.Error(w, "Domain not found", http.StatusNotFound)
		return
	}

	if username != nil {
		d.BasicAuthUsername = *username
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d)
}

func (h *Handlers) updateDomain(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domainID := chi.URLParam(r, "id")

	var d db.Domain
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		UPDATE domains SET
			name = $1, target_url = $2, ssl_enabled = $3, force_https = $4,
			http2_enabled = $5, hsts_enabled = $6, hsts_max_age = $7,
			websocket_enabled = $8, custom_headers = $9, basic_auth_enabled = $10,
			basic_auth_username = $11, enabled = $12, updated_at = CURRENT_TIMESTAMP
		WHERE id = $13`

	args := []interface{}{
		d.Name, d.TargetURL, d.SSLEnabled, d.ForceHTTPS, d.HTTP2Enabled,
		d.HSTSEnabled, d.HSTSMaxAge, d.WebsocketEnabled, d.CustomHeaders,
		d.BasicAuthEnabled, d.BasicAuthUsername, d.Enabled, domainID,
	}

	if d.BasicAuthPassword != "" {
		query = `
			UPDATE domains SET
				name = $1, target_url = $2, ssl_enabled = $3, force_https = $4,
				http2_enabled = $5, hsts_enabled = $6, hsts_max_age = $7,
				websocket_enabled = $8, custom_headers = $9, basic_auth_enabled = $10,
				basic_auth_username = $11, basic_auth_password = $12, enabled = $13,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $14`
		args = []interface{}{
			d.Name, d.TargetURL, d.SSLEnabled, d.ForceHTTPS, d.HTTP2Enabled,
			d.HSTSEnabled, d.HSTSMaxAge, d.WebsocketEnabled, d.CustomHeaders,
			d.BasicAuthEnabled, d.BasicAuthUsername, d.BasicAuthPassword, d.Enabled, domainID,
		}
	}

	result, err := h.db.Exec(ctx, query, args...)
	if err != nil {
		log.Printf("Error updating domain: %v", err)
		http.Error(w, "Failed to update domain", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Domain not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Domain updated successfully"})
}

func (h *Handlers) deleteDomain(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domainID := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(domainID, 10, 64)
	if err != nil {
		http.Error(w, "Invalid domain ID", http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec(ctx, "DELETE FROM domains WHERE id = $1", id)
	if err != nil {
		log.Printf("Error deleting domain: %v", err)
		http.Error(w, "Failed to delete domain", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Domain not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Domain deleted successfully"})
}

func (h *Handlers) toggleDomain(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domainID := chi.URLParam(r, "id")

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec(ctx,
		"UPDATE domains SET enabled = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
		req.Enabled, domainID)
	if err != nil {
		log.Printf("Error toggling domain: %v", err)
		http.Error(w, "Failed to toggle domain", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Domain not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"enabled": req.Enabled})
}
