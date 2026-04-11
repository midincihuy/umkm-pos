package service

import (
	"github.com/google/uuid"
	"umkm-pos/internal/dto"
	"umkm-pos/internal/model"
	"umkm-pos/internal/repository"
)


type ReportService struct {
	reportRepo  *repository.ReportRepo
	accountRepo *repository.AccountRepo
	debtRepo    *repository.DebtRepo
}

func NewReportService(reportRepo *repository.ReportRepo, accountRepo *repository.AccountRepo, debtRepo *repository.DebtRepo) *ReportService {
	return &ReportService{reportRepo: reportRepo, accountRepo: accountRepo, debtRepo: debtRepo}
}

func (s *ReportService) Summary(userID uuid.UUID, month, year int) (*dto.MonthlySummaryResp, error) {
	totalIncome, _ := s.reportRepo.SumByType(userID, model.TxIncome, month, year)
	totalExpense, _ := s.reportRepo.SumByType(userID, model.TxExpense, month, year)
	totalAssets, _ := s.accountRepo.SumBalance(userID)
	totalDebt, _ := s.debtRepo.SumRemaining(userID, model.DebtTypeDebt)
	totalReceivable, _ := s.debtRepo.SumRemaining(userID, model.DebtTypeReceivable)

	return &dto.MonthlySummaryResp{
		Month:           month,
		Year:            year,
		TotalIncome:     totalIncome,
		TotalExpense:    totalExpense,
		SurplusDeficit:  totalIncome - totalExpense,
		TotalAssets:     totalAssets,
		TotalDebt:       totalDebt,
		TotalReceivable: totalReceivable,
		NetWorth:        totalAssets - totalDebt,
	}, nil
}

func (s *ReportService) NetWorth(userID uuid.UUID) (*dto.NetWorthResp, error) {
	accounts, _ := s.accountRepo.FindByUserID(userID, true)
	totalDebt, _ := s.debtRepo.SumRemaining(userID, model.DebtTypeDebt)
	totalReceivable, _ := s.debtRepo.SumRemaining(userID, model.DebtTypeReceivable)

	var totalAssets float64
	var accountResps []dto.AccountResp
	for _, a := range accounts {
		totalAssets += a.CurrentBalance
		accountResps = append(accountResps, toAccountResp(a))
	}

	return &dto.NetWorthResp{
		TotalAssets:     totalAssets,
		TotalDebt:       totalDebt,
		TotalReceivable: totalReceivable,
		NetWorth:        totalAssets - totalDebt,
		Accounts:        accountResps,
	}, nil
}

func (s *ReportService) Breakdown(userID uuid.UUID, txType model.TransactionType, month, year int) (*dto.BreakdownResp, error) {
	rows, err := s.reportRepo.CategoryBreakdown(userID, txType, month, year)
	if err != nil {
		return nil, err
	}

	var total float64
	var items []dto.BreakdownItem
	for _, r := range rows {
		total += r.Total
		items = append(items, dto.BreakdownItem{
			CategoryID:   r.CategoryID,
			CategoryName: r.CategoryName,
			CategoryIcon: r.CategoryIcon,
			Amount:       r.Total,
		})
	}
	// Hitung persentase
	for i := range items {
		if total > 0 {
			items[i].Percentage = (items[i].Amount / total) * 100
		}
	}

	return &dto.BreakdownResp{
		Month:  month,
		Year:   year,
		Type:   string(txType),
		Total:  total,
		Items:  items,
	}, nil
}

func (s *ReportService) Trend(userID uuid.UUID, months int) (*dto.TrendResp, error) {
	if months <= 0 || months > 24 {
		months = 12
	}
	sums, err := s.reportRepo.MonthlySums(userID, months)
	if err != nil {
		return nil, err
	}

	// Gabungkan income & expense per bulan
	type monthKey struct{ Year, Month int }
	grouped := make(map[monthKey]*dto.TrendItem)
	for _, s := range sums {
		k := monthKey{s.Year, s.Month}
		if grouped[k] == nil {
			grouped[k] = &dto.TrendItem{Year: s.Year, Month: s.Month}
		}
		switch s.Type {
		case "income":
			grouped[k].Income = s.Total
		case "expense":
			grouped[k].Expense = s.Total
		}
	}

	var items []dto.TrendItem
	for _, v := range grouped {
		v.Surplus = v.Income - v.Expense
		items = append(items, *v)
	}

	return &dto.TrendResp{Months: months, Items: items}, nil
}
