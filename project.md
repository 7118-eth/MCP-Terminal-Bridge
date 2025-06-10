# MCP Terminal Tester - Technical Design Document

## Project Overview

### Mission
Build an MCP (Model Context Protocol) server in Go that enables AI assistants to test and interact with terminal/TUI applications through a structured API.

### Problem Statement
AI assistants cannot directly see or interact with terminal applications. This server bridges that gap by:
1. Launching terminal applications as subprocesses with proper PTY emulation
2. Capturing visual output with full ANSI support
3. Sending keyboard input including special keys
4. Providing terminal state information

## Architecture

### High-Level Architecture
```
┌─────────────────┐
│   MCP Client    │
│    (Claude)     │
└────────┬────────┘
         │ JSON-RPC
┌────────┴────────┐
│   MCP Server    │
│  (This Project) │
├─────────────────┤
│   Tool Layer    │
├─────────────────┤
│ Session Manager │
├─────────────────┤
│ Terminal Layer  │
├─────────────────┤
│   PTY Layer     │
└────────┬────────┘
         │
┌────────┴────────┐
│ Terminal Apps   │
└─────────────────┘
```

### Component Architecture

#### 1. MCP Server Layer
- Handles MCP protocol communication
- Tool registration and dispatch
- Request/response serialization
- Error handling and protocol compliance

#### 2. Tool Layer
Implements MCP tools as defined handlers:

**Core Tools:**
- `launch_app`: Start new terminal application
- `view_screen`: Get current terminal content (supports plain/raw/ansi/scrollback formats)
- `send_keys`: Send keyboard input
- `get_cursor_position`: Get cursor location
- `get_screen_size`: Get terminal dimensions
- `resize_terminal`: Dynamically resize terminal window
- `restart_app`: Restart existing session
- `stop_app`: Terminate session
- `list_sessions`: List active sessions

#### 3. Session Manager
```go
type SessionManager struct {
    sessions map[string]*Session
    mu       sync.RWMutex
}

type Session struct {
    ID          string
    Command     string
    Args        []string
    Env         map[string]string
    PTY         *PTYWrapper
    Buffer      *ScreenBuffer
    Created     time.Time
    LastActive  time.Time
    State       SessionState
}
```

#### 4. Terminal Layer
Handles terminal emulation and screen buffering:

```go
type ScreenBuffer struct {
    cells    [][]Cell
    width    int
    height   int
    cursorX  int
    cursorY  int
    dirty    bool
    parser   *ANSIParser
}

type Cell struct {
    Rune       rune
    Foreground Color
    Background Color
    Attributes Attributes
}
```

#### 5. PTY Layer
Manages pseudo-terminal creation and I/O:

```go
type PTYWrapper struct {
    pty      *os.File
    process  *os.Process
    reader   *bufio.Reader
    writer   *bufio.Writer
    size     *pty.Winsize
}
```

### Data Flow

#### Launch Application Flow
```
1. Client → launch_app(command, args, env)
2. SessionManager.CreateSession()
3. PTYWrapper.Start()
4. ScreenBuffer.Initialize()
5. Return session_id
```

#### View Screen Flow
```
1. Client → view_screen(session_id, format)
2. SessionManager.GetSession()
3. ScreenBuffer.Render(format)
4. Return formatted content
```

#### Send Keys Flow
```
1. Client → send_keys(session_id, keys)
2. SessionManager.GetSession()
3. KeyMapper.MapKeys()
4. PTYWrapper.Write()
5. Return success
```

## Detailed Design

### Session Management

**Lifecycle:**
1. Creation: Allocate resources, start PTY
2. Active: Handle I/O, maintain buffer
3. Cleanup: Kill process, close PTY, free memory

**Concurrency Model:**
- One goroutine per session for PTY reading
- Mutex protection for session map
- Channel-based communication for updates

**Resource Management:**
- Automatic cleanup on process exit
- Timeout-based cleanup for idle sessions
- Maximum session limit to prevent resource exhaustion

### PTY Handling

**Key Features:**
- Cross-platform support (Linux, macOS, Windows via ConPTY)
- Signal forwarding (SIGINT, SIGTERM, SIGWINCH)
- Non-blocking I/O with buffering
- Proper terminal mode settings

**Implementation Details:**
```go
// PTY creation
func CreatePTY(cmd string, args []string, size *pty.Winsize) (*PTYWrapper, error) {
    // Set up command
    // Create PTY
    // Configure terminal modes
    // Start process
    // Begin I/O goroutines
}
```

### Screen Buffer Design

**Buffer Structure:**
- 2D array of cells (rune + attributes)
- Circular buffer for scrollback
- Dirty region tracking
- Unicode support with wide character handling

**ANSI Processing:**
- State machine for escape sequences
- Support for:
  - Cursor movement (CSI sequences)
  - Colors (SGR sequences)
  - Screen clearing (ED, EL)
  - Scrolling regions
  - Alternative screen buffer

**Output Formats:**
1. **Raw**: Complete with ANSI sequences
2. **Plain**: Stripped of all formatting
3. **ANSI**: With visible markers for debugging

### Input Processing

**Key Mapping:**
```go
var specialKeys = map[string]string{
    "Enter":     "\r",
    "Tab":       "\t",
    "Backspace": "\x7f",
    "Escape":    "\x1b",
    "Up":        "\x1b[A",
    "Down":      "\x1b[B",
    "Right":     "\x1b[C",
    "Left":      "\x1b[D",
    "Ctrl+C":    "\x03",
    "Ctrl+D":    "\x04",
    // ... more mappings
}
```

**Input Queue:**
- Buffered channel for queueing
- Rate limiting to prevent flooding
- Support for paste mode detection

### Error Handling

**Strategy:**
- Graceful degradation
- Detailed error messages
- Recovery mechanisms
- Structured logging

**Error Types:**
1. Session errors (not found, already exists)
2. PTY errors (creation failed, process died)
3. Input errors (invalid keys, encoding issues)
4. Resource errors (out of memory, too many sessions)

## Implementation Details

### Project Structure
```
mcp-terminal-tester/
├── cmd/
│   └── server/
│       └── main.go          # Entry point
├── internal/
│   ├── mcp/
│   │   ├── server.go        # MCP server setup
│   │   └── tools.go         # Tool definitions
│   ├── session/
│   │   ├── manager.go       # Session management
│   │   ├── session.go       # Session struct
│   │   └── store.go         # Session storage
│   ├── terminal/
│   │   ├── pty.go          # PTY wrapper
│   │   ├── buffer.go       # Screen buffer
│   │   ├── ansi.go         # ANSI parser
│   │   └── cell.go         # Cell definition
│   ├── tools/
│   │   ├── launch.go       # launch_app tool
│   │   ├── view.go         # view_screen tool
│   │   ├── input.go        # send_keys tool
│   │   ├── info.go         # cursor/size tools
│   │   └── control.go      # stop/restart tools
│   └── utils/
│       ├── keys.go         # Key mapping
│       └── logger.go       # Logging setup
├── test/
│   ├── apps/               # Test applications
│   └── integration/        # Integration tests
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Configuration

**Server Configuration:**
```go
type Config struct {
    Port           int
    MaxSessions    int
    SessionTimeout time.Duration
    BufferSize     int
    LogLevel       string
}
```

**Default Values:**
- Port: 8080 (or from MCP_PORT env)
- MaxSessions: 100
- SessionTimeout: 30 minutes
- BufferSize: 4096
- LogLevel: "info"

### Security Considerations

1. **Process Isolation**: Each session runs with restricted permissions
2. **Input Validation**: Sanitize all inputs, prevent injection
3. **Resource Limits**: CPU, memory, and file descriptor limits
4. **No Network Access**: Terminal apps run in isolated environment
5. **Audit Logging**: Track all operations with session context

### Performance Optimization

1. **Buffer Management**:
   - Pool allocators for cells
   - Dirty region tracking
   - Incremental updates

2. **I/O Optimization**:
   - Buffered reads/writes
   - Batch processing
   - Non-blocking operations

3. **Concurrency**:
   - Lock-free data structures where possible
   - Read-write locks for session map
   - Worker pools for heavy operations

### Testing Strategy

**Unit Tests:**
- ANSI parser edge cases
- Key mapping validation
- Buffer operations
- Session lifecycle

**Integration Tests:**
- Full tool workflows
- Concurrent session handling
- Error recovery
- Resource cleanup

**Test Applications:**
1. Echo server (simple input/output)
2. Menu system (navigation testing)
3. Progress bar (ANSI animation)
4. Vim-like editor (complex interactions)

### Monitoring and Observability

**Metrics:**
- Active sessions count
- Request latency by tool
- Error rates
- Resource usage

**Logging:**
- Structured JSON logs
- Request/response tracing
- Session event timeline
- Error stack traces

### Future Enhancements

1. **Session Recording**: Record and replay terminal sessions
2. **Multi-window Support**: Handle tmux/screen multiplexing
3. **File Transfer**: Upload/download files to/from sessions
4. **WebSocket Bridge**: Real-time terminal streaming
5. **Session Persistence**: Save/restore sessions across restarts

## API Reference

### Tool Specifications

#### launch_app
```json
{
  "name": "launch_app",
  "params": {
    "command": "string",
    "args": ["string"],
    "env": {"key": "value"}
  },
  "returns": {
    "session_id": "string",
    "success": "boolean"
  }
}
```

#### view_screen
```json
{
  "name": "view_screen",
  "params": {
    "session_id": "string",
    "format": "raw|plain|ansi"
  },
  "returns": {
    "content": "string",
    "cursor": {"row": 0, "col": 0}
  }
}
```

#### send_keys
```json
{
  "name": "send_keys",
  "params": {
    "session_id": "string",
    "keys": "string"
  },
  "returns": {
    "success": "boolean"
  }
}
```

### Error Codes

- `SESSION_NOT_FOUND`: Invalid session ID
- `SESSION_DEAD`: Process has terminated
- `INVALID_INPUT`: Malformed input data
- `RESOURCE_LIMIT`: Too many sessions
- `INTERNAL_ERROR`: Unexpected server error

## Development Workflow

1. **Setup**: Clone repo, install Go 1.21+
2. **Dependencies**: `go mod download`
3. **Build**: `make build`
4. **Test**: `make test`
5. **Run**: `./bin/mcp-terminal-server`
6. **Debug**: Set LOG_LEVEL=debug

## Success Criteria

1. ✓ Launch and control terminal applications
2. ✓ Accurately capture terminal output
3. ✓ Handle all common keyboard inputs
4. ✓ Support TUI frameworks (Bubble Tea, Termui)
5. ✓ Clean session lifecycle management
6. ✓ Concurrent session support
7. ✓ <10ms response times
8. ✓ No resource leaks
9. ✓ Cross-platform compatibility
10. ✓ Comprehensive test coverage