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

func PullTarball(ctx context.Context, repo *remote.Repository, ref string, destDir string) (string, error) {
	parts := strings.Split(ref, "/")
	if len(parts) < 1 {
		return "", fmt.Errorf("invalid reference: %s", ref)
	}

	last := parts[len(parts)-1]
	sub := strings.Split(last, ":")
	if len(sub) != 2 {
		return "", fmt.Errorf("reference must include a tag: %s", ref)
	}

	name := fmt.Sprintf("%s-%s.tgz", sub[0], sub[1])
	localPath := filepath.Join(destDir, name)

	dst, err := file.New(destDir)
	if err != nil {
		panic(err)
	}

	_, err = oras.Copy(ctx, repo, ref, dst, "", oras.DefaultCopyOptions)
	if err != nil {
		return "", fmt.Errorf("oras pull failed: %w", err)
	}

	return localPath, nil
}
