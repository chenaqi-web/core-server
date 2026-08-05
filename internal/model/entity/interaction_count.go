package entity

import (
	"core-server/internal/model/enum"
	"time"
)

type InteractionCount struct {
	ID              uint64               `db:"id" json:"id"`
	CreatedAt       time.Time            `db:"created_at" json:"-"`
	UpdatedAt       time.Time            `db:"updated_at" json:"-"`
	ObjectType      enum.ObjectType      `db:"object_type" json:"object_type"`
	ObjectID        uint64               `db:"object_id" json:"object_id"`
	InteractionType enum.InteractionType `db:"interaction_type" json:"interaction_type"`
	Count           int64                `db:"count" json:"count"`
}

func (InteractionCount) TableName() string {
	return "interaction_count"
}
