package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateAccountReq struct {
	Name           string  `json:"name"            validate:"required,min=1,max=100"`
	Type           string  `json:"type"            validate:"required,oneof=bank cash ewallet investment"`
	OpeningBalance float64 `json:"opening_balance" validate:"min=0"`
	Icon           string  `json:"icon"`
	Color          string  `json:"color"`
}

type UpdateAccountReq struct {
	Name     *string `json:"name"      validate:"omitempty,min=1,max=100"`
	Icon     *string `json:"icon"`
	Color    *string `json:"color"`
	IsActive *bool   `json:"is_active"`
}

type AccountResp struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	OpeningBalance float64   `json:"opening_balance"`
	CurrentBalance float64   `json:"current_balance"`
	Icon           string    `json:"icon"`
	Color          string    `json:"color"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}

type AccountListResp struct {
	Data         []AccountResp `json:"data"`
	TotalBalance float64       `json:"total_balance"`
}
