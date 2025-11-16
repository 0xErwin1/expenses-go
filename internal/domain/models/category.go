package models

import (
	"time"

	"gorm.io/gorm"
)

// Category groups user transactions.
type Category struct {
	CategoryID string          `gorm:"column:category_id;type:uuid;primaryKey"`
	Type       TransactionType `gorm:"column:type"`
	Name       string          `gorm:"column:name"`
	Note       string          `gorm:"column:note"`
	UserID     string          `gorm:"column:user_id"`
	CreatedAt  time.Time       `gorm:"column:created_at"`
	UpdatedAt  time.Time       `gorm:"column:updated_at"`
	DeletedAt  gorm.DeletedAt  `gorm:"column:deleted_at"`
	User       *User           `gorm:"foreignKey:UserID"`
}

func (Category) TableName() string {
	return "categories"
}
