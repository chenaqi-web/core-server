package rpc

import (
	"context"
	"core-server/internal/application"
	"core-server/internal/model/dto"
	"core-server/internal/rpc/userpb"
)

type UserRPC struct {
	userpb.UnimplementedUserServiceServer
	UserService *application.UserService
}

func NewUserRPC(userService *application.UserService) *UserRPC {
	return &UserRPC{UserService: userService}
}

func (u *UserRPC) GetProfile(ctx context.Context, request *userpb.GetProfileRequest) (*userpb.GetProfileResponse, error) {
	res, err := u.UserService.GetProfile(ctx, &dto.GetProfileRequest{UserID: request.GetUserId()})
	if err != nil {
		return nil, err
	}
	return &userpb.GetProfileResponse{
		User: ConvertToUserInfo(res.UserInfo),
	}, nil
}

func (u *UserRPC) UpdateProfile(ctx context.Context, request *userpb.UpdateProfileRequest) (*userpb.UpdateProfileResponse, error) {
	res, err := u.UserService.UpdateProfile(
		ctx,
		&dto.UpdateProfileRequest{UserID: request.GetUserId(), Username: request.GetUsername(), Phone: request.GetPhone(), Sex: request.GetSex(), Age: request.GetAge()},
	)
	if err != nil {
		return nil, err
	}
	return &userpb.UpdateProfileResponse{User: ConvertToUserInfo(res.UserInfo)}, nil
}

func (u *UserRPC) UpdateAvatar(ctx context.Context, request *userpb.UpdateAvatarRequest) (*userpb.UpdateAvatarResponse, error) {
	res, err := u.UserService.UpdateAvatar(ctx, &dto.UpdateAvatarRequest{UserID: request.GetUserId(), Avatar: request.GetAvatar()})
	if err != nil {
		return nil, err
	}
	return &userpb.UpdateAvatarResponse{Url: res.Url}, nil
}

func (u *UserRPC) ListUsers(ctx context.Context, request *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	res, err := u.UserService.List(ctx, &dto.ListUsersRequest{Keyword: request.GetKeyword(), Page: request.GetPage(), PageSize: request.GetPageSize()})
	if err != nil {
		return nil, err
	}
	items := make([]*userpb.UserInfo, 0, len(res.Users))
	for _, user := range res.Users {
		items = append(items, ConvertToUserInfo(user))
	}
	return &userpb.ListUsersResponse{
		Users: items,
		Total: res.Total,
	}, nil
}

func (u *UserRPC) UpdateUserStatus(ctx context.Context, request *userpb.UpdateUserStatusRequest) (*userpb.UpdateUserStatusResponse, error) {
	err := u.UserService.UpdateStatus(ctx, &dto.UpdateUserStatusRequest{UserID: request.GetUserId(), Status: request.GetStatus()})
	if err != nil {
		return nil, err
	}
	return &userpb.UpdateUserStatusResponse{Success: true}, nil
}

// =====================================================================================================================

func ConvertToUserInfo(res *dto.UserInfo) *userpb.UserInfo {
	if res == nil {
		return nil
	}
	return &userpb.UserInfo{
		Id:               res.ID,
		Username:         res.Username,
		Email:            res.Email,
		Phone:            res.Phone,
		Avatar:           res.Avatar,
		Sex:              res.Sex,
		Age:              res.Age,
		Role:             res.Role,
		Status:           res.Status,
		LikeCount:        res.LikeCount,
		ReceiveLikeCount: res.ReceiveLikeCount,
	}
}
