package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/entities"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/storage/db"
)

func (s *PostgresStorage) CreateUser(ctx context.Context, au AuthData) (User, error) {
	user, err := s.Queries.CreateUser(ctx, db.CreateUserParams{
		Login:            au.Login,
		Password:         au.Password,
		BalanceCurrent:   0,
		BalanceWithdrawn: 0,
	})
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505":
				return User{}, entities.ErrAlreadyExists
			default:

				return User{}, errors.Wrap(err, "database error")
			}
		}

		if errors.Is(err, sql.ErrNoRows) {
			return User{}, fmt.Errorf("user creation failed: no rows affected")
		}

		return User{}, errors.Wrap(err, "create user")
	}

	return User{
		AuthData: AuthData{
			Login:    user.Login,
			Password: user.Password,
		},
		ID: int(user.ID),
		Balance: UserBalance{
			Current:   float32(user.BalanceCurrent) / 100,
			Withdrawn: float32(user.BalanceWithdrawn) / 100,
		},
	}, nil
}

func (s *PostgresStorage) LoginUser(ctx context.Context, au AuthData) (User, error) {
	//get user
	user, err := s.Queries.GetUserByLogin(ctx, au.Login)
	if err != nil {
		//if bad login
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, entities.ErrBadLogin
		}

		return User{}, errors.Wrap(err, "get user by login")
	}

	//auth check
	if au.Password != user.Password && au.Login == user.Login {
		return User{}, entities.ErrBadLogin
	}

	return User{
		AuthData: AuthData{
			Login:    user.Login,
			Password: user.Password,
		},
		ID: int(user.ID),
		Balance: UserBalance{
			Current:   float32(user.BalanceCurrent) / 100,
			Withdrawn: float32(user.BalanceWithdrawn) / 100,
		},
	}, nil
}

func (s *PostgresStorage) GetUser(ctx context.Context, id int) (User, error) {
	//get user
	user, err := s.Queries.GetUserByID(ctx, int32(id))
	if err != nil {
		//if bad login
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, entities.ErrBadLogin
		}

		return User{}, errors.Wrap(err, "get user by login")
	}

	//get user bills
	uBillsPG, err := s.Queries.GetBillsByUserID(ctx, user.ID)
	if err != nil {
		//no bills
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, entities.ErrNotFound
		}

		//another
		return User{}, errors.Wrap(err, "get bills by user id")
	}
	//format bill
	var bills []Bill

	for _, bill := range uBillsPG {
		bills = append(bills, Bill{
			Order:       bill.OrderNumber,
			Sum:         float32(bill.Sum) / 100,
			ProcessedAt: bill.ProcessedAt.Format(time.RFC3339),
		})
	}

	//orders
	ordersPG, err := s.Queries.GetOrdersByUserID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, entities.ErrNotFound
		}
		return User{}, errors.Wrap(err, "get orders by user id")
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

	return User{
		AuthData: AuthData{
			Login:    user.Login,
			Password: user.Password,
		},
		ID: int(user.ID),
		Balance: UserBalance{
			Current:   float32(user.BalanceCurrent) / 100,
			Withdrawn: float32(user.BalanceWithdrawn) / 100,
		},
		Orders: orders,
		Bills:  bills,
	}, nil
}

func (s *PostgresStorage) AddToBalanceWithTX(ctx context.Context, uID int, amount float32) error {
	tx, err := s.Postgres.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin tx")
	}
	qtx := s.Queries.WithTx(tx)

	user, err := qtx.GetUserByID(ctx, int32(uID))
	if err != nil {
		_ = tx.Rollback()
		return errors.Wrap(err, "get user")
	}

	user.BalanceCurrent += int32(amount * 100)

	err = qtx.UpdateUserBalance(ctx, db.UpdateUserBalanceParams{
		ID:               user.ID,
		BalanceCurrent:   user.BalanceCurrent,
		BalanceWithdrawn: user.BalanceWithdrawn,
	})
	if err != nil {
		_ = tx.Rollback()
		return errors.Wrap(err, "update balance")
	}

	return tx.Commit()
}

func (s *PostgresStorage) WithdrawFromBalance(ctx context.Context, uID int, amount float32) error {
	tx, err := s.Postgres.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin tx")
	}
	qtx := s.Queries.WithTx(tx)

	user, err := qtx.GetUserByID(ctx, int32(uID))
	if err != nil {
		_ = tx.Rollback()
		return errors.Wrap(err, "get user")
	}

	amountInCopeks := int32(amount * 100)
	if user.BalanceCurrent < amountInCopeks {
		_ = tx.Rollback()
		return entities.ErrPaymentRequired
	}

	user.BalanceCurrent -= amountInCopeks
	user.BalanceWithdrawn += amountInCopeks

	err = qtx.UpdateUserBalance(ctx, db.UpdateUserBalanceParams{
		ID:               user.ID,
		BalanceCurrent:   user.BalanceCurrent,
		BalanceWithdrawn: user.BalanceWithdrawn,
	})
	if err != nil {
		_ = tx.Rollback()
		return errors.Wrap(err, "update balance")
	}

	return tx.Commit()
}
