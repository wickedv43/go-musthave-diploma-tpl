package entities

import "github.com/pkg/errors"

var (
	ErrConflict        = errors.New("conflict")
	ErrNotFound        = errors.New("not found")
	ErrBadLogin        = errors.New("permission denied")
	ErrAlreadyExists   = errors.New("already exists")
	ErrBadOrder        = errors.New("bad order")
	ErrHaveEnoughMoney = errors.New("user have enough money to buy")
)
