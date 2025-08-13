package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	_ "github.com/tursodatabase/libsql-client-go/libsql"

	"github.com/jesses-code-adventures/work/internal/config"
	"github.com/jesses-code-adventures/work/internal/db"
	"github.com/jesses-code-adventures/work/internal/models"
)

type SQLiteDB struct {
	conn     *sql.DB
	queries  *db.Queries
	exitFunc func()
}

func NewDB(cfg *config.Config) (*SQLiteDB, error) {
	conn, err := sql.Open(cfg.DatabaseDriver, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	s := SQLiteDB{
		conn:    conn,
		queries: db.New(conn),
	}
	return &s, nil
}

func NewTursoDBWithEmbeddedReplica(cfg *config.Config) (*SQLiteDB, error) {
	// TODO: implement - https://docs.turso.tech/sdk/go/quickstart
	return nil, fmt.Errorf("not implemented")
	// connector, err := libsql.NewEmbeddedReplicaConnector(cfg.DatabasePath, cfg.DatabaseURL,
	// 	libsql.WithAuthToken(cfg.TursoToken),
	// )
	// if err != nil {
	// 	fmt.Println("Error creating connector:", err)
	// 	os.Exit(1)
	// }
	// conn := sql.OpenDB(connector)
	// s := SQLiteDB{
	// 	conn:    conn,
	// 	queries: db.New(conn),
	// }
	// s.exitFunc = func() {
	// 	if _, err := connector.Sync(); err != nil {
	// 		fmt.Println("Error syncing database:", err)
	// 		os.Exit(1)
	// 	}
	// }
	// return &s, nil
}

func (s *SQLiteDB) Close() error {
	if s.exitFunc != nil {
		s.exitFunc()
	}
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

	return s.convertDBClientToModel(client), nil
}

func (s *SQLiteDB) GetClientByName(ctx context.Context, name string) (*models.Client, error) {
	client, err := s.queries.GetClientByName(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get client by name: %w", err)
	}

	return s.convertDBClientToModel(client), nil
}

func (s *SQLiteDB) ListClients(ctx context.Context) ([]*models.Client, error) {
	clients, err := s.queries.ListClients(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list clients: %w", err)
	}

	result := make([]*models.Client, len(clients))
	for i, client := range clients {
		result[i] = s.convertDBClientToModel(client)
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

func (s *SQLiteDB) ListRecentSessions(ctx context.Context, limit int32) ([]*models.WorkSession, error) {
	sessions, err := s.queries.ListRecentSessions(ctx, int64(limit))
	if err != nil {
		return nil, fmt.Errorf("failed to list recent sessions: %w", err)
	}

	result := make([]*models.WorkSession, len(sessions))
	for i, session := range sessions {
		var sessionRate *float64
		if session.HourlyRate.Valid {
			sessionRate = &session.HourlyRate.Float64
		}

		result[i] = &models.WorkSession{
			ID:          session.ID,
			ClientID:    session.ClientID,
			StartTime:   session.StartTime,
			EndTime:     nullTimeToPtr(session.EndTime),
			Description: nullStringToPtr(session.Description),
			HourlyRate:  sessionRate,
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
		var sessionRate *float64
		if session.HourlyRate.Valid {
			sessionRate = &session.HourlyRate.Float64
		}

		result[i] = &models.WorkSession{
			ID:          session.ID,
			ClientID:    session.ClientID,
			StartTime:   session.StartTime,
			EndTime:     nullTimeToPtr(session.EndTime),
			Description: nullStringToPtr(session.Description),
			HourlyRate:  sessionRate,
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

	return s.convertDBClientToModel(client), nil
}

func (s *SQLiteDB) UpdateClientBilling(ctx context.Context, clientID string, billing *ClientBillingDetails) (*models.Client, error) {
	client, err := s.queries.UpdateClientBilling(ctx, db.UpdateClientBillingParams{
		ID:           clientID,
		CompanyName:  ptrToNullString(billing.CompanyName),
		ContactName:  ptrToNullString(billing.ContactName),
		Email:        ptrToNullString(billing.Email),
		Phone:        ptrToNullString(billing.Phone),
		AddressLine1: ptrToNullString(billing.AddressLine1),
		AddressLine2: ptrToNullString(billing.AddressLine2),
		City:         ptrToNullString(billing.City),
		State:        ptrToNullString(billing.State),
		PostalCode:   ptrToNullString(billing.PostalCode),
		Country:      ptrToNullString(billing.Country),
		TaxNumber:    ptrToNullString(billing.TaxNumber),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update client billing: %w", err)
	}

	return s.convertDBClientToModel(client), nil
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

func (s *SQLiteDB) convertDBClientToModel(client db.Client) *models.Client {
	var rate float64
	if client.HourlyRate.Valid {
		rate = client.HourlyRate.Float64
	}

	return &models.Client{
		ID:           client.ID,
		Name:         client.Name,
		HourlyRate:   rate,
		CompanyName:  nullStringToPtr(client.CompanyName),
		ContactName:  nullStringToPtr(client.ContactName),
		Email:        nullStringToPtr(client.Email),
		Phone:        nullStringToPtr(client.Phone),
		AddressLine1: nullStringToPtr(client.AddressLine1),
		AddressLine2: nullStringToPtr(client.AddressLine2),
		City:         nullStringToPtr(client.City),
		State:        nullStringToPtr(client.State),
		PostalCode:   nullStringToPtr(client.PostalCode),
		Country:      nullStringToPtr(client.Country),
		TaxNumber:    nullStringToPtr(client.TaxNumber),
		CreatedAt:    client.CreatedAt,
		UpdatedAt:    client.UpdatedAt,
	}
}

func ptrToNullString(s *string) sql.NullString {
	if s != nil {
		return sql.NullString{String: *s, Valid: true}
	}
	return sql.NullString{Valid: false}
}
