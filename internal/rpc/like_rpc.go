package rpc

import (
	"context"

	"backend/core-server/internal/application"
	"backend/core-server/internal/rpc/likepb"
)

type LikeRPC struct {
	likepb.UnimplementedLikeServiceServer
	LikeService *application.LikeService
}

func NewLikeRPC(likeService *application.LikeService) *LikeRPC {
	return &LikeRPC{LikeService: likeService}
}

func (l *LikeRPC) ThumbUp(ctx context.Context, request *likepb.ThumbUpRequest) (*likepb.ThumbUpResponse, error) {
	if err := l.LikeService.ThumbUp(
		ctx,
		request.GetUserID(),
		request.GetObjectType(),
		request.GetObjectID(),
		request.GetObjectOwnerID(),
	); err != nil {
		return nil, err
	}
	return &likepb.ThumbUpResponse{Success: true}, nil
}

func (l *LikeRPC) CancelThumbUp(ctx context.Context, request *likepb.CancelThumbUpRequest) (*likepb.CancelThumbUpResponse, error) {
	if err := l.LikeService.CancelThumbUp(
		ctx,
		request.GetUserID(),
		request.GetObjectType(),
		request.GetObjectID(),
		request.GetObjectOwnerID(),
	); err != nil {
		return nil, err
	}
	return &likepb.CancelThumbUpResponse{Success: true}, nil
}
