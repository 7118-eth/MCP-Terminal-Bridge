package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/bioharz/mcp-terminal-tester/internal/session"
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

func (h *Handlers) LaunchApp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	command, ok := args["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command parameter is required")
	}

	// Extract args if provided
	var cmdArgs []string
	if argsParam, exists := args["args"]; exists {
		if argsArray, ok := argsParam.([]interface{}); ok {
			for _, arg := range argsArray {
				if argStr, ok := arg.(string); ok {
					cmdArgs = append(cmdArgs, argStr)
				}
			}
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
	}

	// Create new session
	sess, err := h.sessionManager.CreateSession(command, cmdArgs, env)
	if err != nil {
		return nil, fmt.Errorf("failed to launch app: %w", err)
	}

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
		return nil, fmt.Errorf("session_id parameter is required")
	}

	format := "plain"
	if formatParam, exists := args["format"]; exists {
		if f, ok := formatParam.(string); ok {
			format = f
		}
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

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf(`{"content": %q, "cursor": {"row": %d, "col": %d}}`, content, row, col),
			},
		},
	}, nil
}

func (h *Handlers) SendKeys(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	sessionID, ok := args["session_id"].(string)
	if !ok {
		return nil, fmt.Errorf("session_id parameter is required")
	}

	keys, ok := args["keys"].(string)
	if !ok {
		return nil, fmt.Errorf("keys parameter is required")
	}

	sess, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// Map special keys
	mappedKeys := MapKeys(keys)

	if err := sess.SendKeys(mappedKeys); err != nil {
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
		return nil, fmt.Errorf("session_id parameter is required")
	}

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
		return nil, fmt.Errorf("session_id parameter is required")
	}

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
		return nil, fmt.Errorf("session_id parameter is required")
	}

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
		return nil, fmt.Errorf("session_id parameter is required")
	}

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
	sessions := h.sessionManager.ListSessions()

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