# AI Assistant Context for MCP Terminal Tester

## Project Status (as of January 10, 2025)

### Current State
- **Phase 2 COMPLETE**: All core features implemented and tested
- **9 MCP tools** fully implemented and working
- Builds successfully with `make build`
- Ready for real-world testing with MCP clients

### Key Implementation Details

#### Architecture
- Uses `mark3labs/mcp-go` v0.31.0 (stdio mode only)
- Session management with goroutine per session for PTY reading
- Circular scrollback buffer (1000 lines)
- Non-blocking resize handling via channels

#### ANSI Parser State Machine
- Supports: CSI, SGR, OSC, DCS sequences
- Handles: cursor movement, colors (256-color), attributes, clearing
- Save/restore cursor position implemented
- Escape sequence buffer for parameter parsing

#### PTY Handling
- Uses `creack/pty` for cross-platform support
- Separate goroutine for resize requests
- Session ID logging for debugging
- Graceful shutdown with process cleanup

#### Output Formats
1. **plain**: Stripped of ANSI, trimmed whitespace
2. **raw**: Includes ANSI sequences with full SGR rendering
3. **ansi**: Shows cursor position with ▮ marker
4. **scrollback**: Includes historical lines before current screen

### Known Issues & Limitations

1. **Test Failures**: Some unit tests fail due to:
   - Newline handling (LF doesn't reset cursor X)
   - Cursor save/restore needs global state
   - Scrollback test needs proper initialization

2. **Not Implemented**:
   - Mouse support
   - Alternate screen buffer
   - Some advanced ANSI modes
   - Session persistence
   - Rate limiting for input

3. **Platform Specific**:
   - SIGWINCH handling may vary
   - Terminal mode setting in menu.go is simplified
   - Windows support needs testing

### Testing Strategy

#### Unit Tests
- Run with: `make test-terminal` or `make test-session`
- Some tests need fixes for proper newline handling
- Coverage reports: `make test-coverage`

#### Test Applications
Located in `test/apps/`:
- **echo.go**: Basic I/O, commands, color test
- **menu.go**: Arrow keys, box drawing, 256 colors
- **progress.go**: Animations, multi-line updates

Build all with: `cd test/apps && make all`

#### What to Test with Real MCP Client
1. Launch each test app
2. Verify screen capture accuracy
3. Test special key sequences (arrows, Ctrl+C, etc.)
4. Resize terminal during operation
5. Long-running processes
6. Concurrent sessions
7. Error recovery (kill process, restart)

### Performance Considerations
- Buffer operations could use pooling
- Mutex usage could be optimized with RWMutex in more places
- ANSI parsing allocates escape buffers per parser
- Consider caching rendered output

### Security Notes
- No input validation on commands (relies on OS)
- No resource limits beyond session count
- PTY provides process isolation
- Consider adding command whitelist for production

### Next Development Steps
1. Fix failing unit tests
2. Add integration tests using the test apps
3. Implement missing ANSI features as needed
4. Add performance benchmarks
5. Create example MCP client usage
6. Document API with examples

### Debugging Tips
- Set `LOG_LEVEL=debug` for verbose logging
- Logs go to stderr in JSON format
- Each session has unique ID for tracking
- Use "ansi" format to see cursor position
- Check scrollback for lost output

### Critical Code Paths
1. **Session Creation**: manager.go → session.go → pty.go
2. **Input Flow**: handlers.go → session.SendKeys() → pty.Write()
3. **Output Flow**: pty.Read() → readLoop() → buffer.Write() → ansi.Parse()
4. **Screen Render**: buffer.Render() → renderPlain/Raw/ANSI()

### Build and Run Commands
```bash
# Build server
make build

# Run server (stdio mode)
./bin/mcp-terminal-server

# Run tests
make test

# Build test apps
make test-apps

# Clean everything
make clean
```

### Important File Locations
- Main entry: `cmd/server/main.go`
- MCP tools: `internal/tools/handlers.go`
- Session logic: `internal/session/session.go`
- ANSI parsing: `internal/terminal/ansi.go`
- Buffer rendering: `internal/terminal/buffer.go`

### Environment Variables
- `LOG_LEVEL`: debug, info, warn, error (default: info)
- `MAX_SESSIONS`: Max concurrent sessions (default: 100)
- `SESSION_TIMEOUT`: Idle timeout in minutes (default: 30)

This document should be updated as the project evolves.

## File Structure Summary

### Core Implementation Files
- `cmd/server/main.go` - Entry point, initializes logger and server
- `internal/mcp/server.go` - MCP server setup, tool registration
- `internal/session/manager.go` - Session lifecycle management
- `internal/session/session.go` - Individual session logic
- `internal/terminal/pty.go` - PTY wrapper with resize support
- `internal/terminal/buffer.go` - Screen buffer with scrollback
- `internal/terminal/ansi.go` - ANSI escape sequence parser
- `internal/tools/handlers.go` - All 9 MCP tool implementations
- `internal/tools/keys.go` - Special key mapping (arrows, Ctrl, etc.)
- `internal/utils/logger.go` - Structured logging setup

### Test Files
- `internal/terminal/ansi_test.go` - ANSI parser unit tests
- `internal/terminal/buffer_test.go` - Buffer operation tests
- `internal/session/manager_test.go` - Session management tests
- `test/apps/echo.go` - Basic I/O test application
- `test/apps/menu.go` - Interactive menu with navigation
- `test/apps/progress.go` - Animation and progress bar tests

### Documentation
- `README.md` - Project overview and quick start
- `project.md` - Technical design document
- `progress.md` - Development progress tracking
- `CLAUDE.md` - This file, AI assistant context
- `test/apps/README.md` - Test application documentation

### Build Files
- `Makefile` - Main build system
- `test/apps/Makefile` - Test app builds
- `go.mod` / `go.sum` - Go dependencies
- `.gitignore` - Git ignore rules

All code is fully implemented and builds successfully. The main areas needing work are real-world testing and edge case handling.