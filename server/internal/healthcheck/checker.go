package healthcheck

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "net/netip"
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
            Transport: &http.Transport{
                DisableKeepAlives: true,
                MaxIdleConns: 100,
                IdleConnTimeout: 90 * time.Second,
                TLSHandshakeTimeout: 10 * time.Second,
                ResponseHeaderTimeout: 10 * time.Second,
            },
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

func (c *Checker) Stop() {
    close(c.stopChan)
    c.wg.Wait()
}

func (c *Checker) checkBackendHealth(ctx context.Context, scheme string, ip netip.Addr, port int) string {
    url := fmt.Sprintf("%s://%s:%d/", scheme, ip.String(), port)
    
    // Try up to 2 times with a short delay
    for attempts := 0; attempts < 2; attempts++ {
        req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
        if err != nil {
            log.Printf("Error creating health check request: %v", err)
            continue
        }
        
        // Add standard headers
        req.Header.Set("User-Agent", "ViaCortex-HealthCheck")
        req.Header.Set("Connection", "close")

        resp, err := c.client.Do(req)
        if err != nil {
            log.Printf("Health check failed for %s (attempt %d): %v", url, attempts+1, err)
            if attempts < 1 {
                time.Sleep(time.Second)
                continue
            }
            return "unhealthy"
        }
        defer resp.Body.Close()

        // Any response (even 404) means server is up
        if resp.StatusCode < 600 {
            return "healthy"
        }

        if attempts < 1 {
            time.Sleep(time.Second)
        }
    }

    return "unhealthy"
}

func (c *Checker) checkAllBackends(ctx context.Context) {
    // Get all domains with health checking enabled and their backends
    rows, err := c.db.Query(ctx, `
        SELECT 
            d.id, d.health_check_interval,
            b.id, b.scheme, b.ip::text, b.port
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
        var scheme, ipStr string

        err := rows.Scan(&domainID, &interval, &serverID, &scheme, &ipStr, &port)
        if err != nil {
            log.Printf("Error scanning health check row: %v", err)
            continue
        }

        // Parse IP address
        ip, err := netip.ParseAddr(ipStr)
        if err != nil {
            log.Printf("Error parsing IP address %s: %v", ipStr, err)
            continue
        }

        // Check backend health
        status := c.checkBackendHealth(ctx, scheme, ip, port)

        // Update status in database
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

        // Log status changes
        if err == nil {
            log.Printf("Backend %s:%d health status: %s", ip.String(), port, status)
        }
    }
}