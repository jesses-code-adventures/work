package database

import (
	"context"

	"github.com/jessewilliams/work/internal/models"
)

type DB interface {
	Close() error

	CreateClient(ctx context.Context, name string, hourlyRate float64) (*models.Client, error)
	GetClientByName(ctx context.Context, name string) (*models.Client, error)
	ListClients(ctx context.Context) ([]*models.Client, error)
	UpdateClientRate(ctx context.Context, clientID string, hourlyRate float64) (*models.Client, error)

	CreateWorkSession(ctx context.Context, clientID string, description *string, hourlyRate float64) (*models.WorkSession, error)
	GetActiveSession(ctx context.Context) (*models.WorkSession, error)
	StopWorkSession(ctx context.Context, sessionID string) (*models.WorkSession, error)
	ListRecentSessions(ctx context.Context, limit int32) ([]*models.WorkSession, error)
	ListSessionsWithDateRange(ctx context.Context, fromDate, toDate string, limit int32) ([]*models.WorkSession, error)
	DeleteAllSessions(ctx context.Context) error
	DeleteSessionsByDateRange(ctx context.Context, fromDate, toDate string) error
}
