// internal/api/handlers.go
package api

import (
    "github.com/jackc/pgx/v4/pgxpool"
)

type Handlers struct {
    db *pgxpool.Pool
}

func NewHandlers(db *pgxpool.Pool) *Handlers {
    return &Handlers{db: db}
}