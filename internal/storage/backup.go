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

// Store persists account -> DPAPI-sealed refresh token in SQLite.
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
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	s := &Store{path: path, db: db}
	if err := s.init(); err != nil {
		_ = db.Close()
		return nil, err
	}
	_ = s.migrateOldJSON(filepath.Join(baseDir, oldJSONName))
	_ = s.resealPlaintextRows()
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

// resealPlaintextRows migrates any old plaintext JWT rows to DPAPI blobs.
func (s *Store) resealPlaintextRows() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`SELECT name, token FROM accounts`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type pair struct{ name, tok string }
	var todo []pair
	for rows.Next() {
		var name, tok string
		if err := rows.Scan(&name, &tok); err != nil {
			return err
		}
		if !strings.HasPrefix(tok, tokenPrefix) {
			todo = append(todo, pair{name, tok})
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, p := range todo {
		sealed, err := sealToken(p.tok)
		if err != nil {
			continue
		}
		_, _ = s.db.Exec(
			`UPDATE accounts SET token=?, updated_at=? WHERE name=? COLLATE NOCASE`,
			sealed, time.Now().Unix(), p.name,
		)
	}
	return nil
}

func (s *Store) migrateOldJSON(jsonPath string) error {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil
	}
	m := map[string]string{}
	if err := json.Unmarshal(data, &m); err != nil || len(m) == 0 {
		return err
	}
	if err := s.Merge(m); err != nil {
		return err
	}
	_ = os.Rename(jsonPath, jsonPath+".migrated")
	return nil
}

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
		var name, stored string
		if err := rows.Scan(&name, &stored); err != nil {
			return nil, err
		}
		plain, err := openToken(stored)
		if err != nil {
			// skip undecryptable rows (other user / corrupted)
			continue
		}
		out[name] = plain
	}
	return out, rows.Err()
}

func (s *Store) Get(account string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var stored string
	err := s.db.QueryRow(`SELECT token FROM accounts WHERE name = ? COLLATE NOCASE`, account).Scan(&stored)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	plain, err := openToken(stored)
	if err != nil {
		return "", false, err
	}
	return plain, true, nil
}

func (s *Store) Put(account, token string) error {
	account = strings.TrimSpace(account)
	if account == "" || token == "" {
		return fmt.Errorf("empty account or token")
	}
	sealed, err := sealToken(token)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err = s.db.Exec(
		`INSERT INTO accounts(name, token, updated_at) VALUES(?, ?, ?)
		 ON CONFLICT(name) DO UPDATE SET token=excluded.token, updated_at=excluded.updated_at`,
		account, sealed, time.Now().Unix(),
	)
	return err
}

func (s *Store) Delete(account string) error {
	account = strings.TrimSpace(account)
	if account == "" {
		return fmt.Errorf("empty account")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	res, err := s.db.Exec(`DELETE FROM accounts WHERE lower(name) = lower(?)`, account)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("account not found")
	}
	return nil
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
		sealed, err := sealToken(tok)
		if err != nil {
			return err
		}
		if _, err := stmt.Exec(name, sealed, now); err != nil {
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
