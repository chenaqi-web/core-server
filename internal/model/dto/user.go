package dto

import (
	"core-server/internal/model/entity"
	"core-server/internal/rpc/authpb"
)

func ToLoginResponse(user *entity.User) *authpb.LoginResponse {
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
