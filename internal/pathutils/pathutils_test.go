// code-from-spec: SPEC/golang/tests/os/path_utils@wkRM_mGL3Z8Ro-IOfy-MkIClr5k
package pathutils_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

func testChdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChdir: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("testChdir cleanup: %v", err)
		}
	})
}

func TestPathValidateCfs(t *testing.T) {
	type testCase struct {
		name        string
		input       string
		expectedErr error
	}

	cases := []testCase{
		{
			name:        "TC-PV-01: valid simple relative path",
			input:       "internal/config/config.go",
			expectedErr: nil,
		},
		{
			name:        "TC-PV-02: valid nested path",
			input:       "cmd/framework-mcp/main.go",
			expectedErr: nil,
		},
		{
			name:        "TC-PV-03: valid single filename",
			input:       "main.go",
			expectedErr: nil,
		},
		{
			name:        "TC-PV-04: accepts path with dot segment",
			input:       "internal/./config/config.go",
			expectedErr: nil,
		},
		{
			name:        "TC-PV-05: accepts traversal that resolves within root",
			input:       "a/b/../c",
			expectedErr: nil,
		},
		{
			name:        "TC-PV-06: accepts path with trailing slash",
			input:       "internal/config/",
			expectedErr: nil,
		},
		{
			name:        "TC-PV-07: accepts path with duplicate slashes",
			input:       "internal//config//file.go",
			expectedErr: nil,
		},
		{
			name:        "TC-PV-08: rejects empty string",
			input:       "",
			expectedErr: pathutils.ErrPathEmpty,
		},
		{
			name:        "TC-PV-09: rejects absolute path with leading slash",
			input:       "/etc/passwd",
			expectedErr: pathutils.ErrPathAbsolute,
		},
		{
			name:        "TC-PV-10: rejects absolute path with drive letter",
			input:       "C:/Windows/system32",
			expectedErr: pathutils.ErrPathAbsolute,
		},
		{
			name:        "TC-PV-11: rejects backslash",
			input:       `internal\config\config.go`,
			expectedErr: pathutils.ErrPathContainsBackslash,
		},
		{
			name:        "TC-PV-12: rejects simple traversal",
			input:       "../../etc/passwd",
			expectedErr: pathutils.ErrDirectoryTraversal,
		},
		{
			name:        "TC-PV-13: rejects embedded traversal",
			input:       "internal/../../outside/file.go",
			expectedErr: pathutils.ErrDirectoryTraversal,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := pathutils.PathValidateCfs(tc.input)
			if tc.expectedErr == nil {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error %v, got nil", tc.expectedErr)
			}
			if !isErr(err, tc.expectedErr) {
				t.Fatalf("expected error %v, got: %v", tc.expectedErr, err)
			}
		})
	}
}

func TestPathCfsToOs(t *testing.T) {
	t.Run("TC-CO-01: converts valid path that exists", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		subDir := filepath.Join(tempDir, "internal", "config")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatalf("setup: %v", err)
		}
		if err := os.WriteFile(filepath.Join(subDir, "config.go"), []byte("package config"), 0644); err != nil {
			t.Fatalf("setup: %v", err)
		}

		cfsPath := &pathutils.PathCfs{Value: "internal/config/config.go"}
		result, err := pathutils.PathCfsToOs(cfsPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !filepath.IsAbs(result.Value) {
			t.Fatalf("expected absolute path, got: %s", result.Value)
		}
		expected := filepath.Join("internal", "config", "config.go")
		if !strings.HasSuffix(result.Value, expected) {
			t.Fatalf("expected path ending with %s, got: %s", expected, result.Value)
		}
	})

	t.Run("TC-CO-02: converts valid path that does not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		cfsPath := &pathutils.PathCfs{Value: "internal/newdir/newfile.go"}
		result, err := pathutils.PathCfsToOs(cfsPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !filepath.IsAbs(result.Value) {
			t.Fatalf("expected absolute path, got: %s", result.Value)
		}
		expected := filepath.Join("internal", "newdir", "newfile.go")
		if !strings.HasSuffix(result.Value, expected) {
			t.Fatalf("expected path ending with %s, got: %s", expected, result.Value)
		}
	})

	t.Run("TC-CO-03: converts path with duplicate slashes", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		cfsPath := &pathutils.PathCfs{Value: "internal//config.go"}
		result, err := pathutils.PathCfsToOs(cfsPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !filepath.IsAbs(result.Value) {
			t.Fatalf("expected absolute path, got: %s", result.Value)
		}
	})

	t.Run("TC-CO-04: rejects invalid CfsPath — directory traversal", func(t *testing.T) {
		cfsPath := &pathutils.PathCfs{Value: "../../etc/passwd"}
		result, err := pathutils.PathCfsToOs(cfsPath)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !isErr(err, pathutils.ErrDirectoryTraversal) {
			t.Fatalf("expected ErrDirectoryTraversal, got: %v", err)
		}
		if result != nil {
			t.Fatalf("expected nil result, got: %v", result)
		}
	})

	t.Run("TC-CO-05: rejects symlink escaping project root", func(t *testing.T) {
		outsideDir := t.TempDir()
		outsideFile := filepath.Join(outsideDir, "secret.txt")
		if err := os.WriteFile(outsideFile, []byte("secret"), 0644); err != nil {
			t.Fatalf("setup: %v", err)
		}

		tempDir := t.TempDir()
		testChdir(t, tempDir)

		symlinkPath := filepath.Join(tempDir, "link.txt")
		if err := os.Symlink(outsideFile, symlinkPath); err != nil {
			t.Skip("symlinks not supported:", err)
		}

		cfsPath := &pathutils.PathCfs{Value: "link.txt"}
		_, err := pathutils.PathCfsToOs(cfsPath)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !isErr(err, pathutils.ErrResolvesOutsideRoot) {
			t.Fatalf("expected ErrResolvesOutsideRoot, got: %v", err)
		}
	})

	t.Run("TC-CO-06: roundtrip CfsToOs then OsToCfs", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		subDir := filepath.Join(tempDir, "internal", "config")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatalf("setup: %v", err)
		}
		if err := os.WriteFile(filepath.Join(subDir, "config.go"), []byte("package config"), 0644); err != nil {
			t.Fatalf("setup: %v", err)
		}

		cfsPath := &pathutils.PathCfs{Value: "internal/config/config.go"}
		osPath, err := pathutils.PathCfsToOs(cfsPath)
		if err != nil {
			t.Fatalf("PathCfsToOs: %v", err)
		}

		roundTripped, err := pathutils.PathOsToCfs(osPath)
		if err != nil {
			t.Fatalf("PathOsToCfs: %v", err)
		}

		if roundTripped.Value != "internal/config/config.go" {
			t.Fatalf("expected %q, got %q", "internal/config/config.go", roundTripped.Value)
		}
	})
}

func TestPathOsToCfs(t *testing.T) {
	t.Run("TC-OC-01: converts valid OS path that exists", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		subDir := filepath.Join(tempDir, "somedir")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatalf("setup: %v", err)
		}
		filePath := filepath.Join(subDir, "file.go")
		if err := os.WriteFile(filePath, []byte("package somedir"), 0644); err != nil {
			t.Fatalf("setup: %v", err)
		}

		osPath := &pathutils.PathOs{Value: filePath}
		result, err := pathutils.PathOsToCfs(osPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(result.Value, "\\") {
			t.Fatalf("expected forward slashes, got: %s", result.Value)
		}
		if filepath.IsAbs(result.Value) {
			t.Fatalf("expected relative path, got: %s", result.Value)
		}
	})

	t.Run("TC-OC-02: converts valid OS path that does not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		nonexistentPath := filepath.Join(tempDir, "nonexistent", "file.go")
		osPath := &pathutils.PathOs{Value: nonexistentPath}
		result, err := pathutils.PathOsToCfs(osPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(result.Value, "\\") {
			t.Fatalf("expected forward slashes, got: %s", result.Value)
		}
		if filepath.IsAbs(result.Value) {
			t.Fatalf("expected relative path, got: %s", result.Value)
		}
	})

	t.Run("TC-OC-03: result uses forward slashes", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		subDir := filepath.Join(tempDir, "a", "b")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatalf("setup: %v", err)
		}
		filePath := filepath.Join(subDir, "c.go")
		if err := os.WriteFile(filePath, []byte("package b"), 0644); err != nil {
			t.Fatalf("setup: %v", err)
		}

		osPath := &pathutils.PathOs{Value: filePath}
		result, err := pathutils.PathOsToCfs(osPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(result.Value, "\\") {
			t.Fatalf("result must not contain backslashes, got: %s", result.Value)
		}
	})

	t.Run("TC-OC-04: symlink within root resolving within root", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		targetFile := filepath.Join(tempDir, "real.go")
		if err := os.WriteFile(targetFile, []byte("package main"), 0644); err != nil {
			t.Fatalf("setup: %v", err)
		}

		symlinkPath := filepath.Join(tempDir, "link.go")
		if err := os.Symlink(targetFile, symlinkPath); err != nil {
			t.Skip("symlinks not supported:", err)
		}

		osPath := &pathutils.PathOs{Value: symlinkPath}
		result, err := pathutils.PathOsToCfs(osPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(result.Value, "\\") {
			t.Fatalf("expected forward slashes, got: %s", result.Value)
		}
		if filepath.IsAbs(result.Value) {
			t.Fatalf("expected relative path, got: %s", result.Value)
		}
	})

	t.Run("TC-OC-05: rejects path outside project root", func(t *testing.T) {
		outsideDir := t.TempDir()
		outsidePath := filepath.Join(outsideDir, "outside.go")

		osPath := &pathutils.PathOs{Value: outsidePath}
		_, err := pathutils.PathOsToCfs(osPath)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !isErr(err, pathutils.ErrResolvesOutsideRoot) {
			t.Fatalf("expected ErrResolvesOutsideRoot, got: %v", err)
		}
	})
}

func TestPathGetProjectRoot(t *testing.T) {
	t.Run("TC-PGR-01: returns an absolute path", func(t *testing.T) {
		result, err := pathutils.PathGetProjectRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result.Value == "" {
			t.Fatal("expected non-empty path")
		}
		if !filepath.IsAbs(result.Value) {
			t.Fatalf("expected absolute path, got: %s", result.Value)
		}
	})

	t.Run("TC-PGR-02: matches working directory", func(t *testing.T) {
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("os.Getwd: %v", err)
		}

		result, err := pathutils.PathGetProjectRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wdEval, err := filepath.EvalSymlinks(wd)
		if err != nil {
			wdEval = wd
		}
		resultEval, err := filepath.EvalSymlinks(result.Value)
		if err != nil {
			resultEval = result.Value
		}

		if wdEval != resultEval {
			t.Fatalf("expected %s, got %s", wdEval, resultEval)
		}
	})
}

func isErr(err, target error) bool {
	if err == nil {
		return target == nil
	}
	type unwrapper interface {
		Unwrap() error
	}
	type multiUnwrapper interface {
		Unwrap() []error
	}
	if err == target {
		return true
	}
	if u, ok := err.(unwrapper); ok {
		return isErr(u.Unwrap(), target)
	}
	if mu, ok := err.(multiUnwrapper); ok {
		for _, e := range mu.Unwrap() {
			if isErr(e, target) {
				return true
			}
		}
	}
	return false
}
