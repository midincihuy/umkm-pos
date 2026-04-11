package service

import (
	"github.com/google/uuid"
	"umkm-pos/internal/dto"
	"umkm-pos/internal/model"
	"umkm-pos/internal/repository"
	"umkm-pos/pkg/apperror"
)

type BudgetService struct {
	budgetRepo *repository.BudgetRepo
	txRepo     *repository.TransactionRepo
}

func NewBudgetService(budgetRepo *repository.BudgetRepo, txRepo *repository.TransactionRepo) *BudgetService {
	return &BudgetService{budgetRepo: budgetRepo, txRepo: txRepo}
}

// GetSummary → see budget_service_ext.go

func (s *BudgetService) Upsert(userID uuid.UUID, req *dto.CreateBudgetReq) (*model.Budget, error) {
	freq := model.FreqMonthly
	if req.Frequency != "" {
		freq = model.FrequencyType(req.Frequency)
	}
	b := &model.Budget{
		UserID:        userID,
		CategoryID:    req.CategoryID,
		Month:         req.Month,
		Year:          req.Year,
		PlannedAmount: req.PlannedAmount,
		Frequency:     freq,
	}
	return b, s.budgetRepo.Upsert(b)
}

func (s *BudgetService) Copy(userID uuid.UUID, req *dto.CopyBudgetReq) (*dto.CopyBudgetResp, error) {
	from, err := s.budgetRepo.FindByMonth(userID, req.FromMonth, req.FromYear)
	if err != nil {
		return nil, err
	}
	copied, skipped := 0, 0
	for _, b := range from {
		nb := &model.Budget{
			UserID:        userID,
			CategoryID:    b.CategoryID,
			Month:         req.ToMonth,
			Year:          req.ToYear,
			PlannedAmount: b.PlannedAmount,
			Frequency:     b.Frequency,
		}
		err := s.budgetRepo.Upsert(nb)
		if err != nil {
			skipped++
		} else {
			copied++
		}
	}
	return &dto.CopyBudgetResp{CopiedCount: copied, SkippedCount: skipped}, nil
}

// GetSummary versi lengkap dengan actual_amount per kategori
func (s *BudgetService) GetSummary(userID uuid.UUID, month, year int) (*dto.BudgetSummaryResp, error) {
	budgets, err := s.budgetRepo.FindByMonth(userID, month, year)
	if err != nil {
		return nil, err
	}

	// Ambil realisasi semua kategori dalam bulan ini sekali query
	categorySums, err := s.txRepo.SumByCategory(userID, month, year)
	if err != nil {
		return nil, err
	}
	// Buat map untuk lookup cepat
	actualMap := make(map[uuid.UUID]float64, len(categorySums))
	for _, cs := range categorySums {
		actualMap[cs.CategoryID] += cs.Total
	}

	var budgetResps []dto.BudgetResp
	var totalIncomePlan, totalExpensePlan float64
	var totalIncomeActual, totalExpenseActual float64

	for _, b := range budgets {
		actual := actualMap[b.CategoryID]
		monthly := b.MonthlyEquiv()
		remaining := monthly - actual
		status := computeStatus(monthly, actual, b.Category.Type)

		br := dto.BudgetResp{
			ID:            b.ID,
			CategoryID:    b.CategoryID,
			CategoryName:  b.Category.Name,
			CategoryIcon:  b.Category.Icon,
			CategoryType:  string(b.Category.Type),
			Month:         b.Month,
			Year:          b.Year,
			PlannedAmount: b.PlannedAmount,
			MonthlyEquiv:  monthly,
			ActualAmount:  actual,
			Remaining:     remaining,
			Frequency:     string(b.Frequency),
			Status:        status,
		}

		if b.Category.Type == model.CategoryIncome {
			totalIncomePlan   += monthly
			totalIncomeActual += actual
		} else {
			totalExpensePlan   += monthly
			totalExpenseActual += actual
		}
		budgetResps = append(budgetResps, br)
	}

	debtObligation, _ := s.txRepo.SumByDebtPayment(userID, month, year)

	return &dto.BudgetSummaryResp{
		Month:               month,
		Year:                year,
		TotalIncomePlan:     totalIncomePlan,
		TotalIncomeActual:   totalIncomeActual,
		TotalExpensePlan:    totalExpensePlan,
		TotalExpenseActual:  totalExpenseActual,
		SurplusDeficit:      totalIncomeActual - totalExpenseActual,
		DebtObligations:     debtObligation,
		Budgets:             budgetResps,
	}, nil
}

func computeStatus(planned, actual float64, catType model.CategoryType) string {
	if planned == 0 {
		return "no_data"
	}
	pct := actual / planned
	if catType == model.CategoryExpense {
		if pct >= 1.0    { return "over_budget" }
		if pct >= 0.8    { return "warning" }
		return "safe"
	}
	// income: realisasi < rencana = warning
	if pct < 0.8  { return "warning" }
	if pct < 1.0  { return "safe" }
	return "safe"
}

func (s *BudgetService) Delete(id, userID uuid.UUID) error {
	b, err := s.budgetRepo.FindByID(id, userID)
	if err != nil {
		return apperror.NotFound("Budget tidak ditemukan")
	}
	return s.budgetRepo.Delete(b.ID)
}
