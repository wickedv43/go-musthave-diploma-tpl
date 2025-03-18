package storage

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/storage/db"
)

func (s *PostgresStorage) CreateBill(ctx context.Context, b Bill) error {
	t, err := time.Parse(time.RFC3339, b.ProcessedAt)
	if err != nil {
		return errors.Wrap(err, "failed to parse uploaded at")
	}

	_, err = s.Queries.CreateBill(ctx, db.CreateBillParams{
		OrderNumber: b.Order,
		UserID:      int32(b.UserID),
		Sum:         int32(b.Sum * 100),
		ProcessedAt: t,
	})

	if err != nil {
		s.log.Infoln("Failed to create bill", err)
		return errors.New("failed to create bill")
	}

	return nil
}
