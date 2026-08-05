package rpc

import (
	"context"
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
	if err := l.LikeService.ThumbUp(
		ctx,
		request.GetUserID(),
		request.GetObjectType(),
		request.GetObjectID(),
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
	); err != nil {
		return nil, err
	}
	return &likepb.CancelThumbUpResponse{Success: true}, nil
}

func (l *LikeRPC) PageQueryUserLikeList(ctx context.Context, request *likepb.PageQueryUserLikeListRequest) (*likepb.PageQueryUserLikeListResponse, error) {
	articles, total, err := l.LikeService.UserLikeList(ctx, strconv.Itoa(int(request.GetUserID())), request.GetObjectType(), int(request.GetPage()), int(request.GetPageSize()))
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
