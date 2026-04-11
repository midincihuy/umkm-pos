package service

import (
	"github.com/google/uuid"
	"umkm-pos/internal/dto"
	"umkm-pos/internal/model"
	"umkm-pos/internal/repository"
	"umkm-pos/pkg/apperror"
	"gorm.io/gorm"
)

type TransactionService struct {
	txRepo       *repository.TransactionRepo
	accountRepo  *repository.AccountRepo
	categoryRepo *repository.CategoryRepo
	db           *gorm.DB
}

func NewTransactionService(txRepo *repository.TransactionRepo, accountRepo *repository.AccountRepo, categoryRepo *repository.CategoryRepo, db *gorm.DB) *TransactionService {
	return &TransactionService{txRepo: txRepo, accountRepo: accountRepo, categoryRepo: categoryRepo, db: db}
}

func (s *TransactionService) Create(userID uuid.UUID, req *dto.CreateTransactionReq) (*model.Transaction, error) {
	// Validasi kategori
	cat, err := s.categoryRepo.FindByID(req.CategoryID)
	if err != nil {
		return nil, apperror.NotFound("Kategori tidak ditemukan")
	}
	if cat.IsOthers && req.Notes == "" {
		return nil, apperror.BadRequest("Notes wajib diisi untuk kategori lain-lain")
	}

	var result *model.Transaction

	err = s.db.Transaction(func(tx *gorm.DB) error {
		account, err := s.accountRepo.FindByID(req.AccountID, userID)
		if err != nil {
			return apperror.NotFound("Rekening tidak ditemukan")
		}

		balanceImpact := req.Amount
		if req.Type == "expense" {
			if account.CurrentBalance < req.Amount {
				return apperror.InsufficientBalance("Saldo tidak mencukupi")
			}
			balanceImpact = -req.Amount
		}

		transaction := &model.Transaction{
			UserID:        userID,
			AccountID:     req.AccountID,
			CategoryID:    &req.CategoryID,
			Type:          model.TransactionType(req.Type),
			Amount:        req.Amount,
			BalanceImpact: balanceImpact,
			Date:          req.Date.Time(),
			Description:   req.Description,
			Notes:         req.Notes,
		}
		if err := s.txRepo.CreateTx(tx, transaction); err != nil {
			return err
		}
		if err := s.accountRepo.UpdateBalance(tx, req.AccountID, balanceImpact); err != nil {
			return err
		}
		result = transaction
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Preload relasi agar handler bisa akses Account.Name dan Category.Name
	return s.txRepo.FindByID(result.ID, userID)
}

func (s *TransactionService) Delete(id, userID uuid.UUID) error {
	tx, err := s.txRepo.FindByID(id, userID)
	if err != nil {
		return apperror.NotFound("Transaksi tidak ditemukan")
	}
	if tx.DebtID != nil {
		return apperror.Conflict("Hapus cicilan melalui endpoint /debts/{id}/payments")
	}

	return s.db.Transaction(func(dbTx *gorm.DB) error {
		// Rollback saldo
		if err := s.accountRepo.UpdateBalance(dbTx, tx.AccountID, -tx.BalanceImpact); err != nil {
			return err
		}
		return dbTx.Delete(&model.Transaction{}, "id = ?", id).Error
	})
}
func (s *TransactionService) List(userID uuid.UUID, f repository.TransactionFilter) ([]model.Transaction, int64, error) {
	return s.txRepo.List(userID, f)
}

func (s *TransactionService) GetByID(id, userID uuid.UUID) (*model.Transaction, error) {
	tx, err := s.txRepo.FindByID(id, userID)
	if err != nil {
		return nil, apperror.NotFound("Transaksi tidak ditemukan")
	}
	return tx, nil
}

func (s *TransactionService) Update(id, userID uuid.UUID, req *dto.UpdateTransactionReq) (*model.Transaction, error) {
	tx, err := s.txRepo.FindByID(id, userID)
	if err != nil {
		return nil, apperror.NotFound("Transaksi tidak ditemukan")
	}
	if tx.Type != model.TxIncome && tx.Type != model.TxExpense {
		return nil, apperror.BadRequest("Hanya transaksi income/expense yang dapat diubah")
	}

	err = s.db.Transaction(func(dbTx *gorm.DB) error {
		// Kumpulkan kolom yang akan diupdate secara eksplisit
		updates := map[string]interface{}{}

		if req.Amount != nil && *req.Amount != tx.Amount {
			oldImpact := tx.BalanceImpact
			newAmount := *req.Amount
			var newImpact float64
			if tx.Type == model.TxExpense {
				acct, err := s.accountRepo.FindByIDTx(dbTx, tx.AccountID)
				if err != nil {
					return err
				}
				available := acct.CurrentBalance - oldImpact
				if available < newAmount {
					return apperror.InsufficientBalance("Saldo tidak mencukupi untuk jumlah baru")
				}
				newImpact = -newAmount
			} else {
				newImpact = newAmount
			}
			if err := s.accountRepo.UpdateBalance(dbTx, tx.AccountID, -oldImpact); err != nil {
				return err
			}
			if err := s.accountRepo.UpdateBalance(dbTx, tx.AccountID, newImpact); err != nil {
				return err
			}
			updates["amount"] = newAmount
			updates["balance_impact"] = newImpact
			tx.Amount = newAmount
			tx.BalanceImpact = newImpact
		}

		// ── CategoryID ────────────────────────────────────────────────────────
		// FIX: gunakan update kolom eksplisit, bukan Save() yang bisa
		// tertimpa oleh preloaded association dari FindByID
		if req.CategoryID != nil {
			updates["category_id"] = req.CategoryID
			tx.CategoryID = req.CategoryID
		}

		// ── Date ──────────────────────────────────────────────────────────────
		if req.Date != nil {
			updates["date"] = req.Date
			tx.Date = *req.Date
		}
 
		// ── Description ───────────────────────────────────────────────────────
		if req.Description != nil {
			updates["description"] = req.Description
			tx.Description = *req.Description
		}
 
		// ── Notes ─────────────────────────────────────────────────────────────
		if req.Notes != nil {
			updates["notes"] = req.Notes
			tx.Notes = *req.Notes
		}

		if len(updates) == 0 {
			return nil // tidak ada yang perlu diupdate
		}

		return dbTx.Model(&model.Transaction{}).
			Where("id = ?", tx.ID).
			Updates(updates).Error
	})
	return tx, err
}
