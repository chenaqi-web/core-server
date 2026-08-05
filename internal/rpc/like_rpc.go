package rpc

import (
	"context"
	"fmt"
	"strconv"

	"core-server/internal/application"
	"core-server/internal/rpc/articlepb"
	"core-server/internal/rpc/likepb"
)

type LikeRPC struct {
	likepb.UnimplementedLikeServiceServer
	LikeService *application.LikeService
}

func NewLikeRPC(likeService *application.LikeService) *LikeRPC {
	return &LikeRPC{LikeService: likeService}
}

func (l *LikeRPC) ThumbUp(ctx context.Context, request *likepb.ThumbUpRequest) (*likepb.ThumbUpResponse, error) {
	userID, err := strconv.ParseUint(request.GetUserID(), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid userID: %w", err)
	}
	objectID, err := strconv.ParseUint(request.GetObjectID(), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid objectID: %w", err)
	}
	if err := l.LikeService.ThumbUp(
		ctx,
		userID,
		request.GetObjectType(),
		objectID,
	); err != nil {
		return nil, err
	}
	return &likepb.ThumbUpResponse{Success: true}, nil
}

func (l *LikeRPC) CancelThumbUp(ctx context.Context, request *likepb.CancelThumbUpRequest) (*likepb.CancelThumbUpResponse, error) {
	userID, err := strconv.ParseUint(request.GetUserID(), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid userID: %w", err)
	}
	objectID, err := strconv.ParseUint(request.GetObjectID(), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid objectID: %w", err)
	}
	if err := l.LikeService.CancelThumbUp(
		ctx,
		userID,
		request.GetObjectType(),
		objectID,
	); err != nil {
		return nil, err
	}
	return &likepb.CancelThumbUpResponse{Success: true}, nil
}

func (l *LikeRPC) PageQueryUserLikeList(ctx context.Context, request *likepb.PageQueryUserLikeListRequest) (*likepb.PageQueryUserLikeListResponse, error) {
	userID, err := strconv.ParseUint(request.GetUserID(), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid userID: %w", err)
	}
	articles, total, err := l.LikeService.UserLikeList(ctx, userID, request.GetObjectType(), int(request.GetPage()), int(request.GetPageSize()))
	if err != nil {
		return nil, err
	}
	items := make([]*articlepb.Article, 0, len(articles))
	for _, article := range articles {
		if article == nil || article.Article == nil {
			continue
		}
		items = append(items, toArticlePB(article))
	}
	return &likepb.PageQueryUserLikeListResponse{Articles: items, Total: total}, nil
}
