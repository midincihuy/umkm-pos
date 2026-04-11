package model

import (
	"time"

	"github.com/google/uuid"
)

type Transfer struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index"`
	FromAccountID uuid.UUID `gorm:"type:uuid;not null"`
	ToAccountID   uuid.UUID `gorm:"type:uuid;not null"`
	Amount        float64   `gorm:"type:numeric(15,2);not null"`
	Fee           float64   `gorm:"type:numeric(15,2);not null;default:0"`
	Date          time.Time `gorm:"type:date;not null"`
	Notes         string    `gorm:"type:text"`
	CreatedAt     time.Time

	FromAccount Account `gorm:"foreignKey:FromAccountID"`
	ToAccount   Account `gorm:"foreignKey:ToAccountID"`
}
