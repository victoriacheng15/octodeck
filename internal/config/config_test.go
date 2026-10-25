package config

import (
	"os"
	"testing"
)

func TestResolveHost(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "custom enterprise host preserved",
			input:    "github.mycorp.internal",
			expected: "github.mycorp.internal",
		},
		{
			name:     "empty host defaults to github.com or default host",
			input:    "",
			expected: "github.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveHost(tt.input)
			if tt.input != "" && got != tt.expected {
				t.Errorf("ResolveHost(%q) = %q, want %q", tt.input, got, tt.expected)
			}
			if tt.input == "" && got == "" {
				t.Errorf("ResolveHost(%q) returned empty string, expected valid fallback", tt.input)
			}
		})
	}
}

func TestResolveRepo(t *testing.T) {
	tests := []struct {
		name      string
		flag      string
		wantOwner string
		wantName  string
		wantHost  string
		wantErr   bool
	}{
		{
			name:      "standard owner/repo flag",
			flag:      "charmbracelet/bubbletea",
			wantOwner: "charmbracelet",
			wantName:  "bubbletea",
			wantHost:  "github.com",
			wantErr:   false,
		},
		{
			name:      "enterprise host with owner/repo",
			flag:      "github.corp.net/team/project",
			wantOwner: "team",
			wantName:  "project",
			wantHost:  "github.corp.net",
			wantErr:   false,
		},
		{
			name:    "invalid format with multiple slashes",
			flag:    "a/b/c/d",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, err := ResolveRepo(tt.flag)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for flag %q, got nil", tt.flag)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ctx.Owner != tt.wantOwner || ctx.Name != tt.wantName {
				t.Errorf("ResolveRepo(%q) = %s/%s, want %s/%s", tt.flag, ctx.Owner, ctx.Name, tt.wantOwner, tt.wantName)
			}
			if ctx.Host != tt.wantHost {
				t.Errorf("ResolveRepo(%q).Host = %q, want %q", tt.flag, ctx.Host, tt.wantHost)
			}
		})
	}
}

func TestCheckAuth(t *testing.T) {
	// Temporarily override GH_TOKEN
	origToken := os.Getenv("GH_TOKEN")
	defer os.Setenv("GH_TOKEN", origToken)

	os.Setenv("GH_TOKEN", "mock-token-secret")
	token, err := CheckAuth("github.com")
	if err != nil {
		t.Fatalf("unexpected error with GH_TOKEN set: %v", err)
	}
	if token != "mock-token-secret" {
		t.Errorf("CheckAuth() = %q, want %q", token, "mock-token-secret")
	}
}
