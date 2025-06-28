package packages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tidwall/sjson"
)

type NpmHandler struct{}

// LocatePackage finds the package tarball in the specified directory based on the package name and version.
func (n *NpmHandler) LocatePackage(dir string, packageName string, packageVersion string) (string, error) {
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

// UpdatePackageRef updates the package reference in the project's package.json to point to the local file path.
// It adds or updates the dependency entry for the specified package.
// It should maintain existing format of package.json, including any existing dependencies.
func (n *NpmHandler) UpdatePackageRef(packageName string, packageFilePath string) error {
	// Read entire file as bytes
	pkgJSONPath, err := FindPackageJSON(".")
	if err != nil {
		return fmt.Errorf("finding package.json: %w", err)
	}
	data, err := os.ReadFile(pkgJSONPath)
	if err != nil {
		return err
	}

	// Build the path string for sjson
	depPath := "dependencies." + packageName

	// Add or overwrite dependency
	relPath, err := filepath.Rel(filepath.Dir(pkgJSONPath), packageFilePath)
	if err != nil {
		return fmt.Errorf("calculating relative path: %w", err)
	}

	// Update dependency in the JSON bytes
	updatedData, err := sjson.SetBytes(data, depPath, "file:"+filepath.ToSlash(relPath))
	if err != nil {
		return err
	}

	// Write back to file
	return os.WriteFile(pkgJSONPath, updatedData, 0644)
}

// FindPackageJSON searches for the nearest package.json file starting from the given directory and moving up the directory tree.
func FindPackageJSON(workingDir string) (string, error) {
	for {
		p := filepath.Join(workingDir, "package.json")
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}

		parent := filepath.Dir(workingDir)
		if parent == workingDir {
			break // root reached
		}
		workingDir = parent
	}
	return "", fmt.Errorf("package.json not found")
}
