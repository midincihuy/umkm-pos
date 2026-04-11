package model

import (
	"time"

	"github.com/google/uuid"
)

type CategoryType string

const (
	CategoryIncome  CategoryType = "income"
	CategoryExpense CategoryType = "expense"
)

type Category struct {
	ID        uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    *uuid.UUID   `gorm:"type:uuid;index"` // NULL = kategori sistem
	Name      string       `gorm:"type:varchar(100);not null"`
	Type      CategoryType `gorm:"type:category_type;not null"`
	Icon      string       `gorm:"type:varchar(50)"`
	Color     string       `gorm:"type:varchar(10)"`
	IsSystem  bool         `gorm:"not null;default:false"`
	IsActive  bool         `gorm:"not null;default:true"`
	IsOthers  bool         `gorm:"not null;default:false"` // flag untuk kategori lain-lain
	SortOrder int          `gorm:"default:0"`
	CreatedAt time.Time
}
