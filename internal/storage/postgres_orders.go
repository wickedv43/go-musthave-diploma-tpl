package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
		Status:     "NEW",
		Accrual:    int32(order.Accrual * 100),
		UploadedAt: t,
	})
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505":
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

	s.log.WithFields(logrus.Fields{
		"number":  order.Number,
		"userID":  order.UserID,
		"status":  order.Status,
		"accrual": order.Accrual,
	}).Infoln("Created order successfully")

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

	s.log.WithFields(logrus.Fields{
		"number":  order.Number,
		"userID":  order.UserID,
		"status":  order.Status,
		"accrual": order.Accrual,
	}).Infoln("Updated order successfully")

	return nil
}

func (s *PostgresStorage) ProcessingOrders(ctx context.Context) ([]Order, error) {
	ordersPG, err := s.Queries.GetProcessingOrders(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []Order{}, entities.ErrNotFound
		}
		return []Order{}, errors.Wrap(err, "get orders by user id")
	}

	var orders []Order

	for _, order := range ordersPG {
		orders = append(orders, Order{
			Number:     order.Number,
			Status:     order.Status,
			Accrual:    float32(order.Accrual) / 100,
			UploadedAt: order.UploadedAt.Format(time.RFC3339),
		})
	}

	return orders, nil
}
