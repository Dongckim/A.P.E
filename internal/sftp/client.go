package sftp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// SFTPClient defines the interface for SFTP operations.
type SFTPClient interface {
	ListDirectory(path string) ([]FileInfo, error)
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, content []byte) error
	UploadFile(path string, reader io.Reader) error
	DownloadFile(path string, w io.Writer) error
	DeleteFile(path string) error
	RenameFile(oldPath, newPath string) error
	Stat(path string) (*FileInfo, error)
	MkdirAll(path string) error
	Exec(ctx context.Context, cmd string) (string, int, error)
	Close() error
}

// ConnectConfig holds the parameters needed to establish an SSH connection.
type ConnectConfig struct {
	Host       string
	Port       string
	Username   string
	KeyPath    string
	Passphrase string
}

func (c ConnectConfig) Addr() string {
	return net.JoinHostPort(c.Host, c.Port)
}

// Client wraps an SSH connection and SFTP session.
type Client struct {
	sshClient  *ssh.Client
	sftpClient *sftp.Client
	config     ConnectConfig
}

// Connect establishes an SSH connection and opens an SFTP session.
func Connect(cfg ConnectConfig) (*Client, error) {
	signer, err := loadKey(cfg.KeyPath, cfg.Passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to load SSH key %s: %w", cfg.KeyPath, err)
	}

	hostKeyCallback, err := buildHostKeyCallback()
	if err != nil {
		slog.Warn("known_hosts not available, accepting all host keys", "err", err)
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	sshCfg := &ssh.ClientConfig{
		User: cfg.Username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostKeyCallback,
	}

	sshConn, err := ssh.Dial("tcp", cfg.Addr(), sshCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", cfg.Addr(), err)
	}

	sftpConn, err := sftp.NewClient(sshConn)
	if err != nil {
		sshConn.Close()
		return nil, fmt.Errorf("failed to open SFTP session: %w", err)
	}

	return &Client{
		sshClient:  sshConn,
		sftpClient: sftpConn,
		config:     cfg,
	}, nil
}

// Exec runs a shell command on the remote server via SSH.
// Returns stdout, exit code, and any error. Times out based on context.
func (c *Client) Exec(ctx context.Context, cmd string) (string, int, error) {
	session, err := c.sshClient.NewSession()
	if err != nil {
		return "", -1, fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	// Default 5s timeout if context has no deadline
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	done := make(chan error, 1)
	go func() { done <- session.Run(cmd) }()

	select {
	case <-ctx.Done():
		session.Signal(ssh.SIGKILL)
		return stdout.String(), -1, fmt.Errorf("command timed out: %s", cmd)
	case err := <-done:
		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*ssh.ExitError); ok {
				exitCode = exitErr.ExitStatus()
			} else {
				return stdout.String(), -1, err
			}
		}
		return stdout.String(), exitCode, nil
	}
}

// Close shuts down the SFTP session and SSH connection.
func (c *Client) Close() error {
	if c.sftpClient != nil {
		c.sftpClient.Close()
	}
	if c.sshClient != nil {
		return c.sshClient.Close()
	}
	return nil
}

// Config returns the connection config.
func (c *Client) Config() ConnectConfig {
	return c.config
}

func loadKey(keyPath, passphrase string) (ssh.Signer, error) {
	keyPath = expandHome(keyPath)
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	if passphrase != "" {
		signer, err := ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(passphrase))
		if err != nil {
			return nil, fmt.Errorf("failed to parse key with passphrase: %w", err)
		}
		return signer, nil
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		if strings.Contains(err.Error(), "passphrase") {
			return nil, fmt.Errorf("key is passphrase-protected, please provide a passphrase: %w", err)
		}
		return nil, fmt.Errorf("failed to parse key: %w", err)
	}
	return signer, nil
}

func buildHostKeyCallback() (ssh.HostKeyCallback, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	knownHostsPath := filepath.Join(home, ".ssh", "known_hosts")
	if _, err := os.Stat(knownHostsPath); err != nil {
		return nil, fmt.Errorf("known_hosts not found: %w", err)
	}
	return knownhosts.New(knownHostsPath)
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
