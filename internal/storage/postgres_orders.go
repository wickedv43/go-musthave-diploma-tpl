package storage

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/storage/db"
)

func (s *PostgresStorage) CreateOrder(ctx context.Context, order Order) error {
	t, err := time.Parse(time.RFC3339, order.UploadedAt)
	if err != nil {
		return errors.Wrap(err, "failed to parse uploaded at")
	}

	_, err = s.Queries.CreateOrder(ctx, db.CreateOrderParams{
		Number:     order.Number,
		UserID:     int32(order.UserID),
		Status:     order.Status,
		Accrual:    int32(order.Accrual * 100),
		UploadedAt: t,
	})
	if err != nil {
		//TODO: order number errors?
		//if same number by user
		s.log.Infoln("failed to create order: ", err)
		//if same number by another user

		return errors.Wrap(err, "create order")
	}

	s.log.Infoln("[NEW ORDER] ", order.Number, order.Status, order.Accrual, order.UploadedAt)
	return nil
}

func (s *PostgresStorage) UpdateOrder(ctx context.Context, order Order) error {
	err := s.Queries.UpdateOrder(ctx, db.UpdateOrderParams{
		Number:  order.Number,
		Status:  order.Status,
		Accrual: int32(order.Accrual * 100),
	})
	if err != nil {
		return errors.Wrap(err, "update order")
	}

	return nil
}
