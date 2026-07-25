package entity

import (
	"time"

	"gorm.io/gorm"
)

// CategoryType 一级类型，如：动作、喜剧、爱情。
type CategoryType struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	CreatedAt time.Time      `gorm:"comment:创建时间" json:"-"`
	UpdatedAt time.Time      `gorm:"comment:更新时间" json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index;comment:软删除时间" json:"-"`
	Name      string         `gorm:"size:64;not null;uniqueIndex:uk_category_type_name;comment:一级类型名称" json:"name"`
}

func (CategoryType) TableName() string {
	return "category_type"
}
