package rpc

import (
	"context"
	"core-server/internal/application"
	"core-server/internal/model/dto"
	"core-server/internal/rpc/authpb"
	"errors"

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
	username := request.GetUsername()
	password := request.GetPassword()
	if username == "" || password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	user, err := a.userService.Login(ctx, username, password)
	if err != nil {
		return nil, toLoginError(err)
	}
	return dto.ToLoginResponse(user), nil
}

func (a *AuthRPC) Register(ctx context.Context, request *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	user, err := a.userService.Register(ctx, request.GetUsername(), request.GetEmail(), request.GetPassword(), request.GetConfirmPassword())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &authpb.RegisterResponse{Success: true, User: dto.ToLoginResponse(user).GetUser()}, nil
}

func (a *AuthRPC) EmailLogin(ctx context.Context, request *authpb.EmailLoginRequest) (*authpb.LoginResponse, error) {
	email := request.GetEmail()
	if email == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	user, err := a.userService.EmailLogin(ctx, email)
	if err != nil {
		return nil, toLoginError(err)
	}
	return dto.ToLoginResponse(user), nil
}

func (a *AuthRPC) ForgotPassword(ctx context.Context, request *authpb.ForgotPasswordRequest) (*authpb.ForgotPasswordResponse, error) {
	err := a.userService.ForgotPassword(ctx, request.GetEmail(), request.GetNewPassword(), request.GetConfirmPassword())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &authpb.ForgotPasswordResponse{Success: true}, nil
}

// =====================================================================================================================

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
