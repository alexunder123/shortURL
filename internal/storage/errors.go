package storage

import "errors"

var (
	ErrNoContent error = errors.New("StatusNoContent")
	ErrConflict  error = errors.New("StatusConflict")
	ErrGone      error = errors.New("StatusGone")
)
