package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shopspring/decimal"
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

func (s *SQLiteDB) CreateClient(ctx context.Context, name string, hourlyRate decimal.Decimal, retainerAmount *decimal.Decimal, retainerHours *float64, retainerBasis, dir *string) (*models.Client, error) {
	client, err := s.queries.CreateClient(ctx, db.CreateClientParams{
		ID:   models.NewUUID(),
		Name: name,
		HourlyRate: decimal.NullDecimal{
			Decimal: hourlyRate,
			Valid:   hourlyRate.GreaterThan(decimal.Zero),
		},
		RetainerAmount: ptrToNullDecimal(retainerAmount),
		RetainerHours:  ptrToNullFloat64(retainerHours),
		RetainerBasis:  ptrToNullString(retainerBasis),
		Dir:            ptrToNullString(dir),
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

func (s *SQLiteDB) GetClientByID(ctx context.Context, ID string) (*models.Client, error) {
	client, err := s.queries.GetClientByID(ctx, ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get client by ID: %w", err)
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

func (s *SQLiteDB) GetClientsWithDirectories(ctx context.Context) ([]*models.Client, error) {
	clients, err := s.queries.GetClientsWithDirectories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get clients with directories: %w", err)
	}

	result := make([]*models.Client, len(clients))
	for i, client := range clients {
		result[i] = s.convertDBClientToModel(client)
	}

	return result, nil
}

func (s *SQLiteDB) CreateWorkSession(ctx context.Context, clientID string, description *string, hourlyRate decimal.Decimal, includesGst bool) (*models.WorkSession, error) {
	var desc sql.NullString
	if description != nil {
		desc = sql.NullString{String: *description, Valid: true}
	}

	var rate decimal.NullDecimal
	if hourlyRate.GreaterThan(decimal.Zero) {
		rate = decimal.NullDecimal{Decimal: hourlyRate, Valid: true}
	}

	session, err := s.queries.CreateSession(ctx, db.CreateSessionParams{
		ID:          models.NewUUID(),
		ClientID:    clientID,
		StartTime:   time.Now(),
		Description: desc,
		HourlyRate:  rate,
		IncludesGst: includesGst,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create work session: %w", err)
	}

	return &models.WorkSession{
		ID:          session.ID,
		ClientID:    session.ClientID,
		StartTime:   session.StartTime,
		EndTime:     nullTimeToPtr(session.EndTime),
		Description: nullStringToPtr(session.Description),
		HourlyRate:  nullDecimalToPtr(session.HourlyRate),
		OutsideGit:  nullStringToPtr(session.OutsideGit),
		IncludesGst: session.IncludesGst,
		CreatedAt:   session.CreatedAt,
		UpdatedAt:   session.UpdatedAt,
	}, nil
}

func (s *SQLiteDB) CreateWorkSessionWithStartTime(ctx context.Context, clientID string, startTime time.Time, description *string, hourlyRate decimal.Decimal, includesGst bool) (*models.WorkSession, error) {
	var desc sql.NullString
	if description != nil {
		desc = sql.NullString{String: *description, Valid: true}
	}

	var rate decimal.NullDecimal
	if hourlyRate.GreaterThan(decimal.Zero) {
		rate = decimal.NullDecimal{Decimal: hourlyRate, Valid: true}
	}

	session, err := s.queries.CreateSession(ctx, db.CreateSessionParams{
		ID:          models.NewUUID(),
		ClientID:    clientID,
		StartTime:   startTime,
		Description: desc,
		HourlyRate:  rate,
		IncludesGst: includesGst,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create work session: %w", err)
	}

	return &models.WorkSession{
		ID:          session.ID,
		ClientID:    session.ClientID,
		StartTime:   session.StartTime,
		EndTime:     nullTimeToPtr(session.EndTime),
		Description: nullStringToPtr(session.Description),
		HourlyRate:  nullDecimalToPtr(session.HourlyRate),
		OutsideGit:  nullStringToPtr(session.OutsideGit),
		IncludesGst: session.IncludesGst,
		CreatedAt:   session.CreatedAt,
		UpdatedAt:   session.UpdatedAt,
	}, nil
}

func (s *SQLiteDB) CreateWorkSessionWithTimes(ctx context.Context, clientID string, startTime, endTime time.Time, description *string, hourlyRate decimal.Decimal, includesGst bool) (*models.WorkSession, error) {
	var desc sql.NullString
	if description != nil {
		desc = sql.NullString{String: *description, Valid: true}
	}

	var rate decimal.NullDecimal
	if hourlyRate.GreaterThan(decimal.Zero) {
		rate = decimal.NullDecimal{Decimal: hourlyRate, Valid: true}
	}

	session, err := s.queries.CreateSession(ctx, db.CreateSessionParams{
		ID:          models.NewUUID(),
		ClientID:    clientID,
		StartTime:   startTime,
		Description: desc,
		HourlyRate:  rate,
		IncludesGst: includesGst,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create work session: %w", err)
	}

	// Now update the session with the end time
	updatedSession, err := s.queries.StopSession(ctx, db.StopSessionParams{
		ID:      session.ID,
		EndTime: sql.NullTime{Time: endTime, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set end time on session: %w", err)
	}

	return &models.WorkSession{
		ID:          updatedSession.ID,
		ClientID:    updatedSession.ClientID,
		StartTime:   updatedSession.StartTime,
		EndTime:     nullTimeToPtr(updatedSession.EndTime),
		Description: nullStringToPtr(updatedSession.Description),
		HourlyRate:  nullDecimalToPtr(updatedSession.HourlyRate),
		OutsideGit:  nullStringToPtr(updatedSession.OutsideGit),
		IncludesGst: updatedSession.IncludesGst,
		CreatedAt:   updatedSession.CreatedAt,
		UpdatedAt:   updatedSession.UpdatedAt,
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

	sessionRate := decimal.Zero
	if session.HourlyRate.Valid {
		sessionRate = session.HourlyRate.Decimal
	}

	return &models.WorkSession{
		ID:          session.ID,
		ClientID:    session.ClientID,
		StartTime:   session.StartTime,
		EndTime:     nullTimeToPtr(session.EndTime),
		Description: nullStringToPtr(session.Description),
		HourlyRate:  &sessionRate,
		OutsideGit:  nullStringToPtr(session.OutsideGit),
		IncludesGst: session.IncludesGst,
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
		HourlyRate:  nullDecimalToPtr(session.HourlyRate),
		OutsideGit:  nullStringToPtr(session.OutsideGit),
		IncludesGst: session.IncludesGst,
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
		sessionRate := decimal.Zero
		if session.HourlyRate.Valid {
			sessionRate = session.HourlyRate.Decimal
		}

		result[i] = &models.WorkSession{
			ID:              session.ID,
			ClientID:        session.ClientID,
			StartTime:       session.StartTime,
			EndTime:         nullTimeToPtr(session.EndTime),
			Description:     nullStringToPtr(session.Description),
			HourlyRate:      &sessionRate,
			FullWorkSummary: nullStringToPtr(session.FullWorkSummary),
			OutsideGit:      nullStringToPtr(session.OutsideGit),
			InvoiceID:       nullStringToPtr(session.InvoiceID),
			IncludesGst:     session.IncludesGst,
			CreatedAt:       session.CreatedAt,
			UpdatedAt:       session.UpdatedAt,
			ClientName:      session.ClientName,
		}
	}

	return result, nil
}

func (s *SQLiteDB) ListSessionsWithDateRange(ctx context.Context, fromDate, toDate string, limit int32) ([]*models.WorkSession, error) {
	var startDate, endDate any
	if fromDate != "" {
		startDate = fromDate
	}
	if toDate != "" {
		endDate = toDate
	}

	sessions, err := s.queries.ListSessionsWithDateRange(ctx, db.ListSessionsWithDateRangeParams{
		StartDate:  startDate,
		EndDate:    endDate,
		ClientName: nil, // No client filtering in this method
		LimitCount: int64(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions with date range: %w", err)
	}

	result := make([]*models.WorkSession, len(sessions))
	for i, session := range sessions {
		sessionRate := decimal.Zero
		if session.HourlyRate.Valid {
			sessionRate = session.HourlyRate.Decimal
		}

		result[i] = &models.WorkSession{
			ID:              session.ID,
			ClientID:        session.ClientID,
			StartTime:       session.StartTime,
			EndTime:         nullTimeToPtr(session.EndTime),
			Description:     nullStringToPtr(session.Description),
			HourlyRate:      &sessionRate,
			FullWorkSummary: nullStringToPtr(session.FullWorkSummary),
			OutsideGit:      nullStringToPtr(session.OutsideGit),
			CreatedAt:       session.CreatedAt,
			UpdatedAt:       session.UpdatedAt,
			ClientName:      session.ClientName,
		}
	}

	return result, nil
}

func (s *SQLiteDB) ListSessionsByClient(ctx context.Context, clientName string, limit int32) ([]*models.WorkSession, error) {
	sessions, err := s.queries.ListSessionsWithDateRange(ctx, db.ListSessionsWithDateRangeParams{
		StartDate:  nil,
		EndDate:    nil,
		ClientName: clientName,
		LimitCount: int64(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions by client: %w", err)
	}

	result := make([]*models.WorkSession, len(sessions))
	for i, session := range sessions {
		sessionRate := decimal.Zero
		if session.HourlyRate.Valid {
			sessionRate = session.HourlyRate.Decimal
		}

		result[i] = &models.WorkSession{
			ID:              session.ID,
			ClientID:        session.ClientID,
			StartTime:       session.StartTime,
			EndTime:         nullTimeToPtr(session.EndTime),
			Description:     nullStringToPtr(session.Description),
			HourlyRate:      &sessionRate,
			FullWorkSummary: nullStringToPtr(session.FullWorkSummary),
			OutsideGit:      nullStringToPtr(session.OutsideGit),
			CreatedAt:       session.CreatedAt,
			UpdatedAt:       session.UpdatedAt,
			ClientName:      session.ClientName,
		}
	}

	return result, nil
}

func (s *SQLiteDB) UpdateClient(ctx context.Context, clientID string, updates *ClientUpdateDetails) (*models.Client, error) {
	client, err := s.queries.UpdateClient(ctx, db.UpdateClientParams{
		ID:             clientID,
		HourlyRate:     ptrToNullDecimal(updates.HourlyRate),
		CompanyName:    ptrToNullString(updates.CompanyName),
		ContactName:    ptrToNullString(updates.ContactName),
		Email:          ptrToNullString(updates.Email),
		Phone:          ptrToNullString(updates.Phone),
		AddressLine1:   ptrToNullString(updates.AddressLine1),
		AddressLine2:   ptrToNullString(updates.AddressLine2),
		City:           ptrToNullString(updates.City),
		State:          ptrToNullString(updates.State),
		PostalCode:     ptrToNullString(updates.PostalCode),
		Country:        ptrToNullString(updates.Country),
		Abn:            ptrToNullString(updates.Abn),
		Dir:            ptrToNullString(updates.Dir),
		RetainerAmount: ptrToNullDecimal(updates.RetainerAmount),
		RetainerHours:  ptrToNullFloat64(updates.RetainerHours),
		RetainerBasis:  ptrToNullString(updates.RetainerBasis),
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
	var startDate, endDate any
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

func nullFloat64ToPtr(nf sql.NullFloat64) *float64 {
	if nf.Valid {
		return &nf.Float64
	}
	return nil
}

func (s *SQLiteDB) convertDBClientToModel(client db.Client) *models.Client {
	var rate decimal.Decimal
	if client.HourlyRate.Valid {
		rate = client.HourlyRate.Decimal
	}
	return &models.Client{
		ID:             client.ID,
		Name:           client.Name,
		HourlyRate:     rate,
		CompanyName:    nullStringToPtr(client.CompanyName),
		ContactName:    nullStringToPtr(client.ContactName),
		Email:          nullStringToPtr(client.Email),
		Phone:          nullStringToPtr(client.Phone),
		AddressLine1:   nullStringToPtr(client.AddressLine1),
		AddressLine2:   nullStringToPtr(client.AddressLine2),
		City:           nullStringToPtr(client.City),
		State:          nullStringToPtr(client.State),
		PostalCode:     nullStringToPtr(client.PostalCode),
		Country:        nullStringToPtr(client.Country),
		Abn:            nullStringToPtr(client.Abn),
		Dir:            nullStringToPtr(client.Dir),
		RetainerAmount: nullDecimalToPtr(client.RetainerAmount),
		RetainerHours:  nullFloat64ToPtr(client.RetainerHours),
		RetainerBasis:  nullStringToPtr(client.RetainerBasis),
		CreatedAt:      client.CreatedAt,
		UpdatedAt:      client.UpdatedAt,
	}
}

func ptrToNullString(s *string) sql.NullString {
	if s != nil {
		return sql.NullString{String: *s, Valid: true}
	}
	return sql.NullString{Valid: false}
}

func ptrToNullFloat64(f *float64) sql.NullFloat64 {
	if f != nil {
		return sql.NullFloat64{Float64: *f, Valid: true}
	}
	return sql.NullFloat64{Valid: false}
}

func ptrToNullDecimal(d *decimal.Decimal) decimal.NullDecimal {
	if d != nil {
		return decimal.NullDecimal{Decimal: *d, Valid: true}
	}
	return decimal.NullDecimal{Valid: false}
}

func nullDecimalToPtr(nd decimal.NullDecimal) *decimal.Decimal {
	if nd.Valid {
		return &nd.Decimal
	}
	return nil
}

func (s *SQLiteDB) convertDBSessionToModel(session interface{}) *models.WorkSession {
	switch dbSession := session.(type) {
	case db.Session:
		rate := decimal.Zero
		if dbSession.HourlyRate.Valid {
			rate = dbSession.HourlyRate.Decimal
		}
		return &models.WorkSession{
			ID:              dbSession.ID,
			ClientID:        dbSession.ClientID,
			StartTime:       dbSession.StartTime,
			EndTime:         nullTimeToPtr(dbSession.EndTime),
			Description:     nullStringToPtr(dbSession.Description),
			HourlyRate:      &rate,
			FullWorkSummary: nullStringToPtr(dbSession.FullWorkSummary),
			OutsideGit:      nullStringToPtr(dbSession.OutsideGit),
			IncludesGst:     dbSession.IncludesGst,
			CreatedAt:       dbSession.CreatedAt,
			UpdatedAt:       dbSession.UpdatedAt,
		}
	default:
		return nil
	}
}

func (s *SQLiteDB) GetSessionsWithoutDescription(ctx context.Context, clientName *string, sessionID *string) ([]*models.WorkSession, error) {
	var name any
	if clientName != nil {
		name = *clientName
	}

	var id any
	if sessionID != nil {
		id = *sessionID
	}

	sessions, err := s.queries.GetSessionsWithoutDescription(ctx, db.GetSessionsWithoutDescriptionParams{
		ClientName: name,
		SessionID:  id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions without description: %w", err)
	}

	result := make([]*models.WorkSession, len(sessions))
	for i, session := range sessions {
		sessionRate := decimal.Zero
		if session.HourlyRate.Valid {
			sessionRate = session.HourlyRate.Decimal
		}

		result[i] = &models.WorkSession{
			ID:              session.ID,
			ClientID:        session.ClientID,
			StartTime:       session.StartTime,
			EndTime:         nullTimeToPtr(session.EndTime),
			Description:     nullStringToPtr(session.Description),
			HourlyRate:      &sessionRate,
			FullWorkSummary: nullStringToPtr(session.FullWorkSummary),
			OutsideGit:      nullStringToPtr(session.OutsideGit),
			CreatedAt:       session.CreatedAt,
			UpdatedAt:       session.UpdatedAt,
			ClientName:      session.ClientName,
		}
	}

	return result, nil
}

func (s *SQLiteDB) UpdateSessionDescription(ctx context.Context, sessionID string, description string, fullWorkSummary *string) (*models.WorkSession, error) {
	session, err := s.queries.UpdateSessionDescription(ctx, db.UpdateSessionDescriptionParams{
		ID:              sessionID,
		Description:     sql.NullString{String: description, Valid: true},
		FullWorkSummary: ptrToNullString(fullWorkSummary),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update session description: %w", err)
	}

	sessionRate := decimal.Zero
	if session.HourlyRate.Valid {
		sessionRate = session.HourlyRate.Decimal
	}

	return &models.WorkSession{
		ID:              session.ID,
		ClientID:        session.ClientID,
		StartTime:       session.StartTime,
		EndTime:         nullTimeToPtr(session.EndTime),
		Description:     nullStringToPtr(session.Description),
		HourlyRate:      &sessionRate,
		FullWorkSummary: nullStringToPtr(session.FullWorkSummary),
		OutsideGit:      nullStringToPtr(session.OutsideGit),
		CreatedAt:       session.CreatedAt,
		UpdatedAt:       session.UpdatedAt,
	}, nil
}

func (s *SQLiteDB) GetSessionByID(ctx context.Context, sessionID string) (*models.WorkSession, error) {
	session, err := s.queries.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session by ID: %w", err)
	}

	sessionRate := decimal.Zero
	if session.HourlyRate.Valid {
		sessionRate = session.HourlyRate.Decimal
	}

	return &models.WorkSession{
		ID:              session.ID,
		ClientID:        session.ClientID,
		StartTime:       session.StartTime,
		EndTime:         nullTimeToPtr(session.EndTime),
		Description:     nullStringToPtr(session.Description),
		HourlyRate:      &sessionRate,
		FullWorkSummary: nullStringToPtr(session.FullWorkSummary),
		OutsideGit:      nullStringToPtr(session.OutsideGit),
		CreatedAt:       session.CreatedAt,
		UpdatedAt:       session.UpdatedAt,
		ClientName:      session.ClientName,
	}, nil
}

func (s *SQLiteDB) UpdateSessionOutsideGit(ctx context.Context, sessionID string, outsideGit string) (*models.WorkSession, error) {
	session, err := s.queries.UpdateSessionOutsideGit(ctx, db.UpdateSessionOutsideGitParams{
		ID:         sessionID,
		OutsideGit: sql.NullString{String: outsideGit, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update session outside git: %w", err)
	}

	sessionRate := decimal.Zero
	if session.HourlyRate.Valid {
		sessionRate = session.HourlyRate.Decimal
	}

	return &models.WorkSession{
		ID:              session.ID,
		ClientID:        session.ClientID,
		StartTime:       session.StartTime,
		EndTime:         nullTimeToPtr(session.EndTime),
		Description:     nullStringToPtr(session.Description),
		HourlyRate:      &sessionRate,
		FullWorkSummary: nullStringToPtr(session.FullWorkSummary),
		OutsideGit:      nullStringToPtr(session.OutsideGit),
		CreatedAt:       session.CreatedAt,
		UpdatedAt:       session.UpdatedAt,
	}, nil
}

// Invoice methods

func (s *SQLiteDB) CreateInvoice(ctx context.Context, clientID, invoiceNumber, periodType string, periodStart, periodEnd time.Time, subtotal, gst, total decimal.Decimal) (*models.Invoice, error) {
	invoice, err := s.queries.CreateInvoice(ctx, db.CreateInvoiceParams{
		ID:              models.NewUUID(),
		ClientID:        clientID,
		InvoiceNumber:   invoiceNumber,
		PeriodType:      periodType,
		PeriodStartDate: periodStart,
		PeriodEndDate:   periodEnd,
		SubtotalAmount:  subtotal,
		GstAmount:       gst,
		TotalAmount:     total,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	return s.convertDBInvoiceToModel(invoice), nil
}

func (s *SQLiteDB) GetInvoiceByID(ctx context.Context, invoiceID string) (*models.Invoice, error) {
	invoice, err := s.queries.GetInvoiceByID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice by ID: %w", err)
	}

	return s.convertDBInvoiceRowToModel(invoice), nil
}

func (s *SQLiteDB) GetInvoiceByNumber(ctx context.Context, invoiceNumber string) (*models.Invoice, error) {
	invoice, err := s.queries.GetInvoiceByNumber(ctx, invoiceNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice by number: %w", err)
	}

	return s.convertDBInvoiceByNumberRowToModel(invoice), nil
}

func (s *SQLiteDB) ListInvoices(ctx context.Context, limit int32) ([]*models.Invoice, error) {
	invoices, err := s.queries.ListInvoices(ctx, int64(limit))
	if err != nil {
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}

	result := make([]*models.Invoice, len(invoices))
	for i, invoice := range invoices {
		result[i] = s.convertDBInvoiceListRowToModel(invoice)
	}

	return result, nil
}

func (s *SQLiteDB) GetInvoicesByClient(ctx context.Context, clientName string) ([]*models.Invoice, error) {
	invoices, err := s.queries.GetInvoicesByClient(ctx, clientName)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices by client: %w", err)
	}

	result := make([]*models.Invoice, len(invoices))
	for i, invoice := range invoices {
		result[i] = s.convertDBInvoicesByClientRowToModel(invoice)
	}

	return result, nil
}

func (s *SQLiteDB) GetInvoicesByPeriod(ctx context.Context, periodStart, periodEnd time.Time, periodType string) ([]*models.Invoice, error) {
	invoices, err := s.queries.GetInvoicesByPeriod(ctx, db.GetInvoicesByPeriodParams{
		PeriodStartDate: periodStart,
		PeriodEndDate:   periodEnd,
		PeriodType:      periodType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices by period: %w", err)
	}

	result := make([]*models.Invoice, len(invoices))
	for i, invoice := range invoices {
		result[i] = s.convertDBInvoicesByPeriodRowToModel(invoice)
	}

	return result, nil
}

func (s *SQLiteDB) DeleteInvoice(ctx context.Context, invoiceID string) error {
	err := s.queries.DeleteInvoice(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}
	return nil
}

func (s *SQLiteDB) GetSessionsForPeriodWithoutInvoice(ctx context.Context, startDate, endDate time.Time) ([]*models.WorkSession, error) {
	sessions, err := s.queries.GetSessionsForPeriodWithoutInvoice(ctx, db.GetSessionsForPeriodWithoutInvoiceParams{
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions for period without invoice: %w", err)
	}

	result := make([]*models.WorkSession, len(sessions))
	for i, session := range sessions {
		sessionRate := decimal.Zero
		if session.HourlyRate.Valid {
			sessionRate = session.HourlyRate.Decimal
		}

		result[i] = &models.WorkSession{
			ID:              session.ID,
			ClientID:        session.ClientID,
			StartTime:       session.StartTime,
			EndTime:         nullTimeToPtr(session.EndTime),
			Description:     nullStringToPtr(session.Description),
			HourlyRate:      &sessionRate,
			FullWorkSummary: nullStringToPtr(session.FullWorkSummary),
			OutsideGit:      nullStringToPtr(session.OutsideGit),
			InvoiceID:       nullStringToPtr(session.InvoiceID),
			IncludesGst:     session.IncludesGst,
			CreatedAt:       session.CreatedAt,
			UpdatedAt:       session.UpdatedAt,
			ClientName:      session.ClientName,
		}
	}

	return result, nil
}

func (s *SQLiteDB) GetSessionsByInvoiceID(ctx context.Context, invoiceID string) ([]*models.WorkSession, error) {
	sessions, err := s.queries.GetSessionsByInvoiceID(ctx, sql.NullString{String: invoiceID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions by invoice ID: %w", err)
	}

	result := make([]*models.WorkSession, len(sessions))
	for i, session := range sessions {
		sessionRate := decimal.Zero
		if session.HourlyRate.Valid {
			sessionRate = session.HourlyRate.Decimal
		}

		result[i] = &models.WorkSession{
			ID:              session.ID,
			ClientID:        session.ClientID,
			StartTime:       session.StartTime,
			EndTime:         nullTimeToPtr(session.EndTime),
			Description:     nullStringToPtr(session.Description),
			HourlyRate:      &sessionRate,
			FullWorkSummary: nullStringToPtr(session.FullWorkSummary),
			OutsideGit:      nullStringToPtr(session.OutsideGit),
			InvoiceID:       nullStringToPtr(session.InvoiceID),
			CreatedAt:       session.CreatedAt,
			UpdatedAt:       session.UpdatedAt,
			ClientName:      session.ClientName,
		}
	}

	return result, nil
}

func (s *SQLiteDB) UpdateSessionInvoiceID(ctx context.Context, sessionID, invoiceID string) error {
	err := s.queries.UpdateSessionInvoiceID(ctx, db.UpdateSessionInvoiceIDParams{
		InvoiceID: sql.NullString{String: invoiceID, Valid: true},
		SessionID: sessionID,
	})
	if err != nil {
		return fmt.Errorf("failed to update session invoice ID: %w", err)
	}
	return nil
}

func (s *SQLiteDB) ClearSessionInvoiceIDs(ctx context.Context, invoiceID string) error {
	err := s.queries.ClearSessionInvoiceIDs(ctx, sql.NullString{String: invoiceID, Valid: true})
	if err != nil {
		return fmt.Errorf("failed to clear session invoice IDs: %w", err)
	}
	return nil
}

func (s *SQLiteDB) GetSessionsForPeriodWithoutInvoiceByClient(ctx context.Context, startDate, endDate time.Time, clientName string) ([]*models.WorkSession, error) {
	sessions, err := s.queries.GetSessionsForPeriodWithoutInvoiceByClient(ctx, db.GetSessionsForPeriodWithoutInvoiceByClientParams{
		StartDate:  startDate,
		EndDate:    endDate,
		ClientName: clientName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions for period without invoice by client: %w", err)
	}

	result := make([]*models.WorkSession, len(sessions))
	for i, session := range sessions {
		sessionRate := decimal.Zero
		if session.HourlyRate.Valid {
			sessionRate = session.HourlyRate.Decimal
		}

		result[i] = &models.WorkSession{
			ID:              session.ID,
			ClientID:        session.ClientID,
			StartTime:       session.StartTime,
			EndTime:         nullTimeToPtr(session.EndTime),
			Description:     nullStringToPtr(session.Description),
			HourlyRate:      &sessionRate,
			FullWorkSummary: nullStringToPtr(session.FullWorkSummary),
			OutsideGit:      nullStringToPtr(session.OutsideGit),
			InvoiceID:       nullStringToPtr(session.InvoiceID),
			IncludesGst:     session.IncludesGst,
			CreatedAt:       session.CreatedAt,
			UpdatedAt:       session.UpdatedAt,
			ClientName:      session.ClientName,
		}
	}

	return result, nil
}

func (s *SQLiteDB) GetInvoicesByPeriodAndClient(ctx context.Context, periodStart, periodEnd time.Time, periodType, clientName string) ([]*models.Invoice, error) {
	invoices, err := s.queries.GetInvoicesByPeriodAndClient(ctx, db.GetInvoicesByPeriodAndClientParams{
		PeriodStartDate: periodStart,
		PeriodEndDate:   periodEnd,
		PeriodType:      periodType,
		ClientName:      clientName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices by period and client: %w", err)
	}

	result := make([]*models.Invoice, len(invoices))
	for i, invoice := range invoices {
		result[i] = s.convertDBInvoicesByPeriodAndClientRowToModel(invoice)
	}

	return result, nil
}

func (s *SQLiteDB) PayInvoice(ctx context.Context, param db.PayInvoiceParams) error {
	err := s.queries.PayInvoice(ctx, param)
	if err != nil {
		return fmt.Errorf("failed to pay invoice: %w", err)
	}
	return nil
}

func (s *SQLiteDB) convertDBInvoicesByPeriodAndClientRowToModel(invoice db.GetInvoicesByPeriodAndClientRow) *models.Invoice {
	paymentDate := convertPaymentDate(invoice.PaymentDate)

	return &models.Invoice{
		ID:              invoice.ID,
		ClientID:        invoice.ClientID,
		InvoiceNumber:   invoice.InvoiceNumber,
		PeriodType:      invoice.PeriodType,
		PeriodStartDate: invoice.PeriodStartDate,
		PeriodEndDate:   invoice.PeriodEndDate,
		SubtotalAmount:  invoice.SubtotalAmount,
		GstAmount:       invoice.GstAmount,
		TotalAmount:     invoice.TotalAmount,
		GeneratedDate:   invoice.GeneratedDate,
		AmountPaid:      decimal.NewFromFloat(invoice.AmountPaid),
		PaymentDate:     paymentDate,
		CreatedAt:       invoice.CreatedAt,
		UpdatedAt:       invoice.UpdatedAt,
		ClientName:      invoice.ClientName,
	}
}

// Helper function to convert interface{} to *time.Time
func convertPaymentDate(paymentDate interface{}) *time.Time {
	if paymentDate == nil {
		return nil
	}

	if val, ok := paymentDate.(time.Time); ok {
		return &val
	}

	if val, ok := paymentDate.(string); ok {
		// Try various timestamp formats that SQLite might return
		formats := []string{
			"2006-01-02 15:04:05-07:00",
			"2006-01-02 15:04:05+10:00",
			"2006-01-02 15:04:05Z07:00",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}

		for _, format := range formats {
			if parsedTime, err := time.Parse(format, val); err == nil {
				return &parsedTime
			}
		}
	}

	return nil
}

// Helper methods for converting DB types to models

func (s *SQLiteDB) convertDBInvoiceToModel(invoice db.Invoice) *models.Invoice {
	return &models.Invoice{
		ID:              invoice.ID,
		ClientID:        invoice.ClientID,
		InvoiceNumber:   invoice.InvoiceNumber,
		PeriodType:      invoice.PeriodType,
		PeriodStartDate: invoice.PeriodStartDate,
		PeriodEndDate:   invoice.PeriodEndDate,
		SubtotalAmount:  invoice.SubtotalAmount,
		GstAmount:       invoice.GstAmount,
		TotalAmount:     invoice.TotalAmount,
		GeneratedDate:   invoice.GeneratedDate,
		CreatedAt:       invoice.CreatedAt,
		UpdatedAt:       invoice.UpdatedAt,
	}
}

func (s *SQLiteDB) convertDBInvoiceRowToModel(invoice db.GetInvoiceByIDRow) *models.Invoice {
	paymentDate := convertPaymentDate(invoice.PaymentDate)

	return &models.Invoice{
		ID:              invoice.ID,
		ClientID:        invoice.ClientID,
		InvoiceNumber:   invoice.InvoiceNumber,
		PeriodType:      invoice.PeriodType,
		PeriodStartDate: invoice.PeriodStartDate,
		PeriodEndDate:   invoice.PeriodEndDate,
		SubtotalAmount:  invoice.SubtotalAmount,
		GstAmount:       invoice.GstAmount,
		TotalAmount:     invoice.TotalAmount,
		GeneratedDate:   invoice.GeneratedDate,
		AmountPaid:      decimal.NewFromFloat(invoice.AmountPaid),
		PaymentDate:     paymentDate,
		CreatedAt:       invoice.CreatedAt,
		UpdatedAt:       invoice.UpdatedAt,
		ClientName:      invoice.ClientName,
	}
}

func (s *SQLiteDB) convertDBInvoiceListRowToModel(invoice db.ListInvoicesRow) *models.Invoice {
	paymentDate := convertPaymentDate(invoice.PaymentDate)

	return &models.Invoice{
		ID:              invoice.ID,
		ClientID:        invoice.ClientID,
		InvoiceNumber:   invoice.InvoiceNumber,
		PeriodType:      invoice.PeriodType,
		PeriodStartDate: invoice.PeriodStartDate,
		PeriodEndDate:   invoice.PeriodEndDate,
		SubtotalAmount:  invoice.SubtotalAmount,
		GstAmount:       invoice.GstAmount,
		TotalAmount:     invoice.TotalAmount,
		GeneratedDate:   invoice.GeneratedDate,
		AmountPaid:      decimal.NewFromFloat(invoice.AmountPaid),
		PaymentDate:     paymentDate,
		CreatedAt:       invoice.CreatedAt,
		UpdatedAt:       invoice.UpdatedAt,
		ClientName:      invoice.ClientName,
	}
}

func (s *SQLiteDB) convertDBInvoicesByClientRowToModel(invoice db.GetInvoicesByClientRow) *models.Invoice {
	paymentDate := convertPaymentDate(invoice.PaymentDate)

	return &models.Invoice{
		ID:              invoice.ID,
		ClientID:        invoice.ClientID,
		InvoiceNumber:   invoice.InvoiceNumber,
		PeriodType:      invoice.PeriodType,
		PeriodStartDate: invoice.PeriodStartDate,
		PeriodEndDate:   invoice.PeriodEndDate,
		SubtotalAmount:  invoice.SubtotalAmount,
		GstAmount:       invoice.GstAmount,
		TotalAmount:     invoice.TotalAmount,
		GeneratedDate:   invoice.GeneratedDate,
		AmountPaid:      decimal.NewFromFloat(invoice.AmountPaid),
		PaymentDate:     paymentDate,
		CreatedAt:       invoice.CreatedAt,
		UpdatedAt:       invoice.UpdatedAt,
		ClientName:      invoice.ClientName,
	}
}

func (s *SQLiteDB) convertDBInvoicesByPeriodRowToModel(invoice db.GetInvoicesByPeriodRow) *models.Invoice {
	paymentDate := convertPaymentDate(invoice.PaymentDate)

	return &models.Invoice{
		ID:              invoice.ID,
		ClientID:        invoice.ClientID,
		InvoiceNumber:   invoice.InvoiceNumber,
		PeriodType:      invoice.PeriodType,
		PeriodStartDate: invoice.PeriodStartDate,
		PeriodEndDate:   invoice.PeriodEndDate,
		SubtotalAmount:  invoice.SubtotalAmount,
		GstAmount:       invoice.GstAmount,
		TotalAmount:     invoice.TotalAmount,
		GeneratedDate:   invoice.GeneratedDate,
		AmountPaid:      decimal.NewFromFloat(invoice.AmountPaid),
		PaymentDate:     paymentDate,
		CreatedAt:       invoice.CreatedAt,
		UpdatedAt:       invoice.UpdatedAt,
		ClientName:      invoice.ClientName,
	}
}

func (s *SQLiteDB) convertDBInvoiceByNumberRowToModel(invoice db.GetInvoiceByNumberRow) *models.Invoice {
	paymentDate := convertPaymentDate(invoice.PaymentDate)

	return &models.Invoice{
		ID:              invoice.ID,
		ClientID:        invoice.ClientID,
		InvoiceNumber:   invoice.InvoiceNumber,
		PeriodType:      invoice.PeriodType,
		PeriodStartDate: invoice.PeriodStartDate,
		PeriodEndDate:   invoice.PeriodEndDate,
		SubtotalAmount:  invoice.SubtotalAmount,
		GstAmount:       invoice.GstAmount,
		TotalAmount:     invoice.TotalAmount,
		GeneratedDate:   invoice.GeneratedDate,
		AmountPaid:      decimal.NewFromFloat(invoice.AmountPaid),
		PaymentDate:     paymentDate,
		CreatedAt:       invoice.CreatedAt,
		UpdatedAt:       invoice.UpdatedAt,
		ClientName:      invoice.ClientName,
	}
}

// Expense operations
func (s *SQLiteDB) CreateExpense(ctx context.Context, amount decimal.Decimal, expenseDate time.Time, reference *string, clientID *string, invoiceID *string, description *string) (*models.Expense, error) {
	expense, err := s.queries.CreateExpense(ctx, db.CreateExpenseParams{
		ID:          models.NewUUID(),
		Amount:      amount,
		ExpenseDate: expenseDate,
		Reference:   ptrToNullString(reference),
		ClientID:    ptrToNullString(clientID),
		InvoiceID:   ptrToNullString(invoiceID),
		Description: ptrToNullString(description),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}

	return s.convertDBExpenseToModel(expense), nil
}

func (s *SQLiteDB) GetExpenseByID(ctx context.Context, expenseID string) (*models.Expense, error) {
	expense, err := s.queries.GetExpenseByID(ctx, expenseID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get expense by ID: %w", err)
	}

	return s.convertDBExpenseToModel(expense), nil
}

func (s *SQLiteDB) ListExpenses(ctx context.Context) ([]*models.Expense, error) {
	expenses, err := s.queries.ListExpenses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list expenses: %w", err)
	}

	result := make([]*models.Expense, len(expenses))
	for i, expense := range expenses {
		result[i] = s.convertDBExpenseToModel(expense)
	}

	return result, nil
}

func (s *SQLiteDB) ListExpensesByClient(ctx context.Context, clientID string) ([]*models.Expense, error) {
	expenses, err := s.queries.ListExpensesByClient(ctx, sql.NullString{String: clientID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list expenses by client: %w", err)
	}

	result := make([]*models.Expense, len(expenses))
	for i, expense := range expenses {
		result[i] = s.convertDBExpenseToModel(expense)
	}

	return result, nil
}

func (s *SQLiteDB) ListExpensesByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*models.Expense, error) {
	expenses, err := s.queries.ListExpensesByDateRange(ctx, db.ListExpensesByDateRangeParams{
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list expenses by date range: %w", err)
	}

	result := make([]*models.Expense, len(expenses))
	for i, expense := range expenses {
		result[i] = s.convertDBExpenseToModel(expense)
	}

	return result, nil
}

func (s *SQLiteDB) ListExpensesByClientAndDateRange(ctx context.Context, clientID string, startDate, endDate time.Time) ([]*models.Expense, error) {
	expenses, err := s.queries.ListExpensesByClientAndDateRange(ctx, db.ListExpensesByClientAndDateRangeParams{
		ClientID:  sql.NullString{String: clientID, Valid: true},
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list expenses by client and date range: %w", err)
	}

	result := make([]*models.Expense, len(expenses))
	for i, expense := range expenses {
		result[i] = s.convertDBExpenseToModel(expense)
	}

	return result, nil
}

func (s *SQLiteDB) UpdateExpense(ctx context.Context, expenseID string, amount *decimal.Decimal, expenseDate *time.Time, reference *string, clientID *string, invoiceID *string, description *string) (*models.Expense, error) {
	// Get current expense to preserve existing values
	current, err := s.GetExpenseByID(ctx, expenseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current expense: %w", err)
	}

	updateParams := db.UpdateExpenseParams{
		ID:          expenseID,
		Amount:      current.Amount,
		ExpenseDate: sql.NullTime{Time: current.ExpenseDate, Valid: true},
		Reference:   ptrToNullString(current.Reference),
		ClientID:    ptrToNullString(current.ClientID),
		InvoiceID:   ptrToNullString(current.InvoiceID),
		Description: ptrToNullString(current.Description),
	}

	if amount != nil {
		updateParams.Amount = *amount
	}
	if expenseDate != nil {
		updateParams.ExpenseDate = sql.NullTime{Time: *expenseDate, Valid: true}
	}
	if reference != nil {
		updateParams.Reference = ptrToNullString(reference)
	}
	if clientID != nil {
		updateParams.ClientID = ptrToNullString(clientID)
	}
	if invoiceID != nil {
		updateParams.InvoiceID = ptrToNullString(invoiceID)
	}
	if description != nil {
		updateParams.Description = ptrToNullString(description)
	}

	expense, err := s.queries.UpdateExpense(ctx, updateParams)
	if err != nil {
		return nil, fmt.Errorf("failed to update expense: %w", err)
	}

	return s.convertDBExpenseToModel(expense), nil
}

func (s *SQLiteDB) DeleteExpense(ctx context.Context, expenseID string) error {
	err := s.queries.DeleteExpense(ctx, expenseID)
	if err != nil {
		return fmt.Errorf("failed to delete expense: %w", err)
	}
	return nil
}

func (s *SQLiteDB) GetExpensesByInvoiceID(ctx context.Context, invoiceID string) ([]*models.Expense, error) {
	expenses, err := s.queries.GetExpensesByInvoiceID(ctx, sql.NullString{String: invoiceID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses by invoice ID: %w", err)
	}

	result := make([]*models.Expense, len(expenses))
	for i, expense := range expenses {
		result[i] = s.convertDBExpenseToModel(expense)
	}

	return result, nil
}

func (s *SQLiteDB) GetExpensesWithoutInvoiceByClient(ctx context.Context, clientID string) ([]*models.Expense, error) {
	expenses, err := s.queries.GetExpensesWithoutInvoiceByClient(ctx, sql.NullString{String: clientID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses without invoice by client: %w", err)
	}

	result := make([]*models.Expense, len(expenses))
	for i, expense := range expenses {
		result[i] = s.convertDBExpenseToModel(expense)
	}

	return result, nil
}

func (s *SQLiteDB) GetExpensesWithoutInvoiceByClientAndDateRange(ctx context.Context, clientID string, startDate, endDate time.Time) ([]*models.Expense, error) {
	expenses, err := s.queries.GetExpensesWithoutInvoiceByClientAndDateRange(ctx, db.GetExpensesWithoutInvoiceByClientAndDateRangeParams{
		ClientID:  sql.NullString{String: clientID, Valid: true},
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses without invoice by client and date range: %w", err)
	}

	result := make([]*models.Expense, len(expenses))
	for i, expense := range expenses {
		result[i] = s.convertDBExpenseToModel(expense)
	}

	return result, nil
}

func (s *SQLiteDB) UpdateExpenseInvoiceID(ctx context.Context, expenseID string, invoiceID *string) error {
	err := s.queries.UpdateExpenseInvoiceID(ctx, db.UpdateExpenseInvoiceIDParams{
		ID:        expenseID,
		InvoiceID: ptrToNullString(invoiceID),
	})
	if err != nil {
		return fmt.Errorf("failed to update expense invoice ID: %w", err)
	}
	return nil
}

func (s *SQLiteDB) ClearExpenseInvoiceIDs(ctx context.Context, invoiceID string) error {
	err := s.queries.ClearExpenseInvoiceIDs(ctx, sql.NullString{String: invoiceID, Valid: true})
	if err != nil {
		return fmt.Errorf("failed to clear expense invoice IDs: %w", err)
	}
	return nil
}

func (s *SQLiteDB) convertDBExpenseToModel(expense db.Expense) *models.Expense {
	return &models.Expense{
		ID:          expense.ID,
		Amount:      expense.Amount,
		ExpenseDate: expense.ExpenseDate,
		Reference:   nullStringToPtr(expense.Reference),
		ClientID:    nullStringToPtr(expense.ClientID),
		InvoiceID:   nullStringToPtr(expense.InvoiceID),
		Description: nullStringToPtr(expense.Description),
		CreatedAt:   expense.CreatedAt,
		UpdatedAt:   expense.UpdatedAt,
	}
}
