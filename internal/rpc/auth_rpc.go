package rpc

import (
	"context"
	"errors"
	"net/mail"
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

// 登录
func (a *AuthRPC) Login(ctx context.Context, request *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	username := strings.TrimSpace(request.GetUsername())
	password := request.GetPassword()
	if username == "" || password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	user, err := a.userService.Login(ctx, username, password)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidCredentials):
			return nil, status.Error(codes.Unauthenticated, "authentication failed")
		case errors.Is(err, application.ErrUserDisabled):
			return nil, status.Error(codes.PermissionDenied, "user is disabled")
		default:
			return nil, status.Error(codes.Internal, "login failed")
		}
	}

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
	}, nil
}

// 发送邮箱验证码
func (a *AuthRPC) SendEmailCode(ctx context.Context, request *authpb.SendEmailCodeRequest) (*authpb.SendEmailCodeResponse, error) {
	email := strings.TrimSpace(request.GetEmail())
	address, err := mail.ParseAddress(email)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid email")
	}
	email = address.Address

	purpose := request.GetPurpose()
	if purpose != authpb.EmailCodePurpose_EMAIL_CODE_PURPOSE_REGISTER &&
		purpose != authpb.EmailCodePurpose_EMAIL_CODE_PURPOSE_RESET_PASSWORD {
		return nil, status.Error(codes.InvalidArgument, "invalid email code purpose")
	}

	if err := a.userService.SendEmailCode(ctx, email, int32(purpose)); err != nil {
		if errors.Is(err, application.ErrEmailTooFrequent) {
			return nil, status.Error(codes.ResourceExhausted, "email code sent too frequently")
		}
		return nil, status.Error(codes.Internal, "send email code failed")
	}
	return &authpb.SendEmailCodeResponse{Success: true}, nil
}
