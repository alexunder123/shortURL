// go:build -ldflags "-X main.buildVersion=v.0.1.7 -X main.buildDate=02.03.2023 -X main.buildCommit=test"

package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"shortURL/internal/config"
	pb "shortURL/internal/grpc"
	"shortURL/internal/grpc/proto"
	"shortURL/internal/handler"
	"shortURL/internal/logger"
	"shortURL/internal/router"
	"shortURL/internal/storage"
	"shortURL/internal/worker"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

	logger.Newlogger()
	log.Info().Msg("Start program")
	cnfg, err := config.NewConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("NewConfig read environment error")
	}
	strg := storage.NewStorage(cnfg)
	log.Debug().Msg("storage init")
	deletingWorker := worker.NewWorker()
	hndlr := handler.NewHandler(cnfg, strg, deletingWorker)
	router := router.NewRouter(hndlr)
	log.Debug().Msg("handler init")
	deletingWorker.Run(strg, cnfg.DeletingBufferSize, cnfg.DeletingBufferTimeout)
	srv := http.Server{
		Addr:    cnfg.ServerAddress,
		Handler: router,
	}
	go func() {
		if cnfg.EnableHTTPS {
			cert, privateKey, err := config.NewSertificate(cnfg)
			if err != nil {
				log.Fatal().Err(err).Msg("NewSertificate generating error")
			}
			err = srv.ListenAndServeTLS(cert, privateKey)
			if err != nil {
				log.Error().Msgf("server failed: %s", err)
			}
		} else {
			err = srv.ListenAndServe()
			if err != nil {
				log.Error().Msgf("server failed: %s", err)
			}
		}
	}()
	gRPCconf := pb.NewShortURLsServer(cnfg, strg, deletingWorker)
	gRPCaddr, _, _ := strings.Cut(cnfg.ServerAddress, ":")
	listen, err := net.Listen("tcp", gRPCaddr+":3200")
	if err != nil {
		log.Fatal().Err(err).Msg("gRPC server announce error")
	}
	s := grpc.NewServer()
	proto.RegisterShortURLsServerServer(s, gRPCconf)
	reflection.Register(s)
	go func() {
		if err := s.Serve(listen); err != nil {
			log.Error().Msgf("gRPC server failed: %s", err)
		}
	}()

	//требование statictest: "the channel used with signal.Notify should be buffered"
	sigChan := make(chan os.Signal, 4)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigChan
	log.Info().Msgf("OS cmd received stop signal")
	deletingWorker.Stop()
	strg.CloseDB()
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Error().Msgf("HTTP server Shutdown: %s", err)
	}
	s.GracefulStop()
}
