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
	ErrInvalidUserProfile = errors.New("invalid user profile")
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

func (s *UserService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// 1.判断用户是否存在
	user, err := s.repo.GetByName(ctx, req.Username)
	if err != nil {
		// 用户不存在
		if errors.Is(err, sql.ErrNoRows) {
			s.log.Info("UserService/Login info:", zap.Error(ErrUserNotFound))
			return nil, ErrUserNotFound
		}
		// 数据库错误
		s.log.Error("UserService/Login error:", zap.Error(err))
		return nil, err
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
		if errors.Is(err, sql.ErrNoRows) {
			s.log.Info("UserService/EmailLogin error:", zap.Error(ErrUserNotFound))
			return nil, ErrUserNotFound
		}
		s.log.Error("UserService/EmailLogin error:", zap.Error(err))
		return nil, err
	}

	if user.Status != entity.StatusApproved {
		return nil, ErrUserBlocked
	}
	return dto.ToLoginResponse(user), nil
}

func (s *UserService) Register(ctx context.Context, req *dto.RegisterRequest) error {
	username, email, password := req.Username, req.Email, req.Password
	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.log.Info("UserService/Register user not found", zap.String("username", username))
			return ErrUserNotFound
		}
		s.log.Error("UserService/Register error:", zap.Error(err))
		return err
	}
	if existing != nil {
		return ErrEmailAlreadyInUse
	}

	user := &entity.User{
		Name:     username,
		Email:    email,
		Password: utils.Bcrypt(password),
		Role:     entity.UserRoleUser,
		Status:   entity.StatusApproved,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		s.log.Error("UserService/Register error:", zap.Error(err))
		return err
	}
	return nil
}

func (s *UserService) ForgotPassword(ctx context.Context, req *dto.ForgotPasswordRequest) error {
	email, password, confirm := req.Email, req.Password, req.Confirm
	// 1.校验两次密码是否相同
	if password != confirm {
		return errors.New("passwords do not match")
	}

	// 2.判断该邮箱是否存在
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// 用户不存在
			s.log.Info("UserService/ForgotPassword info:", zap.Error(ErrUserNotFound))
			return ErrUserNotFound
		}
		// 其他数据库错误
		s.log.Error("UserService/ForgotPassword error:", zap.Error(err))
		return err
	}

	// 3.更新密码
	err = s.repo.UpdatePassword(ctx, user.ID, utils.Bcrypt(password))
	if err != nil {
		s.log.Error("UserService/ForgotPassword error:", zap.Error(err))
		return err
	}

	return nil
}

// =====================================================================================================================

func (s *UserService) List(ctx context.Context, req *dto.ListUsersRequest) (*dto.ListUsersResponse, error) {
	users, total, err := s.repo.List(ctx, req.Keyword, req.Page, (req.Page-1)*req.PageSize)
	if err != nil {
		s.log.Error("UserService/List error:", zap.Error(err))
		return nil, err
	}
	return dto.ToListUsersResponse(users, total), nil
}

func (s *UserService) GetProfile(ctx context.Context, req *dto.GetProfileRequest) (*dto.UserMsgResponse, error) {
	user, err := s.repo.GetByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.log.Info("UserService/GetProfile info:", zap.Error(ErrUserNotFound))
			return nil, ErrUserNotFound
		}
		s.log.Error("GetProfile error", zap.Error(err))
		return nil, err
	}
	return dto.ToUserMsgResponse(user), nil
}

func (s *UserService) UpdateProfile(ctx context.Context, req *dto.UpdateProfileRequest) (*dto.UserMsgResponse, error) {
	user, err := s.repo.GetByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		s.log.Error("UpdateProfile error", zap.Error(err))
		return nil, err
	}

	user.Name = req.Username
	user.Phone = req.Phone
	user.Sex = req.Sex
	user.Age = uint64(req.Age)
	if err := s.repo.UpdateProfile(ctx, user); err != nil {
		s.log.Error("UpdateProfile error", zap.Error(err))
		return nil, err
	}
	return dto.ToUserMsgResponse(user), nil
}

func (s *UserService) UpdateAvatar(ctx context.Context, req *dto.UpdateAvatarRequest) (*dto.UserAvatarResponse, error) {
	user, err := s.repo.GetByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		s.log.Error("UpdateAvatar error", zap.Error(err))
		return nil, err
	}
	if err := s.repo.UpdateAvatar(ctx, req.UserID, req.Avatar); err != nil {
		s.log.Error("UpdateAvatar error", zap.Error(err))
		return nil, err
	}
	user.Avatar = req.Avatar
	return dto.ToUserAvatarResponse(user.Avatar), nil
}

func (s *UserService) UpdateStatus(ctx context.Context, req *dto.UpdateUserStatusRequest) error {
	userID, status := req.UserID, req.Status
	if err := s.repo.UpdateStatus(ctx, userID, status); err != nil {
		s.log.Error("UserService/UpdateStatus error:", zap.Error(err))
		return err
	}
	return nil
}
