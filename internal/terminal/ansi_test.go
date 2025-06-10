package terminal

import (
	"testing"
)

func TestANSIParser_BasicText(t *testing.T) {
	buffer := NewScreenBuffer(10, 3)
	parser := NewANSIParser(buffer)

	// Test basic text
	parser.Parse([]byte("Hello"))
	
	// Check that text was written
	for i := 0; i < 5; i++ {
		if buffer.cells[0][i].Rune != rune("Hello"[i]) {
			t.Errorf("Expected '%c' at position %d, got '%c'", "Hello"[i], i, buffer.cells[0][i].Rune)
		}
	}
	
	// Check cursor position
	if buffer.cursorX != 5 || buffer.cursorY != 0 {
		t.Errorf("Expected cursor at (5,0), got (%d,%d)", buffer.cursorX, buffer.cursorY)
	}
}

func TestANSIParser_NewlineHandling(t *testing.T) {
	buffer := NewScreenBuffer(10, 3)
	parser := NewANSIParser(buffer)

	// Test newline
	parser.Parse([]byte("Line1\nLine2"))
	
	// Check first line
	if string(getCellRunes(buffer.cells[0][:5])) != "Line1" {
		t.Error("First line incorrect")
	}
	
	// Check second line
	if string(getCellRunes(buffer.cells[1][:5])) != "Line2" {
		t.Error("Second line incorrect")
	}
	
	// Check cursor position
	if buffer.cursorX != 5 || buffer.cursorY != 1 {
		t.Errorf("Expected cursor at (5,1), got (%d,%d)", buffer.cursorX, buffer.cursorY)
	}
}

func TestANSIParser_CarriageReturn(t *testing.T) {
	buffer := NewScreenBuffer(10, 3)
	parser := NewANSIParser(buffer)

	// Write text then carriage return
	parser.Parse([]byte("Hello\rWorld"))
	
	// "World" should overwrite "Hello"
	if string(getCellRunes(buffer.cells[0][:5])) != "World" {
		t.Error("Carriage return overwrite failed")
	}
}

func TestANSIParser_CursorMovement(t *testing.T) {
	buffer := NewScreenBuffer(20, 10)
	parser := NewANSIParser(buffer)

	tests := []struct {
		name     string
		sequence string
		expectedX int
		expectedY int
	}{
		{"Move up", "\x1b[A", 0, 0}, // Can't go above 0
		{"Move down", "\x1b[5B", 0, 5},
		{"Move right", "\x1b[10C", 10, 5},
		{"Move left", "\x1b[3D", 7, 5},
		{"Move to position", "\x1b[3;8H", 7, 2}, // 1-based to 0-based
		{"Move to column", "\x1b[15G", 14, 2},
		{"Move to row", "\x1b[7d", 14, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser.Parse([]byte(tt.sequence))
			if buffer.cursorX != tt.expectedX || buffer.cursorY != tt.expectedY {
				t.Errorf("Expected cursor at (%d,%d), got (%d,%d)", 
					tt.expectedX, tt.expectedY, buffer.cursorX, buffer.cursorY)
			}
		})
	}
}

func TestANSIParser_ClearOperations(t *testing.T) {
	buffer := NewScreenBuffer(10, 3)
	parser := NewANSIParser(buffer)

	// Fill buffer with 'X'
	for y := 0; y < 3; y++ {
		for x := 0; x < 10; x++ {
			buffer.SetCell(x, y, 'X', Color{Default: true}, Color{Default: true}, Attributes{})
		}
	}

	// Test clear to end of line
	buffer.MoveCursor(5, 1)
	parser.Parse([]byte("\x1b[K"))
	
	// Check that positions 5-9 on line 1 are cleared
	for x := 5; x < 10; x++ {
		if buffer.cells[1][x].Rune != ' ' {
			t.Errorf("Position (%d,1) should be cleared", x)
		}
	}
	
	// Check that other positions are unchanged
	if buffer.cells[1][4].Rune != 'X' {
		t.Error("Position (4,1) should not be cleared")
	}
}

func TestANSIParser_ColorSGR(t *testing.T) {
	buffer := NewScreenBuffer(10, 3)
	parser := NewANSIParser(buffer)

	// Test foreground color
	parser.Parse([]byte("\x1b[31mRed"))
	
	// Check that text has red foreground
	for i := 0; i < 3; i++ {
		cell := buffer.cells[0][i]
		if cell.Foreground.R != 170 || cell.Foreground.G != 0 || cell.Foreground.B != 0 {
			t.Errorf("Expected red color, got R:%d G:%d B:%d", 
				cell.Foreground.R, cell.Foreground.G, cell.Foreground.B)
		}
	}

	// Test reset
	parser.Parse([]byte("\x1b[0m Normal"))
	
	// Check that color is reset
	cell := buffer.cells[0][4] // Space after "Red"
	if !cell.Foreground.Default {
		t.Error("Color should be reset to default")
	}
}

func TestANSIParser_Attributes(t *testing.T) {
	buffer := NewScreenBuffer(20, 3)
	parser := NewANSIParser(buffer)

	tests := []struct {
		sequence string
		checkAttr func(Attributes) bool
		name string
	}{
		{"\x1b[1m", func(a Attributes) bool { return a.Bold }, "Bold"},
		{"\x1b[3m", func(a Attributes) bool { return a.Italic }, "Italic"},
		{"\x1b[4m", func(a Attributes) bool { return a.Underline }, "Underline"},
		{"\x1b[5m", func(a Attributes) bool { return a.Blink }, "Blink"},
		{"\x1b[7m", func(a Attributes) bool { return a.Reverse }, "Reverse"},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser.Parse([]byte(tt.sequence + "X"))
			cell := buffer.cells[0][i]
			if !tt.checkAttr(cell.Attributes) {
				t.Errorf("%s attribute not set", tt.name)
			}
		})
	}
}

func TestANSIParser_256Colors(t *testing.T) {
	buffer := NewScreenBuffer(10, 3)
	parser := NewANSIParser(buffer)

	// Test 256 color foreground
	parser.Parse([]byte("\x1b[38;5;196mX")) // Color 196 is bright red
	
	cell := buffer.cells[0][0]
	if cell.Rune != 'X' {
		t.Error("Character not written")
	}
	
	// For 256 color mode, just check it's not default
	if cell.Foreground.Default {
		t.Error("Foreground color should not be default")
	}
}

func TestANSIParser_ScrollUp(t *testing.T) {
	buffer := NewScreenBuffer(5, 3)
	parser := NewANSIParser(buffer)

	// Fill three lines
	parser.Parse([]byte("Line1\nLine2\nLine3\n"))
	
	// This should cause scroll
	parser.Parse([]byte("Line4"))
	
	// Check that Line1 is gone, Line2 is at top
	if string(getCellRunes(buffer.cells[0][:5])) != "Line2" {
		t.Error("First line should be Line2 after scroll")
	}
	
	if string(getCellRunes(buffer.cells[2][:5])) != "Line4" {
		t.Error("Last line should be Line4")
	}
}

func TestANSIParser_SaveRestoreCursor(t *testing.T) {
	buffer := NewScreenBuffer(10, 10)
	parser := NewANSIParser(buffer)

	// Move cursor and save
	buffer.MoveCursor(5, 3)
	parser.Parse([]byte("\x1b[s")) // Save cursor
	
	// Move cursor elsewhere
	buffer.MoveCursor(8, 7)
	
	// Restore cursor
	parser.Parse([]byte("\x1b[8")) // Restore cursor
	
	if buffer.cursorX != 5 || buffer.cursorY != 3 {
		t.Errorf("Cursor not restored correctly, expected (5,3), got (%d,%d)", 
			buffer.cursorX, buffer.cursorY)
	}
}

// Helper function to get runes from cells
func getCellRunes(cells []Cell) []rune {
	runes := make([]rune, len(cells))
	for i, cell := range cells {
		runes[i] = cell.Rune
	}
	return runes
}