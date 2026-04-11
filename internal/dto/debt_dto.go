package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateDebtReq struct {
	Type        string    `json:"type"         validate:"required,oneof=debt receivable"`
	ContactName string    `json:"contact_name" validate:"required,min=1,max=100"`
	TotalAmount float64   `json:"total_amount" validate:"required,gt=0"`
	AccountID   uuid.UUID `json:"account_id"   validate:"required"`
	StartDate   DateOnly  `json:"start_date"   validate:"required"`
	DueDate     *DateOnly `json:"due_date"`
	Notes       string    `json:"notes"`
}

type UpdateDebtReq struct {
	ContactName *string    `json:"contact_name" validate:"omitempty,min=1"`
	DueDate     *time.Time `json:"due_date"`
	Notes       *string    `json:"notes"`
}

type CreateDebtPaymentReq struct {
	AccountID uuid.UUID `json:"account_id" validate:"required"`
	Amount    float64   `json:"amount"     validate:"required,gt=0"`
	Date      DateOnly  `json:"date"       validate:"required"`
	Notes     string    `json:"notes"`
}

type CancelDebtReq struct {
	Notes string `json:"notes"`
}

type DebtPaymentResp struct {
	ID          uuid.UUID `json:"id"`
	DebtID      uuid.UUID `json:"debt_id"`
	AccountID   uuid.UUID `json:"account_id"`
	AccountName string    `json:"account_name"`
	Amount      float64   `json:"amount"`
	Date        string    `json:"date"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
}

type DebtResp struct {
	ID              uuid.UUID        `json:"id"`
	Type            string           `json:"type"`
	ContactName     string           `json:"contact_name"`
	TotalAmount     float64          `json:"total_amount"`
	PaidAmount      float64          `json:"paid_amount"`
	RemainingAmount float64          `json:"remaining_amount"`
	ProgressPct     float64          `json:"progress_pct"`
	StartDate       string           `json:"start_date"`
	DueDate         *string          `json:"due_date"`
	Status          string           `json:"status"`
	Notes           string           `json:"notes"`
	Payments        []DebtPaymentResp `json:"payments,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
}

type DebtListResp struct {
	Data            []DebtResp `json:"data"`
	TotalDebt       float64    `json:"total_debt"`
	TotalReceivable float64    `json:"total_receivable"`
}

type CreateDebtPaymentResp struct {
	Payment      DebtPaymentResp `json:"payment"`
	Debt         DebtResp        `json:"debt"`
	IsFullyPaid  bool            `json:"is_fully_paid"`
}
