package session

import (
	"fmt"
	"sync"
	"time"

	"github.com/bioharz/mcp-terminal-tester/internal/terminal"
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

	// Create PTY wrapper
	pty, err := terminal.NewPTYWrapper(command, args, env)
	if err != nil {
		return nil, err
	}

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
		return nil, err
	}

	return session, nil
}

func (s *Session) start() error {
	// Start the PTY process
	if err := s.PTY.Start(); err != nil {
		return err
	}

	// Start goroutine to read from PTY and update buffer
	go s.readLoop()

	return nil
}

func (s *Session) readLoop() {
	for {
		data, err := s.PTY.Read()
		if err != nil {
			s.mu.Lock()
			s.State = StateError
			s.mu.Unlock()
			return
		}

		// Update the screen buffer with new data
		s.Buffer.Write(data)
	}
}

func (s *Session) SendKeys(keys string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.State != StateActive {
		return fmt.Errorf("session is not active")
	}

	return s.PTY.Write([]byte(keys))
}

func (s *Session) GetScreen(format string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.State != StateActive {
		return "", fmt.Errorf("session is not active")
	}

	return s.Buffer.Render(format)
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

	// Stop current process
	if err := s.PTY.Stop(); err != nil {
		return err
	}

	// Clear buffer
	s.Buffer.Clear()

	// Create new PTY
	pty, err := terminal.NewPTYWrapper(s.Command, s.Args, s.Env)
	if err != nil {
		return err
	}

	s.PTY = pty
	s.State = StateActive
	s.LastActive = time.Now()

	// Start again
	return s.start()
}

func (s *Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.State = StateStopped
	return s.PTY.Stop()
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