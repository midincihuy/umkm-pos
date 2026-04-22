package repository

import (
	"github.com/google/uuid"
	"umkm-pos/internal/model"
	"gorm.io/gorm"
)

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
