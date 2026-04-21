package middleware

import (
	"net/http"

	"example.com/sample-service/pkg/logging"
	"github.com/gorilla/mux"
)

func Recovery() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logging.FromContext(r.Context()).Named("middleware.Recovery")
			defer func() {
				if p := recover(); p != nil {
					logger.Errorw("http handler panic", "panic", p)
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
