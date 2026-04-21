package serverenv

import (
	"context"

	"example.com/sample-service/pkg/database"
	"example.com/sample-service/pkg/observability"
)

type ServerEnv struct {
	database              *database.DB
	observabilityProvider *observability.Provider
}

type Option func(*ServerEnv) *ServerEnv

func New(ctx context.Context, opts ...Option) *ServerEnv {
	env := &ServerEnv{}
	for _, opt := range opts {
		env = opt(env)
	}
	return env
}

func WithDatabase(db *database.DB) Option {
	return func(env *ServerEnv) *ServerEnv {
		env.database = db
		return env
	}
}

func WithObservabilityProvider(provider *observability.Provider) Option {
	return func(env *ServerEnv) *ServerEnv {
		env.observabilityProvider = provider
		return env
	}
}

func (env *ServerEnv) Database() *database.DB {
	return env.database
}

func (env *ServerEnv) Close(ctx context.Context) error {
	if env == nil {
		return nil
	}
	if env.database != nil {
		env.database.Close(ctx)
	}
	if env.observabilityProvider != nil {
		return env.observabilityProvider.Close(ctx)
	}
	return nil
}
