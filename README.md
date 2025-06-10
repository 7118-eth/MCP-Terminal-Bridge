# MCP Terminal Tester

An MCP (Model Context Protocol) server that enables AI assistants to test and interact with terminal/TUI applications.

## Overview

This server acts as a bridge between AI assistants (like Claude) and terminal applications, providing:
- Terminal application launching with PTY emulation
- Visual output capture with ANSI support
- Keyboard input simulation
- Terminal state information

## Current Implementation Status

### Phase 1: Foundation (COMPLETE ✅)
- ✅ All 9 MCP tools implemented and working
- ✅ Session management with automatic cleanup
- ✅ PTY wrapper for terminal control
- ✅ Screen buffer with basic ANSI support
- ✅ Special key mapping (arrows, function keys, Ctrl sequences)
- ✅ Multiple output formats (plain, raw, ansi)
- ✅ Concurrent session support
- ✅ Build system with Makefile

### Phase 2: Core Features (COMPLETE ✅)
- ✅ Enhanced ANSI parser (supports CSI, SGR, OSC, DCS, and more)
- ✅ Terminal resize support with SIGWINCH handling  
- ✅ Structured logging throughout
- ✅ Scrollback buffer support (1000 lines)
- ✅ All unit tests passing
- ✅ Proper renderRaw() with ANSI sequences
- ✅ Multiple output formats including scrollback

### Phase 3: Testing & Robustness (IN PROGRESS 🚧)
- ✅ Comprehensive integration test framework
- ✅ Tests for all 9 MCP tools
- ✅ Test applications (echo, menu, progress)
- ✅ 13 out of 18 integration tests passing
- 🚧 Error recovery and session management
- 🚧 Performance optimizations

## Quick Start

```bash
# Install dependencies
go mod download

# Build the server
make build

# Build test applications
make test-apps

# Run the server (stdio mode)
./bin/mcp-terminal-server

# Or run directly
go run cmd/server/main.go

# Run tests
make test

# Test with example app (in another terminal)
cd test/apps
./echo
```

## MCP Tools

### launch_app
Launch a new terminal application.
```json
{
  "command": "vim",
  "args": ["test.txt"],
  "env": {"TERM": "xterm-256color"}
}
```

### view_screen
Get the current terminal content.
```json
{
  "session_id": "session-123",
  "format": "plain"  // or "raw", "ansi", "scrollback"
}
```

Output formats:
- `plain`: Text only, no ANSI escape sequences
- `raw`: Full terminal output with ANSI escape sequences
- `ansi`: Debug format showing cursor position with ▮
- `scrollback`: Includes scrollback buffer history

### send_keys
Send keyboard input to the terminal.
```json
{
  "session_id": "session-123",
  "keys": "Hello World"  // or "Enter", "Ctrl+C", etc.
}
```

### Other Tools
- `get_cursor_position`: Get current cursor position
- `get_screen_size`: Get terminal dimensions
- `resize_terminal`: Resize the terminal window
- `restart_app`: Restart a session
- `stop_app`: Terminate a session
- `list_sessions`: List all active sessions

## Configuration

Environment variables:
- `MCP_PORT`: Not used in current stdio implementation
- `MAX_SESSIONS`: Maximum concurrent sessions (default: 100)
- `SESSION_TIMEOUT`: Idle timeout in minutes (default: 30)
- `LOG_LEVEL`: Logging level (default: info)

## Implementation Notes

- Uses `mark3labs/mcp-go` v0.31.0 for MCP protocol
- Uses `creack/pty` v1.1.24 for terminal emulation
- Runs in stdio mode (standard input/output)
- Session cleanup runs every 5 minutes
- Default terminal size: 80x24 (resizable via `resize_terminal` tool)
- Structured JSON logging to stderr (configurable via LOG_LEVEL)
- Enhanced ANSI parser supports most common escape sequences

## Development

See `project.md` for complete technical design and `progress.md` for current development status.

## Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run integration tests
make test-integration

# Run specific test suites
make test-terminal    # Terminal package tests
make test-session     # Session manager tests
```

## License

MIT