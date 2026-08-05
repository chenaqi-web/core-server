package application

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"core-server/internal/domain"
	"core-server/internal/infras/cache"
	"core-server/internal/infras/mail"
	"core-server/internal/model/entity"

	"golang.org/x/crypto/bcrypt"
)

const (
	emailCodeExpire   = 5 * time.Minute
	emailSendInterval = time.Minute
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserDisabled       = errors.New("user is disabled")
	ErrEmailTooFrequent   = errors.New("email code sent too frequently")
)

type UserService struct {
	repo  domain.UserRepoDomain
	cache *cache.CacheClient
	mail  *mail.Sender
}

func NewUserService(
	repo domain.UserRepoDomain,
	cacheClient *cache.CacheClient,
	mailSender *mail.Sender,
) *UserService {
	return &UserService{
		repo:  repo,
		cache: cacheClient,
		mail:  mailSender,
	}
}

// 校验登录
func (s *UserService) Login(ctx context.Context, username, password string) (*entity.User, error) {
	user, err := s.repo.GetByName(ctx, strings.TrimSpace(username))
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	if user.Status != "" && user.Status != "active" {
		return nil, ErrUserDisabled
	}
	return user, nil
}

// 发送邮箱验证码
func (s *UserService) SendEmailCode(ctx context.Context, email string, purpose int32) error {
	purposeValue := strconv.FormatInt(int64(purpose), 10)
	limitKey := fmt.Sprintf("auth:email_code_limit:%s:%s", purposeValue, email)
	codeKey := fmt.Sprintf("auth:email_code:%s:%s", purposeValue, email)

	ok, err := s.cache.Cache.SetNX(ctx, limitKey, 1, emailSendInterval).Result()
	if err != nil {
		return err
	}
	if !ok {
		return ErrEmailTooFrequent
	}

	code, err := newEmailCode()
	if err != nil {
		s.cache.Cache.Del(ctx, limitKey)
		return err
	}
	if err := s.cache.Cache.Set(ctx, codeKey, code, emailCodeExpire).Err(); err != nil {
		s.cache.Cache.Del(ctx, limitKey)
		return err
	}
	if err := s.mail.SendCode(email, code); err != nil {
		s.cache.Cache.Del(ctx, codeKey, limitKey)
		return fmt.Errorf("send email code: %w", err)
	}
	return nil
}

func newEmailCode() (string, error) {
	number, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", number.Int64()), nil
}
