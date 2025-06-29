package oci

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
)

type FakeOrasClient struct{}

func (f *FakeOrasClient) Copy(ctx context.Context, src oras.ReadOnlyTarget, srcRef string, dst oras.Target, dstRef string, options oras.CopyOptions) (v1.Descriptor, error) {
	// Simulate a successful copy operation
	return v1.Descriptor{
		MediaType: "application/vnd.oci.image.manifest.v1+json",
	}, nil
}

func TestPull(t *testing.T) {
	testCases := []struct {
		name        string
		ref         string
		expectedDir string
	}{
		{
			name:        "pull package",
			ref:         "localhost:5000/myorg/mypackage:1.0.0",
			expectedDir: ".universal-packages/myorg/mypackage",
		},
		{
			name:        "pull package without organization",
			ref:         "localhost:5000/mypackage:1.0.0",
			expectedDir: ".universal-packages/mypackage",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			client := &FakeOrasClient{}
			ctx := context.Background()

			repo, err := ConnectToRegistry(testCase.ref)
			if err != nil {
				t.Fatalf("failed to connect to registry: %v", err)
			}

			workingDir, err := Pull(ctx, client, repo, "", "./.universal-packages")
			if err != nil {
				t.Fatalf("failed to pull package: %v", err)
			}

			if workingDir != testCase.expectedDir {
				t.Errorf("expected working directory %s, got %s", testCase.expectedDir, workingDir)
			}
		})
	}
}

func TestPush(t *testing.T) {
	dir := "../../testdata"

	// Create temp file for install location
	tempDir, err := os.MkdirTemp(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to remove temp folder: %v\n", err)
		}
	}() // Clean up
	// Create a fake .tar file in the directory
	packagePath := filepath.Join(tempDir, "mypackage.tgz")
	err = os.WriteFile(packagePath, []byte("{}"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name string
		ref  string
	}{
		{
			name: "push package",
			ref:  "localhost:5000/myorg/mypackage:1.0.0",
		},
		{
			name: "push package without organization",
			ref:  "localhost:5000/mypackage:1.0.0",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			client := &FakeOrasClient{}
			ctx := context.Background()

			err := Push(ctx, client, testCase.ref, packagePath)
			if err != nil {
				t.Fatalf("failed to push package: %v", err)
			}
		})
	}
}

func TestGetPackageNameVersionFromRef(t *testing.T) {
	testCases := []struct {
		ref             string
		expectedName    string
		expectedVersion string
		expectedError   bool
	}{
		{
			ref:             "localhost:5000/myorg/mypackage:1.0.0",
			expectedName:    "mypackage",
			expectedVersion: "1.0.0",
			expectedError:   false,
		},
		{
			ref:             "ghcr.io/myorg/mypackage:1.0.0",
			expectedName:    "mypackage",
			expectedVersion: "1.0.0",
			expectedError:   false,
		},
		{
			ref:             "localhost:5000/mypackage:1.0.0",
			expectedName:    "mypackage",
			expectedVersion: "1.0.0",
			expectedError:   false,
		},
		{
			ref:             "localhost:5000/mypackage",
			expectedName:    "mypackage",
			expectedVersion: "latest",
			expectedError:   false,
		},
		{
			ref:             "invalid-ref",
			expectedName:    "",
			expectedVersion: "",
			expectedError:   true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.ref, func(t *testing.T) {
			name, version, err := GetPackageNameVersionFromRef(testCase.ref)
			if testCase.expectedError {
				if err == nil {
					t.Errorf("expected error for ref %s, got nil", testCase.ref)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error for ref %s: %v", testCase.ref, err)
				return
			}
			if name != testCase.expectedName || version != testCase.expectedVersion {
				t.Errorf("expected (%s, %s), got (%s, %s)", testCase.expectedName, testCase.expectedVersion, name, version)
			}
		})
	}
}

func TestGetBaseName(t *testing.T) {
	dir := "../../testdata"

	// Create temp file for install location
	tempDir, err := os.MkdirTemp(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to remove temp folder: %v\n", err)
		}
	}() // Clean up
	// Create a fake .tar file in the directory
	packagePath := filepath.Join(tempDir, "package.tgz")
	err = os.WriteFile(packagePath, []byte("{}"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		path          string
		expectedName  string
		expectedError bool
	}{
		{
			path:          packagePath,
			expectedName:  "package.tgz",
			expectedError: false,
		},
		{
			path:          filepath.Join(tempDir, "nonexistent.tgz"),
			expectedName:  "",
			expectedError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.path, func(t *testing.T) {
			name, err := getBaseName(testCase.path)
			if testCase.expectedError {
				if err == nil {
					t.Errorf("expected error for path %s, got nil", testCase.path)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error for path %s: %v", testCase.path, err)
				return
			}
			if name != testCase.expectedName {
				t.Errorf("expected %s, got %s", testCase.expectedName, name)
			}
		})
	}
}
