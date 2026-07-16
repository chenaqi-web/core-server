package rpc

import (
	"fmt"
	"log"
	"net"

	"backend/core-server/internal/config"
	"backend/core-server/internal/infras/cache"
	"backend/core-server/internal/infras/mq/kafka"
	"backend/core-server/internal/infras/repo"
	jobdbsync "backend/core-server/internal/jobs/job-dbsync"
	"backend/core-server/internal/rpc/likepb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	cfg          *config.Config
	Engine       *grpc.Server
	dbClient     *repo.DBClient
	cacheClient  *cache.CacheClient
	producer     *kafka.SyncProducer
	kafkaManager *kafka.KafkaManager
	consumer     *jobdbsync.MessageQueueConsumer
}

func NewServer(
	cfg *config.Config,
	dbClient *repo.DBClient,
	cacheClient *cache.CacheClient,
	producer *kafka.SyncProducer,
	kafkaManager *kafka.KafkaManager,
	consumer *jobdbsync.MessageQueueConsumer,
	like *LikeRPC,
) (*Server, error) {
	engine := grpc.NewServer()
	likepb.RegisterLikeServiceServer(engine, like)
	reflection.Register(engine)

	return &Server{
		cfg:          cfg,
		Engine:       engine,
		dbClient:     dbClient,
		cacheClient:  cacheClient,
		producer:     producer,
		kafkaManager: kafkaManager,
		consumer:     consumer,
	}, nil
}

func (s *Server) Start() error {
	if s.consumer != nil {
		if err := s.consumer.Start(); err != nil {
			return fmt.Errorf("start message queue consumer: %w", err)
		}
	}

	lis, err := net.Listen("tcp", s.cfg.Server.Addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.cfg.Server.Addr, err)
	}
	log.Printf("core-server listening on %s", s.cfg.Server.Addr)
	return s.Engine.Serve(lis)
}

func (s *Server) Stop() {
	s.Engine.GracefulStop()

	if s.consumer != nil {
		if err := s.consumer.Stop(); err != nil {
			log.Printf("stop message queue consumer: %v", err)
		}
	}
	if s.producer != nil {
		if err := s.producer.Close(); err != nil {
			log.Printf("close kafka producer: %v", err)
		}
	}
	if s.kafkaManager != nil {
		if err := s.kafkaManager.Close(); err != nil {
			log.Printf("close kafka manager: %v", err)
		}
	}
	if s.dbClient != nil {
		if err := s.dbClient.Close(); err != nil {
			log.Printf("close mysql: %v", err)
		}
	}
	if s.cacheClient != nil {
		if err := s.cacheClient.Close(); err != nil {
			log.Printf("close redis: %v", err)
		}
	}
}

func (s *Server) Addr() string {
	return s.cfg.Server.Addr
}
