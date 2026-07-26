package rpc

import (
	"context"

	"backend/core-server/internal/application"
	"backend/core-server/internal/model/dto"
	"backend/core-server/internal/rpc/commentpb"
)

type CommentRPC struct {
	commentpb.UnimplementedCommentServiceServer
	CommentService *application.CommentService
}

func NewCommentRPC(commentService *application.CommentService) *CommentRPC {
	return &CommentRPC{CommentService: commentService}
}

func (c *CommentRPC) CreateComment(ctx context.Context, req *commentpb.CreateCommentReq) (*commentpb.CreateCommentResp, error) {
	resp, err := c.CommentService.CreateComment(ctx, dto.CreateCommentRequestFromPB(req))
	if err != nil {
		return nil, err
	}
	return &commentpb.CreateCommentResp{Success: resp}, nil
}

func (c *CommentRPC) CreateReply(ctx context.Context, req *commentpb.CreateReplyReq) (*commentpb.CreateReplyResp, error) {
	resp, err := c.CommentService.CreateReply(ctx, dto.CreateReplyRequestFromPB(req))
	if err != nil {
		return nil, err
	}
	return &commentpb.CreateReplyResp{Success: resp}, nil
}

func (c *CommentRPC) DeleteComment(ctx context.Context, req *commentpb.DeleteCommentReq) (*commentpb.DeleteCommentResp, error) {
	if err := c.CommentService.DeleteComment(ctx, dto.DeleteCommentRequestFromPB(req)); err != nil {
		return nil, err
	}
	return &commentpb.DeleteCommentResp{Success: true}, nil
}

func (c *CommentRPC) GetArticleComments(ctx context.Context, req *commentpb.GetArticleCommentsReq) (*commentpb.GetArticleCommentsResp, error) {
	resp, err := c.CommentService.GetArticleComments(ctx, dto.GetArticleCommentsRequestFromPB(req))
	if err != nil {
		return nil, err
	}
	return resp.ToPB(), nil
}

func (c *CommentRPC) GetCommentReplies(ctx context.Context, req *commentpb.GetCommentRepliesReq) (*commentpb.GetCommentRepliesResp, error) {
	resp, err := c.CommentService.GetCommentReplies(ctx, dto.GetCommentRepliesRequestFromPB(req))
	if err != nil {
		return nil, err
	}
	return resp.ToPB(), nil
}

func (c *CommentRPC) GetUserComments(ctx context.Context, req *commentpb.GetUserCommentsReq) (*commentpb.GetUserCommentsResp, error) {
	resp, err := c.CommentService.GetUserComments(ctx, dto.GetUserCommentsRequestFromPB(req))
	if err != nil {
		return nil, err
	}
	return resp.ToPB(), nil
}
