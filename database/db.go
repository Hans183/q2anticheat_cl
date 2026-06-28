package database

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

// DB wraps the SQLite database connection
type DB struct {
	conn *sql.DB
}

// ScreenshotRecord represents a stored screenshot metadata
type ScreenshotRecord struct {
	ID         int64
	ServerAddr string
	PlayerIP   string
	PlayerName string
	ClientID   int
	Width      int
	Height     int
	FilePath   string
	FileSize   int64
	Timestamp  time.Time
	Reviewed   bool
	Notes      string
}

// ViolationRecord represents an anticheat violation
type ViolationRecord struct {
	ID         int64
	ServerAddr string
	PlayerIP   string
	PlayerName string
	ClientID   int
	Type       string
	Reason     string
	Details    string
	Timestamp  time.Time
}

// AdminRecord represents a dashboard admin user
type AdminRecord struct {
	ID        int64
	Username  string
	Password  string
	CreatedAt time.Time
}

// SessionRecord represents a login session
type SessionRecord struct {
	Token     string
	AdminID   int64
	CreatedAt time.Time
	ExpiresAt time.Time
}

// New creates a new database connection and initializes schema
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	log.Printf("[DB] Connected to %s", dbPath)
	return db, nil
}

// migrate creates the database schema
func (db *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS screenshots (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		server_addr TEXT NOT NULL,
		player_ip   TEXT NOT NULL,
		player_name TEXT NOT NULL,
		client_id   INTEGER,
		width       INTEGER NOT NULL,
		height      INTEGER NOT NULL,
		file_path   TEXT NOT NULL,
		file_size   INTEGER NOT NULL,
		timestamp   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		reviewed    BOOLEAN DEFAULT 0,
		notes       TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_screenshots_player ON screenshots(player_ip);
	CREATE INDEX IF NOT EXISTS idx_screenshots_date ON screenshots(timestamp);
	CREATE INDEX IF NOT EXISTS idx_screenshots_server ON screenshots(server_addr);

	CREATE TABLE IF NOT EXISTS violations (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		server_addr TEXT NOT NULL,
		player_ip   TEXT NOT NULL,
		player_name TEXT NOT NULL,
		client_id   INTEGER,
		type        TEXT NOT NULL,
		reason      TEXT NOT NULL,
		details     TEXT,
		timestamp   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_violations_player ON violations(player_ip);
	CREATE INDEX IF NOT EXISTS idx_violations_date ON violations(timestamp);
	CREATE INDEX IF NOT EXISTS idx_violations_type ON violations(type);

	CREATE TABLE IF NOT EXISTS admins (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		username   TEXT UNIQUE NOT NULL,
		password   TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS sessions (
		token      TEXT PRIMARY KEY,
		admin_id   INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME NOT NULL,
		FOREIGN KEY (admin_id) REFERENCES admins(id)
	);
	`

	_, err := db.conn.Exec(schema)
	return err
}

// EnsureDefaultAdmin creates the default admin user if none exists
func (db *DB) EnsureDefaultAdmin(username, passwordHash string) error {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM admins").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err = db.conn.Exec("INSERT INTO admins (username, password) VALUES (?, ?)",
		username, passwordHash)
	return err
}

// GetAdminByUsername retrieves an admin by username
func (db *DB) GetAdminByUsername(username string) (*AdminRecord, error) {
	admin := &AdminRecord{}
	err := db.conn.QueryRow(
		"SELECT id, username, password, created_at FROM admins WHERE username = ?",
		username).Scan(&admin.ID, &admin.Username, &admin.Password, &admin.CreatedAt)
	if err != nil {
		return nil, err
	}
	return admin, nil
}

// CreateSession creates a new login session
func (db *DB) CreateSession(adminID int64) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	_, err := db.conn.Exec(
		"INSERT INTO sessions (token, admin_id, expires_at) VALUES (?, ?, ?)",
		token, adminID, time.Now().Add(24*time.Hour))
	return token, err
}

// GetSession retrieves a valid session by token
func (db *DB) GetSession(token string) (*SessionRecord, error) {
	sess := &SessionRecord{}
	err := db.conn.QueryRow(
		"SELECT token, admin_id, created_at, expires_at FROM sessions WHERE token = ? AND expires_at > ?",
		token, time.Now()).Scan(&sess.Token, &sess.AdminID, &sess.CreatedAt, &sess.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

// DeleteSession removes a session
func (db *DB) DeleteSession(token string) error {
	_, err := db.conn.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}

// CleanupSessions removes expired sessions
func (db *DB) CleanupSessions() {
	db.conn.Exec("DELETE FROM sessions WHERE expires_at <= ?", time.Now())
}

// InsertScreenshot stores a new screenshot record
func (db *DB) InsertScreenshot(record *ScreenshotRecord) (int64, error) {
	result, err := db.conn.Exec(`
		INSERT INTO screenshots (server_addr, player_ip, player_name, client_id,
			width, height, file_path, file_size, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.ServerAddr, record.PlayerIP, record.PlayerName, record.ClientID,
		record.Width, record.Height, record.FilePath, record.FileSize, record.Timestamp)
	if err != nil {
		return 0, fmt.Errorf("insert screenshot: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}

	return id, nil
}

// GetScreenshot retrieves a screenshot by ID
func (db *DB) GetScreenshot(id int64) (*ScreenshotRecord, error) {
	record := &ScreenshotRecord{}
	var ts string
	var notes sql.NullString
	var clientID sql.NullInt64
	err := db.conn.QueryRow(`
		SELECT id, server_addr, player_ip, player_name, client_id,
			width, height, file_path, file_size, timestamp, reviewed, notes
		FROM screenshots WHERE id = ?`, id).Scan(
		&record.ID, &record.ServerAddr, &record.PlayerIP, &record.PlayerName,
		&clientID, &record.Width, &record.Height, &record.FilePath,
		&record.FileSize, &ts, &record.Reviewed, &notes)
	if err != nil {
		return nil, fmt.Errorf("get screenshot: %w", err)
	}
	if clientID.Valid {
		record.ClientID = int(clientID.Int64)
	}
	if notes.Valid {
		record.Notes = notes.String
	}
	record.Timestamp, _ = time.Parse("2006-01-02T15:04:05Z", ts)
	if record.Timestamp.IsZero() {
		record.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
	}
	return record, nil
}

// GetScreenshots retrieves screenshots with optional filters
func (db *DB) GetScreenshots(playerIP, dateFrom, dateTo string, unreviewedOnly bool, page, perPage int) ([]*ScreenshotRecord, int, error) {
	where := "1=1"
	args := []interface{}{}

	if playerIP != "" {
		where += " AND player_ip LIKE ?"
		args = append(args, "%"+playerIP+"%")
	}
	if dateFrom != "" {
		where += " AND timestamp >= ?"
		args = append(args, dateFrom)
	}
	if dateTo != "" {
		where += " AND timestamp <= ?"
		args = append(args, dateTo+" 23:59:59")
	}
	if unreviewedOnly {
		where += " AND reviewed = 0"
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM screenshots WHERE %s", where)
	err := db.conn.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT id, server_addr, player_ip, player_name, client_id,
			width, height, file_path, file_size, timestamp, reviewed, notes
		FROM screenshots WHERE %s
		ORDER BY timestamp DESC LIMIT ? OFFSET ?`, where)
	args = append(args, perPage, offset)

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	records, err := scanScreenshots(rows)
	return records, total, err
}

// GetUnreviewedScreenshots retrieves screenshots that haven't been reviewed
func (db *DB) GetUnreviewedScreenshots(limit int) ([]*ScreenshotRecord, error) {
	rows, err := db.conn.Query(`
		SELECT id, server_addr, player_ip, player_name, client_id,
			width, height, file_path, file_size, timestamp, reviewed, notes
		FROM screenshots WHERE reviewed = 0
		ORDER BY timestamp DESC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("query screenshots: %w", err)
	}
	defer rows.Close()

	return scanScreenshots(rows)
}

// MarkReviewed marks a screenshot as reviewed
func (db *DB) MarkReviewed(id int64, notes string) error {
	_, err := db.conn.Exec(`
		UPDATE screenshots SET reviewed = 1, notes = ? WHERE id = ?`, notes, id)
	return err
}

// InsertViolation stores a new violation record
func (db *DB) InsertViolation(record *ViolationRecord) (int64, error) {
	result, err := db.conn.Exec(`
		INSERT INTO violations (server_addr, player_ip, player_name, client_id,
			type, reason, details, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		record.ServerAddr, record.PlayerIP, record.PlayerName, record.ClientID,
		record.Type, record.Reason, record.Details, record.Timestamp)
	if err != nil {
		return 0, fmt.Errorf("insert violation: %w", err)
	}
	return result.LastInsertId()
}

// GetViolations retrieves violations with optional filters
func (db *DB) GetViolations(playerIP, vType, dateFrom, dateTo string, page, perPage int) ([]*ViolationRecord, int, error) {
	where := "1=1"
	args := []interface{}{}

	if playerIP != "" {
		where += " AND player_ip LIKE ?"
		args = append(args, "%"+playerIP+"%")
	}
	if vType != "" {
		where += " AND type = ?"
		args = append(args, vType)
	}
	if dateFrom != "" {
		where += " AND timestamp >= ?"
		args = append(args, dateFrom)
	}
	if dateTo != "" {
		where += " AND timestamp <= ?"
		args = append(args, dateTo+" 23:59:59")
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM violations WHERE %s", where)
	err := db.conn.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT id, server_addr, player_ip, player_name, client_id,
			type, reason, details, timestamp
		FROM violations WHERE %s
		ORDER BY timestamp DESC LIMIT ? OFFSET ?`, where)
	args = append(args, perPage, offset)

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	records, err := scanViolations(rows)
	return records, total, err
}

// GetStats returns statistics about stored data
func (db *DB) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalCount int
	db.conn.QueryRow("SELECT COUNT(*) FROM screenshots").Scan(&totalCount)
	stats["total_screenshots"] = totalCount

	var unreviewedCount int
	db.conn.QueryRow("SELECT COUNT(*) FROM screenshots WHERE reviewed = 0").Scan(&unreviewedCount)
	stats["unreviewed_screenshots"] = unreviewedCount

	var totalSize int64
	db.conn.QueryRow("SELECT COALESCE(SUM(file_size), 0) FROM screenshots").Scan(&totalSize)
	stats["total_size"] = totalSize

	var totalViolations int
	db.conn.QueryRow("SELECT COUNT(*) FROM violations").Scan(&totalViolations)
	stats["total_violations"] = totalViolations

	var todayViolations int
	db.conn.QueryRow("SELECT COUNT(*) FROM violations WHERE timestamp >= date('now')").Scan(&todayViolations)
	stats["today_violations"] = todayViolations

	// Violations per day for last 7 days
	rows, err := db.conn.Query(`
		SELECT date(timestamp) as day, COUNT(*) as cnt
		FROM violations
		WHERE timestamp >= date('now', '-7 days')
		GROUP BY day ORDER BY day`)
	if err == nil {
		defer rows.Close()
		dailyViolations := make([]map[string]interface{}, 0)
		for rows.Next() {
			var day string
			var cnt int
			if rows.Scan(&day, &cnt) == nil {
				dailyViolations = append(dailyViolations, map[string]interface{}{
					"day": day, "count": cnt,
				})
			}
		}
		stats["daily_violations"] = dailyViolations
	}

	return stats, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

func scanScreenshots(rows *sql.Rows) ([]*ScreenshotRecord, error) {
	var records []*ScreenshotRecord
	for rows.Next() {
		record := &ScreenshotRecord{}
		var ts string
		var notes sql.NullString
		var clientID sql.NullInt64
		err := rows.Scan(
			&record.ID, &record.ServerAddr, &record.PlayerIP, &record.PlayerName,
			&clientID, &record.Width, &record.Height, &record.FilePath,
			&record.FileSize, &ts, &record.Reviewed, &notes)
		if err != nil {
			return nil, err
		}
		if clientID.Valid {
			record.ClientID = int(clientID.Int64)
		}
		if notes.Valid {
			record.Notes = notes.String
		}
		record.Timestamp, _ = time.Parse("2006-01-02T15:04:05Z", ts)
		if record.Timestamp.IsZero() {
			record.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func scanViolations(rows *sql.Rows) ([]*ViolationRecord, error) {
	var records []*ViolationRecord
	for rows.Next() {
		record := &ViolationRecord{}
		var ts string
		var details sql.NullString
		var clientID sql.NullInt64
		err := rows.Scan(
			&record.ID, &record.ServerAddr, &record.PlayerIP, &record.PlayerName,
			&clientID, &record.Type, &record.Reason, &details,
			&ts)
		if err != nil {
			return nil, err
		}
		if clientID.Valid {
			record.ClientID = int(clientID.Int64)
		}
		if details.Valid {
			record.Details = details.String
		}
		record.Timestamp, _ = time.Parse("2006-01-02T15:04:05Z", ts)
		if record.Timestamp.IsZero() {
			record.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
		}
		records = append(records, record)
	}
	return records, rows.Err()
}
