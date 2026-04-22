package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"umkm-pos/internal/dto"
	"umkm-pos/internal/service"
	"umkm-pos/pkg/response"
)

type AccountHandler struct{ svc *service.AccountService }

func NewAccountHandler(svc *service.AccountService) *AccountHandler {
	return &AccountHandler{svc: svc}
}

func (h *AccountHandler) List(c *gin.Context) {
	userID := getUserID(c)
	resp, err := h.svc.List(userID, true)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "OK", resp)
}

func (h *AccountHandler) Create(c *gin.Context) {
	userID := getUserID(c)
	var req dto.CreateAccountReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	acc, err := h.svc.Create(userID, &req)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Rekening berhasil dibuat", acc)
}

func (h *AccountHandler) GetByID(c *gin.Context) {
	userID := getUserID(c)
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	acc, err := h.svc.GetByID(id, userID)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "OK", acc)
}

func (h *AccountHandler) Update(c *gin.Context) {
	userID := getUserID(c)
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateAccountReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	acc, err := h.svc.Update(id, userID, &req)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Rekening diperbarui", acc)
}

func (h *AccountHandler) Delete(c *gin.Context) {
	userID := getUserID(c)
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(id, userID); err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Rekening dihapus", nil)
}
