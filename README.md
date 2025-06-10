# MCP Terminal Tester

An MCP (Model Context Protocol) server that enables AI assistants to test and interact with terminal/TUI applications.

## Overview

This server acts as a bridge between AI assistants (like Claude) and terminal applications, providing:
- Terminal application launching with PTY emulation
- Visual output capture with ANSI support
- Keyboard input simulation
- Terminal state information

## Quick Start

```bash
# Install dependencies
go mod download

# Build the server
make build

# Run the server
./bin/mcp-terminal-server

# Or run directly
go run cmd/server/main.go
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
  "format": "plain"  // or "raw", "ansi"
}
```

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
- `restart_app`: Restart a session
- `stop_app`: Terminate a session
- `list_sessions`: List all active sessions

## Configuration

Environment variables:
- `MCP_PORT`: Server port (default: 8080)
- `MAX_SESSIONS`: Maximum concurrent sessions (default: 100)
- `SESSION_TIMEOUT`: Idle timeout in minutes (default: 30)
- `LOG_LEVEL`: Logging level (default: info)

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
```

## License

MIT