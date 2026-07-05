package entity

import (
	"backend/core-server/internal/model/enum"
	"time"
)

type InteractionCountEntity struct {
	ID              string               `json:"id"`
	CreatedAt       time.Time            `json:"-"`
	UpdatedAt       time.Time            `json:"-"`
	ObjectType      enum.ObjectType      `json:"object_type"`
	ObjectID        string               `json:"object_id"`
	InteractionType enum.InteractionType `json:"interaction_type"`
	Count           int64                `json:"count"`
}
