package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

const dbFileName = "accounts.db"
const oldJSONName = "user_backup.json"

// Store persists account -> refresh token in SQLite next to the executable.
type Store struct {
	path string
	db   *sql.DB
	mu   sync.Mutex
}

func New(baseDir string) (*Store, error) {
	if baseDir == "" {
		baseDir, _ = os.Getwd()
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(baseDir, dbFileName)
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// Reasonable defaults for a local desktop DB
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	s := &Store{path: path, db: db}
	if err := s.init(); err != nil {
		_ = db.Close()
		return nil, err
	}
	// One-time import from older JSON backup if present
	_ = s.migrateOldJSON(filepath.Join(baseDir, oldJSONName))
	return s, nil
}

func (s *Store) Path() string { return s.path }

func (s *Store) init() error {
	stmts := []string{
		`PRAGMA journal_mode=WAL;`,
		`PRAGMA synchronous=NORMAL;`,
		`PRAGMA foreign_keys=ON;`,
		`CREATE TABLE IF NOT EXISTS accounts (
			name TEXT PRIMARY KEY COLLATE NOCASE NOT NULL,
			token TEXT NOT NULL,
			updated_at INTEGER NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_accounts_updated ON accounts(updated_at DESC);`,
	}
	for _, q := range stmts {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("sqlite init: %w", err)
		}
	}
	return nil
}

func (s *Store) migrateOldJSON(jsonPath string) error {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil // nothing to import
	}
	m := map[string]string{}
	if err := json.Unmarshal(data, &m); err != nil || len(m) == 0 {
		return err
	}
	if err := s.Merge(m); err != nil {
		return err
	}
	// Keep a backup copy, remove active JSON so we don't re-import forever
	_ = os.Rename(jsonPath, jsonPath+".migrated")
	return nil
}

// Load returns all accounts as name -> token.
func (s *Store) Load() (map[string]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`SELECT name, token FROM accounts ORDER BY name COLLATE NOCASE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]string{}
	for rows.Next() {
		var name, token string
		if err := rows.Scan(&name, &token); err != nil {
			return nil, err
		}
		out[name] = token
	}
	return out, rows.Err()
}

// Get returns one token by account name.
func (s *Store) Get(account string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var token string
	err := s.db.QueryRow(`SELECT token FROM accounts WHERE name = ? COLLATE NOCASE`, account).Scan(&token)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return token, true, nil
}

func (s *Store) Put(account, token string) error {
	account = strings.TrimSpace(account)
	if account == "" || token == "" {
		return fmt.Errorf("empty account or token")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(
		`INSERT INTO accounts(name, token, updated_at) VALUES(?, ?, ?)
		 ON CONFLICT(name) DO UPDATE SET token=excluded.token, updated_at=excluded.updated_at`,
		account, token, time.Now().Unix(),
	)
	return err
}

func (s *Store) Delete(account string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM accounts WHERE name = ? COLLATE NOCASE`, account)
	return err
}

func (s *Store) Merge(extra map[string]string) error {
	if len(extra) == 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(
		`INSERT INTO accounts(name, token, updated_at) VALUES(?, ?, ?)
		 ON CONFLICT(name) DO UPDATE SET token=excluded.token, updated_at=excluded.updated_at`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().Unix()
	for name, tok := range extra {
		name = strings.TrimSpace(name)
		if name == "" || tok == "" {
			continue
		}
		if _, err := stmt.Exec(name, tok, now); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}
