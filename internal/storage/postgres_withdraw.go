package storage

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/storage/db"
)

func (s *PostgresStorage) CreateBill(ctx context.Context, b Bill) error {
	t, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	_, err = s.Queries.CreateBill(ctx, db.CreateBillParams{
		OrderNumber: b.Order,
		UserID:      int32(b.UserID),
		Sum:         int32(b.Sum * 100),
		ProcessedAt: t,
	})

	if err != nil {
		return errors.New("failed to create bill")
	}

	return nil
}
