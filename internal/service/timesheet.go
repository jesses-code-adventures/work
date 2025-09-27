package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jesses-code-adventures/work/internal/config"
	"github.com/jesses-code-adventures/work/internal/database"
	"github.com/jesses-code-adventures/work/internal/models"
)

type TimesheetService struct {
	db  database.DB
	cfg *config.Config
}

func NewTimesheetService(db database.DB, cfg *config.Config) *TimesheetService {
	return &TimesheetService{db: db, cfg: cfg}
}

func (s *TimesheetService) Config() *config.Config {
	return s.cfg
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

func (s *TimesheetService) StartWorkWithTime(ctx context.Context, clientName string, startTime time.Time, description *string) (*models.WorkSession, error) {
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

	session, err := s.db.CreateWorkSessionWithStartTime(ctx, client.ID, startTime, description, client.HourlyRate)
	if err != nil {
		return nil, fmt.Errorf("failed to create work session: %w", err)
	}

	session.ClientName = clientName
	return session, nil
}

func (s *TimesheetService) CreateSessionWithTimes(ctx context.Context, clientName string, startTime, endTime time.Time, description *string) (*models.WorkSession, error) {
	client, err := s.db.GetClientByName(ctx, clientName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("client '%s' does not exist", clientName)
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	var hourlyRate float64
	if client.HourlyRate > 0 {
		hourlyRate = client.HourlyRate
	}

	session, err := s.db.CreateWorkSessionWithTimes(ctx, client.ID, startTime, endTime, description, hourlyRate)
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

func (s *TimesheetService) ListSessionsByClient(ctx context.Context, clientName string, limit int32) ([]*models.WorkSession, error) {
	return s.db.ListSessionsByClient(ctx, clientName, limit)
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

func (s *TimesheetService) GetClientsWithDirectories(ctx context.Context) ([]*models.Client, error) {
	return s.db.GetClientsWithDirectories(ctx)
}

func (s *TimesheetService) GetClientByName(ctx context.Context, name string) (*models.Client, error) {
	return s.db.GetClientByName(ctx, name)
}

func (s *TimesheetService) GetClientByID(ctx context.Context, ID string) (*models.Client, error) {
	return s.db.GetClientByID(ctx, ID)
}

func (s *TimesheetService) UpdateClient(ctx context.Context, clientName string, updates *database.ClientUpdateDetails) (*models.Client, error) {
	c, err := s.db.GetClientByName(ctx, clientName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("client '%s' does not exist", clientName)
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}
	return s.db.UpdateClient(ctx, c.ID, updates)
}

func (s *TimesheetService) DisplayClient(ctx context.Context, client *models.Client) {
	fmt.Printf("Client: %s\n", client.Name)
	if client.HourlyRate != 0.0 {
		fmt.Printf("Rate: %s\n", s.FormatBillableAmount(client.HourlyRate))
	}
	if client.CompanyName != nil {
		fmt.Printf("Company: %s\n", *client.CompanyName)
	}
	if client.ContactName != nil {
		fmt.Printf("Contact: %s\n", *client.ContactName)
	}
	if client.Email != nil {
		fmt.Printf("Email: %s\n", *client.Email)
	}
	if client.Phone != nil {
		fmt.Printf("Phone: %s\n", *client.Phone)
	}
	if client.AddressLine1 != nil {
		fmt.Printf("Address: %s", *client.AddressLine1)
		if client.AddressLine2 != nil {
			fmt.Printf(", %s", *client.AddressLine2)
		}
		fmt.Printf("\n")
	}
	if client.City != nil || client.State != nil || client.PostalCode != nil {
		fmt.Printf("Location: ")
		if client.City != nil {
			fmt.Printf("%s", *client.City)
		}
		if client.State != nil {
			fmt.Printf(", %s", *client.State)
		}
		if client.PostalCode != nil {
			fmt.Printf(" %s", *client.PostalCode)
		}
		fmt.Printf("\n")
	}
	if client.Country != nil {
		fmt.Printf("Country: %s\n", *client.Country)
	}
	if client.Abn != nil {
		fmt.Printf("ABN: %s\n", *client.Abn)
	}
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

func (s *TimesheetService) GetSessionsWithoutDescription(ctx context.Context, clientName, sessionID *string) ([]*models.WorkSession, error) {
	return s.db.GetSessionsWithoutDescription(ctx, clientName, sessionID)
}

func (s *TimesheetService) GetSessionByID(ctx context.Context, sessionID string) (*models.WorkSession, error) {
	return s.db.GetSessionByID(ctx, sessionID)
}

func (s *TimesheetService) UpdateSessionDescription(ctx context.Context, sessionID string, description string, fullWorkSummary *string) (*models.WorkSession, error) {
	return s.db.UpdateSessionDescription(ctx, sessionID, description, fullWorkSummary)
}

func (s *TimesheetService) AddSessionNote(ctx context.Context, sessionID string, note string) (*models.WorkSession, error) {
	session, err := s.db.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var currentNotes string
	if session.OutsideGit != nil {
		currentNotes = *session.OutsideGit
	}

	newNote := fmt.Sprintf("- %s", note)
	var updatedNotes string
	if currentNotes == "" {
		updatedNotes = newNote
	} else {
		updatedNotes = fmt.Sprintf("%s\n%s", currentNotes, newNote)
	}

	return s.db.UpdateSessionOutsideGit(ctx, sessionID, updatedNotes)
}
