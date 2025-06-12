package terminal

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
)

// Buffer pool for render operations to reduce allocations
var renderBufferPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

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
	
	// Raw data preservation
	rawData         []byte       // Store raw input data with ANSI sequences
	rawDataMu       sync.RWMutex // Separate mutex for raw data
	maxRawDataSize  int          // Maximum size for raw data buffer
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
		cells:          cells,
		width:          width,
		height:         height,
		cursorX:        0,
		cursorY:        0,
		maxScrollback:  1000, // Default scrollback size
		maxRawDataSize: 1024 * 1024, // 1MB max raw data buffer
		rawData:        make([]byte, 0, 4096), // Start with 4KB capacity
	}

	// Initialize scrollback buffer
	sb.scrollback = make([][]Cell, sb.maxScrollback)

	sb.parser = NewANSIParser(sb)
	return sb
}

// Close releases resources associated with the screen buffer
func (sb *ScreenBuffer) Close() {
	if sb.parser != nil {
		sb.parser.Release()
		sb.parser = nil
	}
}

// SetScrollbackSize sets the maximum scrollback buffer size
func (sb *ScreenBuffer) SetScrollbackSize(size int) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	
	if size < 0 {
		size = 0
	}
	
	// Create new scrollback buffer
	newScrollback := make([][]Cell, size)
	
	// Copy existing scrollback if any
	if sb.scrollbackStart > 0 && size > 0 {
		// Calculate how many lines to copy
		linesToCopy := sb.scrollbackStart
		if linesToCopy > size {
			linesToCopy = size
		}
		if linesToCopy > sb.maxScrollback {
			linesToCopy = sb.maxScrollback
		}
		
		// Copy from old to new buffer
		for i := 0; i < linesToCopy; i++ {
			srcIndex := (sb.scrollbackStart - linesToCopy + i) % sb.maxScrollback
			if srcIndex < 0 {
				srcIndex += sb.maxScrollback
			}
			newScrollback[i] = sb.scrollback[srcIndex]
		}
		
		// Update start index
		if sb.scrollbackStart > size {
			sb.scrollbackStart = size
		}
	}
	
	sb.scrollback = newScrollback
	sb.maxScrollback = size
}

func (sb *ScreenBuffer) Write(data []byte) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	// Store raw data for true passthrough
	sb.storeRawData(data)
	
	// Parse ANSI sequences and update buffer
	sb.parser.Parse(data)
}

// storeRawData appends raw data to the buffer with size management
func (sb *ScreenBuffer) storeRawData(data []byte) {
	sb.rawDataMu.Lock()
	defer sb.rawDataMu.Unlock()
	
	// Append new data
	sb.rawData = append(sb.rawData, data...)
	
	// Trim if exceeds max size (keep last 75% when trimming)
	if len(sb.rawData) > sb.maxRawDataSize {
		trimPoint := sb.maxRawDataSize / 4
		sb.rawData = sb.rawData[trimPoint:]
	}
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
	
	// Also clear raw data on full clear
	sb.ClearRawData()
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
	case "scrollback":
		return sb.renderWithScrollback(), nil
	case "passthrough":
		return sb.renderPassthrough(), nil
	default:
		return sb.renderPlain(), nil
	}
}

func (sb *ScreenBuffer) renderPlain() string {
	buf := renderBufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		renderBufferPool.Put(buf)
	}()

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
	buf := renderBufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		renderBufferPool.Put(buf)
	}()
	
	// Track current state to minimize escape sequences
	currentFG := Color{Default: true}
	currentBG := Color{Default: true}
	currentAttrs := Attributes{}
	
	// Start with reset
	buf.WriteString("\x1b[0m")
	
	for y := 0; y < sb.height; y++ {
		for x := 0; x < sb.width; x++ {
			cell := sb.cells[y][x]
			
			// Only emit SGR if attributes changed
			if cell.Foreground != currentFG || cell.Background != currentBG || cell.Attributes != currentAttrs {
				sgr := sb.buildSGRSequence(cell.Foreground, cell.Background, cell.Attributes)
				if sgr != "" {
					buf.WriteString(sgr)
				}
				currentFG = cell.Foreground
				currentBG = cell.Background
				currentAttrs = cell.Attributes
			}
			
			buf.WriteRune(cell.Rune)
		}
		
		if y < sb.height-1 {
			buf.WriteRune('\n')
		}
	}
	
	// Position cursor at the end
	buf.WriteString(fmt.Sprintf("\x1b[%d;%dH", sb.cursorY+1, sb.cursorX+1))
	
	return buf.String()
}

func (sb *ScreenBuffer) renderANSI() string {
	buf := renderBufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		renderBufferPool.Put(buf)
	}()

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
	buf := renderBufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		renderBufferPool.Put(buf)
	}()

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

// renderPassthrough returns the raw data exactly as received, preserving all ANSI sequences
func (sb *ScreenBuffer) renderPassthrough() string {
	sb.rawDataMu.RLock()
	defer sb.rawDataMu.RUnlock()
	
	// Return a copy of the raw data as string
	return string(sb.rawData)
}

// GetRawData returns a copy of the raw data buffer
func (sb *ScreenBuffer) GetRawData() []byte {
	sb.rawDataMu.RLock()
	defer sb.rawDataMu.RUnlock()
	
	// Return a copy to prevent external modifications
	result := make([]byte, len(sb.rawData))
	copy(result, sb.rawData)
	return result
}

// ClearRawData clears the raw data buffer
func (sb *ScreenBuffer) ClearRawData() {
	sb.rawDataMu.Lock()
	defer sb.rawDataMu.Unlock()
	
	sb.rawData = sb.rawData[:0] // Keep capacity
}

// buildSGRSequence builds an ANSI SGR sequence for the given attributes
func (sb *ScreenBuffer) buildSGRSequence(fg, bg Color, attrs Attributes) string {
	// Reset if all defaults
	if fg.Default && bg.Default && attrs == (Attributes{}) {
		return "\x1b[0m"
	}

	var builder strings.Builder
	builder.WriteString("\x1b[")
	hasParam := false

	// Helper to add separator if needed
	addParam := func(param string) {
		if hasParam {
			builder.WriteByte(';')
		}
		builder.WriteString(param)
		hasParam = true
	}

	// Attributes
	if attrs.Bold {
		addParam("1")
	}
	if attrs.Italic {
		addParam("3")
	}
	if attrs.Underline {
		addParam("4")
	}
	if attrs.Blink {
		addParam("5")
	}
	if attrs.Reverse {
		addParam("7")
	}
	if attrs.Hidden {
		addParam("8")
	}

	// Foreground color
	if !fg.Default {
		addParam(fmt.Sprintf("38;2;%d;%d;%d", fg.R, fg.G, fg.B))
	}

	// Background color
	if !bg.Default {
		addParam(fmt.Sprintf("48;2;%d;%d;%d", bg.R, bg.G, bg.B))
	}

	if !hasParam {
		return ""
	}

	builder.WriteByte('m')
	return builder.String()
}