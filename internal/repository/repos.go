package repository

import (
	"time"

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

// ─── ReportRepo ───────────────────────────────────────────────────────────────

type ReportRepo struct{ db *gorm.DB }

func NewReportRepo(db *gorm.DB) *ReportRepo { return &ReportRepo{db: db} }

func (r *ReportRepo) SumByType(userID uuid.UUID, txType model.TransactionType, month, year int) (float64, error) {
	var total float64
	return total, r.db.Model(&model.Transaction{}).
		Where("user_id = ? AND type = ? AND EXTRACT(MONTH FROM date) = ? AND EXTRACT(YEAR FROM date) = ?",
			userID, txType, month, year).
		Select("COALESCE(SUM(amount), 0)").Scan(&total).Error
}

func (r *ReportRepo) MonthlySums(userID uuid.UUID, months int) ([]MonthlySum, error) {
	var res []MonthlySum
	err := r.db.Raw(`
		SELECT EXTRACT(YEAR FROM date)::int  AS year,
		       EXTRACT(MONTH FROM date)::int AS month,
		       type,
		       COALESCE(SUM(amount), 0)      AS total
		FROM transactions
		WHERE user_id = ?
		  AND type IN ('income','expense')
		  AND date >= NOW() - (? * INTERVAL '1 month')
		GROUP BY year, month, type
		ORDER BY year DESC, month DESC
	`, userID, months).Scan(&res).Error
	return res, err
}

func (r *ReportRepo) CategoryBreakdown(userID uuid.UUID, txType model.TransactionType, month, year int) ([]CategoryBreakdownRow, error) {
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

// SumByDebtPayment total pembayaran cicilan dalam bulan tertentu
func (r *TransactionRepo) SumByDebtPayment(userID uuid.UUID, month, year int) (float64, error) {
	var total float64
	return total, r.db.Model(&model.Transaction{}).
		Where("user_id = ? AND type = 'debt_payment' AND EXTRACT(MONTH FROM date) = ? AND EXTRACT(YEAR FROM date) = ?",
			userID, month, year).
		Select("COALESCE(SUM(amount), 0)").Scan(&total).Error
}
