package repository

import (
	"github.com/google/uuid"
	"umkm-pos/internal/model"
	"gorm.io/gorm"
)

// ─── TransferRepo ─────────────────────────────────────────────────────────────

type TransferFilter struct {
	AccountID *uuid.UUID
	Month     int
	Year      int
	Page      int
	PerPage   int
}

type TransferRepo struct{ db *gorm.DB }

func NewTransferRepo(db *gorm.DB) *TransferRepo { return &TransferRepo{db: db} }

func (r *TransferRepo) List(userID uuid.UUID, f TransferFilter) ([]model.Transfer, int64, error) {
	var list []model.Transfer
	var total int64

	q := r.db.Model(&model.Transfer{}).
		Preload("FromAccount").Preload("ToAccount").
		Where("user_id = ?", userID)

	if f.AccountID != nil {
		q = q.Where("from_account_id = ? OR to_account_id = ?", f.AccountID, f.AccountID)
	}
	if f.Month > 0 { q = q.Where("EXTRACT(MONTH FROM date) = ?", f.Month) }
	if f.Year > 0  { q = q.Where("EXTRACT(YEAR FROM date) = ?", f.Year) }

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, perPage := f.Page, f.PerPage
	if page < 1    { page = 1 }
	if perPage < 1 { perPage = 20 }

	err := q.Order("date DESC, created_at DESC").
		Offset((page-1)*perPage).Limit(perPage).Find(&list).Error
	return list, total, err
}

func (r *TransferRepo) FindByID(id, userID uuid.UUID) (*model.Transfer, error) {
	var t model.Transfer
	return &t, r.db.Preload("FromAccount").Preload("ToAccount").
		Where("id = ? AND user_id = ?", id, userID).First(&t).Error
}

// Cari semua transaksi yang terkait dengan transfer (transfer pair)
func (r *TransferRepo) FindTxsByTransferID(transferID uuid.UUID) ([]model.Transaction, error) {
	var txs []model.Transaction
	return txs, r.db.Where("transfer_pair_id = ?", transferID).Find(&txs).Error
}

func (r *TransferRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Transfer{}, "id = ?", id).Error
}
