package application

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"

	"backend/core-server/internal/domain"
	authinfra "backend/core-server/internal/infras/auth"
	"backend/core-server/internal/infras/cache"
	"backend/core-server/internal/infras/clog"
	"backend/core-server/internal/infras/repo"
	"backend/core-server/internal/model/entity"

	"go.uber.org/zap"
)

const (
	maxUsernameCharacters = 64
	maxEmailCharacters    = 128
	maxPhoneCharacters    = 20
	maxAvatarCharacters   = 255
	maxSexCharacters      = 8
	defaultUserSex        = "\u672a\u77e5"
)

type AuthService struct {
	log        *clog.Log
	users      domain.UserRepoDomain
	emailCodes domain.EmailCodeStore
	sessions   domain.AuthSessionStore
	mailSender domain.MailSender
}

type RegisterInput struct {
	Username        string
	Email           string
	Password        string
	ConfirmPassword string
	EmailCode       string
	Phone           string
	Avatar          string
	Sex             string
	Age             uint32
}

type ResetPasswordByEmailInput struct {
	Email           string
	EmailCode       string
	NewPassword     string
	ConfirmPassword string
}

type AuthUser struct {
	ID          uint64
	Username    string
	Email       string
	Phone       string
	Avatar      string
	Sex         string
	Age         uint64
	Role        string
	Status      string
	AuthVersion uint64
}

func NewAuthService(
	log *clog.Log,
	users domain.UserRepoDomain,
	emailCodes domain.EmailCodeStore,
	sessions domain.AuthSessionStore,
	mailSender domain.MailSender,
) (*AuthService, error) {
	if users == nil || emailCodes == nil || sessions == nil || mailSender == nil {
		return nil, fmt.Errorf("auth service dependency is nil")
	}
	return &AuthService{
		log:        log,
		users:      users,
		emailCodes: emailCodes,
		sessions:   sessions,
		mailSender: mailSender,
	}, nil
}

func (s *AuthService) SendEmailCode(
	ctx context.Context,
	email string,
	purpose domain.EmailCodePurpose,
) error {
	normalizedEmail, err := validateAndNormalizeEmail(email)
	if err != nil {
		return err
	}
	if !purpose.Valid() {
		return ErrInvalidEmailPurpose
	}

	reservation, err := s.emailCodes.ReserveCode(ctx, normalizedEmail, purpose)
	if err != nil {
		return mapEmailCodeStoreError(err)
	}
	if err := s.mailSender.SendVerificationCode(ctx, normalizedEmail, reservation.Code, purpose); err != nil {
		if _, cancelErr := s.emailCodes.CancelCode(ctx, normalizedEmail, purpose, reservation); cancelErr != nil {
			return fmt.Errorf("%w: verification code compensation failed", ErrMailUnavailable)
		}
		return ErrMailUnavailable
	}
	return nil
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*AuthUser, error) {
	normalized, err := normalizeRegisterInput(input)
	if err != nil {
		return nil, err
	}

	nameExists, err := s.users.ExistsByName(ctx, normalized.Username)
	if err != nil {
		return nil, fmt.Errorf("check username uniqueness: %w", err)
	}
	if nameExists {
		return nil, ErrUsernameExists
	}
	emailExists, err := s.users.ExistsByEmail(ctx, normalized.Email)
	if err != nil {
		return nil, fmt.Errorf("check email uniqueness: %w", err)
	}
	if emailExists {
		return nil, ErrEmailExists
	}

	passwordHash, err := authinfra.HashPassword(normalized.Password)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPasswordRule, err)
	}
	validCode, err := s.emailCodes.VerifyCode(
		ctx,
		normalized.Email,
		domain.EmailCodePurposeRegister,
		normalized.EmailCode,
	)
	if err != nil {
		return nil, fmt.Errorf("verify registration email code: %w", err)
	}
	if !validCode {
		return nil, ErrEmailCodeInvalid
	}

	user := &entity.User{
		Name:        normalized.Username,
		Password:    passwordHash,
		Phone:       normalized.Phone,
		Avatar:      normalized.Avatar,
		Email:       normalized.Email,
		Role:        entity.UserRoleUser,
		Status:      entity.UserStatusActive,
		AuthVersion: 1,
		Sex:         normalized.Sex,
		Age:         uint64(normalized.Age),
	}
	if err := s.users.Create(ctx, user); err != nil {
		switch {
		case errors.Is(err, repo.ErrUserNameExists):
			return nil, ErrUsernameExists
		case errors.Is(err, repo.ErrUserEmailExists):
			return nil, ErrEmailExists
		default:
			return nil, fmt.Errorf("create user: %w", err)
		}
	}
	return authUserFromEntity(user), nil
}

func (s *AuthService) ResetPasswordByEmail(ctx context.Context, input ResetPasswordByEmailInput) error {
	email, err := validateAndNormalizeEmail(input.Email)
	if err != nil {
		return err
	}
	if !isSixDigitCode(input.EmailCode) {
		return ErrEmailCodeInvalid
	}
	if input.NewPassword != input.ConfirmPassword {
		return fmt.Errorf("%w: passwords do not match", ErrPasswordRule)
	}
	if err := authinfra.ValidatePassword(input.NewPassword); err != nil {
		return fmt.Errorf("%w: %v", ErrPasswordRule, err)
	}

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("find reset account: %w", err)
	}
	if user == nil {
		return ErrEmailCodeInvalid
	}
	passwordHash, err := authinfra.HashPassword(input.NewPassword)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPasswordRule, err)
	}
	validCode, err := s.emailCodes.VerifyCode(
		ctx,
		email,
		domain.EmailCodePurposeResetPassword,
		input.EmailCode,
	)
	if err != nil {
		return fmt.Errorf("verify reset email code: %w", err)
	}
	if !validCode {
		return ErrEmailCodeInvalid
	}

	if err := s.users.WithTransaction(ctx, func(txCtx context.Context) error {
		return s.users.UpdatePasswordAndIncrementAuthVersion(txCtx, user.ID, passwordHash)
	}); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrEmailCodeInvalid
		}
		return fmt.Errorf("reset password transaction: %w", err)
	}
	if err := s.sessions.ClearUserSession(ctx, user.ID); err != nil {
		s.log.Error(
			"failed to clear auth session after password reset",
			zap.Uint64("user_id", user.ID),
		)
	}
	return nil
}

func normalizeRegisterInput(input RegisterInput) (RegisterInput, error) {
	input.Username = strings.TrimSpace(input.Username)
	if input.Username == "" || utf8.RuneCountInString(input.Username) > maxUsernameCharacters || containsControl(input.Username) {
		return RegisterInput{}, fmt.Errorf("%w: username", ErrInvalidAuthInput)
	}
	email, err := validateAndNormalizeEmail(input.Email)
	if err != nil {
		return RegisterInput{}, err
	}
	input.Email = email
	if input.Password != input.ConfirmPassword {
		return RegisterInput{}, fmt.Errorf("%w: passwords do not match", ErrPasswordRule)
	}
	if err := authinfra.ValidatePassword(input.Password); err != nil {
		return RegisterInput{}, fmt.Errorf("%w: %v", ErrPasswordRule, err)
	}
	if !isSixDigitCode(input.EmailCode) {
		return RegisterInput{}, ErrEmailCodeInvalid
	}

	input.Phone = strings.TrimSpace(input.Phone)
	input.Avatar = strings.TrimSpace(input.Avatar)
	input.Sex = strings.TrimSpace(input.Sex)
	if input.Sex == "" {
		input.Sex = defaultUserSex
	}
	if utf8.RuneCountInString(input.Phone) > maxPhoneCharacters || containsControl(input.Phone) ||
		utf8.RuneCountInString(input.Avatar) > maxAvatarCharacters || containsControl(input.Avatar) ||
		utf8.RuneCountInString(input.Sex) > maxSexCharacters || containsControl(input.Sex) {
		return RegisterInput{}, fmt.Errorf("%w: optional profile field", ErrInvalidAuthInput)
	}
	return input, nil
}

func validateAndNormalizeEmail(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || utf8.RuneCountInString(value) > maxEmailCharacters || containsControl(value) {
		return "", ErrInvalidEmail
	}
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value || !strings.Contains(value, "@") {
		return "", ErrInvalidEmail
	}
	return value, nil
}

func containsControl(value string) bool {
	for _, character := range value {
		if unicode.IsControl(character) {
			return true
		}
	}
	return false
}

func isSixDigitCode(code string) bool {
	if len(code) != 6 {
		return false
	}
	for _, character := range code {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

func mapEmailCodeStoreError(err error) error {
	switch {
	case errors.Is(err, cache.ErrInvalidEmailCodePurpose):
		return ErrInvalidEmailPurpose
	case errors.Is(err, cache.ErrEmailCodeCooldown):
		return ErrEmailCodeCooldown
	case errors.Is(err, cache.ErrEmailCodeHourlyLimit):
		return ErrEmailCodeHourlyLimit
	default:
		return fmt.Errorf("reserve email code: %w", err)
	}
}

func authUserFromEntity(user *entity.User) *AuthUser {
	return &AuthUser{
		ID:          user.ID,
		Username:    user.Name,
		Email:       user.Email,
		Phone:       user.Phone,
		Avatar:      user.Avatar,
		Sex:         user.Sex,
		Age:         user.Age,
		Role:        user.Role,
		Status:      user.Status,
		AuthVersion: user.AuthVersion,
	}
}
