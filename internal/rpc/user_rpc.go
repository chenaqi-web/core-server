package rpc

import (
	"context"
	"errors"

	"core-server/internal/application"
	"core-server/internal/model/dto"
	"core-server/internal/rpc/userpb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserRPC struct {
	userpb.UnimplementedUserServiceServer
	UserService *application.UserService
}

func NewUserRPC(userService *application.UserService) *UserRPC {
	return &UserRPC{UserService: userService}
}

func (u *UserRPC) Login(_ context.Context, _ *userpb.LoginReq) (*userpb.LoginResp, error) {
	return &userpb.LoginResp{}, nil
}

func (u *UserRPC) GetProfile(ctx context.Context, request *userpb.GetProfileRequest) (*userpb.GetProfileResponse, error) {
	user, err := u.UserService.GetProfile(ctx, request.GetUserId())
	if err != nil {
		return nil, userError(err)
	}
	return &userpb.GetProfileResponse{User: dto.ToUserInfo(user)}, nil
}

func (u *UserRPC) UpdateProfile(ctx context.Context, request *userpb.UpdateProfileRequest) (*userpb.UpdateProfileResponse, error) {
	user, err := u.UserService.UpdateProfile(
		ctx,
		request.GetUserId(),
		request.GetUsername(),
		request.GetPhone(),
		request.GetSex(),
		request.GetAge(),
	)
	if err != nil {
		return nil, userError(err)
	}
	return &userpb.UpdateProfileResponse{User: dto.ToUserInfo(user)}, nil
}

func (u *UserRPC) UpdateAvatar(ctx context.Context, request *userpb.UpdateAvatarRequest) (*userpb.UpdateAvatarResponse, error) {
	user, err := u.UserService.UpdateAvatar(ctx, request.GetUserId(), request.GetAvatar())
	if err != nil {
		return nil, userError(err)
	}
	return &userpb.UpdateAvatarResponse{User: dto.ToUserInfo(user)}, nil
}

func userError(err error) error {
	switch {
	case errors.Is(err, application.ErrInvalidUserProfile):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, application.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, application.ErrUsernameInUse):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, application.ErrUserDisabled):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, "user operation failed")
	}
}
