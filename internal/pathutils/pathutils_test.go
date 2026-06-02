// code-from-spec: ROOT/golang/tests/os/path_utils@lP0to4UdtY82WCe8TKPh2usTdPI
package pathutils_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
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
			wantErr: pathutils.ErrPathEmpty,
		},
		{
			name:    "rejects absolute path with leading slash",
			input:   "/etc/passwd",
			wantErr: pathutils.ErrPathAbsolute,
		},
		{
			name:    "rejects absolute path with drive letter",
			input:   "C:/Windows/system32",
			wantErr: pathutils.ErrPathAbsolute,
		},
		{
			name:    "rejects backslash",
			input:   `internal\config\config.go`,
			wantErr: pathutils.ErrPathContainsBackslash,
		},
		{
			name:    "rejects simple traversal",
			input:   "../../etc/passwd",
			wantErr: pathutils.ErrDirectoryTraversal,
		},
		{
			name:    "rejects embedded traversal",
			input:   "internal/../../outside/file.go",
			wantErr: pathutils.ErrDirectoryTraversal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := pathutils.PathValidateCfs(tc.input)
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected error %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestPathCfsToOs(t *testing.T) {
	t.Run("converts valid path that exists", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		if err := os.MkdirAll("internal/config", 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile("internal/config/config.go", []byte("package config"), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		result, err := pathutils.PathCfsToOs(&pathutils.PathCfs{Value: "internal/config/config.go"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !filepath.IsAbs(result.Value) {
			t.Fatalf("expected absolute path, got %q", result.Value)
		}
		if !strings.HasSuffix(filepath.ToSlash(result.Value), "internal/config/config.go") {
			t.Fatalf("expected path to end with internal/config/config.go, got %q", result.Value)
		}
	})

	t.Run("converts valid path that does not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		result, err := pathutils.PathCfsToOs(&pathutils.PathCfs{Value: "internal/newdir/newfile.go"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !filepath.IsAbs(result.Value) {
			t.Fatalf("expected absolute path, got %q", result.Value)
		}
		if !strings.HasSuffix(filepath.ToSlash(result.Value), "internal/newdir/newfile.go") {
			t.Fatalf("expected path to end with internal/newdir/newfile.go, got %q", result.Value)
		}
	})

	t.Run("converts path with duplicate slashes", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		result, err := pathutils.PathCfsToOs(&pathutils.PathCfs{Value: "internal//config.go"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !filepath.IsAbs(result.Value) {
			t.Fatalf("expected absolute path, got %q", result.Value)
		}
	})

	t.Run("rejects invalid CfsPath", func(t *testing.T) {
		_, err := pathutils.PathCfsToOs(&pathutils.PathCfs{Value: "../../etc/passwd"})
		if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
			t.Fatalf("expected ErrDirectoryTraversal, got %v", err)
		}
	})

	t.Run("rejects symlink escaping project root", func(t *testing.T) {
		tempDir := t.TempDir()
		outsideDir := t.TempDir()
		testChdir(t, tempDir)

		if err := os.MkdirAll("linkparent", 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		linkPath := filepath.Join(tempDir, "linkparent", "escape")
		if err := os.Symlink(outsideDir, linkPath); err != nil {
			t.Skip("symlinks not supported on this platform")
		}

		_, err := pathutils.PathCfsToOs(&pathutils.PathCfs{Value: "linkparent/escape/secret.txt"})
		if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
			t.Fatalf("expected ErrResolvesOutsideRoot, got %v", err)
		}
	})

	t.Run("roundtrip: CfsToOs then OsToCfs", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		if err := os.MkdirAll("internal/config", 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile("internal/config/config.go", []byte("package config"), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		osPath, err := pathutils.PathCfsToOs(&pathutils.PathCfs{Value: "internal/config/config.go"})
		if err != nil {
			t.Fatalf("PathCfsToOs: %v", err)
		}

		cfsPath, err := pathutils.PathOsToCfs(osPath)
		if err != nil {
			t.Fatalf("PathOsToCfs: %v", err)
		}

		if cfsPath.Value != "internal/config/config.go" {
			t.Fatalf("expected %q, got %q", "internal/config/config.go", cfsPath.Value)
		}
	})
}

func TestPathOsToCfs(t *testing.T) {
	t.Run("converts valid OS path that exists", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		if err := os.MkdirAll("mydir", 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile("mydir/file.go", []byte("package mydir"), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		absPath := filepath.Join(tempDir, "mydir", "file.go")
		result, err := pathutils.PathOsToCfs(&pathutils.PathOs{Value: absPath})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Value != "mydir/file.go" {
			t.Fatalf("expected %q, got %q", "mydir/file.go", result.Value)
		}
	})

	t.Run("converts valid OS path that does not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		absPath := filepath.Join(tempDir, "nonexistent", "file.go")
		result, err := pathutils.PathOsToCfs(&pathutils.PathOs{Value: absPath})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Value != "nonexistent/file.go" {
			t.Fatalf("expected %q, got %q", "nonexistent/file.go", result.Value)
		}
	})

	t.Run("result uses forward slashes", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		if err := os.MkdirAll("a/b/c", 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile("a/b/c/file.go", []byte("package c"), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		absPath := filepath.Join(tempDir, "a", "b", "c", "file.go")
		result, err := pathutils.PathOsToCfs(&pathutils.PathOs{Value: absPath})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(result.Value, `\`) {
			t.Fatalf("result contains backslash: %q", result.Value)
		}
	})

	t.Run("symlink within root resolving within root", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		if err := os.MkdirAll("real/dir", 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile("real/dir/file.go", []byte("package dir"), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		linkPath := filepath.Join(tempDir, "link")
		realPath := filepath.Join(tempDir, "real")
		if err := os.Symlink(realPath, linkPath); err != nil {
			t.Skip("symlinks not supported on this platform")
		}

		absPath := filepath.Join(tempDir, "link", "dir", "file.go")
		_, err := pathutils.PathOsToCfs(&pathutils.PathOs{Value: absPath})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("rejects path outside project root", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		outsideDir := t.TempDir()
		absPath := filepath.Join(outsideDir, "secret.txt")

		_, err := pathutils.PathOsToCfs(&pathutils.PathOs{Value: absPath})
		if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
			t.Fatalf("expected ErrResolvesOutsideRoot, got %v", err)
		}
	})
}

func TestPathGetProjectRoot(t *testing.T) {
	t.Run("returns an absolute path", func(t *testing.T) {
		result, err := pathutils.PathGetProjectRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Value == "" {
			t.Fatal("expected non-empty path")
		}
		if !filepath.IsAbs(result.Value) {
			t.Fatalf("expected absolute path, got %q", result.Value)
		}
	})

	t.Run("matches working directory", func(t *testing.T) {
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Getwd: %v", err)
		}

		result, err := pathutils.PathGetProjectRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		cwdResolved, err := filepath.EvalSymlinks(cwd)
		if err != nil {
			cwdResolved = cwd
		}
		resultResolved, err := filepath.EvalSymlinks(result.Value)
		if err != nil {
			resultResolved = result.Value
		}

		if cwdResolved != resultResolved {
			t.Fatalf("expected %q, got %q", cwdResolved, resultResolved)
		}
	})
}
