package model

import (
	"errors"
	"strings"
	"time"
)

var ErrInvalidItem = errors.New("invalid item")

type Item struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func (i *Item) Validate(maxNameLength int) error {
	if i == nil {
		return ErrInvalidItem
	}
	if strings.TrimSpace(i.Name) == "" {
		return ErrInvalidItem
	}
	if maxNameLength > 0 && len(i.Name) > maxNameLength {
		return ErrInvalidItem
	}
	return nil
}
