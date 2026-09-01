package dto

import (
	"core-server/internal/model/entity"
)

type UserInfo struct {
	ID               uint64
	Username         string
	Email            string
	Phone            string
	Avatar           string
	Sex              string
	Age              uint32
	Role             string
	Status           string
	LikeCount        uint64
	ReceiveLikeCount uint64
}

type LoginRequest struct {
	Username string
	Password string
}
type EmailLoginRequest struct {
	Email string
}
type LoginResponse struct {
	Id               uint64
	Name             string
	Password         string
	Phone            string
	Avatar           string
	Email            string
	Role             string
	Sex              string
	Age              uint32
	LikeCount        uint64
	ReceiveLikeCount uint64
	Status           string
}

type RegisterRequest struct {
	Username,
	Email,
	Password string
}
type ForgotPasswordRequest struct {
	Email,
	Password,
	Confirm string
}

// =====================================================================================================================

type GetProfileRequest struct{ UserID uint64 }
type UpdateProfileRequest struct {
	UserID               uint64
	Username, Phone, Sex string
	Age                  uint32
}
type UpdateAvatarRequest struct {
	UserID uint64
	Avatar string
}

type UpdateUserStatusRequest struct {
	UserID uint64
	Status string
}

type UserMsgResponse struct {
	*UserInfo
}

type UserAvatarResponse struct {
	Url string
}
type ListUsersRequest struct {
	Keyword        string
	Page, PageSize uint32
}

type ListUsersResponse struct {
	Users []*UserInfo
	Total uint64
}

func ToUserInfo(user *entity.User) *UserInfo {
	if user == nil {
		return nil
	}
	return &UserInfo{
		ID:               user.ID,
		Username:         user.Name,
		Email:            user.Email,
		Phone:            user.Phone,
		Avatar:           user.Avatar,
		Sex:              user.Sex,
		Age:              uint32(user.Age),
		Role:             user.Role,
		Status:           user.Status,
		LikeCount:        user.LikeCount,
		ReceiveLikeCount: user.ReceiveLikeCount,
	}
}

func ToLoginResponse(user *entity.User) *LoginResponse {
	return &LoginResponse{
		Id:               user.ID,
		Name:             user.Name,
		Email:            user.Email,
		Phone:            user.Phone,
		Avatar:           user.Avatar,
		Sex:              user.Sex,
		Age:              uint32(user.Age),
		Role:             user.Role,
		LikeCount:        user.LikeCount,
		ReceiveLikeCount: user.ReceiveLikeCount,
		Status:           user.Status,
	}
}

func ToUserMsgResponse(user *entity.User) *UserMsgResponse {
	return &UserMsgResponse{
		ToUserInfo(user),
	}
}

func ToUserAvatarResponse(url string) *UserAvatarResponse {
	return &UserAvatarResponse{
		Url: url,
	}
}

func ToListUsersResponse(users []*entity.User, total uint64) *ListUsersResponse {
	if users == nil {
		return &ListUsersResponse{
			Users: []*UserInfo{},
			Total: total,
		}
	}

	userInfos := make([]*UserInfo, 0, len(users))
	for _, user := range users {
		if userInfo := ToUserInfo(user); userInfo != nil {
			userInfos = append(userInfos, userInfo)
		}
	}

	return &ListUsersResponse{
		Users: userInfos,
		Total: total,
	}
}
