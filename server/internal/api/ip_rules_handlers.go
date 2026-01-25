package api

import (
	"encoding/json"
	"log"
	"net/http"

	"viacortex/internal/db"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) getIPRules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domainID := chi.URLParam(r, "id")

	rows, err := h.db.Query(ctx, `
		SELECT id, domain_id, ip_range, rule_type, description, enabled, created_at, updated_at
		FROM ip_rules
		WHERE domain_id = $1
		ORDER BY created_at DESC
	`, domainID)

	if err != nil {
		log.Printf("Error fetching IP rules: %v", err)
		http.Error(w, "Failed to fetch IP rules", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var rules []db.IPRule
	for rows.Next() {
		var rule db.IPRule
		err := rows.Scan(
			&rule.ID, &rule.DomainID, &rule.IPRange, &rule.RuleType,
			&rule.Description, &rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning IP rule: %v", err)
			continue
		}
		rules = append(rules, rule)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rules)
}

func (h *Handlers) addIPRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domainID := chi.URLParam(r, "id")

	var rule db.IPRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if rule.RuleType != "whitelist" && rule.RuleType != "blacklist" {
		http.Error(w, "Invalid rule type. Must be 'whitelist' or 'blacklist'", http.StatusBadRequest)
		return
	}

	err := h.db.QueryRow(ctx, `
		INSERT INTO ip_rules (domain_id, ip_range, rule_type, description, enabled)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`, domainID, rule.IPRange, rule.RuleType, rule.Description, true).Scan(
		&rule.ID, &rule.CreatedAt, &rule.UpdatedAt,
	)

	if err != nil {
		log.Printf("Error creating IP rule: %v", err)
		http.Error(w, "Failed to create IP rule", http.StatusInternalServerError)
		return
	}

	rule.Enabled = true

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rule)
}

func (h *Handlers) deleteIPRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ruleID := chi.URLParam(r, "ruleID")

	result, err := h.db.Exec(ctx, "DELETE FROM ip_rules WHERE id = $1", ruleID)
	if err != nil {
		log.Printf("Error deleting IP rule: %v", err)
		http.Error(w, "Failed to delete IP rule", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Rule not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "IP rule deleted successfully"})
}
