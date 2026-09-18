package rpc

import (
	"context"
	"core-server/internal/application"
	"core-server/internal/model/dto"
	"core-server/internal/rpc/authpb"
)

type AuthRPC struct {
	authpb.UnimplementedAuthServiceServer
	userService *application.UserService
}

func NewAuthRPC(userService *application.UserService) *AuthRPC {
	return &AuthRPC{userService: userService}
}

func (a *AuthRPC) Login(ctx context.Context, request *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	res, err := a.userService.Login(ctx, &dto.LoginRequest{Username: request.GetUsername(), Password: request.GetPassword()})
	if err != nil {
		return nil, err
	}

	return &authpb.LoginResponse{
		Id:       res.Id,
		Username: res.Name,
		Avatar:   res.Avatar,
		Role:     res.Role,
	}, nil
}

func (a *AuthRPC) EmailLogin(ctx context.Context, request *authpb.EmailLoginRequest) (*authpb.LoginResponse, error) {
	res, err := a.userService.EmailLogin(ctx, &dto.EmailLoginRequest{Email: request.GetEmail()})
	if err != nil {
		return nil, err
	}
	return &authpb.LoginResponse{
		Id:       res.Id,
		Username: res.Name,
		Avatar:   res.Avatar,
		Role:     res.Role,
	}, nil
}

func (a *AuthRPC) Register(ctx context.Context, request *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	err := a.userService.Register(ctx, &dto.RegisterRequest{
		Username: request.GetUsername(),
		Email:    request.GetEmail(),
		Password: request.GetPassword(),
	})
	if err != nil {
		return nil, err
	}
	return &authpb.RegisterResponse{Success: true}, nil
}

func (a *AuthRPC) ForgotPassword(ctx context.Context, request *authpb.ForgotPasswordRequest) (*authpb.ForgotPasswordResponse, error) {
	err := a.userService.ForgotPassword(ctx, &dto.ForgotPasswordRequest{
		Email:    request.GetEmail(),
		Password: request.GetNewPassword(),
		Confirm:  request.GetConfirmPassword(),
	})
	if err != nil {
		return nil, err
	}
	return &authpb.ForgotPasswordResponse{Success: true}, nil
}
