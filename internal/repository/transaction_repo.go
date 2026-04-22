package repository

import (
	"time"

	"github.com/google/uuid"
	"umkm-pos/internal/model"
	"gorm.io/gorm"
)

// ─── TransactionRepo ──────────────────────────────────────────────────────────

type TransactionFilter struct {
	AccountID  *uuid.UUID
	CategoryID *uuid.UUID
	Type       string
	Month      int
	Year       int
	DateFrom   *time.Time
	DateTo     *time.Time
	Search     string
	Page       int
	PerPage    int
}

type TransactionRepo struct{ db *gorm.DB }

func NewTransactionRepo(db *gorm.DB) *TransactionRepo { return &TransactionRepo{db: db} }

func (r *TransactionRepo) List(userID uuid.UUID, f TransactionFilter) ([]model.Transaction, int64, error) {
	var list []model.Transaction
	var total int64

	q := r.db.Model(&model.Transaction{}).
		Preload("Account").Preload("Category").
		Where("transactions.user_id = ?", userID)

	if f.AccountID != nil  { q = q.Where("transactions.account_id = ?", f.AccountID) }
	if f.CategoryID != nil { q = q.Where("transactions.category_id = ?", f.CategoryID) }
	if f.Type != ""        { q = q.Where("transactions.type = ?", f.Type) }
	if f.Month > 0         { q = q.Where("EXTRACT(MONTH FROM transactions.date) = ?", f.Month) }
	if f.Year > 0          { q = q.Where("EXTRACT(YEAR FROM transactions.date) = ?", f.Year) }
	if f.DateFrom != nil   { q = q.Where("transactions.date >= ?", f.DateFrom) }
	if f.DateTo != nil     { q = q.Where("transactions.date <= ?", f.DateTo) }
	if f.Search != "" {
		like := "%" + f.Search + "%"
		q = q.Where("transactions.description ILIKE ? OR transactions.notes ILIKE ?", like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, perPage := f.Page, f.PerPage
	if page < 1    { page = 1 }
	if perPage < 1 { perPage = 20 }

	err := q.Order("transactions.date DESC, transactions.created_at DESC").
		Offset((page - 1) * perPage).Limit(perPage).Find(&list).Error
	return list, total, err
}

func (r *TransactionRepo) CreateTx(db *gorm.DB, tx *model.Transaction) error { return db.Create(tx).Error }

func (r *TransactionRepo) FindByID(id, userID uuid.UUID) (*model.Transaction, error) {
	var tx model.Transaction
	return &tx, r.db.Preload("Account").Preload("Category").
		Where("id = ? AND user_id = ?", id, userID).First(&tx).Error
}

func (r *TransactionRepo) Update(tx *model.Transaction) error { return r.db.Save(tx).Error }

func (r *TransactionRepo) SumByCategory(userID uuid.UUID, month, year int) ([]CategorySum, error) {
	var res []CategorySum
	err := r.db.Model(&model.Transaction{}).
		Select("category_id, type, COALESCE(SUM(amount), 0) as total").
		Where("user_id = ? AND type IN ('income','expense') AND category_id IS NOT NULL AND EXTRACT(MONTH FROM date) = ? AND EXTRACT(YEAR FROM date) = ?",
			userID, month, year).
		Group("category_id, type").Scan(&res).Error
	return res, err
}

type CategorySum struct {
	CategoryID uuid.UUID
	Type       string
	Total      float64
}

func (r *TransactionRepo) CategoryBreakdown(userID uuid.UUID, txType model.TransactionType, month, year int) ([]CategoryBreakdownRow, error) {
	var res []CategoryBreakdownRow
	err := r.db.Model(&model.Transaction{}).
		Select("categories.id as category_id, categories.name as category_name, categories.icon as category_icon, COALESCE(SUM(transactions.amount), 0) as total").
		Joins("LEFT JOIN categories ON categories.id = transactions.category_id").
		Where("transactions.user_id = ? AND transactions.type = ? AND EXTRACT(MONTH FROM transactions.date) = ? AND EXTRACT(YEAR FROM transactions.date) = ?",
			userID, txType, month, year).
		Group("categories.id, categories.name, categories.icon").
		Order("total DESC").Scan(&res).Error
	return res, err
}

type CategoryBreakdownRow struct {
	CategoryID   uuid.UUID
	CategoryName string
	CategoryIcon string
	Total        float64
}

func (r *TransactionRepo) MonthlySums(userID uuid.UUID, months int) ([]MonthlySum, error) {
	var res []MonthlySum
	err := r.db.Raw(`
		SELECT EXTRACT(YEAR FROM date)::int as year,
		       EXTRACT(MONTH FROM date)::int as month,
		       type,
		       COALESCE(SUM(amount), 0) as total
		FROM transactions
		WHERE user_id = ? AND type IN ('income','expense')
		  AND date >= NOW() - INTERVAL '1 month' * ?
		GROUP BY year, month, type
		ORDER BY year DESC, month DESC
	`, userID, months).Scan(&res).Error
	return res, err
}

type MonthlySum struct {
	Year  int
	Month int
	Type  string
	Total float64
}

// SumByDebtPayment total pembayaran cicilan dalam bulan tertentu
func (r *TransactionRepo) SumByDebtPayment(userID uuid.UUID, month, year int) (float64, error) {
	var total float64
	return total, r.db.Model(&model.Transaction{}).
		Where("user_id = ? AND type = 'debt_payment' AND EXTRACT(MONTH FROM date) = ? AND EXTRACT(YEAR FROM date) = ?",
			userID, month, year).
		Select("COALESCE(SUM(amount), 0)").Scan(&total).Error
}
