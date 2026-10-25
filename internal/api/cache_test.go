package api

import (
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     any
		ttl       time.Duration
		sleep     time.Duration
		wantFound bool
		wantVal   any
	}{
		{
			name:      "retrieve existing item within TTL",
			key:       "valid_key",
			value:     "hello",
			ttl:       1 * time.Second,
			sleep:     0,
			wantFound: true,
			wantVal:   "hello",
		},
		{
			name:      "expire item after TTL elapses",
			key:       "expired_key",
			value:     "temp",
			ttl:       10 * time.Millisecond,
			sleep:     20 * time.Millisecond,
			wantFound: false,
			wantVal:   nil,
		},
		{
			name:      "lookup nonexistent key",
			key:       "missing_key",
			value:     nil,
			ttl:       0,
			sleep:     0,
			wantFound: false,
			wantVal:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCache(1 * time.Second)
			if tt.value != nil {
				c.Set(tt.key, tt.value, tt.ttl)
			}
			if tt.sleep > 0 {
				time.Sleep(tt.sleep)
			}

			got, found := c.Get(tt.key)
			if found != tt.wantFound {
				t.Fatalf("Get(%q) found = %v, want %v", tt.key, found, tt.wantFound)
			}
			if tt.wantFound && got != tt.wantVal {
				t.Errorf("Get(%q) = %v, want %v", tt.key, got, tt.wantVal)
			}
		})
	}
}

func TestCache_DeleteAndClear(t *testing.T) {
	tests := []struct {
		name        string
		setupKeys   map[string]any
		action      func(c *Cache)
		checkKey    string
		expectFound bool
	}{
		{
			name:      "delete specific key removes only that key",
			setupKeys: map[string]any{"k1": "v1", "k2": "v2"},
			action: func(c *Cache) {
				c.Delete("k1")
			},
			checkKey:    "k1",
			expectFound: false,
		},
		{
			name:      "delete leaves other keys intact",
			setupKeys: map[string]any{"k1": "v1", "k2": "v2"},
			action: func(c *Cache) {
				c.Delete("k1")
			},
			checkKey:    "k2",
			expectFound: true,
		},
		{
			name:      "clear flushes all cached items",
			setupKeys: map[string]any{"k1": "v1", "k2": "v2"},
			action: func(c *Cache) {
				c.Clear()
			},
			checkKey:    "k1",
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCache(1 * time.Minute)
			for k, v := range tt.setupKeys {
				c.Set(k, v, 0)
			}

			tt.action(c)

			_, found := c.Get(tt.checkKey)
			if found != tt.expectFound {
				t.Errorf("Get(%q) found = %v, want %v", tt.checkKey, found, tt.expectFound)
			}
		})
	}
}
