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
	ProcessedAt string  `json:"processed_at"`
}

type DataKeeper interface {
	//user
	RegisterUser(context.Context, AuthData) (User, error)
	LoginUser(context.Context, AuthData) (User, error)
	UserData(context.Context, int) (User, error)

	//order
	CreateOrder(context.Context, Order) error

	//withdraw

	//di
	HealthCheck() error
	Close() error
}
