package storage

import "errors"

// Переменные для передачи хэндлену идентификатора ошибки.
var (
	ErrNoContent     error = errors.New("StatusNoContent")
	ErrConflict      error = errors.New("StatusConflict")
	ErrGone          error = errors.New("StatusGone")
	ErrUnsupported   error = errors.New("StatusUnsupportedMediaType")
	ErrBadRequest    error = errors.New("StatusBadRequest")
	ErrUnauthorized  error = errors.New("StatusUnauthorized")
	ErrInternalError error = errors.New("ErrInternalServerError")
	ErrForbidden     error = errors.New("StatusForbidden")
	ErrUnavailable   error = errors.New("StatusServiceUnavailable")
)
