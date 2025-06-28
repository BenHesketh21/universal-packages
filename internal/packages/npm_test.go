package packages

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLocatePackage(t *testing.T) {
	handler := &NpmHandler{}

	// Adding lodash package for testing
	dir := "../../testdata"
	packageFile := "lodash-4.17.21.tgz"

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
	tarFilePath := filepath.Join(tempDir, packageFile)
	err = os.WriteFile(tarFilePath, []byte("fake tar content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		packageName    string
		packageVersion string
		expectedPath   string
		expectedError  bool
	}{
		{
			packageName:    "lodash",
			packageVersion: "4.17.21",
			expectedPath:   filepath.Join(tempDir, "lodash-4.17.21.tgz"),
			expectedError:  false,
		},
		{
			packageName:    "nonexistent-package",
			packageVersion: "1.0.0",
			expectedPath:   "",
			expectedError:  true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.packageName, func(t *testing.T) {
			filePath, err := handler.LocatePackage(tempDir, testCase.packageName, testCase.packageVersion)
			if testCase.expectedError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if filePath != testCase.expectedPath {
					t.Errorf("expected %s, got %s", testCase.expectedPath, filePath)
				}
			}
		})
	}
}

func TestUpdatePackageRef(t *testing.T) {
	testCases := []struct {
		name         string
		inputJSON    string
		packageName  string
		version      string
		expectedJSON string
	}{
		{
			name: "updates existing dependency",
			inputJSON: `{
                "dependencies": {
                    "lodash": "4.17.0"
                }
            }`,
			packageName: "lodash",
			version:     "4.18.0",
			expectedJSON: `{
  "dependencies": {
    "lodash": "file:.upkg/lodash-4.18.0.tgz"
  }
}`,
		},
		{
			name: "updates existing dependency",
			inputJSON: `{
                "dependencies": {
                    "lodash": "file:.upkg/lodash-4.17.0.tgz"
                }
            }`,
			packageName: "lodash",
			version:     "4.18.0",
			expectedJSON: `{
  "dependencies": {
    "lodash": "file:.upkg/lodash-4.18.0.tgz"
  }
}`,
		},
		{
			name: "adds new dependency",
			inputJSON: `{
                "dependencies": {
                    "express": "4.17.1"
                }
            }`,
			packageName: "lodash",
			version:     "4.18.0",
			expectedJSON: `{
  "dependencies": {
    "express": "4.17.1",
    "lodash": "file:.upkg/lodash-4.18.0.tgz"
  }
}`,
		},
		{
			name:        "creates dependencies section",
			inputJSON:   `{}`,
			packageName: "lodash",
			version:     "4.18.0",
			expectedJSON: `{
  "dependencies": {
    "lodash": "file:.upkg/lodash-4.18.0.tgz"
  }
}`,
		},
	}

	handler := &NpmHandler{}
	dir := "../../testdata"

	// Create temp file for install location
	installTempDir, err := os.MkdirTemp(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(installTempDir); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to remove temp folder: %v\n", err)
		}
	}() // Clean up
	// Create a fake .tar file in the directory
	installPackageJSONPath := filepath.Join(installTempDir, "package.json")
	err = os.WriteFile(installPackageJSONPath, []byte("{}"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			// Write initial input JSON
			err = os.WriteFile(installPackageJSONPath, []byte(testCase.inputJSON), 0644)
			if err != nil {
				t.Fatal(err)
			}

			// Run the function
			packageLocation := filepath.Join(installTempDir, ".upkg", fmt.Sprintf("%s-%s.tgz", testCase.packageName, testCase.version))
			err = handler.UpdatePackageRef(testCase.packageName, packageLocation, installPackageJSONPath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Read updated file
			updated, err := os.ReadFile(installPackageJSONPath)
			if err != nil {
				t.Fatal(err)
			}

			// Normalize JSON for comparison (in case of field order)
			var got, expected map[string]interface{}
			if err := json.Unmarshal(updated, &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(testCase.expectedJSON), &expected); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, expected) {
				t.Errorf("expected %v, got %v", expected, got)
			}
		})
	}
}

func TestFindPackageJSON(t *testing.T) {
	testCases := []struct {
		name string
		dir  string
	}{
		{
			name: "updates existing dependency",
			dir:  "../../testdata",
		},
	}
	dir := "../../testdata"

	// Create temp file for install location
	installTempDir, err := os.MkdirTemp(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(installTempDir); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to remove temp folder: %v\n", err)
		}
	}() // Clean up
	// Create a fake .tar file in the directory
	installPackageJSONPath := filepath.Join(installTempDir, "package.json")
	err = os.WriteFile(installPackageJSONPath, []byte("{}"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			// Run the function
			packageJSONPath, err := FindPackageJSON(installTempDir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			expectedFilePath := filepath.Join(installTempDir, "package.json")
			if packageJSONPath != expectedFilePath {
				t.Fatalf("expected %s, got %s", expectedFilePath, packageJSONPath)
			}
		})
	}
}
