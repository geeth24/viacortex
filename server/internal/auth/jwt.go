package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenPair struct {
    AccessToken  string `json:"access_token"`
	AccessTokenValidUntil time.Time `json:"access_token_valid_until"`
    RefreshToken string `json:"refresh_token"`
	RefreshTokenValidUntil time.Time `json:"refresh_token_valid_until"`
}

type Claims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    Role   string `json:"role"`
    Type   string `json:"type"` // "access" or "refresh"
    jwt.RegisteredClaims
}

func GenerateTokenPair(userID, email, role string) (*TokenPair, error) {
    // Access token - short lived (15 minutes)
    accessToken, err := generateToken(userID, email, role, "access", 15*time.Minute)
    if err != nil {
        return nil, fmt.Errorf("failed to generate access token: %v", err)
    }

    // Refresh token - long lived (7 days)
    refreshToken, err := generateToken(userID, email, role, "refresh", 168*time.Hour)
    if err != nil {
        return nil, fmt.Errorf("failed to generate refresh token: %v", err)
    }

    return &TokenPair{
        AccessToken:  accessToken,
		AccessTokenValidUntil: time.Now().Add(15*time.Minute),
        RefreshToken: refreshToken,
		RefreshTokenValidUntil: time.Now().Add(168*time.Hour),
    }, nil
}

func generateToken(userID, email, role, tokenType string, expiry time.Duration) (string, error) {
    secret := []byte(os.Getenv("JWT_SECRET"))
    claims := Claims{
        UserID: userID,
        Email:  email,
        Role:   role,
        Type:   tokenType,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(secret)
}

func ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
        if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method")
        }
        return []byte(os.Getenv("JWT_SECRET")), nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, fmt.Errorf("invalid token")
}