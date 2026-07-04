package rpc

import (
	"context"

	"backend/core-server/internal/rpc/healthpb"
)

type HealthPRC struct {
	// 每一个RPC下包含
	// 1. rpc的server
	// 2. application的service
	// todo 3. logger日志暂未接入
	healthpb.UnimplementedHealthServiceServer
}

func NewHealthPRC() *HealthPRC {
	return &HealthPRC{}
}

func (h *HealthPRC) Ping(_ context.Context, _ *healthpb.PingRequest) (*healthpb.PingResponse, error) {
	return &healthpb.PingResponse{Message: "pong"}, nil
}
