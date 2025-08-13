package models

import (
	"time"

	"github.com/google/uuid"
)

type Client struct {
	ID         string    `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	HourlyRate float64   `json:"hourly_rate" db:"hourly_rate"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type WorkSession struct {
	ID          string     `json:"id" db:"id"`
	ClientID    string     `json:"client_id" db:"client_id"`
	StartTime   time.Time  `json:"start_time" db:"start_time"`
	EndTime     *time.Time `json:"end_time,omitempty" db:"end_time"`
	Description *string    `json:"description,omitempty" db:"description"`
	HourlyRate  *float64   `json:"hourly_rate,omitempty" db:"hourly_rate"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`

	ClientName string `json:"client_name,omitempty" db:"client_name"`
}

func NewUUID() string {
	return uuid.Must(uuid.NewV7()).String()
}
