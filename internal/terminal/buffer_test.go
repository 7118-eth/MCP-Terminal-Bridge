package terminal

import (
	"strings"
	"testing"
)

func TestScreenBuffer_Creation(t *testing.T) {
	buffer := NewScreenBuffer(80, 24)
	
	if buffer.width != 80 || buffer.height != 24 {
		t.Errorf("Expected size 80x24, got %dx%d", buffer.width, buffer.height)
	}
	
	if buffer.cursorX != 0 || buffer.cursorY != 0 {
		t.Errorf("Expected cursor at (0,0), got (%d,%d)", buffer.cursorX, buffer.cursorY)
	}
	
	// Check all cells are initialized with spaces
	for y := 0; y < buffer.height; y++ {
		for x := 0; x < buffer.width; x++ {
			if buffer.cells[y][x].Rune != ' ' {
				t.Errorf("Cell at (%d,%d) not initialized with space", x, y)
			}
		}
	}
}

func TestScreenBuffer_SetCell(t *testing.T) {
	buffer := NewScreenBuffer(10, 10)
	
	fg := Color{R: 255, G: 0, B: 0}
	bg := Color{R: 0, G: 0, B: 255}
	attrs := Attributes{Bold: true, Underline: true}
	
	buffer.SetCell(5, 3, 'X', fg, bg, attrs)
	
	cell := buffer.cells[3][5]
	if cell.Rune != 'X' {
		t.Errorf("Expected 'X', got '%c'", cell.Rune)
	}
	
	if cell.Foreground != fg {
		t.Error("Foreground color mismatch")
	}
	
	if cell.Background != bg {
		t.Error("Background color mismatch")
	}
	
	if cell.Attributes != attrs {
		t.Error("Attributes mismatch")
	}
}

func TestScreenBuffer_MoveCursor(t *testing.T) {
	buffer := NewScreenBuffer(10, 10)
	
	// Normal move
	buffer.MoveCursor(5, 7)
	if buffer.cursorX != 5 || buffer.cursorY != 7 {
		t.Errorf("Expected cursor at (5,7), got (%d,%d)", buffer.cursorX, buffer.cursorY)
	}
	
	// Test clamping
	buffer.MoveCursor(-5, 20)
	if buffer.cursorX != 0 || buffer.cursorY != 9 {
		t.Errorf("Expected cursor clamped to (0,9), got (%d,%d)", buffer.cursorX, buffer.cursorY)
	}
	
	buffer.MoveCursor(15, -3)
	if buffer.cursorX != 9 || buffer.cursorY != 0 {
		t.Errorf("Expected cursor clamped to (9,0), got (%d,%d)", buffer.cursorX, buffer.cursorY)
	}
}

func TestScreenBuffer_Clear(t *testing.T) {
	buffer := NewScreenBuffer(5, 5)
	
	// Fill with 'X'
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			buffer.SetCell(x, y, 'X', Color{}, Color{}, Attributes{})
		}
	}
	buffer.MoveCursor(3, 4)
	
	// Clear
	buffer.Clear()
	
	// Check all cells are spaces
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			if buffer.cells[y][x].Rune != ' ' {
				t.Errorf("Cell at (%d,%d) not cleared", x, y)
			}
		}
	}
	
	// Check cursor reset
	if buffer.cursorX != 0 || buffer.cursorY != 0 {
		t.Error("Cursor not reset to origin")
	}
}

func TestScreenBuffer_ScrollUp(t *testing.T) {
	buffer := NewScreenBuffer(5, 3)
	
	// Fill lines
	for x := 0; x < 5; x++ {
		buffer.SetCell(x, 0, '1', Color{}, Color{}, Attributes{})
		buffer.SetCell(x, 1, '2', Color{}, Color{}, Attributes{})
		buffer.SetCell(x, 2, '3', Color{}, Color{}, Attributes{})
	}
	
	// Scroll up
	buffer.ScrollUp()
	
	// Check that line 1 became line 0, etc.
	for x := 0; x < 5; x++ {
		if buffer.cells[0][x].Rune != '2' {
			t.Errorf("Line 0 should have '2', got '%c'", buffer.cells[0][x].Rune)
		}
		if buffer.cells[1][x].Rune != '3' {
			t.Errorf("Line 1 should have '3', got '%c'", buffer.cells[1][x].Rune)
		}
		if buffer.cells[2][x].Rune != ' ' {
			t.Errorf("Line 2 should be empty, got '%c'", buffer.cells[2][x].Rune)
		}
	}
}

func TestScreenBuffer_Resize(t *testing.T) {
	buffer := NewScreenBuffer(5, 5)
	
	// Fill with pattern
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			buffer.SetCell(x, y, rune('0'+y), Color{}, Color{}, Attributes{})
		}
	}
	buffer.MoveCursor(4, 4)
	
	// Resize smaller
	buffer.Resize(3, 3)
	
	if buffer.width != 3 || buffer.height != 3 {
		t.Errorf("Expected size 3x3, got %dx%d", buffer.width, buffer.height)
	}
	
	// Check content preserved
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			if buffer.cells[y][x].Rune != rune('0'+y) {
				t.Errorf("Content at (%d,%d) not preserved", x, y)
			}
		}
	}
	
	// Check cursor clamped
	if buffer.cursorX != 2 || buffer.cursorY != 2 {
		t.Errorf("Expected cursor clamped to (2,2), got (%d,%d)", buffer.cursorX, buffer.cursorY)
	}
	
	// Resize larger
	buffer.Resize(7, 7)
	
	// Check new cells are spaces
	for y := 0; y < 7; y++ {
		for x := 0; x < 7; x++ {
			if x >= 3 || y >= 3 {
				if buffer.cells[y][x].Rune != ' ' {
					t.Errorf("New cell at (%d,%d) not initialized with space", x, y)
				}
			}
		}
	}
}

func TestScreenBuffer_RenderPlain(t *testing.T) {
	buffer := NewScreenBuffer(5, 3)
	
	// Create pattern
	buffer.SetCell(0, 0, 'H', Color{}, Color{}, Attributes{})
	buffer.SetCell(1, 0, 'e', Color{}, Color{}, Attributes{})
	buffer.SetCell(2, 0, 'l', Color{}, Color{}, Attributes{})
	buffer.SetCell(3, 0, 'l', Color{}, Color{}, Attributes{})
	buffer.SetCell(4, 0, 'o', Color{}, Color{}, Attributes{})
	
	buffer.SetCell(0, 1, 'W', Color{}, Color{}, Attributes{})
	buffer.SetCell(1, 1, 'o', Color{}, Color{}, Attributes{})
	buffer.SetCell(2, 1, 'r', Color{}, Color{}, Attributes{})
	buffer.SetCell(3, 1, 'l', Color{}, Color{}, Attributes{})
	buffer.SetCell(4, 1, 'd', Color{}, Color{}, Attributes{})
	
	rendered, err := buffer.Render("plain")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	
	lines := strings.Split(rendered, "\n")
	if len(lines) < 2 {
		t.Fatal("Expected at least 2 lines")
	}
	
	if !strings.HasPrefix(lines[0], "Hello") {
		t.Errorf("First line should start with 'Hello', got '%s'", lines[0])
	}
	
	if !strings.HasPrefix(lines[1], "World") {
		t.Errorf("Second line should start with 'World', got '%s'", lines[1])
	}
}

func TestScreenBuffer_Scrollback(t *testing.T) {
	buffer := NewScreenBuffer(5, 3)
	buffer.maxScrollback = 10 // Small for testing
	
	// Add some lines that will go to scrollback
	for i := 0; i < 5; i++ {
		for x := 0; x < 5; x++ {
			buffer.SetCell(x, 0, rune('A'+i), Color{}, Color{}, Attributes{})
		}
		buffer.ScrollUp()
	}
	
	// Get scrollback
	scrollback := buffer.GetScrollback()
	
	if len(scrollback) != 5 {
		t.Errorf("Expected 5 lines in scrollback, got %d", len(scrollback))
	}
	
	// Check content
	for i, line := range scrollback {
		if line[0].Rune != rune('A'+i) {
			t.Errorf("Scrollback line %d should start with '%c', got '%c'", 
				i, 'A'+i, line[0].Rune)
		}
	}
}