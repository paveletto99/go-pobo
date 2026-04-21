package setup

import (
	"context"
	"fmt"

	"example.com/sample-service/internal/serverenv"
	"example.com/sample-service/pkg/database"
	"example.com/sample-service/pkg/logging"
	"example.com/sample-service/pkg/observability"
	"github.com/sethvargo/go-envconfig"
)

type DatabaseConfigProvider interface {
	DatabaseConfig() *database.Config
}

type ObservabilityConfigProvider interface {
	ObservabilityConfig() *observability.Config
}

func Setup(ctx context.Context, config interface{}) (*serverenv.ServerEnv, error) {
	return SetupWith(ctx, config, envconfig.OsLookuper())
}

func SetupWith(ctx context.Context, config interface{}, lookuper envconfig.Lookuper) (*serverenv.ServerEnv, error) {
	logger := logging.FromContext(ctx)

	if err := envconfig.ProcessWith(ctx, config, lookuper); err != nil {
		return nil, fmt.Errorf("error loading environment variables: %w", err)
	}
	logger.Infow("provided", "config", config)

	var opts []serverenv.Option

	if provider, ok := config.(DatabaseConfigProvider); ok {
		logger.Info("configuring database")
		db, err := database.NewFromEnv(ctx, provider.DatabaseConfig())
		if err != nil {
			return nil, fmt.Errorf("unable to connect to database: %w", err)
		}
		opts = append(opts, serverenv.WithDatabase(db))
		logger.Infow("database", "config", provider.DatabaseConfig())
	}

	if provider, ok := config.(ObservabilityConfigProvider); ok {
		logger.Info("configuring observability")
		otelProvider, err := observability.New(ctx, provider.ObservabilityConfig())
		if err != nil {
			return nil, fmt.Errorf("unable to configure observability: %w", err)
		}
		opts = append(opts, serverenv.WithObservabilityProvider(otelProvider))
		logger.Infow("observability", "config", provider.ObservabilityConfig())
	}

	return serverenv.New(ctx, opts...), nil
}
