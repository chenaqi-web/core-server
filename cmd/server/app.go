package main

import (
	"log"

	jobdbsync "backend/core-server/internal/jobs/job-dbsync"
	"backend/core-server/internal/rpc"
)

type App struct {
	// rpc的服务协程
	Server *rpc.Server

	// 消费者的协程
	Consumer *jobdbsync.MessageQueueConsumer
}

func NewApp(server *rpc.Server, consumer *jobdbsync.MessageQueueConsumer) *App {
	return &App{
		Server:   server,
		Consumer: consumer,
	}
}

func (a *App) Stop() {
	a.Server.Stop()
	if err := a.Consumer.Stop(); err != nil {
		log.Printf("stop like consumer: %v", err)
	}
}
