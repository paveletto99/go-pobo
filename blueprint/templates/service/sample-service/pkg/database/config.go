package database

import (
	"net/url"
	"strconv"
	"time"
)

type Config struct {
	Name               string        `env:"DB_NAME" json:",omitempty"`
	User               string        `env:"DB_USER" json:",omitempty"`
	Host               string        `env:"DB_HOST, default=localhost" json:",omitempty"`
	Port               string        `env:"DB_PORT, default=5432" json:",omitempty"`
	SSLMode            string        `env:"DB_SSLMODE, default=disable" json:",omitempty"`
	ConnectionTimeout  int           `env:"DB_CONNECT_TIMEOUT" json:",omitempty"`
	Password           string        `env:"DB_PASSWORD" json:"-"`
	PoolMinConnections string        `env:"DB_POOL_MIN_CONNS" json:",omitempty"`
	PoolMaxConnections string        `env:"DB_POOL_MAX_CONNS" json:",omitempty"`
	PoolMaxConnLife    time.Duration `env:"DB_POOL_MAX_CONN_LIFETIME, default=5m" json:",omitempty"`
	PoolMaxConnIdle    time.Duration `env:"DB_POOL_MAX_CONN_IDLE_TIME, default=1m" json:",omitempty"`
	PoolHealthCheck    time.Duration `env:"DB_POOL_HEALTH_CHECK_PERIOD, default=1m" json:",omitempty"`
}

func (c *Config) ConnectionURL() string {
	if c == nil {
		return ""
	}

	host := c.Host
	if c.Port != "" {
		host = host + ":" + c.Port
	}

	u := &url.URL{
		Scheme: "postgres",
		Host:   host,
		Path:   c.Name,
	}
	if c.User != "" || c.Password != "" {
		u.User = url.UserPassword(c.User, c.Password)
	}

	q := u.Query()
	if c.ConnectionTimeout > 0 {
		q.Add("connect_timeout", strconv.Itoa(c.ConnectionTimeout))
	}
	if c.SSLMode != "" {
		q.Add("sslmode", c.SSLMode)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
