// code-from-spec: ROOT/golang/tests/os/path_utils@ruhBpVP5vmi7GmFkUUhOulg01As
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
	testCases := []struct {
		name      string
		input     string
		wantErr   error
		wantNoErr bool
	}{
		{
			name:      "valid simple relative path",
			input:     "internal/config/config.go",
			wantNoErr: true,
		},
		{
			name:      "valid nested path",
			input:     "cmd/framework-mcp/main.go",
			wantNoErr: true,
		},
		{
			name:      "valid single filename",
			input:     "main.go",
			wantNoErr: true,
		},
		{
			name:      "accepts path with dot segment",
			input:     "internal/./config/config.go",
			wantNoErr: true,
		},
		{
			name:      "accepts traversal that resolves within root",
			input:     "a/b/../c",
			wantNoErr: true,
		},
		{
			name:      "accepts path with trailing slash",
			input:     "internal/config/",
			wantNoErr: true,
		},
		{
			name:      "accepts path with duplicate slashes",
			input:     "internal//config//file.go",
			wantNoErr: true,
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := pathutils.PathValidateCfs(tc.input)
			if tc.wantNoErr {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				return
			}
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("expected error %v, got %v", tc.wantErr, err)
				}
			}
		})
	}
}

func TestPathCfsToOs(t *testing.T) {
	t.Run("converts valid path that exists", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		if err := os.MkdirAll("internal/config", 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile("internal/config/config.go", []byte("package config"), 0644); err != nil {
			t.Fatalf("writefile: %v", err)
		}

		cfsPath := &pathutils.PathCfs{Value: "internal/config/config.go"}
		osPath, err := pathutils.PathCfsToOs(cfsPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !filepath.IsAbs(osPath.Value) {
			t.Errorf("expected absolute path, got %q", osPath.Value)
		}
		if !strings.HasSuffix(osPath.Value, filepath.Join("internal", "config", "config.go")) {
			t.Errorf("path %q does not end with expected suffix", osPath.Value)
		}
	})

	t.Run("converts valid path that does not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		cfsPath := &pathutils.PathCfs{Value: "internal/newdir/newfile.go"}
		osPath, err := pathutils.PathCfsToOs(cfsPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !filepath.IsAbs(osPath.Value) {
			t.Errorf("expected absolute path, got %q", osPath.Value)
		}
		if !strings.HasSuffix(osPath.Value, filepath.Join("internal", "newdir", "newfile.go")) {
			t.Errorf("path %q does not end with expected suffix", osPath.Value)
		}
	})

	t.Run("converts path with duplicate slashes", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		cfsPath := &pathutils.PathCfs{Value: "internal//config.go"}
		osPath, err := pathutils.PathCfsToOs(cfsPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !filepath.IsAbs(osPath.Value) {
			t.Errorf("expected absolute path, got %q", osPath.Value)
		}
	})

	t.Run("rejects invalid CfsPath", func(t *testing.T) {
		cfsPath := &pathutils.PathCfs{Value: "../../etc/passwd"}
		_, err := pathutils.PathCfsToOs(cfsPath)
		if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
			t.Errorf("expected ErrDirectoryTraversal, got %v", err)
		}
	})

	t.Run("rejects symlink escaping project root", func(t *testing.T) {
		tempDir := t.TempDir()
		outsideDir := t.TempDir()
		testChdir(t, tempDir)

		outsideFile := filepath.Join(outsideDir, "secret.txt")
		if err := os.WriteFile(outsideFile, []byte("secret"), 0644); err != nil {
			t.Fatalf("writefile: %v", err)
		}

		if err := os.Symlink(outsideFile, filepath.Join(tempDir, "symlink.txt")); err != nil {
			t.Skip("symlinks not supported on this platform")
		}

		cfsPath := &pathutils.PathCfs{Value: "symlink.txt"}
		_, err := pathutils.PathCfsToOs(cfsPath)
		if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
			t.Errorf("expected ErrResolvesOutsideRoot, got %v", err)
		}
	})

	t.Run("roundtrip CfsToOs then OsToCfs", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		cfsPath := &pathutils.PathCfs{Value: "internal/config/config.go"}
		osPath, err := pathutils.PathCfsToOs(cfsPath)
		if err != nil {
			t.Fatalf("PathCfsToOs: %v", err)
		}

		backCfs, err := pathutils.PathOsToCfs(osPath)
		if err != nil {
			t.Fatalf("PathOsToCfs: %v", err)
		}

		if backCfs.Value != "internal/config/config.go" {
			t.Errorf("expected %q, got %q", "internal/config/config.go", backCfs.Value)
		}
	})
}

func TestPathOsToCfs(t *testing.T) {
	t.Run("converts valid OS path that exists", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		if err := os.MkdirAll("subdir", 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile("subdir/file.go", []byte("package subdir"), 0644); err != nil {
			t.Fatalf("writefile: %v", err)
		}

		absPath := filepath.Join(tempDir, "subdir", "file.go")
		osPath := &pathutils.PathOs{Value: absPath}

		cfsPath, err := pathutils.PathOsToCfs(osPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfsPath.Value != "subdir/file.go" {
			t.Errorf("expected %q, got %q", "subdir/file.go", cfsPath.Value)
		}
	})

	t.Run("converts valid OS path that does not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		absPath := filepath.Join(tempDir, "nonexistent", "file.go")
		osPath := &pathutils.PathOs{Value: absPath}

		cfsPath, err := pathutils.PathOsToCfs(osPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfsPath.Value != "nonexistent/file.go" {
			t.Errorf("expected %q, got %q", "nonexistent/file.go", cfsPath.Value)
		}
	})

	t.Run("result uses forward slashes", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		absPath := filepath.Join(tempDir, "some", "nested", "file.go")
		osPath := &pathutils.PathOs{Value: absPath}

		cfsPath, err := pathutils.PathOsToCfs(osPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(cfsPath.Value, `\`) {
			t.Errorf("PathCfs contains backslash: %q", cfsPath.Value)
		}
	})

	t.Run("symlink within root resolving within root", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		if err := os.WriteFile("target.go", []byte("package main"), 0644); err != nil {
			t.Fatalf("writefile: %v", err)
		}

		symlinkPath := filepath.Join(tempDir, "link.go")
		targetPath := filepath.Join(tempDir, "target.go")
		if err := os.Symlink(targetPath, symlinkPath); err != nil {
			t.Skip("symlinks not supported on this platform")
		}

		osPath := &pathutils.PathOs{Value: symlinkPath}
		_, err := pathutils.PathOsToCfs(osPath)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("rejects path outside project root", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		outsideDir := t.TempDir()
		outsidePath := filepath.Join(outsideDir, "outside.go")
		osPath := &pathutils.PathOs{Value: outsidePath}

		_, err := pathutils.PathOsToCfs(osPath)
		if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
			t.Errorf("expected ErrResolvesOutsideRoot, got %v", err)
		}
	})
}

func TestPathGetProjectRoot(t *testing.T) {
	t.Run("returns an absolute path", func(t *testing.T) {
		root, err := pathutils.PathGetProjectRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if root == nil {
			t.Fatal("got nil PathOs")
		}
		if root.Value == "" {
			t.Error("got empty path")
		}
		if !filepath.IsAbs(root.Value) {
			t.Errorf("expected absolute path, got %q", root.Value)
		}
	})

	t.Run("matches working directory", func(t *testing.T) {
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("os.Getwd: %v", err)
		}

		root, err := pathutils.PathGetProjectRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wdResolved, err := filepath.EvalSymlinks(wd)
		if err != nil {
			wdResolved = wd
		}
		rootResolved, err := filepath.EvalSymlinks(root.Value)
		if err != nil {
			rootResolved = root.Value
		}

		if wdResolved != rootResolved {
			t.Errorf("expected %q, got %q", wdResolved, rootResolved)
		}
	})
}
