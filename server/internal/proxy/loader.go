package proxy

import (
	"context"
	"log"
	"net"
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
	if err := l.LoadAllDomains(); err != nil {
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
			if err := l.LoadAllDomains(); err != nil {
				log.Printf("Domain reload error: %v", err)
			}
		}
	}
}

func (l *Loader) LoadAllDomains() error {
	ctx := context.Background()

	// Query all enabled domains with new simplified schema
	rows, err := l.db.Query(ctx, `
		SELECT
			id,
			name,
			target_url,
			ssl_enabled,
			force_https,
			http2_enabled,
			hsts_enabled,
			hsts_max_age,
			websocket_enabled,
			custom_headers,
			basic_auth_enabled,
			basic_auth_username,
			basic_auth_password,
			enabled
		FROM domains
		WHERE enabled = true
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	loadedDomains := make(map[string]struct{})

	for rows.Next() {
		var (
			domainID          int64
			name              string
			targetURL         string
			sslEnabled        bool
			forceHTTPS        bool
			http2Enabled      bool
			hstsEnabled       bool
			hstsMaxAge        int
			websocketEnabled  bool
			customHeaders     []byte
			basicAuthEnabled  bool
			basicAuthUsername string
			basicAuthPassword string
			enabled           bool
		)

		err := rows.Scan(
			&domainID,
			&name,
			&targetURL,
			&sslEnabled,
			&forceHTTPS,
			&http2Enabled,
			&hstsEnabled,
			&hstsMaxAge,
			&websocketEnabled,
			&customHeaders,
			&basicAuthEnabled,
			&basicAuthUsername,
			&basicAuthPassword,
			&enabled,
		)
		if err != nil {
			log.Printf("Error scanning domain row: %v", err)
			continue
		}

		config := &DomainConfig{
			Domain:            name,
			TargetURL:         targetURL,
			SSLEnabled:        sslEnabled,
			ForceHTTPS:        forceHTTPS,
			HTTP2Enabled:      http2Enabled,
			HSTSEnabled:       hstsEnabled,
			HSTSMaxAge:        hstsMaxAge,
			WebsocketEnabled:  websocketEnabled,
			CustomHeaders:     customHeaders,
			BasicAuthEnabled:  basicAuthEnabled,
			BasicAuthUsername: basicAuthUsername,
			BasicAuthPassword: basicAuthPassword,
		}

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

		// Load redirects
		redirects, err := l.loadRedirects(ctx, domainID)
		if err != nil {
			log.Printf("Error loading redirects for domain %s: %v", name, err)
		}
		config.Redirects = redirects

		// Update proxy configuration
		l.proxy.UpdateDomain(config.Domain, config)
		log.Printf("Loaded domain %s -> %s (SSL: %v)", config.Domain, config.TargetURL, config.SSLEnabled)
		loadedDomains[config.Domain] = struct{}{}
	}

	// Remove domains that no longer exist
	l.proxy.domains.Range(func(key, _ interface{}) bool {
		domain := key.(string)
		if _, exists := loadedDomains[domain]; !exists {
			l.proxy.DeleteDomain(domain)
			log.Printf("Removed domain %s", domain)
		}
		return true
	})

	return nil
}

func (l *Loader) loadIPRules(ctx context.Context, domainID int64) ([]*IPRule, error) {
	rows, err := l.db.Query(ctx, `
		SELECT id, ip_range, rule_type, description
		FROM ip_rules
		WHERE domain_id = $1 AND enabled = true
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
		SELECT id, requests_per_second, burst_size, enabled
		FROM rate_limits
		WHERE domain_id = $1 AND enabled = true
		LIMIT 1
	`, domainID).Scan(&r.ID, &r.RequestsPerSecond, &r.BurstSize, &r.Enabled)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}

	return &r, nil
}

func (l *Loader) loadRedirects(ctx context.Context, domainID int64) ([]*Redirect, error) {
	rows, err := l.db.Query(ctx, `
		SELECT id, source_path, target_url, redirect_type, preserve_query, enabled
		FROM redirects
		WHERE domain_id = $1 AND enabled = true
		ORDER BY source_path DESC
	`, domainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var redirects []*Redirect
	for rows.Next() {
		var r Redirect
		err := rows.Scan(&r.ID, &r.SourcePath, &r.TargetURL, &r.RedirectType, &r.PreserveQuery, &r.Enabled)
		if err != nil {
			return nil, err
		}
		redirects = append(redirects, &r)
	}

	return redirects, nil
}
