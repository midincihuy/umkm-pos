package dto

import "github.com/google/uuid"

type CreateCategoryReq struct {
	Name  string `json:"name"  validate:"required,min=1,max=100"`
	Type  string `json:"type"  validate:"required,oneof=income expense"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

type CategoryResp struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Icon      string    `json:"icon"`
	Color     string    `json:"color"`
	IsSystem  bool      `json:"is_system"`
	IsActive  bool      `json:"is_active"`
	SortOrder int       `json:"sort_order"`
}

type CategoryListResp struct {
	Income  []CategoryResp `json:"income"`
	Expense []CategoryResp `json:"expense"`
}
