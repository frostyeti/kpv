package kvconf

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()
	if config == nil {
		t.Fatal("expected config, got nil")
	}
	if config.dict == nil {
		t.Fatal("expected dict to be initialized")
	}
}

func TestConfigSet(t *testing.T) {
	config := NewConfig()
	config.Set("key1", "value1")
	val, exists := config.Get("key1")
	if !exists || val != "value1" {
		t.Fatalf("expected value1, got %s (exists: %v)", val, exists)
	}
}

func TestConfigSetUpdate(t *testing.T) {
	config := NewConfig()
	config.Set("key1", "value1")
	config.Set("key1", "value2")
	val, _ := config.Get("key1")
	if val != "value2" {
		t.Fatalf("expected value2, got %s", val)
	}
}

func TestConfigParse(t *testing.T) {
	config := NewConfig()
	input := "key1=value1\nkey2=value2\n"
	config.Parse(input)
	val1, _ := config.Get("key1")
	val2, _ := config.Get("key2")
	if val1 != "value1" || val2 != "value2" {
		t.Fatalf("expected value1 and value2, got %s and %s", val1, val2)
	}
}

func TestConfigParseMultiLine(t *testing.T) {
	config := NewConfig()
	input := "key1=EOF\nline1\nline2\nEOF\n"
	config.Parse(input)
	val, _ := config.Get("key1")
	expected := "line1\nline2"
	if val != expected {
		t.Fatalf("expected %q, got %q", expected, val)
	}
}

func TestConfigString(t *testing.T) {
	config := NewConfig()
	config.Set("key1", "value1")
	result := config.String()
	expected := "key1=value1\n"
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestConfigAddLine(t *testing.T) {
	config := NewConfig()
	config.AddLine()
	result := config.String()
	if result != "\n" {
		t.Fatalf("expected newline, got %q", result)
	}
}

func TestConfigAddComment(t *testing.T) {
	config := NewConfig()
	config.AddComment("test comment")
	result := config.String()
	expected := "# test comment\n"
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestConfigGetMissing(t *testing.T) {
	config := NewConfig()
	_, exists := config.Get("nonexistent")
	if exists {
		t.Fatal("expected key to not exist")
	}
}
