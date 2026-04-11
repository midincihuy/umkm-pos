package dto

import (
	"time"

	"github.com/google/uuid"
)

type ReconciliationPreviewReq struct {
	AccountID     uuid.UUID `json:"account_id"     validate:"required"`
	ActualBalance float64   `json:"actual_balance" validate:"min=0"`
}

type CreateReconciliationReq struct {
	AccountID     uuid.UUID `json:"account_id"     validate:"required"`
	ActualBalance float64   `json:"actual_balance" validate:"min=0"`
	Date          DateOnly  `json:"date"           validate:"required"`
	Notes         string    `json:"notes"`
}

type ReconciliationPreviewResp struct {
	AccountID       uuid.UUID `json:"account_id"`
	AccountName     string    `json:"account_name"`
	BalanceBefore   float64   `json:"balance_before"`
	ActualBalance   float64   `json:"actual_balance"`
	Difference      float64   `json:"difference"`
	NeedsAdjustment bool      `json:"needs_adjustment"`
	Warning         *string   `json:"warning"`
}

type CashAdjustmentResp struct {
	ID            uuid.UUID `json:"id"`
	AccountID     uuid.UUID `json:"account_id"`
	AccountName   string    `json:"account_name"`
	BalanceBefore float64   `json:"balance_before"`
	ActualBalance float64   `json:"actual_balance"`
	Difference    float64   `json:"difference"`
	Date          string    `json:"date"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
}
