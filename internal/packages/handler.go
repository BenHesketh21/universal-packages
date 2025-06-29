package packages

import (
	"fmt"
)

type PackageHandler interface {
	// LocatePackage finds the package file in the specified directory
	// based on the package name and version.
	// Returns the file path if found, or an error if not.
	LocatePackage(dir string, packageName string, packageVersion string) (string, error)
	// UpdatePackageRef updates the package reference in the project's
	// package file (e.g., package.json for npm) to point to the local file
	UpdatePackageRef(packageName string, packageFilePath string, packageRefFilePath string) error
}

// Registry of supported handlers by name
var handlers = map[string]PackageHandler{
	"npm": &NpmHandler{},
	// "pypi": &PyPiHandler{},
	// Add more here
}

func GetHandler(packageType string) (PackageHandler, error) {
	if h, ok := handlers[packageType]; ok {
		return h, nil
	}
	return nil, fmt.Errorf("unsupported package type: %s", packageType)
}
