package samplejob

import (
	"fmt"

	"example.com/sample-service/internal/setup"
	"example.com/sample-service/pkg/observability"
	"github.com/hashicorp/go-multierror"
)

var _ setup.ObservabilityConfigProvider = (*Config)(nil)

type Config struct {
	Observability observability.Config

	WorkItems int  `env:"SAMPLE_JOB_WORK_ITEMS, default=1"`
	Fail      bool `env:"SAMPLE_JOB_FAIL, default=false"`
}

func (c *Config) ObservabilityConfig() *observability.Config {
	return &c.Observability
}

func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("missing config")
	}

	var result *multierror.Error

	if c.WorkItems < 0 {
		result = multierror.Append(result, fmt.Errorf("SAMPLE_JOB_WORK_ITEMS must be >= 0"))
	}

	return result.ErrorOrNil()
}
