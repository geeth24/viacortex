package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"viacortex/internal/auth"

	"github.com/go-chi/cors"
)

// Define context keys to avoid string-based keys
type contextKey string
const (
    UserIDKey contextKey = "user_id"
    RoleKey   contextKey = "role"
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
			next.ServeHTTP(w, r)
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

		// Add claims to request context using typed keys
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper functions to get values from context
func GetUserIDFromContext(ctx context.Context) string {
    if id, ok := ctx.Value(UserIDKey).(string); ok {
        return id
    }
    return ""
}

func GetRoleFromContext(ctx context.Context) string {
    if role, ok := ctx.Value(RoleKey).(string); ok {
        return role
    }
    return ""
}