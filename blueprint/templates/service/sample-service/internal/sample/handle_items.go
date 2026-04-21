package sample

import (
	"errors"
	"net/http"
	"time"

	"example.com/sample-service/internal/jsonutil"
	"example.com/sample-service/internal/sample/model"
	coredatabase "example.com/sample-service/pkg/database"
	"example.com/sample-service/pkg/logging"
	"github.com/gorilla/mux"
)

type createItemRequest struct {
	Name string `json:"name"`
}

func (s *Server) handleCreateItem() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const operation = "create"
		ctx := r.Context()
		logger := logging.FromContext(ctx).Named("handleCreateItem")
		start := time.Now()
		s.metrics.recordRequest(ctx, operation)
		defer s.metrics.recordHandlerLatency(ctx, operation, start)

		var req createItemRequest
		code, err := jsonutil.Unmarshal(w, r, &req)
		if err != nil {
			s.metrics.recordError(ctx, operation, "decode")
			s.h.RenderJSON(w, code, err)
			return
		}

		item, err := s.service.CreateItem(ctx, &model.Item{Name: req.Name})
		if err != nil {
			if errors.Is(err, model.ErrInvalidItem) {
				s.metrics.recordError(ctx, operation, "invalid")
				s.h.RenderJSON(w, http.StatusBadRequest, err)
				return
			}
			s.metrics.recordError(ctx, operation, "internal")
			logger.Errorw("failed to create item", "error", err)
			s.h.RenderJSON(w, http.StatusInternalServerError, nil)
			return
		}

		s.metrics.recordItemCreated(ctx)
		s.h.RenderJSON(w, http.StatusCreated, item)
	})
}

func (s *Server) handleGetItem() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const operation = "get"
		ctx := r.Context()
		logger := logging.FromContext(ctx).Named("handleGetItem")
		start := time.Now()
		s.metrics.recordRequest(ctx, operation)
		defer s.metrics.recordHandlerLatency(ctx, operation, start)

		id := mux.Vars(r)["id"]
		item, err := s.service.GetItem(ctx, id)
		if err != nil {
			switch {
			case errors.Is(err, model.ErrInvalidItem):
				s.metrics.recordError(ctx, operation, "invalid")
				s.h.RenderJSON(w, http.StatusBadRequest, err)
			case errors.Is(err, coredatabase.ErrNotFound):
				s.metrics.recordLookup(ctx, false)
				s.h.RenderJSON(w, http.StatusNotFound, nil)
			default:
				s.metrics.recordError(ctx, operation, "internal")
				logger.Errorw("failed to get item", "error", err, "id", id)
				s.h.RenderJSON(w, http.StatusInternalServerError, nil)
			}
			return
		}

		s.metrics.recordLookup(ctx, true)
		s.h.RenderJSON(w, http.StatusOK, item)
	})
}
