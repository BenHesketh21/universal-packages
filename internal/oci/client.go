package oci

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
)

func ConnectToRegistry(ref string) (*remote.Repository, error) {
	parts := strings.SplitN(ref, "/", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid OCI reference: %s", ref)
	}

	repo, err := remote.NewRepository(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to create remote repository: %w", err)
	}

	// This enables credential helpers (Docker config, GitHub token, etc.)
	repo.PlainHTTP = false // Set true for local or insecure registries

	return repo, nil
}

func PullPackage(ctx context.Context, repo *remote.Repository, ref string, upRootDir string) (string, error) {

	repositoryName := repo.Reference.Repository

	workingDir := filepath.Join(upRootDir, repositoryName)
	dst, err := file.New(workingDir)
	if err != nil {
		panic(err)
	}

	_, err = oras.Copy(ctx, repo, ref, dst, "", oras.DefaultCopyOptions)
	if err != nil {
		return "", fmt.Errorf("oras pull failed: %w", err)
	}

	return workingDir, nil
}
