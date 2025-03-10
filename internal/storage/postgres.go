package storage

import (
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/config"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/entities"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/logger"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/storage/db"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/util"

	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	Postgres *sql.DB
	Queries  *db.Queries
	log      *logrus.Entry
	cfg      *config.Config
}

func NewPostgresStorage(i do.Injector) (*PostgresStorage, error) {
	storage, err := do.InvokeStruct[PostgresStorage](i)
	log := do.MustInvoke[*logger.Logger](i).WithField("component", "postgres")
	cfg := do.MustInvoke[*config.Config](i)

	if err != nil {
		return nil, errors.Wrap(err, "invoke struct")
	}

	storage.log = log
	storage.cfg = cfg

	pgDB, err := sql.Open("postgres", storage.cfg.Database.DSN)
	if err != nil {
		return nil, errors.Wrap(err, "connect to postgres")
	}

	storage.Postgres = pgDB

	err = storage.Migrate()
	if err != nil {
		return nil, errors.Wrap(err, "migrate")
	}

	storage.Queries = db.New(pgDB)

	return storage, err
}

func (s *PostgresStorage) Close() error {
	return s.Postgres.Close()
}

func (s *PostgresStorage) HealthCheck() error {
	return s.Postgres.Ping()
}

func (s *PostgresStorage) Migrate() error {
	//create tables in db
	//open schema file
	query, err := os.ReadFile("./internal/storage/schema/schema.sql")
	if err != nil {
		return errors.Wrap(err, "read schema")
	}

	//exec query
	_, err = s.Postgres.Exec(string(query))
	if err != nil {
		//TODO: fix migration
		return nil
	}

	return nil
}

// TODO: uID int32?
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
			Current:   int(user.BalanceCurrent),
			Withdrawn: int(user.BalanceWithdrawn),
		},
		Orders: nil,
		Bills:  nil,
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
	if au.Password != user.Password || au.Login != user.Login {
		return User{}, entities.ErrBadLogin
	}

	return User{
		AuthData: AuthData{
			Login:    user.Login,
			Password: user.Password,
		},
		ID: int(user.ID),
		Balance: UserBalance{
			Current:   int(user.BalanceCurrent),
			Withdrawn: int(user.BalanceWithdrawn),
		},
		Orders: nil,
		Bills:  nil,
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
			Sum:         int(bill.Sum),
			ProcessedAt: bill.ProcessedAt.String(),
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
			Accrual:    int(order.Accrual),
			UploadedAt: order.UploadedAt.String(),
		})
	}

	return User{
		AuthData: AuthData{
			Login:    user.Login,
			Password: user.Password,
		},
		ID: int(user.ID),
		Balance: UserBalance{
			Current:   int(user.BalanceCurrent),
			Withdrawn: int(user.BalanceWithdrawn),
		},
		Orders: orders,
		Bills:  bills,
	}, nil
}

func (s *PostgresStorage) CreateOrder(ctx context.Context, order Order) error {
	uploadedAt, err := time.Parse(time.RFC3339, order.UploadedAt)
	if err != nil {
		return errors.Wrap(err, "parse uploaded at")
	}

	_, err = s.Queries.CreateOrder(ctx, db.CreateOrderParams{
		Number:     order.Number,
		UserID:     int32(order.UserID),
		Status:     "NEW",
		Accrual:    int32(order.Accrual),
		UploadedAt: uploadedAt,
	})
	if err != nil {
		//TODO: order number errors?
		//if same number by user

		//if same number by another user

		return errors.Wrap(err, "create order")
	}

	return nil
}

func (s *PostgresStorage) ProcessPayment(ctx context.Context, bill Bill) error {
	if !util.LuhnCheck(bill.Order) {
		return errors.New("bill number incorrect")
	}

	//process payment with tx
	tx, err := s.Postgres.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin transaction")
	}

	queriesWithTX := db.New(tx)

	//get order to get userID
	order, err := queriesWithTX.GetOrderByNumber(ctx, bill.Order)
	if err != nil {
		//mb log?
		tx.Rollback()
		return errors.Wrap(err, "get order")
	}

	//get user with balance
	user, err := queriesWithTX.GetUserByID(ctx, order.UserID)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "get user")
	}

	//check
	if int(user.BalanceCurrent)-bill.Sum < 0 {
		tx.Rollback()
		return errors.New("insufficient balance")
	}

	user.BalanceCurrent -= int32(bill.Sum)

	err = queriesWithTX.UpdateUserBalance(ctx, db.UpdateUserBalanceParams{
		ID:               user.ID,
		BalanceCurrent:   user.BalanceCurrent,
		BalanceWithdrawn: user.BalanceWithdrawn,
	})
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "update user balance")
	}

	//if all success
	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "commit transaction")
	}

	return nil
}
