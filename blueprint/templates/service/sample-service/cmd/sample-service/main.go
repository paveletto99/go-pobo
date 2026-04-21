package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"example.com/sample-service/internal/buildinfo"
	"example.com/sample-service/internal/sample"
	"example.com/sample-service/internal/setup"
	"example.com/sample-service/pkg/logging"
	"example.com/sample-service/pkg/server"
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
	logger := logging.FromContext(ctx)

	var cfg sample.Config
	env, err := setup.Setup(ctx, &cfg)
	if err != nil {
		return fmt.Errorf("setup.Setup: %w", err)
	}
	defer env.Close(ctx)

	sampleServer, err := sample.NewServer(&cfg, env)
	if err != nil {
		return fmt.Errorf("sample.NewServer: %w", err)
	}

	srv, err := server.New(cfg.Port)
	if err != nil {
		return fmt.Errorf("server.New: %w", err)
	}
	logger.Infow("server listening", "port", cfg.Port)
	return srv.ServeHTTPHandler(ctx, sampleServer.Routes(ctx))
}
