package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"umkm-pos/internal/dto"
	"umkm-pos/internal/service"
	"umkm-pos/pkg/response"
)

type BudgetHandler struct{ svc *service.BudgetService }

func NewBudgetHandler(svc *service.BudgetService) *BudgetHandler { return &BudgetHandler{svc: svc} }

func (h *BudgetHandler) Get(c *gin.Context) {
	userID := getUserID(c)
	now := time.Now()
	month := intQuery(c, "month", int(now.Month()))
	year := intQuery(c, "year", now.Year())
	if month < 1 || month > 12 {
		response.Error(c, http.StatusBadRequest, "Bulan tidak valid (1-12)", nil)
		return
	}
	resp, err := h.svc.GetSummary(userID, month, year)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "OK", resp)
}

func (h *BudgetHandler) Upsert(c *gin.Context) {
	userID := getUserID(c)
	var req dto.CreateBudgetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	b, err := h.svc.Upsert(userID, &req)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Budget disimpan", b)
}

func (h *BudgetHandler) Copy(c *gin.Context) {
	userID := getUserID(c)
	var req dto.CopyBudgetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	resp, err := h.svc.Copy(userID, &req)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Budget disalin", resp)
}

func (h *BudgetHandler) Delete(c *gin.Context) {
	userID := getUserID(c)
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(id, userID); err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Budget dihapus", nil)
}
