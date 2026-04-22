package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"umkm-pos/internal/dto"
	"umkm-pos/internal/service"
	"umkm-pos/pkg/response"
)

type ReconciliationHandler struct{ svc *service.ReconService }

func NewReconHandler(svc *service.ReconService) *ReconciliationHandler {
	return &ReconciliationHandler{svc: svc}
}

func (h *ReconciliationHandler) Preview(c *gin.Context) {
	userID := getUserID(c)
	var req dto.ReconciliationPreviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	resp, err := h.svc.Preview(userID, &req)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "OK", resp)
}

func (h *ReconciliationHandler) Create(c *gin.Context) {
	userID := getUserID(c)
	var req dto.CreateReconciliationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	resp, err := h.svc.Create(userID, &req)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Rekonsiliasi berhasil", resp)
}

func (h *ReconciliationHandler) List(c *gin.Context) {
	userID := getUserID(c)
	var accountID *uuid.UUID
	if v := c.Query("account_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			accountID = &id
		}
	}
	list, err := h.svc.List(userID, accountID)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "OK", list)
}
