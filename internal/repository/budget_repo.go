package repository

import (
	"github.com/google/uuid"
	"umkm-pos/internal/model"
	"gorm.io/gorm"
)

// ─── BudgetRepo ───────────────────────────────────────────────────────────────

type BudgetRepo struct{ db *gorm.DB }

func NewBudgetRepo(db *gorm.DB) *BudgetRepo { return &BudgetRepo{db: db} }

func (r *BudgetRepo) FindByMonth(userID uuid.UUID, month, year int) ([]model.Budget, error) {
	var list []model.Budget
	return list, r.db.Preload("Category").
		Where("user_id = ? AND month = ? AND year = ?", userID, month, year).
		Find(&list).Error
}

func (r *BudgetRepo) FindByID(id, userID uuid.UUID) (*model.Budget, error) {
	var b model.Budget
	return &b, r.db.Preload("Category").
		Where("id = ? AND user_id = ?", id, userID).First(&b).Error
}

func (r *BudgetRepo) Upsert(b *model.Budget) error {
	// Cek apakah sudah ada
	var existing model.Budget
	err := r.db.Where("user_id = ? AND category_id = ? AND month = ? AND year = ?",
		b.UserID, b.CategoryID, b.Month, b.Year).First(&existing).Error

	if err == nil {
		// Update existing
		existing.PlannedAmount = b.PlannedAmount
		existing.Frequency = b.Frequency
		*b = existing
		return r.db.Save(&existing).Error
	}

	// Create new
	return r.db.Create(b).Error
}

func (r *BudgetRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Budget{}, "id = ?", id).Error
}
