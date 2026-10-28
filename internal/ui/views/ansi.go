package views

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// ansiControlRegex matches screen-clearing, cursor-moving, and terminal mode escape codes (excluding SGR color 'm').
var ansiControlRegex = regexp.MustCompile(`\x1b\[(?:\?|\>)?\d*(?:;\d+)*[A-HJ-KSTsuhl]`)

// StripANSI removes all ANSI escape sequences from the string.
func StripANSI(s string) string {
	return ansi.Strip(s)
}

// SanitizeLogLine cleans a single line of raw log output by normalizing tabs,
// removing carriage returns and dangerous terminal control codes, while preserving colors.
func SanitizeLogLine(line string) string {
	// Remove carriage returns
	clean := strings.ReplaceAll(line, "\r", "")
	// Expand tabs to 4 spaces
	clean = strings.ReplaceAll(clean, "\t", "    ")
	// Strip null bytes and bells
	clean = strings.ReplaceAll(clean, "\x00", "")
	clean = strings.ReplaceAll(clean, "\a", "")
	// Remove screen/cursor control codes
	clean = ansiControlRegex.ReplaceAllString(clean, "")

	return clean
}

// SanitizeLogs splits raw log text into sanitized, display-ready lines.
func SanitizeLogs(raw string) []string {
	if raw == "" {
		return nil
	}

	// Normalize windows newlines
	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	rawLines := strings.Split(normalized, "\n")

	// If the last line is empty (trailing newline), drop it
	if len(rawLines) > 0 && rawLines[len(rawLines)-1] == "" {
		rawLines = rawLines[:len(rawLines)-1]
	}

	lines := make([]string, len(rawLines))
	for i, line := range rawLines {
		lines[i] = SanitizeLogLine(line)
	}

	return lines
}
