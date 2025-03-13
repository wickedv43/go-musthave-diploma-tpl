package storage

import (
	"context"

	"github.com/pkg/errors"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/storage/db"
)

func (s *PostgresStorage) CreateOrder(ctx context.Context, order Order) error {
	_, err := s.Queries.CreateOrder(ctx, db.CreateOrderParams{
		Number:     order.Number,
		UserID:     int32(order.UserID),
		Status:     order.Status,
		Accrual:    int32(order.Accrual * 100),
		UploadedAt: order.UploadedAt,
	})
	if err != nil {
		//TODO: order number errors?
		//if same number by user

		//if same number by another user

		return errors.Wrap(err, "create order")
	}

	return nil
}
