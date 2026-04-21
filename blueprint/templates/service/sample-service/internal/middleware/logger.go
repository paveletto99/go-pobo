package middleware

import (
	"net/http"

	"example.com/sample-service/pkg/logging"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func PopulateLogger(originalLogger *zap.SugaredLogger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := originalLogger
			if id := RequestIDFromContext(ctx); id != "" {
				logger = logger.With("request_id", id)
			}
			if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
				logger = logger.With(
					"trace_id", spanCtx.TraceID().String(),
					"span_id", spanCtx.SpanID().String(),
				)
			}
			ctx = logging.WithLogger(ctx, logger)
			next.ServeHTTP(w, r.Clone(ctx))
		})
	}
}
