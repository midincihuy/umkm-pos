package dto

import "github.com/google/uuid"

type CreateBudgetReq struct {
	CategoryID    uuid.UUID `json:"category_id"    validate:"required"`
	Month         int       `json:"month"          validate:"required,min=1,max=12"`
	Year          int       `json:"year"           validate:"required,min=2000"`
	PlannedAmount float64   `json:"planned_amount" validate:"min=0"`
	Frequency     string    `json:"frequency"      validate:"omitempty,oneof=monthly quarterly yearly"`
}

type CopyBudgetReq struct {
	FromMonth int `json:"from_month" validate:"required,min=1,max=12"`
	FromYear  int `json:"from_year"  validate:"required"`
	ToMonth   int `json:"to_month"   validate:"required,min=1,max=12"`
	ToYear    int `json:"to_year"    validate:"required"`
}

type BudgetResp struct {
	ID            uuid.UUID `json:"id"`
	CategoryID    uuid.UUID `json:"category_id"`
	CategoryName  string    `json:"category_name"`
	CategoryIcon  string    `json:"category_icon"`
	CategoryType  string    `json:"category_type"`
	Month         int       `json:"month"`
	Year          int       `json:"year"`
	PlannedAmount float64   `json:"planned_amount"`
	MonthlyEquiv  float64   `json:"monthly_equiv"`
	ActualAmount  float64   `json:"actual_amount"`
	Remaining     float64   `json:"remaining"`
	Frequency     string    `json:"frequency"`
	Status        string    `json:"status"` // safe, warning, over_budget, achieved, not_started
}

type BudgetSummaryResp struct {
	Month               int          `json:"month"`
	Year                int          `json:"year"`
	TotalIncomePlan     float64      `json:"total_income_plan"`
	TotalIncomeActual   float64      `json:"total_income_actual"`
	TotalExpensePlan    float64      `json:"total_expense_plan"`
	TotalExpenseActual  float64      `json:"total_expense_actual"`
	SurplusDeficit      float64      `json:"surplus_deficit"`
	DebtObligations     float64      `json:"debt_obligations"`
	Budgets             []BudgetResp `json:"budgets"`
}

type CopyBudgetResp struct {
	CopiedCount  int `json:"copied_count"`
	SkippedCount int `json:"skipped_count"`
}
