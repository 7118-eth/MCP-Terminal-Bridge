# AI Assistant Context for MCP Terminal Tester

## Project Status (as of June 11, 2025)

### Current State
- **Phase 3 IN PROGRESS**: Integration testing and robustness improvements
- **9 MCP tools** fully implemented and working
- **Integration tests**: 13 out of 18 passing
- Builds successfully with `make build`
- Unit tests: All passing ✅
- Integration tests: Created comprehensive test framework

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

### Recent Changes (Phase 3)

1. **Integration Test Framework**:
   - Created `test/integration/framework_test.go` with test harness
   - Added mock CallToolRequest implementation
   - Tests simulate real MCP client interactions

2. **Parameter Type Fixes**:
   - Fixed ResizeTerminal to accept both int and float64 for width/height
   - Added debug logging to track parameter types
   - Fixed argument extraction in LaunchApp

3. **Test Improvements**:
   - Updated tests to use `sh -c` with sleep to keep sessions alive
   - Fixed immediate command termination issues
   - Added proper error handling in test framework

### Known Issues & Limitations

1. **Integration Test Failures** (5 out of 18):
   - **TestRestartApp**: Session becomes inactive after restart (readLoop lifecycle issue)
   - **TestScrollbackFormat**: Not detecting historical content
   - **TestAnsiOutput**: Raw format not preserving ANSI escape sequences
   - **TestMenuApp**: Timing out (5s) - test app interaction issue
   - **TestProgressApp**: Timing out (5s) - test app interaction issue

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
- Run with: `make test`
- All unit tests passing ✅
- Coverage reports: `make test-coverage`
- Specific suites: `make test-terminal` or `make test-session`

#### Integration Tests
- Run with: `make test-integration`
- 13/18 tests passing
- Framework in `test/integration/framework_test.go`
- Tool tests in `test/integration/tools_test.go`
- Test app tests in `test/integration/testapps_test.go`

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
1. Fix remaining integration test failures (5 tests)
2. Add error recovery from PTY crashes (in progress)
3. Implement input validation for all tools
4. Add performance optimizations (buffer pooling, RWMutex)
5. Create vim-like test application
6. Write API documentation with examples
7. Profile and optimize hot paths
8. Add graceful session cleanup on errors

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

# Run all unit tests
make test

# Run integration tests
make test-integration

# Run all tests (unit + integration)
make test-all

# Build test apps
make test-apps

# Clean everything
make clean

# With debug logging
LOG_LEVEL=debug ./bin/mcp-terminal-server
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

All code is fully implemented and builds successfully. The main areas needing work are:
1. Fixing remaining integration test failures
2. Error recovery and robustness improvements
3. Performance optimizations
4. Real-world testing with actual MCP clients

## Current TODO List
1. ✅ Create integration test framework for end-to-end testing
2. ✅ Test all 9 MCP tools in integration tests
3. 🚧 Add error recovery from PTY crashes (in progress)
4. ⏳ Implement input validation for all tools
5. ⏳ Add buffer pooling for performance
6. ⏳ Convert appropriate mutexes to RWMutex
7. ⏳ Create vim-like test application
8. ⏳ Write API documentation with examples
9. ⏳ Profile code and optimize hot paths
10. ⏳ Add graceful session cleanup on errors