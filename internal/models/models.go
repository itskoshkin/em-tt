package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Subscription struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	ServiceName string         `json:"service_name"`
	Price       int            `json:"price"`
	UserID      uuid.UUID      `json:"user_id"`
	StartDate   time.Time      `json:"start_date"`
	EndDate     *time.Time     `json:"end_date,omitempty"`
	CreatedAt   time.Time      `json:"-" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"-" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}
