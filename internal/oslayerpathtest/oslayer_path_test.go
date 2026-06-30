// code-from-spec: SPEC/golang/test/cases/oslayer/path@MbWnnIw7DUu1S1TXunVI2bfv8D0
package oslayerpathtest

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"
)

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
					t.Errorf("expected no error, got %v", err)
				}
			} else {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("expected error %v, got %v", tc.wantErr, err)
				}
			}
		})
	}
}

func TestCfsPathToOs(t *testing.T) {
	t.Run("converts valid path that exists", func(t *testing.T) {
		testutils.Chdir(t)
		if err := os.MkdirAll("internal/config", 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("internal/config/config.go", []byte("package config"), 0644); err != nil {
			t.Fatal(err)
		}
		result, err := oslayer.CfsPathToOs("internal/config/config.go")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !filepath.IsAbs(string(result)) {
			t.Errorf("expected absolute path, got %q", result)
		}
		if !strings.HasSuffix(filepath.ToSlash(string(result)), "internal/config/config.go") {
			t.Errorf("expected path to end with internal/config/config.go, got %q", result)
		}
	})

	t.Run("converts valid path that does not exist", func(t *testing.T) {
		testutils.Chdir(t)
		result, err := oslayer.CfsPathToOs("internal/newdir/newfile.go")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !filepath.IsAbs(string(result)) {
			t.Errorf("expected absolute path, got %q", result)
		}
	})

	t.Run("converts path with duplicate slashes", func(t *testing.T) {
		testutils.Chdir(t)
		result, err := oslayer.CfsPathToOs("internal//config.go")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !filepath.IsAbs(string(result)) {
			t.Errorf("expected absolute path, got %q", result)
		}
	})

	t.Run("rejects invalid CfsPath", func(t *testing.T) {
		testutils.Chdir(t)
		_, err := oslayer.CfsPathToOs("../../etc/passwd")
		if !errors.Is(err, oslayer.ErrDirectoryTraversal) {
			t.Errorf("expected ErrDirectoryTraversal, got %v", err)
		}
	})

	t.Run("rejects symlink escaping project root", func(t *testing.T) {
		dir := testutils.Chdir(t)

		outsideDir := t.TempDir()
		outsideFile := filepath.Join(outsideDir, "secret.txt")
		if err := os.WriteFile(outsideFile, []byte("secret"), 0644); err != nil {
			t.Fatal(err)
		}

		if err := os.MkdirAll("internal", 0755); err != nil {
			t.Fatal(err)
		}
		symlinkPath := filepath.Join(dir, "internal", "escape.txt")
		if err := os.Symlink(outsideFile, symlinkPath); err != nil {
			if errors.Is(err, os.ErrInvalid) || strings.Contains(err.Error(), "not supported") {
				t.Skip("symlinks not supported on this platform")
			}
			t.Fatal(err)
		}

		_, err := oslayer.CfsPathToOs("internal/escape.txt")
		if !errors.Is(err, oslayer.ErrResolvesOutsideRoot) {
			t.Errorf("expected ErrResolvesOutsideRoot, got %v", err)
		}
	})

	t.Run("rejects path whose root is a prefix but not a boundary", func(t *testing.T) {
		baseDir := t.TempDir()

		projectDir := filepath.Join(baseDir, "project")
		otherDir := filepath.Join(baseDir, "projectother")

		if err := os.MkdirAll(projectDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(otherDir, 0755); err != nil {
			t.Fatal(err)
		}

		otherFile := filepath.Join(otherDir, "file.txt")
		if err := os.WriteFile(otherFile, []byte("other"), 0644); err != nil {
			t.Fatal(err)
		}

		original, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Chdir(projectDir); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := os.Chdir(original); err != nil {
				t.Errorf("failed to restore working directory: %v", err)
			}
		})

		_, err = oslayer.CfsPathToOs("../projectother/file.txt")
		if !errors.Is(err, oslayer.ErrDirectoryTraversal) && !errors.Is(err, oslayer.ErrResolvesOutsideRoot) {
			t.Errorf("expected ErrDirectoryTraversal or ErrResolvesOutsideRoot, got %v", err)
		}
	})

	t.Run("roundtrip CfsPathToOs then OsPathToCfs", func(t *testing.T) {
		testutils.Chdir(t)
		if err := os.MkdirAll("internal/config", 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("internal/config/config.go", []byte("package config"), 0644); err != nil {
			t.Fatal(err)
		}

		original := oslayer.CfsPath("internal/config/config.go")
		osPath, err := oslayer.CfsPathToOs(original)
		if err != nil {
			t.Fatalf("CfsPathToOs failed: %v", err)
		}
		result, err := oslayer.OsPathToCfs(osPath)
		if err != nil {
			t.Fatalf("OsPathToCfs failed: %v", err)
		}
		if result != original {
			t.Errorf("roundtrip mismatch: got %q, want %q", result, original)
		}
	})
}

func TestOsPathToCfs(t *testing.T) {
	t.Run("converts valid OS path that exists", func(t *testing.T) {
		dir := testutils.Chdir(t)
		if err := os.MkdirAll("internal/config", 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("internal/config/config.go", []byte("package config"), 0644); err != nil {
			t.Fatal(err)
		}
		absPath := oslayer.OsPath(filepath.Join(dir, "internal", "config", "config.go"))
		result, err := oslayer.OsPathToCfs(absPath)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result != "internal/config/config.go" {
			t.Errorf("expected %q, got %q", "internal/config/config.go", result)
		}
	})

	t.Run("converts valid OS path that does not exist", func(t *testing.T) {
		dir := testutils.Chdir(t)
		absPath := oslayer.OsPath(filepath.Join(dir, "internal", "nonexistent", "file.go"))
		result, err := oslayer.OsPathToCfs(absPath)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !strings.HasPrefix(string(result), "internal/") {
			t.Errorf("unexpected result: %q", result)
		}
	})

	t.Run("result uses forward slashes", func(t *testing.T) {
		dir := testutils.Chdir(t)
		absPath := oslayer.OsPath(filepath.Join(dir, "some", "path", "file.go"))
		result, err := oslayer.OsPathToCfs(absPath)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if strings.Contains(string(result), "\\") {
			t.Errorf("result contains backslashes: %q", result)
		}
	})

	t.Run("symlink within root resolving within root", func(t *testing.T) {
		dir := testutils.Chdir(t)
		if err := os.MkdirAll("internal/real", 0755); err != nil {
			t.Fatal(err)
		}
		realFile := filepath.Join(dir, "internal", "real", "file.go")
		if err := os.WriteFile(realFile, []byte("package real"), 0644); err != nil {
			t.Fatal(err)
		}
		symlinkPath := filepath.Join(dir, "internal", "link.go")
		if err := os.Symlink(realFile, symlinkPath); err != nil {
			if errors.Is(err, os.ErrInvalid) || strings.Contains(err.Error(), "not supported") {
				t.Skip("symlinks not supported on this platform")
			}
			t.Fatal(err)
		}
		_, err := oslayer.OsPathToCfs(oslayer.OsPath(symlinkPath))
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("rejects path outside project root", func(t *testing.T) {
		testutils.Chdir(t)
		outsideDir := t.TempDir()
		absPath := oslayer.OsPath(filepath.Join(outsideDir, "secret.txt"))
		_, err := oslayer.OsPathToCfs(absPath)
		if !errors.Is(err, oslayer.ErrResolvesOutsideRoot) {
			t.Errorf("expected ErrResolvesOutsideRoot, got %v", err)
		}
	})

	t.Run("rejects OS path whose root is a prefix but not a boundary", func(t *testing.T) {
		baseDir := t.TempDir()

		projectDir := filepath.Join(baseDir, "project")
		otherDir := filepath.Join(baseDir, "projectother")

		if err := os.MkdirAll(projectDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(otherDir, 0755); err != nil {
			t.Fatal(err)
		}

		otherFile := filepath.Join(otherDir, "file.txt")
		if err := os.WriteFile(otherFile, []byte("other"), 0644); err != nil {
			t.Fatal(err)
		}

		original, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Chdir(projectDir); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := os.Chdir(original); err != nil {
				t.Errorf("failed to restore working directory: %v", err)
			}
		})

		absPath := oslayer.OsPath(otherFile)
		_, err = oslayer.OsPathToCfs(absPath)
		if !errors.Is(err, oslayer.ErrResolvesOutsideRoot) {
			t.Errorf("expected ErrResolvesOutsideRoot, got %v", err)
		}
	})
}

func TestGetProjectRoot(t *testing.T) {
	t.Run("returns an absolute path", func(t *testing.T) {
		testutils.Chdir(t)
		result, err := oslayer.GetProjectRoot()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if string(result) == "" {
			t.Error("expected non-empty path")
		}
		if !filepath.IsAbs(string(result)) {
			t.Errorf("expected absolute path, got %q", result)
		}
	})

	t.Run("matches working directory", func(t *testing.T) {
		dir := testutils.Chdir(t)
		result, err := oslayer.GetProjectRoot()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		wantAbs, err := filepath.EvalSymlinks(dir)
		if err != nil {
			wantAbs = dir
		}
		gotAbs, err := filepath.EvalSymlinks(string(result))
		if err != nil {
			gotAbs = string(result)
		}
		if gotAbs != wantAbs {
			t.Errorf("expected %q, got %q", wantAbs, gotAbs)
		}
	})
}
