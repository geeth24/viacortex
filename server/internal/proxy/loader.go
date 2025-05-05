package proxy

import (
	"context"
	"database/sql"
	"log"
	"net"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Loader struct {
    db    *pgxpool.Pool
    proxy *ProxyServer
}

func NewLoader(dbPool *pgxpool.Pool, proxy *ProxyServer) *Loader {
    return &Loader{
        db:    dbPool,
        proxy: proxy,
    }
}

func (l *Loader) Start(ctx context.Context) {
    // Initial load
    if err := l.LoadAllDomains(); err != nil {  // Changed this line
        log.Printf("Initial domain load error: %v", err)
    }

    // Periodic reload every 30 seconds
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := l.LoadAllDomains(); err != nil {  // Changed this line
                log.Printf("Domain reload error: %v", err)
            }
        }
    }
}

func (l *Loader) LoadAllDomains() error {

    ctx := context.Background()

    // Query all active domains
    rows, err := l.db.Query(ctx, `
        SELECT 
            d.id,
            d.name,
            d.target_url,
            d.ssl_enabled,
            d.health_check_enabled,
            d.health_check_interval
        FROM domains d
    `)
    if err != nil {
        return err
    }
    defer rows.Close()

    loadedDomains := make(map[string]struct{})

    for rows.Next() {
        var (
            domainID            int64
            name               string
            targetURL          string
            sslEnabled         bool
            healthCheckEnabled bool
            healthCheckInterval int
        )

        err := rows.Scan(
            &domainID,
            &name,
            &targetURL,
            &sslEnabled,
            &healthCheckEnabled,
            &healthCheckInterval,
        )
        if err != nil {
            return err
        }

        // For TCP domains, use the name instead of targetURL to avoid protocol prefix issues
        domainKey := targetURL
        if strings.HasPrefix(targetURL, "tcp://") {
            domainKey = name
            log.Printf("Using domain name %s instead of %s for TCP", name, targetURL)
        }

        config := &DomainConfig{
            Domain:             domainKey,
            SSLEnabled:        sslEnabled,
            HealthCheckEnabled: healthCheckEnabled,
        }

        // Load backends
        backends, err := l.loadBackends(ctx, domainID)
        if err != nil {
            log.Printf("Error loading backends for domain %s: %v", name, err)
            continue
        }
        config.Backends = backends

        // Load IP rules
        ipRules, err := l.loadIPRules(ctx, domainID)
        if err != nil {
            log.Printf("Error loading IP rules for domain %s: %v", name, err)
        }
        config.IPRules = ipRules

        // Load rate limit
        rateLimit, err := l.loadRateLimit(ctx, domainID)
        if err != nil {
            log.Printf("Error loading rate limit for domain %s: %v", name, err)
        }
        config.RateLimit = rateLimit

        // Update proxy configuration
        l.proxy.UpdateDomain(config.Domain, config)
        log.Printf("Loaded domain %s with SSL enabled: %v", config.Domain, config.SSLEnabled)
        loadedDomains[config.Domain] = struct{}{}
    }

    // Remove domains that no longer exist
    l.proxy.domains.Range(func(key, _ interface{}) bool {
        domain := key.(string)
        if _, exists := loadedDomains[domain]; !exists {
            l.proxy.DeleteDomain(domain)
        }
        return true
    })

    return nil
}

func (l *Loader) loadBackends(ctx context.Context, domainID int64) ([]*BackendServer, error) {
    rows, err := l.db.Query(ctx, `
        SELECT 
            id, scheme, host(ip::inet), port, weight, is_active,
            last_health_check, health_status
        FROM backend_servers
        WHERE domain_id = $1
    `, domainID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var backends []*BackendServer
    for rows.Next() {
        var b BackendServer
        var ipStr string
        var healthStatus sql.NullString  // Use sql.NullString for potentially NULL health_status
        err := rows.Scan(
            &b.ID,
            &b.Scheme,
            &ipStr,
            &b.Port,
            &b.Weight,
            &b.IsActive,
            &b.LastHealthCheck,
            &healthStatus,
        )
        if err != nil {
            return nil, err
        }

        // Convert health status if it's not null
        if healthStatus.Valid {
            status := healthStatus.String
            b.HealthStatus = &status
        }

        b.IP = net.ParseIP(ipStr).To4()
		log.Printf("Loaded backend %d with IP: %s", b.ID, b.IP)
        if b.IP == nil {
            log.Printf("Warning: Invalid IP address for backend %d: %s", b.ID, ipStr)
            continue
        }

        backends = append(backends, &b)
    }

    return backends, nil
}
func (l *Loader) loadIPRules(ctx context.Context, domainID int64) ([]*IPRule, error) {
    rows, err := l.db.Query(ctx, `
        SELECT id, ip_range, rule_type, description
        FROM ip_rules
        WHERE domain_id = $1
    `, domainID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var rules []*IPRule
    for rows.Next() {
        var r IPRule
        var ipRangeStr string
        err := rows.Scan(&r.ID, &ipRangeStr, &r.RuleType, &r.Description)
        if err != nil {
            return nil, err
        }

        _, ipNet, err := net.ParseCIDR(ipRangeStr)
        if err != nil {
            log.Printf("Warning: Invalid CIDR for rule %d: %s", r.ID, ipRangeStr)
            continue
        }
        r.IPRange = *ipNet

        rules = append(rules, &r)
    }

    return rules, nil
}

func (l *Loader) loadRateLimit(ctx context.Context, domainID int64) (*RateLimit, error) {
    var r RateLimit
    err := l.db.QueryRow(ctx, `
        SELECT id, requests_per_second, burst_size, per_ip
        FROM rate_limits
        WHERE domain_id = $1
        LIMIT 1
    `, domainID).Scan(&r.ID, &r.RequestsPerSecond, &r.BurstSize, &r.PerIP)

    if err != nil {
        if err.Error() == "no rows in result set" {
            return nil, nil
        }
        return nil, err
    }

    return &r, nil
}
