package sftp

import (
	"testing"
)

func TestConnectConfigAddr(t *testing.T) {
	tests := []struct {
		name string
		cfg  ConnectConfig
		want string
	}{
		{"standard", ConnectConfig{Host: "54.123.45.67", Port: "22"}, "54.123.45.67:22"},
		{"custom port", ConnectConfig{Host: "10.0.0.1", Port: "2222"}, "10.0.0.1:2222"},
		{"hostname", ConnectConfig{Host: "ec2.example.com", Port: "22"}, "ec2.example.com:22"},
		{"empty host", ConnectConfig{Host: "", Port: "22"}, ":22"},
		{"empty port", ConnectConfig{Host: "10.0.0.1", Port: ""}, "10.0.0.1:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.Addr()
			if got != tt.want {
				t.Errorf("Addr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExpandHome(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTilde bool // if false, result should NOT start with ~/
	}{
		{"absolute path unchanged", "/usr/local/bin", false},
		{"relative no tilde", "keys/my.pem", false},
		{"tilde expanded", "~/some/path", false},
		{"tilde no slash unchanged", "~other", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandHome(tt.input)
			if tt.input == "/usr/local/bin" && got != "/usr/local/bin" {
				t.Errorf("absolute path changed: %q -> %q", tt.input, got)
			}
			if tt.input == "keys/my.pem" && got != "keys/my.pem" {
				t.Errorf("relative path changed: %q -> %q", tt.input, got)
			}
			if tt.input == "~/some/path" && got == "~/some/path" {
				t.Errorf("tilde path was not expanded: %q", got)
			}
			if tt.input == "~other" && got != "~other" {
				t.Errorf("~other should not be expanded, got: %q", got)
			}
		})
	}
}
