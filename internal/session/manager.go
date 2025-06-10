package session

import (
	"fmt"
	"sync"
	"time"
)

type Manager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	maxSessions int
	sessionTimeout time.Duration
}

func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
		maxSessions: 100,
		sessionTimeout: 30 * time.Minute,
	}
}

func (m *Manager) CreateSession(command string, args []string, env map[string]string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.sessions) >= m.maxSessions {
		return nil, fmt.Errorf("maximum number of sessions (%d) reached", m.maxSessions)
	}

	session, err := NewSession(command, args, env)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	m.sessions[session.ID] = session
	return session, nil
}

func (m *Manager) GetSession(id string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", id)
	}

	// Update last active time
	session.UpdateLastActive()

	return session, nil
}

func (m *Manager) RemoveSession(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[id]
	if !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	// Clean up the session
	if err := session.Close(); err != nil {
		return fmt.Errorf("failed to close session: %w", err)
	}

	delete(m.sessions, id)
	return nil
}

func (m *Manager) ListSessions() []*SessionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sessions []*SessionInfo
	for _, session := range m.sessions {
		sessions = append(sessions, session.GetInfo())
	}

	return sessions
}

func (m *Manager) CleanupIdleSessions() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for id, session := range m.sessions {
		if now.Sub(session.LastActive) > m.sessionTimeout {
			if err := session.Close(); err != nil {
				// Log error but continue cleanup
				fmt.Printf("Error closing idle session %s: %v\n", id, err)
			}
			delete(m.sessions, id)
		}
	}
}

func (m *Manager) StartCleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			m.CleanupIdleSessions()
		}
	}()
}