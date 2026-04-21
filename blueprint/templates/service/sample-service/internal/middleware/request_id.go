package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type contextKey string

const contextKeyRequestID = contextKey("request_id")

func PopulateRequestID() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if RequestIDFromContext(ctx) == "" {
				id, err := uuid.NewRandom()
				if err != nil {
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
				ctx = context.WithValue(ctx, contextKeyRequestID, id.String())
				r = r.Clone(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequestIDFromContext(ctx context.Context) string {
	v := ctx.Value(contextKeyRequestID)
	if v == nil {
		return ""
	}
	id, ok := v.(string)
	if !ok {
		return ""
	}
	return id
}
