package application

import (
	"context"
	"core-server/internal/domain"
	"core-server/internal/infras/clog"
	"core-server/internal/model/entity"
	"core-server/internal/utils"
	"database/sql"
	"errors"

	"go.uber.org/zap"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserDisabled       = errors.New("user is disabled")
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
	if user.Status != "active" {
		return nil, ErrUserDisabled
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

	if user.Status != "active" {
		return nil, ErrUserDisabled
	}
	return user, nil
}

func (s *UserService) Register(ctx context.Context, username, email, password, confirm string) (*entity.User, error) {
	if password != confirm {
		return nil, errors.New("passwords do not match")
	}

	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailAlreadyInUse
	}

	user := &entity.User{
		Name:        username,
		Email:       email,
		Password:    utils.Bcrypt(password),
		Role:        entity.UserRoleUser,
		Status:      "active",
		AuthVersion: 1,
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

	if user.Status != "active" {
		return ErrUserDisabled
	}
	return nil
}
