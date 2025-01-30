package healthcheck

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Checker struct {
    db        *pgxpool.Pool
    client    *http.Client
    stopChan  chan struct{}
    wg        sync.WaitGroup
}

func NewChecker(db *pgxpool.Pool) *Checker {
    return &Checker{
        db: db,
        client: &http.Client{
            Timeout: 5 * time.Second,
        },
        stopChan: make(chan struct{}),
    }
}

func (c *Checker) Start(ctx context.Context) {
    c.wg.Add(1)
    go func() {
        defer c.wg.Done()
        
        // Check immediately on startup
        c.checkAllBackends(ctx)
        
        // Then set up periodic checks
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                return
            case <-c.stopChan:
                return
            case <-ticker.C:
                c.checkAllBackends(ctx)
            }
        }
    }()
}

func (c *Checker) checkAllBackends(ctx context.Context) {
    // Get all domains with health checking enabled and their backends
    rows, err := c.db.Query(ctx, `
        SELECT 
            d.id, d.health_check_interval,
            b.id, b.scheme, b.ip, b.port
        FROM domains d
        JOIN backend_servers b ON b.domain_id = d.id
        WHERE d.health_check_enabled = true 
        AND b.is_active = true
    `)
    if err != nil {
        log.Printf("Health check query error: %v", err)
        return
    }
    defer rows.Close()

    for rows.Next() {
        var domainID, interval, serverID, port int
        var scheme, ip string

        err := rows.Scan(&domainID, &interval, &serverID, &scheme, &ip, &port)
        if err != nil {
            log.Printf("Error scanning health check row: %v", err)
            continue
        }

        // Check each backend
        status := "unhealthy"
        url := fmt.Sprintf("%s://%s:%d/health", scheme, ip, port)
        
        req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
        if err != nil {
            log.Printf("Error creating health check request: %v", err)
        } else {
            resp, err := c.client.Do(req)
            if err == nil {
                if resp.StatusCode >= 200 && resp.StatusCode < 300 {
                    status = "healthy"
                }
                resp.Body.Close()
            }
        }

        // Update backend status
        _, err = c.db.Exec(ctx, `
            UPDATE backend_servers 
            SET 
                health_status = $1,
                last_health_check = CURRENT_TIMESTAMP
            WHERE id = $2
        `, status, serverID)
        
        if err != nil {
            log.Printf("Error updating backend status: %v", err)
        }
    }
}

func (c *Checker) Stop() {
    close(c.stopChan)
    c.wg.Wait()
}