package models

import (
	"time"

	"github.com/google/uuid"
)

type Client struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	HourlyRate   float64   `json:"hourly_rate" db:"hourly_rate"`
	CompanyName  *string   `json:"company_name,omitempty" db:"company_name"`
	ContactName  *string   `json:"contact_name,omitempty" db:"contact_name"`
	Email        *string   `json:"email,omitempty" db:"email"`
	Phone        *string   `json:"phone,omitempty" db:"phone"`
	AddressLine1 *string   `json:"address_line1,omitempty" db:"address_line1"`
	AddressLine2 *string   `json:"address_line2,omitempty" db:"address_line2"`
	City         *string   `json:"city,omitempty" db:"city"`
	State        *string   `json:"state,omitempty" db:"state"`
	PostalCode   *string   `json:"postal_code,omitempty" db:"postal_code"`
	Country      *string   `json:"country,omitempty" db:"country"`
	TaxNumber    *string   `json:"tax_number,omitempty" db:"tax_number"`
	Dir          *string   `json:"dir,omitempty" db:"dir"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type WorkSession struct {
	ID              string     `json:"id" db:"id"`
	ClientID        string     `json:"client_id" db:"client_id"`
	StartTime       time.Time  `json:"start_time" db:"start_time"`
	EndTime         *time.Time `json:"end_time,omitempty" db:"end_time"`
	Description     *string    `json:"description,omitempty" db:"description"`
	HourlyRate      *float64   `json:"hourly_rate,omitempty" db:"hourly_rate"`
	FullWorkSummary *string    `json:"full_work_summary,omitempty" db:"full_work_summary"`
	OutsideGit      *string    `json:"outside_git,omitempty" db:"outside_git"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`

	ClientName string `json:"client_name,omitempty" db:"client_name"`
}

func NewUUID() string {
	return uuid.Must(uuid.NewV7()).String()
}
