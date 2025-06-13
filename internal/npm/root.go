package npm

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
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
	defer f.Close()

	gzReader, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("gzip reader: %w", err)
	}
	defer gzReader.Close()

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

// func UpdatePackageJSONWithFileDep(pkgJSONPath, depName, filePath string) error {
// 	// Read existing package.json
// 	f, err := os.ReadFile(pkgJSONPath)
// 	if err != nil {
// 		return fmt.Errorf("reading package.json: %w", err)
// 	}

// 	// Unmarshal to map first
// 	var data map[string]any
// 	if err := json.Unmarshal(f, &data); err != nil {
// 		return err
// 	}

// 	// Marshal again to put known fields into struct
// 	knownFields, err := json.Marshal(data)
// 	if err != nil {
// 		return err
// 	}

// 	var pkg PackageJSON
// 	if err := json.Unmarshal(knownFields, &pkg); err != nil {
// 		return err
// 	}

// 	// Save extra unknown fields
// 	pkg.Extras = make(map[string]any)
// 	for k, v := range data {
// 		if !isKnownPackageJSONField(k) {
// 			pkg.Extras[k] = v
// 		}
// 	}

// 	log.Printf("Processing field: %s", pkg.Extras)

// 	// Ensure dependencies map exists
// 	if pkg.Dependencies == nil {
// 		pkg.Dependencies = make(map[string]string)
// 	}

// 	// Add or overwrite dependency
// 	relPath, err := filepath.Rel(filepath.Dir(pkgJSONPath), filePath)
// 	if err != nil {
// 		return fmt.Errorf("calculating relative path: %w", err)
// 	}
// 	pkg.Dependencies[depName] = "file:" + filepath.ToSlash(relPath)

// 	if err := SavePackageJSON(pkgJSONPath, &pkg); err != nil {
// 		return fmt.Errorf("writing updated package.json: %w", err)
// 	}

// 	return nil
// }

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

func isKnownPackageJSONField(key string) bool {
	switch key {
	case "name", "version", "description", "main",
		"scripts", "dependencies", "devDependencies", "peerDependencies",
		"keywords", "author", "license":
		return true
	default:
		return false
	}
}

type PackageJSON struct {
	Name             string            `json:"name,omitempty"`
	Version          string            `json:"version,omitempty"`
	Description      string            `json:"description,omitempty"`
	License          string            `json:"license,omitempty"`
	Main             string            `json:"main,omitempty"`
	Scripts          map[string]string `json:"scripts,omitempty"`
	Dependencies     map[string]string `json:"dependencies,omitempty"`
	DevDependencies  map[string]string `json:"devDependencies,omitempty"`
	PeerDependencies map[string]string `json:"peerDependencies,omitempty"`
	Keywords         []string          `json:"keywords,omitempty"`
	Author           any               `json:"author,omitempty"` // can be string or object

	// This is to preserve unknown fields if needed
	Extras map[string]any `json:"-"`
}

func SavePackageJSON(path string, pkg *PackageJSON) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")  // pretty print with 2-space indent
	encoder.SetEscapeHTML(false) // disable escaping of &, <, >

	log.Printf("%s", pkg)

	return encoder.Encode(pkg)
}
