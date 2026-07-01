package runclient

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
	_ "modernc.org/sqlite"
)

type RuntimeInspection struct {
	DatabaseSource           string
	DatabaseExists           bool
	DatabaseInitialized      bool
	VaultRoot                string
	DocumentCount            int
	ChunkCount               int
	DuplicateProjectionCount int
	Tables                   map[string]bool
	Projections              map[string]RuntimeProjectionInspection
	Synthesis                RuntimeSynthesisInspection
	Modules                  []SemanticModuleConfig
}

type RuntimeProjectionInspection struct {
	Total   int
	Fresh   int
	Stale   int
	Unknown int
}

type RuntimeSynthesisInspection struct {
	FreshCount               int
	StaleCount               int
	MissingSourceRefCount    int
	SupersededSourceRefCount int
}

func InspectExistingRuntime(ctx context.Context, cfg Config) (RuntimeInspection, error) {
	databasePath, databaseSource, err := resolveDatabasePathWithSource(cfg)
	if err != nil {
		return RuntimeInspection{}, err
	}
	databasePath = filepath.Clean(databasePath)
	inspection := RuntimeInspection{
		DatabaseSource: databaseSource,
		Tables:         map[string]bool{},
		Projections:    map[string]RuntimeProjectionInspection{},
	}
	info, err := os.Stat(databasePath)
	if err == nil {
		if info.IsDir() {
			return inspection, domain.ValidationError("database path must be a file", map[string]any{"database_path": databasePath})
		}
		inspection.DatabaseExists = true
	} else if errors.Is(err, os.ErrNotExist) {
		return inspection, nil
	} else {
		return inspection, domain.InternalError("inspect OpenClerk database", err)
	}

	db, err := sql.Open("sqlite", readOnlySQLiteDSN(databasePath))
	if err != nil {
		return inspection, domain.InternalError("open OpenClerk database read-only", err)
	}
	defer func() {
		_ = db.Close()
	}()
	for _, statement := range []string{
		`PRAGMA busy_timeout = 5000;`,
		`PRAGMA foreign_keys = ON;`,
	} {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return inspection, domain.InternalError("configure read-only SQLite inspection", err)
		}
	}

	for _, table := range []string{"runtime_config", "documents", "chunks", "projection_states"} {
		exists, err := inspectTableExists(ctx, db, table)
		if err != nil {
			return inspection, err
		}
		inspection.Tables[table] = exists
	}
	if !inspection.Tables["runtime_config"] {
		return inspection, nil
	}

	runtimeValues, err := inspectRuntimeConfigValues(ctx, db)
	if err != nil {
		return inspection, err
	}
	inspection.VaultRoot = filepath.Clean(strings.TrimSpace(runtimeValues["vault_root"]))
	inspection.DatabaseInitialized = inspection.VaultRoot != "" && inspection.VaultRoot != "."

	if inspection.Tables["documents"] {
		count, err := inspectTableCount(ctx, db, "documents")
		if err != nil {
			return inspection, err
		}
		inspection.DocumentCount = count
	}
	if inspection.Tables["chunks"] {
		count, err := inspectTableCount(ctx, db, "chunks")
		if err != nil {
			return inspection, err
		}
		inspection.ChunkCount = count
	}
	if inspection.Tables["projection_states"] {
		projections, synthesis, duplicateProjectionCount, err := inspectProjectionPosture(ctx, db)
		if err != nil {
			return inspection, err
		}
		inspection.Projections = projections
		inspection.Synthesis = synthesis
		inspection.DuplicateProjectionCount = duplicateProjectionCount
	}
	inspection.Modules = inspectConfiguredModulesFromValues(runtimeValues)
	return inspection, nil
}

func readOnlySQLiteDSN(path string) string {
	return (&url.URL{
		Scheme:   "file",
		Path:     path,
		RawQuery: "mode=ro&immutable=1",
	}).String()
}

func inspectTableExists(ctx context.Context, db *sql.DB, name string) (bool, error) {
	var tableName string
	err := db.QueryRowContext(ctx, `SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, name).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, domain.InternalError("inspect OpenClerk schema", err)
	}
	return tableName == name, nil
}

func inspectTableCount(ctx context.Context, db *sql.DB, table string) (int, error) {
	var count int
	if err := db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count); err != nil {
		return 0, domain.InternalError("count "+table, err)
	}
	return count, nil
}

func inspectRuntimeConfigValues(ctx context.Context, db *sql.DB) (map[string]string, error) {
	rows, err := db.QueryContext(ctx, `SELECT key_name, value_text FROM runtime_config`)
	if err != nil {
		return nil, domain.InternalError("read runtime config", err)
	}
	defer func() {
		_ = rows.Close()
	}()
	values := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, domain.InternalError("scan runtime config", err)
		}
		values[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, domain.InternalError("iterate runtime config", err)
	}
	return values, nil
}

func inspectProjectionPosture(ctx context.Context, db *sql.DB) (map[string]RuntimeProjectionInspection, RuntimeSynthesisInspection, int, error) {
	rows, err := db.QueryContext(ctx, `SELECT projection_name, freshness, details_json FROM projection_states`)
	if err != nil {
		return nil, RuntimeSynthesisInspection{}, 0, domain.InternalError("read projection states", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	projections := map[string]RuntimeProjectionInspection{}
	var synthesis RuntimeSynthesisInspection
	duplicateProjectionCount := 0
	for rows.Next() {
		var projectionName, freshness, detailsJSON string
		if err := rows.Scan(&projectionName, &freshness, &detailsJSON); err != nil {
			return nil, RuntimeSynthesisInspection{}, 0, domain.InternalError("scan projection state", err)
		}
		projection := projections[projectionName]
		projection.Total++
		switch freshness {
		case "fresh":
			projection.Fresh++
		case "stale":
			projection.Stale++
		default:
			projection.Unknown++
		}
		projections[projectionName] = projection
		if projectionName == "synthesis" {
			switch freshness {
			case "fresh":
				synthesis.FreshCount++
			case "stale":
				synthesis.StaleCount++
			}
		}
		var details map[string]string
		if err := json.Unmarshal([]byte(detailsJSON), &details); err == nil {
			if details["duplicate_ref_id"] != "" {
				duplicateProjectionCount++
			}
			if projectionName == "synthesis" {
				synthesis.MissingSourceRefCount += countCommaList(details["missing_source_refs"])
				synthesis.SupersededSourceRefCount += countCommaList(details["superseded_source_refs"])
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, RuntimeSynthesisInspection{}, 0, domain.InternalError("iterate projection states", err)
	}
	return projections, synthesis, duplicateProjectionCount, nil
}

func countCommaList(value string) int {
	if strings.TrimSpace(value) == "" {
		return 0
	}
	count := 0
	for _, part := range strings.Split(value, ",") {
		if strings.TrimSpace(part) != "" {
			count++
		}
	}
	return count
}

func inspectConfiguredModulesFromValues(values map[string]string) []SemanticModuleConfig {
	modules := []SemanticModuleConfig{}
	for _, provider := range []string{SemanticModuleProviderOllama, SemanticModuleProviderGemini} {
		config := semanticModuleConfigFromValues(ModuleKindEmbeddingProvider, provider, prefixedRuntimeValues(values, semanticModuleKey(provider, "")))
		if strings.TrimSpace(config.ModuleName) != "" {
			modules = append(modules, config)
		}
	}
	ocrConfig := semanticModuleConfigFromValues(ModuleKindOCRProvider, OCRModuleProviderTesseract, prefixedRuntimeValues(values, ocrModuleKey(OCRModuleProviderTesseract, "")))
	if strings.TrimSpace(ocrConfig.ModuleName) != "" {
		modules = append(modules, ocrConfig)
	}
	return modules
}

func prefixedRuntimeValues(values map[string]string, prefix string) map[string]string {
	result := map[string]string{}
	for key, value := range values {
		if strings.HasPrefix(key, prefix) {
			result[strings.TrimPrefix(key, prefix)] = value
		}
	}
	return result
}
