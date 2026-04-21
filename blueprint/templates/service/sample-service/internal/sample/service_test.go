package sample

import (
	"context"
	"errors"
	"testing"

	"example.com/sample-service/internal/sample/model"
)

type fakeRepository struct {
	insertErr error
	getErr    error
	inserted  *model.Item
	item      *model.Item
}

func (f *fakeRepository) InsertItem(ctx context.Context, item *model.Item) error {
	f.inserted = item
	return f.insertErr
}

func (f *fakeRepository) GetItem(ctx context.Context, id string) (*model.Item, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.item, nil
}

func TestServiceCreateItem(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		repo := &fakeRepository{}
		service := NewService(repo, 20)

		item, err := service.CreateItem(ctx, &model.Item{Name: "  hello  "})
		if err != nil {
			t.Fatal(err)
		}
		if item.ID == "" {
			t.Fatal("expected generated id")
		}
		if got, want := item.Name, "hello"; got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
		if repo.inserted != item {
			t.Fatal("expected repository to receive created item")
		}
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		repo := &fakeRepository{}
		service := NewService(repo, 20)

		if _, err := service.CreateItem(ctx, &model.Item{}); !errors.Is(err, model.ErrInvalidItem) {
			t.Fatalf("got %v, want ErrInvalidItem", err)
		}
	})
}
