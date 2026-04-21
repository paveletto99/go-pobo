package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"example.com/sample-service/pkg/logging"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func NewFromEnv(ctx context.Context, cfg *Config) (*DB, error) {
	pgxConfig, err := pgxpool.ParseConfig(dbDSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}
	pgxConfig.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
		return conn.Ping(ctx) == nil
	}

	pool, err := pgxpool.ConnectConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	return &DB{Pool: pool}, nil
}

func (db *DB) Close(ctx context.Context) {
	logger := logging.FromContext(ctx)
	logger.Info("closing connection pool")
	db.Pool.Close()
}

func dbDSN(cfg *Config) string {
	vals := dbValues(cfg)
	parts := make([]string, 0, len(vals))
	for k, v := range vals {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, " ")
}

func dbValues(cfg *Config) map[string]string {
	values := map[string]string{}
	setIfNotEmpty(values, "dbname", cfg.Name)
	setIfNotEmpty(values, "user", cfg.User)
	setIfNotEmpty(values, "host", cfg.Host)
	setIfNotEmpty(values, "port", cfg.Port)
	setIfNotEmpty(values, "sslmode", cfg.SSLMode)
	setIfPositive(values, "connect_timeout", cfg.ConnectionTimeout)
	setIfNotEmpty(values, "password", cfg.Password)
	setIfNotEmpty(values, "pool_min_conns", cfg.PoolMinConnections)
	setIfNotEmpty(values, "pool_max_conns", cfg.PoolMaxConnections)
	setIfPositiveDuration(values, "pool_max_conn_lifetime", cfg.PoolMaxConnLife)
	setIfPositiveDuration(values, "pool_max_conn_idle_time", cfg.PoolMaxConnIdle)
	setIfPositiveDuration(values, "pool_health_check_period", cfg.PoolHealthCheck)
	return values
}

func setIfNotEmpty(m map[string]string, key, val string) {
	if val != "" {
		m[key] = val
	}
}

func setIfPositive(m map[string]string, key string, val int) {
	if val > 0 {
		m[key] = fmt.Sprintf("%d", val)
	}
}

func setIfPositiveDuration(m map[string]string, key string, d time.Duration) {
	if d > 0 {
		m[key] = d.String()
	}
}
