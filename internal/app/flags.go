package app

import "flag"

func ReadFlags(P *Param) {
	if P.Server == "" {
		flag.StringVar(&P.Server, "a", "127.0.0.1:8080", "Адрес запускаемого сервера")
	}
	if P.URL == "" {
		flag.StringVar(&P.URL, "b", "http://127.0.0.1:8080", "Базовый адрес результирующего URL")
	}
	if P.Storage == "" {
		flag.StringVar(&P.Storage, "f", "jsonDB.json", "Хранилище URL")
	}
	flag.Parse()
}
