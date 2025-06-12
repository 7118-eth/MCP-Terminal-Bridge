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
	
	// Test setting a cell
	fg := Color{R: 255, G: 0, B: 0}
	bg := Color{R: 0, G: 255, B: 0}
	attrs := Attributes{Bold: true}
	
	buffer.SetCell(5, 5, 'X', fg, bg, attrs)
	
	cell := buffer.cells[5][5]
	if cell.Rune != 'X' {
		t.Errorf("Expected rune 'X', got '%c'", cell.Rune)
	}
	if cell.Foreground != fg {
		t.Errorf("Foreground color mismatch")
	}
	if cell.Background != bg {
		t.Errorf("Background color mismatch")
	}
	if cell.Attributes != attrs {
		t.Errorf("Attributes mismatch")
	}
	
	// Test out of bounds
	buffer.SetCell(-1, 0, 'A', fg, bg, attrs) // Should not panic
	buffer.SetCell(0, -1, 'B', fg, bg, attrs) // Should not panic
	buffer.SetCell(10, 0, 'C', fg, bg, attrs) // Should not panic
	buffer.SetCell(0, 10, 'D', fg, bg, attrs) // Should not panic
}

func TestScreenBuffer_MoveCursor(t *testing.T) {
	buffer := NewScreenBuffer(80, 24)
	
	// Test normal movement
	buffer.MoveCursor(10, 5)
	if buffer.cursorX != 10 || buffer.cursorY != 5 {
		t.Errorf("Expected cursor at (10,5), got (%d,%d)", buffer.cursorX, buffer.cursorY)
	}
	
	// Test clamping
	buffer.MoveCursor(100, 30)
	if buffer.cursorX != 79 || buffer.cursorY != 23 {
		t.Errorf("Expected cursor clamped to (79,23), got (%d,%d)", buffer.cursorX, buffer.cursorY)
	}
	
	buffer.MoveCursor(-5, -5)
	if buffer.cursorX != 0 || buffer.cursorY != 0 {
		t.Errorf("Expected cursor clamped to (0,0), got (%d,%d)", buffer.cursorX, buffer.cursorY)
	}
}

func TestScreenBuffer_Clear(t *testing.T) {
	buffer := NewScreenBuffer(10, 10)
	
	// Set some cells
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			buffer.SetCell(x, y, 'A', Color{}, Color{}, Attributes{})
		}
	}
	
	// Move cursor
	buffer.MoveCursor(5, 5)
	
	// Clear
	buffer.Clear()
	
	// Check all cells are spaces
	for y := 0; y < buffer.height; y++ {
		for x := 0; x < buffer.width; x++ {
			if buffer.cells[y][x].Rune != ' ' {
				t.Errorf("Cell at (%d,%d) not cleared", x, y)
			}
		}
	}
	
	// Check cursor reset
	if buffer.cursorX != 0 || buffer.cursorY != 0 {
		t.Errorf("Cursor not reset, at (%d,%d)", buffer.cursorX, buffer.cursorY)
	}
}

func TestScreenBuffer_ScrollUp(t *testing.T) {
	buffer := NewScreenBuffer(5, 3)
	
	// Fill buffer with different lines
	for y := 0; y < 3; y++ {
		for x := 0; x < 5; x++ {
			buffer.SetCell(x, y, rune('A'+y), Color{}, Color{}, Attributes{})
		}
	}
	
	// Scroll up
	buffer.ScrollUp()
	
	// Check first line is now second line
	if buffer.cells[0][0].Rune != 'B' {
		t.Errorf("First line should have 'B' after scroll, got '%c'", buffer.cells[0][0].Rune)
	}
	
	// Check second line is now third line
	if buffer.cells[1][0].Rune != 'C' {
		t.Errorf("Second line should have 'C' after scroll, got '%c'", buffer.cells[1][0].Rune)
	}
	
	// Check last line is cleared
	for x := 0; x < 5; x++ {
		if buffer.cells[2][x].Rune != ' ' {
			t.Errorf("Last line should be cleared, got '%c' at x=%d", buffer.cells[2][x].Rune, x)
		}
	}
}

func TestScreenBuffer_Resize(t *testing.T) {
	buffer := NewScreenBuffer(10, 10)
	
	// Fill with data
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			buffer.SetCell(x, y, rune('0'+(x+y)%10), Color{}, Color{}, Attributes{})
		}
	}
	
	// Place cursor
	buffer.MoveCursor(5, 5)
	
	// Resize smaller
	buffer.Resize(5, 5)
	
	if buffer.width != 5 || buffer.height != 5 {
		t.Errorf("Expected size 5x5, got %dx%d", buffer.width, buffer.height)
	}
	
	// Check cursor is clamped
	if buffer.cursorX != 4 || buffer.cursorY != 4 {
		t.Errorf("Expected cursor clamped to (4,4), got (%d,%d)", buffer.cursorX, buffer.cursorY)
	}
	
	// Check data preservation
	if buffer.cells[0][0].Rune != '0' {
		t.Errorf("Expected preserved data '0', got '%c'", buffer.cells[0][0].Rune)
	}
	
	// Resize larger
	buffer.Resize(15, 15)
	
	// Check new cells are spaces
	if buffer.cells[10][10].Rune != ' ' {
		t.Errorf("New cells should be spaces, got '%c'", buffer.cells[10][10].Rune)
	}
}

func TestScreenBuffer_RenderPlain(t *testing.T) {
	buffer := NewScreenBuffer(10, 3)
	
	// First line
	buffer.SetCell(0, 0, 'H', Color{}, Color{}, Attributes{})
	buffer.SetCell(1, 0, 'e', Color{}, Color{}, Attributes{})
	buffer.SetCell(2, 0, 'l', Color{}, Color{}, Attributes{})
	buffer.SetCell(3, 0, 'l', Color{}, Color{}, Attributes{})
	buffer.SetCell(4, 0, 'o', Color{}, Color{}, Attributes{})
	
	// Second line
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
	buffer.SetScrollbackSize(10) // Small for testing
	
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

func TestScreenBuffer_Passthrough(t *testing.T) {
	sb := NewScreenBuffer(80, 24)
	
	// Write data with ANSI sequences
	testData := "\x1b[31mRed Text\x1b[0m Normal \x1b[1;32mBold Green\x1b[0m"
	sb.Write([]byte(testData))
	
	// Test passthrough render preserves original sequences
	passthrough, _ := sb.Render("passthrough")
	if passthrough != testData {
		t.Errorf("Passthrough render should preserve original data.\nExpected: %q\nGot: %q", testData, passthrough)
	}
	
	// Test GetRawData method
	rawData := sb.GetRawData()
	if string(rawData) != testData {
		t.Errorf("GetRawData should return original data.\nExpected: %q\nGot: %q", testData, string(rawData))
	}
	
	// Test Clear also clears raw data
	sb.Clear()
	passthrough, _ = sb.Render("passthrough")
	if passthrough != "" {
		t.Errorf("Clear should also clear raw data, but got: %q", passthrough)
	}
	
	// Test raw data size limit
	sb = NewScreenBuffer(80, 24)
	// Write data to exceed max size
	largeData := strings.Repeat("A", 512*1024) // 512KB
	sb.Write([]byte(largeData))
	sb.Write([]byte(largeData)) // Total 1MB
	sb.Write([]byte("END"))      // This should trigger trimming
	
	rawData = sb.GetRawData()
	if len(rawData) > sb.maxRawDataSize {
		t.Errorf("Raw data size %d exceeds max %d", len(rawData), sb.maxRawDataSize)
	}
	
	// Should contain the END marker after trimming
	if !strings.HasSuffix(string(rawData), "END") {
		t.Error("Raw data should preserve latest data after trimming")
	}
}