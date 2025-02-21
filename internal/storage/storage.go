package storage

import (
	"context"
)

type User struct {
	Login    string      `json:"login"`
	Password string      `json:"password"`
	Balance  UserBalance `json:"balance"`
	Orders   []Order     `json:"orders"`
}

type UserBalance struct {
	Current   int `json:"current"`
	Withdrawn int `json:"withdrawn"`
}

// Order status | NEW | PROCESSING | INVALID | PROCESSED
type Order struct {
	Number     string `json:"number"`
	Status     string `json:"status"`
	Accrual    int    `json:"accrual"`
	UploadedAt string `json:"uploaded_at"`
}

type DataKeeper interface {
	RegisterUser(context.Context, User) error
	LoginUser(context.Context, User) error
	CreateOrder(context.Context, Order) error

	HealthCheck() error

	Close() error
}
