package models

import (
	"time"

	"gorm.io/gorm"
)

// Transaction represents a monetary movement inside the platform.
type Transaction struct {
	TransactionID string          `gorm:"column:transaction_id;type:uuid;primaryKey"`
	Type          TransactionType `gorm:"column:type"`
	Amount        float64         `gorm:"column:amount"`
	Currency      Currency        `gorm:"column:currency"`
	Note          string          `gorm:"column:note"`
	Day           *int            `gorm:"column:day"`
	Month         Month           `gorm:"column:month"`
	Year          int             `gorm:"column:year"`
	ExchangeRate  *float64        `gorm:"column:exchange_rate"`
	UserID        string          `gorm:"column:user_id"`
	CategoryID    *string         `gorm:"column:category_id"`
	GoalID        *string         `gorm:"column:goal_id"`
	CreatedAt     time.Time       `gorm:"column:created_at"`
	UpdatedAt     time.Time       `gorm:"column:updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"column:deleted_at"`

	Category *Category `gorm:"foreignKey:CategoryID"`
	User     *User     `gorm:"foreignKey:UserID"`
}

func (Transaction) TableName() string {
	return "transactions"
}
