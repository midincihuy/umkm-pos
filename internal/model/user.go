package model

import (
	"time"

	"github.com/google/uuid"
)

// User memetakan tabel public.users.
// Data diisi otomatis via Supabase trigger saat user register.
// ID selalu sama dengan auth.users.id dari Supabase Auth.
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"type:varchar(100);not null;default:''"`
	Email     string    `gorm:"type:varchar(150);not null;uniqueIndex"`
	Currency  string    `gorm:"type:varchar(5);not null;default:'IDR'"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
