package database

import (
	"context"
	"errors"
	"fmt"

	"example.com/sample-service/internal/sample/model"
	coredatabase "example.com/sample-service/pkg/database"
	"github.com/jackc/pgx/v4"
)

type ItemDB struct {
	db *coredatabase.DB
}

func New(db *coredatabase.DB) *ItemDB {
	return &ItemDB{db: db}
}

func (db *ItemDB) InsertItem(ctx context.Context, item *model.Item) error {
	return db.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO items
				(id, name, created_at)
			VALUES
				($1, $2, $3)
		`, item.ID, item.Name, item.CreatedAt)
		if err != nil {
			return fmt.Errorf("inserting item: %w", err)
		}
		return nil
	})
}

func (db *ItemDB) GetItem(ctx context.Context, id string) (*model.Item, error) {
	var item model.Item

	if err := db.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, `
			SELECT id, name, created_at
			FROM items
			WHERE id = $1
			LIMIT 1
		`, id)

		if err := row.Scan(&item.ID, &item.Name, &item.CreatedAt); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return coredatabase.ErrNotFound
			}
			return fmt.Errorf("scanning item: %w", err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("getting item: %w", err)
	}

	return &item, nil
}
