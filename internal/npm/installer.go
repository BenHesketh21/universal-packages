package npm

import (
	"fmt"
)

func InstallTarball(tarballPath string) error {
	// Optionally: Validate that it's a valid NPM package (has package.json, etc.)

	// For now, just untar into ./node_modules/<package-name> (future improvement)
	fmt.Printf("Installing %s...\n", tarballPath)
	// TODO: parse package.json and name correctly

	return nil
}
