package runner_test

import (
	"context"
	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestResolvePathsUsesDatabaseAnchoredConfig(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "env-db", "openclerk.sqlite")
	t.Setenv("OPENCLERK_DATABASE_PATH", dbPath)

	result, err := runner.RunDocumentTask(context.Background(), runclient.Config{}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionResolvePaths,
	})
	if err != nil {
		t.Fatalf("resolve paths: %v", err)
	}
	if result.Paths == nil ||
		result.Paths.DatabasePath != dbPath ||
		result.Paths.VaultRoot != filepath.Join(filepath.Dir(dbPath), "vault") {
		t.Fatalf("paths = %+v", result.Paths)
	}

	boundVaultRoot := filepath.Join(t.TempDir(), "wiki")
	initialized, err := runclient.InitializePaths(runclient.Config{DatabasePath: dbPath}, boundVaultRoot)
	if err != nil {
		t.Fatalf("initialize paths: %v", err)
	}
	if initialized.VaultRoot != boundVaultRoot {
		t.Fatalf("initialized paths = %+v, want vault %q", initialized, boundVaultRoot)
	}

	explicit, err := runner.RunDocumentTask(context.Background(), runclient.Config{
		DatabasePath: filepath.Join(t.TempDir(), "explicit-db", "openclerk.sqlite"),
	}, runner.DocumentTaskRequest{Action: runner.DocumentTaskActionResolvePaths})
	if err != nil {
		t.Fatalf("resolve explicit paths: %v", err)
	}
	if explicit.Paths.DatabasePath == os.Getenv("OPENCLERK_DATABASE_PATH") ||
		explicit.Paths.VaultRoot == boundVaultRoot {
		t.Fatalf("explicit config did not take precedence: %+v", explicit.Paths)
	}

	again, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: dbPath}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionResolvePaths,
	})
	if err != nil {
		t.Fatalf("resolve persisted paths: %v", err)
	}
	if again.Paths == nil || again.Paths.VaultRoot != boundVaultRoot {
		t.Fatalf("persisted paths = %+v, want vault %q", again.Paths, boundVaultRoot)
	}
}

func TestStoredVaultRootSurvivesRunnerReplacementAndRetiredEnv(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	vaultRoot := filepath.Join(t.TempDir(), "wiki")
	initialized, err := runclient.InitializePaths(runclient.Config{DatabasePath: dbPath}, vaultRoot)
	if err != nil {
		t.Fatalf("initialize paths: %v", err)
	}
	if initialized.DatabasePath != dbPath || initialized.VaultRoot != vaultRoot {
		t.Fatalf("initialized paths = %+v, want db %q vault %q", initialized, dbPath, vaultRoot)
	}

	t.Setenv("OPENCLERK_DATABASE_PATH", dbPath)
	t.Setenv("XDG_DATA_HOME", filepath.Join(t.TempDir(), "changed-xdg"))
	t.Setenv("OPENCLERK_DATA_DIR", filepath.Join(t.TempDir(), "retired-data"))
	t.Setenv("OPENCLERK_VAULT_ROOT", filepath.Join(t.TempDir(), "retired-vault"))

	resolved, err := runner.RunDocumentTask(context.Background(), runclient.Config{}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionResolvePaths,
	})
	if err != nil {
		t.Fatalf("resolve paths after replacement: %v", err)
	}
	if resolved.Paths == nil ||
		resolved.Paths.DatabasePath != dbPath ||
		resolved.Paths.VaultRoot != vaultRoot {
		t.Fatalf("resolved paths = %+v, want db %q vault %q", resolved.Paths, dbPath, vaultRoot)
	}

	layout, err := runner.RunDocumentTask(context.Background(), runclient.Config{}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionInspectLayout,
	})
	if err != nil {
		t.Fatalf("inspect layout after replacement: %v", err)
	}
	if layout.Layout == nil ||
		layout.Layout.Paths.DatabasePath != dbPath ||
		layout.Layout.Paths.VaultRoot != vaultRoot {
		t.Fatalf("layout paths = %+v, want db %q vault %q", layout.Layout, dbPath, vaultRoot)
	}
}

func TestResolvePathsZeroConfigCreatesDefaultDatabaseAndVaultConfig(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", filepath.Join(t.TempDir(), "xdg"))
	t.Setenv("OPENCLERK_DATABASE_PATH", "")
	t.Setenv("OPENCLERK_DATA_DIR", filepath.Join(t.TempDir(), "retired-data"))
	t.Setenv("OPENCLERK_VAULT_ROOT", filepath.Join(t.TempDir(), "retired-vault"))

	result, err := runner.RunDocumentTask(context.Background(), runclient.Config{}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionResolvePaths,
	})
	if err != nil {
		t.Fatalf("resolve paths: %v", err)
	}
	wantDB := filepath.Join(os.Getenv("XDG_DATA_HOME"), "openclerk", "openclerk.sqlite")
	wantVault := filepath.Join(filepath.Dir(wantDB), "vault")
	if result.Paths == nil ||
		result.Paths.DatabasePath != wantDB ||
		result.Paths.VaultRoot != wantVault {
		t.Fatalf("paths = %+v, want db %q vault %q", result.Paths, wantDB, wantVault)
	}
	if _, err := os.Stat(wantDB); err != nil {
		t.Fatalf("default database was not created: %v", err)
	}
	if _, err := os.Stat(wantVault); err != nil {
		t.Fatalf("default vault was not created: %v", err)
	}
}

func TestParallelFreshStartupReadActions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	runConcurrent(t, 24, func(i int) error {
		switch i % 3 {
		case 0:
			_, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{Action: runner.DocumentTaskActionResolvePaths})
			return err
		case 1:
			_, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
				Action: runner.DocumentTaskActionList,
				List:   runner.DocumentListOptions{PathPrefix: "notes/", Limit: 10},
			})
			return err
		default:
			_, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
				Action: runner.RetrievalTaskActionSearch,
				Search: runner.SearchOptions{
					Text:  "runner",
					Limit: 10,
				},
			})
			return err
		}
	})
}

func TestReadOnlyActionsDoNotTakeRunnerWriteLock(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	if _, err := runclient.InitializePaths(config, ""); err != nil {
		t.Fatalf("initialize paths: %v", err)
	}
	lockPath := config.DatabasePath + ".runner-write.lock"
	if err := os.WriteFile(lockPath, []byte("pid=1\n"), 0o644); err != nil {
		t.Fatalf("write fake runner lock: %v", err)
	}
	defer func() {
		_ = os.Remove(lockPath)
	}()

	if _, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:  "runner",
			Limit: 10,
		},
	}); err != nil {
		t.Fatalf("read-only retrieval should not take runner write lock: %v", err)
	}
	if _, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: "notes/", Limit: 10},
	}); err != nil {
		t.Fatalf("read-only document task should not take runner write lock: %v", err)
	}
}

func TestParallelReadWorkflowsAfterSeed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	doc := createDocument(t, ctx, config, "notes/projects/concurrency.md", "Concurrency", "# Concurrency\n\n## Summary\nParallel runner safe read evidence.\n")
	createDocument(t, ctx, config, "records/services/parallel-runner.md", "Parallel runner service", "---\nservice_id: parallel-runner\nservice_name: Parallel runner\nservice_status: active\nservice_owner: runner\nservice_interface: JSON runner\n---\n# Parallel runner\n\n## Summary\nParallel read service evidence.\n")
	createDocument(t, ctx, config, "docs/architecture/parallel-runner-decision.md", "Parallel runner decision", "---\ndecision_id: adr-parallel-runner\ndecision_title: Parallel runner reads\ndecision_status: accepted\ndecision_scope: runner\ndecision_owner: platform\ndecision_date: 2026-04-29\n---\n# Parallel runner decision\n\n## Summary\nAllow safe parallel read commands.\n")

	runConcurrent(t, 30, func(i int) error {
		switch i % 6 {
		case 0:
			_, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
				Action: runner.DocumentTaskActionList,
				List:   runner.DocumentListOptions{PathPrefix: "notes/", Limit: 10},
			})
			return err
		case 1:
			_, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
				Action: runner.DocumentTaskActionGet,
				DocID:  doc.DocID,
			})
			return err
		case 2:
			_, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
				Action: runner.RetrievalTaskActionSearch,
				Search: runner.SearchOptions{
					Text:  "parallel runner",
					Limit: 10,
				},
			})
			return err
		case 3:
			_, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
				Action: runner.RetrievalTaskActionServicesLookup,
				Services: runner.ServiceLookupOptions{
					Text:  "Parallel runner",
					Limit: 10,
				},
			})
			return err
		case 4:
			_, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
				Action: runner.RetrievalTaskActionDecisionsLookup,
				Decisions: runner.DecisionLookupOptions{
					Text:  "Parallel runner",
					Limit: 10,
				},
			})
			return err
		default:
			_, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
				Action: runner.RetrievalTaskActionProvenanceEvents,
				Provenance: runner.ProvenanceEventOptions{
					RefKind: "document",
					RefID:   doc.DocID,
					Limit:   10,
				},
			})
			return err
		}
	})
}

func TestReadWriteOverlapAndConflictingWrites(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	doc := createDocument(t, ctx, config, "notes/projects/overlap.md", "Overlap", "# Overlap\n\n## Summary\nRead write overlap evidence.\n")

	runConcurrent(t, 24, func(i int) error {
		if i == 0 {
			_, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
				Action:  runner.DocumentTaskActionAppend,
				DocID:   doc.DocID,
				Content: "## Update\nSerialized write completed.\n",
			})
			return err
		}
		if i%2 == 0 {
			_, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
				Action: runner.DocumentTaskActionList,
				List:   runner.DocumentListOptions{PathPrefix: "notes/", Limit: 10},
			})
			return err
		}
		_, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
			Action: runner.RetrievalTaskActionSearch,
			Search: runner.SearchOptions{
				Text:  "overlap evidence",
				Limit: 10,
			},
		})
		return err
	})

	var successes int
	var conflicts int
	var mu sync.Mutex
	runConcurrent(t, 8, func(i int) error {
		_, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
			Action: runner.DocumentTaskActionCreate,
			Document: runner.DocumentInput{
				Path:  "notes/projects/conflicting-create.md",
				Title: "Conflicting create",
				Body:  "# Conflicting create\n\n## Summary\nOnly one write should win.\n",
			},
		})
		mu.Lock()
		defer mu.Unlock()
		if err == nil {
			successes++
			return nil
		}
		if isClearConcurrencyConflict(err) {
			conflicts++
			return nil
		}
		return err
	})
	if successes != 1 || conflicts != 7 {
		t.Fatalf("conflicting creates: successes=%d conflicts=%d, want 1/7", successes, conflicts)
	}
}

func TestMutatingDocumentTaskSyncsVaultBeforeWrite(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	create := createDocument(t, ctx, config, "notes/projects/manual-edit.md", "Manual edit", "# Manual edit\n\n## Summary\nOriginal registry body.\n")
	paths, err := runclient.ResolvePaths(config)
	if err != nil {
		t.Fatalf("resolve paths: %v", err)
	}
	manualBody := "# Manual edit\n\n## Summary\nEdited outside the runner before append.\n"
	if err := os.WriteFile(filepath.Join(paths.VaultRoot, "notes", "projects", "manual-edit.md"), []byte(manualBody), 0o644); err != nil {
		t.Fatalf("write manual edit: %v", err)
	}

	appendResult, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action:  runner.DocumentTaskActionAppend,
		DocID:   create.DocID,
		Content: "## Follow-up\nRunner append preserved manual edits.\n",
	})
	if err != nil {
		t.Fatalf("append document task: %v", err)
	}
	if appendResult.Document == nil ||
		!strings.Contains(appendResult.Document.Body, "Edited outside the runner before append.") ||
		!strings.Contains(appendResult.Document.Body, "Runner append preserved manual edits.") {
		t.Fatalf("append result body = %q", appendResult.Document.Body)
	}
}

func runConcurrent(t *testing.T, count int, fn func(int) error) {
	t.Helper()

	var wg sync.WaitGroup
	errs := make(chan error, count)
	for i := 0; i < count; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := fn(i); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "sqlite") ||
				strings.Contains(strings.ToLower(err.Error()), "runtime config") ||
				strings.Contains(strings.ToLower(err.Error()), "upsert") {
				t.Fatalf("raw concurrency failure leaked: %v", err)
			}
			t.Fatalf("concurrent task failed: %v", err)
		}
	}
}
