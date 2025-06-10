package session

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/bioharz/mcp-terminal-tester/internal/terminal"
	"github.com/bioharz/mcp-terminal-tester/internal/utils"
	"github.com/google/uuid"
)

type SessionState int

const (
	StateActive SessionState = iota
	StateStopped
	StateError
)

type Session struct {
	ID         string
	Command    string
	Args       []string
	Env        map[string]string
	PTY        *terminal.PTYWrapper
	Buffer     *terminal.ScreenBuffer
	Created    time.Time
	LastActive time.Time
	State      SessionState
	mu         sync.RWMutex
}

type SessionInfo struct {
	ID         string            `json:"id"`
	Command    string            `json:"command"`
	Args       []string          `json:"args"`
	Created    time.Time         `json:"created"`
	LastActive time.Time         `json:"last_active"`
	State      string            `json:"state"`
}

func NewSession(command string, args []string, env map[string]string) (*Session, error) {
	// Generate unique session ID
	id := uuid.New().String()

	slog.Debug("Creating new session",
		slog.String("session_id", id),
		slog.String("command", command),
		slog.Any("args", args),
	)

	// Create PTY wrapper
	pty, err := terminal.NewPTYWrapper(command, args, env)
	if err != nil {
		utils.LogError(err, "Failed to create PTY", slog.String("session_id", id))
		return nil, err
	}
	
	// Set session ID for logging
	pty.SetSessionID(id)

	// Create screen buffer
	buffer := terminal.NewScreenBuffer(80, 24)

	session := &Session{
		ID:         id,
		Command:    command,
		Args:       args,
		Env:        env,
		PTY:        pty,
		Buffer:     buffer,
		Created:    time.Now(),
		LastActive: time.Now(),
		State:      StateActive,
	}

	// Start PTY and connect it to the buffer
	if err := session.start(); err != nil {
		utils.LogError(err, "Failed to start session", slog.String("session_id", id))
		return nil, err
	}

	slog.Info("Session created successfully",
		slog.String("session_id", id),
		slog.String("command", command),
	)

	return session, nil
}

func (s *Session) start() error {
	// Start the PTY process
	if err := s.PTY.Start(); err != nil {
		return err
	}

	slog.Debug("PTY started", slog.String("session_id", s.ID))

	// Start goroutine to read from PTY and update buffer
	go s.readLoop()

	return nil
}

func (s *Session) readLoop() {
	slog.Debug("Starting read loop", slog.String("session_id", s.ID))
	
	for {
		data, err := s.PTY.Read()
		if err != nil {
			s.mu.Lock()
			s.State = StateError
			s.mu.Unlock()
			
			if err.Error() != "EOF" {
				utils.LogError(err, "Read loop error", slog.String("session_id", s.ID))
			} else {
				slog.Debug("Read loop ended (EOF)", slog.String("session_id", s.ID))
			}
			return
		}

		// Update the screen buffer with new data
		s.Buffer.Write(data)
		slog.Debug("Buffer updated",
			slog.String("session_id", s.ID),
			slog.Int("bytes", len(data)),
		)
	}
}

func (s *Session) SendKeys(keys string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.State != StateActive {
		err := fmt.Errorf("session is not active")
		slog.Debug("Cannot send keys to inactive session",
			slog.String("session_id", s.ID),
			slog.String("state", s.getStateString()),
		)
		return err
	}

	err := s.PTY.Write([]byte(keys))
	if err != nil {
		utils.LogError(err, "Failed to send keys",
			slog.String("session_id", s.ID),
			slog.Int("key_length", len(keys)),
		)
	} else {
		slog.Debug("Keys sent",
			slog.String("session_id", s.ID),
			slog.Int("key_length", len(keys)),
		)
	}
	return err
}

func (s *Session) GetScreen(format string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.State != StateActive {
		err := fmt.Errorf("session is not active")
		slog.Debug("Cannot get screen from inactive session",
			slog.String("session_id", s.ID),
			slog.String("state", s.getStateString()),
		)
		return "", err
	}

	content, err := s.Buffer.Render(format)
	if err != nil {
		utils.LogError(err, "Failed to render screen",
			slog.String("session_id", s.ID),
			slog.String("format", format),
		)
	} else {
		slog.Debug("Screen rendered",
			slog.String("session_id", s.ID),
			slog.String("format", format),
			slog.Int("content_length", len(content)),
		)
	}
	return content, err
}

func (s *Session) GetCursorPosition() (int, int) {
	return s.Buffer.GetCursorPosition()
}

func (s *Session) GetScreenSize() (int, int) {
	return s.Buffer.GetSize()
}

func (s *Session) Restart() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	slog.Info("Restarting session", slog.String("session_id", s.ID))

	// Stop current process
	if err := s.PTY.Stop(); err != nil {
		utils.LogError(err, "Failed to stop PTY during restart", slog.String("session_id", s.ID))
		return err
	}

	// Clear buffer
	s.Buffer.Clear()

	// Create new PTY
	pty, err := terminal.NewPTYWrapper(s.Command, s.Args, s.Env)
	if err != nil {
		utils.LogError(err, "Failed to create new PTY during restart", slog.String("session_id", s.ID))
		return err
	}
	
	// Set session ID for logging
	pty.SetSessionID(s.ID)

	s.PTY = pty
	s.State = StateActive
	s.LastActive = time.Now()

	// Start again
	err = s.start()
	if err != nil {
		utils.LogError(err, "Failed to start session after restart", slog.String("session_id", s.ID))
		s.State = StateError
	} else {
		// Give the process a moment to start before the readLoop begins reading
		time.Sleep(50 * time.Millisecond)
		slog.Info("Session restarted successfully", slog.String("session_id", s.ID))
	}
	return err
}

func (s *Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	slog.Debug("Closing session", slog.String("session_id", s.ID))

	s.State = StateStopped
	err := s.PTY.Stop()
	if err != nil {
		utils.LogError(err, "Failed to stop PTY during close", slog.String("session_id", s.ID))
	} else {
		slog.Info("Session closed", slog.String("session_id", s.ID))
	}
	return err
}

func (s *Session) UpdateLastActive() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastActive = time.Now()
}

func (s *Session) GetInfo() *SessionInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state := "active"
	switch s.State {
	case StateStopped:
		state = "stopped"
	case StateError:
		state = "error"
	}

	return &SessionInfo{
		ID:         s.ID,
		Command:    s.Command,
		Args:       s.Args,
		Created:    s.Created,
		LastActive: s.LastActive,
		State:      state,
	}
}

func (s *Session) getStateString() string {
	switch s.State {
	case StateActive:
		return "active"
	case StateStopped:
		return "stopped"
	case StateError:
		return "error"
	default:
		return "unknown"
	}
}

// Resize resizes the terminal
func (s *Session) Resize(width, height int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.State != StateActive {
		err := fmt.Errorf("session is not active")
		slog.Debug("Cannot resize inactive session",
			slog.String("session_id", s.ID),
			slog.String("state", s.getStateString()),
		)
		return err
	}

	// Resize the PTY
	err := s.PTY.Resize(uint16(height), uint16(width))
	if err != nil {
		utils.LogError(err, "Failed to resize PTY",
			slog.String("session_id", s.ID),
			slog.Int("width", width),
			slog.Int("height", height),
		)
		return err
	}

	// Resize the buffer
	s.Buffer.Resize(width, height)

	slog.Info("Session resized",
		slog.String("session_id", s.ID),
		slog.Int("width", width),
		slog.Int("height", height),
	)

	return nil
}