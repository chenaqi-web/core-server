package dto

import (
	"core-server/internal/model/entity"
	"core-server/internal/model/enum"
	"time"
)

type LoginRequest struct {
	Username string
	Password string
}
type EmailLoginRequest struct {
	Email string
}
type LoginResponse struct {
	Id     uint64
	Name   string
	Avatar string
	Status enum.UserStatus
	Role   enum.UserRole
}

type RegisterRequest struct {
	Username string
	Email    string
	Password string
}
type ForgotPasswordRequest struct {
	Email    string
	Password string
	Confirm  string
}

// =====================================================================================================================
type GetProfileRequest struct {
	UserID uint64
}
type GetProfileResponse struct {
	ID        uint64
	Username  string
	Email     string
	Phone     string
	Avatar    string
	Sex       enum.UserSex
	Signature string
	Birthday  time.Time
	Role      enum.UserRole
	Status    enum.UserStatus

	ArticleCount      uint64
	FollowersCount    uint64
	FollowingCount    uint64
	LikeCount         uint64
	ReceiveLikeCount  uint64
	FavorCount        uint64
	ReceiveFavorCount uint64
}

type UpdateProfileRequest struct {
	UserID                     uint64
	Username, Phone, Signature string
	Sex                        enum.UserSex
	Birthday                   time.Time
}

type UpdateAvatarRequest struct {
	UserID uint64
	Avatar string
}

type UserAvatarResponse struct {
	Url string
}

// =====================================================================================================================

type UpdateUserStatusRequest struct {
	UserID uint64
	Status enum.UserStatus
}

type ListUsersRequest struct {
	Page, PageSize int32
}

type UserInfo struct {
	ID        uint64
	Username  string
	Email     string
	Phone     string
	Avatar    string
	Sex       enum.UserSex
	Signature string
	Birthday  time.Time
	Role      enum.UserRole
	Status    enum.UserStatus
}

type ListUsersResponse struct {
	Users []*UserInfo
	Total uint64
}

type SearchUsersRequest struct {
	Keyword        string
	Page, PageSize int32
}

type SearchUsersResponse struct {
	Users []*UserInfo
	Total uint64
}

// =====================================================================================================================

func ToGetProfileResponse(user *entity.User, stat *entity.UserStat) *GetProfileResponse {
	return &GetProfileResponse{
		ID:                user.ID,
		Username:          user.Name,
		Email:             user.Email,
		Phone:             user.Phone,
		Avatar:            user.Avatar,
		Sex:               user.Sex,
		Signature:         user.Signature,
		Birthday:          user.Birthday,
		Role:              user.Role,
		Status:            user.Status,
		ArticleCount:      stat.ArticleCount,
		FollowersCount:    stat.FollowersCount,
		FollowingCount:    stat.FollowingCount,
		LikeCount:         stat.LikeCount,
		ReceiveLikeCount:  stat.ReceiveLikeCount,
		FavorCount:        stat.FavorCount,
		ReceiveFavorCount: stat.ReceiveFavorCount,
	}
}

func ToLoginResponse(user *entity.User) *LoginResponse {
	return &LoginResponse{
		Id:     user.ID,
		Name:   user.Name,
		Avatar: user.Avatar,
		Status: user.Status,
		Role:   user.Role,
	}
}

func ToUserAvatarResponse(url string) *UserAvatarResponse {
	return &UserAvatarResponse{
		Url: url,
	}
}

func ToUserInfo(user *entity.User) *UserInfo {
	if user == nil {
		return nil
	}
	return &UserInfo{
		ID:       user.ID,
		Username: user.Name,
		Email:    user.Email,
		Phone:    user.Phone,
		Avatar:   user.Avatar,
		Sex:      user.Sex,
		Birthday: user.Birthday,
		Role:     user.Role,
		Status:   user.Status,
	}
}

func ToListUsersResponse(users []*entity.User, total uint64) *ListUsersResponse {
	if users == nil {
		return nil
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

func ToSearchUsersResponse(users []*entity.User, total uint64) *SearchUsersResponse {
	listResponse := ToListUsersResponse(users, total)
	return &SearchUsersResponse{
		Users: listResponse.Users,
		Total: listResponse.Total,
	}
}
