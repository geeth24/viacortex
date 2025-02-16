package middleware

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"os"

	"viacortex/internal/auth"

	"github.com/go-chi/cors"
)

// Define context keys to avoid string-based keys
type contextKey string
const (
    UserIDKey   contextKey = "userID"  // Changed to match the key used in handlers
    EmailKey    contextKey = "userEmail"
    RoleKey     contextKey = "userRole"
)

func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        w.Header().Set("Content-Security-Policy", "default-src 'self'")
        next.ServeHTTP(w, r)
    })
}

func Cors() func(http.Handler) http.Handler {
    return cors.Handler(cors.Options{
        AllowedOrigins:   []string{"http://localhost:*", "https://*.viacortex.com"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Refresh-Token"},
        ExposedHeaders:   []string{"Link"},
        AllowCredentials: true,
        MaxAge:          300,
    })
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if env := os.Getenv("ENV"); env != "production" {
			// For development, still set a test user ID
			ctx := context.WithValue(r.Context(), UserIDKey, int64(1))
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

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

		// Verify it's an access token, not a refresh token
		if claims.Type != "access" {
			http.Error(w, "Invalid token type", http.StatusUnauthorized)
			return
		}

		// Convert user ID from string to int64
		userID, err := strconv.ParseInt(claims.UserID, 10, 64)
		if err != nil {
			log.Printf("Error converting user ID: %v", err)
			http.Error(w, "Invalid user ID", http.StatusUnauthorized)
			return
		}

		log.Printf("Setting userID in context: %d", userID) // Debug log

		// Add claims to request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, userID)
		ctx = context.WithValue(ctx, EmailKey, claims.Email)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Update helper functions to return correct types
func GetUserIDFromContext(ctx context.Context) int64 {
    if id, ok := ctx.Value(UserIDKey).(int64); ok {
        return id
    }
    return 0
}

func GetEmailFromContext(ctx context.Context) string {
    if email, ok := ctx.Value(EmailKey).(string); ok {
        return email
    }
    return ""
}

func GetRoleFromContext(ctx context.Context) string {
    if role, ok := ctx.Value(RoleKey).(string); ok {
        return role
    }
    return ""
}