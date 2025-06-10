package terminal

import (
	"bytes"
	"strings"
	"sync"
)

type Cell struct {
	Rune       rune
	Foreground Color
	Background Color 
	Attributes Attributes
}

type Color struct {
	R, G, B uint8
	Default bool
}

type Attributes struct {
	Bold      bool
	Italic    bool
	Underline bool
	Blink     bool
	Reverse   bool
	Hidden    bool
}

type ScreenBuffer struct {
	cells    [][]Cell
	width    int
	height   int
	cursorX  int
	cursorY  int
	parser   *ANSIParser
	mu       sync.RWMutex
}

func NewScreenBuffer(width, height int) *ScreenBuffer {
	cells := make([][]Cell, height)
	for i := range cells {
		cells[i] = make([]Cell, width)
		for j := range cells[i] {
			cells[i][j] = Cell{
				Rune:       ' ',
				Foreground: Color{Default: true},
				Background: Color{Default: true},
			}
		}
	}

	sb := &ScreenBuffer{
		cells:   cells,
		width:   width,
		height:  height,
		cursorX: 0,
		cursorY: 0,
	}

	sb.parser = NewANSIParser(sb)
	return sb
}

func (sb *ScreenBuffer) Write(data []byte) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	// Parse ANSI sequences and update buffer
	sb.parser.Parse(data)
}

func (sb *ScreenBuffer) SetCell(x, y int, r rune, fg, bg Color, attrs Attributes) {
	if x < 0 || x >= sb.width || y < 0 || y >= sb.height {
		return
	}

	sb.cells[y][x] = Cell{
		Rune:       r,
		Foreground: fg,
		Background: bg,
		Attributes: attrs,
	}
}

func (sb *ScreenBuffer) MoveCursor(x, y int) {
	sb.cursorX = x
	sb.cursorY = y

	// Clamp to bounds
	if sb.cursorX < 0 {
		sb.cursorX = 0
	}
	if sb.cursorX >= sb.width {
		sb.cursorX = sb.width - 1
	}
	if sb.cursorY < 0 {
		sb.cursorY = 0
	}
	if sb.cursorY >= sb.height {
		sb.cursorY = sb.height - 1
	}
}

func (sb *ScreenBuffer) Clear() {
	for y := 0; y < sb.height; y++ {
		for x := 0; x < sb.width; x++ {
			sb.cells[y][x] = Cell{
				Rune:       ' ',
				Foreground: Color{Default: true},
				Background: Color{Default: true},
			}
		}
	}
	sb.cursorX = 0
	sb.cursorY = 0
}

func (sb *ScreenBuffer) ClearLine(y int) {
	if y < 0 || y >= sb.height {
		return
	}

	for x := 0; x < sb.width; x++ {
		sb.cells[y][x] = Cell{
			Rune:       ' ',
			Foreground: Color{Default: true},
			Background: Color{Default: true},
		}
	}
}

func (sb *ScreenBuffer) ScrollUp() {
	// Move all lines up by one
	for y := 0; y < sb.height-1; y++ {
		sb.cells[y] = sb.cells[y+1]
	}

	// Clear the bottom line
	sb.cells[sb.height-1] = make([]Cell, sb.width)
	for x := 0; x < sb.width; x++ {
		sb.cells[sb.height-1][x] = Cell{
			Rune:       ' ',
			Foreground: Color{Default: true},
			Background: Color{Default: true},
		}
	}
}

func (sb *ScreenBuffer) Render(format string) (string, error) {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	switch format {
	case "plain":
		return sb.renderPlain(), nil
	case "raw":
		return sb.renderRaw(), nil
	case "ansi":
		return sb.renderANSI(), nil
	default:
		return sb.renderPlain(), nil
	}
}

func (sb *ScreenBuffer) renderPlain() string {
	var buf bytes.Buffer

	for y := 0; y < sb.height; y++ {
		for x := 0; x < sb.width; x++ {
			buf.WriteRune(sb.cells[y][x].Rune)
		}
		// Don't add newline after the last line
		if y < sb.height-1 {
			buf.WriteRune('\n')
		}
	}

	return strings.TrimRight(buf.String(), " \n")
}

func (sb *ScreenBuffer) renderRaw() string {
	// For now, raw format is the same as plain
	// In the future, this could include ANSI sequences
	return sb.renderPlain()
}

func (sb *ScreenBuffer) renderANSI() string {
	var buf bytes.Buffer

	for y := 0; y < sb.height; y++ {
		for x := 0; x < sb.width; x++ {
			cell := sb.cells[y][x]
			
			// Show cursor position with a marker
			if x == sb.cursorX && y == sb.cursorY {
				buf.WriteString("▮")
			} else if cell.Rune == ' ' {
				buf.WriteString("·")
			} else {
				buf.WriteRune(cell.Rune)
			}
		}
		if y < sb.height-1 {
			buf.WriteRune('\n')
		}
	}

	return buf.String()
}

func (sb *ScreenBuffer) GetCursorPosition() (int, int) {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return sb.cursorX, sb.cursorY
}

func (sb *ScreenBuffer) GetSize() (int, int) {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return sb.width, sb.height
}

func (sb *ScreenBuffer) Resize(width, height int) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	// Create new cells
	newCells := make([][]Cell, height)
	for i := range newCells {
		newCells[i] = make([]Cell, width)
		for j := range newCells[i] {
			newCells[i][j] = Cell{
				Rune:       ' ',
				Foreground: Color{Default: true},
				Background: Color{Default: true},
			}
		}
	}

	// Copy existing content
	minHeight := height
	if sb.height < minHeight {
		minHeight = sb.height
	}
	minWidth := width
	if sb.width < minWidth {
		minWidth = sb.width
	}

	for y := 0; y < minHeight; y++ {
		for x := 0; x < minWidth; x++ {
			newCells[y][x] = sb.cells[y][x]
		}
	}

	sb.cells = newCells
	sb.width = width
	sb.height = height

	// Adjust cursor position if needed
	if sb.cursorX >= width {
		sb.cursorX = width - 1
	}
	if sb.cursorY >= height {
		sb.cursorY = height - 1
	}
}