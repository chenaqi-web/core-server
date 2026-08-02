package rpc

import (
	"context"
	"core-server/internal/application"
	"core-server/internal/rpc/userpb"
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
