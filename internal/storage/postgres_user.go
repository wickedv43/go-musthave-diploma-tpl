package storage

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/entities"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/storage/db"
)

func (s *PostgresStorage) RegisterUser(ctx context.Context, au AuthData) (User, error) {
	user, err := s.Queries.CreateUser(ctx, db.CreateUserParams{
		Login:            au.Login,
		Password:         au.Password,
		BalanceCurrent:   0,
		BalanceWithdrawn: 0,
	})
	if err != nil {
		//if login already exist?

		//if another problems
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

func (s *PostgresStorage) UserData(ctx context.Context, id int) (User, error) {
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
			Sum:         float32(bill.Sum),
			ProcessedAt: bill.ProcessedAt,
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
			UploadedAt: order.UploadedAt,
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
