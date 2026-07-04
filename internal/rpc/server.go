package rpc

import (
	"backend/core-server/internal/rpc/healthpb"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"backend/core-server/internal/config"
	"backend/core-server/internal/infras/repo"
)

type Server struct {
	cfg    *config.Config
	Engine *grpc.Server
	db     *repo.SQLClient
}

func NewServer(
	cfg *config.Config,
	db *repo.SQLClient,
	health *HealthPRC,
) (*Server, error) {

	s := &Server{
		cfg:    cfg,
		Engine: grpc.NewServer(),
		db:     db,
	}

	// 在这个下面注册rpc...
	healthpb.RegisterHealthServiceServer(s.Engine, health)

	reflection.Register(s.Engine)
	return s, nil
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.cfg.Server.Addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.cfg.Server.Addr, err)
	}
	log.Printf("core-server listening on %s", s.cfg.Server.Addr)
	return s.Engine.Serve(lis)
}

func (s *Server) Stop() {
	s.Engine.GracefulStop()
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			log.Printf("close database: %v", err)
		}
	}
}

func (s *Server) Addr() string {
	return s.cfg.Server.Addr
}
