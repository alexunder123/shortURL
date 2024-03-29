// Модуль собирает данные на удаление адресов от хэндлеров.
// Модуль передает в хранилище данные на удаление пакетами.
package worker

import (
	"context"
	"shortURL/internal/storage"
	"time"

	"github.com/rs/zerolog/log"
)

// Worker - структура с каналами для обмена данными с хэндлерами.
type Worker struct {
	InputCh  chan ToDelete
	finished chan struct{}
	Closed   bool
}

// ToDelete - структура, задающая формат обмена данными с хэндлерами.
type ToDelete struct {
	Keys []string
	ID   string
}

// NewWorker функция создает и возвращает ссылку на обработчик
func NewWorker() *Worker {
	worker := Worker{
		InputCh:  make(chan ToDelete),
		finished: make(chan struct{}),
		Closed:   false,
	}

	return &worker
}

// Run метод запускает работу обработчика
func (w *Worker) Run(strg storage.Storager, buffer int, delay time.Duration) {

	go func() {
		//update
		log.Debug().Msg("DeletingWorker started")
		userIDbuf := make([]string, 0, buffer)
		keyBuf := make([]string, 0, buffer)
		for {
			ctx, timeout := context.WithTimeout(context.Background(), delay)
			defer timeout()
			userIDbuf = userIDbuf[:0]
			keyBuf = keyBuf[:0]
		loop:
			for {
				select {
				case toDelele, ok := <-w.InputCh:
					if !ok {
						if len(userIDbuf) > 0 {
							strg.MarkDeleted(keyBuf, userIDbuf)
						}
						close(w.finished)
						log.Debug().Msg("DeletingWorker finished")
						return
					}
					for _, key := range toDelele.Keys {
						userIDbuf = append(userIDbuf, toDelele.ID)
						keyBuf = append(keyBuf, key)
						//flush
						if len(userIDbuf) == buffer {
							log.Debug().Msg("DeletingWorker flush")
							strg.MarkDeleted(keyBuf, userIDbuf)
							userIDbuf = userIDbuf[:0]
							keyBuf = keyBuf[:0]
						}
					}
				case <-ctx.Done():
					if len(userIDbuf) > 0 {
						log.Debug().Msg("DeletingWorker flush from cancel context")
						strg.MarkDeleted(keyBuf, userIDbuf)
					}
					break loop
				}
			}
		}
	}()

}

// Stop метод останавливает работу обработчика
func (w *Worker) Stop() {
	w.Closed = true
	close(w.InputCh)
	<-w.finished
	log.Info().Msg("inputCh closed")
}

// Add метод обрабатывает входящий JSON с данными и добавляет полученные значения
// в канал обработчика.
func (w *Worker) Add(urls []string, userID string) error {
	if w.Closed {
		return storage.ErrUnavailable
	}
	w.InputCh <- ToDelete{Keys: urls, ID: userID}
	return nil
}
