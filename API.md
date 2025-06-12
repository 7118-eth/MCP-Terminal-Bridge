# MCP Terminal Tester - API Documentation

The MCP Terminal Tester provides 9 tools for interacting with terminal applications through the Model Context Protocol (MCP).

## Tools Overview

| Tool | Purpose | Parameters |
|------|---------|------------|
| `launch_app` | Start a new terminal application | command, args, env |
| `view_screen` | Get terminal content | session_id, format |
| `send_keys` | Send keyboard input | session_id, keys |
| `get_cursor_position` | Get cursor coordinates | session_id |
| `get_screen_size` | Get terminal dimensions | session_id |
| `resize_terminal` | Change terminal size | session_id, width, height |
| `restart_app` | Restart an application | session_id |
| `stop_app` | Terminate an application | session_id |
| `list_sessions` | List all active sessions | none |

## Tool Reference

### launch_app

Starts a new terminal application and returns a session ID for further interactions.

**Parameters:**
- `command` (string, required): The command to execute
- `args` (array of strings, optional): Command line arguments
- `env` (object, optional): Environment variables as key-value pairs

**Returns:**
- `session_id`: Unique identifier for the session
- `success`: Boolean indicating success

**Example:**
```json
{
  "name": "launch_app",
  "arguments": {
    "command": "vim",
    "args": ["test.txt"],
    "env": {
      "TERM": "xterm-256color",
      "EDITOR": "vim"
    }
  }
}
```

**Response:**
```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "success": true
}
```

### view_screen

Captures the current terminal screen content in various formats.

**Parameters:**
- `session_id` (string, required): Session identifier
- `format` (string, optional): Output format (default: "plain")
  - `plain`: Text only, ANSI sequences stripped
  - `raw`: Full output with ANSI escape sequences reconstructed from cell attributes
  - `ansi`: Debug format showing cursor position with ▮
  - `scrollback`: Includes scrollback buffer history
  - `passthrough`: Original data exactly as received, preserving all ANSI sequences

**Returns:**
- `content`: The screen content
- `cursor`: Object with cursor position (`row`, `col`)

**Example:**
```json
{
  "name": "view_screen",
  "arguments": {
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "format": "plain"
  }
}
```

**Response:**
```json
{
  "content": "Hello, World!\nThis is line 2\n                ",
  "cursor": {
    "row": 1,
    "col": 0
  }
}
```

### send_keys

Sends keyboard input to the terminal application.

**Parameters:**
- `session_id` (string, required): Session identifier
- `keys` (string, required): Keys to send

**Special Key Sequences:**
- `Enter`: Enter/Return key
- `Tab`: Tab key
- `Escape`: Escape key
- `Backspace`: Backspace key
- `Up`, `Down`, `Left`, `Right`: Arrow keys
- `Ctrl+C`: Control+C
- `Ctrl+D`: Control+D
- `F1`-`F12`: Function keys

**Returns:**
- `success`: Boolean indicating success

**Example:**
```json
{
  "name": "send_keys",
  "arguments": {
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "keys": "iHello, World!Escape"
  }
}
```

### get_cursor_position

Gets the current cursor position in the terminal.

**Parameters:**
- `session_id` (string, required): Session identifier

**Returns:**
- `row`: Cursor row (0-based)
- `col`: Cursor column (0-based)

**Example:**
```json
{
  "name": "get_cursor_position",
  "arguments": {
    "session_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response:**
```json
{
  "row": 5,
  "col": 12
}
```

### get_screen_size

Gets the current terminal dimensions.

**Parameters:**
- `session_id` (string, required): Session identifier

**Returns:**
- `width`: Terminal width in columns
- `height`: Terminal height in rows

**Example:**
```json
{
  "name": "get_screen_size",
  "arguments": {
    "session_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response:**
```json
{
  "width": 80,
  "height": 24
}
```

### resize_terminal

Changes the terminal size. Applications will receive a SIGWINCH signal.

**Parameters:**
- `session_id` (string, required): Session identifier
- `width` (number, required): New width in columns (1-1000)
- `height` (number, required): New height in rows (1-1000)

**Returns:**
- `success`: Boolean indicating success

**Example:**
```json
{
  "name": "resize_terminal",
  "arguments": {
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "width": 120,
    "height": 30
  }
}
```

### restart_app

Restarts the application with the same command, arguments, and environment.

**Parameters:**
- `session_id` (string, required): Session identifier

**Returns:**
- `success`: Boolean indicating success

**Example:**
```json
{
  "name": "restart_app",
  "arguments": {
    "session_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

### stop_app

Terminates the application and removes the session.

**Parameters:**
- `session_id` (string, required): Session identifier

**Returns:**
- `success`: Boolean indicating success

**Example:**
```json
{
  "name": "stop_app",
  "arguments": {
    "session_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

### list_sessions

Lists all active sessions with their information.

**Parameters:** None

**Returns:**
- `sessions`: Array of session objects

**Example:**
```json
{
  "name": "list_sessions",
  "arguments": {}
}
```

**Response:**
```json
{
  "sessions": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "command": "vim",
      "args": ["test.txt"],
      "created": "2025-01-11T10:30:00Z",
      "last_active": "2025-01-11T10:35:00Z",
      "state": "active"
    }
  ]
}
```

## Common Workflows

### Testing a Text Editor

```json
// 1. Launch vim
{
  "name": "launch_app",
  "arguments": {
    "command": "vim",
    "args": ["test.txt"]
  }
}

// 2. Wait for vim to load and view the screen
{
  "name": "view_screen",
  "arguments": {
    "session_id": "session-id",
    "format": "plain"
  }
}

// 3. Enter insert mode and type text
{
  "name": "send_keys",
  "arguments": {
    "session_id": "session-id",
    "keys": "iHello, World!"
  }
}

// 4. Exit insert mode and save
{
  "name": "send_keys",
  "arguments": {
    "session_id": "session-id",
    "keys": "Escape:wqEnter"
  }
}
```

### Testing Terminal Applications with Colors

```json
// 1. Launch a colorized application
{
  "name": "launch_app",
  "arguments": {
    "command": "ls",
    "args": ["--color=always"],
    "env": {
      "TERM": "xterm-256color"
    }
  }
}

// 2. View with raw format to see ANSI sequences
{
  "name": "view_screen",
  "arguments": {
    "session_id": "session-id",
    "format": "raw"
  }
}
```

### Testing Interactive Menus

```json
// 1. Launch menu application
{
  "name": "launch_app",
  "arguments": {
    "command": "./menu"
  }
}

// 2. Navigate with arrow keys
{
  "name": "send_keys",
  "arguments": {
    "session_id": "session-id",
    "keys": "DownDownEnter"
  }
}

// 3. Check the result
{
  "name": "view_screen",
  "arguments": {
    "session_id": "session-id",
    "format": "plain"
  }
}
```

### Testing Terminal Resize Behavior

```json
// 1. Launch an application
{
  "name": "launch_app",
  "arguments": {
    "command": "htop"
  }
}

// 2. Check initial size
{
  "name": "get_screen_size",
  "arguments": {
    "session_id": "session-id"
  }
}

// 3. Resize terminal
{
  "name": "resize_terminal",
  "arguments": {
    "session_id": "session-id",
    "width": 120,
    "height": 40
  }
}

// 4. Verify the application adapted to new size
{
  "name": "view_screen",
  "arguments": {
    "session_id": "session-id",
    "format": "plain"
  }
}
```

## Error Handling

All tools return errors in a consistent format:

```json
{
  "error": {
    "message": "session_id parameter is required",
    "code": "INVALID_PARAMETER"
  }
}
```

### Common Error Conditions

- **Invalid session_id**: Session not found or invalid UUID format
- **Session not active**: Application has terminated
- **Invalid parameters**: Missing required parameters or invalid values
- **Command not found**: Specified command doesn't exist
- **Permission denied**: Insufficient permissions to execute command
- **Invalid format**: Unsupported output format specified

## Input Validation

The MCP Terminal Tester includes comprehensive input validation:

### Session IDs
- Must be valid UUID format
- Must reference an existing, active session

### Commands
- Cannot contain command injection characters (`;`, `|`, `&`)
- Cannot contain path traversal sequences (`..`)
- Must be valid executable names or paths

### Arguments
- Maximum 1000 characters per argument
- Cannot contain path traversal sequences

### Environment Variables
- Keys: Maximum 100 characters, no `=` or null bytes
- Values: Maximum 1000 characters

### Keys Parameter
- Maximum 10000 characters
- Supports special key sequences as documented

### Format Parameter
- Must be one of: `plain`, `raw`, `ansi`, `scrollback`

### Dimensions
- Width and height must be between 1 and 1000

## Performance Considerations

- **Buffer Pooling**: The server uses buffer pools to reduce garbage collection
- **Concurrent Sessions**: Supports up to 100 concurrent sessions by default
- **Session Cleanup**: Idle sessions are automatically cleaned up after 30 minutes
- **Memory Management**: Uses efficient data structures for screen buffers and ANSI parsing

## Security Features

- **Command Validation**: Prevents command injection attacks
- **Path Traversal Protection**: Blocks directory traversal attempts
- **Resource Limits**: Enforces limits on parameter sizes and session counts
- **Input Sanitization**: Validates all user inputs before processing