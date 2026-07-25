package entity

import (
	"database/sql"
	"time"
)

const RootCategoryParentID uint64 = 0

// Category 二级分类节点（固定两级）。
// parent_id = 0 为一级分类（如：动作、喜剧），parent_id > 0 为二级分类（如：全部、爱情喜剧）。
type Category struct {
	ID        uint64       `db:"id" json:"id"`
	CreatedAt time.Time    `db:"created_at" json:"-"`
	UpdatedAt time.Time    `db:"updated_at" json:"-"`
	DeletedAt sql.NullTime `db:"deleted_at" json:"-"`
	ParentID  uint64       `db:"parent_id" json:"parent_id"`
	Name      string       `db:"name" json:"name"`
}

func (Category) TableName() string {
	return "category"
}

func (c *Category) IsRoot() bool {
	return c.ParentID == RootCategoryParentID
}
