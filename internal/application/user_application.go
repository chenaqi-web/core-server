package application

import (
	"context"
	"core-server/internal/domain"
	"core-server/internal/infras/clog"
	"core-server/internal/model/entity"
	"core-server/internal/utils"
	"errors"
	"strings"

	"go.uber.org/zap"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserDisabled       = errors.New("user is disabled")
	ErrEmailAlreadyInUse  = errors.New("email is already registered")
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
	user, err := s.repo.GetByName(ctx, strings.TrimSpace(username))
	if err != nil {
		s.log.Error("Login error :", zap.Error(err))
		return nil, err
	}

	if user.Status != "active" {
		return nil, ErrUserDisabled
	}

	if user.Password != utils.Bcrypt(password) {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func (s *UserService) EmailLogin(ctx context.Context, email string) (*entity.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		s.log.Error("EmailLogin error :", zap.Error(err))
		return nil, err
	}

	if user.Status != "active" {
		return nil, ErrUserDisabled
	}
	return user, nil
}

func (s *UserService) Register(ctx context.Context, username, email, password string) (*entity.User, error) {
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
