// code-from-spec: SPEC/golang/tests/oslayer/path@4GZwutg1hMyyRmsidS-fm0WduJw
package oslayer_test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
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
		name    string
		input   string
		wantErr error
	}{
		{
			name:    "valid simple relative path",
			input:   "internal/config/config.go",
			wantErr: nil,
		},
		{
			name:    "valid nested path",
			input:   "cmd/framework-mcp/main.go",
			wantErr: nil,
		},
		{
			name:    "valid single filename",
			input:   "main.go",
			wantErr: nil,
		},
		{
			name:    "accepts path with dot segment",
			input:   "internal/./config/config.go",
			wantErr: nil,
		},
		{
			name:    "accepts traversal that resolves within root",
			input:   "a/b/../c",
			wantErr: nil,
		},
		{
			name:    "accepts path with trailing slash",
			input:   "internal/config/",
			wantErr: nil,
		},
		{
			name:    "accepts path with duplicate slashes",
			input:   "internal//config//file.go",
			wantErr: nil,
		},
		{
			name:    "rejects empty string",
			input:   "",
			wantErr: oslayer.ErrPathEmpty,
		},
		{
			name:    "rejects absolute path with leading slash",
			input:   "/etc/passwd",
			wantErr: oslayer.ErrPathAbsolute,
		},
		{
			name:    "rejects absolute path with drive letter",
			input:   "C:/Windows/system32",
			wantErr: oslayer.ErrPathAbsolute,
		},
		{
			name:    "rejects backslash",
			input:   `internal\config\config.go`,
			wantErr: oslayer.ErrPathContainsBackslash,
		},
		{
			name:    "rejects simple traversal",
			input:   "../../etc/passwd",
			wantErr: oslayer.ErrDirectoryTraversal,
		},
		{
			name:    "rejects embedded traversal",
			input:   "internal/../../outside/file.go",
			wantErr: oslayer.ErrDirectoryTraversal,
		},
		{
			name:    "rejects lowercase drive letter",
			input:   "c:/Windows/system32",
			wantErr: oslayer.ErrPathAbsolute,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := oslayer.ValidateCfsPath(tc.input)
			if tc.wantErr == nil {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("expected error %v, got: %v", tc.wantErr, err)
			}
		})
	}
}

func TestCfsPathToOs(t *testing.T) {
	t.Run("converts valid path that exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		testChdir(t, tmpDir)
		if err := os.MkdirAll("internal/config", 0755); err != nil {
			t.Fatalf("setup: %v", err)
		}
		if err := os.WriteFile("internal/config/config.go", []byte("package config"), 0644); err != nil {
			t.Fatalf("setup: %v", err)
		}
		result, err := oslayer.CfsPathToOs(oslayer.CfsPath("internal/config/config.go"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !filepath.IsAbs(string(result)) {
			t.Errorf("expected absolute path, got: %s", result)
		}
	})

	t.Run("converts valid path that does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		testChdir(t, tmpDir)
		result, err := oslayer.CfsPathToOs(oslayer.CfsPath("internal/newdir/newfile.go"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !filepath.IsAbs(string(result)) {
			t.Errorf("expected absolute path, got: %s", result)
		}
	})

	t.Run("converts path with duplicate slashes", func(t *testing.T) {
		tmpDir := t.TempDir()
		testChdir(t, tmpDir)
		result, err := oslayer.CfsPathToOs(oslayer.CfsPath("internal//config.go"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !filepath.IsAbs(string(result)) {
			t.Errorf("expected absolute path, got: %s", result)
		}
	})

	t.Run("rejects invalid CfsPath", func(t *testing.T) {
		tmpDir := t.TempDir()
		testChdir(t, tmpDir)
		_, err := oslayer.CfsPathToOs(oslayer.CfsPath("../../etc/passwd"))
		if !errors.Is(err, oslayer.ErrDirectoryTraversal) {
			t.Errorf("expected ErrDirectoryTraversal, got: %v", err)
		}
	})

	t.Run("rejects symlink escaping project root", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("symlink creation may require elevated privileges on Windows")
		}
		tmpDir := t.TempDir()
		testChdir(t, tmpDir)
		outsideDir := t.TempDir()
		outsideFile := filepath.Join(outsideDir, "secret.txt")
		if err := os.WriteFile(outsideFile, []byte("secret"), 0644); err != nil {
			t.Fatalf("setup: %v", err)
		}
		if err := os.MkdirAll("internal", 0755); err != nil {
			t.Fatalf("setup: %v", err)
		}
		if err := os.Symlink(outsideFile, filepath.Join(tmpDir, "internal", "escape.go")); err != nil {
			t.Skipf("symlink not supported: %v", err)
		}
		_, err := oslayer.CfsPathToOs(oslayer.CfsPath("internal/escape.go"))
		if !errors.Is(err, oslayer.ErrResolvesOutsideRoot) {
			t.Errorf("expected ErrResolvesOutsideRoot, got: %v", err)
		}
	})

	t.Run("roundtrip CfsPathToOs then OsPathToCfs", func(t *testing.T) {
		tmpDir := t.TempDir()
		testChdir(t, tmpDir)
		if err := os.MkdirAll("internal/config", 0755); err != nil {
			t.Fatalf("setup: %v", err)
		}
		if err := os.WriteFile("internal/config/config.go", []byte("package config"), 0644); err != nil {
			t.Fatalf("setup: %v", err)
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
			t.Errorf("roundtrip mismatch: got %q, want %q", result, original)
		}
	})
}

func TestOsPathToCfs(t *testing.T) {
	t.Run("converts valid OS path that exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		testChdir(t, tmpDir)
		if err := os.MkdirAll("internal/config", 0755); err != nil {
			t.Fatalf("setup: %v", err)
		}
		if err := os.WriteFile("internal/config/config.go", []byte("package config"), 0644); err != nil {
			t.Fatalf("setup: %v", err)
		}
		osPath := oslayer.OsPath(filepath.Join(tmpDir, "internal", "config", "config.go"))
		result, err := oslayer.OsPathToCfs(osPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != oslayer.CfsPath("internal/config/config.go") {
			t.Errorf("got %q, want %q", result, "internal/config/config.go")
		}
	})

	t.Run("converts valid OS path that does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		testChdir(t, tmpDir)
		osPath := oslayer.OsPath(filepath.Join(tmpDir, "internal", "newdir", "newfile.go"))
		result, err := oslayer.OsPathToCfs(osPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(result) == "" {
			t.Errorf("expected non-empty CfsPath")
		}
	})

	t.Run("result uses forward slashes", func(t *testing.T) {
		tmpDir := t.TempDir()
		testChdir(t, tmpDir)
		osPath := oslayer.OsPath(filepath.Join(tmpDir, "internal", "config", "config.go"))
		result, err := oslayer.OsPathToCfs(osPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(string(result), "\\") {
			t.Errorf("result contains backslash: %s", result)
		}
	})

	t.Run("symlink within root resolving within root", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("symlink creation may require elevated privileges on Windows")
		}
		tmpDir := t.TempDir()
		testChdir(t, tmpDir)
		if err := os.MkdirAll("internal/config", 0755); err != nil {
			t.Fatalf("setup: %v", err)
		}
		target := filepath.Join(tmpDir, "internal", "config", "config.go")
		if err := os.WriteFile(target, []byte("package config"), 0644); err != nil {
			t.Fatalf("setup: %v", err)
		}
		linkPath := filepath.Join(tmpDir, "internal", "config", "link.go")
		if err := os.Symlink(target, linkPath); err != nil {
			t.Skipf("symlink not supported: %v", err)
		}
		result, err := oslayer.OsPathToCfs(oslayer.OsPath(linkPath))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(string(result), "\\") {
			t.Errorf("result contains backslash: %s", result)
		}
	})

	t.Run("rejects path outside project root", func(t *testing.T) {
		tmpDir := t.TempDir()
		testChdir(t, tmpDir)
		outsideDir := t.TempDir()
		outsidePath := oslayer.OsPath(filepath.Join(outsideDir, "outside.go"))
		_, err := oslayer.OsPathToCfs(outsidePath)
		if !errors.Is(err, oslayer.ErrResolvesOutsideRoot) {
			t.Errorf("expected ErrResolvesOutsideRoot, got: %v", err)
		}
	})
}

func TestGetProjectRoot(t *testing.T) {
	t.Run("returns an absolute path", func(t *testing.T) {
		result, err := oslayer.GetProjectRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(result) == "" {
			t.Errorf("expected non-empty OsPath")
		}
		if !filepath.IsAbs(string(result)) {
			t.Errorf("expected absolute path, got: %s", result)
		}
	})

	t.Run("matches working directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		testChdir(t, tmpDir)
		result, err := oslayer.GetProjectRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("os.Getwd: %v", err)
		}
		wdResolved, err := filepath.EvalSymlinks(wd)
		if err != nil {
			wdResolved = wd
		}
		resultResolved, err := filepath.EvalSymlinks(string(result))
		if err != nil {
			resultResolved = string(result)
		}
		if resultResolved != wdResolved {
			t.Errorf("got %q, want %q", resultResolved, wdResolved)
		}
	})
}
