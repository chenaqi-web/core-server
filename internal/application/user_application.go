package application

import (
	"context"
	"core-server/internal/domain"
	"core-server/internal/infras/clog"
	"core-server/internal/model/dto"
	"core-server/internal/model/entity"
	"core-server/internal/utils"
	"database/sql"
	"errors"

	"go.uber.org/zap"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserBlocked        = errors.New("USER_BLOCKED")
	ErrEmailAlreadyInUse  = errors.New("email is already registered")
	ErrUserNotFound       = errors.New("user not found")
)

type UserService struct {
	repo domain.UserRepoDomain
	log  *clog.Log
}

func NewUserService(
	repo domain.UserRepoDomain,
	log *clog.Log,
) *UserService {
	return &UserService{
		repo: repo,
		log:  log,
	}
}

// 用户登入方面

func (s *UserService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// 1.判断用户是否存在
	user, err := s.repo.GetByName(ctx, req.Username)
	if err != nil {
		// 数据库错误
		s.log.Error("UserService/Login error:", zap.Error(err))
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// 2.判断密码是否正确
	if user.Password != utils.Bcrypt(req.Password) {
		return nil, ErrInvalidCredentials
	}

	// 3.判断用户是否被拉黑
	if user.Status != entity.StatusApproved {
		return nil, ErrUserBlocked
	}
	return dto.ToLoginResponse(user), nil
}

func (s *UserService) EmailLogin(ctx context.Context, req *dto.EmailLoginRequest) (*dto.LoginResponse, error) {
	email := req.Email
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		s.log.Error("UserService/EmailLogin error:", zap.Error(err))
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if user.Status != entity.StatusApproved {
		return nil, ErrUserBlocked
	}
	return dto.ToLoginResponse(user), nil
}

func (s *UserService) Register(ctx context.Context, req *dto.RegisterRequest) error {
	existing, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.log.Error("UserService/Register error:", zap.Error(err))
		return err
	}
	if existing != nil {
		return ErrEmailAlreadyInUse
	}

	user := &entity.User{
		Name:     req.Username,
		Email:    req.Email,
		Password: utils.Bcrypt(req.Password),
		Role:     entity.UserRoleUser,
		Status:   entity.StatusApproved,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		s.log.Error("UserService/Register error:", zap.Error(err))
		return err
	}

	return nil
}

func (s *UserService) ForgotPassword(ctx context.Context, req *dto.ForgotPasswordRequest) error {
	// 1.校验两次密码是否相同
	if req.Password != req.Confirm {
		return errors.New("passwords do not match")
	}

	// 2.判断该邮箱是否存在
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.log.Error("UserService/ForgotPassword error:", zap.Error(err))
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// 3.更新密码
	err = s.repo.UpdatePassword(ctx, user.ID, utils.Bcrypt(req.Password))
	if err != nil {
		s.log.Error("UserService/ForgotPassword error:", zap.Error(err))
		return err
	}

	return nil
}

// =====================================================================================================================

// 用户信息方面

func (s *UserService) GetProfile(ctx context.Context, req *dto.GetProfileRequest) (*dto.GetProfileResponse, error) {
	// 1.拿到用户的基础信息
	userMsg, err := s.repo.GetByID(ctx, req.UserID)
	if err != nil {
		s.log.Error("GetProfile error", zap.Error(err))
		return nil, err
	}
	if userMsg == nil {
		return nil, ErrUserNotFound
	}

	// 2.拿到用户的计数信息
	userStat, err := s.repo.GetStat(ctx, req.UserID)
	if err != nil {
		s.log.Error("GetProfile error", zap.Error(err))
		return nil, err
	}

	return dto.ToGetProfileResponse(userMsg, userStat), nil
}

func (s *UserService) UpdateProfile(ctx context.Context, req *dto.UpdateProfileRequest) error {
	user, err := s.repo.GetByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		s.log.Error("UpdateProfile error", zap.Error(err))
		return err
	}

	user.Name = req.Username
	user.Phone = req.Phone
	user.Sex = req.Sex
	user.Age = uint64(req.Age)
	if err := s.repo.UpdateProfile(ctx, user); err != nil {
		s.log.Error("UpdateProfile error", zap.Error(err))
		return err
	}
	return nil
}

// todo 后续修改
func (s *UserService) UpdateAvatar(ctx context.Context, req *dto.UpdateAvatarRequest) (*dto.UserAvatarResponse, error) {
	user, err := s.repo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if err := s.repo.UpdateAvatar(ctx, req.UserID, req.Avatar); err != nil {
		s.log.Error("UpdateAvatar error", zap.Error(err))
		return nil, err
	}
	user.Avatar = req.Avatar
	return dto.ToUserAvatarResponse(user.Avatar), nil
}

// =====================================================================================================================

// 管理用户方面

func (s *UserService) UpdateStatus(ctx context.Context, req *dto.UpdateUserStatusRequest) error {
	if err := s.repo.UpdateStatus(ctx, req.UserID, req.Status); err != nil {
		s.log.Error("UserService/UpdateStatus error:", zap.Error(err))
		return err
	}
	return nil
}

func (s *UserService) List(ctx context.Context, req *dto.ListUsersRequest) (*dto.ListUsersResponse, error) {
	users, total, err := s.repo.List(ctx, req.PageSize, (req.Page-1)*req.PageSize)
	if err != nil {
		s.log.Error("UserService/List error:", zap.Error(err))
		return nil, err
	}
	return dto.ToListUsersResponse(users, total), nil
}

func (s *UserService) SearchUser(ctx context.Context, req *dto.SearchUsersRequest) (*dto.SearchUsersResponse, error) {
	users, total, err := s.repo.Search(ctx, req.Keyword, req.PageSize, (req.Page-1)*req.PageSize)
	if err != nil {
		s.log.Error("UserService/Search error:", zap.Error(err))
		return nil, err
	}
	return dto.ToSearchUsersResponse(users, total), nil
}
