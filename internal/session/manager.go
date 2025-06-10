package session

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/bioharz/mcp-terminal-tester/internal/utils"
)

type Manager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	maxSessions int
	sessionTimeout time.Duration
}

func NewManager() *Manager {
	m := &Manager{
		sessions: make(map[string]*Session),
		maxSessions: 100,
		sessionTimeout: 30 * time.Minute,
	}
	slog.Info("Session manager created",
		slog.Int("max_sessions", m.maxSessions),
		slog.Duration("session_timeout", m.sessionTimeout),
	)
	return m
}

func (m *Manager) CreateSession(command string, args []string, env map[string]string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.sessions) >= m.maxSessions {
		err := fmt.Errorf("maximum number of sessions (%d) reached", m.maxSessions)
		slog.Error("Failed to create session", 
			slog.String("error", err.Error()),
			slog.Int("current_sessions", len(m.sessions)),
		)
		return nil, err
	}

	session, err := NewSession(command, args, env)
	if err != nil {
		utils.LogError(err, "Failed to create session",
			slog.String("command", command),
			slog.Any("args", args),
		)
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	m.sessions[session.ID] = session
	utils.LogSessionEvent(session.ID, "created",
		slog.String("command", command),
		slog.Any("args", args),
		slog.Int("total_sessions", len(m.sessions)),
	)
	return session, nil
}

func (m *Manager) GetSession(id string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[id]
	if !exists {
		err := fmt.Errorf("session not found: %s", id)
		slog.Debug("Session lookup failed",
			slog.String("session_id", id),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	// Update last active time
	session.UpdateLastActive()
	slog.Debug("Session accessed", slog.String("session_id", id))

	return session, nil
}

func (m *Manager) RemoveSession(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[id]
	if !exists {
		err := fmt.Errorf("session not found: %s", id)
		slog.Debug("Cannot remove non-existent session",
			slog.String("session_id", id),
			slog.String("error", err.Error()),
		)
		return err
	}

	// Clean up the session
	if err := session.Close(); err != nil {
		utils.LogError(err, "Failed to close session", slog.String("session_id", id))
		return fmt.Errorf("failed to close session: %w", err)
	}

	delete(m.sessions, id)
	utils.LogSessionEvent(id, "removed",
		slog.Int("remaining_sessions", len(m.sessions)),
	)
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
	cleaned := 0
	for id, session := range m.sessions {
		idleTime := now.Sub(session.LastActive)
		if idleTime > m.sessionTimeout {
			if err := session.Close(); err != nil {
				utils.LogError(err, "Error closing idle session",
					slog.String("session_id", id),
					slog.Duration("idle_time", idleTime),
				)
			}
			delete(m.sessions, id)
			utils.LogSessionEvent(id, "cleaned_idle",
				slog.Duration("idle_time", idleTime),
			)
			cleaned++
		}
	}
	if cleaned > 0 {
		slog.Info("Idle session cleanup completed",
			slog.Int("cleaned", cleaned),
			slog.Int("remaining", len(m.sessions)),
		)
	}
}

func (m *Manager) StartCleanupRoutine() {
	interval := 5 * time.Minute
	slog.Info("Starting session cleanup routine", slog.Duration("interval", interval))
	
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			slog.Debug("Running idle session cleanup")
			m.CleanupIdleSessions()
		}
	}()
}