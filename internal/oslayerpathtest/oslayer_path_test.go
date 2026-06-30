// code-from-spec: SPEC/golang/test/cases/oslayer/path@5Y5IjjuIMVDmhW-xppS5zVCjPq0
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
			t.Errorf("expected absolute path, got %q", result)
		}
		if !strings.HasSuffix(string(result), filepath.Join("internal", "config", "config.go")) {
			t.Errorf("unexpected path %q", result)
		}
	})

	t.Run("converts valid path that does not exist", func(t *testing.T) {
		testutils.Chdir(t)
		result, err := oslayer.CfsPathToOs("internal/newdir/newfile.go")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !filepath.IsAbs(string(result)) {
			t.Errorf("expected absolute path, got %q", result)
		}
	})

	t.Run("converts path with duplicate slashes", func(t *testing.T) {
		testutils.Chdir(t)
		result, err := oslayer.CfsPathToOs("internal//config.go")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
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
		err := os.WriteFile(outsideFile, []byte("secret"), 0644)
		if err != nil {
			t.Fatalf("write outside file: %v", err)
		}
		err = os.MkdirAll(filepath.Join(dir, "internal"), 0755)
		if err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		linkPath := filepath.Join(dir, "internal", "symlink.txt")
		err = os.Symlink(outsideFile, linkPath)
		if err != nil {
			t.Skipf("symlink not supported: %v", err)
		}
		_, err = oslayer.CfsPathToOs("internal/symlink.txt")
		if !errors.Is(err, oslayer.ErrResolvesOutsideRoot) {
			t.Errorf("expected ErrResolvesOutsideRoot, got %v", err)
		}
	})

	t.Run("rejects path whose root is a prefix but not a boundary", func(t *testing.T) {
		parent := t.TempDir()
		projectDir := filepath.Join(parent, "project")
		otherDir := filepath.Join(parent, "projectother")
		err := os.MkdirAll(projectDir, 0755)
		if err != nil {
			t.Fatalf("mkdir project: %v", err)
		}
		err = os.MkdirAll(otherDir, 0755)
		if err != nil {
			t.Fatalf("mkdir projectother: %v", err)
		}
		otherFile := filepath.Join(otherDir, "file.txt")
		err = os.WriteFile(otherFile, []byte("data"), 0644)
		if err != nil {
			t.Fatalf("write file: %v", err)
		}
		original, err := os.Getwd()
		if err != nil {
			t.Fatalf("getwd: %v", err)
		}
		err = os.Chdir(projectDir)
		if err != nil {
			t.Fatalf("chdir: %v", err)
		}
		t.Cleanup(func() {
			if cerr := os.Chdir(original); cerr != nil {
				t.Errorf("cleanup chdir: %v", cerr)
			}
		})
		_, err = oslayer.CfsPathToOs("../projectother/file.txt")
		if !errors.Is(err, oslayer.ErrDirectoryTraversal) && !errors.Is(err, oslayer.ErrResolvesOutsideRoot) {
			t.Errorf("expected ErrDirectoryTraversal or ErrResolvesOutsideRoot, got %v", err)
		}
	})

	t.Run("roundtrip CfsPathToOs then OsPathToCfs", func(t *testing.T) {
		testutils.Chdir(t)
		err := os.MkdirAll("internal/config", 0755)
		if err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		err = os.WriteFile("internal/config/config.go", []byte("package config"), 0644)
		if err != nil {
			t.Fatalf("write: %v", err)
		}
		osPath, err := oslayer.CfsPathToOs("internal/config/config.go")
		if err != nil {
			t.Fatalf("CfsPathToOs: %v", err)
		}
		cfsPath, err := oslayer.OsPathToCfs(osPath)
		if err != nil {
			t.Fatalf("OsPathToCfs: %v", err)
		}
		if cfsPath != "internal/config/config.go" {
			t.Errorf("roundtrip mismatch: got %q, want %q", cfsPath, "internal/config/config.go")
		}
	})
}

func TestOsPathToCfs(t *testing.T) {
	t.Run("converts valid OS path that exists", func(t *testing.T) {
		dir := testutils.Chdir(t)
		err := os.MkdirAll("subdir", 0755)
		if err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		filePath := filepath.Join(dir, "subdir", "file.go")
		err = os.WriteFile(filePath, []byte("package x"), 0644)
		if err != nil {
			t.Fatalf("write: %v", err)
		}
		result, err := oslayer.OsPathToCfs(oslayer.OsPath(filePath))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(string(result), "/") || strings.Contains(string(result), "\\") {
			t.Errorf("expected forward slashes, got %q", result)
		}
		if result != "subdir/file.go" {
			t.Errorf("expected %q, got %q", "subdir/file.go", result)
		}
	})

	t.Run("converts valid OS path that does not exist", func(t *testing.T) {
		dir := testutils.Chdir(t)
		nonexistentPath := filepath.Join(dir, "nonexistent", "file.go")
		result, err := oslayer.OsPathToCfs(oslayer.OsPath(nonexistentPath))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !filepath.IsAbs(string(oslayer.OsPath(nonexistentPath))) {
			t.Errorf("input was not absolute")
		}
		if strings.Contains(string(result), "\\") {
			t.Errorf("expected forward slashes only, got %q", result)
		}
	})

	t.Run("result uses forward slashes", func(t *testing.T) {
		dir := testutils.Chdir(t)
		err := os.MkdirAll("a/b/c", 0755)
		if err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		filePath := filepath.Join(dir, "a", "b", "c", "file.go")
		err = os.WriteFile(filePath, []byte("package x"), 0644)
		if err != nil {
			t.Fatalf("write: %v", err)
		}
		result, err := oslayer.OsPathToCfs(oslayer.OsPath(filePath))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(string(result), "\\") {
			t.Errorf("result contains backslash: %q", result)
		}
	})

	t.Run("symlink within root resolving within root", func(t *testing.T) {
		dir := testutils.Chdir(t)
		err := os.MkdirAll("real", 0755)
		if err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		realFile := filepath.Join(dir, "real", "file.go")
		err = os.WriteFile(realFile, []byte("package x"), 0644)
		if err != nil {
			t.Fatalf("write: %v", err)
		}
		linkPath := filepath.Join(dir, "link.go")
		err = os.Symlink(realFile, linkPath)
		if err != nil {
			t.Skipf("symlink not supported: %v", err)
		}
		_, err = oslayer.OsPathToCfs(oslayer.OsPath(linkPath))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("rejects path outside project root", func(t *testing.T) {
		testutils.Chdir(t)
		outsidePath := t.TempDir()
		outsideFile := filepath.Join(outsidePath, "outside.go")
		err := os.WriteFile(outsideFile, []byte("package x"), 0644)
		if err != nil {
			t.Fatalf("write: %v", err)
		}
		_, err = oslayer.OsPathToCfs(oslayer.OsPath(outsideFile))
		if !errors.Is(err, oslayer.ErrResolvesOutsideRoot) {
			t.Errorf("expected ErrResolvesOutsideRoot, got %v", err)
		}
	})

	t.Run("rejects OS path whose root is a prefix but not a boundary", func(t *testing.T) {
		parent := t.TempDir()
		projectDir := filepath.Join(parent, "project")
		otherDir := filepath.Join(parent, "projectother")
		err := os.MkdirAll(projectDir, 0755)
		if err != nil {
			t.Fatalf("mkdir project: %v", err)
		}
		err = os.MkdirAll(otherDir, 0755)
		if err != nil {
			t.Fatalf("mkdir projectother: %v", err)
		}
		otherFile := filepath.Join(otherDir, "file.txt")
		err = os.WriteFile(otherFile, []byte("data"), 0644)
		if err != nil {
			t.Fatalf("write file: %v", err)
		}
		original, err := os.Getwd()
		if err != nil {
			t.Fatalf("getwd: %v", err)
		}
		err = os.Chdir(projectDir)
		if err != nil {
			t.Fatalf("chdir: %v", err)
		}
		t.Cleanup(func() {
			if cerr := os.Chdir(original); cerr != nil {
				t.Errorf("cleanup chdir: %v", cerr)
			}
		})
		_, err = oslayer.OsPathToCfs(oslayer.OsPath(otherFile))
		if !errors.Is(err, oslayer.ErrResolvesOutsideRoot) {
			t.Errorf("expected ErrResolvesOutsideRoot, got %v", err)
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
			t.Fatalf("unexpected error: %v", err)
		}
		evalDir, err := filepath.EvalSymlinks(dir)
		if err != nil {
			evalDir = dir
		}
		evalResult, err := filepath.EvalSymlinks(string(result))
		if err != nil {
			evalResult = string(result)
		}
		if evalResult != evalDir {
			t.Errorf("expected %q, got %q", evalDir, evalResult)
		}
	})
}
