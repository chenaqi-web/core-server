package rpc

import (
	"context"
	"core-server/internal/application"
	"core-server/internal/model/dto"
	"core-server/internal/rpc/userpb"
	"time"
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
		Id:                res.ID,
		Username:          res.Username,
		Email:             res.Email,
		Phone:             res.Phone,
		Avatar:            res.Avatar,
		Sex:               res.Sex,
		Birthday:          formatBirthday(res.Birthday),
		Role:              res.Role,
		Status:            res.Status,
		FollowersCount:    res.FollowersCount,
		FollowingCount:    res.FollowingCount,
		LikeCount:         res.LikeCount,
		ReceiveLikeCount:  res.ReceiveLikeCount,
		FavorCount:        res.FavorCount,
		ReceiveFavorCount: res.ReceiveFavorCount,
	}, nil
}

func (u *UserRPC) UpdateProfile(ctx context.Context, request *userpb.UpdateProfileRequest) (*userpb.UpdateProfileResponse, error) {
	err := u.UserService.UpdateProfile(ctx, &dto.UpdateProfileRequest{
		UserID:   request.GetUserId(),
		Username: request.GetUsername(),
		Phone:    request.GetPhone(),
		Sex:      request.GetSex(),
		Birthday: parseBirthday(request.GetBirthday()),
	})
	if err != nil {
		return nil, err
	}
	return &userpb.UpdateProfileResponse{Success: true}, nil
}

func (u *UserRPC) UpdateAvatar(ctx context.Context, request *userpb.UpdateAvatarRequest) (*userpb.UpdateAvatarResponse, error) {
	res, err := u.UserService.UpdateAvatar(ctx, &dto.UpdateAvatarRequest{UserID: request.GetUserId(), Avatar: request.GetAvatar()})
	if err != nil {
		return nil, err
	}
	return &userpb.UpdateAvatarResponse{Url: res.Url}, nil
}

func (u *UserRPC) ListUsers(ctx context.Context, request *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	res, err := u.UserService.List(ctx, &dto.ListUsersRequest{Page: request.GetPage(), PageSize: request.GetPageSize()})
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

func (u *UserRPC) SearchUsers(ctx context.Context, request *userpb.SearchUsersRequest) (*userpb.SearchUsersResponse, error) {
	res, err := u.UserService.SearchUser(ctx, &dto.SearchUsersRequest{
		Keyword:  request.GetKeyword(),
		Page:     request.GetPage(),
		PageSize: request.GetPageSize(),
	})
	if err != nil {
		return nil, err
	}
	items := make([]*userpb.UserInfo, 0, len(res.Users))
	for _, user := range res.Users {
		items = append(items, ConvertToUserInfo(user))
	}
	return &userpb.SearchUsersResponse{
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
		Id:       res.ID,
		Username: res.Username,
		Email:    res.Email,
		Phone:    res.Phone,
		Avatar:   res.Avatar,
		Sex:      res.Sex,
		Birthday: formatBirthday(res.Birthday),
		Role:     res.Role,
		Status:   res.Status,
	}
}

func parseBirthday(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		parsed, _ = time.Parse(time.RFC3339, value)
	}
	return parsed
}

func formatBirthday(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02")
}
