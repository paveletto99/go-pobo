package sample

import (
	"fmt"

	"example.com/sample-service/internal/setup"
	"example.com/sample-service/pkg/database"
	"example.com/sample-service/pkg/observability"
	"github.com/hashicorp/go-multierror"
)

var (
	_ setup.DatabaseConfigProvider      = (*Config)(nil)
	_ setup.ObservabilityConfigProvider = (*Config)(nil)
)

type Config struct {
	Database      database.Config
	Observability observability.Config

	Port        string `env:"PORT, default=8080"`
	Maintenance bool   `env:"MAINTENANCE_MODE, default=false"`

	MaxItemNameLength int `env:"MAX_ITEM_NAME_LENGTH, default=120"`
}

func (c *Config) DatabaseConfig() *database.Config {
	return &c.Database
}

func (c *Config) ObservabilityConfig() *observability.Config {
	return &c.Observability
}

func (c *Config) Validate() error {
	var result *multierror.Error

	if c.MaxItemNameLength <= 0 {
		result = multierror.Append(result, fmt.Errorf("MAX_ITEM_NAME_LENGTH must be > 0"))
	}

	return result.ErrorOrNil()
}
