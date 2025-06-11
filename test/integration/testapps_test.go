package integration

import (
	"strings"
	"testing"
	"time"
)

func TestEchoApp(t *testing.T) {
	tf := NewTestFramework(t)
	defer tf.Cleanup()
	
	// Launch the echo test app
	sessionID := tf.LaunchApp("/Users/bioharz/git/2025/mcp-terminal-tester/test/apps/echo", []string{})
	
	// Wait for prompt
	if !tf.WaitForContent(sessionID, "Echo Test Application", 5*time.Second) {
		content := tf.ViewScreen(sessionID, "plain")
		t.Fatalf("Echo app didn't start properly: %s", content)
	}
	
	// Test echo command
	tf.SendKeys(sessionID, "Hello World")
	tf.SendKeys(sessionID, "Enter")
	
	if !tf.WaitForContent(sessionID, "You typed: Hello World", 2*time.Second) {
		content := tf.ViewScreen(sessionID, "plain")
		t.Errorf("Echo didn't work: %s", content)
	}
	
	// Test clear command
	tf.SendKeys(sessionID, "clear")
	tf.SendKeys(sessionID, "Enter")
	time.Sleep(200 * time.Millisecond)
	
	content := tf.ViewScreen(sessionID, "plain")
	// After clear, screen should be mostly empty
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) > 2 {
		t.Errorf("Clear didn't work, too many lines: %d", len(lines))
	}
	
	// Test exit
	tf.SendKeys(sessionID, "exit")
	tf.SendKeys(sessionID, "Enter")
	
	// App should terminate
	time.Sleep(500 * time.Millisecond)
	_, err := tf.CallTool("view_screen", map[string]interface{}{
		"session_id": sessionID,
		"format":     "plain",
	})
	if err == nil {
		t.Error("App should have exited")
	}
}

func TestMenuApp(t *testing.T) {
	tf := NewTestFramework(t)
	defer tf.Cleanup()
	
	// Launch the menu test app
	sessionID := tf.LaunchApp("/Users/bioharz/git/2025/mcp-terminal-tester/test/apps/menu", []string{})
	
	// Wait for menu to appear
	if !tf.WaitForContent(sessionID, "Terminal Test Menu System", 5*time.Second) {
		content := tf.ViewScreen(sessionID, "plain")
		t.Fatalf("Menu app didn't start properly: %s", content)
	}
	
	// Select first option (Show System Info) - no navigation needed, already at index 0
	tf.SendKeys(sessionID, "Enter")
	
	// Should show system info
	if !tf.WaitForContent(sessionID, "System Information:", 2*time.Second) {
		content := tf.ViewScreen(sessionID, "plain")
		t.Errorf("System info display didn't work: %s", content)
	}
	
	// Press any key to continue
	tf.SendKeys(sessionID, "Enter")
	time.Sleep(100 * time.Millisecond)
	
	// Navigate to Exit (index 6 - need to go down 6 times from 0)
	for i := 0; i < 6; i++ {
		tf.SendKeys(sessionID, "Down")
		time.Sleep(50 * time.Millisecond)
	}
	
	// Exit
	tf.SendKeys(sessionID, "Enter")
	
	// App should terminate
	time.Sleep(500 * time.Millisecond)
	_, err := tf.CallTool("view_screen", map[string]interface{}{
		"session_id": sessionID,
		"format":     "plain",
	})
	if err == nil {
		t.Error("App should have exited")
	}
}

func TestProgressApp(t *testing.T) {
	tf := NewTestFramework(t)
	defer tf.Cleanup()
	
	// Launch the progress test app
	sessionID := tf.LaunchApp("/Users/bioharz/git/2025/mcp-terminal-tester/test/apps/progress", []string{})
	
	// Wait a bit and then check what content we get
	time.Sleep(2 * time.Second)
	content := tf.ViewScreen(sessionID, "plain")
	t.Logf("Progress app content: %q", content)
	
	// Just check if we have percentage sign indicating progress is running
	if !strings.Contains(content, "%") {
		t.Errorf("Progress app doesn't seem to be showing progress: %s", content)
	}
	
	// The raw format should contain ANSI sequences for the progress bar (check while active)
	rawContent := tf.ViewScreen(sessionID, "raw")
	if !strings.Contains(rawContent, "\033[") && !strings.Contains(rawContent, "\x1b[") {
		t.Error("Raw format should contain ANSI sequences for colors")
	}
	
	// Wait for the app to complete (it will exit on its own)
	timeout := time.Now().Add(30 * time.Second)
	for time.Now().Before(timeout) {
		_, err := tf.CallTool("view_screen", map[string]interface{}{
			"session_id": sessionID,
			"format":     "plain",
		})
		if err != nil && strings.Contains(err.Error(), "session is not active") {
			// App completed successfully
			return
		}
		time.Sleep(1 * time.Second)
	}
	t.Error("Progress app didn't complete within timeout")
}

func TestAnsiFormatShowsCursor(t *testing.T) {
	tf := NewTestFramework(t)
	defer tf.Cleanup()
	
	// Launch cat for interactive input
	sessionID := tf.LaunchApp("cat", []string{})
	time.Sleep(100 * time.Millisecond)
	
	// Type some text without pressing enter
	tf.SendKeys(sessionID, "Hello")
	time.Sleep(100 * time.Millisecond)
	
	// ANSI format should show cursor marker
	ansiContent := tf.ViewScreen(sessionID, "ansi")
	if !strings.Contains(ansiContent, "▮") {
		t.Error("ANSI format should show cursor marker ▮")
	}
	
	// Send Ctrl+C to exit
	tf.SendKeys(sessionID, "Ctrl+C")
}

func TestScrollbackFormat(t *testing.T) {
	tf := NewTestFramework(t)
	defer tf.Cleanup()
	
	// Launch a command that produces multiple lines (add sleep to keep session alive)
	sessionID := tf.LaunchApp("sh", []string{"-c", "for i in $(seq 1 30); do echo Line$i; done; sleep 2"})
	
	// Wait for completion
	time.Sleep(1 * time.Second)
	
	// Regular view might not show all lines if terminal is small
	plainContent := tf.ViewScreen(sessionID, "plain")
	plainLines := strings.Split(strings.TrimSpace(plainContent), "\n")
	
	// Scrollback should include historical lines
	scrollbackContent := tf.ViewScreen(sessionID, "scrollback")
	scrollbackLines := strings.Split(strings.TrimSpace(scrollbackContent), "\n")
	
	// Scrollback should have more content if some scrolled off screen
	if len(scrollbackLines) <= len(plainLines) {
		t.Logf("Plain lines: %d, Scrollback lines: %d", len(plainLines), len(scrollbackLines))
		t.Error("Scrollback format should include historical content")
	}
	
	// Verify early lines are in scrollback
	if !strings.Contains(scrollbackContent, "Line1") {
		t.Error("Scrollback should contain Line1")
	}
}