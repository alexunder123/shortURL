package app

import (
	"log"
	"os"
	"os/signal"
	"shortURL/internal/storage"
	"syscall"
)

var (
	Stop = make(chan int)
)

func CloserDB(S storage.Storager) {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for s := range sigChan {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				log.Println("Начинаем выход из программы")
				S.CloseDB()
				close(Stop)
			}
		}
	}()
}
