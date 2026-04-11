package service

import (
	"github.com/google/uuid"
	"umkm-pos/internal/dto"
	"umkm-pos/internal/model"
	"umkm-pos/internal/repository"
	"umkm-pos/pkg/apperror"
	"gorm.io/gorm"
)
type ReconService struct {
	reconRepo   *repository.ReconRepo
	accountRepo *repository.AccountRepo
	txRepo      *repository.TransactionRepo
	db          *gorm.DB
}

func NewReconService(reconRepo *repository.ReconRepo, accountRepo *repository.AccountRepo, txRepo *repository.TransactionRepo, db *gorm.DB) *ReconService {
	return &ReconService{reconRepo: reconRepo, accountRepo: accountRepo, txRepo: txRepo, db: db}
}

func (s *ReconService) List(userID uuid.UUID, accountID *uuid.UUID) ([]dto.CashAdjustmentResp, error) {
	list, err := s.reconRepo.FindByUserID(userID, accountID)
	if err != nil {
		return nil, err
	}
	var result []dto.CashAdjustmentResp
	for _, ca := range list {
		result = append(result, toCashAdjResp(ca))
	}
	return result, nil
}

func toCashAdjResp(ca model.CashAdjustment) dto.CashAdjustmentResp {
	r := dto.CashAdjustmentResp{
		ID:            ca.ID,
		AccountID:     ca.AccountID,
		BalanceBefore: ca.BalanceBefore,
		ActualBalance: ca.ActualBalance,
		Difference:    ca.Difference,
		Date:          ca.Date.Format("2006-01-02"),
		Notes:         ca.Notes,
		CreatedAt:     ca.CreatedAt,
	}
	if ca.Account.ID != uuid.Nil {
		r.AccountName = ca.Account.Name
	}
	return r
}

func (s *ReconService) Preview(userID uuid.UUID, req *dto.ReconciliationPreviewReq) (*dto.ReconciliationPreviewResp, error) {
	account, err := s.accountRepo.FindByID(req.AccountID, userID)
	if err != nil {
		return nil, apperror.NotFound("Rekening tidak ditemukan")
	}
	difference := req.ActualBalance - account.CurrentBalance
	resp := &dto.ReconciliationPreviewResp{
		AccountID:       account.ID,
		AccountName:     account.Name,
		BalanceBefore:   account.CurrentBalance,
		ActualBalance:   req.ActualBalance,
		Difference:      difference,
		NeedsAdjustment: difference != 0,
	}
	// Warning jika selisih > 20%
	if account.CurrentBalance > 0 {
		pct := (difference / account.CurrentBalance) * 100
		if pct > 20 || pct < -20 {
			msg := "Selisih cukup besar (>20%), pastikan inputan sudah benar"
			resp.Warning = &msg
		}
	}
	return resp, nil
}

func (s *ReconService) Create(userID uuid.UUID, req *dto.CreateReconciliationReq) (*dto.CashAdjustmentResp, error) {
	account, err := s.accountRepo.FindByID(req.AccountID, userID)
	if err != nil {
		return nil, apperror.NotFound("Rekening tidak ditemukan")
	}
	if req.ActualBalance < 0 {
		return nil, apperror.BadRequest("Saldo tidak boleh negatif")
	}

	difference := req.ActualBalance - account.CurrentBalance
	var result *model.CashAdjustment

	err = s.db.Transaction(func(tx *gorm.DB) error {
		adjustment := &model.CashAdjustment{
			UserID:        userID,
			AccountID:     req.AccountID,
			BalanceBefore: account.CurrentBalance,
			ActualBalance: req.ActualBalance,
			Difference:    difference,
			Date:          req.Date.Time(),
			Notes:         req.Notes,
		}
		if err := s.reconRepo.CreateTx(tx, adjustment); err != nil {
			return err
		}

		if err := s.accountRepo.SetBalance(tx, req.AccountID, req.ActualBalance); err != nil {
			return err
		}

		// Catat ke transactions jika ada selisih
		if difference != 0 {
			txType := model.TxAdjustment
			adjTx := &model.Transaction{
				UserID:        userID,
				AccountID:     req.AccountID,
				Type:          txType,
				Amount:        abs(difference),
				BalanceImpact: difference,
				Date:          req.Date.Time(),
				Description:   "Penyesuaian kas",
				Notes:         req.Notes,
			}
			if err := tx.Create(adjTx).Error; err != nil {
				return err
			}
		}

		result = adjustment
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &dto.CashAdjustmentResp{
		ID:            result.ID,
		AccountID:     result.AccountID,
		AccountName:   account.Name,
		BalanceBefore: result.BalanceBefore,
		ActualBalance: result.ActualBalance,
		Difference:    result.Difference,
		Date:          result.Date.Format("2006-01-02"),
		Notes:         result.Notes,
		CreatedAt:     result.CreatedAt,
	}, nil
}