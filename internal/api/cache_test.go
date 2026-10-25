package api

import (
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	c := NewCache(1 * time.Second)

	c.Set("key1", "val1", 0)
	got, found := c.Get("key1")
	if !found {
		t.Fatalf("expected key1 to be found")
	}
	if got.(string) != "val1" {
		t.Errorf("got %v, want val1", got)
	}

	// Missing key
	_, found = c.Get("nonexistent")
	if found {
		t.Errorf("expected nonexistent key to return false")
	}
}

func TestCache_Expiration(t *testing.T) {
	c := NewCache(50 * time.Millisecond)

	c.Set("short", "data", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)

	_, found := c.Get("short")
	if found {
		t.Errorf("expected key to expire, but was found")
	}
}

func TestCache_Delete(t *testing.T) {
	c := NewCache(1 * time.Minute)
	c.Set("to_delete", "val", 0)

	c.Delete("to_delete")
	_, found := c.Get("to_delete")
	if found {
		t.Errorf("expected deleted key to not be found")
	}
}

func TestCache_Clear(t *testing.T) {
	c := NewCache(1 * time.Minute)
	c.Set("k1", "v1", 0)
	c.Set("k2", "v2", 0)

	c.Clear()
	if _, found := c.Get("k1"); found {
		t.Errorf("k1 found after Clear()")
	}
	if _, found := c.Get("k2"); found {
		t.Errorf("k2 found after Clear()")
	}
}
