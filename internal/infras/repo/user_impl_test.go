package repo

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"backend/core-server/internal/model/entity"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func newMockUserRepo(t *testing.T) (*UserRepo, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewUserRepo(&DBClient{DB: sqlx.NewDb(db, "sqlmock")}), mock
}

func TestUserRepoFindByNameUsesBoundArgument(t *testing.T) {
	repo, mock := newMockUserRepo(t)
	query := `
SELECT ` + authUserColumns + `
FROM user
WHERE name = ? AND deleted_at IS NULL
LIMIT 1`
	createdAt := time.Now()
	maliciousName := "name' OR 1=1 --"
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(maliciousName).
		WillReturnRows(userRows(createdAt).AddRow(
			uint64(1), createdAt, createdAt, nil, maliciousName, "hash", "", "", "user@qq.com", "user", "active", uint64(1), "未知", uint64(0),
		))

	user, err := repo.FindByName(context.Background(), maliciousName)
	if err != nil {
		t.Fatalf("FindByName() error = %v", err)
	}
	if user == nil || user.Name != maliciousName || user.Password != "hash" {
		t.Fatalf("FindByName() = %+v", user)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserRepoFindByIDExcludesSoftDeletedUsers(t *testing.T) {
	repo, mock := newMockUserRepo(t)
	query := `
SELECT ` + authUserColumns + `
FROM user
WHERE id = ? AND deleted_at IS NULL
LIMIT 1`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(uint64(8)).WillReturnError(sql.ErrNoRows)
	user, err := repo.FindByID(context.Background(), 8)
	if err != nil || user != nil {
		t.Fatalf("FindByID() = %+v, %v", user, err)
	}
}

func TestUserRepoFindByEmailNotFound(t *testing.T) {
	repo, mock := newMockUserRepo(t)
	query := `
SELECT ` + authUserColumns + `
FROM user
WHERE email = ? AND deleted_at IS NULL
LIMIT 1`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("missing@qq.com").WillReturnError(sql.ErrNoRows)
	user, err := repo.FindByEmail(context.Background(), "missing@qq.com")
	if err != nil || user != nil {
		t.Fatalf("FindByEmail() = %+v, %v", user, err)
	}
}

func TestUserRepoExistsQueriesUseBoundArguments(t *testing.T) {
	repo, mock := newMockUserRepo(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM user WHERE name = ? LIMIT 1)`)).
		WithArgs("user").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM user WHERE email = ? LIMIT 1)`)).
		WithArgs("user@qq.com").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	nameExists, err := repo.ExistsByName(context.Background(), "user")
	if err != nil || !nameExists {
		t.Fatalf("ExistsByName() = %v, %v", nameExists, err)
	}
	emailExists, err := repo.ExistsByEmail(context.Background(), "user@qq.com")
	if err != nil || emailExists {
		t.Fatalf("ExistsByEmail() = %v, %v", emailExists, err)
	}
}

func TestUserRepoCreateUsesBoundArguments(t *testing.T) {
	repo, mock := newMockUserRepo(t)
	query := `
INSERT INTO user
    (created_at, updated_at, name, password, phone, avatar, email, role, status, auth_version, sex, age)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	user := &entity.User{
		Name: "user", Password: "hash", Phone: "", Avatar: "", Email: "user@qq.com",
		Role: entity.UserRoleUser, Status: entity.UserStatusActive, AuthVersion: 1, Sex: "未知", Age: 0,
	}
	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), user.Name, user.Password, user.Phone, user.Avatar, user.Email, user.Role, user.Status, user.AuthVersion, user.Sex, user.Age).
		WillReturnResult(sqlmock.NewResult(15, 1))
	if err := repo.Create(context.Background(), user); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if user.ID != 15 || user.CreatedAt.IsZero() || user.UpdatedAt.IsZero() {
		t.Fatalf("created user = %+v", user)
	}
}

func TestUserRepoCreateMapsDuplicateEmail(t *testing.T) {
	repo, mock := newMockUserRepo(t)
	query := `
INSERT INTO user
    (created_at, updated_at, name, password, phone, avatar, email, role, status, auth_version, sex, age)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	user := &entity.User{Name: "user", Password: "hash", Email: "user@qq.com", Role: "user", Status: "active", AuthVersion: 1, Sex: "未知"}
	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), user.Name, user.Password, user.Phone, user.Avatar, user.Email, user.Role, user.Status, user.AuthVersion, user.Sex, user.Age).
		WillReturnError(&mysql.MySQLError{Number: 1062, Message: "Duplicate entry for key 'uk_user_email'"})
	if err := repo.Create(context.Background(), user); !errors.Is(err, ErrUserEmailExists) {
		t.Fatalf("Create() error = %v, want %v", err, ErrUserEmailExists)
	}
}

func TestUserRepoUpdatePasswordAndIncrementAuthVersion(t *testing.T) {
	repo, mock := newMockUserRepo(t)
	query := `
UPDATE user
SET password = ?, auth_version = auth_version + 1, updated_at = ?
WHERE id = ? AND deleted_at IS NULL`
	mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs("new-hash", sqlmock.AnyArg(), uint64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.UpdatePasswordAndIncrementAuthVersion(context.Background(), 5, "new-hash"); err != nil {
		t.Fatalf("UpdatePasswordAndIncrementAuthVersion() error = %v", err)
	}
}

func TestUserRepoUpdatePasswordReturnsNotFound(t *testing.T) {
	repo, mock := newMockUserRepo(t)
	query := `
UPDATE user
SET password = ?, auth_version = auth_version + 1, updated_at = ?
WHERE id = ? AND deleted_at IS NULL`
	mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs("new-hash", sqlmock.AnyArg(), uint64(99)).WillReturnResult(sqlmock.NewResult(0, 0))
	if err := repo.UpdatePasswordAndIncrementAuthVersion(context.Background(), 99, "new-hash"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("UpdatePasswordAndIncrementAuthVersion() error = %v, want %v", err, ErrNotFound)
	}
}

func userRows(_ time.Time) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "name", "password", "phone", "avatar", "email", "role", "status", "auth_version", "sex", "age",
	})
}
