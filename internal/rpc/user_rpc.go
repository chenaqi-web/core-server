package rpc

import (
	"backend/core-server/internal/application"
	"backend/core-server/internal/rpc/userpb"
	"context"
)

type UserRPC struct {
	userpb.UnimplementedUserServiceServer
	UserService *application.UserService
}

func NewUserRPC(userService *application.UserService) *UserRPC {
	return &UserRPC{UserService: userService}
}

func (u *UserRPC) Login(ctx context.Context, request *userpb.LoginReq) (*userpb.LoginResp, error) {
	// todo
	return nil, nil
}
