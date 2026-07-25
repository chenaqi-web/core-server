package entity

import (
	"time"

	"gorm.io/gorm"
)

// Category 二级分类，挂在某个 CategoryType 下，如：全部、爱情喜剧。
type Category struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	CreatedAt time.Time      `gorm:"comment:创建时间" json:"-"`
	UpdatedAt time.Time      `gorm:"comment:更新时间" json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index;comment:软删除时间" json:"-"`
	TypeID    uint64         `gorm:"not null;uniqueIndex:uk_category_type_id_name,priority:1;index:idx_category_type_id;comment:一级类型ID" json:"type_id"`
	Name      string         `gorm:"size:64;not null;uniqueIndex:uk_category_type_id_name,priority:2;comment:二级分类名称" json:"name"`
}

func (Category) TableName() string {
	return "category"
}
