package router

import (
	"bytes"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"

	"shortURL/internal/config"
	"shortURL/internal/handler"
	"shortURL/internal/storage"
	"shortURL/internal/worker"
)

func ExampleNewRouter() {

	// Задаем адрес сервера
	os.Setenv("SERVER_ADDRESS", "127.0.0.1:8888")
	os.Setenv("BASE_URL", "http://127.0.0.1:8888")

	// Подготовка роутера
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	strg := storage.NewStorage(cfg)
	wrkr := worker.NewWorker()
	hndlr := handler.NewHandler(cfg, strg, wrkr)

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
