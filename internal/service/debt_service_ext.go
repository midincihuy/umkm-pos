package service

import (
	"time"
	"github.com/google/uuid"
	"umkm-pos/internal/dto"
	"umkm-pos/internal/model"
	"umkm-pos/internal/repository"
	"umkm-pos/pkg/apperror"
	"gorm.io/gorm"
)

type DebtService struct {
	debtRepo    *repository.DebtRepo
	accountRepo *repository.AccountRepo
	txRepo      *repository.TransactionRepo
	db          *gorm.DB
}

func NewDebtService(debtRepo *repository.DebtRepo, accountRepo *repository.AccountRepo, txRepo *repository.TransactionRepo, db *gorm.DB) *DebtService {
	return &DebtService{debtRepo: debtRepo, accountRepo: accountRepo, txRepo: txRepo, db: db}
}

func (s *DebtService) Create(userID uuid.UUID, req *dto.CreateDebtReq) (*model.Debt, error) {
	var result *model.Debt
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if req.Type == "receivable" {
			account, err := s.accountRepo.FindByID(req.AccountID, userID)
			if err != nil {
				return apperror.NotFound("Rekening tidak ditemukan")
			}
			if account.CurrentBalance < req.TotalAmount {
				return apperror.InsufficientBalance("Saldo tidak mencukupi untuk memberikan pinjaman")
			}
			if err := s.accountRepo.UpdateBalance(tx, req.AccountID, -req.TotalAmount); err != nil {
				return err
			}
		} else {
			// debt: terima pinjaman → saldo bertambah
			if err := s.accountRepo.UpdateBalance(tx, req.AccountID, req.TotalAmount); err != nil {
				return err
			}
		}
		var due *time.Time
		if req.DueDate != nil {
			t := req.DueDate.Time()
			due = &t
		}
		debt := &model.Debt{
			UserID:      userID,
			Type:        model.DebtType(req.Type),
			ContactName: req.ContactName,
			TotalAmount: req.TotalAmount,
			StartDate:   req.StartDate.Time(),
			DueDate:     due,
			Notes:       req.Notes,
		}
		if err := tx.Create(debt).Error; err != nil {
			return err
		}
		result = debt
		return nil
	})
	return result, err
}

func (s *DebtService) CreatePayment(debtID, userID uuid.UUID, req *dto.CreateDebtPaymentReq) (*dto.CreateDebtPaymentResp, error) {
	debt, err := s.debtRepo.FindByID(debtID, userID)
	if err != nil {
		return nil, apperror.NotFound("Utang/piutang tidak ditemukan")
	}
	if debt.Status != model.DebtStatusActive {
		return nil, apperror.New(apperror.ErrDebtNotActive, "Utang/piutang sudah lunas atau dibatalkan")
	}
	if req.Amount > debt.RemainingAmount() {
		return nil, apperror.New(apperror.ErrPaymentExceedsDebt, "Nominal melebihi sisa kewajiban")
	}

	var paymentResult *model.DebtPayment

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Validasi saldo untuk debt payment
		if debt.Type == model.DebtTypeDebt {
			account, err := s.accountRepo.FindByID(req.AccountID, userID)
			if err != nil {
				return apperror.NotFound("Rekening tidak ditemukan")
			}
			if account.CurrentBalance < req.Amount {
				return apperror.InsufficientBalance("Saldo tidak mencukupi")
			}
			if err := s.accountRepo.UpdateBalance(tx, req.AccountID, -req.Amount); err != nil {
				return err
			}
		} else {
			// receivable: terima cicilan → saldo bertambah
			if err := s.accountRepo.UpdateBalance(tx, req.AccountID, req.Amount); err != nil {
				return err
			}
		}

		payment := &model.DebtPayment{
			DebtID:    debtID,
			AccountID: req.AccountID,
			Amount:    req.Amount,
			Date:      req.Date.Time(),
			Notes:     req.Notes,
		}
		if err := s.debtRepo.CreatePaymentTx(tx, payment); err != nil {
			return err
		}

		debt.PaidAmount += req.Amount
		if debt.PaidAmount >= debt.TotalAmount {
			debt.Status = model.DebtStatusPaid
		}
		if err := s.debtRepo.UpdateTx(tx, debt); err != nil {
			return err
		}

		paymentResult = payment
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &dto.CreateDebtPaymentResp{
		Payment:     toDebtPaymentResp(*paymentResult),
		Debt:        toDebtResp(*debt),
		IsFullyPaid: debt.Status == model.DebtStatusPaid,
	}, nil
}

func toDebtPaymentResp(p model.DebtPayment) dto.DebtPaymentResp {
	return dto.DebtPaymentResp{
		ID:        p.ID,
		DebtID:    p.DebtID,
		AccountID: p.AccountID,
		Amount:    p.Amount,
		Date:      p.Date.Format("2006-01-02"),
		Notes:     p.Notes,
		CreatedAt: p.CreatedAt,
	}
}

func toDebtResp(d model.Debt) dto.DebtResp {
	resp := dto.DebtResp{
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
		resp.DueDate = &s
	}
	return resp
}

func (s *DebtService) List(userID uuid.UUID, f repository.DebtFilter) ([]model.Debt, int64, error) {
	return s.debtRepo.List(userID, f)
}

func (s *DebtService) GetByID(id, userID uuid.UUID) (*dto.DebtResp, error) {
	d, err := s.debtRepo.FindByID(id, userID)
	if err != nil {
		return nil, apperror.NotFound("Utang/piutang tidak ditemukan")
	}
	resp := toDebtResp(*d)
	return &resp, nil
}

func (s *DebtService) Update(id, userID uuid.UUID, req *dto.UpdateDebtReq) (*dto.DebtResp, error) {
	d, err := s.debtRepo.FindByID(id, userID)
	if err != nil {
		return nil, apperror.NotFound("Utang/piutang tidak ditemukan")
	}
	if req.ContactName != nil { d.ContactName = *req.ContactName }
	if req.DueDate != nil     { d.DueDate = req.DueDate }
	if req.Notes != nil       { d.Notes = *req.Notes }
	if err := s.debtRepo.Update(d); err != nil {
		return nil, err
	}
	resp := toDebtResp(*d)
	return &resp, nil
}

func (s *DebtService) Cancel(id, userID uuid.UUID) error {
	d, err := s.debtRepo.FindByID(id, userID)
	if err != nil {
		return apperror.NotFound("Utang/piutang tidak ditemukan")
	}
	if d.Status != model.DebtStatusActive {
		return apperror.New(apperror.ErrDebtNotActive, "Hanya utang/piutang aktif yang dapat dibatalkan")
	}
	d.Status = model.DebtStatusCancelled
	err = s.debtRepo.Update(d)
	if err != nil {
		return err
	}
	return nil
}
 // ganti disini
// DeletePayment rollback saldo rekening dan kurangi paid_amount
func (s *DebtService) DeletePayment(paymentID, debtID, userID uuid.UUID) error {
	debt, err := s.debtRepo.FindByID(debtID, userID)
	if err != nil {
		return apperror.NotFound("Utang/piutang tidak ditemukan")
	}

	payment, err := s.debtRepo.FindPaymentByID(paymentID)
	if err != nil {
		return apperror.NotFound("Cicilan tidak ditemukan")
	}
	if payment.DebtID != debtID {
		return apperror.Forbidden("Akses ditolak")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Rollback saldo
		if debt.Type == model.DebtTypeDebt {
			// Bayar cicilan hutang → saldo berkurang → rollback: tambah balik
			if err := s.accountRepo.UpdateBalance(tx, payment.AccountID, payment.Amount); err != nil {
				return err
			}
		} else {
			// Terima cicilan piutang → saldo bertambah → rollback: kurangi
			if err := s.accountRepo.UpdateBalance(tx, payment.AccountID, -payment.Amount); err != nil {
				return err
			}
		}
		// Update paid_amount
		debt.PaidAmount -= payment.Amount
		if debt.PaidAmount < 0 {
			debt.PaidAmount = 0
		}
		if debt.Status == model.DebtStatusPaid {
			debt.Status = model.DebtStatusActive
		}
		if err := s.debtRepo.UpdateTx(tx, debt); err != nil {
			return err
		}
		return s.debtRepo.DeletePayment(tx, paymentID)
	})
}
