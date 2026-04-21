package samplejob

import (
	"context"
	"fmt"
	"time"

	"example.com/sample-service/pkg/logging"
)

type Runner struct {
	config  *Config
	metrics *metrics
}

func NewRunner(cfg *Config) (*Runner, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	metrics, err := newMetrics()
	if err != nil {
		return nil, fmt.Errorf("metrics: %w", err)
	}

	return &Runner{
		config:  cfg,
		metrics: metrics,
	}, nil
}

func (r *Runner) RunOnce(ctx context.Context) (err error) {
	logger := logging.FromContext(ctx).Named("samplejob")
	start := time.Now()
	result := "success"
	defer func() {
		if err != nil {
			result = "error"
		}
		r.metrics.recordDuration(ctx, result, start)
	}()

	if r.config.Fail {
		r.metrics.recordError(ctx, "configured_failure")
		return fmt.Errorf("sample job configured to fail")
	}

	r.metrics.recordWork(ctx, "completed", r.config.WorkItems)
	r.metrics.recordSuccess(ctx)
	logger.Infow("sample job completed", "work_items", r.config.WorkItems)
	return nil
}
