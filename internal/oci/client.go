package oci

import (
	"context"
	"fmt"
	"os"
	"path"
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

func Push(ref string, tarballPath string) error {
	// 0. Create a file store
	fs, err := file.New("")
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}
	defer fs.Close()
	ctx := context.Background()

	// 1. Add files to the file store
	fileName := getBaseName(tarballPath)
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

	// 3. Connect to a remote repository
	repo, err := remote.NewRepository(ref)
	if err != nil {
		return fmt.Errorf("failed to parse ref: %w", err)
	}

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

	tag := repo.Reference.Reference
	if err = fs.Tag(ctx, manifestDescriptor, tag); err != nil {
		return fmt.Errorf("failed to tag artifact: %w", err)
	}

	// 4. Copy from the file store to the remote repository
	_, err = oras.Copy(ctx, fs, tag, repo, tag, oras.DefaultCopyOptions)
	if err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	return nil
}

// Utility function to extract filename from path
func getBaseName(path string) string {
	if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
		return stat.Name()
	}
	return "package.tgz"
}

func GetPackageNameFromRef(ref string) string {
	trimmed := strings.Split(ref, "@")[0]
	trimmed = strings.Split(trimmed, ":")[0]
	return path.Base(trimmed)
}
