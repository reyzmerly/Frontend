package routes

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kwizera/university-invoices/internal/config"
	"github.com/kwizera/university-invoices/internal/middleware"
	admincontrollers "github.com/kwizera/university-invoices/services/admin/controllers"
	authcontrollers "github.com/kwizera/university-invoices/services/auth/controllers"
	feecontrollers "github.com/kwizera/university-invoices/services/fee/controllers"
	financecontrollers "github.com/kwizera/university-invoices/services/finance/controllers"
	invoicecontrollers "github.com/kwizera/university-invoices/services/invoice/controllers"
	paymentcontrollers "github.com/kwizera/university-invoices/services/payment/controllers"
	receiptcontrollers "github.com/kwizera/university-invoices/services/receipt/controllers"
	studentcontrollers "github.com/kwizera/university-invoices/services/student/controllers"
	webhookcontrollers "github.com/kwizera/university-invoices/services/webhook/controllers"
)

func SetupRouter(db *sql.DB, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)

	r.Use(middleware.CORS)

	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Post("/auth/login", authcontrollers.NewAuthController(db, cfg.JWTSecret).Login)
			r.Post("/auth/forgot-password", authcontrollers.NewAuthController(db, cfg.JWTSecret).ForgotPassword)
		})

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Use(middleware.RequireRole("student", "finance", "admin"))

			r.Post("/auth/logout", authcontrollers.NewAuthController(db, cfg.JWTSecret).Logout)
			r.Get("/auth/me", authcontrollers.NewAuthController(db, cfg.JWTSecret).GetMe)

			r.Route("/student", func(r chi.Router) {
				r.Use(middleware.RequireRole("student"))
				r.Get("/dashboard", studentcontrollers.NewStudentController(db).GetDashboard)
			})

			r.Route("/invoices", func(r chi.Router) {
				r.Use(middleware.RequireRole("student", "finance"))
				r.Get("/", invoicecontrollers.NewInvoiceController(db).ListInvoices)
				r.Post("/", invoicecontrollers.NewInvoiceController(db).CreateInvoice)
				r.Route("/{invoiceId}", func(r chi.Router) {
					r.Get("/", invoicecontrollers.NewInvoiceController(db).GetInvoice)
					r.Post("/issue", invoicecontrollers.NewInvoiceController(db).IssueInvoice)
					r.Put("/", invoicecontrollers.NewInvoiceController(db).UpdateInvoice)
					r.Post("/payments", paymentcontrollers.NewPaymentController(db, cfg).CreatePayment)
				})
			})

			r.Route("/payments", func(r chi.Router) {
				r.Use(middleware.RequireRole("student", "finance"))
				r.Get("/", paymentcontrollers.NewPaymentController(db, cfg).ListPayments)
				r.Route("/{paymentId}", func(r chi.Router) {
					r.Get("/", paymentcontrollers.NewPaymentController(db, cfg).GetPayment)
					r.Post("/cancel", paymentcontrollers.NewPaymentController(db, cfg).CancelPayment)
					r.Get("/notes", paymentcontrollers.NewPaymentController(db, cfg).ListPaymentNotes)
					r.Post("/notes", paymentcontrollers.NewPaymentController(db, cfg).CreatePaymentNote)
				})
			})

			r.Route("/receipts", func(r chi.Router) {
				r.Use(middleware.RequireRole("student", "finance"))
				r.Route("/{receiptId}", func(r chi.Router) {
					r.Get("/", receiptcontrollers.NewReceiptController(db).GetReceipt)
					r.Get("/pdf", receiptcontrollers.NewReceiptController(db).GetReceiptPDF)
				})
				r.Get("/payment/{paymentId}", receiptcontrollers.NewReceiptController(db).GetReceiptByPaymentID)
			})

			r.Route("/finance", func(r chi.Router) {
				r.Use(middleware.RequireRole("finance"))
				r.Get("/dashboard", financecontrollers.NewFinanceController(db).GetDashboard)
				r.Get("/reports/summary", financecontrollers.NewFinanceController(db).GetReportSummary)
				r.Get("/reports/accounts", financecontrollers.NewFinanceController(db).GetAccountsReport)
				r.Get("/reports/export", financecontrollers.NewFinanceController(db).ExportReport)
				r.Get("/students", financecontrollers.NewFinanceController(db).SearchStudents)
				r.Get("/programmes", financecontrollers.NewFinanceController(db).ListProgrammes)
				r.Get("/semesters", financecontrollers.NewFinanceController(db).ListSemesters)
			})

			r.Route("/fee-types", func(r chi.Router) {
				r.Use(middleware.RequireRole("finance"))
				r.Get("/", feecontrollers.NewFeeController(db).ListFeeTypes)
				r.Post("/", feecontrollers.NewFeeController(db).CreateFeeType)
				r.Route("/{feeTypeId}", func(r chi.Router) {
					r.Get("/", feecontrollers.NewFeeController(db).GetFeeType)
					r.Put("/", feecontrollers.NewFeeController(db).UpdateFeeType)
				})
			})

			r.Route("/admin", func(r chi.Router) {
				r.Use(middleware.RequireRole("admin"))
				r.Get("/users", admincontrollers.NewAdminController(db).ListUsers)
				r.Post("/users", admincontrollers.NewAdminController(db).CreateUser)
				r.Route("/users/{userId}", func(r chi.Router) {
					r.Patch("/access", admincontrollers.NewAdminController(db).UpdateUserAccess)
				})
				r.Get("/roles", admincontrollers.NewAdminController(db).ListRoles)
				r.Get("/audit-logs", admincontrollers.NewAdminController(db).ListAuditLogs)
			})
		})

		r.Post("/webhooks/hdev", webhookcontrollers.NewWebhookController(db).HandleMoPayWebhook)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return r
}
