package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Client struct {
	ID             string           `json:"id" db:"id"`
	Name           string           `json:"name" db:"name"`
	HourlyRate     decimal.Decimal  `json:"hourly_rate" db:"hourly_rate"`
	CompanyName    *string          `json:"company_name,omitempty" db:"company_name"`
	ContactName    *string          `json:"contact_name,omitempty" db:"contact_name"`
	Email          *string          `json:"email,omitempty" db:"email"`
	Phone          *string          `json:"phone,omitempty" db:"phone"`
	AddressLine1   *string          `json:"address_line1,omitempty" db:"address_line1"`
	AddressLine2   *string          `json:"address_line2,omitempty" db:"address_line2"`
	City           *string          `json:"city,omitempty" db:"city"`
	State          *string          `json:"state,omitempty" db:"state"`
	PostalCode     *string          `json:"postal_code,omitempty" db:"postal_code"`
	Country        *string          `json:"country,omitempty" db:"country"`
	Abn            *string          `json:"abn,omitempty" db:"abn"`
	Dir            *string          `json:"dir,omitempty" db:"dir"`
	RetainerAmount *decimal.Decimal `json:"retainer_amount,omitempty" db:"retainer_amount"`
	RetainerHours  *float64         `json:"retainer_hours,omitempty" db:"retainer_hours"`
	RetainerBasis  *string          `json:"retainer_basis,omitempty" db:"retainer_basis"`
	CreatedAt      time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at" db:"updated_at"`
}

type WorkSession struct {
	ID              string           `json:"id" db:"id"`
	ClientID        string           `json:"client_id" db:"client_id"`
	StartTime       time.Time        `json:"start_time" db:"start_time"`
	EndTime         *time.Time       `json:"end_time,omitempty" db:"end_time"`
	Description     *string          `json:"description,omitempty" db:"description"`
	HourlyRate      *decimal.Decimal `json:"hourly_rate,omitempty" db:"hourly_rate"`
	FullWorkSummary *string          `json:"full_work_summary,omitempty" db:"full_work_summary"`
	OutsideGit      *string          `json:"outside_git,omitempty" db:"outside_git"`
	InvoiceID       *string          `json:"invoice_id,omitempty" db:"invoice_id"`
	CreatedAt       time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at" db:"updated_at"`

	ClientName string `json:"client_name,omitempty" db:"client_name"`
}

type Invoice struct {
	ID              string          `json:"id" db:"id"`
	ClientID        string          `json:"client_id" db:"client_id"`
	InvoiceNumber   string          `json:"invoice_number" db:"invoice_number"`
	PeriodType      string          `json:"period_type" db:"period_type"`
	PeriodStartDate time.Time       `json:"period_start_date" db:"period_start_date"`
	PeriodEndDate   time.Time       `json:"period_end_date" db:"period_end_date"`
	SubtotalAmount  decimal.Decimal `json:"subtotal_amount" db:"subtotal_amount"`
	GstAmount       decimal.Decimal `json:"gst_amount" db:"gst_amount"`
	TotalAmount     decimal.Decimal `json:"total_amount" db:"total_amount"`
	AmountPaid      decimal.Decimal `json:"amount_paid" db:"amount_paid"`
	PaymentDate     *time.Time      `json:"payment_date,omitempty" db:"payment_date"`
	GeneratedDate   time.Time       `json:"generated_date" db:"generated_date"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`

	ClientName string `json:"client_name,omitempty" db:"client_name"`
}

func NewUUID() string {
	return uuid.Must(uuid.NewV7()).String()
}
