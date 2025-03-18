package storage

import (
	"context"
	"time"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/entities"
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
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation (дубликат номера заказа)
				s.log.Infoln("Order number already exists:", order.Number)
				var ord db.Order

				ord, err = s.Queries.GetOrderByNumber(ctx, order.Number)
				if err != nil {
					return errors.Wrap(err, "failed to get order by number")
				}
				if order.UserID == int(ord.UserID) {
					return entities.ErrAlreadyExists
				}

				return entities.ErrConflict
			default:
				s.log.Errorln("Database error:", pqErr.Message)
				return errors.Wrap(err, "database error")
			}
		}
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
