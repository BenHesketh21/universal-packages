package packages

import (
	"fmt"
)

type PackageHandler interface {
	LocateArtefact(dir string, packageName string, packageVersion string) (string, error)
}

// Registry of supported handlers by name
var handlers = map[string]PackageHandler{
	"npm": &NpmHandler{},
	// "pypi": &PyPiHandler{},
	// Add more here
}

func GetHandler(lang string) (PackageHandler, error) {
	if h, ok := handlers[lang]; ok {
		return h, nil
	}
	return nil, fmt.Errorf("unsupported language: %s", lang)
}
