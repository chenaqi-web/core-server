package rpc

import (
	"context"
	"core-server/internal/model/entity"
	"errors"
	"strings"

	"core-server/internal/application"
	"core-server/internal/rpc/authpb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthRPC struct {
	authpb.UnimplementedAuthServiceServer
	userService *application.UserService
}

func NewAuthRPC(userService *application.UserService) *AuthRPC {
	return &AuthRPC{userService: userService}
}

func (a *AuthRPC) Login(ctx context.Context, request *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	username := strings.TrimSpace(request.GetUsername())
	password := request.GetPassword()
	if username == "" || password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	user, err := a.userService.Login(ctx, username, password)
	if err != nil {
		return nil, toLoginError(err)
	}
	return buildLoginResponse(user), nil
}

func (a *AuthRPC) EmailLogin(ctx context.Context, request *authpb.EmailLoginRequest) (*authpb.LoginResponse, error) {
	email := strings.TrimSpace(request.GetEmail())
	if email == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	user, err := a.userService.EmailLogin(ctx, email)
	if err != nil {
		return nil, toLoginError(err)
	}
	return buildLoginResponse(user), nil
}

// =====================================================================================================================

func buildLoginResponse(user *entity.User) *authpb.LoginResponse {
	return &authpb.LoginResponse{
		User: &authpb.UserInfo{
			Id:          user.ID,
			Username:    user.Name,
			Email:       user.Email,
			Phone:       user.Phone,
			Avatar:      user.Avatar,
			Sex:         user.Sex,
			Age:         uint32(user.Age),
			Role:        user.Role,
			Status:      user.Status,
			AuthVersion: user.AuthVersion,
		},
	}
}

func toLoginError(err error) error {
	switch {
	case errors.Is(err, application.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "authentication failed")
	case errors.Is(err, application.ErrUserDisabled):
		return status.Error(codes.PermissionDenied, "user is disabled")
	default:
		return status.Error(codes.Internal, "login failed")
	}
}
