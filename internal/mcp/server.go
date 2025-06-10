package mcp

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bioharz/mcp-terminal-tester/internal/session"
	"github.com/bioharz/mcp-terminal-tester/internal/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	mcpServer       *server.MCPServer
	sessionManager  *session.Manager
}

func NewServer() (*Server, error) {
	slog.Info("Creating MCP server")
	
	// Create session manager
	sm := session.NewManager()

	// Create MCP server instance
	mcpServer := server.NewMCPServer(
		"mcp-terminal-tester",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	s := &Server{
		mcpServer:      mcpServer,
		sessionManager: sm,
	}

	// Register tools
	if err := s.registerTools(); err != nil {
		slog.Error("Failed to register tools", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	// Start session cleanup routine
	sm.StartCleanupRoutine()

	slog.Info("MCP server created successfully", slog.Int("tools_registered", 8))
	return s, nil
}

func (s *Server) registerTools() error {
	slog.Debug("Registering MCP tools")
	
	// Create tool handlers with session manager
	toolHandlers := tools.NewHandlers(s.sessionManager)

	// Register launch_app tool
	launchTool := mcp.NewTool("launch_app",
		mcp.WithDescription("Launch a new terminal application"),
		mcp.WithString("command",
			mcp.Required(),
			mcp.Description("The command to execute"),
		),
		mcp.WithArray("args",
			mcp.Description("Command arguments"),
			mcp.Items(map[string]any{"type": "string"}),
		),
		mcp.WithObject("env",
			mcp.Description("Environment variables"),
		),
	)
	s.mcpServer.AddTool(launchTool, toolHandlers.LaunchApp)

	// Register view_screen tool
	viewTool := mcp.NewTool("view_screen",
		mcp.WithDescription("Get the current terminal screen content"),
		mcp.WithString("session_id",
			mcp.Required(),
			mcp.Description("The session ID"),
		),
		mcp.WithString("format",
			mcp.Description("Output format"),
			mcp.Enum("plain", "raw", "ansi"),
			mcp.DefaultString("plain"),
		),
	)
	s.mcpServer.AddTool(viewTool, toolHandlers.ViewScreen)

	// Register send_keys tool
	sendKeysTool := mcp.NewTool("send_keys",
		mcp.WithDescription("Send keyboard input to the terminal"),
		mcp.WithString("session_id",
			mcp.Required(),
			mcp.Description("The session ID"),
		),
		mcp.WithString("keys",
			mcp.Required(),
			mcp.Description("The keys to send"),
		),
	)
	s.mcpServer.AddTool(sendKeysTool, toolHandlers.SendKeys)

	// Register get_cursor_position tool
	cursorTool := mcp.NewTool("get_cursor_position",
		mcp.WithDescription("Get the current cursor position"),
		mcp.WithString("session_id",
			mcp.Required(),
			mcp.Description("The session ID"),
		),
	)
	s.mcpServer.AddTool(cursorTool, toolHandlers.GetCursorPosition)

	// Register get_screen_size tool
	sizeTool := mcp.NewTool("get_screen_size",
		mcp.WithDescription("Get the terminal screen dimensions"),
		mcp.WithString("session_id",
			mcp.Required(),
			mcp.Description("The session ID"),
		),
	)
	s.mcpServer.AddTool(sizeTool, toolHandlers.GetScreenSize)

	// Register restart_app tool
	restartTool := mcp.NewTool("restart_app",
		mcp.WithDescription("Restart a terminal session"),
		mcp.WithString("session_id",
			mcp.Required(),
			mcp.Description("The session ID"),
		),
	)
	s.mcpServer.AddTool(restartTool, toolHandlers.RestartApp)

	// Register stop_app tool
	stopTool := mcp.NewTool("stop_app",
		mcp.WithDescription("Stop a terminal session"),
		mcp.WithString("session_id",
			mcp.Required(),
			mcp.Description("The session ID"),
		),
	)
	s.mcpServer.AddTool(stopTool, toolHandlers.StopApp)

	// Register list_sessions tool
	listTool := mcp.NewTool("list_sessions",
		mcp.WithDescription("List all active terminal sessions"),
	)
	s.mcpServer.AddTool(listTool, toolHandlers.ListSessions)

	// Register resize_terminal tool
	resizeTool := mcp.NewTool("resize_terminal",
		mcp.WithDescription("Resize the terminal window"),
		mcp.WithString("session_id",
			mcp.Required(),
			mcp.Description("The session ID"),
		),
		mcp.WithNumber("width",
			mcp.Required(),
			mcp.Description("Terminal width in columns"),
			mcp.Min(1),
			mcp.Max(500),
		),
		mcp.WithNumber("height",
			mcp.Required(),
			mcp.Description("Terminal height in rows"),
			mcp.Min(1),
			mcp.Max(200),
		),
	)
	s.mcpServer.AddTool(resizeTool, toolHandlers.ResizeTerminal)

	slog.Debug("All tools registered successfully")
	return nil
}

func (s *Server) Run(ctx context.Context) error {
	slog.Info("Starting MCP server in stdio mode")
	err := server.ServeStdio(s.mcpServer)
	if err != nil {
		slog.Error("MCP server error", slog.String("error", err.Error()))
	}
	return err
}