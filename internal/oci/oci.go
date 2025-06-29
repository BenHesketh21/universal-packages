package oci

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"
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
	// Set up Docker-compatible credential store
	// 3. Connect to a remote repository

	storeOpts := credentials.StoreOptions{}
	credStore, err := credentials.NewStoreFromDocker(storeOpts)
	if err != nil {
		panic(err)
	}

	repo.Client = &auth.Client{
		Client:     retry.DefaultClient,
		Cache:      auth.NewCache(),
		Credential: credentials.Credential(credStore),
	}

	return repo, nil
}

func Pull(ctx context.Context, client OrasClient, ref string, upRootDir string) (string, error) {

	repo, err := ConnectToRegistry(ref)
	if err != nil {
		return "", fmt.Errorf("failed to connect to registry: %w", err)
	}

	repositoryName := repo.Reference.Repository

	workingDir := filepath.Join(upRootDir, repositoryName)
	dst, err := file.New(workingDir)
	if err != nil {
		panic(err)
	}
	_, err = client.Copy(ctx, repo, ref, dst, "", oras.DefaultCopyOptions)
	if err != nil {
		return "", fmt.Errorf("oras pull failed: %w", err)
	}

	return workingDir, nil
}

func Push(ctx context.Context, orasClient OrasClient, ref string, tarballPath string) error {
	// 0. Create a file store
	fs, err := file.New("")
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}
	defer func() {
		if err := fs.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to close file: %v\n", err)
		}
	}()

	// 1. Add files to the file store
	fileName, err := getBaseName(tarballPath)
	if err != nil {
		return fmt.Errorf("failed to get base name from tarball path: %w", err)
	}
	mediaType := "application/vnd.test.file"
	fileDescriptors := make([]v1.Descriptor, 0, 1)
	fileDescriptor, err := fs.Add(ctx, fileName, mediaType, tarballPath)
	if err != nil {
		return fmt.Errorf("failed to add tarball: %w", err)
	}
	fileDescriptors = append(fileDescriptors, fileDescriptor)

	// 2. Pack the files and tag the packed manifest
	artifactType := "application/vnd.test.artifact"
	opts := oras.PackManifestOptions{
		Layers: fileDescriptors,
	}
	manifestDescriptor, err := oras.PackManifest(ctx, fs, oras.PackManifestVersion1_1, artifactType, opts)
	if err != nil {
		return fmt.Errorf("failed to pack manifest: %w", err)
	}

	repo, err := ConnectToRegistry(ref)
	if err != nil {
		return fmt.Errorf("failed to connect to registry: %w", err)
	}

	tag := repo.Reference.Reference
	if err = fs.Tag(ctx, manifestDescriptor, tag); err != nil {
		return fmt.Errorf("failed to tag artifact: %w", err)
	}

	// 4. Copy from the file store to the remote repository
	_, err = orasClient.Copy(ctx, fs, tag, repo, tag, oras.DefaultCopyOptions)
	if err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	return nil
}

// Utility function to extract filename from path
func getBaseName(path string) (string, error) {
	if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
		return stat.Name(), nil
	}
	return "", fmt.Errorf("no file found: %s", path)
}

func GetPackageNameVersionFromRef(ref string) (string, string, error) {
	if ref == "" {
		return "", "", fmt.Errorf("empty reference")
	}

	// Default tag if not explicitly provided
	tag := "latest"
	path := ref

	// Check if a tag is specified
	lastColon := strings.LastIndex(ref, ":")
	if lastColon == -1 {
		return "", "", fmt.Errorf("missing tag in reference: %s", ref)
	}
	lastSlash := strings.LastIndex(ref, "/")

	if lastColon > lastSlash {
		// colon after last slash -> it's a tag
		path = ref[:lastColon]
		tag = ref[lastColon+1:]
	}

	// Split host from repo path
	slashParts := strings.SplitN(path, "/", 3)

	repoName := ""
	if len(slashParts) == 2 {
		// No org, just repo name
		repoName = slashParts[1]
	} else if len(slashParts) == 3 {
		// Org/repo format
		repoName = slashParts[2]
	} else {
		return "", "", fmt.Errorf("invalid reference format: %s", ref)
	}

	if repoName == "" || tag == "" {
		return "", "", fmt.Errorf("invalid reference format: %s", ref)
	}

	return repoName, tag, nil
}
