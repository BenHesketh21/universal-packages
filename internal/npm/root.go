package npm

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/tidwall/sjson"
)

func FindFirstTGZInDir(dir string) (string, error) {
	var tgzPath string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // abort on access error
		}
		if !info.IsDir() && filepath.Ext(path) == ".tgz" {
			tgzPath = path
			return filepath.SkipDir // stop once found
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("walking directory: %w", err)
	}
	if tgzPath == "" {
		return "", fmt.Errorf("no .tgz file found in directory %s", dir)
	}
	return tgzPath, nil
}

func GetPackageName(ociPath string) (string, error) {
	f, err := os.Open(ociPath)
	if err != nil {
		return "", fmt.Errorf("opening tarball: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalf("unable to close file: %v", err)
		}
	}()

	gzReader, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("gzip reader: %w", err)
	}
	defer func() {
		if err := gzReader.Close(); err != nil {
			log.Fatalf("unable to close file: %v", err)
		}
	}()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return "", fmt.Errorf("tar read error: %w", err)
		}

		// Looking for package/package.json inside the tarball
		if strings.HasSuffix(header.Name, "package.json") {
			var pkgJSON struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			}
			if err := json.NewDecoder(tarReader).Decode(&pkgJSON); err != nil {
				return "", fmt.Errorf("failed to decode package.json: %w", err)
			}

			return pkgJSON.Name, nil
		}
	}

	return "", fmt.Errorf("package.json not found in tarball")
}

func UpdatePackageJSONWithFileDep(pkgJSONPath, depName, filePath string) error {
	// Read entire file as bytes
	data, err := os.ReadFile(pkgJSONPath)
	if err != nil {
		return err
	}

	// Build the path string for sjson
	depPath := "dependencies." + depName

	// Add or overwrite dependency
	relPath, err := filepath.Rel(filepath.Dir(pkgJSONPath), filePath)
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
