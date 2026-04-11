package repository

import (
	"github.com/google/uuid"
	"umkm-pos/internal/model"
	"gorm.io/gorm"
)

type AccountRepo struct {
	db *gorm.DB
}

func NewAccountRepo(db *gorm.DB) *AccountRepo {
	return &AccountRepo{db: db}
}

func (r *AccountRepo) FindByUserID(userID uuid.UUID, activeOnly bool) ([]model.Account, error) {
	var accounts []model.Account
	q := r.db.Where("user_id = ?", userID)
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	return accounts, q.Order("created_at ASC").Find(&accounts).Error
}

func (r *AccountRepo) FindByID(id, userID uuid.UUID) (*model.Account, error) {
	var account model.Account
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&account).Error
	return &account, err
}

func (r *AccountRepo) Create(account *model.Account) error {
	return r.db.Create(account).Error
}

func (r *AccountRepo) Update(account *model.Account) error {
	return r.db.Save(account).Error
}

// UpdateBalance digunakan dalam transaksi DB — tx harus dari db.Transaction()
func (r *AccountRepo) UpdateBalance(tx *gorm.DB, id uuid.UUID, delta float64) error {
	return tx.Model(&model.Account{}).
		Where("id = ?", id).
		Update("current_balance", gorm.Expr("current_balance + ?", delta)).Error
}

func (r *AccountRepo) SetBalance(tx *gorm.DB, id uuid.UUID, balance float64) error {
	return tx.Model(&model.Account{}).
		Where("id = ?", id).
		Update("current_balance", balance).Error
}

func (r *AccountRepo) SumBalance(userID uuid.UUID) (float64, error) {
	var total float64
	err := r.db.Model(&model.Account{}).
		Where("user_id = ? AND is_active = true", userID).
		Select("COALESCE(SUM(current_balance), 0)").
		Scan(&total).Error
	return total, err
}

func (r *AccountRepo) HasTransactions(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.Transaction{}).Where("account_id = ?", id).Count(&count).Error
	return count > 0, err
}

// FindByIDTx digunakan di dalam database transaction
func (r *AccountRepo) FindByIDTx(tx *gorm.DB, id uuid.UUID) (*model.Account, error) {
	var a model.Account
	return &a, tx.First(&a, "id = ?", id).Error
}
