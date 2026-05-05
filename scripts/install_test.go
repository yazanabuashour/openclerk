package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const testInstallVersion = "v0.2.4"

func TestInstallDefaultsToHomeLocalBinWhenPathContainsWritableTempDir(t *testing.T) {
	toolsDir := writeInstallTestTools(t)
	home := t.TempDir()
	codexTmp := filepath.Join(t.TempDir(), ".codex", "tmp", "arg0", "codex-arg08w8fRj")
	if err := os.MkdirAll(codexTmp, 0o755); err != nil {
		t.Fatalf("mkdir codex temp path: %v", err)
	}

	result := runInstallScript(t, installScriptEnv{
		home: home,
		path: joinPath(toolsDir, codexTmp, systemPath()),
	})

	wantPath := filepath.Join(home, ".local", "bin", "openclerk")
	assertFileExists(t, wantPath)
	assertFileMissing(t, filepath.Join(codexTmp, "openclerk"))
	assertContains(t, result, "Installed openclerk runner to "+wantPath)
	assertContains(t, result, "Add this directory to PATH before using the skill:")
}

func TestInstallDirOverrideWins(t *testing.T) {
	toolsDir := writeInstallTestTools(t)
	home := t.TempDir()
	installDir := filepath.Join(t.TempDir(), "custom-bin")

	result := runInstallScript(t, installScriptEnv{
		home:       home,
		path:       joinPath(toolsDir, systemPath()),
		installDir: installDir,
	})

	wantPath := filepath.Join(installDir, "openclerk")
	assertFileExists(t, wantPath)
	assertContains(t, result, "Installed openclerk runner to "+wantPath)
}

func TestInstallUpgradesExistingDurablePathBinary(t *testing.T) {
	toolsDir := writeInstallTestTools(t)
	home := t.TempDir()
	scratch := repoLocalScratchDir(t)
	durableBin := filepath.Join(scratch, "bin")
	if err := os.MkdirAll(durableBin, 0o755); err != nil {
		t.Fatalf("mkdir durable bin: %v", err)
	}
	writeExecutable(t, filepath.Join(durableBin, "openclerk"), "#!/bin/sh\necho old-openclerk\n")

	result := runInstallScript(t, installScriptEnv{
		home: home,
		path: joinPath(toolsDir, durableBin, systemPath()),
	})

	wantPath := filepath.Join(durableBin, "openclerk")
	assertFileExists(t, wantPath)
	assertContains(t, result, "Installed openclerk runner to "+wantPath)
	assertContains(t, result, "Runner version: openclerk "+testInstallVersion)
}

func TestInstallSkipsExistingEphemeralPathBinary(t *testing.T) {
	toolsDir := writeInstallTestTools(t)
	home := t.TempDir()
	ephemeralBin := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(ephemeralBin, 0o755); err != nil {
		t.Fatalf("mkdir ephemeral bin: %v", err)
	}
	writeExecutable(t, filepath.Join(ephemeralBin, "openclerk"), "#!/bin/sh\necho old-temp-openclerk\n")

	result := runInstallScript(t, installScriptEnv{
		home: home,
		path: joinPath(toolsDir, ephemeralBin, systemPath()),
	})

	wantPath := filepath.Join(home, ".local", "bin", "openclerk")
	assertFileExists(t, wantPath)
	assertContains(t, result, "Installed openclerk runner to "+wantPath)
}

func TestInstallSkipsExistingPrivateVarFoldersPathBinary(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("/private/var/folders is a macOS temp path")
	}

	toolsDir := writeInstallTestTools(t)
	home := t.TempDir()
	tempDir := t.TempDir()
	if !strings.HasPrefix(tempDir, "/var/folders/") {
		t.Skipf("temp dir is not under /var/folders: %s", tempDir)
	}
	ephemeralBin := filepath.Join("/private", tempDir, "bin")
	if err := os.MkdirAll(ephemeralBin, 0o755); err != nil {
		t.Fatalf("mkdir private var folders bin: %v", err)
	}
	writeExecutable(t, filepath.Join(ephemeralBin, "openclerk"), "#!/bin/sh\necho old-temp-openclerk\n")

	result := runInstallScript(t, installScriptEnv{
		home: home,
		path: joinPath(toolsDir, ephemeralBin, systemPath()),
	})

	wantPath := filepath.Join(home, ".local", "bin", "openclerk")
	assertFileExists(t, wantPath)
	assertContains(t, result, "Installed openclerk runner to "+wantPath)
}

func TestInstallModuleInstallsEmbeddingPackageAndPrintsRegistration(t *testing.T) {
	toolsDir := writeModuleInstallTestTools(t, "ollama-embeddings", true)
	home := t.TempDir()
	installDir := filepath.Join(t.TempDir(), "bin")
	moduleDir := filepath.Join(t.TempDir(), "modules")

	result := runInstallModuleScript(t, moduleInstallScriptEnv{
		home:       home,
		path:       joinPath(toolsDir, systemPath()),
		installDir: installDir,
		moduleDir:  moduleDir,
		module:     "ollama-embeddings",
		version:    "latest",
	})

	assertFileExists(t, filepath.Join(installDir, "semantic-retrieval-adapter"))
	assertFileExists(t, filepath.Join(moduleDir, "ollama-embeddings", "module.json"))
	assertFileExists(t, filepath.Join(moduleDir, "ollama-embeddings", "skill", "ollama-embeddings", "SKILL.md"))
	assertContains(t, result, "Installed module files to "+filepath.Join(moduleDir, "ollama-embeddings"))
	assertContains(t, result, `"manifest_path":"`+filepath.Join(moduleDir, "ollama-embeddings", "module.json")+`"`)
	assertContains(t, result, `"command":"`+filepath.Join(installDir, "semantic-retrieval-adapter")+`"`)
	assertContains(t, result, `"provider":"ollama"`)
}

func TestInstallModuleDoesNotBundleExternalTesseractBinary(t *testing.T) {
	toolsDir := writeModuleInstallTestTools(t, "tesseract-ocr", false)
	home := t.TempDir()
	installDir := filepath.Join(t.TempDir(), "bin")
	moduleDir := filepath.Join(t.TempDir(), "modules")

	result := runInstallModuleScript(t, moduleInstallScriptEnv{
		home:       home,
		path:       joinPath(toolsDir, systemPath()),
		installDir: installDir,
		moduleDir:  moduleDir,
		module:     "tesseract-ocr",
		version:    "v0.1.1",
	})

	assertFileMissing(t, filepath.Join(installDir, "tesseract"))
	assertFileExists(t, filepath.Join(moduleDir, "tesseract-ocr", "module.json"))
	assertContains(t, result, `"kind":"ocr_provider"`)
	assertContains(t, result, `"command":"tesseract"`)
}

func TestBuildModuleReleaseBundlePackagesAdapterAndTesseractSeparately(t *testing.T) {
	toolsDir := writeModuleBuildTestTools(t)
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	outDir := filepath.Join(t.TempDir(), "dist")

	cmd := exec.Command("bash", "./scripts/build-module-release-bundle.sh", "ollama-embeddings", "v0.1.1", outDir)
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "PATH="+joinPath(toolsDir, systemPath()))
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build ollama module bundle: %v\n%s", err, output)
	}
	ollamaArchive := filepath.Join(outDir, "openclerk-module-ollama-embeddings_0.1.1_darwin_arm64.tar.gz")
	assertFileExists(t, ollamaArchive)
	assertFileExists(t, filepath.Join(outDir, "openclerk-module-ollama-embeddings_0.1.1_skill.tar.gz"))
	installerPath := filepath.Join(outDir, "install-module.sh")
	assertFileExists(t, installerPath)
	installer, err := os.ReadFile(installerPath)
	if err != nil {
		t.Fatalf("read generated installer: %v", err)
	}
	assertContains(t, string(installer), `default_version="v0.1.1"`)
	ollamaContents := tarList(t, ollamaArchive)
	assertContains(t, ollamaContents, "bin/semantic-retrieval-adapter")
	assertContains(t, ollamaContents, "modules/ollama-embeddings/module.json")
	assertContains(t, ollamaContents, "modules/ollama-embeddings/skill/ollama-embeddings/SKILL.md")

	outDir = filepath.Join(t.TempDir(), "dist")
	cmd = exec.Command("bash", "./scripts/build-module-release-bundle.sh", "tesseract-ocr", "v0.1.1", outDir)
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "PATH="+joinPath(toolsDir, systemPath()))
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build tesseract module bundle: %v\n%s", err, output)
	}
	tesseractArchive := filepath.Join(outDir, "openclerk-module-tesseract-ocr_0.1.1_darwin_arm64.tar.gz")
	assertFileExists(t, tesseractArchive)
	tesseractContents := tarList(t, tesseractArchive)
	if strings.Contains(tesseractContents, "/bin/") {
		t.Fatalf("tesseract module package should not bundle OCR binaries:\n%s", tesseractContents)
	}
	assertContains(t, tesseractContents, "modules/tesseract-ocr/module.json")
}

type installScriptEnv struct {
	home       string
	path       string
	installDir string
}

func runInstallScript(t *testing.T, env installScriptEnv) string {
	t.Helper()

	scriptPath, err := filepath.Abs("install.sh")
	if err != nil {
		t.Fatalf("resolve install.sh: %v", err)
	}
	cmd := exec.Command("sh", scriptPath)
	cmd.Env = append(os.Environ(),
		"HOME="+env.home,
		"PATH="+env.path,
		"OPENCLERK_VERSION="+testInstallVersion,
		"OPENCLERK_INSTALL_DIR="+env.installDir,
		"TMPDIR="+t.TempDir(),
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("install.sh failed: %v\n%s", err, output)
	}
	return string(output)
}

type moduleInstallScriptEnv struct {
	home       string
	path       string
	installDir string
	moduleDir  string
	module     string
	version    string
}

func runInstallModuleScript(t *testing.T, env moduleInstallScriptEnv) string {
	t.Helper()

	scriptPath, err := filepath.Abs("install-module.sh")
	if err != nil {
		t.Fatalf("resolve install-module.sh: %v", err)
	}
	cmd := exec.Command("sh", scriptPath)
	cmd.Env = append(os.Environ(),
		"HOME="+env.home,
		"PATH="+env.path,
		"OPENCLERK_MODULE="+env.module,
		"OPENCLERK_MODULE_VERSION="+env.version,
		"OPENCLERK_INSTALL_DIR="+env.installDir,
		"OPENCLERK_MODULE_DIR="+env.moduleDir,
		"TMPDIR="+t.TempDir(),
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("install-module.sh failed: %v\n%s", err, output)
	}
	return string(output)
}

func writeInstallTestTools(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	writeExecutable(t, filepath.Join(dir, "curl"), `#!/bin/sh
output=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "-o" ]; then
    shift
    output="$1"
  fi
  shift || true
done
if [ -z "$output" ]; then
  printf '%s\n' '{"tag_name":"v0.2.4"}'
  exit 0
fi
case "$output" in
  *checksums.txt)
    {
      printf '%s  %s\n' mock openclerk_0.2.4_darwin_amd64.tar.gz
      printf '%s  %s\n' mock openclerk_0.2.4_darwin_arm64.tar.gz
      printf '%s  %s\n' mock openclerk_0.2.4_linux_amd64.tar.gz
      printf '%s  %s\n' mock openclerk_0.2.4_linux_arm64.tar.gz
    } > "$output"
    ;;
  *)
    : > "$output"
    ;;
esac
`)
	writeExecutable(t, filepath.Join(dir, "tar"), `#!/bin/sh
archive=""
for arg in "$@"; do
  case "$arg" in
    *.tar.gz) archive="$arg" ;;
  esac
done
[ -n "$archive" ] || exit 1
asset_dir="${archive%.tar.gz}"
mkdir -p "$asset_dir"
cat > "$asset_dir/openclerk" <<'EOF'
#!/bin/sh
case "$1" in
  --version) printf '%s\n' 'openclerk v0.2.4' ;;
  --help) printf '%s\n' 'OpenClerk help' ;;
  *) printf '%s\n' 'OpenClerk mock' ;;
esac
EOF
chmod 755 "$asset_dir/openclerk"
`)
	writeExecutable(t, filepath.Join(dir, "shasum"), "#!/bin/sh\nexit 0\n")
	return dir
}

func writeModuleInstallTestTools(t *testing.T, module string, withBinary bool) string {
	t.Helper()

	dir := t.TempDir()
	writeExecutable(t, filepath.Join(dir, "curl"), `#!/bin/sh
output=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "-o" ]; then
    shift
    output="$1"
  fi
  shift || true
done
if [ -z "$output" ]; then
  printf '%s\n' '[{"tag_name":"`+module+`/v0.1.1"}]'
  exit 0
fi
case "$output" in
  *checksums.txt)
    {
      printf '%s  %s\n' mock "openclerk-module-`+module+`_0.1.1_darwin_amd64.tar.gz"
      printf '%s  %s\n' mock "openclerk-module-`+module+`_0.1.1_darwin_arm64.tar.gz"
      printf '%s  %s\n' mock "openclerk-module-`+module+`_0.1.1_linux_amd64.tar.gz"
      printf '%s  %s\n' mock "openclerk-module-`+module+`_0.1.1_linux_arm64.tar.gz"
    } > "$output"
    ;;
  *)
    : > "$output"
    ;;
esac
`)
	binaryBlock := ""
	if withBinary {
		binaryBlock = `
mkdir -p "$asset_dir/bin"
cat > "$asset_dir/bin/semantic-retrieval-adapter" <<'EOF'
#!/bin/sh
printf '%s\n' semantic-retrieval-adapter
EOF
chmod 755 "$asset_dir/bin/semantic-retrieval-adapter"
`
	}
	writeExecutable(t, filepath.Join(dir, "tar"), `#!/bin/sh
archive=""
for arg in "$@"; do
  case "$arg" in
    *.tar.gz) archive="$arg" ;;
  esac
done
[ -n "$archive" ] || exit 1
asset_dir="${archive%.tar.gz}"
mkdir -p "$asset_dir/modules/`+module+`/skill/`+module+`"
printf '%s\n' '{"schema_version":"openclerk-module.v1"}' > "$asset_dir/modules/`+module+`/module.json"
printf '%s\n' 'module skill' > "$asset_dir/modules/`+module+`/skill/`+module+`/SKILL.md"
`+binaryBlock+`
`)
	writeExecutable(t, filepath.Join(dir, "shasum"), "#!/bin/sh\nexit 0\n")
	return dir
}

func writeModuleBuildTestTools(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	writeExecutable(t, filepath.Join(dir, "go"), `#!/bin/sh
case "$1" in
  build)
    output=""
    while [ "$#" -gt 0 ]; do
      if [ "$1" = "-o" ]; then
        shift
        output="$1"
      fi
      shift || true
    done
    [ -n "$output" ] || exit 1
    mkdir -p "$(dirname "$output")"
    printf '%s\n' '#!/bin/sh' 'printf "%s\n" semantic-retrieval-adapter' > "$output"
    chmod 755 "$output"
    ;;
  run)
    output=""
    while [ "$#" -gt 0 ]; do
      if [ "$1" = "-output" ]; then
        shift
        output="$1"
      fi
      shift || true
    done
    [ -n "$output" ] || exit 1
    printf '%s\n' '{"bomFormat":"CycloneDX"}' > "$output"
    ;;
  *)
    exit 1
    ;;
esac
`)
	return dir
}

func tarList(t *testing.T, archive string) string {
	t.Helper()

	output, err := exec.Command("tar", "-tzf", archive).CombinedOutput()
	if err != nil {
		t.Fatalf("list tar archive %s: %v\n%s", archive, err, output)
	}
	return string(output)
}

func writeExecutable(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write executable %s: %v", path, err)
	}
}

func repoLocalScratchDir(t *testing.T) string {
	t.Helper()

	dir, err := os.MkdirTemp(".", ".install-test-")
	if err != nil {
		t.Fatalf("mkdir repo-local scratch dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("remove repo-local scratch dir: %v", err)
		}
	})
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatalf("abs repo-local scratch dir: %v", err)
	}
	return abs
}

func joinPath(parts ...string) string {
	var clean []string
	for _, part := range parts {
		if part != "" {
			clean = append(clean, part)
		}
	}
	return strings.Join(clean, string(os.PathListSeparator))
}

func systemPath() string {
	if runtime.GOOS == "darwin" {
		return "/usr/bin:/bin:/usr/sbin:/sbin"
	}
	return "/usr/bin:/bin"
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}

func assertFileMissing(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be missing, stat err=%v", path, err)
	}
}

func assertContains(t *testing.T, output string, want string) {
	t.Helper()

	if !strings.Contains(output, want) {
		t.Fatalf("output missing %q:\n%s", want, output)
	}
}
