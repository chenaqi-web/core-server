package application

import (
	"context"
	"database/sql/driver"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"backend/core-server/internal/domain"
	authinfra "backend/core-server/internal/infras/auth"
	"backend/core-server/internal/infras/cache"
	"backend/core-server/internal/infras/repo"
	"backend/core-server/internal/model/entity"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type stubUserRepo struct {
	nameExists      bool
	emailExists     bool
	findByEmailUser *entity.User
	createErr       error
	createUser      *entity.User
	transactionErr  error
	transactionRuns int
	updateErr       error
	updatedUserID   uint64
	updatedHash     string
}

func (r *stubUserRepo) GetByID(context.Context, uint64) (*entity.User, error) { return nil, nil }
func (r *stubUserRepo) ListByIDs(context.Context, []uint64) ([]*entity.User, error) {
	return nil, nil
}
func (r *stubUserRepo) FindByID(context.Context, uint64) (*entity.User, error) { return nil, nil }
func (r *stubUserRepo) FindByName(context.Context, string) (*entity.User, error) {
	return nil, nil
}
func (r *stubUserRepo) FindByEmail(context.Context, string) (*entity.User, error) {
	return r.findByEmailUser, nil
}
func (r *stubUserRepo) ExistsByName(context.Context, string) (bool, error) {
	return r.nameExists, nil
}
func (r *stubUserRepo) ExistsByEmail(context.Context, string) (bool, error) {
	return r.emailExists, nil
}
func (r *stubUserRepo) Create(_ context.Context, user *entity.User) error {
	r.createUser = user
	if r.createErr == nil {
		user.ID = 42
	}
	return r.createErr
}
func (r *stubUserRepo) UpdatePasswordAndIncrementAuthVersion(_ context.Context, id uint64, hash string) error {
	r.updatedUserID = id
	r.updatedHash = hash
	return r.updateErr
}
func (r *stubUserRepo) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	r.transactionRuns++
	if r.transactionErr != nil {
		return r.transactionErr
	}
	return fn(ctx)
}

type stubEmailCodeStore struct {
	reservation   *domain.EmailCodeReservation
	reserveErr    error
	reserveCalls  int
	verifyValid   bool
	verifyErr     error
	verifyEmail   string
	verifyCode    string
	verifyPurpose domain.EmailCodePurpose
	cancelCalls   int
}

func (s *stubEmailCodeStore) ReserveCode(
	context.Context,
	string,
	domain.EmailCodePurpose,
) (*domain.EmailCodeReservation, error) {
	s.reserveCalls++
	if s.reserveErr != nil {
		return nil, s.reserveErr
	}
	if s.reservation == nil {
		s.reservation = &domain.EmailCodeReservation{Code: "123456", ID: "reservation", HourBucket: "2030010110"}
	}
	return s.reservation, nil
}

func TestAuthServiceSendEmailCodeRejectsInvalidInputBeforeReserve(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		purpose  domain.EmailCodePurpose
		expected error
	}{
		{name: "email", email: "invalid", purpose: domain.EmailCodePurposeRegister, expected: ErrInvalidEmail},
		{name: "purpose", email: "user@qq.com", purpose: "unknown", expected: ErrInvalidEmailPurpose},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			codes := &stubEmailCodeStore{}
			mailSender := &stubMailSender{}
			service, err := NewAuthService(nil, &stubUserRepo{}, codes, &stubSessionStore{}, mailSender)
			if err != nil {
				t.Fatalf("NewAuthService() error = %v", err)
			}
			err = service.SendEmailCode(context.Background(), test.email, test.purpose)
			if !errors.Is(err, test.expected) || codes.reserveCalls != 0 || mailSender.calls != 0 {
				t.Fatalf("SendEmailCode() error = %v, reserves = %d, mail calls = %d", err, codes.reserveCalls, mailSender.calls)
			}
		})
	}
}
func (s *stubEmailCodeStore) CancelCode(
	context.Context,
	string,
	domain.EmailCodePurpose,
	*domain.EmailCodeReservation,
) (bool, error) {
	s.cancelCalls++
	return true, nil
}
func (s *stubEmailCodeStore) VerifyCode(
	_ context.Context,
	email string,
	purpose domain.EmailCodePurpose,
	code string,
) (bool, error) {
	s.verifyEmail = email
	s.verifyPurpose = purpose
	s.verifyCode = code
	valid := s.verifyValid
	if valid && s.verifyErr == nil {
		s.verifyValid = false
	}
	return valid, s.verifyErr
}

type stubSessionStore struct {
	clearedUserID uint64
	err           error
}

func (s *stubSessionStore) ClearUserSession(_ context.Context, userID uint64) error {
	s.clearedUserID = userID
	return s.err
}

type stubMailSender struct {
	err       error
	calls     int
	recipient string
	code      string
	purpose   domain.EmailCodePurpose
}

func (s *stubMailSender) SendVerificationCode(
	_ context.Context,
	recipient, code string,
	purpose domain.EmailCodePurpose,
) error {
	s.calls++
	s.recipient = recipient
	s.code = code
	s.purpose = purpose
	return s.err
}

func TestAuthServiceSendEmailCodeCompensatesFailedDelivery(t *testing.T) {
	server := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })
	emailCodes, err := cache.NewEmailCodeStore(&cache.CacheClient{Cache: redisClient})
	if err != nil {
		t.Fatalf("NewEmailCodeStore() error = %v", err)
	}
	mailSender := &stubMailSender{err: errors.New("smtp secret detail")}
	service, err := NewAuthService(nil, &stubUserRepo{}, emailCodes, &stubSessionStore{}, mailSender)
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}

	err = service.SendEmailCode(context.Background(), "User@QQ.com", domain.EmailCodePurposeRegister)
	if !errors.Is(err, ErrMailUnavailable) {
		t.Fatalf("SendEmailCode() error = %v", err)
	}
	if strings.Contains(err.Error(), "smtp secret detail") {
		t.Fatalf("SMTP error leaked: %v", err)
	}

	mailSender.err = nil
	if err := service.SendEmailCode(context.Background(), "user@qq.com", domain.EmailCodePurposeRegister); err != nil {
		t.Fatalf("immediate retry after SMTP compensation error = %v", err)
	}
	if mailSender.calls != 2 || mailSender.recipient != "user@qq.com" {
		t.Fatalf("mail calls = %d, recipient = %q", mailSender.calls, mailSender.recipient)
	}
}

func TestAuthServiceSendEmailCodeMapsRateLimits(t *testing.T) {
	tests := []struct {
		name     string
		storeErr error
		expected error
	}{
		{name: "cooldown", storeErr: cache.ErrEmailCodeCooldown, expected: ErrEmailCodeCooldown},
		{name: "hourly", storeErr: cache.ErrEmailCodeHourlyLimit, expected: ErrEmailCodeHourlyLimit},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service, err := NewAuthService(
				nil,
				&stubUserRepo{},
				&stubEmailCodeStore{reserveErr: test.storeErr},
				&stubSessionStore{},
				&stubMailSender{},
			)
			if err != nil {
				t.Fatalf("NewAuthService() error = %v", err)
			}
			if err := service.SendEmailCode(context.Background(), "user@qq.com", domain.EmailCodePurposeRegister); !errors.Is(err, test.expected) {
				t.Fatalf("SendEmailCode() error = %v, want %v", err, test.expected)
			}
		})
	}
}

func TestAuthServiceRegisterForcesServerOwnedFields(t *testing.T) {
	users := &stubUserRepo{}
	codes := &stubEmailCodeStore{verifyValid: true}
	service, err := NewAuthService(nil, users, codes, &stubSessionStore{}, &stubMailSender{})
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}
	result, err := service.Register(context.Background(), RegisterInput{
		Username:        "  test_user  ",
		Email:           "Example@QQ.com ",
		Password:        "abc12345",
		ConfirmPassword: "abc12345",
		EmailCode:       "123456",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	created := users.createUser
	if created == nil || created.Role != entity.UserRoleUser || created.Status != entity.UserStatusActive || created.AuthVersion != 1 {
		t.Fatalf("server-owned fields = %+v", created)
	}
	if created.Phone != "" || created.Avatar != "" || created.Sex != defaultUserSex || created.Age != 0 {
		t.Fatalf("profile defaults = %+v", created)
	}
	if err := authinfra.VerifyPassword(created.Password, "abc12345"); err != nil {
		t.Fatalf("stored password is not a bcrypt hash: %v", err)
	}
	if result.Username != "test_user" || result.Email != "example@qq.com" || result.Role != entity.UserRoleUser {
		t.Fatalf("Register() result = %+v", result)
	}
	if codes.verifyPurpose != domain.EmailCodePurposeRegister || codes.verifyEmail != "example@qq.com" {
		t.Fatalf("verification = %q, %q", codes.verifyPurpose, codes.verifyEmail)
	}
}

func TestAuthServiceRegisterHandlesPrecheckAndDatabaseConflicts(t *testing.T) {
	tests := []struct {
		name     string
		users    *stubUserRepo
		expected error
	}{
		{name: "username precheck", users: &stubUserRepo{nameExists: true}, expected: ErrUsernameExists},
		{name: "email precheck", users: &stubUserRepo{emailExists: true}, expected: ErrEmailExists},
		{name: "username insert race", users: &stubUserRepo{createErr: repo.ErrUserNameExists}, expected: ErrUsernameExists},
		{name: "email insert race", users: &stubUserRepo{createErr: repo.ErrUserEmailExists}, expected: ErrEmailExists},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service, err := NewAuthService(nil, test.users, &stubEmailCodeStore{verifyValid: true}, &stubSessionStore{}, &stubMailSender{})
			if err != nil {
				t.Fatalf("NewAuthService() error = %v", err)
			}
			_, err = service.Register(context.Background(), RegisterInput{
				Username: "test_user", Email: "user@qq.com", Password: "abc12345",
				ConfirmPassword: "abc12345", EmailCode: "123456",
			})
			if !errors.Is(err, test.expected) {
				t.Fatalf("Register() error = %v, want %v", err, test.expected)
			}
		})
	}
}

func TestAuthServiceRegisterDatabaseFailureKeepsCodeConsumed(t *testing.T) {
	users := &stubUserRepo{createErr: errors.New("database unavailable")}
	codes := &stubEmailCodeStore{verifyValid: true}
	service, err := NewAuthService(nil, users, codes, &stubSessionStore{}, &stubMailSender{})
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}
	_, err = service.Register(context.Background(), RegisterInput{
		Username: "test_user", Email: "user@qq.com", Password: "abc12345",
		ConfirmPassword: "abc12345", EmailCode: "123456",
	})
	if err == nil {
		t.Fatal("Register() expected a database error")
	}
	valid, verifyErr := codes.VerifyCode(context.Background(), "user@qq.com", domain.EmailCodePurposeRegister, "123456")
	if verifyErr != nil || valid {
		t.Fatalf("database failure restored consumed code: valid = %v, error = %v", valid, verifyErr)
	}
}

func TestAuthServiceRegisterRejectsInvalidCode(t *testing.T) {
	users := &stubUserRepo{}
	service, err := NewAuthService(nil, users, &stubEmailCodeStore{verifyValid: false}, &stubSessionStore{}, &stubMailSender{})
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}
	_, err = service.Register(context.Background(), RegisterInput{
		Username: "test_user", Email: "user@qq.com", Password: "abc12345",
		ConfirmPassword: "abc12345", EmailCode: "123456",
	})
	if !errors.Is(err, ErrEmailCodeInvalid) || users.createUser != nil {
		t.Fatalf("Register() error = %v, created = %+v", err, users.createUser)
	}
}

func TestAuthServiceRegisterValidatesPasswordConfirmation(t *testing.T) {
	users := &stubUserRepo{}
	codes := &stubEmailCodeStore{verifyValid: true}
	service, err := NewAuthService(nil, users, codes, &stubSessionStore{}, &stubMailSender{})
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}
	_, err = service.Register(context.Background(), RegisterInput{
		Username: "test_user", Email: "user@qq.com", Password: "abc12345",
		ConfirmPassword: "different123", EmailCode: "123456",
	})
	if !errors.Is(err, ErrPasswordRule) || codes.verifyEmail != "" || users.createUser != nil {
		t.Fatalf("Register() error = %v, verified email = %q, created = %+v", err, codes.verifyEmail, users.createUser)
	}
}

type bcryptHashArgument struct {
	password string
}

func (argument bcryptHashArgument) Match(value driver.Value) bool {
	hash, ok := value.(string)
	return ok && authinfra.VerifyPassword(hash, argument.password) == nil
}

func TestAuthServiceResetPasswordSucceedsAfterCommitWhenSessionCleanupFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	users := repo.NewUserRepo(&repo.DBClient{DB: sqlx.NewDb(db, "sqlmock")})
	createdAt := time.Now()
	mock.ExpectQuery(`(?s)SELECT .*FROM user.*WHERE email = \? AND deleted_at IS NULL.*LIMIT 1`).
		WithArgs("user@qq.com").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "deleted_at", "name", "password", "phone", "avatar", "email", "role", "status", "auth_version", "sex", "age",
		}).AddRow(uint64(7), createdAt, createdAt, nil, "user", "old-hash", "", "", "user@qq.com", "user", "active", uint64(3), defaultUserSex, uint64(0)))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`
UPDATE user
SET password = ?, auth_version = auth_version + 1, updated_at = ?
WHERE id = ? AND deleted_at IS NULL`)).
		WithArgs(bcryptHashArgument{password: "newabc123"}, sqlmock.AnyArg(), uint64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	codes := &stubEmailCodeStore{verifyValid: true}
	sessions := &stubSessionStore{err: errors.New("redis internal secret")}
	service, err := NewAuthService(nil, users, codes, sessions, &stubMailSender{})
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}
	if err := service.ResetPasswordByEmail(context.Background(), ResetPasswordByEmailInput{
		Email: "User@QQ.com", EmailCode: "654321", NewPassword: "newabc123", ConfirmPassword: "newabc123",
	}); err != nil {
		t.Fatalf("ResetPasswordByEmail() error = %v", err)
	}
	if codes.verifyPurpose != domain.EmailCodePurposeResetPassword || codes.verifyCode != "654321" {
		t.Fatalf("verification = %q, %q", codes.verifyPurpose, codes.verifyCode)
	}
	if sessions.clearedUserID != 7 {
		t.Fatalf("cleared user ID = %d", sessions.clearedUserID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAuthServiceResetPasswordDoesNotStartTransactionForInvalidCode(t *testing.T) {
	users := &stubUserRepo{findByEmailUser: &entity.User{ID: 7}}
	service, err := NewAuthService(nil, users, &stubEmailCodeStore{verifyValid: false}, &stubSessionStore{}, &stubMailSender{})
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}
	err = service.ResetPasswordByEmail(context.Background(), ResetPasswordByEmailInput{
		Email: "user@qq.com", EmailCode: "123456", NewPassword: "newabc123", ConfirmPassword: "newabc123",
	})
	if !errors.Is(err, ErrEmailCodeInvalid) || users.transactionRuns != 0 {
		t.Fatalf("ResetPasswordByEmail() error = %v, transactions = %d", err, users.transactionRuns)
	}
}

func TestAuthServiceResetPasswordRollsBackAndDoesNotClearSessionOnUpdateFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	users := repo.NewUserRepo(&repo.DBClient{DB: sqlx.NewDb(db, "sqlmock")})
	createdAt := time.Now()
	mock.ExpectQuery(`(?s)SELECT .*FROM user.*WHERE email = \? AND deleted_at IS NULL.*LIMIT 1`).
		WithArgs("user@qq.com").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "deleted_at", "name", "password", "phone", "avatar", "email", "role", "status", "auth_version", "sex", "age",
		}).AddRow(uint64(7), createdAt, createdAt, nil, "user", "old-hash", "", "", "user@qq.com", "user", "active", uint64(3), defaultUserSex, uint64(0)))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`
UPDATE user
SET password = ?, auth_version = auth_version + 1, updated_at = ?
WHERE id = ? AND deleted_at IS NULL`)).
		WithArgs(bcryptHashArgument{password: "newabc123"}, sqlmock.AnyArg(), uint64(7)).
		WillReturnError(errors.New("update failed"))
	mock.ExpectRollback()

	sessions := &stubSessionStore{}
	codes := &stubEmailCodeStore{verifyValid: true}
	service, err := NewAuthService(nil, users, codes, sessions, &stubMailSender{})
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}
	err = service.ResetPasswordByEmail(context.Background(), ResetPasswordByEmailInput{
		Email: "user@qq.com", EmailCode: "654321", NewPassword: "newabc123", ConfirmPassword: "newabc123",
	})
	if err == nil || sessions.clearedUserID != 0 {
		t.Fatalf("ResetPasswordByEmail() error = %v, cleared user ID = %d", err, sessions.clearedUserID)
	}
	valid, verifyErr := codes.VerifyCode(context.Background(), "user@qq.com", domain.EmailCodePurposeResetPassword, "654321")
	if verifyErr != nil || valid {
		t.Fatalf("database failure restored consumed code: valid = %v, error = %v", valid, verifyErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
