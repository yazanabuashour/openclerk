package sqlite

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
)

func (s *Store) vaultAbsPath(relPath string) string {
	return filepath.Join(s.vaultRoot, filepath.FromSlash(relPath))
}

func (s *Store) vaultCreateAbsPath(relPath string, label string) (string, error) {
	absPath := s.vaultAbsPath(relPath)
	if err := validateVaultCreatePath(s.vaultRoot, absPath, label); err != nil {
		return "", err
	}
	return absPath, nil
}

func (s *Store) vaultExistingAbsPath(relPath string, label string) (string, error) {
	absPath := s.vaultAbsPath(relPath)
	if err := validateExistingVaultFile(s.vaultRoot, absPath, label); err != nil {
		return "", err
	}
	return absPath, nil
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
