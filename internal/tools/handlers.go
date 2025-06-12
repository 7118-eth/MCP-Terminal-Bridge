package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/bioharz/mcp-terminal-tester/internal/session"
	"github.com/bioharz/mcp-terminal-tester/internal/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

type Handlers struct {
	sessionManager *session.Manager
}

func NewHandlers(sm *session.Manager) *Handlers {
	return &Handlers{
		sessionManager: sm,
	}
}

// Input validation functions
func validateSessionID(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session_id parameter is required")
	}
	// Basic UUID format validation
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !uuidRegex.MatchString(sessionID) {
		return fmt.Errorf("session_id must be a valid UUID")
	}
	return nil
}

func validateCommand(command string) error {
	if command == "" {
		return fmt.Errorf("command parameter is required")
	}
	// Prevent command injection and ensure safe commands
	if strings.Contains(command, ";") || strings.Contains(command, "|") || strings.Contains(command, "&") {
		return fmt.Errorf("command contains invalid characters (;|&)")
	}
	// Prevent path traversal
	if strings.Contains(command, "..") {
		return fmt.Errorf("command contains path traversal (..)")
	}
	return nil
}

func validateArguments(args []string) error {
	for i, arg := range args {
		if len(arg) > 1000 {
			return fmt.Errorf("argument %d exceeds maximum length (1000 characters)", i)
		}
		// Prevent certain dangerous arguments
		if strings.Contains(arg, "../") || strings.Contains(arg, "..\\") {
			return fmt.Errorf("argument %d contains path traversal", i)
		}
	}
	return nil
}

func validateEnvironment(env map[string]string) error {
	for key, value := range env {
		if len(key) > 100 {
			return fmt.Errorf("environment key '%s' exceeds maximum length (100 characters)", key)
		}
		if len(value) > 1000 {
			return fmt.Errorf("environment value for '%s' exceeds maximum length (1000 characters)", key)
		}
		// Prevent environment variable injection
		if strings.Contains(key, "=") || strings.Contains(key, "\x00") {
			return fmt.Errorf("environment key '%s' contains invalid characters", key)
		}
	}
	return nil
}

func validateKeys(keys string) error {
	if keys == "" {
		return fmt.Errorf("keys parameter is required")
	}
	if len(keys) > 10000 {
		return fmt.Errorf("keys parameter exceeds maximum length (10000 characters)")
	}
	return nil
}

func validateFormat(format string) error {
	validFormats := []string{"plain", "raw", "ansi", "scrollback", "passthrough"}
	for _, valid := range validFormats {
		if format == valid {
			return nil
		}
	}
	return fmt.Errorf("format must be one of: %s", strings.Join(validFormats, ", "))
}

func validateDimensions(width, height float64) error {
	if width < 1 || width > 1000 {
		return fmt.Errorf("width must be between 1 and 1000")
	}
	if height < 1 || height > 1000 {
		return fmt.Errorf("height must be between 1 and 1000")
	}
	return nil
}

func (h *Handlers) LaunchApp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	utils.LogToolCall("launch_app", "")
	
	args := request.GetArguments()
	command, ok := args["command"].(string)
	if !ok {
		err := fmt.Errorf("command parameter is required")
		slog.Error("Invalid tool call", 
			slog.String("tool", "launch_app"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	// Validate command
	if err := validateCommand(command); err != nil {
		slog.Error("Invalid command", 
			slog.String("tool", "launch_app"),
			slog.String("command", command),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	// Extract args if provided
	var cmdArgs []string
	if argsParam, exists := args["args"]; exists {
		// Try []interface{} first
		if argsArray, ok := argsParam.([]interface{}); ok {
			for _, arg := range argsArray {
				if argStr, ok := arg.(string); ok {
					cmdArgs = append(cmdArgs, argStr)
				}
			}
		} else if argsArray, ok := argsParam.([]string); ok {
			// Also try []string directly
			cmdArgs = argsArray
		}
		slog.Debug("Extracted args", 
			slog.String("tool", "launch_app"),
			slog.Any("args", cmdArgs),
			slog.Any("raw_args", argsParam),
		)
		
		// Validate arguments
		if err := validateArguments(cmdArgs); err != nil {
			slog.Error("Invalid arguments", 
				slog.String("tool", "launch_app"),
				slog.Any("args", cmdArgs),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
	}

	// Extract env if provided
	env := make(map[string]string)
	if envParam, exists := args["env"]; exists {
		if envMap, ok := envParam.(map[string]interface{}); ok {
			for k, v := range envMap {
				if vStr, ok := v.(string); ok {
					env[k] = vStr
				}
			}
		}
		
		// Validate environment
		if err := validateEnvironment(env); err != nil {
			slog.Error("Invalid environment", 
				slog.String("tool", "launch_app"),
				slog.Any("env", env),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
	}

	// Create new session
	sess, err := h.sessionManager.CreateSession(command, cmdArgs, env)
	if err != nil {
		utils.LogError(err, "Failed to launch app",
			slog.String("tool", "launch_app"),
			slog.String("command", command),
		)
		return nil, fmt.Errorf("failed to launch app: %w", err)
	}

	slog.Info("App launched successfully",
		slog.String("tool", "launch_app"),
		slog.String("session_id", sess.ID),
		slog.String("command", command),
	)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf(`{"session_id": "%s", "success": true}`, sess.ID),
			},
		},
	}, nil
}

func (h *Handlers) ViewScreen(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	sessionID, ok := args["session_id"].(string)
	if !ok {
		err := fmt.Errorf("session_id parameter is required")
		slog.Error("Invalid tool call",
			slog.String("tool", "view_screen"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	// Validate session ID
	if err := validateSessionID(sessionID); err != nil {
		slog.Error("Invalid session ID",
			slog.String("tool", "view_screen"),
			slog.String("session_id", sessionID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	utils.LogToolCall("view_screen", sessionID)

	format := "plain"
	if formatParam, exists := args["format"]; exists {
		if f, ok := formatParam.(string); ok {
			format = f
		}
	}
	
	// Validate format
	if err := validateFormat(format); err != nil {
		slog.Error("Invalid format",
			slog.String("tool", "view_screen"),
			slog.String("format", format),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	sess, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	content, err := sess.GetScreen(format)
	if err != nil {
		return nil, err
	}

	row, col := sess.GetCursorPosition()

	// Create response object and marshal to JSON properly
	response := map[string]interface{}{
		"content": content,
		"cursor": map[string]interface{}{
			"row": row,
			"col": col,
		},
	}
	
	respData, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(respData),
			},
		},
	}, nil
}

func (h *Handlers) SendKeys(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	sessionID, ok := args["session_id"].(string)
	if !ok {
		err := fmt.Errorf("session_id parameter is required")
		slog.Error("Invalid tool call",
			slog.String("tool", "send_keys"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	// Validate session ID
	if err := validateSessionID(sessionID); err != nil {
		slog.Error("Invalid session ID",
			slog.String("tool", "send_keys"),
			slog.String("session_id", sessionID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	keys, ok := args["keys"].(string)
	if !ok {
		err := fmt.Errorf("keys parameter is required")
		slog.Error("Invalid tool call",
			slog.String("tool", "send_keys"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	// Validate keys
	if err := validateKeys(keys); err != nil {
		slog.Error("Invalid keys",
			slog.String("tool", "send_keys"),
			slog.String("keys", keys),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	utils.LogToolCall("send_keys", sessionID, slog.Int("key_count", len(keys)))

	sess, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// Map special keys
	mappedKeys := MapKeys(keys)
	if mappedKeys != keys {
		slog.Debug("Keys mapped",
			slog.String("original", keys),
			slog.String("mapped", fmt.Sprintf("%q", mappedKeys)),
		)
	}

	if err := sess.SendKeys(mappedKeys); err != nil {
		utils.LogError(err, "Failed to send keys",
			slog.String("tool", "send_keys"),
			slog.String("session_id", sessionID),
		)
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: `{"success": true}`,
			},
		},
	}, nil
}

func (h *Handlers) GetCursorPosition(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	sessionID, ok := args["session_id"].(string)
	if !ok {
		err := fmt.Errorf("session_id parameter is required")
		slog.Error("Invalid tool call",
			slog.String("tool", "get_cursor_position"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	// Validate session ID
	if err := validateSessionID(sessionID); err != nil {
		slog.Error("Invalid session ID",
			slog.String("tool", "get_cursor_position"),
			slog.String("session_id", sessionID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	utils.LogToolCall("get_cursor_position", sessionID)

	sess, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	row, col := sess.GetCursorPosition()

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf(`{"row": %d, "col": %d}`, row, col),
			},
		},
	}, nil
}

func (h *Handlers) GetScreenSize(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	sessionID, ok := args["session_id"].(string)
	if !ok {
		err := fmt.Errorf("session_id parameter is required")
		slog.Error("Invalid tool call",
			slog.String("tool", "get_screen_size"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	// Validate session ID
	if err := validateSessionID(sessionID); err != nil {
		slog.Error("Invalid session ID",
			slog.String("tool", "get_screen_size"),
			slog.String("session_id", sessionID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	utils.LogToolCall("get_screen_size", sessionID)

	sess, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	width, height := sess.GetScreenSize()

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf(`{"width": %d, "height": %d}`, width, height),
			},
		},
	}, nil
}

func (h *Handlers) RestartApp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	sessionID, ok := args["session_id"].(string)
	if !ok {
		err := fmt.Errorf("session_id parameter is required")
		slog.Error("Invalid tool call",
			slog.String("tool", "restart_app"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	// Validate session ID
	if err := validateSessionID(sessionID); err != nil {
		slog.Error("Invalid session ID",
			slog.String("tool", "restart_app"),
			slog.String("session_id", sessionID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	utils.LogToolCall("restart_app", sessionID)

	sess, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if err := sess.Restart(); err != nil {
		return nil, fmt.Errorf("failed to restart app: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: `{"success": true}`,
			},
		},
	}, nil
}

func (h *Handlers) StopApp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	sessionID, ok := args["session_id"].(string)
	if !ok {
		err := fmt.Errorf("session_id parameter is required")
		slog.Error("Invalid tool call",
			slog.String("tool", "stop_app"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	// Validate session ID
	if err := validateSessionID(sessionID); err != nil {
		slog.Error("Invalid session ID",
			slog.String("tool", "stop_app"),
			slog.String("session_id", sessionID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	utils.LogToolCall("stop_app", sessionID)

	if err := h.sessionManager.RemoveSession(sessionID); err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: `{"success": true}`,
			},
		},
	}, nil
}

func (h *Handlers) ListSessions(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	utils.LogToolCall("list_sessions", "")
	
	sessions := h.sessionManager.ListSessions()
	
	slog.Debug("Sessions listed",
		slog.String("tool", "list_sessions"),
		slog.Int("count", len(sessions)),
	)

	// Convert sessions to JSON string
	var sessionStrings []string
	for _, s := range sessions {
		sessionStrings = append(sessionStrings, fmt.Sprintf(`{"id": %q, "command": %q, "state": %q, "created": %q}`, 
			s.ID, s.Command, s.State, s.Created.Format("2006-01-02T15:04:05Z")))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf(`{"sessions": [%s]}`, strings.Join(sessionStrings, ", ")),
			},
		},
	}, nil
}

func (h *Handlers) ResizeTerminal(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	
	// Debug logging
	slog.Debug("ResizeTerminal called", 
		slog.String("tool", "resize_terminal"),
		slog.Any("args", args),
	)
	
	sessionID, ok := args["session_id"].(string)
	if !ok {
		err := fmt.Errorf("session_id parameter is required")
		slog.Error("Invalid tool call",
			slog.String("tool", "resize_terminal"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	// Validate session ID
	if err := validateSessionID(sessionID); err != nil {
		slog.Error("Invalid session ID",
			slog.String("tool", "resize_terminal"),
			slog.String("session_id", sessionID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	// Try to get width as float64 or int
	var width float64
	if w, ok := args["width"].(float64); ok {
		width = w
	} else if w, ok := args["width"].(int); ok {
		width = float64(w)
	} else {
		err := fmt.Errorf("width parameter is required")
		slog.Error("Invalid tool call",
			slog.String("tool", "resize_terminal"),
			slog.String("error", err.Error()),
			slog.Any("width_type", fmt.Sprintf("%T", args["width"])),
		)
		return nil, err
	}

	// Try to get height as float64 or int
	var height float64
	if h, ok := args["height"].(float64); ok {
		height = h
	} else if h, ok := args["height"].(int); ok {
		height = float64(h)
	} else {
		err := fmt.Errorf("height parameter is required")
		slog.Error("Invalid tool call",
			slog.String("tool", "resize_terminal"),
			slog.String("error", err.Error()),
			slog.Any("height_type", fmt.Sprintf("%T", args["height"])),
		)
		return nil, err
	}
	
	// Validate dimensions
	if err := validateDimensions(width, height); err != nil {
		slog.Error("Invalid dimensions",
			slog.String("tool", "resize_terminal"),
			slog.Float64("width", width),
			slog.Float64("height", height),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	utils.LogToolCall("resize_terminal", sessionID,
		slog.Int("width", int(width)),
		slog.Int("height", int(height)),
	)

	sess, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if err := sess.Resize(int(width), int(height)); err != nil {
		utils.LogError(err, "Failed to resize terminal",
			slog.String("tool", "resize_terminal"),
			slog.String("session_id", sessionID),
		)
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf(`{"success": true, "width": %d, "height": %d}`, int(width), int(height)),
			},
		},
	}, nil
}