package terminal

import (
	"bytes"
	"fmt"
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
	cells           [][]Cell
	width           int
	height          int
	cursorX         int
	cursorY         int
	parser          *ANSIParser
	scrollback      [][]Cell
	maxScrollback   int
	scrollbackStart int // Index of first line in circular buffer
	mu              sync.RWMutex
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
	// Save the top line to scrollback
	sb.addToScrollback(sb.cells[0])

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

// ScrollDown scrolls the buffer content down by one line
func (sb *ScreenBuffer) ScrollDown() {
	// Move all lines down by one
	for y := sb.height - 1; y > 0; y-- {
		sb.cells[y] = sb.cells[y-1]
	}

	// Clear the top line
	sb.cells[0] = make([]Cell, sb.width)
	for x := 0; x < sb.width; x++ {
		sb.cells[0][x] = Cell{
			Rune:       ' ',
			Foreground: Color{Default: true},
			Background: Color{Default: true},
		}
	}
}

// InsertLines inserts n blank lines at position y
func (sb *ScreenBuffer) InsertLines(y, n int) {
	if y < 0 || y >= sb.height || n <= 0 {
		return
	}

	// Limit n to available space
	if y + n > sb.height {
		n = sb.height - y
	}

	// Shift lines down
	for i := sb.height - 1; i >= y + n; i-- {
		sb.cells[i] = sb.cells[i-n]
	}

	// Clear inserted lines
	for i := y; i < y + n && i < sb.height; i++ {
		sb.ClearLine(i)
	}
}

// DeleteLines deletes n lines starting at position y
func (sb *ScreenBuffer) DeleteLines(y, n int) {
	if y < 0 || y >= sb.height || n <= 0 {
		return
	}

	// Limit n to available lines
	if y + n > sb.height {
		n = sb.height - y
	}

	// Shift lines up
	for i := y; i < sb.height - n; i++ {
		sb.cells[i] = sb.cells[i+n]
	}

	// Clear bottom lines
	for i := sb.height - n; i < sb.height; i++ {
		sb.ClearLine(i)
	}
}

// InsertChars inserts n blank characters at position (x, y)
func (sb *ScreenBuffer) InsertChars(x, y, n int) {
	if x < 0 || x >= sb.width || y < 0 || y >= sb.height || n <= 0 {
		return
	}

	// Limit n to available space
	if x + n > sb.width {
		n = sb.width - x
	}

	// Shift characters right
	for i := sb.width - 1; i >= x + n; i-- {
		sb.cells[y][i] = sb.cells[y][i-n]
	}

	// Clear inserted characters
	for i := x; i < x + n && i < sb.width; i++ {
		sb.cells[y][i] = Cell{
			Rune:       ' ',
			Foreground: Color{Default: true},
			Background: Color{Default: true},
		}
	}
}

// DeleteChars deletes n characters at position (x, y)
func (sb *ScreenBuffer) DeleteChars(x, y, n int) {
	if x < 0 || x >= sb.width || y < 0 || y >= sb.height || n <= 0 {
		return
	}

	// Limit n to available characters
	if x + n > sb.width {
		n = sb.width - x
	}

	// Shift characters left
	for i := x; i < sb.width - n; i++ {
		sb.cells[y][i] = sb.cells[y][i+n]
	}

	// Clear end of line
	for i := sb.width - n; i < sb.width; i++ {
		sb.cells[y][i] = Cell{
			Rune:       ' ',
			Foreground: Color{Default: true},
			Background: Color{Default: true},
		}
	}
}

// addToScrollback adds a line to the scrollback buffer
func (sb *ScreenBuffer) addToScrollback(line []Cell) {
	if sb.maxScrollback == 0 {
		return
	}

	// Copy the line
	lineCopy := make([]Cell, len(line))
	copy(lineCopy, line)

	// Add to circular buffer
	index := sb.scrollbackStart % sb.maxScrollback
	sb.scrollback[index] = lineCopy
	sb.scrollbackStart++
}

// GetScrollback returns the scrollback buffer contents
func (sb *ScreenBuffer) GetScrollback() [][]Cell {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	if sb.scrollbackStart == 0 {
		return nil
	}

	// Calculate how many lines we have
	lineCount := sb.scrollbackStart
	if lineCount > sb.maxScrollback {
		lineCount = sb.maxScrollback
	}

	// Extract lines from circular buffer
	result := make([][]Cell, lineCount)
	for i := 0; i < lineCount; i++ {
		// Calculate actual index in circular buffer
		index := (sb.scrollbackStart - lineCount + i) % sb.maxScrollback
		if index < 0 {
			index += sb.maxScrollback
		}
		result[i] = sb.scrollback[index]
	}

	return result
}

// renderWithScrollback renders the buffer including scrollback history
func (sb *ScreenBuffer) renderWithScrollback() string {
	var buf bytes.Buffer

	// First render scrollback
	scrollbackLines := sb.GetScrollback()
	for _, line := range scrollbackLines {
		for _, cell := range line {
			buf.WriteRune(cell.Rune)
		}
		buf.WriteRune('\n')
	}

	// Then render current screen
	buf.WriteString(sb.renderPlain())

	return buf.String()
}

// buildSGRSequence builds an ANSI SGR sequence for the given attributes
func (sb *ScreenBuffer) buildSGRSequence(fg, bg Color, attrs Attributes) string {
	var params []string

	// Reset if all defaults
	if fg.Default && bg.Default && attrs == (Attributes{}) {
		return "\x1b[0m"
	}

	// Attributes
	if attrs.Bold {
		params = append(params, "1")
	}
	if attrs.Italic {
		params = append(params, "3")
	}
	if attrs.Underline {
		params = append(params, "4")
	}
	if attrs.Blink {
		params = append(params, "5")
	}
	if attrs.Reverse {
		params = append(params, "7")
	}
	if attrs.Hidden {
		params = append(params, "8")
	}

	// Foreground color
	if !fg.Default {
		params = append(params, fmt.Sprintf("38;2;%d;%d;%d", fg.R, fg.G, fg.B))
	}

	// Background color
	if !bg.Default {
		params = append(params, fmt.Sprintf("48;2;%d;%d;%d", bg.R, bg.G, bg.B))
	}

	if len(params) == 0 {
		return ""
	}

	return fmt.Sprintf("\x1b[%sm", strings.Join(params, ";"))
}