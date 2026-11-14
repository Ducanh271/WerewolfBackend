package mysql

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrations embed.FS

// migration is one numbered file of migrations/.
type migration struct {
	version int
	name    string
	stmts   []string
}

// loadMigrations reads migrations/NNN_name.sql in order. A file holds plain
// statements separated by ";" at the end of a line (no procedures).
func loadMigrations() ([]migration, error) {
	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return nil, err
	}
	var out []migration
	for _, e := range entries {
		num, _, ok := strings.Cut(e.Name(), "_")
		v, err := strconv.Atoi(num)
		if !ok || err != nil || v <= 0 {
			return nil, fmt.Errorf("mysql: migration %q is not named NNN_name.sql", e.Name())
		}
		body, err := migrations.ReadFile("migrations/" + e.Name())
		if err != nil {
			return nil, err
		}
		out = append(out, migration{version: v, name: e.Name(), stmts: splitStatements(string(body))})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].version < out[j].version })
	for i := 1; i < len(out); i++ {
		if out[i].version == out[i-1].version {
			return nil, fmt.Errorf("mysql: two migrations numbered %d", out[i].version)
		}
	}
	return out, nil
}

func splitStatements(body string) []string {
	var stmts []string
	var cur strings.Builder
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		cur.WriteString(line)
		cur.WriteString("\n")
		if strings.HasSuffix(trimmed, ";") {
			stmts = append(stmts, strings.TrimSuffix(strings.TrimSpace(cur.String()), ";"))
			cur.Reset()
		}
	}
	if s := strings.TrimSpace(cur.String()); s != "" {
		stmts = append(stmts, s)
	}
	return stmts
}

// migrate applies every migration newer than the database. MySQL commits DDL
// on its own, so a migration that fails half way is reported and the server
// does not start; fix it by hand (docs/database.md).
func migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_version (
		version    INT UNSIGNED NOT NULL PRIMARY KEY,
		name       VARCHAR(128) NOT NULL,
		applied_at DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`); err != nil {
		return fmt.Errorf("mysql: schema_version: %w", err)
	}
	var have int
	if err := db.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_version`).Scan(&have); err != nil {
		return fmt.Errorf("mysql: read schema_version: %w", err)
	}
	all, err := loadMigrations()
	if err != nil {
		return err
	}
	for _, m := range all {
		if m.version <= have {
			continue
		}
		for _, stmt := range m.stmts {
			if _, err := db.ExecContext(ctx, stmt); err != nil {
				return fmt.Errorf("mysql: migration %s: %w", m.name, err)
			}
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO schema_version (version, name) VALUES (?, ?)`, m.version, m.name); err != nil {
			return fmt.Errorf("mysql: record migration %s: %w", m.name, err)
		}
	}
	return nil
}
