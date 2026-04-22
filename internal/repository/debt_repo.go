package repository

import (
	"github.com/google/uuid"
	"umkm-pos/internal/model"
	"gorm.io/gorm"
)

// ─── DebtRepo ─────────────────────────────────────────────────────────────────

type DebtFilter struct {
	Type    string
	Status  string
	Page    int
	PerPage int
}

type DebtRepo struct{ db *gorm.DB }

func NewDebtRepo(db *gorm.DB) *DebtRepo { return &DebtRepo{db: db} }

func (r *DebtRepo) List(userID uuid.UUID, f DebtFilter) ([]model.Debt, int64, error) {
	var list []model.Debt
	var total int64

	q := r.db.Model(&model.Debt{}).Where("user_id = ?", userID)
	if f.Type != ""   { q = q.Where("type = ?", f.Type) }
	if f.Status != "" { q = q.Where("status = ?", f.Status) }

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, perPage := f.Page, f.PerPage
	if page < 1    { page = 1 }
	if perPage < 1 { perPage = 20 }

	err := q.Preload("Payments").
		Order("CASE status WHEN 'active' THEN 0 ELSE 1 END, created_at DESC").
		Offset((page-1)*perPage).Limit(perPage).Find(&list).Error
	return list, total, err
}

func (r *DebtRepo) FindByID(id, userID uuid.UUID) (*model.Debt, error) {
	var d model.Debt
	return &d, r.db.Preload("Payments.Account").
		Where("id = ? AND user_id = ?", id, userID).First(&d).Error
}

func (r *DebtRepo) Create(d *model.Debt) error          { return r.db.Create(d).Error }
func (r *DebtRepo) Update(d *model.Debt) error          { return r.db.Save(d).Error }
func (r *DebtRepo) UpdateTx(tx *gorm.DB, d *model.Debt) error { return tx.Save(d).Error }

func (r *DebtRepo) CreatePaymentTx(tx *gorm.DB, p *model.DebtPayment) error {
	return tx.Create(p).Error
}
func (r *DebtRepo) FindPaymentByID(id uuid.UUID) (*model.DebtPayment, error) {
	var p model.DebtPayment
	return &p, r.db.First(&p, "id = ?", id).Error
}
func (r *DebtRepo) DeletePayment(tx *gorm.DB, id uuid.UUID) error {
	return tx.Delete(&model.DebtPayment{}, "id = ?", id).Error
}

func (r *DebtRepo) SumRemaining(userID uuid.UUID, debtType model.DebtType) (float64, error) {
	var total float64
	return total, r.db.Model(&model.Debt{}).
		Where("user_id = ? AND type = ? AND status = 'active'", userID, debtType).
		Select("COALESCE(SUM(total_amount - paid_amount), 0)").Scan(&total).Error
}
