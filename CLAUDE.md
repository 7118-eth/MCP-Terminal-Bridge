# AI Assistant Context for MCP Terminal Tester

## Project Status (as of January 11, 2025)

### Current State - PHASE 3 COMPLETE ✅
- **Phase 3 COMPLETE**: Core production features implemented
- **9 MCP tools** fully implemented and working
- **Integration tests**: ALL 18 PASSING ✅ 
- **Unit tests**: All passing ✅
- **Error recovery**: ✅ Panic recovery in readLoop, graceful cleanup
- **Input validation**: ✅ Comprehensive validation for all tools (command injection, path traversal, UUID format, etc.)
- **Performance optimizations**: ⚠️ Buffer pool defined but not actively used
- **API documentation**: ✅ Complete with examples (API.md)
- **Test applications**: ✅ 4 apps including fully functional vim-like editor

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

### Recent Changes (Phase 3 Completion)

1. **Input Validation** ✅:
   - Command injection prevention (blocks `;|&` characters)
   - Path traversal protection
   - UUID format validation for session IDs
   - Environment variable key/value validation
   - Length limits on all string inputs
   - Format validation for output types

2. **Error Recovery** ✅:
   - Panic recovery in session readLoop
   - Graceful PTY cleanup on errors
   - Proper WaitGroup management for goroutines
   - Non-blocking resize operations

3. **Session Restart Fix** ✅:
   - Proper readLoop lifecycle management
   - Clean done channel handling
   - Wait for old readLoop before starting new one

4. **vim Test Application** ✅:
   - Full vim-like editor implementation
   - Normal, Insert, Command modes
   - Navigation (hjkl, 0, $, g, G)
   - Basic editing (i, a, o, O, x, d)
   - File save/load support

5. **All Integration Tests Passing** ✅:
   - Fixed TestRestartApp with proper lifecycle management
   - Fixed TestAnsiOutput with proper SGR rendering
   - Fixed TestScrollbackFormat with correct buffer handling
   - Fixed TestMenuApp and TestProgressApp timing issues

### Known Issues & Limitations

1. **Performance Optimizations Partial**:
   - Buffer pool defined in ansi.go but not actively used
   - Most operations still use regular Mutex instead of RWMutex
   - No output caching implemented
   - Escape buffer allocations still happen per parser

2. **Not Implemented**:
   - Mouse support
   - Alternate screen buffer
   - Some advanced ANSI modes (DEC private modes)
   - Session persistence across server restarts
   - Rate limiting for input
   - True raw ANSI passthrough (currently regenerates SGR sequences)

3. **Platform Specific**:
   - SIGWINCH handling may vary on different OS
   - Terminal mode setting is simplified
   - Windows support needs testing

### Testing Strategy

#### Unit Tests
- Run with: `make test`
- All unit tests passing ✅
- Coverage reports: `make test-coverage`
- Specific suites: `make test-terminal` or `make test-session`

#### Integration Tests
- Run with: `make test-integration`
- ALL 18/18 tests passing ✅
- Framework in `test/integration/framework_test.go`
- Tool tests in `test/integration/tools_test.go`
- Test app tests in `test/integration/testapps_test.go`

#### Test Applications
Located in `test/apps/`:
- **echo.go**: Basic I/O, commands, color test
- **menu.go**: Arrow keys, box drawing, 256 colors
- **progress.go**: Animations, multi-line updates
- **vim.go**: Full vim-like editor with modes and file operations

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
- Buffer pool defined but not actively used - need to implement Get/Put calls
- Most operations use regular Mutex - convert read-heavy ops to RWMutex
- ANSI parser has pool but doesn't use it for escape buffers
- No output caching implemented yet
- Consider using strings.Builder for render operations

### Security Notes
- ✅ Input validation prevents command injection (`;|&` blocked) and path traversal
- ✅ UUID validation for session IDs prevents injection
- ✅ Environment variable validation (key/value limits and character checks)
- No resource limits beyond session count
- PTY provides process isolation
- Consider adding command whitelist for production use

### Next Development Steps (Phase 4)
1. ✅ ~~Fix remaining integration test failures~~ - ALL TESTS PASSING
2. ✅ ~~Add error recovery from PTY crashes~~ - Implemented
3. ✅ ~~Implement input validation for all tools~~ - Complete
4. ✅ ~~Create vim-like test application~~ - Implemented
5. ✅ ~~Write API documentation with examples~~ - API.md created
6. ⏳ Activate buffer pooling (pool defined but not used)
7. ⏳ Convert Mutex to RWMutex for read operations
8. ⏳ Add true raw ANSI passthrough mode
9. ⏳ Implement performance benchmarks
10. ⏳ Add mouse support and alternate screen buffer

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

## Phase 3 Achievements ✅
1. ✅ Created integration test framework for end-to-end testing
2. ✅ Tested all 9 MCP tools in integration tests
3. ✅ Added error recovery from PTY crashes with panic handling
4. ✅ Implemented comprehensive input validation for all tools
5. ✅ Created vim-like test application with full functionality
6. ✅ Wrote complete API documentation with examples (API.md)
7. ✅ Fixed all integration test failures (18/18 passing)
8. ✅ Added graceful session cleanup on errors

## Phase 4 TODO List - Performance & Advanced Features
1. ⏳ Activate buffer pooling (currently defined but unused)
2. ⏳ Convert Mutex to RWMutex for read-heavy operations
3. ⏳ Add true raw ANSI passthrough (preserve original sequences)
4. ⏳ Implement performance benchmarks
5. ⏳ Add output caching for frequently accessed screens
6. ⏳ Implement mouse support
7. ⏳ Add alternate screen buffer support
8. ⏳ Implement session persistence
9. ⏳ Add rate limiting for input
10. ⏳ Create profiling benchmarks