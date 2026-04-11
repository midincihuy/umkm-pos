package main

import (
	"log"

	"umkm-pos/config"
	"umkm-pos/internal/handler"
	"umkm-pos/internal/repository"
	"umkm-pos/internal/router"
	"umkm-pos/internal/service"
	"umkm-pos/pkg/jwt"
)

func main() {
	// 1. Load config dari .env
	cfg := config.Load()

	// 2. Inisialisasi koneksi database
	db := config.InitDB(cfg)

	// 3. JWT helper — hanya untuk validasi token Supabase
	jwtHelper, err := jwt.NewJWT(cfg.SupabaseJWKSUrl)

	// 4. Wire: repo → service → handler
	if err != nil {
		log.Fatal(err)
	}
	// Repository layer
	userRepo        := repository.NewUserRepo(db)
	accountRepo     := repository.NewAccountRepo(db)
	categoryRepo    := repository.NewCategoryRepo(db)
	transactionRepo := repository.NewTransactionRepo(db)
	transferRepo    := repository.NewTransferRepo(db)
	debtRepo        := repository.NewDebtRepo(db)
	budgetRepo      := repository.NewBudgetRepo(db)
	reconRepo       := repository.NewReconRepo(db)
	reportRepo      := repository.NewReportRepo(db)

	// Service layer
	// Auth service hanya untuk GET/PATCH profil — login/register dihandle Supabase
	authService     := service.NewAuthService(userRepo)
	accountService  := service.NewAccountService(accountRepo, transactionRepo, db)
	categoryService := service.NewCategoryService(categoryRepo)
	txService       := service.NewTransactionService(transactionRepo, accountRepo, categoryRepo, db)
	transferService := service.NewTransferService(transferRepo, accountRepo, categoryRepo, db)
	debtService     := service.NewDebtService(debtRepo, accountRepo, transactionRepo, db)
	budgetService   := service.NewBudgetService(budgetRepo, transactionRepo)
	reconService    := service.NewReconService(reconRepo, accountRepo, transactionRepo, db)
	reportService   := service.NewReportService(reportRepo, accountRepo, debtRepo)

	// Handler layer
	authHandler     := handler.NewAuthHandler(authService)
	accountHandler  := handler.NewAccountHandler(accountService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	txHandler       := handler.NewTransactionHandler(txService)
	transferHandler := handler.NewTransferHandler(transferService)
	debtHandler     := handler.NewDebtHandler(debtService)
	budgetHandler   := handler.NewBudgetHandler(budgetService)
	reconHandler    := handler.NewReconHandler(reconService)
	reportHandler   := handler.NewReportHandler(reportService)

	// 5. Setup router & jalankan server
	r := router.Setup(router.Handlers{
		Auth:           authHandler,
		Account:        accountHandler,
		Category:       categoryHandler,
		Transaction:    txHandler,
		Transfer:       transferHandler,
		Debt:           debtHandler,
		Budget:         budgetHandler,
		Reconciliation: reconHandler,
		Report:         reportHandler,
	}, jwtHelper, cfg.AllowedOrigin)

	log.Printf("Server running on :%s (env: %s)", cfg.Port, cfg.Env)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
