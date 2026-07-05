package rpc

import (
	"backend/core-server/internal/application"
	"backend/core-server/internal/rpc/likepb"
	"context"
)

type LikeRPC struct {
	likepb.UnimplementedLikeServiceServer
	LikeService *application.LikeService
}

func NewLikeRPC(likeService *application.LikeService) *LikeRPC {
	return &LikeRPC{
		UnimplementedLikeServiceServer: likepb.UnimplementedLikeServiceServer{},
		LikeService:                    likeService,
	}
}

func (l *LikeRPC) ThumbUp(ctx context.Context, request *likepb.ThumbUpRequest) (*likepb.ThumbUpResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *LikeRPC) CancelThumbUp(ctx context.Context, request *likepb.CancelThumbUpRequest) (*likepb.CancelThumbUpResponse, error) {
	//TODO implement me
	panic("implement me")
}
