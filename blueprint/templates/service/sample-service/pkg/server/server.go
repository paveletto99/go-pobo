package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"example.com/sample-service/pkg/logging"
	"github.com/hashicorp/go-multierror"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Server struct {
	ip       string
	port     string
	listener net.Listener
}

func New(port string) (*Server, error) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener on :%s: %w", port, err)
	}
	addr := listener.Addr().(*net.TCPAddr)
	return &Server{
		ip:       addr.IP.String(),
		port:     strconv.Itoa(addr.Port),
		listener: listener,
	}, nil
}

func (s *Server) ServeHTTP(ctx context.Context, srv *http.Server) error {
	logger := logging.FromContext(ctx)

	errCh := make(chan error, 1)
	go func() {
		<-ctx.Done()
		shutdownCtx, done := context.WithTimeout(context.Background(), 5*time.Second)
		defer done()
		errCh <- srv.Shutdown(shutdownCtx)
	}()

	if err := srv.Serve(s.listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to serve: %w", err)
	}

	var merr *multierror.Error
	if err := <-errCh; err != nil {
		merr = multierror.Append(merr, fmt.Errorf("failed to shutdown server: %w", err))
	}

	logger.Debug("server stopped")
	return merr.ErrorOrNil()
}

func (s *Server) ServeHTTPHandler(ctx context.Context, handler http.Handler) error {
	return s.ServeHTTP(ctx, &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
		Handler: otelhttp.NewHandler(handler, "http.server",
			otelhttp.WithFilter(func(r *http.Request) bool {
				return r.URL.Path != "/health"
			}),
		),
	})
}

func (s *Server) Addr() string {
	return net.JoinHostPort(s.ip, s.port)
}
