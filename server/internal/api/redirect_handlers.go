package api

import (
	"encoding/json"
	"log"
	"net/http"

	"viacortex/internal/db"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) getRedirects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domainID := chi.URLParam(r, "id")

	rows, err := h.db.Query(ctx, `
		SELECT id, domain_id, source_path, target_url, redirect_type,
			preserve_query, enabled, created_at, updated_at
		FROM redirects
		WHERE domain_id = $1
		ORDER BY created_at DESC
	`, domainID)

	if err != nil {
		log.Printf("Error fetching redirects: %v", err)
		http.Error(w, "Failed to fetch redirects", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var redirects []db.Redirect
	for rows.Next() {
		var rd db.Redirect
		err := rows.Scan(
			&rd.ID, &rd.DomainID, &rd.SourcePath, &rd.TargetURL,
			&rd.RedirectType, &rd.PreserveQuery, &rd.Enabled,
			&rd.CreatedAt, &rd.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning redirect: %v", err)
			continue
		}
		redirects = append(redirects, rd)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(redirects)
}

func (h *Handlers) addRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domainID := chi.URLParam(r, "id")

	var rd db.Redirect
	if err := json.NewDecoder(r.Body).Decode(&rd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if rd.RedirectType == 0 {
		rd.RedirectType = 301
	}

	err := h.db.QueryRow(ctx, `
		INSERT INTO redirects (domain_id, source_path, target_url, redirect_type, preserve_query, enabled)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`, domainID, rd.SourcePath, rd.TargetURL, rd.RedirectType, rd.PreserveQuery, true).Scan(
		&rd.ID, &rd.CreatedAt, &rd.UpdatedAt,
	)

	if err != nil {
		log.Printf("Error creating redirect: %v", err)
		http.Error(w, "Failed to create redirect", http.StatusInternalServerError)
		return
	}

	rd.Enabled = true

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rd)
}

func (h *Handlers) updateRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	redirectID := chi.URLParam(r, "redirectID")

	var rd db.Redirect
	if err := json.NewDecoder(r.Body).Decode(&rd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec(ctx, `
		UPDATE redirects SET
			source_path = $1, target_url = $2, redirect_type = $3,
			preserve_query = $4, enabled = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
	`, rd.SourcePath, rd.TargetURL, rd.RedirectType, rd.PreserveQuery, rd.Enabled, redirectID)

	if err != nil {
		log.Printf("Error updating redirect: %v", err)
		http.Error(w, "Failed to update redirect", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Redirect not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Redirect updated successfully"})
}

func (h *Handlers) deleteRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	redirectID := chi.URLParam(r, "redirectID")

	result, err := h.db.Exec(ctx, "DELETE FROM redirects WHERE id = $1", redirectID)
	if err != nil {
		log.Printf("Error deleting redirect: %v", err)
		http.Error(w, "Failed to delete redirect", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Redirect not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Redirect deleted successfully"})
}
