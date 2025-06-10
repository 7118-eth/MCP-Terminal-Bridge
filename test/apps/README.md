# Test Applications

This directory contains test applications for the MCP Terminal Tester.

## Applications

### echo.go
A simple echo application that tests basic input/output and ANSI colors.

**Features:**
- Echo user input
- Clear screen command
- ANSI color test
- Help command

**Build & Run:**
```bash
go build -o echo echo.go
./echo
```

### menu.go
An interactive menu system that tests cursor movement and terminal UI features.

**Features:**
- Arrow key navigation
- Box drawing characters
- Cursor positioning
- Color attributes
- 256-color palette display

**Build & Run:**
```bash
go build -o menu menu.go
./menu
```

### progress.go
Progress bars and animation tests for ANSI escape sequence handling.

**Features:**
- Simple progress bar
- Colored progress bar
- Spinner animations
- Multi-line progress tracking
- Wave animation

**Build & Run:**
```bash
go build -o progress progress.go
./progress
```

## Testing with MCP

To test these applications with the MCP Terminal Tester:

1. Start the MCP server
2. Use the `launch_app` tool to start an application:
   ```json
   {
     "command": "./test/apps/echo",
     "args": []
   }
   ```
3. Use `view_screen` to see the output
4. Use `send_keys` to interact with the application
5. Test different output formats: "plain", "raw", "ansi", "scrollback"

## What to Test

1. **Basic I/O**: Can the app receive input and display output?
2. **ANSI Support**: Are colors and formatting preserved?
3. **Cursor Control**: Does cursor positioning work correctly?
4. **Special Keys**: Do arrow keys, Enter, Escape work?
5. **Screen Updates**: Are animations and progress bars captured?
6. **Scrollback**: Is historical output preserved?