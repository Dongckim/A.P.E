package api

import (
	"context"
	"net/http"
	"time"

	"github.com/dongchankim/ape/internal/postgres"
)

type RDSHandler struct {
	factory *postgres.Factory
}

func NewRDSHandler(factory *postgres.Factory) *RDSHandler {
	return &RDSHandler{factory: factory}
}

type RDSOverview struct {
	Version      string               `json:"version"`
	CurrentDB    string               `json:"current_db"`
	SchemaCount  int                  `json:"schema_count"`
	TableCount   int                  `json:"table_count"`
	Schemas      []RDSSchemaSummary   `json:"schemas"`
	Databases    []RDSDatabaseSummary `json:"databases"`
	Connected    bool                 `json:"connected"`
	ErrorMessage string               `json:"error,omitempty"`
}

type RDSSchemaSummary struct {
	Name       string `json:"name"`
	TableCount int    `json:"table_count"`
}

type RDSDatabaseSummary struct {
	Name       string `json:"name"`
	SizeBytes  int64  `json:"size_bytes"`
	SizePretty string `json:"size_pretty"`
	IsCurrent  bool   `json:"is_current"`
}

type RDSTableSummary struct {
	Name        string `json:"name"`
	RowEstimate int64  `json:"row_estimate"`
	SizeBytes   int64  `json:"size_bytes"`
	SizePretty  string `json:"size_pretty"`
}

type RDSTablesResponse struct {
	Database string            `json:"database"`
	Schema   string            `json:"schema"`
	Tables   []RDSTableSummary `json:"tables"`
}

// HandleOverview handles GET /api/rds/overview[?db=<dbname>].
//
// The optional `db` query parameter switches the connection to a different
// database on the same server (using the cached Factory). If omitted, the
// factory's default database (the one specified during the initial CLI
// prompt) is used.
func (h *RDSHandler) HandleOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if h.factory == nil {
		writeJSON(w, http.StatusOK, rdsOverviewError("PostgreSQL is not configured. Restart A.P.E and connect to RDS at startup."))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	requestedDB := r.URL.Query().Get("db")
	client, err := h.factory.Get(ctx, requestedDB)
	if err != nil {
		writeJSON(w, http.StatusOK, rdsOverviewError("failed to connect to database: "+err.Error()))
		return
	}

	out := RDSOverview{
		Connected: true,
		Schemas:   []RDSSchemaSummary{},
		Databases: []RDSDatabaseSummary{},
	}

	if err := client.QueryRowContext(ctx, "SELECT version()").Scan(&out.Version); err != nil {
		writeJSON(w, http.StatusOK, rdsOverviewError("failed to read postgres version: "+err.Error()))
		return
	}
	if err := client.QueryRowContext(ctx, "SELECT current_database()").Scan(&out.CurrentDB); err != nil {
		writeJSON(w, http.StatusOK, rdsOverviewError("failed to read current database: "+err.Error()))
		return
	}
	if err := client.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM information_schema.schemata
		WHERE schema_name NOT IN ('pg_catalog', 'information_schema')
		  AND schema_name NOT LIKE 'pg_toast%'
		  AND schema_name NOT LIKE 'pg_temp_%'
	`).Scan(&out.SchemaCount); err != nil {
		writeJSON(w, http.StatusOK, rdsOverviewError("failed to count schemas: "+err.Error()))
		return
	}
	if err := client.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_type='BASE TABLE'
		  AND table_schema NOT IN ('pg_catalog', 'information_schema')
		  AND table_schema NOT LIKE 'pg_toast%'
		  AND table_schema NOT LIKE 'pg_temp_%'
	`).Scan(&out.TableCount); err != nil {
		writeJSON(w, http.StatusOK, rdsOverviewError("failed to count tables: "+err.Error()))
		return
	}

	rows, err := client.QueryContext(ctx, `
		SELECT table_schema, COUNT(*) AS table_count
		FROM information_schema.tables
		WHERE table_type='BASE TABLE'
		  AND table_schema NOT IN ('pg_catalog', 'information_schema')
		GROUP BY table_schema
		ORDER BY table_count DESC, table_schema ASC
		LIMIT 20
	`)
	if err != nil {
		writeJSON(w, http.StatusOK, rdsOverviewError("failed to query schema overview: "+err.Error()))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var s RDSSchemaSummary
		if err := rows.Scan(&s.Name, &s.TableCount); err != nil {
			writeJSON(w, http.StatusOK, rdsOverviewError("failed to parse schema overview: "+err.Error()))
			return
		}
		out.Schemas = append(out.Schemas, s)
	}
	if err := rows.Err(); err != nil {
		writeJSON(w, http.StatusOK, rdsOverviewError("failed to read schema overview rows: "+err.Error()))
		return
	}

	// List all user-accessible databases on this instance, with size.
	dbRows, err := client.QueryContext(ctx, `
		SELECT datname,
		       pg_database_size(datname) AS size_bytes,
		       pg_size_pretty(pg_database_size(datname)) AS size_pretty
		FROM pg_database
		WHERE NOT datistemplate
		  AND datallowconn
		  AND datname NOT IN ('rdsadmin')
		ORDER BY pg_database_size(datname) DESC, datname ASC
	`)
	if err != nil {
		writeJSON(w, http.StatusOK, rdsOverviewError("failed to query databases: "+err.Error()))
		return
	}
	defer dbRows.Close()
	for dbRows.Next() {
		var d RDSDatabaseSummary
		if err := dbRows.Scan(&d.Name, &d.SizeBytes, &d.SizePretty); err != nil {
			writeJSON(w, http.StatusOK, rdsOverviewError("failed to parse database row: "+err.Error()))
			return
		}
		d.IsCurrent = d.Name == out.CurrentDB
		out.Databases = append(out.Databases, d)
	}
	if err := dbRows.Err(); err != nil {
		writeJSON(w, http.StatusOK, rdsOverviewError("failed to read database rows: "+err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, out)
}

func rdsOverviewError(msg string) RDSOverview {
	return RDSOverview{
		Connected:    false,
		ErrorMessage: msg,
		Schemas:      []RDSSchemaSummary{},
		Databases:    []RDSDatabaseSummary{},
	}
}

// HandleTables handles GET /api/rds/tables?db=<name>&schema=<name>.
//
// Returns the list of base/partitioned tables in the given schema with row
// count estimates and total size. Row counts come from pg_class.reltuples
// (updated by ANALYZE/autovacuum) so they are approximations, not exact.
func (h *RDSHandler) HandleTables(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if h.factory == nil {
		writeError(w, http.StatusServiceUnavailable, "PostgreSQL is not configured")
		return
	}

	schema := r.URL.Query().Get("schema")
	if schema == "" {
		writeError(w, http.StatusBadRequest, "schema parameter is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	requestedDB := r.URL.Query().Get("db")
	client, err := h.factory.Get(ctx, requestedDB)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to connect to database: "+err.Error())
		return
	}

	var currentDB string
	_ = client.QueryRowContext(ctx, "SELECT current_database()").Scan(&currentDB)

	rows, err := client.QueryContext(ctx, `
		SELECT c.relname,
		       c.reltuples::bigint AS row_estimate,
		       pg_total_relation_size(c.oid) AS size_bytes,
		       pg_size_pretty(pg_total_relation_size(c.oid)) AS size_pretty
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = $1
		  AND c.relkind IN ('r', 'p')
		ORDER BY pg_total_relation_size(c.oid) DESC, c.relname ASC
		LIMIT 200
	`, schema)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query tables: "+err.Error())
		return
	}
	defer rows.Close()

	out := RDSTablesResponse{
		Database: currentDB,
		Schema:   schema,
		Tables:   []RDSTableSummary{},
	}
	for rows.Next() {
		var t RDSTableSummary
		if err := rows.Scan(&t.Name, &t.RowEstimate, &t.SizeBytes, &t.SizePretty); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to parse table row: "+err.Error())
			return
		}
		out.Tables = append(out.Tables, t)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read table rows: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, out)
}
