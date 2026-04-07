package api

import (
	"context"
	"net/http"
	"time"

	"github.com/dongchankim/ape/internal/postgres"
)

type RDSHandler struct {
	client postgres.Client
}

func NewRDSHandler(client postgres.Client) *RDSHandler {
	return &RDSHandler{client: client}
}

type RDSOverview struct {
	Version      string             `json:"version"`
	CurrentDB    string             `json:"current_db"`
	SchemaCount  int                `json:"schema_count"`
	TableCount   int                `json:"table_count"`
	Schemas      []RDSSchemaSummary `json:"schemas"`
	Connected    bool               `json:"connected"`
	ErrorMessage string             `json:"error,omitempty"`
}

type RDSSchemaSummary struct {
	Name       string `json:"name"`
	TableCount int    `json:"table_count"`
}

// HandleOverview handles GET /api/rds/overview.
func (h *RDSHandler) HandleOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if h.client == nil {
		writeJSON(w, http.StatusOK, rdsOverviewError("PostgreSQL is not configured. Set APE_PG_DSN and restart A.P.E."))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	out := RDSOverview{Connected: true, Schemas: []RDSSchemaSummary{}}

	if err := h.client.QueryRowContext(ctx, "SELECT version()").Scan(&out.Version); err != nil {
		writeJSON(w, http.StatusOK, rdsOverviewError("failed to read postgres version: "+err.Error()))
		return
	}
	if err := h.client.QueryRowContext(ctx, "SELECT current_database()").Scan(&out.CurrentDB); err != nil {
		writeJSON(w, http.StatusOK, rdsOverviewError("failed to read current database: "+err.Error()))
		return
	}
	if err := h.client.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.schemata").Scan(&out.SchemaCount); err != nil {
		writeJSON(w, http.StatusOK, rdsOverviewError("failed to count schemas: "+err.Error()))
		return
	}
	if err := h.client.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_type='BASE TABLE'").Scan(&out.TableCount); err != nil {
		writeJSON(w, http.StatusOK, rdsOverviewError("failed to count tables: "+err.Error()))
		return
	}

	rows, err := h.client.QueryContext(ctx, `
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

	writeJSON(w, http.StatusOK, out)
}

func rdsOverviewError(msg string) RDSOverview {
	return RDSOverview{
		Connected:    false,
		ErrorMessage: msg,
		Schemas:      []RDSSchemaSummary{},
	}
}
