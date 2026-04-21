package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"example.com/sample-service/internal/buildinfo"
	"example.com/sample-service/internal/jobs/samplejob"
	"example.com/sample-service/internal/setup"
	"example.com/sample-service/pkg/logging"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	logger := logging.NewLoggerFromEnv().
		With("build_id", buildinfo.BuildID).
		With("build_tag", buildinfo.BuildTag)
	ctx = logging.WithLogger(ctx, logger)

	defer func() {
		done()
		if r := recover(); r != nil {
			logger.Fatalw("application panic", "panic", r)
		}
	}()

	err := realMain(ctx)
	done()

	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("successful shutdown")
}

func realMain(ctx context.Context) error {
	var cfg samplejob.Config
	env, err := setup.Setup(ctx, &cfg)
	if err != nil {
		return fmt.Errorf("setup.Setup: %w", err)
	}
	defer env.Close(ctx)

	runner, err := samplejob.NewRunner(&cfg)
	if err != nil {
		return fmt.Errorf("samplejob.NewRunner: %w", err)
	}

	return runner.RunOnce(ctx)
}
