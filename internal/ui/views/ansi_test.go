package views

import (
	"reflect"
	"testing"
)

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain string without escape codes",
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			name:     "string with SGR color escape codes",
			input:    "\x1b[32mSuccess\x1b[0m",
			expected: "Success",
		},
		{
			name:     "complex compound styling",
			input:    "\x1b[1;31;40mERROR: failed to connect\x1b[0m",
			expected: "ERROR: failed to connect",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripANSI(tt.input)
			if got != tt.expected {
				t.Errorf("StripANSI(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSanitizeLogLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain text",
			input:    "Standard output message",
			expected: "Standard output message",
		},
		{
			name:     "expands tabs to spaces",
			input:    "col1\tcol2\tcol3",
			expected: "col1    col2    col3",
		},
		{
			name:     "removes carriage return",
			input:    "progress: 50%\rprogress: 100%",
			expected: "progress: 50%progress: 100%",
		},
		{
			name:     "strips screen-clearing sequence",
			input:    "\x1b[2JStarting build...",
			expected: "Starting build...",
		},
		{
			name:     "strips cursor-movement sequence",
			input:    "\x1b[1;1H\x1b[2KRunning step 1",
			expected: "Running step 1",
		},
		{
			name:     "preserves readable ANSI colors",
			input:    "\x1b[32m✓\x1b[0m Step passed",
			expected: "\x1b[32m✓\x1b[0m Step passed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeLogLine(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeLogLine(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSanitizeLogs(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected []string
	}{
		{
			name:     "empty input returns nil",
			raw:      "",
			expected: nil,
		},
		{
			name:     "unix newline log lines",
			raw:      "line 1\nline 2\nline 3\n",
			expected: []string{"line 1", "line 2", "line 3"},
		},
		{
			name:     "windows CRLF log lines",
			raw:      "line 1\r\nline 2\r\nline 3\r\n",
			expected: []string{"line 1", "line 2", "line 3"},
		},
		{
			name:     "tabs and colors across lines",
			raw:      "Step 1:\t\x1b[32mOK\x1b[0m\nStep 2:\t\x1b[31mFAIL\x1b[0m",
			expected: []string{"Step 1:    \x1b[32mOK\x1b[0m", "Step 2:    \x1b[31mFAIL\x1b[0m"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeLogs(tt.raw)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("SanitizeLogs() = %#v, want %#v", got, tt.expected)
			}
		})
	}
}
