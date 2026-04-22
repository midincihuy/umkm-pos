package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"umkm-pos/internal/model"
	"umkm-pos/internal/service"
	"umkm-pos/pkg/response"
)

type ReportHandler struct{ svc *service.ReportService }

func NewReportHandler(svc *service.ReportService) *ReportHandler { return &ReportHandler{svc: svc} }

func (h *ReportHandler) Summary(c *gin.Context) {
	userID := getUserID(c)
	now := time.Now()
	month := intQuery(c, "month", int(now.Month()))
	year := intQuery(c, "year", now.Year())
	if month < 1 || month > 12 {
		response.Error(c, http.StatusBadRequest, "Bulan tidak valid (1-12)", nil)
		return
	}
	resp, err := h.svc.Summary(userID, month, year)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "OK", resp)
}

func (h *ReportHandler) ExpenseBreakdown(c *gin.Context) {
	userID := getUserID(c)
	now := time.Now()
	month := intQuery(c, "month", int(now.Month()))
	year := intQuery(c, "year", now.Year())
	resp, err := h.svc.Breakdown(userID, model.TxExpense, month, year)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "OK", resp)
}

func (h *ReportHandler) IncomeBreakdown(c *gin.Context) {
	userID := getUserID(c)
	now := time.Now()
	month := intQuery(c, "month", int(now.Month()))
	year := intQuery(c, "year", now.Year())
	resp, err := h.svc.Breakdown(userID, model.TxIncome, month, year)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "OK", resp)
}

func (h *ReportHandler) Trend(c *gin.Context) {
	userID := getUserID(c)
	months := intQuery(c, "months", 12)
	if months < 1 || months > 24 {
		months = 12
	}
	resp, err := h.svc.Trend(userID, months)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "OK", resp)
}

func (h *ReportHandler) NetWorth(c *gin.Context) {
	userID := getUserID(c)
	resp, err := h.svc.NetWorth(userID)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "OK", resp)
}
