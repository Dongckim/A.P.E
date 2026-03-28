package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/dongchankim/ape/internal/api"
	"github.com/dongchankim/ape/internal/config"
	"github.com/dongchankim/ape/internal/s3"
	"github.com/dongchankim/ape/internal/server"
	"github.com/dongchankim/ape/internal/sftp"
)

// ANSI color (only used for the ⏺ dot)
const (
	cReset  = "\033[0m"
	cRed    = "\033[31m"
	cGreen  = "\033[32m"
	cYellow = "\033[33m"
)

// Status dots
var (
	dotOK   = cGreen + "⏺" + cReset
	dotFail = cRed + "⏺" + cReset
	dotWarn = cYellow + "⏺" + cReset
)

const banner = `
          ▄▄██████████▄▄
        ▄████████████████▄
       ████████████████████
       ███  (◕)    (◕)  ███
       ████     ▄▄     ████
        ████ ┌──────┐ ████
         ████│ ━━━━ │████
          ▀██└──────┘██▀
            ▀████████▀

       ██████  ██████  ██████
       ██  ██  ██  ██  ██
       ██████  ██████  ████
       ██  ██  ██      ██
       ██  ██  ██      ██████

         AWS Platform Explorer
               v0.1.0

`

const helpText = `
──────────────────────────────────────────
A.P.E is running. Commands:
  /add     — connect additional EC2
  /list    — list active connections
  /status  — show connection info
  /h       — show this help
  /q       — quit A.P.E
──────────────────────────────────────────
`

var reader *bufio.Reader

func Execute() {
	reader = bufio.NewReader(os.Stdin)

	fmt.Print("\033[2J\033[H") // clear screen, cursor to top
	fmt.Print(banner)

	connMgr := api.NewConnectionManager()

	// Initialize S3 client (non-fatal if AWS credentials are missing)
	var s3Client s3.S3Client
	s3c, err := s3.New(context.Background(), "")
	if err != nil {
		fmt.Printf("  %s S3 not available: %s\n\n", dotWarn, err.Error())
	} else {
		s3Client = s3c
		fmt.Printf("  %s S3 client initialized\n\n", dotOK)
	}

	srv := server.New("127.0.0.1:9000", connMgr, s3Client)

	cfg := promptOrSelectConnection()
	connectAndRegister(cfg, connMgr)

	go func() {
		if err := srv.Start(); err != nil && err.Error() != "http: Server closed" {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	fmt.Printf("\n  %s Web UI ready at http://localhost:9000\n", dotOK)
	openBrowser("http://localhost:9000")

	fmt.Print(helpText)
	repl(srv, connMgr)
}

func connectAndRegister(cfg sftp.ConnectConfig, connMgr *api.ConnectionManager) {
	fmt.Printf("  Connecting to %s@%s:%s ...\n", cfg.Username, cfg.Host, cfg.Port)

	client, err := sftp.Connect(cfg)
	if err != nil {
		fmt.Printf("  %s Connection failed: %s\n\n", dotFail, err)
		fmt.Println("  Retrying — fix the failing field:")

		errMsg := err.Error()
		for {
			if strings.Contains(errMsg, "failed to load SSH key") || strings.Contains(errMsg, "passphrase") {
				cfg.KeyPath = promptWithDefault("SSH key path", "~/.ssh/id_rsa")
				if strings.Contains(errMsg, "passphrase") {
					cfg.Passphrase = promptRequired("Key passphrase")
				}
			} else if strings.Contains(errMsg, "failed to connect") {
				cfg.Host = promptRequired("EC2 host")
				cfg.Port = promptWithDefault("SSH port", cfg.Port)
				cfg.Username = promptWithDefault("Username", cfg.Username)
			} else {
				cfg = promptConnectionFull()
			}

			fmt.Printf("\n  Retrying %s@%s:%s ...\n", cfg.Username, cfg.Host, cfg.Port)
			client, err = sftp.Connect(cfg)
			if err == nil {
				break
			}
			fmt.Printf("  %s Connection failed: %s\n\n", dotFail, err)
			errMsg = err.Error()
		}
	}

	id := connMgr.Add(client)
	fmt.Printf("  %s Connected!\n\n", dotOK)
	slog.Debug("connection registered", "id", id)

	promptSaveConnection(cfg)
}

func promptSaveConnection(cfg sftp.ConnectConfig) {
	savedCfg, err := config.Load()
	if err != nil {
		return
	}

	// Skip if already saved with same host/port/username
	for _, c := range savedCfg.Connections {
		if c.Host == cfg.Host && c.Port == cfg.Port && c.Username == cfg.Username {
			return
		}
	}

	save := promptWithDefault("Save this connection? (y/n)", "y")
	if !strings.EqualFold(save, "y") {
		return
	}

	name := promptRequired("Connection name")
	if name == "" {
		name = fmt.Sprintf("%s@%s", cfg.Username, cfg.Host)
	}

	savedCfg.AddConnection(config.SavedConnection{
		Name:     name,
		Host:     cfg.Host,
		Port:     cfg.Port,
		Username: cfg.Username,
		KeyPath:  cfg.KeyPath,
	})

	if err := config.Save(savedCfg); err != nil {
		fmt.Printf("  %s Failed to save connection: %s\n", dotWarn, err)
		return
	}
	fmt.Printf("  %s Connection saved as \"%s\"\n", dotOK, name)
}

func promptOrSelectConnection() sftp.ConnectConfig {
	savedCfg, err := config.Load()
	if err != nil {
		slog.Debug("failed to load config", "err", err)
		return promptConnection()
	}

	if len(savedCfg.Connections) == 0 {
		return promptConnection()
	}

	fmt.Println("  Saved connections:")
	for i, c := range savedCfg.Connections {
		fmt.Printf("    [%d] %s — %s@%s:%s\n", i+1, c.Name, c.Username, c.Host, c.Port)
	}
	fmt.Printf("    [N] New connection\n\n")

	choice := promptWithDefault("Select connection", "N")

	if strings.EqualFold(choice, "n") || choice == "" {
		return promptConnection()
	}

	idx := 0
	if _, err := fmt.Sscanf(choice, "%d", &idx); err != nil || idx < 1 || idx > len(savedCfg.Connections) {
		fmt.Printf("  %s Invalid choice, starting new connection\n\n", dotWarn)
		return promptConnection()
	}

	sc := savedCfg.Connections[idx-1]
	fmt.Printf("  Using saved connection: %s\n\n", sc.Name)
	return sftp.ConnectConfig{
		Host:     sc.Host,
		Port:     sc.Port,
		Username: sc.Username,
		KeyPath:  expandHome(sc.KeyPath),
	}
}

func promptConnection() sftp.ConnectConfig {
	keyPath := promptWithDefault("SSH key path", "~/.ssh/id_rsa")
	expanded := expandHome(keyPath)
	if _, err := os.Stat(expanded); err != nil {
		fmt.Printf("  %s Key file not found: %s (will verify on connect)\n\n", dotWarn, keyPath)
	} else {
		keyPath = expanded
	}

	var host string
	for attempts := 0; attempts < 3; attempts++ {
		host = promptRequired("EC2 host")
		if host != "" {
			break
		}
		fmt.Printf("  %s Host cannot be empty\n", dotFail)
	}
	if host == "" {
		fmt.Printf("  %s No host provided, exiting.\n", dotFail)
		os.Exit(1)
	}

	username := promptWithDefault("Username", "ubuntu")

	port := promptWithDefault("SSH port", "22")
	if !isValidPort(port) {
		fmt.Printf("  %s Invalid port '%s', using default 22\n", dotWarn, port)
		port = "22"
	}

	return sftp.ConnectConfig{
		Host:     host,
		Port:     port,
		Username: username,
		KeyPath:  keyPath,
	}
}

func promptConnectionFull() sftp.ConnectConfig {
	return promptConnection()
}

func repl(srv *server.Server, connMgr *api.ConnectionManager) {
	for {
		fmt.Print("\nape ▸ ")
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		input = strings.TrimSpace(input)

		switch input {
		case "/add":
			fmt.Println("\n— Add new EC2 connection —")
			cfg := promptConnection()
			connectAndRegister(cfg, connMgr)

		case "/list":
			conns := connMgr.List()
			if len(conns) == 0 {
				fmt.Println("No active connections.")
				continue
			}
			fmt.Println("\nActive connections:")
			for i, c := range conns {
				fmt.Printf("  %s [%d] %s (%s)\n", dotOK, i+1, c.ID, c.Host)
			}

		case "/status":
			conns := connMgr.List()
			if len(conns) == 0 {
				fmt.Println("No active connections.")
				continue
			}
			c := conns[0]
			fmt.Printf("\nCurrent connection:\n")
			fmt.Printf("  %s Host:     %s\n", dotOK, c.Host)
			fmt.Printf("    Username: %s\n", c.Username)
			fmt.Printf("    Port:     %s\n", c.Port)
			fmt.Printf("    Web UI:   http://localhost:9000\n")

		case "/h":
			fmt.Print(helpText)

		case "/q":
			fmt.Println("Shutting down...")
			if err := srv.Shutdown(); err != nil {
				slog.Error("shutdown error", "err", err)
			}
			fmt.Println("Goodbye! 🦍")
			os.Exit(0)

		case "":
			continue

		default:
			fmt.Printf("  %s Unknown command: %s (type /h for help)\n", dotFail, input)
		}
	}
}

func promptWithDefault(label, defaultVal string) string {
	fmt.Printf("? %s (%s): ", label, defaultVal)
	input, _ := reader.ReadString('\n')
	input = cleanInput(input)
	if input == "" {
		return defaultVal
	}
	return input
}

func promptRequired(label string) string {
	fmt.Printf("? %s: ", label)
	input, _ := reader.ReadString('\n')
	return cleanInput(input)
}

func cleanInput(s string) string {
	s = strings.TrimSpace(s)
	// Strip surrounding quotes (single or double)
	if len(s) >= 2 && ((s[0] == '\'' && s[len(s)-1] == '\'') || (s[0] == '"' && s[len(s)-1] == '"')) {
		s = s[1 : len(s)-1]
	}
	return s
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home + path[1:]
	}
	return path
}

func isValidPort(port string) bool {
	_, err := net.LookupPort("tcp", port)
	return err == nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return
	}
	_ = cmd.Start()
}
