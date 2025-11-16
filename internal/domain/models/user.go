package models

import (
	"time"

	"gorm.io/gorm"
)

// User mirrors the users table present in the original project.
type User struct {
	UserID       string         `gorm:"column:user_id;type:uuid;primaryKey"`
	Email        string         `gorm:"column:email;uniqueIndex"`
	FirstName    string         `gorm:"column:first_name"`
	LastName     string         `gorm:"column:last_name"`
	Password     string         `gorm:"column:password"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at"`
	Categories   []Category     `gorm:"foreignKey:UserID"`
	Transactions []Transaction  `gorm:"foreignKey:UserID"`
}

// TableName tells GORM the exact table name, avoiding pluralization issues.
func (User) TableName() string {
	return "users"
}
