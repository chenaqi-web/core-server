package rpc

import (
	"context"
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
		request.UserID,
		request.GetObjectType(),
		request.ObjectID,
	); err != nil {
		return nil, err
	}
	return &likepb.ThumbUpResponse{Success: true}, nil
}

func (l *LikeRPC) CancelThumbUp(ctx context.Context, request *likepb.CancelThumbUpRequest) (*likepb.CancelThumbUpResponse, error) {
	if err := l.LikeService.CancelThumbUp(
		ctx,
		request.UserID,
		request.GetObjectType(),
		request.ObjectID,
	); err != nil {
		return nil, err
	}
	return &likepb.CancelThumbUpResponse{Success: true}, nil
}

func (l *LikeRPC) PageQueryUserLikeList(ctx context.Context, request *likepb.PageQueryUserLikeListRequest) (*likepb.PageQueryUserLikeListResponse, error) {
	articles, total, err := l.LikeService.UserLikeList(ctx, request.GetUserID(), request.GetObjectType(), int(request.GetPage()), int(request.GetPageSize()))
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

func (l *LikeRPC) HasLike(ctx context.Context, request *likepb.HasArticleLikeRequest) (*likepb.HasArticleLikeResponse, error) {
	isLiked, err := l.LikeService.HasThumbUp(ctx, request.GetUserID(), request.GetObjectType(), request.GetObjectID())
	if err != nil {
		return nil, err
	}
	return &likepb.HasArticleLikeResponse{IsLiked: isLiked}, nil
}

func (l *LikeRPC) BatchLikeStatus(ctx context.Context, request *likepb.BatchCommentLikeStatusRequest) (*likepb.BatchLikeStatusResponse, error) {
	statuses, err := l.LikeService.BatchHasThumbUp(ctx, request.GetUserID(), request.GetObjectType(), request.GetObjectIDs())
	if err != nil {
		return nil, err
	}
	items := make([]*likepb.LikeStatus, 0, len(statuses))
	for _, objectId := range request.GetObjectIDs() {
		items = append(items, &likepb.LikeStatus{
			ObjectID: objectId,
			IsLiked:  statuses[objectId],
		})
	}
	return &likepb.BatchLikeStatusResponse{Items: items}, nil
}
