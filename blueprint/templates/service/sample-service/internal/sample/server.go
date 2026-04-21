package sample

import (
	"context"
	"fmt"
	"net/http"

	"example.com/sample-service/internal/middleware"
	sampledb "example.com/sample-service/internal/sample/database"
	"example.com/sample-service/internal/sample/model"
	"example.com/sample-service/internal/serverenv"
	"example.com/sample-service/pkg/logging"
	"example.com/sample-service/pkg/render"
	"example.com/sample-service/pkg/server"
	"github.com/gorilla/mux"
)

type itemService interface {
	CreateItem(context.Context, *model.Item) (*model.Item, error)
	GetItem(context.Context, string) (*model.Item, error)
}

type Server struct {
	config  *Config
	env     *serverenv.ServerEnv
	service itemService
	metrics *metrics
	h       *render.Renderer
}

func NewServer(cfg *Config, env *serverenv.ServerEnv) (*Server, error) {
	if env.Database() == nil {
		return nil, fmt.Errorf("missing database in server environment")
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	repository := sampledb.New(env.Database())
	service := NewService(repository, cfg.MaxItemNameLength)
	metrics, err := newMetrics()
	if err != nil {
		return nil, fmt.Errorf("metrics: %w", err)
	}

	return &Server{
		config:  cfg,
		env:     env,
		service: service,
		metrics: metrics,
		h:       render.NewRenderer(),
	}, nil
}

func (s *Server) Routes(ctx context.Context) *mux.Router {
	logger := logging.FromContext(ctx).Named("sample")

	r := mux.NewRouter()
	r.Use(middleware.Recovery())
	r.Use(middleware.PopulateRequestID())
	r.Use(middleware.PopulateLogger(logger))

	r.Handle("/health", server.HandleHealthz(s.env.Database())).Methods(http.MethodGet)
	r.Handle("/v1/items", s.handleCreateItem()).Methods(http.MethodPost)
	r.Handle("/v1/items/{id}", s.handleGetItem()).Methods(http.MethodGet)

	return r
}
