package repository

import (
	"github.com/google/uuid"
	"umkm-pos/internal/model"
	"gorm.io/gorm"
)

// ─── CategoryRepo ─────────────────────────────────────────────────────────────

type CategoryRepo struct{ db *gorm.DB }

func NewCategoryRepo(db *gorm.DB) *CategoryRepo { return &CategoryRepo{db: db} }

func (r *CategoryRepo) FindAll(userID uuid.UUID, includeInactive bool) ([]model.Category, error) {
	var cats []model.Category
	q := r.db.Where("user_id = ? OR user_id IS NULL", userID)
	if !includeInactive {
		q = q.Where("is_active = true")
	}
	return cats, q.Order("sort_order ASC, name ASC").Find(&cats).Error
}
func (r *CategoryRepo) FindByID(id uuid.UUID) (*model.Category, error) {
	var c model.Category
	return &c, r.db.First(&c, "id = ?", id).Error
}
func (r *CategoryRepo) Create(c *model.Category) error { return r.db.Create(c).Error }
func (r *CategoryRepo) Update(c *model.Category) error { return r.db.Save(c).Error }
func (r *CategoryRepo) IsUsed(id uuid.UUID) (bool, error) {
	var n int64
	return n > 0, r.db.Model(&model.Transaction{}).Where("category_id = ?", id).Count(&n).Error
}
