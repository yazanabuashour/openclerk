package sqlite

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
)

var (
	vaultWriteNewFile = func(store *Store, relPath string, data []byte, label string) error {
		return store.writeNewVaultFileDirect(relPath, data, label)
	}
	vaultWriteExistingFile = func(store *Store, relPath string, data []byte, label string) error {
		return store.writeExistingVaultFileDirect(relPath, data, label)
	}
)

func (s *Store) vaultAbsPath(relPath string) string {
	return filepath.Join(s.vaultRoot, filepath.FromSlash(relPath))
}

func (s *Store) vaultCreateAbsPath(relPath string, label string) (string, error) {
	if err := s.validateVaultPathNotIgnored(relPath); err != nil {
		return "", err
	}
	absPath := s.vaultAbsPath(relPath)
	if err := validateVaultCreatePath(s.vaultRoot, absPath, label); err != nil {
		return "", err
	}
	return absPath, nil
}

func (s *Store) vaultExistingAbsPath(relPath string, label string) (string, error) {
	if err := s.validateVaultPathNotIgnored(relPath); err != nil {
		return "", err
	}
	absPath := s.vaultAbsPath(relPath)
	if err := validateExistingVaultFile(s.vaultRoot, absPath, label); err != nil {
		return "", err
	}
	return absPath, nil
}

func (s *Store) validateVaultPathNotIgnored(relPath string) error {
	if s == nil || !s.vaultIgnoreMatcher.Matches(relPath) {
		return nil
	}
	return domain.ValidationError("document path is ignored by vault sync configuration", map[string]any{
		"path": relPath,
	})
}

func validateVaultCreatePath(vaultRoot string, absPath string, label string) error {
	rootAbs, err := filepath.Abs(vaultRoot)
	if err != nil {
		return domain.InternalError(label, err)
	}
	rootReal, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return domain.InternalError(label, err)
	}
	candidateAbs, err := filepath.Abs(absPath)
	if err != nil {
		return domain.InternalError(label, err)
	}
	rel, err := filepath.Rel(rootAbs, candidateAbs)
	if err != nil {
		return domain.InternalError(label, err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return domain.ValidationError("document path must stay inside the vault root", nil)
	}
	parentRel := filepath.Dir(rel)
	if parentRel == "." {
		return nil
	}
	current := rootReal
	for _, component := range strings.Split(parentRel, string(filepath.Separator)) {
		if component == "" || component == "." {
			continue
		}
		current = filepath.Join(current, component)
		info, err := osLstat(current)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			return nil
		case err != nil:
			return domain.InternalError(label, err)
		case info.Mode()&fs.ModeSymlink != 0:
			return domain.ValidationError("document path must stay inside the vault root", nil)
		case !info.IsDir():
			return domain.ValidationError("document path parent must be a directory", nil)
		}
	}
	return nil
}

func (s *Store) openVaultRoot(label string) (*os.Root, error) {
	root, err := os.OpenRoot(s.vaultRoot)
	if err != nil {
		return nil, domain.InternalError(label, err)
	}
	return root, nil
}

func vaultRootName(relPath string) string {
	if relPath == "" || relPath == "." {
		return "."
	}
	return filepath.FromSlash(relPath)
}

func vaultRootOpError(label string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, fs.ErrExist) || errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if strings.Contains(strings.ToLower(err.Error()), "escapes") {
		return domain.ValidationError("vault path must stay inside the vault root", nil)
	}
	return domain.InternalError(label, err)
}

func (s *Store) lstatVaultPath(relPath string, label string) (fs.FileInfo, error) {
	root, err := s.openVaultRoot(label)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = root.Close()
	}()
	info, err := root.Lstat(vaultRootName(relPath))
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (s *Store) statVaultPath(relPath string, label string) (fs.FileInfo, error) {
	root, err := s.openVaultRoot(label)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = root.Close()
	}()
	info, err := root.Stat(vaultRootName(relPath))
	if err != nil {
		return nil, vaultRootOpError(label, err)
	}
	return info, nil
}

func (s *Store) readVaultFile(relPath string, label string) ([]byte, error) {
	root, err := s.openVaultRoot(label)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = root.Close()
	}()
	body, err := root.ReadFile(vaultRootName(relPath))
	if err != nil {
		return nil, vaultRootOpError(label, err)
	}
	return body, nil
}

func (s *Store) readVaultFileLimited(relPath string, maxBytes int64, label string) ([]byte, error) {
	info, err := s.statVaultPath(relPath, label)
	if err != nil {
		return nil, err
	}
	if info.Size() > maxBytes {
		return nil, domain.ValidationError("source asset exceeds maximum supported size", map[string]any{"max_bytes": maxBytes})
	}
	body, err := s.readVaultFile(relPath, label)
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxBytes {
		return nil, domain.ValidationError("source asset exceeds maximum supported size", map[string]any{"max_bytes": maxBytes})
	}
	return body, nil
}

func (s *Store) writeNewVaultFile(relPath string, data []byte, label string) error {
	return vaultWriteNewFile(s, relPath, data, label)
}

func (s *Store) writeNewVaultFileDirect(relPath string, data []byte, label string) error {
	root, err := s.openVaultRoot(label)
	if err != nil {
		return err
	}
	defer func() {
		_ = root.Close()
	}()
	name := vaultRootName(relPath)
	if parent := filepath.Dir(name); parent != "." {
		if err := root.MkdirAll(parent, 0o700); err != nil {
			return vaultRootOpError(label, err)
		}
	}
	file, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return vaultRootOpError(label, err)
	}
	writeErr := writeAll(file, data)
	closeErr := file.Close()
	if writeErr != nil {
		return domain.InternalError(label, writeErr)
	}
	if closeErr != nil {
		return domain.InternalError(label, closeErr)
	}
	return nil
}

func (s *Store) writeExistingVaultFile(relPath string, data []byte, label string) error {
	return vaultWriteExistingFile(s, relPath, data, label)
}

func (s *Store) writeExistingVaultFileDirect(relPath string, data []byte, label string) error {
	root, err := s.openVaultRoot(label)
	if err != nil {
		return err
	}
	defer func() {
		_ = root.Close()
	}()
	file, err := root.OpenFile(vaultRootName(relPath), os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return vaultRootOpError(label, err)
	}
	writeErr := writeAll(file, data)
	closeErr := file.Close()
	if writeErr != nil {
		return domain.InternalError(label, writeErr)
	}
	if closeErr != nil {
		return domain.InternalError(label, closeErr)
	}
	return nil
}

func writeAll(file *os.File, data []byte) error {
	for len(data) > 0 {
		n, err := file.Write(data)
		if err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
}

func (s *Store) removeVaultPath(relPath string, label string) error {
	root, err := s.openVaultRoot(label)
	if err != nil {
		return err
	}
	defer func() {
		_ = root.Close()
	}()
	if err := root.Remove(vaultRootName(relPath)); err != nil {
		return vaultRootOpError(label, err)
	}
	return nil
}

func (s *Store) renameVaultPath(oldRelPath string, newRelPath string, label string) error {
	root, err := s.openVaultRoot(label)
	if err != nil {
		return err
	}
	defer func() {
		_ = root.Close()
	}()
	newName := vaultRootName(newRelPath)
	if parent := filepath.Dir(newName); parent != "." {
		if err := root.MkdirAll(parent, 0o700); err != nil {
			return vaultRootOpError(label, err)
		}
	}
	if err := root.Rename(vaultRootName(oldRelPath), newName); err != nil {
		return vaultRootOpError(label, err)
	}
	return nil
}
