package tools

import (
	"strings"
)

var specialKeys = map[string]string{
	"Enter":     "\r",
	"Tab":       "\t", 
	"Backspace": "\x7f",
	"Escape":    "\x1b",
	"Space":     " ",
	"Delete":    "\x1b[3~",
	
	// Arrow keys
	"Up":    "\x1b[A",
	"Down":  "\x1b[B",
	"Right": "\x1b[C",
	"Left":  "\x1b[D",
	
	// Control keys
	"Ctrl+A": "\x01",
	"Ctrl+B": "\x02",
	"Ctrl+C": "\x03",
	"Ctrl+D": "\x04",
	"Ctrl+E": "\x05",
	"Ctrl+F": "\x06",
	"Ctrl+G": "\x07",
	"Ctrl+H": "\x08",
	"Ctrl+I": "\x09",
	"Ctrl+J": "\x0a",
	"Ctrl+K": "\x0b",
	"Ctrl+L": "\x0c",
	"Ctrl+M": "\x0d",
	"Ctrl+N": "\x0e",
	"Ctrl+O": "\x0f",
	"Ctrl+P": "\x10",
	"Ctrl+Q": "\x11",
	"Ctrl+R": "\x12",
	"Ctrl+S": "\x13",
	"Ctrl+T": "\x14",
	"Ctrl+U": "\x15",
	"Ctrl+V": "\x16",
	"Ctrl+W": "\x17",
	"Ctrl+X": "\x18",
	"Ctrl+Y": "\x19",
	"Ctrl+Z": "\x1a",
	
	// Function keys
	"F1":  "\x1bOP",
	"F2":  "\x1bOQ",
	"F3":  "\x1bOR",
	"F4":  "\x1bOS",
	"F5":  "\x1b[15~",
	"F6":  "\x1b[17~",
	"F7":  "\x1b[18~",
	"F8":  "\x1b[19~",
	"F9":  "\x1b[20~",
	"F10": "\x1b[21~",
	"F11": "\x1b[23~",
	"F12": "\x1b[24~",
	
	// Navigation keys
	"Home":     "\x1b[H",
	"End":      "\x1b[F",
	"PageUp":   "\x1b[5~",
	"PageDown": "\x1b[6~",
	"Insert":   "\x1b[2~",
}

// MapKeys converts special key names to their terminal sequences
func MapKeys(input string) string {
	// Check if the entire input is a special key
	if seq, ok := specialKeys[input]; ok {
		return seq
	}
	
	// Check for lowercase versions
	if seq, ok := specialKeys[strings.Title(strings.ToLower(input))]; ok {
		return seq
	}
	
	// Return the input as-is if it's not a special key
	return input
}