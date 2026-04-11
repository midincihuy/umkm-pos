package dto

import "github.com/google/uuid"

type MonthlySummaryResp struct {
	Month           int     `json:"month"`
	Year            int     `json:"year"`
	TotalIncome     float64 `json:"total_income"`
	TotalExpense    float64 `json:"total_expense"`
	SurplusDeficit  float64 `json:"surplus_deficit"`
	TotalAssets     float64 `json:"total_assets"`
	TotalDebt       float64 `json:"total_debt"`
	TotalReceivable float64 `json:"total_receivable"`
	NetWorth        float64 `json:"net_worth"`
}

type BreakdownItem struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	CategoryIcon string    `json:"category_icon"`
	Amount       float64   `json:"amount"`
	Percentage   float64   `json:"percentage"`
}

type BreakdownResp struct {
	Month int             `json:"month"`
	Year  int             `json:"year"`
	Type  string          `json:"type"`
	Total float64         `json:"total"`
	Items []BreakdownItem `json:"items"`
}

type TrendItem struct {
	Year    int     `json:"year"`
	Month   int     `json:"month"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Surplus float64 `json:"surplus"`
}

type TrendResp struct {
	Months int         `json:"months"`
	Items  []TrendItem `json:"items"`
}


type NetWorthResp struct {
	TotalAssets     float64       `json:"total_assets"`
	TotalDebt       float64       `json:"total_debt"`
	TotalReceivable float64       `json:"total_receivable"`
	NetWorth        float64       `json:"net_worth"`
	Accounts        []AccountResp `json:"accounts"`
}
