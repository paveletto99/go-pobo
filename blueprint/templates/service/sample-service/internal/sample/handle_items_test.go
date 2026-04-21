package sample

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example.com/sample-service/internal/sample/model"
	"example.com/sample-service/pkg/render"
)

type fakeItemService struct {
	createItem *model.Item
	createErr  error
	getItem    *model.Item
	getErr     error
}

func (f *fakeItemService) CreateItem(ctx context.Context, item *model.Item) (*model.Item, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.createItem = &model.Item{
		ID:        "item-1",
		Name:      item.Name,
		CreatedAt: time.Unix(100, 0).UTC(),
	}
	return f.createItem, nil
}

func (f *fakeItemService) GetItem(ctx context.Context, id string) (*model.Item, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.getItem, nil
}

func TestHandleCreateItem(t *testing.T) {
	t.Parallel()

	service := &fakeItemService{}
	server := &Server{
		config:  &Config{},
		service: service,
		h:       render.NewRenderer(),
	}

	body := bytes.NewBufferString(`{"name":"hello"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/items", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	server.handleCreateItem().ServeHTTP(rr, req)

	if got, want := rr.Code, http.StatusCreated; got != want {
		t.Fatalf("got status %d, want %d", got, want)
	}

	var got model.Item
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != "item-1" || got.Name != "hello" {
		t.Fatalf("unexpected response: %+v", got)
	}
}
