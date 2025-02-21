package entities

import "github.com/pkg/errors"

var ErrConflict = errors.New("conflict")
var ErrNotFound = errors.New("not found")
var ErrBadLogin = errors.New("permission denied")
