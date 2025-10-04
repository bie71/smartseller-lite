package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const (
	defaultRetryAttempts = 10
	defaultRetryDelay    = time.Second
	pingTimeout          = 5 * time.Second
)

// openMySQLWithRetry opens a MySQL connection with simple retry + ping logic.
func openMySQLWithRetry(dsn string) (*sql.DB, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return nil, fmt.Errorf("mysql dsn is required")
	}

	var lastErr error
	for attempt := 1; attempt <= defaultRetryAttempts; attempt++ {
		conn, err := sql.Open("mysql", dsn)
		if err != nil {
			lastErr = classifyAuthError(err)
			if _, isUnsupported := lastErr.(*UnsupportedAuthPluginError); isUnsupported {
				return nil, lastErr
			}
			time.Sleep(defaultRetryDelay)
			continue
		}

		configurePool(conn)

		ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
		pingErr := conn.PingContext(ctx)
		cancel()
		if pingErr == nil {
			return conn, nil
		}

		_ = conn.Close()
		lastErr = classifyAuthError(pingErr)
		if _, isUnsupported := lastErr.(*UnsupportedAuthPluginError); isUnsupported {
			return nil, lastErr
		}
		time.Sleep(defaultRetryDelay)
	}

	return nil, fmt.Errorf("open mysql failed after retries: %w", lastErr)
}

func configurePool(conn *sql.DB) {
	conn.SetMaxOpenConns(5)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxIdleTime(15 * time.Minute)
	conn.SetConnMaxLifetime(30 * time.Minute)
}
