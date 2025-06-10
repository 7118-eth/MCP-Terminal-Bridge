# MCP Terminal Testing Server - Development Prompt

## Mission
Build an MCP (Model Context Protocol) server in Go that allows AI assistants to test and interact with terminal/TUI applications. This server will act as a bridge between Claude and terminal applications, providing visibility and control over terminal UI testing.

## Problem Statement
AI assistants cannot directly see or interact with terminal applications. This MCP server will:
1. Launch terminal applications as subprocesses
2. Capture their visual output
3. Send keyboard input
4. Provide terminal state information

## Technical Requirements

### Core Technology
- **Language**: Go
- **MCP SDK**: github.com/mark3labs/mcp-go
- **Terminal Handling**: Use pseudo-terminal (PTY) for proper terminal emulation
- **ANSI Processing**: Parse and optionally strip ANSI escape codes

### MCP Tools to Implement

1. **launch_app**
   - Parameters: command (string), args ([]string), env (map[string]string)
   - Returns: session_id, success status
   - Launches app in PTY with specified size (80x24 default)

2. **view_screen**
   - Parameters: session_id (string), format (enum: "raw", "plain", "ansi")
   - Returns: Current terminal content
   - Formats: raw (with ANSI), plain (stripped), ansi (with markers)

3. **send_keys**
   - Parameters: session_id (string), keys (string)
   - Special keys: "Enter", "Esc", "Tab", "Backspace", "Ctrl+C", etc.
   - Returns: Success status

4. **get_cursor_position**
   - Parameters: session_id (string)
   - Returns: {row: int, col: int}

5. **get_screen_size**
   - Parameters: session_id (string)
   - Returns: {rows: int, cols: int}

6. **restart_app**
   - Parameters: session_id (string)
   - Returns: New session_id

7. **stop_app**
   - Parameters: session_id (string)
   - Returns: Success status

8. **list_sessions**
   - Returns: Array of active sessions with metadata

### Implementation Details

1. **Session Management**
   - Support multiple concurrent sessions
   - Each session has unique ID
   - Clean up resources on disconnect

2. **PTY Handling**
   - Use `github.com/creack/pty` or similar
   - Handle window resize events
   - Proper signal forwarding

3. **Output Buffering**
   - Maintain screen buffer per session
   - Handle partial ANSI sequences
   - Detect screen updates

4. **Input Processing**
   - Convert special key names to proper sequences
   - Support key combinations (Ctrl+, Alt+, etc.)
   - Queue inputs if needed

### Example Usage Flow
```
1. Claude: launch_app("go", ["run", "cmd/budget/main.go"])
   → Returns: session_123

2. Claude: view_screen("session_123", "plain")
   → Returns: "Budget Tracker\n\n1. View Assets\n2. Add Asset..."

3. Claude: send_keys("session_123", "2")
   → Returns: success

4. Claude: view_screen("session_123", "plain")
   → Returns: "Add New Asset\n\nEnter asset symbol: _"
```

### Testing the MCP Server
1. Create simple test apps (echo input, menu system)
2. Test special key handling
3. Verify ANSI code processing
4. Test concurrent sessions

### Project Structure
```
mcp-terminal-test/
├── cmd/server/main.go
├── internal/
│   ├── terminal/     # PTY and terminal handling
│   ├── session/      # Session management
│   └── tools/        # MCP tool implementations
├── go.mod
└── README.md
```

### Success Criteria
1. Can launch and control terminal applications
2. Accurately captures terminal output
3. Handles all common keyboard inputs
4. Supports TUI frameworks (Bubble Tea, etc.)
5. Clean session lifecycle management

### Resources
- MCP Go SDK: https://github.com/mark3labs/mcp-go
- PTY handling: https://github.com/creack/pty
- ANSI parsing: Consider github.com/Azure/go-ansiterm

## Getting Started
1. Initialize Go module: `go mod init mcp-terminal-test`
2. Install dependencies
3. Implement basic launch_app and view_screen tools first
4. Test with simple applications
5. Iterate on remaining tools

Build this server to enable AI assistants to effectively test and interact with terminal applications!