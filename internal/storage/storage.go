package storage

import (
	"context"
)

// TODO: userID primary key
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
	Current   int `json:"current"`
	Withdrawn int `json:"withdrawn"`
}

// Order status | NEW | PROCESSING | INVALID | PROCESSED
type Order struct {
	UserID     int    `json:"-"`
	Number     string `json:"number"`
	Status     string `json:"status"`
	Accrual    int    `json:"accrual,omitempty"`
	UploadedAt string `json:"uploaded_at"`
}

type Bill struct {
	Order       string `json:"order"`
	Sum         int    `json:"sum"`
	ProcessedAt string `json:"processed_at"`
}

// TODO: postgres sqlc or gorm?
type DataKeeper interface {
	//user
	RegisterUser(context.Context, AuthData) (User, error)
	LoginUser(context.Context, AuthData) (User, error)
	UserData(context.Context, int) (User, error)

	//order
	CreateOrder(context.Context, Order) error

	//payment
	ProcessPayment(context.Context, Bill) error

	//di
	HealthCheck() error
	Close() error
}
