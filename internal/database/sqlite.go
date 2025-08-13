package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/jessewilliams/work/internal/db"
	"github.com/jessewilliams/work/internal/models"
)

type SQLiteDB struct {
	conn    *sql.DB
	queries *db.Queries
}

func NewSQLiteDB(databaseURL string) (*SQLiteDB, error) {
	conn, err := sql.Open("sqlite3", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &SQLiteDB{
		conn:    conn,
		queries: db.New(conn),
	}, nil
}

func (s *SQLiteDB) Close() error {
	return s.conn.Close()
}

func (s *SQLiteDB) GetConnection() *sql.DB {
	return s.conn
}

func (s *SQLiteDB) CreateClient(ctx context.Context, name string, hourlyRate float64) (*models.Client, error) {
	client, err := s.queries.CreateClient(ctx, db.CreateClientParams{
		ID:   models.NewUUID(),
		Name: name,
		HourlyRate: sql.NullFloat64{
			Float64: hourlyRate,
			Valid:   hourlyRate > 0,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	var rate float64
	if client.HourlyRate.Valid {
		rate = client.HourlyRate.Float64
	}

	return &models.Client{
		ID:         client.ID,
		Name:       client.Name,
		HourlyRate: rate,
		CreatedAt:  client.CreatedAt,
		UpdatedAt:  client.UpdatedAt,
	}, nil
}

func (s *SQLiteDB) GetClientByName(ctx context.Context, name string) (*models.Client, error) {
	client, err := s.queries.GetClientByName(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get client by name: %w", err)
	}

	var rate float64
	if client.HourlyRate.Valid {
		rate = client.HourlyRate.Float64
	}

	return &models.Client{
		ID:         client.ID,
		Name:       client.Name,
		HourlyRate: rate,
		CreatedAt:  client.CreatedAt,
		UpdatedAt:  client.UpdatedAt,
	}, nil
}

func (s *SQLiteDB) ListClients(ctx context.Context) ([]*models.Client, error) {
	clients, err := s.queries.ListClients(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list clients: %w", err)
	}

	result := make([]*models.Client, len(clients))
	for i, client := range clients {
		var rate float64
		if client.HourlyRate.Valid {
			rate = client.HourlyRate.Float64
		}

		result[i] = &models.Client{
			ID:         client.ID,
			Name:       client.Name,
			HourlyRate: rate,
			CreatedAt:  client.CreatedAt,
			UpdatedAt:  client.UpdatedAt,
		}
	}

	return result, nil
}

func (s *SQLiteDB) CreateWorkSession(ctx context.Context, clientID string, description *string, hourlyRate float64) (*models.WorkSession, error) {
	var desc sql.NullString
	if description != nil {
		desc = sql.NullString{String: *description, Valid: true}
	}

	var rate sql.NullFloat64
	if hourlyRate > 0 {
		rate = sql.NullFloat64{Float64: hourlyRate, Valid: true}
	}

	session, err := s.queries.CreateSession(ctx, db.CreateSessionParams{
		ID:          models.NewUUID(),
		ClientID:    clientID,
		StartTime:   time.Now(),
		Description: desc,
		HourlyRate:  rate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create work session: %w", err)
	}

	var sessionRate *float64
	if session.HourlyRate.Valid {
		sessionRate = &session.HourlyRate.Float64
	}

	return &models.WorkSession{
		ID:          session.ID,
		ClientID:    session.ClientID,
		StartTime:   session.StartTime,
		EndTime:     nullTimeToPtr(session.EndTime),
		Description: nullStringToPtr(session.Description),
		HourlyRate:  sessionRate,
		CreatedAt:   session.CreatedAt,
		UpdatedAt:   session.UpdatedAt,
	}, nil
}

func (s *SQLiteDB) GetActiveSession(ctx context.Context) (*models.WorkSession, error) {
	session, err := s.queries.GetActiveSession(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get active session: %w", err)
	}

	return &models.WorkSession{
		ID:          session.ID,
		ClientID:    session.ClientID,
		StartTime:   session.StartTime,
		EndTime:     nullTimeToPtr(session.EndTime),
		Description: nullStringToPtr(session.Description),
		CreatedAt:   session.CreatedAt,
		UpdatedAt:   session.UpdatedAt,
		ClientName:  session.ClientName,
	}, nil
}

func (s *SQLiteDB) StopWorkSession(ctx context.Context, sessionID string) (*models.WorkSession, error) {
	session, err := s.queries.StopSession(ctx, db.StopSessionParams{
		ID:      sessionID,
		EndTime: sql.NullTime{Time: time.Now(), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to stop work session: %w", err)
	}

	return &models.WorkSession{
		ID:          session.ID,
		ClientID:    session.ClientID,
		StartTime:   session.StartTime,
		EndTime:     nullTimeToPtr(session.EndTime),
		Description: nullStringToPtr(session.Description),
		CreatedAt:   session.CreatedAt,
		UpdatedAt:   session.UpdatedAt,
	}, nil
}

func (s *SQLiteDB) ListRecentSessions(ctx context.Context, limit int32) ([]*models.WorkSession, error) {
	sessions, err := s.queries.ListRecentSessions(ctx, int64(limit))
	if err != nil {
		return nil, fmt.Errorf("failed to list recent sessions: %w", err)
	}

	result := make([]*models.WorkSession, len(sessions))
	for i, session := range sessions {
		result[i] = &models.WorkSession{
			ID:          session.ID,
			ClientID:    session.ClientID,
			StartTime:   session.StartTime,
			EndTime:     nullTimeToPtr(session.EndTime),
			Description: nullStringToPtr(session.Description),
			CreatedAt:   session.CreatedAt,
			UpdatedAt:   session.UpdatedAt,
			ClientName:  session.ClientName,
		}
	}

	return result, nil
}

func (s *SQLiteDB) ListSessionsWithDateRange(ctx context.Context, fromDate, toDate string, limit int32) ([]*models.WorkSession, error) {
	var startDate, endDate interface{}
	if fromDate != "" {
		startDate = fromDate
	}
	if toDate != "" {
		endDate = toDate
	}

	sessions, err := s.queries.ListSessionsWithDateRange(ctx, db.ListSessionsWithDateRangeParams{
		StartDate:  startDate,
		EndDate:    endDate,
		LimitCount: int64(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions with date range: %w", err)
	}

	result := make([]*models.WorkSession, len(sessions))
	for i, session := range sessions {
		result[i] = &models.WorkSession{
			ID:          session.ID,
			ClientID:    session.ClientID,
			StartTime:   session.StartTime,
			EndTime:     nullTimeToPtr(session.EndTime),
			Description: nullStringToPtr(session.Description),
			CreatedAt:   session.CreatedAt,
			UpdatedAt:   session.UpdatedAt,
			ClientName:  session.ClientName,
		}
	}

	return result, nil
}

func (s *SQLiteDB) UpdateClientRate(ctx context.Context, clientID string, hourlyRate float64) (*models.Client, error) {
	client, err := s.queries.UpdateClientRate(ctx, db.UpdateClientRateParams{
		ID: clientID,
		HourlyRate: sql.NullFloat64{
			Float64: hourlyRate,
			Valid:   hourlyRate > 0,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update client rate: %w", err)
	}

	var rate float64
	if client.HourlyRate.Valid {
		rate = client.HourlyRate.Float64
	}

	return &models.Client{
		ID:         client.ID,
		Name:       client.Name,
		HourlyRate: rate,
		CreatedAt:  client.CreatedAt,
		UpdatedAt:  client.UpdatedAt,
	}, nil
}

func (s *SQLiteDB) DeleteAllSessions(ctx context.Context) error {
	err := s.queries.DeleteAllSessions(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete all sessions: %w", err)
	}
	return nil
}

func (s *SQLiteDB) DeleteSessionsByDateRange(ctx context.Context, fromDate, toDate string) error {
	var startDate, endDate interface{}
	if fromDate != "" {
		startDate = fromDate
	}
	if toDate != "" {
		endDate = toDate
	}

	err := s.queries.DeleteSessionsByDateRange(ctx, db.DeleteSessionsByDateRangeParams{
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return fmt.Errorf("failed to delete sessions by date range: %w", err)
	}
	return nil
}

func nullTimeToPtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

func nullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}
