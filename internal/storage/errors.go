package storage

import "errors"

// Переменные для передачи хэндлену идентификатора ошибки.
var (
	ErrNoContent error = errors.New("StatusNoContent")
	ErrConflict  error = errors.New("StatusConflict")
	ErrGone      error = errors.New("StatusGone")
)
