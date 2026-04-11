package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateTransactionReq struct {
	AccountID   uuid.UUID `json:"account_id"  validate:"required"`
	CategoryID  uuid.UUID `json:"category_id" validate:"required"`
	Type        string    `json:"type"        validate:"required,oneof=income expense"`
	Amount      float64   `json:"amount"      validate:"required,gt=0"`
	Date        DateOnly  `json:"date"        validate:"required"`
	Description string    `json:"description"`
	Notes       string    `json:"notes"`
}

type UpdateTransactionReq struct {
	CategoryID  *uuid.UUID `json:"category_id"`
	Amount      *float64   `json:"amount"      validate:"omitempty,gt=0"`
	Date        *time.Time `json:"date"`
	Description *string    `json:"description"`
	Notes       *string    `json:"notes"`
}

type TransactionFilter struct {
	Month      *int       `form:"month"       validate:"omitempty,min=1,max=12"`
	Year       *int       `form:"year"`
	Type       *string    `form:"type"`
	CategoryID *uuid.UUID `form:"category_id"`
	AccountID  *uuid.UUID `form:"account_id"`
	DateFrom   *time.Time `form:"date_from"`
	DateTo     *time.Time `form:"date_to"`
	Search     *string    `form:"search"`
	Page       int        `form:"page"        validate:"min=1"`
	PerPage    int        `form:"per_page"    validate:"min=1,max=100"`
}

type TransactionResp struct {
	ID            uuid.UUID  `json:"id"`
	AccountID     uuid.UUID  `json:"account_id"`
	AccountName   string     `json:"account_name"`
	CategoryID    *uuid.UUID `json:"category_id"`
	CategoryName  *string    `json:"category_name"`
	Type          string     `json:"type"`
	Amount        float64    `json:"amount"`
	BalanceImpact float64    `json:"balance_impact"`
	Date          string     `json:"date"`
	Description   string     `json:"description"`
	Notes         string     `json:"notes"`
	CreatedAt     time.Time  `json:"created_at"`
}
