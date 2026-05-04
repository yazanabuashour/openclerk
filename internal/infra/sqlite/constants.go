package sqlite

import "regexp"

const (
	configKeyVaultRoot                = "vault_root"
	configKeyLayoutConventionVersion  = "layout_convention_version"
	configKeyProjectionRebuildPending = "projection_rebuild_pending"
	configKeyFTSRebuildPending        = "fts_rebuild_pending"
	defaultLayoutConventionVersion    = "root_v1"
	rootSynthesisPathPrefix           = "synthesis/"
	maxSourceDownloadBytes            = 50 << 20
	sourceURLModeCreate               = "create"
	sourceURLModeUpdate               = "update"
	sourceTypePDF                     = "pdf"
	sourceTypeWeb                     = "web"
	evalSourceFixtureRootEnv          = "OPENCLERK_EVAL_SOURCE_FIXTURE_ROOT"
	evalSourceFixtureHost             = "openclerk-eval.local"
)

var (
	headingPattern = regexp.MustCompile(`^(#{1,6})\s+(.*?)\s*$`)
	linkPattern    = regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)
	wordPattern    = regexp.MustCompile(`[A-Za-z0-9_]+`)
)
