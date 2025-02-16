package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

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

    // if users are = 0, then we need to create an admin user
    var count int
    err := h.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
    if err != nil {
        log.Printf("Error checking users: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }
    if count == 0 {
        req.Role = "admin"
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

    // Get the created user's data
    var user db.User
    var nullableName sql.NullString

    err = h.db.QueryRow(ctx, `
        SELECT id, email, name, role, active, last_login, created_at, updated_at
        FROM users 
        WHERE id = $1
    `, userID).Scan(
        &user.ID, &user.Email, &nullableName, &user.Role, &user.Active,
        &user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
    )

    if err != nil {
        log.Printf("Error fetching created user: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    // Set the name from nullable field
    if nullableName.Valid {
        user.Name = nullableName.String
    } else {
        user.Name = ""
    }

    // Generate tokens
    tokens, err := auth.GenerateTokenPair(fmt.Sprintf("%d", userID), req.Email, req.Role)
    if err != nil {
        log.Printf("Error generating tokens: %v", err)
        http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
        return
    }

    response := map[string]interface{}{
        "access_token": tokens.AccessToken,
        "refresh_token": tokens.RefreshToken,
        "user": map[string]interface{}{
            "id": user.ID,
            "email": user.Email,
            "role": user.Role,
            "active": user.Active,
            "name": user.Name,
        },
    }

    if user.LastLogin.Valid {
        response["user"].(map[string]interface{})["last_login"] = user.LastLogin.Time
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(response)
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
    var nullableName sql.NullString

    err = tx.QueryRow(ctx, `
        SELECT id, email, password_hash, role, active, name 
        FROM users 
        WHERE email = $1
    `, req.Email).Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.Active, &nullableName)

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
    
    // After the scan, set the name
    if nullableName.Valid {
        user.Name = nullableName.String
    } else {
        user.Name = "" // Set empty string if NULL
    }

    // Generate tokens
    tokens, err := auth.GenerateTokenPair(fmt.Sprintf("%d", user.ID), user.Email, user.Role)
    if err != nil {
        http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
        return
    }

    response := map[string]interface{}{
        "access_token": tokens.AccessToken,
        "refresh_token": tokens.RefreshToken,
        "user": map[string]interface{}{
            "id": user.ID,
            "email": user.Email,
            "role": user.Role,
            "active": user.Active,
            "name": user.Name,
        },
    }

    if user.LastLogin.Valid {
        response["user"].(map[string]interface{})["last_login"] = user.LastLogin.Time
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
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

func (h *Handlers) verifyToken(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    tokenParts := strings.Split(authHeader, " ")
    if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
        http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
        return
    }

    claims, err := auth.ValidateToken(tokenParts[1])
    if err != nil {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    // Get user data
    var user db.User
    var nullableName sql.NullString

    err = h.db.QueryRow(ctx, `
        SELECT id, email, name, role, active, last_login, created_at, updated_at
        FROM users 
        WHERE id = $1 AND active = true
    `, claims.UserID).Scan(
        &user.ID, &user.Email, &nullableName, &user.Role, &user.Active,
        &user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
    )

    if err != nil {
        log.Printf("Error fetching user: %v", err)
        http.Error(w, "User not found", http.StatusUnauthorized)
        return
    }

    // After the scan, set the name
    if nullableName.Valid {
        user.Name = nullableName.String
    } else {
        user.Name = ""
    }

    // Create response with proper handling of NULL values
    response := map[string]interface{}{
        "user": map[string]interface{}{
            "id": user.ID,
            "email": user.Email,
            "role": user.Role,
            "active": user.Active,
            "name": user.Name,
        },
    }

    // Only include last_login if it's not NULL
    if user.LastLogin.Valid {
        response["user"].(map[string]interface{})["last_login"] = user.LastLogin.Time
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (h *Handlers) checkUsers(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    var count int
    err := h.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
    if err != nil {
        log.Printf("Error checking users: %v", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]int{"count": count})
}

func (h *Handlers) handleVerify(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Get token from Authorization header
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    tokenParts := strings.Split(authHeader, " ")
    if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
        http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
        return
    }

    // Validate token
    claims, err := auth.ValidateToken(tokenParts[1])
    if err != nil {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    // Get user data
    var user db.User
    var nullableName sql.NullString

    err = h.db.QueryRow(ctx, `
        SELECT id, email, name, role, active, last_login, created_at, updated_at
        FROM users 
        WHERE id = $1 AND active = true
    `, claims.UserID).Scan(
        &user.ID, &user.Email, &nullableName, &user.Role, &user.Active,
        &user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
    )

    if err != nil {
        log.Printf("Error fetching user: %v", err)
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    // After the scan, set the name
    if nullableName.Valid {
        user.Name = nullableName.String
    } else {
        user.Name = ""
    }

    response := map[string]interface{}{
        "user": map[string]interface{}{
            "id": user.ID,
            "email": user.Email,
            "role": user.Role,
            "active": user.Active,
            "name": user.Name,
        },
    }

    if user.LastLogin.Valid {
        response["user"].(map[string]interface{})["last_login"] = user.LastLogin.Time
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}