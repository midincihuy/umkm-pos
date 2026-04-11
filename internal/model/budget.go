package model

import (
	"time"

	"github.com/google/uuid"
)

type FrequencyType string

const (
	FreqMonthly   FrequencyType = "monthly"
	FreqQuarterly FrequencyType = "quarterly"
	FreqYearly    FrequencyType = "yearly"
)

type Budget struct {
	ID            uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID        uuid.UUID     `gorm:"type:uuid;not null;index"`
	CategoryID    uuid.UUID     `gorm:"type:uuid;not null"`
	Month         int           `gorm:"not null"`
	Year          int           `gorm:"not null"`
	PlannedAmount float64       `gorm:"type:numeric(15,2);not null"`
	Frequency     FrequencyType `gorm:"type:frequency_type;not null;default:'monthly'"`
	CreatedAt     time.Time
	UpdatedAt     time.Time

	Category Category `gorm:"foreignKey:CategoryID"`
}

// MonthlyEquiv menghitung nilai bulanan efektif berdasarkan frekuensi
func (b *Budget) MonthlyEquiv() float64 {
	switch b.Frequency {
	case FreqYearly:
		return b.PlannedAmount / 12
	case FreqQuarterly:
		return b.PlannedAmount / 3
	default:
		return b.PlannedAmount
	}
}
