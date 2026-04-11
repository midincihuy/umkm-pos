package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"umkm-pos/internal/dto"
	"umkm-pos/internal/model"
	"umkm-pos/internal/repository"
	"umkm-pos/internal/service"
	"umkm-pos/pkg/apperror"
	"umkm-pos/pkg/response"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

func getUserID(c *gin.Context) uuid.UUID {
	id, _ := c.Get("userID")
	return id.(uuid.UUID)
}

func getUserClaims(c *gin.Context) (interface{}, error) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, errors.New("user claims tidak ditemukan di context")
	}
	return claims, nil
}

func parseUUID(c *gin.Context, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID tidak valid", nil)
		return uuid.Nil, false
	}
	return id, true
}

func intQuery(c *gin.Context, key string, def int) int {
	v := c.Query(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return def
	}
	return n
}

func handleErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperror.ErrNotFound):
		response.Error(c, http.StatusNotFound, err.Error(), nil)
	case errors.Is(err, apperror.ErrUnauthorized):
		response.Error(c, http.StatusUnauthorized, err.Error(), nil)
	case errors.Is(err, apperror.ErrForbidden):
		response.Error(c, http.StatusForbidden, err.Error(), nil)
	case errors.Is(err, apperror.ErrInsufficientBalance):
		response.Error(c, http.StatusUnprocessableEntity, err.Error(), nil)
	case errors.Is(err, apperror.ErrConflict):
		response.Error(c, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, apperror.ErrSystemCategory):
		response.Error(c, http.StatusForbidden, err.Error(), nil)
	case errors.Is(err, apperror.ErrDebtNotActive):
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, apperror.ErrPaymentExceedsDebt):
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, apperror.ErrBadRequest):
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
	default:
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server", nil)
	}
}

// ─── AuthHandler ─────────────────────────────────────────────────────────────

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

// ─── AccountHandler ───────────────────────────────────────────────────────────

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


// ─── CategoryHandler ─────────────────────────────────────────────────────────

type CategoryHandler struct{ svc *service.CategoryService }

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

func (h *CategoryHandler) List(c *gin.Context) {
	userID := getUserID(c)
	catType := c.Query("type")
	resp, err := h.svc.List(userID, nilStr(catType), false)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "OK", resp)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	userID := getUserID(c)
	var req dto.CreateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	cat, err := h.svc.Create(userID, &req)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Kategori berhasil dibuat", cat)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	var req dto.CreateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	cat, err := h.svc.Update(id, &req)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Kategori diperbarui", cat)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(id); err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Kategori dihapus", nil)
}

// ─── TransactionHandler ──────────────────────────────────────────────────────

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

// mapper
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

// ─── TransferHandler ─────────────────────────────────────────────────────────

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

// ─── DebtHandler ─────────────────────────────────────────────────────────────

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

// ─── BudgetHandler ───────────────────────────────────────────────────────────

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

// ─── ReconciliationHandler ───────────────────────────────────────────────────

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

// ─── ReportHandler ───────────────────────────────────────────────────────────

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

// ─── misc helpers ─────────────────────────────────────────────────────────────

func nilStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
