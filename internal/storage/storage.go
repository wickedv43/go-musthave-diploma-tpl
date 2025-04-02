package storage

import (
	"context"
)

type User struct {
	AuthData

	ID      int         `json:"id"`
	Balance UserBalance `json:"balance"`
	Orders  []Order     `json:"orders"`
	Bills   []Bill      `json:"bills"`
}

type AuthData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserBalance struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

// Order status | REGISTERED | PROCESSING | INVALID | PROCESSED
type Order struct {
	UserID     int     `json:"-"`
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float32 `json:"accrual"`
	UploadedAt string  `json:"uploaded_at"`
}

type Bill struct {
	Order       string  `json:"order"`
	Sum         float32 `json:"sum"`
	UserID      int     `json:"-"`
	ProcessedAt string  `json:"processed_at"`
}

type DataKeeper interface {
	//user
	CreateUser(context.Context, AuthData) (User, error)
	GetUser(context.Context, int) (User, error)
	UpdateUserBalance(context.Context, User) error
	LoginUser(context.Context, AuthData) (User, error)

	//tx
	AddToBalanceWithTX(context.Context, int, float32) error
	WithdrawFromBalance(context.Context, int, float32) error

	//order
	CreateOrder(context.Context, Order) error
	UpdateOrder(context.Context, Order) error
	ProcessingOrders(context.Context) ([]Order, error)

	//withdraw
	CreateBill(context.Context, Bill) error

	//di
	HealthCheck() error
	Close() error
}
