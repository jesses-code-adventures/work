package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jesses-code-adventures/work/internal/database"
	"github.com/jesses-code-adventures/work/internal/models"
)

type TimesheetService struct {
	db database.DB
}

func NewTimesheetService(db database.DB) *TimesheetService {
	return &TimesheetService{db: db}
}

func (s *TimesheetService) StartWork(ctx context.Context, clientName string, description *string) (*models.WorkSession, error) {
	activeSession, err := s.db.GetActiveSession(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check for active session: %w", err)
	}

	if activeSession != nil {
		fmt.Printf("Stopping current session for %s (started at %s)\n",
			activeSession.ClientName,
			activeSession.StartTime.Format("15:04:05"))

		_, err := s.db.StopWorkSession(ctx, activeSession.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to stop active session: %w", err)
		}
	}

	client, err := s.db.GetClientByName(ctx, clientName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("client '%s' does not exist", clientName)
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	session, err := s.db.CreateWorkSession(ctx, client.ID, description, client.HourlyRate)
	if err != nil {
		return nil, fmt.Errorf("failed to create work session: %w", err)
	}

	session.ClientName = clientName
	return session, nil
}

func (s *TimesheetService) StopWork(ctx context.Context) (*models.WorkSession, error) {
	activeSession, err := s.db.GetActiveSession(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check for active session: %w", err)
	}

	if activeSession == nil {
		return nil, fmt.Errorf("no active work session to stop")
	}

	stoppedSession, err := s.db.StopWorkSession(ctx, activeSession.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to stop work session: %w", err)
	}

	stoppedSession.ClientName = activeSession.ClientName
	return stoppedSession, nil
}

func (s *TimesheetService) GetActiveSession(ctx context.Context) (*models.WorkSession, error) {
	return s.db.GetActiveSession(ctx)
}

func (s *TimesheetService) ListRecentSessions(ctx context.Context, limit int32) ([]*models.WorkSession, error) {
	return s.db.ListRecentSessions(ctx, limit)
}

func (s *TimesheetService) ListSessionsWithDateRange(ctx context.Context, fromDate, toDate string, limit int32) ([]*models.WorkSession, error) {
	from := s.formatDateForQuery(fromDate, true)
	to := s.formatDateForQuery(toDate, false)
	return s.db.ListSessionsWithDateRange(ctx, from, to, limit)
}

func (s *TimesheetService) DeleteAllSessions(ctx context.Context) error {
	return s.db.DeleteAllSessions(ctx)
}

func (s *TimesheetService) DeleteSessionsByDateRange(ctx context.Context, fromDate, toDate string) error {
	from := s.formatDateForQuery(fromDate, true)
	to := s.formatDateForQuery(toDate, false)
	return s.db.DeleteSessionsByDateRange(ctx, from, to)
}

func (s *TimesheetService) CreateClient(ctx context.Context, name string, hourlyRate float64) (*models.Client, error) {
	existing, err := s.db.GetClientByName(ctx, name)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check for existing client: %w", err)
	}

	if existing != nil {
		return nil, fmt.Errorf("client '%s' already exists", name)
	}

	return s.db.CreateClient(ctx, name, hourlyRate)
}

func (s *TimesheetService) ListClients(ctx context.Context) ([]*models.Client, error) {
	return s.db.ListClients(ctx)
}

func (s *TimesheetService) GetClientByName(ctx context.Context, name string) (*models.Client, error) {
	return s.db.GetClientByName(ctx, name)
}

func (s *TimesheetService) UpdateClient(ctx context.Context, client string, rate float64) (*models.Client, error) {
	if rate == 0.0 {
		return nil, nil
	}
	c, err := s.db.GetClientByName(ctx, client)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("client '%s' does not exist", client)
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}
	return s.db.UpdateClientRate(ctx, c.ID, rate)
}

func (s *TimesheetService) UpdateClientBilling(ctx context.Context, clientName string, billing *database.ClientBillingDetails) (*models.Client, error) {
	c, err := s.db.GetClientByName(ctx, clientName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("client '%s' does not exist", clientName)
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}
	return s.db.UpdateClientBilling(ctx, c.ID, billing)
}

func (s *TimesheetService) CalculateDuration(session *models.WorkSession) time.Duration {
	if session.EndTime == nil {
		return time.Since(session.StartTime)
	}
	return session.EndTime.Sub(session.StartTime)
}

func (s *TimesheetService) FormatDuration(d time.Duration) string {
	hours := d / time.Hour
	minutes := (d % time.Hour) / time.Minute
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func (s *TimesheetService) CalculateBillableAmount(session *models.WorkSession) float64 {
	if session.HourlyRate == nil || *session.HourlyRate <= 0 {
		return 0.0
	}

	duration := s.CalculateDuration(session)
	hours := duration.Hours()
	return hours * (*session.HourlyRate)
}

func (s *TimesheetService) FormatBillableAmount(amount float64) string {
	if amount <= 0 {
		return "$0.00"
	}
	return fmt.Sprintf("$%.2f", amount)
}

func (s *TimesheetService) formatDateForQuery(dateStr string, isStart bool) string {
	if dateStr == "" {
		return ""
	}

	if len(dateStr) == 10 {
		if isStart {
			return dateStr + " 00:00:00"
		} else {
			return dateStr + " 23:59:59"
		}
	}

	return dateStr
}
