package application

import (
	"context"
	"core-server/internal/domain"
	"core-server/internal/infras/clog"
	"core-server/internal/model/entity"
	"core-server/internal/utils"
	"database/sql"
	"errors"
	"strings"
	"unicode/utf8"

	"go.uber.org/zap"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserDisabled       = errors.New("user is disabled")
	ErrUserBlocked        = errors.New("USER_BLOCKED")
	ErrEmailAlreadyInUse  = errors.New("email is already registered")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidUserProfile = errors.New("invalid user profile")
	ErrUsernameInUse      = errors.New("username is already in use")
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

func (s *UserService) Login(ctx context.Context, username, password string) (*entity.User, error) {
	// 1.判断用户是否存在
	user, err := s.repo.GetByName(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// 用户不存在
			s.log.Info("Login error:", zap.Error(ErrUserNotFound))
			return nil, ErrUserNotFound
		}
		// 数据库错误
		s.log.Error("Login error:", zap.Error(err))
		return nil, err
	}

	// 2.判断密码是否正确
	if user.Password != utils.Bcrypt(password) {
		return nil, ErrInvalidCredentials
	}

	// 状态是否是active
	if user.Status != entity.StatusApproved {
		return nil, userStatusError(user.Status)
	}
	return user, nil
}

func (s *UserService) EmailLogin(ctx context.Context, email string) (*entity.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// 用户不存在
			s.log.Info("Login error:", zap.Error(ErrUserNotFound))
			return nil, ErrUserNotFound
		}
		// 数据库错误
		s.log.Error("Login error:", zap.Error(err))
		return nil, err
	}

	if user.Status != entity.StatusApproved {
		return nil, userStatusError(user.Status)
	}
	return user, nil
}

func (s *UserService) Register(ctx context.Context, username, email, password string) (*entity.User, error) {
	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailAlreadyInUse
	}

	user := &entity.User{
		Name:     username,
		Email:    email,
		Password: utils.Bcrypt(password),
		Role:     entity.UserRoleUser,
		Status:   entity.StatusApproved,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) ForgotPassword(ctx context.Context, email, password, confirm string) error {
	if password != confirm {
		return errors.New("passwords do not match")
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// 用户不存在
			return ErrUserNotFound
		}
		// 其他数据库错误
		s.log.Error("Login error:", zap.Error(err))
		return err
	}

	if user.Status != entity.StatusApproved {
		return userStatusError(user.Status)
	}
	return s.repo.UpdatePassword(ctx, user.ID, utils.Bcrypt(password))
}

// =====================================================================================================================

func (s *UserService) List(ctx context.Context, keyword string, page, pageSize uint32) ([]*entity.User, uint64, error) {
	if page == 0 {
		page = 1
	}
	if pageSize == 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.List(ctx, keyword, pageSize, (page-1)*pageSize)
}

func (s *UserService) UpdateStatus(ctx context.Context, userID uint64, status string) error {
	if userID == 0 || !entity.IsValidUserStatus(status) {
		return ErrInvalidUserProfile
	}
	if err := s.repo.UpdateStatus(ctx, userID, status); err != nil {
		return err
	}
	return nil
}

func (s *UserService) GetProfile(ctx context.Context, userID uint64) (*entity.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		s.log.Error("GetProfile error", zap.Error(err))
		return nil, err
	}
	if user.Status != entity.StatusApproved {
		return nil, userStatusError(user.Status)
	}
	return user, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uint64, username, phone, sex string, age uint32) (*entity.User, error) {
	username = strings.TrimSpace(username)
	phone = strings.TrimSpace(phone)
	sex = strings.TrimSpace(sex)
	usernameLength := utf8.RuneCountInString(username)
	if userID == 0 || usernameLength < 2 || usernameLength > 50 || len(phone) > 20 || age > 150 {
		return nil, ErrInvalidUserProfile
	}
	if sex != "" && sex != "male" && sex != "female" {
		return nil, ErrInvalidUserProfile
	}

	user, err := s.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	if username != user.Name {
		existing, findErr := s.repo.GetByName(ctx, username)
		if findErr != nil && !errors.Is(findErr, sql.ErrNoRows) {
			return nil, findErr
		}
		if existing != nil && existing.ID != userID {
			return nil, ErrUsernameInUse
		}
	}

	user.Name = username
	user.Phone = phone
	user.Sex = sex
	user.Age = uint64(age)
	if err := s.repo.UpdateProfile(ctx, user); err != nil {
		s.log.Error("UpdateProfile error", zap.Error(err))
		return nil, err
	}
	return user, nil
}

func (s *UserService) UpdateAvatar(ctx context.Context, userID uint64, avatar string) (*entity.User, error) {
	avatar = strings.TrimSpace(avatar)
	if userID == 0 || avatar == "" || len(avatar) > 500 {
		return nil, ErrInvalidUserProfile
	}

	user, err := s.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdateAvatar(ctx, userID, avatar); err != nil {
		s.log.Error("UpdateAvatar error", zap.Error(err))
		return nil, err
	}
	user.Avatar = avatar
	return user, nil
}

func userStatusError(status string) error {
	if status == entity.StatusBlocked {
		return ErrUserBlocked
	}
	return ErrUserDisabled
}
