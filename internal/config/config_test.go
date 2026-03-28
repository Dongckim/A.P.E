package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadNoFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Connections) != 0 {
		t.Fatalf("expected 0 connections, got %d", len(cfg.Connections))
	}
}

func TestLoadValidYAML(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir := filepath.Join(tmp, ".ape")
	os.MkdirAll(dir, 0700)

	data := []byte("connections:\n  - name: test\n    host: 1.2.3.4\n    port: \"22\"\n    username: ubuntu\n    key_path: ~/.ssh/id_rsa\n")
	os.WriteFile(filepath.Join(dir, "config.yaml"), data, 0600)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Connections) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(cfg.Connections))
	}
	if cfg.Connections[0].Name != "test" {
		t.Fatalf("expected name 'test', got '%s'", cfg.Connections[0].Name)
	}
	if cfg.Connections[0].Host != "1.2.3.4" {
		t.Fatalf("expected host '1.2.3.4', got '%s'", cfg.Connections[0].Host)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir := filepath.Join(tmp, ".ape")
	os.MkdirAll(dir, 0700)
	os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(":\n\t:::bad"), 0600)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestSaveCreatesDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg := &Config{
		Connections: []SavedConnection{
			{Name: "dev", Host: "10.0.0.1", Port: "22", Username: "ec2-user", KeyPath: "~/.ssh/key.pem"},
		},
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := filepath.Join(tmp, ".ape", "config.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	original := &Config{
		Connections: []SavedConnection{
			{Name: "prod", Host: "54.1.2.3", Port: "2222", Username: "admin", KeyPath: "/keys/prod.pem"},
			{Name: "staging", Host: "10.0.0.5", Port: "22", Username: "ubuntu", KeyPath: "~/.ssh/id_rsa"},
		},
	}

	if err := Save(original); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("load error: %v", err)
	}

	if len(loaded.Connections) != 2 {
		t.Fatalf("expected 2 connections, got %d", len(loaded.Connections))
	}
	if loaded.Connections[0].Name != "prod" || loaded.Connections[1].Name != "staging" {
		t.Fatal("round-trip data mismatch")
	}
}

func TestAddConnection(t *testing.T) {
	cfg := &Config{}

	cfg.AddConnection(SavedConnection{Name: "a", Host: "1.1.1.1"})
	if len(cfg.Connections) != 1 {
		t.Fatalf("expected 1, got %d", len(cfg.Connections))
	}

	cfg.AddConnection(SavedConnection{Name: "b", Host: "2.2.2.2"})
	if len(cfg.Connections) != 2 {
		t.Fatalf("expected 2, got %d", len(cfg.Connections))
	}
}

func TestAddConnectionOverwrite(t *testing.T) {
	cfg := &Config{}

	cfg.AddConnection(SavedConnection{Name: "a", Host: "1.1.1.1"})
	cfg.AddConnection(SavedConnection{Name: "a", Host: "9.9.9.9"})

	if len(cfg.Connections) != 1 {
		t.Fatalf("expected 1 after overwrite, got %d", len(cfg.Connections))
	}
	if cfg.Connections[0].Host != "9.9.9.9" {
		t.Fatalf("expected overwritten host '9.9.9.9', got '%s'", cfg.Connections[0].Host)
	}
}

func TestRemoveConnection(t *testing.T) {
	cfg := &Config{}
	cfg.AddConnection(SavedConnection{Name: "a", Host: "1.1.1.1"})
	cfg.AddConnection(SavedConnection{Name: "b", Host: "2.2.2.2"})

	if !cfg.RemoveConnection("a") {
		t.Fatal("expected true when removing existing connection")
	}
	if len(cfg.Connections) != 1 {
		t.Fatalf("expected 1 after remove, got %d", len(cfg.Connections))
	}
	if cfg.Connections[0].Name != "b" {
		t.Fatalf("expected remaining connection 'b', got '%s'", cfg.Connections[0].Name)
	}
}

func TestRemoveConnectionNotFound(t *testing.T) {
	cfg := &Config{}
	cfg.AddConnection(SavedConnection{Name: "a"})

	if cfg.RemoveConnection("nonexistent") {
		t.Fatal("expected false when removing non-existent connection")
	}
	if len(cfg.Connections) != 1 {
		t.Fatalf("expected 1, got %d", len(cfg.Connections))
	}
}
