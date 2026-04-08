package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type SavedConnection struct {
	Name     string `yaml:"name"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	KeyPath  string `yaml:"key_path"`
}

// SavedRDSConnection persists everything needed to reconnect to an RDS
// PostgreSQL instance EXCEPT the password, which is always re-prompted.
// Tunnel=true means the connection should go through whichever EC2 SSH
// session is active at startup.
type SavedRDSConnection struct {
	Name     string `yaml:"name"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"sslmode"`
	Tunnel   bool   `yaml:"tunnel"`
}

type Config struct {
	Connections    []SavedConnection    `yaml:"connections"`
	RDSConnections []SavedRDSConnection `yaml:"rds_connections,omitempty"`
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ape", "config.yaml"), nil
}

// Load reads the config file. Returns empty config if file doesn't exist.
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return &Config{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &cfg, nil
}

// Save writes the config to disk, creating ~/.ape/ if needed.
func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}

// AddConnection appends a connection and saves. Overwrites if same name exists.
func (c *Config) AddConnection(conn SavedConnection) {
	for i, existing := range c.Connections {
		if existing.Name == conn.Name {
			c.Connections[i] = conn
			return
		}
	}
	c.Connections = append(c.Connections, conn)
}

// RemoveConnection removes a connection by name.
func (c *Config) RemoveConnection(name string) bool {
	for i, conn := range c.Connections {
		if conn.Name == name {
			c.Connections = append(c.Connections[:i], c.Connections[i+1:]...)
			return true
		}
	}
	return false
}

// AddRDSConnection appends an RDS connection and saves. Overwrites if same
// name exists.
func (c *Config) AddRDSConnection(conn SavedRDSConnection) {
	for i, existing := range c.RDSConnections {
		if existing.Name == conn.Name {
			c.RDSConnections[i] = conn
			return
		}
	}
	c.RDSConnections = append(c.RDSConnections, conn)
}

// RemoveRDSConnection removes an RDS connection by name.
func (c *Config) RemoveRDSConnection(name string) bool {
	for i, conn := range c.RDSConnections {
		if conn.Name == name {
			c.RDSConnections = append(c.RDSConnections[:i], c.RDSConnections[i+1:]...)
			return true
		}
	}
	return false
}
