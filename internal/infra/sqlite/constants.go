package sqlite

import "regexp"

const (
	configKeyVaultRoot                = "vault_root"
	configKeyVaultIgnorePaths         = "vault_ignore_paths"
	configKeyLayoutConventionVersion  = "layout_convention_version"
	configKeyProjectionRebuildPending = "projection_rebuild_pending"
	configKeyFTSRebuildPending        = "fts_rebuild_pending"
	defaultLayoutConventionVersion    = "root_v1"
	rootSynthesisPathPrefix           = "synthesis/"
	maxSourceDownloadBytes            = 50 << 20
	maxVaultMarkdownFiles             = 10000
	maxVaultMarkdownDocumentBytes     = 5 << 20
	maxVaultMarkdownSections          = 1000
	maxVaultMarkdownMetadataFields    = 256
	maxLexicalFallbackCandidateRows   = 1000
	sourceURLModeCreate               = "create"
	sourceURLModeUpdate               = "update"
	sourceTypePDF                     = "pdf"
	sourceTypeWeb                     = "web"
	evalSourceFixtureRootEnv          = "OPENCLERK_EVAL_SOURCE_FIXTURE_ROOT"
	evalSourceFixtureEnableEnv        = "OPENCLERK_ENABLE_EVAL_SOURCE_FIXTURES"
	evalSourceFixtureHost             = "openclerk-eval.local"
)

var (
	headingPattern = regexp.MustCompile(`^(#{1,6})\s+(.*?)\s*$`)
	linkPattern    = regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)
	wordPattern    = regexp.MustCompile(`[A-Za-z0-9_]+`)
)
