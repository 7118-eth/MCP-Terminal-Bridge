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
	sessionID := tf.LaunchApp("../../test/apps/echo", []string{})
	
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
	sessionID := tf.LaunchApp("../../test/apps/menu", []string{})
	
	// Wait for menu to appear
	if !tf.WaitForContent(sessionID, "Terminal Menu Test", 5*time.Second) {
		content := tf.ViewScreen(sessionID, "plain")
		t.Fatalf("Menu app didn't start properly: %s", content)
	}
	
	// Navigate down
	tf.SendKeys(sessionID, "Down")
	time.Sleep(100 * time.Millisecond)
	
	// Navigate down again
	tf.SendKeys(sessionID, "Down")
	time.Sleep(100 * time.Millisecond)
	
	// Select option 3 (Show Time)
	tf.SendKeys(sessionID, "Enter")
	
	// Should show current time
	if !tf.WaitForContent(sessionID, "Current time:", 2*time.Second) {
		content := tf.ViewScreen(sessionID, "plain")
		t.Errorf("Time display didn't work: %s", content)
	}
	
	// Press any key to continue
	tf.SendKeys(sessionID, "Enter")
	time.Sleep(100 * time.Millisecond)
	
	// Navigate to Exit
	tf.SendKeys(sessionID, "Down")
	time.Sleep(100 * time.Millisecond)
	tf.SendKeys(sessionID, "Down")
	time.Sleep(100 * time.Millisecond)
	
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
	sessionID := tf.LaunchApp("../../test/apps/progress", []string{})
	
	// Wait for progress bar to appear
	if !tf.WaitForContent(sessionID, "Progress Bar Demo", 5*time.Second) {
		content := tf.ViewScreen(sessionID, "plain")
		t.Fatalf("Progress app didn't start properly: %s", content)
	}
	
	// Wait for some progress
	time.Sleep(2 * time.Second)
	
	// Check that progress is being made
	content := tf.ViewScreen(sessionID, "plain")
	if !strings.Contains(content, "%") {
		t.Error("Progress percentage not shown")
	}
	
	// The raw format should contain ANSI sequences for the progress bar
	rawContent := tf.ViewScreen(sessionID, "raw")
	if !strings.Contains(rawContent, "\033[") {
		t.Error("Raw format should contain ANSI sequences for colors")
	}
	
	// Wait for completion
	if !tf.WaitForContent(sessionID, "All tasks completed!", 15*time.Second) {
		content := tf.ViewScreen(sessionID, "plain")
		t.Errorf("Progress didn't complete: %s", content)
	}
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
	
	// Launch a command that produces multiple lines
	sessionID := tf.LaunchApp("sh", []string{"-c", "for i in $(seq 1 30); do echo Line$i; done"})
	
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