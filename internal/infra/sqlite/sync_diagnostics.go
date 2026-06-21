package sqlite

import (
	"encoding/json"
	"math"
	"path/filepath"
	"time"

	"github.com/yazanabuashour/openclerk/internal/domain"
)

const maxSyncDiagnosticsIgnoredPaths = 100

type SyncDiagnostics struct {
	Status                     string                         `json:"status"`
	LastPhase                  string                         `json:"last_phase"`
	Mode                       string                         `json:"mode"`
	PathsScanned               int                            `json:"paths_scanned"`
	PathsIgnored               int                            `json:"paths_ignored"`
	IgnoredDirectories         int                            `json:"ignored_directories"`
	IgnoredFiles               int                            `json:"ignored_files"`
	IgnoredPathRules           []string                       `json:"ignored_path_rules,omitempty"`
	IgnoredPaths               []string                       `json:"ignored_paths,omitempty"`
	IgnoredPathsTruncated      bool                           `json:"ignored_paths_truncated,omitempty"`
	DocumentsCreated           int                            `json:"documents_created"`
	DocumentsUpdated           int                            `json:"documents_updated"`
	DocumentsUnchanged         int                            `json:"documents_unchanged"`
	DocumentsPruned            int                            `json:"documents_pruned"`
	BytesRead                  int64                          `json:"bytes_read"`
	ChunksWritten              int                            `json:"chunks_written"`
	FTSRowsWritten             int                            `json:"fts_rows_written"`
	FTSStrategy                string                         `json:"fts_strategy"`
	FTSBootstrap               bool                           `json:"fts_bootstrap"`
	FTSRebuildPending          bool                           `json:"fts_rebuild_pending"`
	FTSRebuildSkipped          bool                           `json:"fts_rebuild_skipped"`
	ProjectionBootstrap        bool                           `json:"projection_bootstrap"`
	ProjectionRebuildSkipped   bool                           `json:"projection_rebuild_skipped"`
	ScanSeconds                float64                        `json:"scan_seconds"`
	PruneSeconds               float64                        `json:"prune_seconds"`
	DocumentReadParseSeconds   float64                        `json:"document_read_parse_seconds"`
	DocumentWriteSeconds       float64                        `json:"document_write_seconds"`
	DocumentRecordWriteSeconds float64                        `json:"document_record_write_seconds"`
	ChunkWriteSeconds          float64                        `json:"chunk_write_seconds"`
	ProvenanceWriteSeconds     float64                        `json:"provenance_write_seconds"`
	IncrementalFTSWriteSeconds float64                        `json:"incremental_fts_write_seconds"`
	BulkFTSRebuildSeconds      float64                        `json:"bulk_fts_rebuild_seconds"`
	ProjectionRebuildSeconds   float64                        `json:"projection_rebuild_seconds"`
	TotalSeconds               float64                        `json:"total_seconds"`
	ProjectionRebuilds         []ProjectionRebuildDiagnostics `json:"projection_rebuilds,omitempty"`
	ReducedReportSafe          bool                           `json:"reduced_report_safe"`
	EvidencePosture            string                         `json:"evidence_posture"`
}

type ProjectionRebuildDiagnostics struct {
	Projection string  `json:"projection"`
	Seconds    float64 `json:"seconds"`
}

func newSyncDiagnostics() SyncDiagnostics {
	return SyncDiagnostics{
		Status:            "running",
		LastPhase:         "starting",
		Mode:              "vault_sync",
		FTSStrategy:       "pending",
		ReducedReportSafe: true,
		EvidencePosture:   "reduced counters and timings plus bounded ignored vault-relative paths only; excludes titles, snippets, raw content, database paths, vault roots, and machine-absolute paths",
	}
}

func (d *SyncDiagnostics) recordIgnoredPath(relPath string, directory bool) {
	if d == nil {
		return
	}
	d.PathsIgnored++
	if directory {
		d.IgnoredDirectories++
	} else {
		d.IgnoredFiles++
	}
	if len(d.IgnoredPaths) < maxSyncDiagnosticsIgnoredPaths {
		d.IgnoredPaths = append(d.IgnoredPaths, relPath)
		return
	}
	d.IgnoredPathsTruncated = true
}

func (d *SyncDiagnostics) changedDocuments() int {
	if d == nil {
		return 0
	}
	return d.DocumentsCreated + d.DocumentsUpdated + d.DocumentsPruned
}

func (s *Store) LatestSyncDiagnostics() (SyncDiagnostics, bool) {
	if s == nil {
		return SyncDiagnostics{}, false
	}
	if s.lastSyncDiagnostics == nil {
		return SyncDiagnostics{
			IgnoredPathRules: append([]string(nil), s.vaultIgnorePaths...),
		}, false
	}
	diagnostics := *s.lastSyncDiagnostics
	diagnostics.IgnoredPathRules = append([]string(nil), diagnostics.IgnoredPathRules...)
	diagnostics.IgnoredPaths = append([]string(nil), diagnostics.IgnoredPaths...)
	diagnostics.ProjectionRebuilds = append([]ProjectionRebuildDiagnostics(nil), diagnostics.ProjectionRebuilds...)
	return diagnostics, true
}

func writeSyncDiagnostics(path string, diagnostics SyncDiagnostics) error {
	if path == "" {
		return nil
	}
	if err := osMkdirAll(filepath.Dir(path), 0o755); err != nil {
		return domain.InternalError("create sync diagnostics directory", err)
	}
	diagnostics = roundedSyncDiagnostics(diagnostics)
	content, err := json.MarshalIndent(diagnostics, "", "  ")
	if err != nil {
		return domain.InternalError("encode sync diagnostics", err)
	}
	content = append(content, '\n')
	if err := osWriteBytes(path, content); err != nil {
		return domain.InternalError("write sync diagnostics", err)
	}
	return nil
}

func syncSecondsSince(start time.Time) float64 {
	return time.Since(start).Seconds()
}

func roundedSyncDiagnostics(diagnostics SyncDiagnostics) SyncDiagnostics {
	diagnostics.ScanSeconds = roundSyncSeconds(diagnostics.ScanSeconds)
	diagnostics.PruneSeconds = roundSyncSeconds(diagnostics.PruneSeconds)
	diagnostics.DocumentReadParseSeconds = roundSyncSeconds(diagnostics.DocumentReadParseSeconds)
	diagnostics.DocumentWriteSeconds = roundSyncSeconds(diagnostics.DocumentWriteSeconds)
	diagnostics.DocumentRecordWriteSeconds = roundSyncSeconds(diagnostics.DocumentRecordWriteSeconds)
	diagnostics.ChunkWriteSeconds = roundSyncSeconds(diagnostics.ChunkWriteSeconds)
	diagnostics.ProvenanceWriteSeconds = roundSyncSeconds(diagnostics.ProvenanceWriteSeconds)
	diagnostics.IncrementalFTSWriteSeconds = roundSyncSeconds(diagnostics.IncrementalFTSWriteSeconds)
	diagnostics.BulkFTSRebuildSeconds = roundSyncSeconds(diagnostics.BulkFTSRebuildSeconds)
	diagnostics.ProjectionRebuildSeconds = roundSyncSeconds(diagnostics.ProjectionRebuildSeconds)
	diagnostics.TotalSeconds = roundSyncSeconds(diagnostics.TotalSeconds)
	for index := range diagnostics.ProjectionRebuilds {
		diagnostics.ProjectionRebuilds[index].Seconds = roundSyncSeconds(diagnostics.ProjectionRebuilds[index].Seconds)
	}
	return diagnostics
}

func roundSyncSeconds(seconds float64) float64 {
	return math.Round(seconds*100) / 100
}
