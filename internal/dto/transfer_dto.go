package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateTransferReq struct {
	FromAccountID uuid.UUID `json:"from_account_id" validate:"required"`
	ToAccountID   uuid.UUID `json:"to_account_id"   validate:"required"`
	Amount        float64   `json:"amount"          validate:"required,gt=0"`
	Fee           float64   `json:"fee"             validate:"min=0"`
	Date          DateOnly  `json:"date"            validate:"required"`
	Notes         string    `json:"notes"`
}

type TransferResp struct {
	ID            uuid.UUID `json:"id"`
	FromAccountID uuid.UUID `json:"from_account_id"`
	FromAccount   string    `json:"from_account"`
	ToAccountID   uuid.UUID `json:"to_account_id"`
	ToAccount     string    `json:"to_account"`
	Amount        float64   `json:"amount"`
	Fee           float64   `json:"fee"`
	Date          string    `json:"date"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
}
