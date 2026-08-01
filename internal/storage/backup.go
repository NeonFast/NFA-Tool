package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Store persists account -> refresh token map next to the executable.
type Store struct {
	path string
	mu   sync.Mutex
}

func New(baseDir string) *Store {
	if baseDir == "" {
		baseDir, _ = os.Getwd()
	}
	return &Store{path: filepath.Join(baseDir, "user_backup.json")}
}

func (s *Store) Path() string { return s.path }

func (s *Store) Load() (map[string]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadUnlocked()
}

func (s *Store) loadUnlocked() (map[string]string, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	out := map[string]string{}
	if len(data) == 0 {
		return out, nil
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Store) Save(m map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveUnlocked(m)
}

func (s *Store) saveUnlocked(m map[string]string) error {
	if m == nil {
		m = map[string]string{}
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}

func (s *Store) Put(account, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, err := s.loadUnlocked()
	if err != nil {
		return err
	}
	m[account] = token
	return s.saveUnlocked(m)
}

func (s *Store) Delete(account string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, err := s.loadUnlocked()
	if err != nil {
		return err
	}
	delete(m, account)
	return s.saveUnlocked(m)
}

func (s *Store) Merge(extra map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, err := s.loadUnlocked()
	if err != nil {
		return err
	}
	for k, v := range extra {
		if k != "" && v != "" {
			m[k] = v
		}
	}
	return s.saveUnlocked(m)
}
