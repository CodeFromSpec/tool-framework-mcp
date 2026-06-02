// code-from-spec: ROOT/golang/tests/os/path_utils@DSnqhHzBJVM1Gomslrm8qlEciuQ
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

func TestPathCfsToOs(t *testing.T) {
	t.Run("converts valid path that exists", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		dir := filepath.Join(tempDir, "internal", "config")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "config.go"), []byte("package config"), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		result, err := pathutils.PathCfsToOs(&pathutils.PathCfs{Value: "internal/config/config.go"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !filepath.IsAbs(result.Value) {
			t.Errorf("expected absolute path, got: %s", result.Value)
		}
		suffix := filepath.Join("internal", "config", "config.go")
		if !strings.HasSuffix(result.Value, suffix) {
			t.Errorf("expected path to end with %s, got: %s", suffix, result.Value)
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
			t.Errorf("expected absolute path, got: %s", result.Value)
		}
		suffix := filepath.Join("internal", "newdir", "newfile.go")
		if !strings.HasSuffix(result.Value, suffix) {
			t.Errorf("expected path to end with %s, got: %s", suffix, result.Value)
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
			t.Errorf("expected absolute path, got: %s", result.Value)
		}
	})

	t.Run("rejects invalid CfsPath", func(t *testing.T) {
		_, err := pathutils.PathCfsToOs(&pathutils.PathCfs{Value: "../../etc/passwd"})
		if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
			t.Errorf("expected ErrDirectoryTraversal, got: %v", err)
		}
	})

	t.Run("rejects symlink escaping project root", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		outsideDir := t.TempDir()
		outsideFile := filepath.Join(outsideDir, "secret.txt")
		if err := os.WriteFile(outsideFile, []byte("secret"), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		symlinkDir := filepath.Join(tempDir, "links")
		if err := os.MkdirAll(symlinkDir, 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		symlinkPath := filepath.Join(symlinkDir, "escape")
		if err := os.Symlink(outsideDir, symlinkPath); err != nil {
			t.Skip("symlinks not supported on this platform")
		}

		_, err := pathutils.PathCfsToOs(&pathutils.PathCfs{Value: "links/escape/secret.txt"})
		if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
			t.Errorf("expected ErrResolvesOutsideRoot, got: %v", err)
		}
	})

	t.Run("roundtrip: CfsToOs then OsToCfs", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		dir := filepath.Join(tempDir, "internal", "config")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "config.go"), []byte("package config"), 0644); err != nil {
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
			t.Errorf("expected internal/config/config.go, got: %s", cfsPath.Value)
		}
	})
}

func TestPathOsToCfs(t *testing.T) {
	t.Run("converts valid OS path that exists", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		dir := filepath.Join(tempDir, "mydir")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		filePath := filepath.Join(dir, "myfile.go")
		if err := os.WriteFile(filePath, []byte("package mydir"), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		result, err := pathutils.PathOsToCfs(&pathutils.PathOs{Value: filePath})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Value != "mydir/myfile.go" {
			t.Errorf("expected mydir/myfile.go, got: %s", result.Value)
		}
	})

	t.Run("converts valid OS path that does not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		nonExistentPath := filepath.Join(tempDir, "nonexistent", "file.go")

		result, err := pathutils.PathOsToCfs(&pathutils.PathOs{Value: nonExistentPath})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Value != "nonexistent/file.go" {
			t.Errorf("expected nonexistent/file.go, got: %s", result.Value)
		}
	})

	t.Run("result uses forward slashes", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		dir := filepath.Join(tempDir, "a", "b", "c")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		filePath := filepath.Join(dir, "file.go")
		if err := os.WriteFile(filePath, []byte("package c"), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		result, err := pathutils.PathOsToCfs(&pathutils.PathOs{Value: filePath})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(result.Value, `\`) {
			t.Errorf("expected no backslashes, got: %s", result.Value)
		}
	})

	t.Run("symlink within root resolving within root", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		dir := filepath.Join(tempDir, "realdir")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		realFile := filepath.Join(dir, "real.go")
		if err := os.WriteFile(realFile, []byte("package realdir"), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		symlinkPath := filepath.Join(tempDir, "link.go")
		if err := os.Symlink(realFile, symlinkPath); err != nil {
			t.Skip("symlinks not supported on this platform")
		}

		result, err := pathutils.PathOsToCfs(&pathutils.PathOs{Value: symlinkPath})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Value == "" {
			t.Error("expected non-empty PathCfs value")
		}
	})

	t.Run("rejects path outside project root", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		outsideDir := t.TempDir()
		outsidePath := filepath.Join(outsideDir, "file.go")

		_, err := pathutils.PathOsToCfs(&pathutils.PathOs{Value: outsidePath})
		if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
			t.Errorf("expected ErrResolvesOutsideRoot, got: %v", err)
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
			t.Error("expected non-empty path")
		}
		if !filepath.IsAbs(result.Value) {
			t.Errorf("expected absolute path, got: %s", result.Value)
		}
	})

	t.Run("matches working directory", func(t *testing.T) {
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Getwd: %v", err)
		}

		result, err := pathutils.PathGetProjectRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wdResolved, err := filepath.EvalSymlinks(wd)
		if err != nil {
			wdResolved = wd
		}
		resultResolved, err := filepath.EvalSymlinks(result.Value)
		if err != nil {
			resultResolved = result.Value
		}

		if resultResolved != wdResolved {
			t.Errorf("expected %s, got: %s", wdResolved, resultResolved)
		}
	})
}
