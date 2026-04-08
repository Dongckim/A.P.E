package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/dongchankim/ape/internal/api"
	"github.com/dongchankim/ape/internal/config"
	"github.com/dongchankim/ape/internal/postgres"
	"github.com/dongchankim/ape/internal/s3"
	"github.com/dongchankim/ape/internal/server"
	"github.com/dongchankim/ape/internal/sftp"
	"golang.org/x/term"
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

	// EC2 connection first — its SSH client may be used to tunnel RDS.
	cfg := promptOrSelectConnection()
	bastion := connectAndRegister(cfg, connMgr)

	// Initialize PostgreSQL factory (optional; interactive prompt or APE_PG_DSN).
	// If an EC2 connection is up, ape can open an in-process SSH tunnel so
	// the user does not have to run `ssh -L` manually. The factory lets the
	// UI switch databases on the same server without re-prompting.
	pgFactory := promptRDSConnection(bastion)

	srv := server.New("127.0.0.1:9000", connMgr, s3Client, pgFactory)
	if srv.FrontendStaleHint != "" {
		fmt.Printf("  %s %s\n\n", dotWarn, srv.FrontendStaleHint)
	}

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

func connectAndRegister(cfg sftp.ConnectConfig, connMgr *api.ConnectionManager) *sftp.Client {
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
	return client
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

// promptRDSConnection asks the user whether to connect to an RDS PostgreSQL
// instance and gathers credentials interactively. Returns nil if the user
// declines or retries are exhausted. If APE_PG_DSN is set, it is used directly.
//
// If saved RDS connections exist in ~/.ape/config.yaml, the user is shown a
// menu to pick one (the password is always re-prompted; only host/port/user/
// db/sslmode/tunnel are persisted).
//
// If bastion is non-nil and the user opts in, ape opens an in-process SSH
// tunnel through the EC2 connection (no separate `ssh -L` terminal needed):
// it binds a random local port on 127.0.0.1, forwards each accepted
// connection through SSH to the RDS endpoint.
//
// The returned *postgres.Factory caches per-database clients, so the UI can
// switch databases on the same server without re-prompting for credentials.
func promptRDSConnection(bastion *sftp.Client) *postgres.Factory {
	if dsn := os.Getenv("APE_PG_DSN"); dsn != "" {
		factory, err := postgres.NewFactory(dsn)
		if err != nil {
			fmt.Printf("  %s PostgreSQL (APE_PG_DSN) failed: %s\n\n", dotWarn, err)
			return nil
		}
		if _, err := factory.Get(context.Background(), ""); err != nil {
			fmt.Printf("  %s PostgreSQL (APE_PG_DSN) failed: %s\n\n", dotWarn, err)
			return nil
		}
		fmt.Printf("  %s PostgreSQL connected (from APE_PG_DSN)\n\n", dotOK)
		return factory
	}

	// Show saved RDS connections if any.
	savedCfg, _ := config.Load()
	saved := []config.SavedRDSConnection{}
	if savedCfg != nil {
		saved = savedCfg.RDSConnections
	}

	if len(saved) > 0 {
		fmt.Println()
		fmt.Println("  Saved RDS connections:")
		for i, c := range saved {
			tunnelMark := ""
			if c.Tunnel {
				tunnelMark = " [tunneled]"
			}
			fmt.Printf("    [%d] %s — %s@%s:%s/%s%s\n", i+1, c.Name, c.Username, c.Host, c.Port, c.Database, tunnelMark)
		}
		fmt.Printf("    [N] New RDS connection\n")
		fmt.Printf("    [S] Skip RDS\n\n")

		choice := promptWithDefault("Select RDS connection", "S")
		if strings.EqualFold(choice, "s") || choice == "" {
			fmt.Printf("  %s RDS skipped\n\n", dotWarn)
			return nil
		}
		if !strings.EqualFold(choice, "n") {
			idx := 0
			if _, err := fmt.Sscanf(choice, "%d", &idx); err == nil && idx >= 1 && idx <= len(saved) {
				if f := connectFromSavedRDS(saved[idx-1], bastion); f != nil {
					return f
				}
				// Fall through to new connection prompt on failure.
				fmt.Println("  Falling back to new connection setup.")
			} else {
				fmt.Printf("  %s Invalid choice, starting new connection\n", dotWarn)
			}
		}
	} else {
		answer := promptWithDefault("Connect to RDS PostgreSQL? (y/N)", "N")
		if !strings.EqualFold(answer, "y") {
			fmt.Printf("  %s RDS skipped\n\n", dotWarn)
			return nil
		}
	}

	// Decide whether to tunnel through the EC2 SSH connection.
	useTunnel := false
	if bastion != nil {
		fmt.Printf("\n  %s EC2 SSH connection active (%s).\n", dotOK, bastion.Config().Host)
		fmt.Println("    For RDS in a private subnet, ape can tunnel through this EC2.")
		ans := promptWithDefault("Tunnel RDS through this EC2? (Y/n)", "Y")
		useTunnel = strings.EqualFold(ans, "y") || ans == ""
	}

	for {
		rdsHost := promptRequired("RDS endpoint host")
		if rdsHost == "" {
			fmt.Printf("  %s Host required, skipping RDS\n\n", dotWarn)
			return nil
		}
		rdsPort := promptWithDefault("RDS port", "5432")
		if !isValidPort(rdsPort) {
			fmt.Printf("  %s Invalid port '%s', using 5432\n", dotWarn, rdsPort)
			rdsPort = "5432"
		}
		username := promptWithDefault("Username", "postgres")
		database := promptWithDefault("Database", "postgres")
		password := promptPassword("Password")

		// If tunneling, open the local forwarder and rewrite the connect target
		// to 127.0.0.1:<localPort>. SSL still happens end-to-end with the real
		// RDS server (the tunnel only relays TCP bytes), so sslmode=require
		// works — and is in fact required because RDS rejects unencrypted
		// connections via pg_hba.conf. We use `require` (not verify-full)
		// because pgx would otherwise try to verify the cert against
		// "127.0.0.1" instead of the RDS hostname.
		dialHost, dialPort := rdsHost, rdsPort
		sslmode := "require"
		tunnelLabel := ""
		if useTunnel {
			remoteAddr := net.JoinHostPort(rdsHost, rdsPort)
			fmt.Printf("\n  Opening SSH tunnel to %s ...\n", remoteAddr)
			localAddr, err := bastion.StartLocalForward(remoteAddr)
			if err != nil {
				fmt.Printf("  %s Failed to open tunnel: %s\n", dotFail, err)
				retry := promptWithDefault("Retry? (y/N)", "N")
				if !strings.EqualFold(retry, "y") {
					return nil
				}
				continue
			}
			lh, lp, _ := net.SplitHostPort(localAddr)
			dialHost, dialPort = lh, lp
			tunnelLabel = fmt.Sprintf(" (tunneled via %s)", bastion.Config().Host)
			fmt.Printf("  %s Tunnel ready: %s -> %s\n", dotOK, localAddr, remoteAddr)
		} else {
			// Direct connection — let user override sslmode
			sslmode = promptWithDefault("SSL mode", "require")
		}

		dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			url.QueryEscape(username),
			url.QueryEscape(password),
			dialHost, dialPort,
			url.PathEscape(database),
			url.QueryEscape(sslmode),
		)

		fmt.Printf("\n  Connecting to %s@%s:%s/%s%s ...\n", username, rdsHost, rdsPort, database, tunnelLabel)
		factory, err := postgres.NewFactory(dsn)
		if err == nil {
			if _, perr := factory.Get(context.Background(), ""); perr == nil {
				fmt.Printf("  %s PostgreSQL connected%s\n\n", dotOK, tunnelLabel)
				promptSaveRDSConnection(rdsHost, rdsPort, username, database, sslmode, useTunnel)
				return factory
			} else {
				err = perr
			}
		}
		fmt.Printf("  %s Connection failed: %s\n", dotFail, err)

		retry := promptWithDefault("Retry? (y/N)", "N")
		if !strings.EqualFold(retry, "y") {
			fmt.Printf("  %s RDS skipped\n\n", dotWarn)
			return nil
		}
	}
}

// connectFromSavedRDS reuses a saved RDS connection: it opens the SSH tunnel
// (if the saved entry was tunneled), prompts only for the password, and
// returns the resulting Factory. Returns nil on failure (caller may fall
// back to the new-connection prompt).
func connectFromSavedRDS(sc config.SavedRDSConnection, bastion *sftp.Client) *postgres.Factory {
	fmt.Printf("\n  Using saved RDS: %s (%s@%s:%s/%s)\n", sc.Name, sc.Username, sc.Host, sc.Port, sc.Database)

	dialHost, dialPort := sc.Host, sc.Port
	tunnelLabel := ""
	if sc.Tunnel {
		if bastion == nil {
			fmt.Printf("  %s This RDS was saved with tunnel=true but no EC2 connection is active.\n", dotWarn)
			return nil
		}
		remoteAddr := net.JoinHostPort(sc.Host, sc.Port)
		fmt.Printf("  Opening SSH tunnel to %s ...\n", remoteAddr)
		localAddr, err := bastion.StartLocalForward(remoteAddr)
		if err != nil {
			fmt.Printf("  %s Failed to open tunnel: %s\n", dotFail, err)
			return nil
		}
		lh, lp, _ := net.SplitHostPort(localAddr)
		dialHost, dialPort = lh, lp
		tunnelLabel = fmt.Sprintf(" (tunneled via %s)", bastion.Config().Host)
		fmt.Printf("  %s Tunnel ready: %s -> %s\n", dotOK, localAddr, remoteAddr)
	}

	// Password is never persisted — always re-prompt.
	password := promptPassword("Password")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		url.QueryEscape(sc.Username),
		url.QueryEscape(password),
		dialHost, dialPort,
		url.PathEscape(sc.Database),
		url.QueryEscape(sc.SSLMode),
	)

	fmt.Printf("\n  Connecting to %s@%s:%s/%s%s ...\n", sc.Username, sc.Host, sc.Port, sc.Database, tunnelLabel)
	factory, err := postgres.NewFactory(dsn)
	if err == nil {
		if _, perr := factory.Get(context.Background(), ""); perr == nil {
			fmt.Printf("  %s PostgreSQL connected%s\n\n", dotOK, tunnelLabel)
			return factory
		} else {
			err = perr
		}
	}
	fmt.Printf("  %s Connection failed: %s\n", dotFail, err)
	return nil
}

// promptSaveRDSConnection asks the user whether to persist a successful RDS
// connection (everything except the password) to ~/.ape/config.yaml.
func promptSaveRDSConnection(host, port, user, database, sslmode string, tunnel bool) {
	savedCfg, err := config.Load()
	if err != nil {
		return
	}

	// Skip if already saved with same host/port/user/db.
	for _, c := range savedCfg.RDSConnections {
		if c.Host == host && c.Port == port && c.Username == user && c.Database == database {
			return
		}
	}

	save := promptWithDefault("Save this RDS connection? (y/n)", "y")
	if !strings.EqualFold(save, "y") {
		return
	}

	name := promptRequired("Connection name")
	if name == "" {
		name = fmt.Sprintf("%s@%s/%s", user, host, database)
	}

	savedCfg.AddRDSConnection(config.SavedRDSConnection{
		Name:     name,
		Host:     host,
		Port:     port,
		Username: user,
		Database: database,
		SSLMode:  sslmode,
		Tunnel:   tunnel,
	})

	if err := config.Save(savedCfg); err != nil {
		fmt.Printf("  %s Failed to save RDS connection: %s\n", dotWarn, err)
		return
	}
	fmt.Printf("  %s RDS connection saved as \"%s\" (password not stored)\n", dotOK, name)
}

func promptPassword(label string) string {
	fmt.Printf("? %s: ", label)
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		input, _ := reader.ReadString('\n')
		return cleanInput(input)
	}
	bytePwd, err := term.ReadPassword(fd)
	fmt.Println()
	if err != nil {
		return ""
	}
	return string(bytePwd)
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
