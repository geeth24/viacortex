package api

import (
	"encoding/json"
	"log"
	"net/http"

	"viacortex/internal/db"

	"github.com/go-chi/chi/v5"
)

// getIPRules returns all IP rules for a domain
func (h *Handlers) getIPRules(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    domainID := chi.URLParam(r, "id")

    rows, err := h.db.Query(ctx, `
        SELECT id, ip_range, rule_type, description, created_at, updated_at
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

    rules := []db.IPRule{}
    for rows.Next() {
        var rule db.IPRule
        err := rows.Scan(
            &rule.ID, &rule.IPRange, &rule.RuleType,
            &rule.Description, &rule.CreatedAt, &rule.UpdatedAt,
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

// addIPRule adds a new IP rule to a domain
func (h *Handlers) addIPRule(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    domainID := chi.URLParam(r, "id")

    var rule db.IPRule
    if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Validate rule type
    if rule.RuleType != "whitelist" && rule.RuleType != "blacklist" {
        http.Error(w, "Invalid rule type", http.StatusBadRequest)
        return
    }

    var ruleID int64
    err := h.db.QueryRow(ctx, `
        INSERT INTO ip_rules (domain_id, ip_range, rule_type, description)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `, domainID, rule.IPRange, rule.RuleType, rule.Description).Scan(&ruleID)

    if err != nil {
        log.Printf("Error creating IP rule: %v", err)
        http.Error(w, "Failed to create IP rule", http.StatusInternalServerError)
        return
    }

    // Record audit log
    userID := getUserIDFromContext(ctx)
    if err := h.recordAudit(ctx, userID, "create", "ip_rule", ruleID, rule); err != nil {
        log.Printf("Error recording audit: %v", err)
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "id": ruleID,
        "message": "IP rule created successfully",
    })
}

// deleteIPRule deletes an IP rule
func (h *Handlers) deleteIPRule(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    ruleID := chi.URLParam(r, "ruleID")

    // Get rule details for audit log before deletion
    var oldRule db.IPRule
    err := h.db.QueryRow(ctx, `
        SELECT ip_range, rule_type, description 
        FROM ip_rules WHERE id = $1
    `, ruleID).Scan(&oldRule.IPRange, &oldRule.RuleType, &oldRule.Description)
    
    if err != nil {
        log.Printf("Error fetching IP rule: %v", err)
        http.Error(w, "Rule not found", http.StatusNotFound)
        return
    }

    result, err := h.db.Exec(ctx, "DELETE FROM ip_rules WHERE id = $1", ruleID)
    if err != nil {
        log.Printf("Error deleting IP rule: %v", err)
        http.Error(w, "Failed to delete IP rule", http.StatusInternalServerError)
        return
    }

    if rowsAffected := result.RowsAffected(); rowsAffected == 0 {
        http.Error(w, "Rule not found", http.StatusNotFound)
        return
    }

    // Record audit log
    userID := getUserIDFromContext(ctx)
    if err := h.recordAudit(ctx, userID, "delete", "ip_rule", 
        mustParseInt64(ruleID), oldRule); err != nil {
        log.Printf("Error recording audit: %v", err)
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "IP rule deleted successfully",
    })
}