package dto

import "github.com/google/uuid"

// UpdateUserReq untuk PATCH /auth/me
type UpdateUserReq struct {
	Name     string `json:"name"     validate:"omitempty,min=2,max=100"`
	Currency string `json:"currency" validate:"omitempty,len=3"`
}

// UserResp untuk GET /auth/me dan PATCH /auth/me
type UserResp struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Currency string    `json:"currency"`
}

// ChangePasswordReq — tidak digunakan di backend.
// Ganti password via: supabase.auth.updateUser({ password: newPassword })
type ChangePasswordReq struct{}
