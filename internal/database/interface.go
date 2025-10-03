package database

import (
	"context"
	"time"

	"github.com/jesses-code-adventures/work/internal/db"
	"github.com/jesses-code-adventures/work/internal/models"
	"github.com/shopspring/decimal"
)

type ClientUpdateDetails struct {
	HourlyRate     *decimal.Decimal
	CompanyName    *string
	ContactName    *string
	Email          *string
	Phone          *string
	AddressLine1   *string
	AddressLine2   *string
	City           *string
	State          *string
	PostalCode     *string
	Country        *string
	Abn            *string
	Dir            *string
	RetainerAmount *decimal.Decimal
	RetainerHours  *float64
	RetainerBasis  *string
}

type DB interface {
	Close() error

	CreateClient(ctx context.Context, name string, hourlyRate decimal.Decimal, retainerAmount *decimal.Decimal, retainerHours *float64, retainerBasis, dir *string) (*models.Client, error)
	GetClientByName(ctx context.Context, name string) (*models.Client, error)
	GetClientByID(ctx context.Context, ID string) (*models.Client, error)
	ListClients(ctx context.Context) ([]*models.Client, error)
	GetClientsWithDirectories(ctx context.Context) ([]*models.Client, error)
	UpdateClient(ctx context.Context, clientID string, billing *ClientUpdateDetails) (*models.Client, error)

	CreateWorkSession(ctx context.Context, clientID string, description *string, hourlyRate decimal.Decimal, includesGst bool) (*models.WorkSession, error)
	CreateWorkSessionWithStartTime(ctx context.Context, clientID string, startTime time.Time, description *string, hourlyRate decimal.Decimal, includesGst bool) (*models.WorkSession, error)
	CreateWorkSessionWithTimes(ctx context.Context, clientID string, startTime, endTime time.Time, description *string, hourlyRate decimal.Decimal, includesGst bool) (*models.WorkSession, error)
	GetActiveSession(ctx context.Context) (*models.WorkSession, error)
	StopWorkSession(ctx context.Context, sessionID string) (*models.WorkSession, error)
	ListRecentSessions(ctx context.Context, limit int32) ([]*models.WorkSession, error)
	ListSessionsWithDateRange(ctx context.Context, fromDate, toDate string, limit int32) ([]*models.WorkSession, error)
	ListSessionsByClient(ctx context.Context, clientName string, limit int32) ([]*models.WorkSession, error)
	GetSessionsWithoutDescription(ctx context.Context, clientName *string, sessionID *string) ([]*models.WorkSession, error)
	GetSessionByID(ctx context.Context, sessionID string) (*models.WorkSession, error)
	UpdateSessionDescription(ctx context.Context, sessionID string, description string, fullWorkSummary *string) (*models.WorkSession, error)
	UpdateSessionOutsideGit(ctx context.Context, sessionID string, outsideGit string) (*models.WorkSession, error)
	DeleteAllSessions(ctx context.Context) error
	DeleteSessionsByDateRange(ctx context.Context, fromDate, toDate string) error

	// Invoice operations
	CreateInvoice(ctx context.Context, clientID, invoiceNumber, periodType string, periodStart, periodEnd time.Time, subtotal, gst, total decimal.Decimal) (*models.Invoice, error)
	GetInvoiceByID(ctx context.Context, invoiceID string) (*models.Invoice, error)
	PayInvoice(ctx context.Context, param db.PayInvoiceParams) error
	GetInvoiceByNumber(ctx context.Context, invoiceNumber string) (*models.Invoice, error)
	ListInvoices(ctx context.Context, limit int32) ([]*models.Invoice, error)
	GetInvoicesByClient(ctx context.Context, clientName string) ([]*models.Invoice, error)
	GetInvoicesByPeriod(ctx context.Context, periodStart, periodEnd time.Time, periodType string) ([]*models.Invoice, error)
	DeleteInvoice(ctx context.Context, invoiceID string) error
	GetSessionsForPeriodWithoutInvoice(ctx context.Context, startDate, endDate time.Time) ([]*models.WorkSession, error)
	GetSessionsForPeriodWithoutInvoiceByClient(ctx context.Context, startDate, endDate time.Time, clientName string) ([]*models.WorkSession, error)
	GetSessionsByInvoiceID(ctx context.Context, invoiceID string) ([]*models.WorkSession, error)
	GetInvoicesByPeriodAndClient(ctx context.Context, periodStart, periodEnd time.Time, periodType, clientName string) ([]*models.Invoice, error)
	UpdateSessionInvoiceID(ctx context.Context, sessionID, invoiceID string) error
	ClearSessionInvoiceIDs(ctx context.Context, invoiceID string) error

	// Expense operations
	CreateExpense(ctx context.Context, amount decimal.Decimal, expenseDate time.Time, reference *string, clientID *string, invoiceID *string, description *string) (*models.Expense, error)
	GetExpenseByID(ctx context.Context, expenseID string) (*models.Expense, error)
	ListExpenses(ctx context.Context) ([]*models.Expense, error)
	ListExpensesByClient(ctx context.Context, clientID string) ([]*models.Expense, error)
	ListExpensesByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*models.Expense, error)
	ListExpensesByClientAndDateRange(ctx context.Context, clientID string, startDate, endDate time.Time) ([]*models.Expense, error)
	GetExpensesByInvoiceID(ctx context.Context, invoiceID string) ([]*models.Expense, error)
	GetExpensesWithoutInvoiceByClient(ctx context.Context, clientID string) ([]*models.Expense, error)
	GetExpensesWithoutInvoiceByClientAndDateRange(ctx context.Context, clientID string, startDate, endDate time.Time) ([]*models.Expense, error)
	UpdateExpense(ctx context.Context, expenseID string, amount *decimal.Decimal, expenseDate *time.Time, reference *string, clientID *string, invoiceID *string, description *string) (*models.Expense, error)
	UpdateExpenseInvoiceID(ctx context.Context, expenseID string, invoiceID *string) error
	ClearExpenseInvoiceIDs(ctx context.Context, invoiceID string) error
	DeleteExpense(ctx context.Context, expenseID string) error
}
