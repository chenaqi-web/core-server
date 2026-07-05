package rpc

import (
	"backend/core-server/internal/rpc/likepb"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"backend/core-server/internal/config"
)

type Server struct {
	cfg    *config.Config
	Engine *grpc.Server
}

func NewServer(
	cfg *config.Config,
	like *LikeRPC,
) (*Server, error) {

	s := &Server{
		cfg:    cfg,
		Engine: grpc.NewServer(),
	}

	// 在这个下面注册rpc...
	likepb.RegisterLikeServiceServer(s.Engine, like)

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
}

func (s *Server) Addr() string {
	return s.cfg.Server.Addr
}
