package terminal

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
)

type PTYWrapper struct {
	cmd      *exec.Cmd
	pty      *os.File
	process  *os.Process
	reader   *bufio.Reader
	writer   *bufio.Writer
	size     *pty.Winsize
	mu       sync.Mutex
	stopChan chan struct{}
}

func NewPTYWrapper(command string, args []string, env map[string]string) (*PTYWrapper, error) {
	// Create command
	cmd := exec.Command(command, args...)
	
	// Set environment variables
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Default terminal size
	size := &pty.Winsize{
		Rows: 24,
		Cols: 80,
	}

	return &PTYWrapper{
		cmd:      cmd,
		size:     size,
		stopChan: make(chan struct{}),
	}, nil
}

func (p *PTYWrapper) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Start command with PTY
	ptmx, err := pty.StartWithSize(p.cmd, p.size)
	if err != nil {
		return fmt.Errorf("failed to start PTY: %w", err)
	}

	p.pty = ptmx
	p.process = p.cmd.Process
	p.reader = bufio.NewReader(ptmx)
	p.writer = bufio.NewWriter(ptmx)

	return nil
}

func (p *PTYWrapper) Read() ([]byte, error) {
	if p.reader == nil {
		return nil, fmt.Errorf("PTY not started")
	}

	// Read up to 4KB at a time
	buf := make([]byte, 4096)
	n, err := p.reader.Read(buf)
	if err != nil {
		if err == io.EOF {
			// Process has exited
			return nil, err
		}
		return nil, fmt.Errorf("failed to read from PTY: %w", err)
	}

	return buf[:n], nil
}

func (p *PTYWrapper) Write(data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.writer == nil {
		return fmt.Errorf("PTY not started")
	}

	_, err := p.writer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to PTY: %w", err)
	}

	return p.writer.Flush()
}

func (p *PTYWrapper) Resize(rows, cols uint16) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pty == nil {
		return fmt.Errorf("PTY not started")
	}

	p.size.Rows = rows
	p.size.Cols = cols

	return pty.Setsize(p.pty, p.size)
}

func (p *PTYWrapper) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Signal stop
	close(p.stopChan)

	// Kill the process if it's still running
	if p.process != nil {
		if err := p.process.Kill(); err != nil {
			// Process might already be dead
			if !os.IsPermission(err) {
				return fmt.Errorf("failed to kill process: %w", err)
			}
		}
		
		// Wait for process to exit
		_, _ = p.process.Wait()
	}

	// Close PTY
	if p.pty != nil {
		if err := p.pty.Close(); err != nil {
			return fmt.Errorf("failed to close PTY: %w", err)
		}
	}

	return nil
}

func (p *PTYWrapper) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.process == nil {
		return false
	}

	// Check if process is still running
	// This is a non-blocking check
	return p.process.Signal(os.Signal(nil)) == nil
}