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

	app, err := InitializeApp(cfg)
	if err != nil {
		log.Fatalf("initialize app: %v", err)
	}

	// 启动消费者协程
	if err := app.Consumer.Start(); err != nil {
		log.Fatalf("start like consumer: %v", err)
	}

	// 启动http服务
	go func() {
		if err := app.Server.Start(); err != nil {
			log.Fatalf("grpc server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down core-server...")
	app.Stop()
}
