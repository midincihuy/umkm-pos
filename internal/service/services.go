package service

// Petunjuk implementasi:
// - Setiap service menerima repo + db sebagai dependency
// - Operasi yang menyentuh >1 tabel wajib dibungkus db.Transaction()
// - Error dikembalikan menggunakan pkg/apperror
// File ini berisi struct + constructor. Implementasikan method di file terpisah per service.

import (
	"time"

	"github.com/google/uuid"
	"umkm-pos/internal/dto"
	"umkm-pos/internal/model"
	"umkm-pos/internal/repository"
	"umkm-pos/pkg/apperror"
	"gorm.io/gorm"
)

// ─── AuthService ─────────────────────────────────────────────────────────────
// Login, register, Google OAuth, dan refresh token sepenuhnya dihandle Supabase Auth.
// Service ini hanya mengurus profil user di tabel public.users.

type AuthService struct {
	userRepo *repository.UserRepo
}

func NewAuthService(userRepo *repository.UserRepo) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) GetByID(userID uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, apperror.NotFound("User tidak ditemukan")
	}
	return user, nil
}

func (s *AuthService) UpdateUser(userID uuid.UUID, req *dto.UpdateUserReq) (*model.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, apperror.NotFound("User tidak ditemukan")
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Currency != "" {
		user.Currency = req.Currency
	}
	return user, s.userRepo.Update(user)
}

// ChangePassword tidak relevan lagi — password dikelola Supabase Auth.
// Gunakan Supabase SDK di frontend: supabase.auth.updateUser({password: newPassword})
func (s *AuthService) ChangePassword(_ uuid.UUID, _ *dto.ChangePasswordReq) error {
	return apperror.BadRequest("Ganti password melalui Supabase Auth di frontend")
}

// ─── AccountService ──────────────────────────────────────────────────────────

type AccountService struct {
	accountRepo *repository.AccountRepo
	txRepo      *repository.TransactionRepo
	db          *gorm.DB
}

func NewAccountService(accountRepo *repository.AccountRepo, txRepo *repository.TransactionRepo, db *gorm.DB) *AccountService {
	return &AccountService{accountRepo: accountRepo, txRepo: txRepo, db: db}
}

func (s *AccountService) List(userID uuid.UUID, activeOnly bool) (*dto.AccountListResp, error) {
	accounts, err := s.accountRepo.FindByUserID(userID, activeOnly)
	if err != nil {
		return nil, err
	}
	totalBalance, _ := s.accountRepo.SumBalance(userID)

	var resp []dto.AccountResp
	for _, a := range accounts {
		resp = append(resp, toAccountResp(a))
	}
	return &dto.AccountListResp{Data: resp, TotalBalance: totalBalance}, nil
}

func (s *AccountService) Create(userID uuid.UUID, req *dto.CreateAccountReq) (*dto.AccountResp, error) {
	var result *model.Account
	err := s.db.Transaction(func(tx *gorm.DB) error {
		account := &model.Account{
			UserID:         userID,
			Name:           req.Name,
			Type:           model.AccountType(req.Type),
			OpeningBalance: req.OpeningBalance,
			CurrentBalance: req.OpeningBalance,
			Icon:           req.Icon,
			Color:          req.Color,
		}
		if err := tx.Create(account).Error; err != nil {
			return err
		}
		// Catat opening balance sebagai transaksi
		if req.OpeningBalance > 0 {
			openingTx := &model.Transaction{
				UserID:        userID,
				AccountID:     account.ID,
				Type:          model.TxOpeningBalance,
				Amount:        req.OpeningBalance,
				BalanceImpact: req.OpeningBalance,
				Date:          time.Now(),
				Description:   "Saldo awal " + req.Name,
			}
			if err := tx.Create(openingTx).Error; err != nil {
				return err
			}
		}
		result = account
		return nil
	})
	if err != nil {
		return nil, err
	}
	r := toAccountResp(*result)
	return &r, nil
}

func (s *AccountService) GetByID(id, userID uuid.UUID) (*dto.AccountResp, error) {
	account, err := s.accountRepo.FindByID(id, userID)
	if err != nil {
		return nil, apperror.NotFound("Rekening tidak ditemukan")
	}
	r := toAccountResp(*account)
	return &r, nil
}

func (s *AccountService) Update(id, userID uuid.UUID, req *dto.UpdateAccountReq) (*dto.AccountResp, error) {
	account, err := s.accountRepo.FindByID(id, userID)
	if err != nil {
		return nil, apperror.NotFound("Rekening tidak ditemukan")
	}
	if req.Name != nil {
		account.Name = *req.Name
	}
	if req.Icon != nil {
		account.Icon = *req.Icon
	}
	if req.Color != nil {
		account.Color = *req.Color
	}
	if req.IsActive != nil {
		account.IsActive = *req.IsActive
	}
	if err := s.accountRepo.Update(account); err != nil {
		return nil, err
	}
	r := toAccountResp(*account)
	return &r, nil
}

func (s *AccountService) Delete(id, userID uuid.UUID) error {
	account, err := s.accountRepo.FindByID(id, userID)
	if err != nil {
		return apperror.NotFound("Rekening tidak ditemukan")
	}
	hasTransactions, _ := s.accountRepo.HasTransactions(id)
	if hasTransactions && account.CurrentBalance != 0 {
		// Soft delete — nonaktifkan saja
		account.IsActive = false
		return s.accountRepo.Update(account)
	}
	account.IsActive = false
	return s.accountRepo.Update(account)
}

func toAccountResp(a model.Account) dto.AccountResp {
	return dto.AccountResp{
		ID:             a.ID,
		Name:           a.Name,
		Type:           string(a.Type),
		OpeningBalance: a.OpeningBalance,
		CurrentBalance: a.CurrentBalance,
		Icon:           a.Icon,
		Color:          a.Color,
		IsActive:       a.IsActive,
		CreatedAt:      a.CreatedAt,
	}
}

func toTransactionResp(a model.Transaction) dto.TransactionResp {
	var categoryName *string
	if a.Category != nil {
		categoryName = &a.Category.Name
	}
	return dto.TransactionResp{
		ID:             a.ID,
		AccountID:             a.AccountID,
		AccountName:           a.Account.Name,
		CategoryID:             a.CategoryID,
		CategoryName:           categoryName,
		Type:           string(a.Type),
		Amount: a.Amount,
		BalanceImpact: a.BalanceImpact,
		Date: a.Date.Format("2006-01-02"),
		Description: a.Description,
		Notes:	a.Notes,
		CreatedAt:      a.CreatedAt,
	}
}

// ─── CategoryService ─────────────────────────────────────────────────────────

type CategoryService struct {
	categoryRepo *repository.CategoryRepo
}

func NewCategoryService(categoryRepo *repository.CategoryRepo) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

func (s *CategoryService) List(userID uuid.UUID, catType *string, includeInactive bool) (*dto.CategoryListResp, error) {
	cats, err := s.categoryRepo.FindAll(userID, includeInactive)
	if err != nil {
		return nil, err
	}
	var income, expense []dto.CategoryResp
	for _, c := range cats {
		r := dto.CategoryResp{
			ID:        c.ID,
			Name:      c.Name,
			Type:      string(c.Type),
			Icon:      c.Icon,
			Color:     c.Color,
			IsSystem:  c.IsSystem,
			IsActive:  c.IsActive,
			SortOrder: c.SortOrder,
		}
		if c.Type == model.CategoryIncome {
			income = append(income, r)
		} else {
			expense = append(expense, r)
		}
	}
	return &dto.CategoryListResp{Income: income, Expense: expense}, nil
}

func (s *CategoryService) Create(userID uuid.UUID, req *dto.CreateCategoryReq) (*dto.CategoryResp, error) {
	cat := &model.Category{
		UserID: &userID,
		Name:   req.Name,
		Type:   model.CategoryType(req.Type),
		Icon:   req.Icon,
		Color:  req.Color,
	}
	if err := s.categoryRepo.Create(cat); err != nil {
		return nil, err
	}
	return &dto.CategoryResp{
		ID:       cat.ID,
		Name:     cat.Name,
		Type:     string(cat.Type),
		Icon:     cat.Icon,
		Color:    cat.Color,
		IsSystem: false,
		IsActive: true,
	}, nil
}

func (s *CategoryService) Update(id uuid.UUID, req *dto.CreateCategoryReq) (*dto.CategoryResp, error) {
	cat, err := s.categoryRepo.FindByID(id)
	if err != nil {
		return nil, apperror.NotFound("Kategori tidak ditemukan")
	}
	if cat.IsSystem {
		return nil, apperror.New(apperror.ErrSystemCategory, "Kategori sistem tidak dapat diedit")
	}
	cat.Name = req.Name
	cat.Icon = req.Icon
	cat.Color = req.Color
	if err := s.categoryRepo.Update(cat); err != nil {
		return nil, err
	}
	return &dto.CategoryResp{
		ID:       cat.ID,
		Name:     cat.Name,
		Type:     string(cat.Type),
		Icon:     cat.Icon,
		Color:    cat.Color,
		IsSystem: cat.IsSystem,
		IsActive: cat.IsActive,
	}, nil
}

func (s *CategoryService) Delete(id uuid.UUID) error {
	cat, err := s.categoryRepo.FindByID(id)
	if err != nil {
		return apperror.NotFound("Kategori tidak ditemukan")
	}
	if cat.IsSystem {
		return apperror.New(apperror.ErrSystemCategory, "Kategori sistem tidak dapat dihapus")
	}
	isUsed, _ := s.categoryRepo.IsUsed(id)
	if isUsed {
		return apperror.Conflict("Kategori masih digunakan oleh transaksi atau budget")
	}
	cat.IsActive = false
	return s.categoryRepo.Update(cat)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

