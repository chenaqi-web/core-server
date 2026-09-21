package entity

import (
	"core-server/internal/model/enum"
	"database/sql"
	"time"
)

type User struct {
	ID        uint64          `db:"id"`
	CreatedAt time.Time       `db:"created_at"`
	UpdatedAt time.Time       `db:"updated_at"`
	Birthday  time.Time       `db:"birthday"`
	DeletedAt sql.NullTime    `db:"deleted_at"`
	Name      string          `db:"name"`
	Password  string          `db:"password"`
	Phone     string          `db:"phone"`
	Avatar    string          `db:"avatar"`
	Email     string          `db:"email"`
	Signature string          `db:"signature"`
	Role      enum.UserRole   `db:"role"`
	Sex       enum.UserSex    `db:"sex"`
	Status    enum.UserStatus `db:"status"`
}

func (User) TableName() string {
	return "user"
}
