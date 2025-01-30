package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"viacortex/internal/auth"
	"viacortex/internal/db"

	"github.com/jackc/pgx/v4"
	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type registerRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
    Role     string `json:"role"`
}


func (h *Handlers) handleRegister(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    var req registerRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Printf("Error decoding request: %v", err)
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Validate role
    if req.Role == "" {
        req.Role = "user" // Default role
    }
    if !isValidRole(req.Role) {
        http.Error(w, "Invalid role", http.StatusBadRequest)
        return
    }

    // Start transaction
    tx, err := h.db.Begin(ctx)
    if err != nil {
        log.Printf("Error starting transaction: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }
    defer tx.Rollback(ctx)

    // Check if email already exists
    var exists bool
    err = tx.QueryRow(ctx, 
        "SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)", 
        req.Email,
    ).Scan(&exists)
    
    if err != nil {
        log.Printf("Error checking email existence: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }
    if exists {
        http.Error(w, "Email already exists", http.StatusConflict)
        return
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        log.Printf("Error hashing password: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    // Create user
    var userID int64
    err = tx.QueryRow(ctx, `
        INSERT INTO users (
            email, 
            password_hash, 
            role,
            active,
            last_login
        ) VALUES ($1, $2, $3, true, NULL)
        RETURNING id
    `, req.Email, string(hashedPassword), req.Role).Scan(&userID)

    if err != nil {
        log.Printf("Error inserting user: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    // Add audit log
    changes := map[string]interface{}{
        "email": req.Email,
        "role":  req.Role,
    }
    changesJSON, _ := json.Marshal(changes)
    
    _, err = tx.Exec(ctx, `
        INSERT INTO audit_logs (
            user_id, 
            action, 
            entity_type, 
            entity_id, 
            changes
        ) VALUES ($1, $2, $3, $4, $5)
    `, userID, "register", "user", userID, changesJSON)

    if err != nil {
        log.Printf("Error creating audit log: %v", err)
    }

    // Commit transaction
    if err := tx.Commit(ctx); err != nil {
        log.Printf("Error committing transaction: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    // Generate tokens
    tokens, err := auth.GenerateTokenPair(fmt.Sprintf("%d", userID), req.Email, req.Role)
    if err != nil {
        log.Printf("Error generating tokens: %v", err)
        http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    if err := json.NewEncoder(w).Encode(tokens); err != nil {
        log.Printf("Error encoding response: %v", err)
    }
}

func (h *Handlers) handleLogin(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    var req loginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Start transaction
    tx, err := h.db.Begin(ctx)
    if err != nil {
        log.Printf("Error starting transaction: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }
    defer tx.Rollback(ctx)

    var user db.User
    err = tx.QueryRow(ctx, `
        SELECT id, email, password_hash, role, active 
        FROM users 
        WHERE email = $1
    `, req.Email).Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.Active)

    if err == pgx.ErrNoRows {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }
    if err != nil {
        log.Printf("Error querying user: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    // Check if user is active
    if !user.Active {
        http.Error(w, "Account is deactivated", http.StatusForbidden)
        return
    }

    // Verify password
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // Update last login time
    _, err = tx.Exec(ctx, `
        UPDATE users 
        SET last_login = CURRENT_TIMESTAMP 
        WHERE id = $1
    `, user.ID)
    
    if err != nil {
        log.Printf("Error updating last login: %v", err)
    }

    // Add audit log
    changes := map[string]string{"action": "login"}
    changesJSON, _ := json.Marshal(changes)
    
    _, err = tx.Exec(ctx, `
        INSERT INTO audit_logs (
            user_id, 
            action, 
            entity_type, 
            entity_id, 
            changes
        ) VALUES ($1, $2, $3, $4, $5)
    `, user.ID, "login", "user", user.ID, changesJSON)

    if err != nil {
        log.Printf("Error creating audit log: %v", err)
    }

    // Commit transaction
    if err := tx.Commit(ctx); err != nil {
        log.Printf("Error committing transaction: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }
    
    // Generate tokens
    tokens, err := auth.GenerateTokenPair(fmt.Sprintf("%d", user.ID), user.Email, user.Role)
    if err != nil {
        http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(tokens)
}

func (h *Handlers) handleRefresh(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    refreshToken := r.Header.Get("X-Refresh-Token")
    if refreshToken == "" {
        http.Error(w, "Refresh token required", http.StatusBadRequest)
        return
    }

    // Validate refresh token
    claims, err := auth.ValidateToken(refreshToken)
    if err != nil {
        http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
        return
    }

    // Verify user still exists and is active
    var active bool
    err = h.db.QueryRow(ctx, `
        SELECT active FROM users WHERE id = $1
    `, claims.UserID).Scan(&active)

    if err == pgx.ErrNoRows {
        http.Error(w, "User not found", http.StatusUnauthorized)
        return
    }
    if err != nil {
        log.Printf("Error querying user: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    if !active {
        http.Error(w, "Account is deactivated", http.StatusForbidden)
        return
    }

    // Generate new token pair
    tokens, err := auth.GenerateTokenPair(claims.UserID, claims.Email, claims.Role)
    if err != nil {
        http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(tokens)
}