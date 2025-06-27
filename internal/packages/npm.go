package packages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type NpmHandler struct{}

func (n *NpmHandler) LocateArtefact(dir string, packageName string, packageVersion string) (string, error) {
	// Normalize package name: "@scope/pkg" → "scope-pkg"
	normalized := strings.ReplaceAll(strings.TrimPrefix(packageName, "@"), "/", "-")

	filename := fmt.Sprintf("%s-%s.tgz", normalized, packageVersion)
	packagePath := filepath.Join(dir, filename)

	// Check if file exists
	if _, err := os.Stat(packagePath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("expected package file not found: %s", filename)
		}
		return "", fmt.Errorf("error checking package file: %w", err)
	}
	return packagePath, nil
}
