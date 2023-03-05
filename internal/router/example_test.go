package router

import (
	"bytes"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"time"

	"shortURL/internal/config"
	"shortURL/internal/handler"
	"shortURL/internal/storage"
	"shortURL/internal/worker"
)

func ExampleNewRouter() {

	// Подготовка роутера
	var (
		cfg config.Config = config.Config{
			ServerAddress:         "127.0.0.1:8888",
			BaseURL:               "http://127.0.0.1:8888",
			SavePlace:             config.SaveMemory,
			DeletingBufferSize:    10,
			DeletingBufferTimeout: 100 * time.Millisecond,
		}
		strg  storage.Storager = storage.NewStorage(&cfg)
		wrkr  *worker.Worker   = worker.NewWorker()
		hndlr *handler.Handler = handler.NewHandler(&cfg, strg, wrkr)
	)
	router := NewRouter(hndlr)

	// Запуск воркера
	wrkr.Run(strg, cfg.DeletingBufferSize, cfg.DeletingBufferTimeout)

	// запуск сервера
	listener, err := net.Listen("tcp", cfg.ServerAddress)
	if err != nil {
		log.Fatal(err)
	}
	testServer := httptest.NewUnstartedServer(router)
	testServer.Listener.Close()
	testServer.Listener = listener
	testServer.Start()
	defer testServer.Close()

	// Отправка POST сообщения на "/" с адресом в теле запроса
	request, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:8888/", bytes.NewReader([]byte(`/postgrespro.ru/docs/postgrespro/13/sql-syntax`)))
	if err != nil {
		log.Fatal(err)
	}
	result, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	// Считываем полученный сокращенный адрес в теле ответа
	userResult, err := io.ReadAll(result.Body)
	result.Body.Close()

	// Отправка GET запроса на "/<id>" для получения полного адреса
	request, err = http.NewRequest(http.MethodGet, string(userResult), nil)
	if err != nil {
		log.Fatal(err)
	}
	result, err = http.DefaultTransport.RoundTrip(request)
	if err != nil {
		log.Fatal(err)
	}
	// Считываем исходный адрес в заголовке ответа
	_ = result.Header.Get("Location")
	result.Body.Close()

	// Останавливаем воркер перед выходом из приложения
	wrkr.Stop()
}
