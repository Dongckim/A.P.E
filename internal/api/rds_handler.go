package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

type RDSColumnInfo struct {
	Name         string  `json:"name"`
	DataType     string  `json:"data_type"`
	IsNullable   bool    `json:"is_nullable"`
	DefaultValue *string `json:"default_value"`
	IsPrimaryKey bool    `json:"is_primary_key"`
	Position     int     `json:"position"`
}

type RDSTableDetail struct {
	Database    string          `json:"database"`
	Schema      string          `json:"schema"`
	Table       string          `json:"table"`
	Columns     []RDSColumnInfo `json:"columns"`
	SampleRows  [][]any         `json:"sample_rows"`
	SampleLimit int             `json:"sample_limit"`
	RowEstimate int64           `json:"row_estimate"`
}

type RDSERDColumn struct {
	Name         string `json:"name"`
	DataType     string `json:"data_type"`
	IsPrimaryKey bool   `json:"is_primary_key"`
	IsForeignKey bool   `json:"is_foreign_key"`
	IsNullable   bool   `json:"is_nullable"`
}

type RDSERDTable struct {
	Name    string         `json:"name"`
	Columns []RDSERDColumn `json:"columns"`
}

type RDSERDEdge struct {
	ConstraintName string `json:"constraint_name"`
	FromTable      string `json:"from_table"`
	FromColumn     string `json:"from_column"`
	ToTable        string `json:"to_table"`
	ToColumn       string `json:"to_column"`
}

type RDSERDResponse struct {
	Database   string        `json:"database"`
	Schema     string        `json:"schema"`
	Tables     []RDSERDTable `json:"tables"`
	Edges      []RDSERDEdge  `json:"edges"`
	Truncated  bool          `json:"truncated"`
	TableLimit int           `json:"table_limit"`
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

// quoteIdent safely quotes a PostgreSQL identifier (schema, table, or column
// name) for inclusion in dynamic SQL. PostgreSQL placeholder parameters
// cannot be used for identifiers, so we wrap the value in double quotes and
// escape any embedded double quotes.
func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

// HandleTableDetail handles GET /api/rds/table?db=&schema=&table=&sample_limit=N.
//
// Returns the column metadata (name, type, nullable, default, primary key)
// and a small sample of rows for the given table. The sample query uses
// quoted identifiers (placeholder parameters cannot be used for identifiers
// in PostgreSQL), and the limit is capped server-side at 100 rows.
func (h *RDSHandler) HandleTableDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if h.factory == nil {
		writeError(w, http.StatusServiceUnavailable, "PostgreSQL is not configured")
		return
	}

	q := r.URL.Query()
	schema := q.Get("schema")
	table := q.Get("table")
	if schema == "" || table == "" {
		writeError(w, http.StatusBadRequest, "schema and table parameters are required")
		return
	}
	sampleLimit := 10
	if v := q.Get("sample_limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			sampleLimit = n
		}
	}
	if sampleLimit > 100 {
		sampleLimit = 100
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	requestedDB := q.Get("db")
	client, err := h.factory.Get(ctx, requestedDB)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to connect to database: "+err.Error())
		return
	}

	out := RDSTableDetail{
		Schema:      schema,
		Table:       table,
		Columns:     []RDSColumnInfo{},
		SampleRows:  [][]any{},
		SampleLimit: sampleLimit,
	}
	_ = client.QueryRowContext(ctx, "SELECT current_database()").Scan(&out.Database)

	// Verify the table exists in this DB and grab its row estimate. This also
	// guards against quoting attempts on non-existent identifiers.
	qualified := schema + "." + table
	if err := client.QueryRowContext(ctx, `
		SELECT COALESCE(c.reltuples::bigint, 0)
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = $1 AND c.relname = $2
		  AND c.relkind IN ('r', 'p', 'v', 'm')
	`, schema, table).Scan(&out.RowEstimate); err != nil {
		writeError(w, http.StatusNotFound, "table not found: "+qualified)
		return
	}

	// Primary key columns (set lookup).
	pkCols := map[string]bool{}
	pkRows, err := client.QueryContext(ctx, `
		SELECT a.attname
		FROM pg_index i
		JOIN pg_class c ON c.oid = i.indrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace
		JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum = ANY(i.indkey)
		WHERE n.nspname = $1 AND c.relname = $2 AND i.indisprimary
	`, schema, table)
	if err == nil {
		for pkRows.Next() {
			var name string
			if err := pkRows.Scan(&name); err == nil {
				pkCols[name] = true
			}
		}
		pkRows.Close()
	}

	// Column metadata.
	colRows, err := client.QueryContext(ctx, `
		SELECT column_name,
		       data_type,
		       is_nullable,
		       column_default,
		       ordinal_position
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position
	`, schema, table)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query columns: "+err.Error())
		return
	}
	defer colRows.Close()
	for colRows.Next() {
		var (
			c          RDSColumnInfo
			isNullable string
			defVal     *string
		)
		if err := colRows.Scan(&c.Name, &c.DataType, &isNullable, &defVal, &c.Position); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to parse column row: "+err.Error())
			return
		}
		c.IsNullable = strings.EqualFold(isNullable, "YES")
		c.DefaultValue = defVal
		c.IsPrimaryKey = pkCols[c.Name]
		out.Columns = append(out.Columns, c)
	}
	if err := colRows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read column rows: "+err.Error())
		return
	}

	// Sample rows. Identifiers must be quoted (placeholders are for values).
	sampleSQL := fmt.Sprintf("SELECT * FROM %s.%s LIMIT $1", quoteIdent(schema), quoteIdent(table))
	sRows, err := client.QueryContext(ctx, sampleSQL, sampleLimit)
	if err != nil {
		// Don't fail the whole request if sample fails (e.g. permission denied);
		// return columns and an empty sample with a hint in the column list.
		writeJSON(w, http.StatusOK, out)
		return
	}
	defer sRows.Close()

	colNames, err := sRows.Columns()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read sample columns: "+err.Error())
		return
	}
	for sRows.Next() {
		vals := make([]any, len(colNames))
		ptrs := make([]any, len(colNames))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := sRows.Scan(ptrs...); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to scan sample row: "+err.Error())
			return
		}
		// Convert []byte to string so JSON encoding doesn't base64 it.
		for i, v := range vals {
			if b, ok := v.([]byte); ok {
				vals[i] = string(b)
			}
		}
		out.SampleRows = append(out.SampleRows, vals)
	}
	if err := sRows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read sample rows: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, out)
}

// HandleERD handles GET /api/rds/erd?db=<name>&schema=<name>&limit=<n>.
//
// Returns the data needed to render an Entity Relationship Diagram for a
// schema: each table's columns (with PK/FK marks) and the foreign key edges
// between them. The number of tables is capped server-side to keep the
// rendered diagram readable.
func (h *RDSHandler) HandleERD(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if h.factory == nil {
		writeError(w, http.StatusServiceUnavailable, "PostgreSQL is not configured")
		return
	}

	q := r.URL.Query()
	schema := q.Get("schema")
	if schema == "" {
		writeError(w, http.StatusBadRequest, "schema parameter is required")
		return
	}
	tableLimit := 50
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			tableLimit = n
		}
	}
	if tableLimit > 200 {
		tableLimit = 200
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	requestedDB := q.Get("db")
	client, err := h.factory.Get(ctx, requestedDB)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to connect to database: "+err.Error())
		return
	}

	out := RDSERDResponse{
		Schema:     schema,
		Tables:     []RDSERDTable{},
		Edges:      []RDSERDEdge{},
		TableLimit: tableLimit,
	}
	_ = client.QueryRowContext(ctx, "SELECT current_database()").Scan(&out.Database)

	// Step 1: list tables in this schema (LIMIT applied) + a count to know if truncated.
	var totalTables int
	if err := client.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = $1 AND c.relkind IN ('r', 'p')
	`, schema).Scan(&totalTables); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to count tables: "+err.Error())
		return
	}
	out.Truncated = totalTables > tableLimit

	tableRows, err := client.QueryContext(ctx, `
		SELECT c.relname
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = $1 AND c.relkind IN ('r', 'p')
		ORDER BY c.relname
		LIMIT $2
	`, schema, tableLimit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tables: "+err.Error())
		return
	}
	tableSet := map[string]bool{}
	tableOrder := []string{}
	for tableRows.Next() {
		var name string
		if err := tableRows.Scan(&name); err != nil {
			tableRows.Close()
			writeError(w, http.StatusInternalServerError, "failed to scan table name: "+err.Error())
			return
		}
		tableSet[name] = true
		tableOrder = append(tableOrder, name)
	}
	tableRows.Close()

	// Step 2: pull all columns for the schema, with PK/FK marks via LEFT JOIN.
	// Filter to the truncated set in Go.
	colRows, err := client.QueryContext(ctx, `
		SELECT c.table_name,
		       c.column_name,
		       c.data_type,
		       c.is_nullable,
		       COALESCE(pk.is_pk, false) AS is_pk,
		       COALESCE(fk.is_fk, false) AS is_fk,
		       c.ordinal_position
		FROM information_schema.columns c
		LEFT JOIN (
		    SELECT kcu.table_name, kcu.column_name, true AS is_pk
		    FROM information_schema.table_constraints tc
		    JOIN information_schema.key_column_usage kcu
		      ON tc.constraint_name = kcu.constraint_name
		     AND tc.table_schema = kcu.table_schema
		    WHERE tc.constraint_type = 'PRIMARY KEY'
		      AND tc.table_schema = $1
		) pk ON pk.table_name = c.table_name AND pk.column_name = c.column_name
		LEFT JOIN (
		    SELECT kcu.table_name, kcu.column_name, true AS is_fk
		    FROM information_schema.table_constraints tc
		    JOIN information_schema.key_column_usage kcu
		      ON tc.constraint_name = kcu.constraint_name
		     AND tc.table_schema = kcu.table_schema
		    WHERE tc.constraint_type = 'FOREIGN KEY'
		      AND tc.table_schema = $1
		) fk ON fk.table_name = c.table_name AND fk.column_name = c.column_name
		WHERE c.table_schema = $1
		ORDER BY c.table_name, c.ordinal_position
	`, schema)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query columns: "+err.Error())
		return
	}
	defer colRows.Close()

	tablesByName := map[string]*RDSERDTable{}
	for _, name := range tableOrder {
		tbl := &RDSERDTable{Name: name, Columns: []RDSERDColumn{}}
		tablesByName[name] = tbl
	}
	for colRows.Next() {
		var (
			tableName  string
			c          RDSERDColumn
			isNullable string
			ordinal    int
		)
		if err := colRows.Scan(&tableName, &c.Name, &c.DataType, &isNullable, &c.IsPrimaryKey, &c.IsForeignKey, &ordinal); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to parse column row: "+err.Error())
			return
		}
		if !tableSet[tableName] {
			continue
		}
		c.IsNullable = strings.EqualFold(isNullable, "YES")
		tablesByName[tableName].Columns = append(tablesByName[tableName].Columns, c)
	}
	if err := colRows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read column rows: "+err.Error())
		return
	}
	for _, name := range tableOrder {
		out.Tables = append(out.Tables, *tablesByName[name])
	}

	// Step 3: foreign-key edges. Filter to edges where both endpoints are in
	// the truncated table set.
	edgeRows, err := client.QueryContext(ctx, `
		SELECT tc.constraint_name,
		       tc.table_name      AS from_table,
		       kcu.column_name    AS from_column,
		       ccu.table_name     AS to_table,
		       ccu.column_name    AS to_column
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
		  ON tc.constraint_name = kcu.constraint_name
		 AND tc.table_schema = kcu.table_schema
		JOIN information_schema.constraint_column_usage ccu
		  ON ccu.constraint_name = tc.constraint_name
		 AND ccu.table_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY'
		  AND tc.table_schema = $1
		ORDER BY tc.table_name, tc.constraint_name
	`, schema)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query foreign keys: "+err.Error())
		return
	}
	defer edgeRows.Close()
	for edgeRows.Next() {
		var e RDSERDEdge
		if err := edgeRows.Scan(&e.ConstraintName, &e.FromTable, &e.FromColumn, &e.ToTable, &e.ToColumn); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to parse edge row: "+err.Error())
			return
		}
		if !tableSet[e.FromTable] || !tableSet[e.ToTable] {
			continue
		}
		out.Edges = append(out.Edges, e)
	}
	if err := edgeRows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read edge rows: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, out)
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
