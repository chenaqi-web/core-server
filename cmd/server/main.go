package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"backend/core-server/internal/config"
	grpcserver "backend/core-server/internal/infras/grpc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	srv, err := grpcserver.New(cfg)
	if err != nil {
		log.Fatalf("create grpc server: %v", err)
	}

	go func() {
		log.Printf("core-server listening on %s", cfg.Server.Addr)
		if err := srv.Start(); err != nil {
			log.Fatalf("grpc server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down core-server...")
	srv.Stop()
}
