package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// Client is the PostgreSQL abstraction used by API handlers.
type Client interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	Close() error
}

// New initializes a single-database PostgreSQL client from a DSN string and
// verifies connectivity. Most callers should use Factory instead — it shares
// connection params and supports switching databases on the same server.
func New(ctx context.Context, dsn string) (Client, error) {
	if dsn == "" {
		return nil, fmt.Errorf("empty PostgreSQL DSN")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return db, nil
}

// Factory holds the parsed connection params for a PostgreSQL server and
// hands out (lazily-created, cached) Clients per database name. It lets the
// HTTP handlers switch databases on the same server without re-prompting the
// user for credentials.
type Factory struct {
	base *pgx.ConnConfig

	mu      sync.Mutex
	clients map[string]Client
}

// NewFactory parses a DSN and returns a Factory. Connectivity is NOT verified
// here — call Get(ctx, "") to open and ping the default database.
func NewFactory(dsn string) (*Factory, error) {
	if dsn == "" {
		return nil, fmt.Errorf("empty PostgreSQL DSN")
	}
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PostgreSQL DSN: %w", err)
	}
	return &Factory{
		base:    cfg,
		clients: make(map[string]Client),
	}, nil
}

// DefaultDB returns the database name from the original DSN.
func (f *Factory) DefaultDB() string {
	return f.base.Database
}

// Get returns (and caches) a Client connected to dbname on the same server.
// An empty dbname falls back to the factory's default database.
func (f *Factory) Get(ctx context.Context, dbname string) (Client, error) {
	if dbname == "" {
		dbname = f.base.Database
	}

	f.mu.Lock()
	if c, ok := f.clients[dbname]; ok {
		f.mu.Unlock()
		return c, nil
	}
	f.mu.Unlock()

	cfg := f.base.Copy()
	cfg.Database = dbname
	registered := stdlib.RegisterConnConfig(cfg)

	db, err := sql.Open("pgx", registered)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection to %q: %w", dbname, err)
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL on %q: %w", dbname, err)
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	// Re-check in case of race.
	if c, ok := f.clients[dbname]; ok {
		_ = db.Close()
		return c, nil
	}
	f.clients[dbname] = db
	return db, nil
}

// Close closes all cached clients.
func (f *Factory) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range f.clients {
		_ = c.Close()
	}
	f.clients = nil
	return nil
}

// NewFromEnv initializes a PostgreSQL client from APE_PG_DSN.
func NewFromEnv(ctx context.Context) (Client, error) {
	dsn := os.Getenv("APE_PG_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("set APE_PG_DSN to enable RDS PostgreSQL")
	}
	return New(ctx, dsn)
}
