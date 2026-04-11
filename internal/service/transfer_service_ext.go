package service

import (
	"github.com/google/uuid"
	"umkm-pos/internal/dto"
	"umkm-pos/internal/model"
	"umkm-pos/internal/repository"
	"umkm-pos/pkg/apperror"
	"gorm.io/gorm"
)

type TransferService struct {
	transferRepo *repository.TransferRepo
	accountRepo  *repository.AccountRepo
	categoryRepo *repository.CategoryRepo
	db           *gorm.DB
}

func NewTransferService(transferRepo *repository.TransferRepo, accountRepo *repository.AccountRepo, categoryRepo *repository.CategoryRepo, db *gorm.DB) *TransferService {
	return &TransferService{transferRepo: transferRepo, accountRepo: accountRepo, categoryRepo: categoryRepo, db: db}
}

func (s *TransferService) Create(userID uuid.UUID, req *dto.CreateTransferReq) (*model.Transfer, error) {
	if req.FromAccountID == req.ToAccountID {
		return nil, apperror.BadRequest("Rekening asal dan tujuan tidak boleh sama")
	}

	var result *model.Transfer

	err := s.db.Transaction(func(tx *gorm.DB) error {
		from, err := s.accountRepo.FindByID(req.FromAccountID, userID)
		if err != nil {
			return apperror.NotFound("Rekening asal tidak ditemukan")
		}
		if from.CurrentBalance < (req.Amount + req.Fee) {
			return apperror.InsufficientBalance("Saldo tidak mencukupi")
		}
		to, err := s.accountRepo.FindByID(req.ToAccountID, userID)
		if err != nil {
			return apperror.NotFound("Rekening tujuan tidak ditemukan")
		}

		transfer := &model.Transfer{
			UserID:        userID,
			FromAccountID: req.FromAccountID,
			ToAccountID:   req.ToAccountID,
			Amount:        req.Amount,
			Fee:           req.Fee,
			Date:          req.Date.Time(),
			Notes:         req.Notes,
		}
		if err := tx.Create(transfer).Error; err != nil {
			return err
		}

		// Kurangi saldo rekening asal (amount + fee)
		if err := s.accountRepo.UpdateBalance(tx, req.FromAccountID, -(req.Amount + req.Fee)); err != nil {
			return err
		}
		// Tambah saldo rekening tujuan
		if err := s.accountRepo.UpdateBalance(tx, req.ToAccountID, req.Amount); err != nil {
			return err
		}

		// Catat 2 transaksi transfer
		fromTx := &model.Transaction{
			UserID:         userID,
			AccountID:      req.FromAccountID,
			TransferPairID: &transfer.ID,
			Type:           model.TxTransfer,
			Amount:         req.Amount,
			BalanceImpact:  -(req.Amount + req.Fee),
			Date:           req.Date.Time(),
			Description:    "Transfer ke " + to.Name,
			Notes:          req.Notes,
		}
		toTx := &model.Transaction{
			UserID:         userID,
			AccountID:      req.ToAccountID,
			TransferPairID: &transfer.ID,
			Type:           model.TxTransfer,
			Amount:         req.Amount,
			BalanceImpact:  req.Amount,
			Date:           req.Date.Time(),
			Description:    "Transfer dari " + from.Name,
			Notes:          req.Notes,
		}
		if err := tx.Create(fromTx).Error; err != nil {
			return err
		}
		if err := tx.Create(toTx).Error; err != nil {
			return err
		}

		result = transfer
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Preload FromAccount dan ToAccount agar handler bisa akses nama rekening
	return s.transferRepo.FindByID(result.ID, userID)
}

func (s *TransferService) List(userID uuid.UUID, f repository.TransferFilter) ([]model.Transfer, int64, error) {
	return s.transferRepo.List(userID, f)
}

func (s *TransferService) GetByID(id, userID uuid.UUID) (*model.Transfer, error) {
	t, err := s.transferRepo.FindByID(id, userID)
	if err != nil {
		return nil, apperror.NotFound("Transfer tidak ditemukan")
	}
	return t, nil
}

// Delete rollback saldo kedua rekening dan hapus transaksi terkait
func (s *TransferService) Delete(id, userID uuid.UUID) error {
	transfer, err := s.transferRepo.FindByID(id, userID)
	if err != nil {
		return apperror.NotFound("Transfer tidak ditemukan")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Rollback saldo: kembalikan dari sumber, kurangi dari tujuan
		if err := s.accountRepo.UpdateBalance(tx, transfer.FromAccountID, transfer.Amount+transfer.Fee); err != nil {
			return err
		}
		if err := s.accountRepo.UpdateBalance(tx, transfer.ToAccountID, -transfer.Amount); err != nil {
			return err
		}
		// Hapus transaksi yang terkait
		if err := tx.Where("transfer_pair_id = ?", id).Delete(&model.Transaction{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Transfer{}, "id = ?", id).Error
	})
}
