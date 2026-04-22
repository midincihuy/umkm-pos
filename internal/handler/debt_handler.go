package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"umkm-pos/internal/dto"
	"umkm-pos/internal/model"
	"umkm-pos/internal/repository"
	"umkm-pos/internal/service"
	"umkm-pos/pkg/response"
)

type DebtHandler struct{ svc *service.DebtService }

func NewDebtHandler(svc *service.DebtService) *DebtHandler { return &DebtHandler{svc: svc} }

func (h *DebtHandler) List(c *gin.Context) {
	userID := getUserID(c)
	f := repository.DebtFilter{
		Type:    c.Query("type"),
		Status:  c.Query("status"),
		Page:    intQuery(c, "page", 1),
		PerPage: intQuery(c, "per_page", 20),
	}
	list, total, err := h.svc.List(userID, f)
	if err != nil {
		handleErr(c, err)
		return
	}
	totalPages := int(total) / f.PerPage
	if int(total)%f.PerPage != 0 {
		totalPages++
	}
	// Build summary
	var totalDebt, totalReceivable float64
	resps := make([]dto.DebtResp, len(list))
	for i, d := range list {
		resps[i] = toDebtRespFull(d)
		if d.Type == model.DebtTypeDebt {
			totalDebt += d.RemainingAmount()
		} else {
			totalReceivable += d.RemainingAmount()
		}
	}
	response.Success(c, http.StatusOK, "OK", gin.H{
		"data":             resps,
		"total_debt":       totalDebt,
		"total_receivable": totalReceivable,
		"meta":             gin.H{"page": f.Page, "per_page": f.PerPage, "total": total, "total_pages": totalPages},
	})
}

func (h *DebtHandler) Create(c *gin.Context) {
	userID := getUserID(c)
	var req dto.CreateDebtReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	debt, err := h.svc.Create(userID, &req)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Utang/piutang dicatat", toDebtRespFull(*debt))
}

func (h *DebtHandler) GetByID(c *gin.Context) {
	userID := getUserID(c)
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	resp, err := h.svc.GetByID(id, userID)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "OK", resp)
}

func (h *DebtHandler) Update(c *gin.Context) {
	userID := getUserID(c)
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateDebtReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	resp, err := h.svc.Update(id, userID, &req)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Utang/piutang diperbarui", resp)
}

func (h *DebtHandler) Cancel(c *gin.Context) {
	userID := getUserID(c)
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Cancel(id, userID); err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Utang/piutang dibatalkan", nil)
}

func (h *DebtHandler) CreatePayment(c *gin.Context) {
	userID := getUserID(c)
	debtID, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	var req dto.CreateDebtPaymentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	resp, err := h.svc.CreatePayment(debtID, userID, &req)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Cicilan berhasil dicatat", resp)
}

func (h *DebtHandler) DeletePayment(c *gin.Context) {
	userID := getUserID(c)
	debtID, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	paymentID, ok := parseUUID(c, "payment_id")
	if !ok {
		return
	}
	if err := h.svc.DeletePayment(paymentID, debtID, userID); err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Cicilan dihapus dan saldo dikembalikan", nil)
}

func toDebtRespFull(d model.Debt) dto.DebtResp {
	r := dto.DebtResp{
		ID:              d.ID,
		Type:            string(d.Type),
		ContactName:     d.ContactName,
		TotalAmount:     d.TotalAmount,
		PaidAmount:      d.PaidAmount,
		RemainingAmount: d.RemainingAmount(),
		ProgressPct:     d.ProgressPct(),
		StartDate:       d.StartDate.Format("2006-01-02"),
		Status:          string(d.Status),
		Notes:           d.Notes,
		CreatedAt:       d.CreatedAt,
	}
	if d.DueDate != nil {
		s := d.DueDate.Format("2006-01-02")
		r.DueDate = &s
	}
	for _, p := range d.Payments {
		r.Payments = append(r.Payments, dto.DebtPaymentResp{
			ID:        p.ID,
			DebtID:    p.DebtID,
			AccountID: p.AccountID,
			Amount:    p.Amount,
			Date:      p.Date.Format("2006-01-02"),
			Notes:     p.Notes,
			CreatedAt: p.CreatedAt,
		})
	}
	return r
}
