package router

import (
	"github.com/gin-gonic/gin"
	"umkm-pos/internal/handler"
	"umkm-pos/internal/middleware"
	"umkm-pos/pkg/jwt"
)

type Handlers struct {
	Auth           *handler.AuthHandler
	Account        *handler.AccountHandler
	Category       *handler.CategoryHandler
	Transaction    *handler.TransactionHandler
	Transfer       *handler.TransferHandler
	Debt           *handler.DebtHandler
	Budget         *handler.BudgetHandler
	Reconciliation *handler.ReconciliationHandler
	Report         *handler.ReportHandler
	Spreadsheet    *handler.SpreadsheetHandler
}

func Setup(h Handlers, jwtHelper *jwt.JWT, allowedOrigin string) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORSMiddleware(allowedOrigin))
	r.Use(middleware.LoggerMiddleware())

	v1 := r.Group("/v1")

	// Semua route protected — login/register dihandle Supabase Auth di frontend
	api := v1.Group("")
	api.Use(middleware.AuthMiddleware(jwtHelper))
	{
		// Auth — hanya profil user, bukan login/register
		api.GET("/auth/me", h.Auth.Me)
		api.PATCH("/auth/me", h.Auth.UpdateMe)
		api.PATCH("/auth/me/password", h.Auth.ChangePassword)

		// Accounts
		api.GET("/accounts", h.Account.List)
		api.POST("/accounts", h.Account.Create)
		api.GET("/accounts/:id", h.Account.GetByID)
		api.PATCH("/accounts/:id", h.Account.Update)
		api.DELETE("/accounts/:id", h.Account.Delete)
		api.GET("/accounts/:id/transactions", func(c *gin.Context) {
			// inject account_id ke query param lalu forward ke Transaction.List
			q := c.Request.URL.Query()
			q.Set("account_id", c.Param("id"))
			c.Request.URL.RawQuery = q.Encode()
			h.Transaction.List(c)
		})

		// Categories
		api.GET("/categories", h.Category.List)
		api.POST("/categories", h.Category.Create)
		api.PATCH("/categories/:id", h.Category.Update)
		api.DELETE("/categories/:id", h.Category.Delete)

		// Transactions
		api.GET("/transactions", h.Transaction.List)
		api.POST("/transactions", h.Transaction.Create)
		api.GET("/transactions/:id", h.Transaction.GetByID)
		api.PATCH("/transactions/:id", h.Transaction.Update)
		api.DELETE("/transactions/:id", h.Transaction.Delete)

		// Transfers
		api.GET("/transfers", h.Transfer.List)
		api.POST("/transfers", h.Transfer.Create)
		api.GET("/transfers/:id", h.Transfer.GetByID)
		api.DELETE("/transfers/:id", h.Transfer.Delete)

		// Debts
		api.GET("/debts", h.Debt.List)
		api.POST("/debts", h.Debt.Create)
		api.GET("/debts/:id", h.Debt.GetByID)
		api.PATCH("/debts/:id", h.Debt.Update)
		api.POST("/debts/:id/cancel", h.Debt.Cancel)
		api.POST("/debts/:id/payments", h.Debt.CreatePayment)
		api.DELETE("/debts/:id/payments/:payment_id", h.Debt.DeletePayment)

		// Budgets
		api.GET("/budgets", h.Budget.Get)
		api.POST("/budgets", h.Budget.Upsert)
		api.POST("/budgets/copy", h.Budget.Copy)
		api.DELETE("/budgets/:id", h.Budget.Delete)

		// Reconciliation
		api.POST("/reconciliations/preview", h.Reconciliation.Preview)
		api.GET("/reconciliations", h.Reconciliation.List)
		api.POST("/reconciliations", h.Reconciliation.Create)

		// Reports
		api.GET("/reports/summary", h.Report.Summary)
		api.GET("/reports/expense-breakdown", h.Report.ExpenseBreakdown)
		api.GET("/reports/income-breakdown", h.Report.IncomeBreakdown)
		api.GET("/reports/trend", h.Report.Trend)
		api.GET("/reports/net-worth", h.Report.NetWorth)

		// Spreadsheet
		sheetGroup := api.Group("/spreadsheet")
		{
			sheetGroup.POST("/read", h.Spreadsheet.ReadData)
			sheetGroup.POST("/write", h.Spreadsheet.WriteData)
			sheetGroup.POST("/append", h.Spreadsheet.AppendData)
			sheetGroup.POST("/update-cell", h.Spreadsheet.UpdateCell)
		}
	}

	return r
}