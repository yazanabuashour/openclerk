package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	maturityModeScaleLadder = "scale-ladder"
	maturityModeRealVault   = "real-vault"

	scaleTier10MB  = "10mb"
	scaleTier100MB = "100mb"
	scaleTier1GB   = "1gb"
)

type maturityConfig struct {
	Mode             string
	Tier             string
	TargetBytes      int64
	Seed             int64
	RunRoot          string
	ReportDir        string
	ReportName       string
	PrivateVaultRoot string
	QueryCSV         string
	Allow1GB         bool
	SkipReopen       bool
}

func parseMaturityConfig(args []string, stderr io.Writer) (maturityConfig, error) {
	if len(args) == 0 {
		return maturityConfig{}, errors.New("usage: ockp maturity scale-ladder|real-vault [flags]")
	}
	mode := strings.TrimSpace(args[0])
	switch mode {
	case maturityModeScaleLadder, maturityModeRealVault:
	default:
		return maturityConfig{}, fmt.Errorf("unsupported maturity mode %q", mode)
	}

	fs := flag.NewFlagSet("ockp maturity "+mode, flag.ContinueOnError)
	fs.SetOutput(stderr)
	config := maturityConfig{
		Mode:      mode,
		Tier:      scaleTier10MB,
		Seed:      53,
		ReportDir: filepath.Join("docs", "evals", "results"),
	}
	fs.StringVar(&config.Tier, "tier", config.Tier, "scale tier: 10mb, 100mb, or 1gb")
	fs.Int64Var(&config.TargetBytes, "target-bytes", 0, "override generated corpus bytes for test or calibration runs")
	fs.Int64Var(&config.Seed, "seed", config.Seed, "deterministic corpus seed")
	fs.StringVar(&config.RunRoot, "run-root", "", "directory for local generated/private maturity artifacts")
	fs.StringVar(&config.ReportDir, "report-dir", config.ReportDir, "directory for reduced reports")
	fs.StringVar(&config.ReportName, "report-name", "", "base filename for reduced reports, without extension")
	fs.StringVar(&config.PrivateVaultRoot, "vault-root", "", "private/local vault root for real-vault maturity reports")
	fs.StringVar(&config.QueryCSV, "query", "", "comma-separated read probes for real-vault reports")
	fs.BoolVar(&config.Allow1GB, "allow-1gb", false, "allow the 1 GB scale tier after smaller tiers justify it")
	fs.BoolVar(&config.SkipReopen, "skip-reopen", false, "skip the second sync/rebuild timing pass")
	if err := fs.Parse(args[1:]); err != nil {
		return maturityConfig{}, err
	}
	if fs.NArg() != 0 {
		return maturityConfig{}, fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	if config.RunRoot == "" {
		config.RunRoot = filepath.Join(os.TempDir(), fmt.Sprintf("openclerk-ockp-maturity-%d", time.Now().UnixNano()))
	}
	if config.Mode == maturityModeScaleLadder {
		config.Tier = strings.ToLower(strings.TrimSpace(config.Tier))
	}
	if strings.TrimSpace(config.ReportName) == "" {
		config.ReportName = defaultMaturityReportName(config)
	}
	if config.Mode == maturityModeScaleLadder {
		target, err := scaleTierBytes(config.Tier)
		if err != nil {
			return maturityConfig{}, err
		}
		if config.Tier == scaleTier1GB && !config.Allow1GB {
			return maturityConfig{}, errors.New("--tier 1gb requires --allow-1gb after 10 MB and 100 MB reports show the run is meaningful")
		}
		if config.TargetBytes == 0 {
			config.TargetBytes = target
		}
		if config.TargetBytes < 1024 {
			return maturityConfig{}, errors.New("--target-bytes must be at least 1024")
		}
	}
	if config.Mode == maturityModeRealVault && strings.TrimSpace(config.PrivateVaultRoot) == "" {
		return maturityConfig{}, errors.New("real-vault maturity reports require --vault-root")
	}
	return config, nil
}

func defaultMaturityReportName(config maturityConfig) string {
	if config.Mode == maturityModeScaleLadder {
		return "ockp-scale-ladder-" + strings.ToLower(config.Tier)
	}
	return "ockp-real-vault-dogfood"
}

func scaleTierBytes(tier string) (int64, error) {
	switch strings.ToLower(strings.TrimSpace(tier)) {
	case scaleTier10MB:
		return 10 * 1024 * 1024, nil
	case scaleTier100MB:
		return 100 * 1024 * 1024, nil
	case scaleTier1GB:
		return 1024 * 1024 * 1024, nil
	default:
		return 0, fmt.Errorf("--tier must be %q, %q, or %q", scaleTier10MB, scaleTier100MB, scaleTier1GB)
	}
}

func maturityQueries(config maturityConfig) []string {
	if strings.TrimSpace(config.QueryCSV) != "" {
		return splitCSV(config.QueryCSV)
	}
	if config.Mode == maturityModeScaleLadder {
		return []string{
			fmt.Sprintf("scale ladder authority marker seed %d", config.Seed),
			"scale ladder synthesis freshness marker",
			"scale ladder duplicate candidate marker",
		}
	}
	return []string{
		"source linked synthesis provenance freshness",
		"decision record source refs citations",
		"duplicate candidate stale source authority",
	}
}
