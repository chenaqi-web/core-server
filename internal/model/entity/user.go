package entity

import (
	"database/sql"
	"time"
)

type User struct {
	ID        uint64       `db:"id"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
	Name      string       `db:"name"`
	Password  string       `db:"password"`
	Phone     string       `db:"phone"`
	Avatar    string       `db:"avatar"`
	Email     string       `db:"email"`
	Role      string       `db:"role"`
	Sex       string       `db:"sex"`
	Status    string       `db:"status"`
	Birthday  time.Time    `db:"birthday"`
}

func (User) TableName() string {
	return "user"
}

//======================================================================================================================
// 用户角色

const (
	UserRoleAdmin = "admin"
	UserRoleUser  = "user"
)

const (
	StatusApproved = "approved" // 已通过
	StatusBlocked  = "blocked"  // 拉黑中
)
