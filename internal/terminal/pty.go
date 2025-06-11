package terminal

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"github.com/creack/pty"
	"github.com/bioharz/mcp-terminal-tester/internal/utils"
)

// Buffer pool for PTY reads to reduce GC pressure
var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 4096)
	},
}

type PTYWrapper struct {
	cmd         *exec.Cmd
	pty         *os.File
	process     *os.Process
	reader      *bufio.Reader
	writer      *bufio.Writer
	size        *pty.Winsize
	mu          sync.Mutex
	stopChan    chan struct{}
	resizeChan  chan *pty.Winsize
	sessionID   string // For logging
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
		cmd:        cmd,
		size:       size,
		stopChan:   make(chan struct{}),
		resizeChan: make(chan *pty.Winsize, 1),
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

	// Start resize handler
	go p.handleResize()

	slog.Debug("PTY started",
		slog.String("session_id", p.sessionID),
		slog.Int("rows", int(p.size.Rows)),
		slog.Int("cols", int(p.size.Cols)),
	)

	return nil
}

func (p *PTYWrapper) Read() ([]byte, error) {
	if p.reader == nil {
		return nil, fmt.Errorf("PTY not started")
	}

	// Get buffer from pool to reduce allocations
	buf := bufferPool.Get().([]byte)
	n, err := p.reader.Read(buf)
	if err != nil {
		if err == io.EOF {
			// Process has exited
			bufferPool.Put(buf) // Return buffer to pool
			return nil, err
		}
		bufferPool.Put(buf) // Return buffer to pool
		return nil, fmt.Errorf("failed to read from PTY: %w", err)
	}

	// Create a copy of the data since we need to return the buffer to pool
	result := make([]byte, n)
	copy(result, buf[:n])
	bufferPool.Put(buf) // Return buffer to pool

	return result, nil
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
	newSize := &pty.Winsize{
		Rows: rows,
		Cols: cols,
	}

	// Send resize request to handler goroutine
	select {
	case p.resizeChan <- newSize:
		slog.Debug("Resize requested",
			slog.String("session_id", p.sessionID),
			slog.Int("rows", int(rows)),
			slog.Int("cols", int(cols)),
		)
		return nil
	case <-p.stopChan:
		return fmt.Errorf("PTY is stopped")
	default:
		// Resize channel is full, skip this resize
		slog.Debug("Resize skipped (channel full)",
			slog.String("session_id", p.sessionID),
		)
		return nil
	}
}

func (p *PTYWrapper) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Signal stop only once
	select {
	case <-p.stopChan:
		// Already stopped
		return nil
	default:
		close(p.stopChan)
	}

	// Kill the process if it's still running
	if p.process != nil {
		if err := p.process.Kill(); err != nil {
			// Process might already be dead
			if !os.IsPermission(err) {
				utils.LogError(err, "Failed to kill process",
					slog.String("session_id", p.sessionID),
				)
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

// SetSessionID sets the session ID for logging
func (p *PTYWrapper) SetSessionID(id string) {
	p.sessionID = id
}

// handleResize handles resize requests in a separate goroutine
func (p *PTYWrapper) handleResize() {
	for {
		select {
		case newSize := <-p.resizeChan:
			p.mu.Lock()
			if p.pty != nil {
				oldRows, oldCols := p.size.Rows, p.size.Cols
				p.size = newSize
				
				err := pty.Setsize(p.pty, p.size)
				if err != nil {
					utils.LogError(err, "Failed to resize PTY",
						slog.String("session_id", p.sessionID),
						slog.Int("rows", int(newSize.Rows)),
						slog.Int("cols", int(newSize.Cols)),
					)
				} else {
					slog.Info("PTY resized",
						slog.String("session_id", p.sessionID),
						slog.Int("old_rows", int(oldRows)),
						slog.Int("old_cols", int(oldCols)),
						slog.Int("new_rows", int(newSize.Rows)),
						slog.Int("new_cols", int(newSize.Cols)),
					)
				}
			}
			p.mu.Unlock()
		case <-p.stopChan:
			return
		}
	}
}

// StartSIGWINCHHandler starts monitoring for terminal size changes
// This is mainly for when the MCP server itself is running in a terminal
func StartSIGWINCHHandler() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	
	go func() {
		for range ch {
			// In a real implementation, you would get the new terminal size
			// and propagate it to active sessions
			slog.Debug("SIGWINCH received")
		}
	}()
}