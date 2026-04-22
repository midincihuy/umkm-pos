package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"umkm-pos/internal/dto"
	"umkm-pos/internal/model"
	"umkm-pos/internal/repository"
	"umkm-pos/internal/service"
	"umkm-pos/pkg/response"
)

type TransactionHandler struct{ svc *service.TransactionService }

func NewTransactionHandler(svc *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{svc: svc}
}

func (h *TransactionHandler) List(c *gin.Context) {
	userID := getUserID(c)

	f := repository.TransactionFilter{
		Month:   intQuery(c, "month", 0),
		Year:    intQuery(c, "year", 0),
		Page:    intQuery(c, "page", 1),
		PerPage: intQuery(c, "per_page", 20),
	}
	if f.PerPage > 100 {
		f.PerPage = 100
	}
	if t := c.Query("type"); t != "" {
		f.Type = t
	}
	if s := c.Query("search"); s != "" {
		f.Search = s
	}
	if v := c.Query("account_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			f.AccountID = &id
		}
	}
	if v := c.Query("category_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			f.CategoryID = &id
		}
	}
	if v := c.Query("date_from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			f.DateFrom = &t
		}
	}
	if v := c.Query("date_to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			f.DateTo = &t
		}
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

	response.Success(c, http.StatusOK, "OK", gin.H{
		"data": toTxRespList(list),
		"meta": gin.H{
			"page":        f.Page,
			"per_page":    f.PerPage,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func (h *TransactionHandler) Create(c *gin.Context) {
	userID := getUserID(c)
	var req dto.CreateTransactionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	tx, err := h.svc.Create(userID, &req)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Transaksi berhasil dicatat", toTxResp(*tx))
}

func (h *TransactionHandler) GetByID(c *gin.Context) {
	userID := getUserID(c)
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	tx, err := h.svc.GetByID(id, userID)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "OK", toTxResp(*tx))
}

func (h *TransactionHandler) Update(c *gin.Context) {
	userID := getUserID(c)
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateTransactionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	tx, err := h.svc.Update(id, userID, &req)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Transaksi diperbarui", toTxResp(*tx))
}

func (h *TransactionHandler) Delete(c *gin.Context) {
	userID := getUserID(c)
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(id, userID); err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Transaksi dihapus", nil)
}

func toTxResp(tx model.Transaction) dto.TransactionResp {
	r := dto.TransactionResp{
		ID:            tx.ID,
		AccountID:     tx.AccountID,
		AccountName:   tx.Account.Name,
		CategoryID:    tx.CategoryID,
		Type:          string(tx.Type),
		Amount:        tx.Amount,
		BalanceImpact: tx.BalanceImpact,
		Date:          tx.Date.Format("2006-01-02"),
		Description:   tx.Description,
		Notes:         tx.Notes,
		CreatedAt:     tx.CreatedAt,
	}
	if tx.Category != nil {
		name := tx.Category.Name
		r.CategoryName = &name
	}
	return r
}

func toTxRespList(list []model.Transaction) []dto.TransactionResp {
	out := make([]dto.TransactionResp, len(list))
	for i, tx := range list {
		out[i] = toTxResp(tx)
	}
	return out
}
