// code-from-spec: ROOT/golang/tests/os/path_utils@2XiBvXcJTme3vVNTM9k6_2aWMjg
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
					t.Errorf("unexpected error: %v", err)
				}
			} else {
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
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile("internal/config/config.go", []byte("package config"), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		cfsPath := &pathutils.PathCfs{Value: "internal/config/config.go"}
		osPath, err := pathutils.PathCfsToOs(cfsPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if osPath == nil {
			t.Fatal("expected non-nil PathOs")
		}
		if !filepath.IsAbs(osPath.Value) {
			t.Errorf("expected absolute path, got %q", osPath.Value)
		}
		suffix := filepath.Join("internal", "config", "config.go")
		if !strings.HasSuffix(osPath.Value, suffix) {
			t.Errorf("path %q does not end with %q", osPath.Value, suffix)
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
		if osPath == nil {
			t.Fatal("expected non-nil PathOs")
		}
		if !filepath.IsAbs(osPath.Value) {
			t.Errorf("expected absolute path, got %q", osPath.Value)
		}
		suffix := filepath.Join("internal", "newdir", "newfile.go")
		if !strings.HasSuffix(osPath.Value, suffix) {
			t.Errorf("path %q does not end with %q", osPath.Value, suffix)
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
		if osPath == nil {
			t.Fatal("expected non-nil PathOs")
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

		outsideFile := filepath.Join(outsideDir, "secret.txt")
		if err := os.WriteFile(outsideFile, []byte("secret"), 0644); err != nil {
			t.Fatalf("WriteFile outside: %v", err)
		}

		testChdir(t, tempDir)

		linkName := "escaping-link"
		if err := os.Symlink(outsideFile, linkName); err != nil {
			t.Skipf("symlinks not supported: %v", err)
		}

		cfsPath := &pathutils.PathCfs{Value: linkName}
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

		backToCfs, err := pathutils.PathOsToCfs(osPath)
		if err != nil {
			t.Fatalf("PathOsToCfs: %v", err)
		}
		if backToCfs.Value != "internal/config/config.go" {
			t.Errorf("expected %q, got %q", "internal/config/config.go", backToCfs.Value)
		}
	})
}

func TestPathOsToCfs(t *testing.T) {
	t.Run("converts valid OS path that exists", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		if err := os.WriteFile("somefile.go", []byte("package main"), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		absPath := filepath.Join(tempDir, "somefile.go")
		osPath := &pathutils.PathOs{Value: absPath}
		cfsPath, err := pathutils.PathOsToCfs(osPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfsPath == nil {
			t.Fatal("expected non-nil PathCfs")
		}
		if strings.ContainsRune(cfsPath.Value, '/') == false && cfsPath.Value != "somefile.go" {
			t.Errorf("unexpected CFS path %q", cfsPath.Value)
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
		if cfsPath == nil {
			t.Fatal("expected non-nil PathCfs")
		}
		if !strings.HasSuffix(cfsPath.Value, "nonexistent/file.go") {
			t.Errorf("unexpected CFS path %q", cfsPath.Value)
		}
	})

	t.Run("result uses forward slashes", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		if err := os.MkdirAll(filepath.Join("a", "b"), 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}

		absPath := filepath.Join(tempDir, "a", "b", "file.go")
		osPath := &pathutils.PathOs{Value: absPath}
		cfsPath, err := pathutils.PathOsToCfs(osPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(cfsPath.Value, `\`) {
			t.Errorf("CFS path contains backslash: %q", cfsPath.Value)
		}
	})

	t.Run("symlink within root resolving within root", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		if err := os.WriteFile("target.go", []byte("package main"), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		linkPath := filepath.Join(tempDir, "link.go")
		targetPath := filepath.Join(tempDir, "target.go")
		if err := os.Symlink(targetPath, linkPath); err != nil {
			t.Skipf("symlinks not supported: %v", err)
		}

		osPath := &pathutils.PathOs{Value: linkPath}
		cfsPath, err := pathutils.PathOsToCfs(osPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfsPath == nil {
			t.Fatal("expected non-nil PathCfs")
		}
		if strings.Contains(cfsPath.Value, `\`) {
			t.Errorf("CFS path contains backslash: %q", cfsPath.Value)
		}
	})

	t.Run("rejects path outside project root", func(t *testing.T) {
		tempDir := t.TempDir()
		testChdir(t, tempDir)

		outsideDir := t.TempDir()
		absPath := filepath.Join(outsideDir, "outside.go")

		osPath := &pathutils.PathOs{Value: absPath}
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
			t.Fatal("expected non-nil PathOs")
		}
		if root.Value == "" {
			t.Error("expected non-empty path")
		}
		if !filepath.IsAbs(root.Value) {
			t.Errorf("expected absolute path, got %q", root.Value)
		}
	})

	t.Run("matches working directory", func(t *testing.T) {
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Getwd: %v", err)
		}

		root, err := pathutils.PathGetProjectRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wdEval, err := filepath.EvalSymlinks(wd)
		if err != nil {
			wdEval = wd
		}
		rootEval, err := filepath.EvalSymlinks(root.Value)
		if err != nil {
			rootEval = root.Value
		}

		if wdEval != rootEval {
			t.Errorf("expected %q, got %q", wdEval, rootEval)
		}
	})
}
