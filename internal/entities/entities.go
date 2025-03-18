package entities

import "github.com/pkg/errors"

var (
	ErrConflict      = errors.New("conflict")
	ErrNotFound      = errors.New("not found")
	ErrBadLogin      = errors.New("permission denied")
	ErrAlreadyExists = errors.New("already exists")
)
