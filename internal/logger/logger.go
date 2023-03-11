// Модуль опередляет и конфигурирует внешний пакет для логирования работы сервиса.
package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Newlogger конфигурирует формат и уровень отображения логов.
func Newlogger() {
	zerolog.TimeFieldFormat = ""

	zerolog.TimestampFunc = func() time.Time {
		return time.Date(2008, 1, 8, 17, 5, 05, 0, time.UTC)
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}
