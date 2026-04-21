package sample

import (
	"context"
	"fmt"
	"strings"
	"time"

	"example.com/sample-service/internal/sample/model"
	"github.com/google/uuid"
)

type itemRepository interface {
	InsertItem(context.Context, *model.Item) error
	GetItem(context.Context, string) (*model.Item, error)
}

type Service struct {
	repository    itemRepository
	maxNameLength int
}

func NewService(repository itemRepository, maxNameLength int) *Service {
	return &Service{
		repository:    repository,
		maxNameLength: maxNameLength,
	}
}

func (s *Service) CreateItem(ctx context.Context, item *model.Item) (*model.Item, error) {
	if item == nil {
		return nil, model.ErrInvalidItem
	}

	item.Name = strings.TrimSpace(item.Name)
	if item.ID == "" {
		item.ID = uuid.NewString()
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now().UTC()
	}
	if err := item.Validate(s.maxNameLength); err != nil {
		return nil, err
	}

	if err := s.repository.InsertItem(ctx, item); err != nil {
		return nil, fmt.Errorf("repository.InsertItem: %w", err)
	}
	return item, nil
}

func (s *Service) GetItem(ctx context.Context, id string) (*model.Item, error) {
	if strings.TrimSpace(id) == "" {
		return nil, model.ErrInvalidItem
	}

	item, err := s.repository.GetItem(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("repository.GetItem: %w", err)
	}
	return item, nil
}
