package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"umkm-pos/internal/dto"
	"umkm-pos/internal/service"
	"umkm-pos/pkg/response"
)

type AuthHandler struct{ svc *service.AuthService }

func NewAuthHandler(svc *service.AuthService) *AuthHandler { return &AuthHandler{svc: svc} }

func (h *AuthHandler) Me(c *gin.Context) {
	userID := getUserID(c)
	user, err := h.svc.GetByID(userID)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "OK", user)
}

func (h *AuthHandler) UpdateMe(c *gin.Context) {
	userID := getUserID(c)
	var req dto.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	user, err := h.svc.UpdateUser(userID, &req)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Profil diperbarui", user)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	response.Error(c, http.StatusGone, "Gunakan Supabase Auth SDK untuk mengubah password", nil)
}
