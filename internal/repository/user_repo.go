package repository

import (
	"github.com/google/uuid"
	"umkm-pos/internal/model"
	"gorm.io/gorm"
)

// ─── UserRepo ────────────────────────────────────────────────────────────────

type UserRepo struct{ db *gorm.DB }

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) FindByEmail(email string) (*model.User, error) {
	var u model.User
	return &u, r.db.Where("email = ?", email).First(&u).Error
}
func (r *UserRepo) FindByID(id uuid.UUID) (*model.User, error) {
	var u model.User
	return &u, r.db.First(&u, "id = ?", id).Error
}
func (r *UserRepo) Create(u *model.User) error { return r.db.Create(u).Error }
func (r *UserRepo) Update(u *model.User) error { return r.db.Save(u).Error }
