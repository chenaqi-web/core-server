package dto

import (
	"time"

	"backend/core-server/internal/model/entity"
	"backend/core-server/internal/rpc/commentpb"
)

// 创建评论

type CreateCommentRequest struct {
	ArticleID uint64
	UserID    uint64
	Content   string
}

func CreateCommentRequestFromPB(req *commentpb.CreateCommentReq) *CreateCommentRequest {
	if req == nil {
		return nil
	}
	return &CreateCommentRequest{
		ArticleID: req.GetArticleId(),
		UserID:    req.GetUserId(),
		Content:   req.GetContent(),
	}
}

//创建回复

type CreateReplyRequest struct {
	ArticleID   uint64
	RootID      uint64
	UserID      uint64
	ReplyToID   uint64
	ReplyToName string
	Content     string
}

type CreateReplyResponse struct {
	ReplyID uint64
}

type DeleteCommentRequest struct {
	ID     uint64
	UserID uint64
}

type GetArticleCommentsRequest struct {
	ArticleID uint64
	Page      int32
	Size      int32
	OrderBy   string
}

type GetArticleCommentsResponse struct {
	Comments []*CommentInfoDTO
	Page     int32
	Size     int32
	Total    int32
}

type GetCommentRepliesRequest struct {
	RootID uint64
	Page   int32
	Size   int32
}

type GetCommentRepliesResponse struct {
	Replies []*CommentInfoDTO
	Page    int32
	Size    int32
	Total   int32
}

type GetUserCommentsRequest struct {
	UserID uint64
	Page   int32
	Size   int32
}

type GetUserCommentsResponse struct {
	Comments []*CommentInfoDTO
	Page     int32
	Size     int32
	Total    int32
}

type CommentInfoDTO struct {
	ID          uint64
	ArticleID   uint64
	UserID      uint64
	ParentID    uint64
	RootID      uint64
	ReplyToID   uint64
	ReplyToName string
	Content     string
	LikeCount   uint32
	ChildCount  uint32
	CreatedAt   string
	UserName    string
	UserAvatar  string
	Replies     []*CommentInfoDTO
}

func CreateReplyRequestFromPB(req *commentpb.CreateReplyReq) *CreateReplyRequest {
	if req == nil {
		return nil
	}
	return &CreateReplyRequest{

		RootID:      req.GetRootId(),
		UserID:      req.GetUserId(),
		ReplyToID:   req.GetReplyToId(),
		ReplyToName: req.GetReplyToName(),
		Content:     req.GetContent(),
	}
}

func (r *CreateReplyResponse) ToPB() *commentpb.CreateReplyResp {
	if r == nil {
		return nil
	}
	return &commentpb.CreateReplyResp{ReplyId: r.ReplyID}
}

func DeleteCommentRequestFromPB(req *commentpb.DeleteCommentReq) *DeleteCommentRequest {
	if req == nil {
		return nil
	}
	return &DeleteCommentRequest{
		ID:     req.GetId(),
		UserID: req.GetUserId(),
	}
}

func GetArticleCommentsRequestFromPB(req *commentpb.GetArticleCommentsReq) *GetArticleCommentsRequest {
	if req == nil {
		return nil
	}
	return &GetArticleCommentsRequest{
		ArticleID: req.GetArticleId(),
		Page:      req.GetPage(),
		Size:      req.GetSize(),
		OrderBy:   req.GetOrderBy(),
	}
}

func (r *GetArticleCommentsResponse) ToPB() *commentpb.GetArticleCommentsResp {
	if r == nil {
		return nil
	}
	return &commentpb.GetArticleCommentsResp{
		Comments: CommentInfoListToPB(r.Comments),
		Page:     r.Page,
		Size:     r.Size,
		Total:    r.Total,
	}
}

func GetCommentRepliesRequestFromPB(req *commentpb.GetCommentRepliesReq) *GetCommentRepliesRequest {
	if req == nil {
		return nil
	}
	return &GetCommentRepliesRequest{
		RootID: req.GetRootId(),
		Page:   req.GetPage(),
		Size:   req.GetSize(),
	}
}

func (r *GetCommentRepliesResponse) ToPB() *commentpb.GetCommentRepliesResp {
	if r == nil {
		return nil
	}
	return &commentpb.GetCommentRepliesResp{
		Replies: CommentInfoListToPB(r.Replies),
		Page:    r.Page,
		Size:    r.Size,
		Total:   r.Total,
	}
}

func GetUserCommentsRequestFromPB(req *commentpb.GetUserCommentsReq) *GetUserCommentsRequest {
	if req == nil {
		return nil
	}
	return &GetUserCommentsRequest{
		UserID: req.GetUserId(),
		Page:   req.GetPage(),
		Size:   req.GetSize(),
	}
}

func (r *GetUserCommentsResponse) ToPB() *commentpb.GetUserCommentsResp {
	if r == nil {
		return nil
	}
	return &commentpb.GetUserCommentsResp{
		Comments: CommentInfoListToPB(r.Comments),
		Page:     r.Page,
		Size:     r.Size,
		Total:    r.Total,
	}
}

func CommentInfoFromEntity(c *entity.Comment, author *entity.User) *CommentInfoDTO {
	if c == nil {
		return nil
	}
	info := &CommentInfoDTO{
		ID:          c.ID,
		ArticleID:   c.ArticleID,
		UserID:      c.UserID,
		ParentID:    c.ParentID,
		RootID:      c.RootID,
		ReplyToID:   c.ReplyToID,
		ReplyToName: c.ReplyToName,
		Content:     c.Content,
		LikeCount:   c.LikeCount,
		ChildCount:  c.ChildCount,
		CreatedAt:   c.CreatedAt.Format(time.DateTime),
	}
	if author != nil {
		info.UserName = author.Name
		info.UserAvatar = author.Avatar
	}
	return info
}

func CommentInfoToPB(info *CommentInfoDTO) *commentpb.CommentInfo {
	if info == nil {
		return nil
	}
	pb := &commentpb.CommentInfo{
		Id:          info.ID,
		ArticleId:   info.ArticleID,
		UserId:      info.UserID,
		ParentId:    info.ParentID,
		RootId:      info.RootID,
		ReplyToId:   info.ReplyToID,
		ReplyToName: info.ReplyToName,
		Content:     info.Content,
		LikeCount:   info.LikeCount,
		ChildCount:  info.ChildCount,
		CreatedAt:   info.CreatedAt,
		UserName:    info.UserName,
		UserAvatar:  info.UserAvatar,
	}
	if len(info.Replies) > 0 {
		pb.Replies = CommentInfoListToPB(info.Replies)
	}
	return pb
}

func CommentInfoListToPB(items []*CommentInfoDTO) []*commentpb.CommentInfo {
	if len(items) == 0 {
		return nil
	}
	result := make([]*commentpb.CommentInfo, 0, len(items))
	for _, item := range items {
		result = append(result, CommentInfoToPB(item))
	}
	return result
}
