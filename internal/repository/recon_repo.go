package repository

import (
	"github.com/google/uuid"
	"umkm-pos/internal/model"
	"gorm.io/gorm"
)

// ─── ReconRepo ────────────────────────────────────────────────────────────────

type ReconRepo struct{ db *gorm.DB }

func NewReconRepo(db *gorm.DB) *ReconRepo { return &ReconRepo{db: db} }

func (r *ReconRepo) CreateTx(tx *gorm.DB, ca *model.CashAdjustment) error { return tx.Create(ca).Error }

func (r *ReconRepo) FindByUserID(userID uuid.UUID, accountID *uuid.UUID) ([]model.CashAdjustment, error) {
	var list []model.CashAdjustment
	q := r.db.Preload("Account").Where("user_id = ?", userID)
	if accountID != nil { q = q.Where("account_id = ?", accountID) }
	return list, q.Order("created_at DESC").Find(&list).Error
}
