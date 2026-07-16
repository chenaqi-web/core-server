package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"backend/core-server/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	srv, err := InitializeServer(cfg)
	if err != nil {
		log.Fatalf("initialize server: %v", err)
	}

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down core-server...")
	srv.Stop()
}
