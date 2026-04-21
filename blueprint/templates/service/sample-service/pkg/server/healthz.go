package server

import (
	"fmt"
	"net/http"

	"example.com/sample-service/pkg/database"
	"example.com/sample-service/pkg/logging"
)

func HandleHealthz(db *database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logging.FromContext(ctx).Named("server.HandleHealthz")

		conn, err := db.Pool.Acquire(ctx)
		if err != nil {
			logger.Errorw("failed to acquire database connection", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer conn.Release()

		if err := conn.Conn().Ping(ctx); err != nil {
			logger.Errorw("failed to ping database", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "ok"}`)
	})
}
