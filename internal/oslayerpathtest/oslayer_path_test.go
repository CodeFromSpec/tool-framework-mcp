// code-from-spec: SPEC/golang/tests/oslayer/path@iDB96jy2SnJjsGMFeF1TZYha-aE
package oslayerpathtest_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
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

func TestValidateCfsPath(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr error
	}{
		{
			name:        "valid simple relative path",
			input:       "internal/config/config.go",
			expectedErr: nil,
		},
		{
			name:        "valid nested path",
			input:       "cmd/framework-mcp/main.go",
			expectedErr: nil,
		},
		{
			name:        "valid single filename",
			input:       "main.go",
			expectedErr: nil,
		},
		{
			name:        "accepts path with dot segment",
			input:       "internal/./config/config.go",
			expectedErr: nil,
		},
		{
			name:        "accepts traversal that resolves within root",
			input:       "a/b/../c",
			expectedErr: nil,
		},
		{
			name:        "accepts path with trailing slash",
			input:       "internal/config/",
			expectedErr: nil,
		},
		{
			name:        "accepts path with duplicate slashes",
			input:       "internal//config//file.go",
			expectedErr: nil,
		},
		{
			name:        "rejects empty string",
			input:       "",
			expectedErr: oslayer.ErrPathEmpty,
		},
		{
			name:        "rejects absolute path with leading slash",
			input:       "/etc/passwd",
			expectedErr: oslayer.ErrPathAbsolute,
		},
		{
			name:        "rejects absolute path with drive letter",
			input:       "C:/Windows/system32",
			expectedErr: oslayer.ErrPathAbsolute,
		},
		{
			name:        "rejects backslash",
			input:       `internal\config\config.go`,
			expectedErr: oslayer.ErrPathContainsBackslash,
		},
		{
			name:        "rejects simple traversal",
			input:       "../../etc/passwd",
			expectedErr: oslayer.ErrDirectoryTraversal,
		},
		{
			name:        "rejects embedded traversal",
			input:       "internal/../../outside/file.go",
			expectedErr: oslayer.ErrDirectoryTraversal,
		},
		{
			name:        "rejects lowercase drive letter",
			input:       "c:/Windows/system32",
			expectedErr: oslayer.ErrPathAbsolute,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := oslayer.ValidateCfsPath(tc.input)
			if tc.expectedErr == nil {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
			} else {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got: %v", tc.expectedErr, err)
				}
			}
		})
	}
}

func TestCfsPathToOs(t *testing.T) {
	t.Run("converts valid path that exists", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		err := os.MkdirAll("internal/config", 0755)
		if err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		err = os.WriteFile("internal/config/config.go", []byte("package config"), 0644)
		if err != nil {
			t.Fatalf("write: %v", err)
		}

		result, err := oslayer.CfsPathToOs("internal/config/config.go")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !filepath.IsAbs(string(result)) {
			t.Fatalf("expected absolute path, got: %s", result)
		}
		if !strings.HasSuffix(string(result), filepath.Join("internal", "config", "config.go")) {
			t.Fatalf("unexpected result: %s", result)
		}
	})

	t.Run("converts valid path that does not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		result, err := oslayer.CfsPathToOs("internal/newdir/newfile.go")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !filepath.IsAbs(string(result)) {
			t.Fatalf("expected absolute path, got: %s", result)
		}
	})

	t.Run("converts path with duplicate slashes", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		result, err := oslayer.CfsPathToOs("internal//config.go")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !filepath.IsAbs(string(result)) {
			t.Fatalf("expected absolute path, got: %s", result)
		}
	})

	t.Run("rejects invalid CfsPath", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		_, err := oslayer.CfsPathToOs("../../etc/passwd")
		if !errors.Is(err, oslayer.ErrDirectoryTraversal) {
			t.Fatalf("expected ErrDirectoryTraversal, got: %v", err)
		}
	})

	t.Run("rejects symlink escaping project root", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		outsideDir := t.TempDir()
		outsideFile := filepath.Join(outsideDir, "secret.txt")
		err := os.WriteFile(outsideFile, []byte("secret"), 0644)
		if err != nil {
			t.Fatalf("write outside file: %v", err)
		}

		err = os.WriteFile("link.txt", []byte("placeholder"), 0644)
		if err != nil {
			t.Fatalf("write placeholder: %v", err)
		}
		err = os.Remove("link.txt")
		if err != nil {
			t.Fatalf("remove placeholder: %v", err)
		}
		err = os.Symlink(outsideFile, "link.txt")
		if err != nil {
			if errors.Is(err, os.ErrInvalid) || strings.Contains(err.Error(), "not supported") || strings.Contains(err.Error(), "privilege") {
				t.Skip("symlinks not supported on this platform")
			}
			t.Fatalf("symlink: %v", err)
		}

		_, err = oslayer.CfsPathToOs("link.txt")
		if !errors.Is(err, oslayer.ErrResolvesOutsideRoot) {
			t.Fatalf("expected ErrResolvesOutsideRoot, got: %v", err)
		}
	})

	t.Run("rejects path whose root is a prefix but not a boundary", func(t *testing.T) {
		tempDir := t.TempDir()

		projectDir := filepath.Join(tempDir, "project")
		projectOtherDir := filepath.Join(tempDir, "projectother")
		err := os.MkdirAll(projectDir, 0755)
		if err != nil {
			t.Fatalf("mkdir project: %v", err)
		}
		err = os.MkdirAll(projectOtherDir, 0755)
		if err != nil {
			t.Fatalf("mkdir projectother: %v", err)
		}

		testChdir(t, projectDir)

		outsideFile := filepath.Join(projectOtherDir, "file.txt")
		err = os.WriteFile(outsideFile, []byte("data"), 0644)
		if err != nil {
			t.Fatalf("write: %v", err)
		}

		_, err = oslayer.CfsPathToOs("../projectother/file.txt")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, oslayer.ErrDirectoryTraversal) && !errors.Is(err, oslayer.ErrResolvesOutsideRoot) {
			t.Fatalf("expected ErrDirectoryTraversal or ErrResolvesOutsideRoot, got: %v", err)
		}
	})

	t.Run("roundtrip CfsPathToOs then OsPathToCfs", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		err := os.MkdirAll("internal/config", 0755)
		if err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		err = os.WriteFile("internal/config/config.go", []byte("package config"), 0644)
		if err != nil {
			t.Fatalf("write: %v", err)
		}

		original := oslayer.CfsPath("internal/config/config.go")
		osPath, err := oslayer.CfsPathToOs(original)
		if err != nil {
			t.Fatalf("CfsPathToOs: %v", err)
		}
		result, err := oslayer.OsPathToCfs(osPath)
		if err != nil {
			t.Fatalf("OsPathToCfs: %v", err)
		}
		if result != original {
			t.Fatalf("roundtrip mismatch: got %q, want %q", result, original)
		}
	})
}

func TestOsPathToCfs(t *testing.T) {
	t.Run("converts valid OS path that exists", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		err := os.MkdirAll("subdir", 0755)
		if err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		err = os.WriteFile("subdir/file.go", []byte("data"), 0644)
		if err != nil {
			t.Fatalf("write: %v", err)
		}

		osPath := oslayer.OsPath(filepath.Join(tempDir, "subdir", "file.go"))
		result, err := oslayer.OsPathToCfs(osPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(result) != "subdir/file.go" {
			t.Fatalf("expected subdir/file.go, got: %s", result)
		}
	})

	t.Run("converts valid OS path that does not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		osPath := oslayer.OsPath(filepath.Join(tempDir, "nonexistent", "file.go"))
		result, err := oslayer.OsPathToCfs(osPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(string(result), "\\") {
			t.Fatalf("result contains backslashes: %s", result)
		}
	})

	t.Run("result uses forward slashes", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		err := os.MkdirAll("a/b/c", 0755)
		if err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		err = os.WriteFile("a/b/c/file.go", []byte("data"), 0644)
		if err != nil {
			t.Fatalf("write: %v", err)
		}

		osPath := oslayer.OsPath(filepath.Join(tempDir, "a", "b", "c", "file.go"))
		result, err := oslayer.OsPathToCfs(osPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(string(result), "\\") {
			t.Fatalf("result contains backslashes: %s", result)
		}
		if string(result) != "a/b/c/file.go" {
			t.Fatalf("expected a/b/c/file.go, got: %s", result)
		}
	})

	t.Run("symlink within root resolving within root", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		err := os.WriteFile("realfile.go", []byte("data"), 0644)
		if err != nil {
			t.Fatalf("write: %v", err)
		}

		err = os.Symlink(filepath.Join(tempDir, "realfile.go"), filepath.Join(tempDir, "linkfile.go"))
		if err != nil {
			if errors.Is(err, os.ErrInvalid) || strings.Contains(err.Error(), "not supported") || strings.Contains(err.Error(), "privilege") {
				t.Skip("symlinks not supported on this platform")
			}
			t.Fatalf("symlink: %v", err)
		}

		osPath := oslayer.OsPath(filepath.Join(tempDir, "linkfile.go"))
		_, err = oslayer.OsPathToCfs(osPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("rejects path outside project root", func(t *testing.T) {
		tempDir := t.TempDir()
		projectDir := filepath.Join(tempDir, "project")
		err := os.MkdirAll(projectDir, 0755)
		if err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		testChdir(t, projectDir)

		outsideFile := filepath.Join(tempDir, "outside.go")
		err = os.WriteFile(outsideFile, []byte("data"), 0644)
		if err != nil {
			t.Fatalf("write: %v", err)
		}

		_, err = oslayer.OsPathToCfs(oslayer.OsPath(outsideFile))
		if !errors.Is(err, oslayer.ErrResolvesOutsideRoot) {
			t.Fatalf("expected ErrResolvesOutsideRoot, got: %v", err)
		}
	})

	t.Run("rejects OS path whose root is a prefix but not a boundary", func(t *testing.T) {
		tempDir := t.TempDir()

		projectDir := filepath.Join(tempDir, "project")
		projectOtherDir := filepath.Join(tempDir, "projectother")
		err := os.MkdirAll(projectDir, 0755)
		if err != nil {
			t.Fatalf("mkdir project: %v", err)
		}
		err = os.MkdirAll(projectOtherDir, 0755)
		if err != nil {
			t.Fatalf("mkdir projectother: %v", err)
		}

		testChdir(t, projectDir)

		outsideFile := filepath.Join(projectOtherDir, "file.txt")
		err = os.WriteFile(outsideFile, []byte("data"), 0644)
		if err != nil {
			t.Fatalf("write: %v", err)
		}

		_, err = oslayer.OsPathToCfs(oslayer.OsPath(outsideFile))
		if !errors.Is(err, oslayer.ErrResolvesOutsideRoot) {
			t.Fatalf("expected ErrResolvesOutsideRoot, got: %v", err)
		}
	})
}

func TestGetProjectRoot(t *testing.T) {
	t.Run("returns an absolute path", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		root, err := oslayer.GetProjectRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(root) == "" {
			t.Fatal("expected non-empty path")
		}
		if !filepath.IsAbs(string(root)) {
			t.Fatalf("expected absolute path, got: %s", root)
		}
	})

	t.Run("matches working directory", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		root, err := oslayer.GetProjectRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Getwd: %v", err)
		}

		rootEval, err := filepath.EvalSymlinks(string(root))
		if err != nil {
			rootEval = string(root)
		}
		wdEval, err := filepath.EvalSymlinks(wd)
		if err != nil {
			wdEval = wd
		}

		if rootEval != wdEval {
			t.Fatalf("root %q does not match working directory %q", rootEval, wdEval)
		}
	})
}
