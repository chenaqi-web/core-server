package entity

import (
	"backend/core-server/internal/model/enum"
	"time"
)

type InteractionCount struct {
	ID              string               `gorm:"primaryKey;size:64" json:"id"`
	CreatedAt       time.Time            `gorm:"comment:创建时间" json:"-"`
	UpdatedAt       time.Time            `gorm:"comment:更新时间" json:"-"`
	ObjectType      enum.ObjectType      `gorm:"size:32;not null;uniqueIndex:uk_count_object,priority:1" json:"object_type"`
	ObjectID        string               `gorm:"size:64;not null;uniqueIndex:uk_count_object,priority:2" json:"object_id"`
	InteractionType enum.InteractionType `gorm:"size:32;not null;uniqueIndex:uk_count_object,priority:3" json:"interaction_type"`
	Count           int64                `gorm:"not null;default:0" json:"count"`
}

func (InteractionCount) TableName() string {
	return "interaction_count"
}
