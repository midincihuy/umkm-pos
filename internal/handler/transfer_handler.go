package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"umkm-pos/internal/dto"
	"umkm-pos/internal/model"
	"umkm-pos/internal/repository"
	"umkm-pos/internal/service"
	"umkm-pos/pkg/response"
)

type TransferHandler struct{ svc *service.TransferService }

func NewTransferHandler(svc *service.TransferService) *TransferHandler {
	return &TransferHandler{svc: svc}
}

func (h *TransferHandler) List(c *gin.Context) {
	userID := getUserID(c)
	f := repository.TransferFilter{
		Month:   intQuery(c, "month", 0),
		Year:    intQuery(c, "year", 0),
		Page:    intQuery(c, "page", 1),
		PerPage: intQuery(c, "per_page", 20),
	}
	if v := c.Query("account_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			f.AccountID = &id
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
		"data": toTransferRespList(list),
		"meta": gin.H{"page": f.Page, "per_page": f.PerPage, "total": total, "total_pages": totalPages},
	})
}

func (h *TransferHandler) Create(c *gin.Context) {
	userID := getUserID(c)
	var req dto.CreateTransferReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	t, err := h.svc.Create(userID, &req)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Transfer berhasil", toTransferResp(*t))
}

func (h *TransferHandler) GetByID(c *gin.Context) {
	userID := getUserID(c)
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	t, err := h.svc.GetByID(id, userID)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "OK", toTransferResp(*t))
}

func (h *TransferHandler) Delete(c *gin.Context) {
	userID := getUserID(c)
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(id, userID); err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Transfer dihapus dan saldo dikembalikan", nil)
}

func toTransferResp(t model.Transfer) dto.TransferResp {
	return dto.TransferResp{
		ID:            t.ID,
		FromAccountID: t.FromAccountID,
		FromAccount:   t.FromAccount.Name,
		ToAccountID:   t.ToAccountID,
		ToAccount:     t.ToAccount.Name,
		Amount:        t.Amount,
		Fee:           t.Fee,
		Date:          t.Date.Format("2006-01-02"),
		Notes:         t.Notes,
		CreatedAt:     t.CreatedAt,
	}
}

func toTransferRespList(list []model.Transfer) []dto.TransferResp {
	out := make([]dto.TransferResp, len(list))
	for i, t := range list {
		out[i] = toTransferResp(t)
	}
	return out
}
