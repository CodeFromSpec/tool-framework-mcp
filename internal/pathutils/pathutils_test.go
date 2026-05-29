// code-from-spec: ROOT/golang/tests/os/path_utils@goYPwb2pDOkheYMtYzRgMjKDNRA

package pathutils_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// testChdir changes the working directory to dir and registers a cleanup
// to restore it after the test.
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

// ---------------------------------------------------------------------------
// PathValidateCfs
// ---------------------------------------------------------------------------

func TestPathValidateCfs(t *testing.T) {
	type testCase struct {
		name      string
		input     string
		wantErr   error
		wantNoErr bool
	}

	tests := []testCase{
		// TC-PV-01
		{
			name:      "TC-PV-01: valid simple relative path",
			input:     "internal/config/config.go",
			wantNoErr: true,
		},
		// TC-PV-02
		{
			name:      "TC-PV-02: valid nested path",
			input:     "cmd/framework-mcp/main.go",
			wantNoErr: true,
		},
		// TC-PV-03
		{
			name:      "TC-PV-03: valid single filename",
			input:     "main.go",
			wantNoErr: true,
		},
		// TC-PV-04
		{
			name:      "TC-PV-04: accepts path with dot segment",
			input:     "internal/./config/config.go",
			wantNoErr: true,
		},
		// TC-PV-05
		{
			name:      "TC-PV-05: accepts traversal that resolves within root",
			input:     "a/b/../c",
			wantNoErr: true,
		},
		// TC-PV-06
		{
			name:      "TC-PV-06: accepts path with trailing slash",
			input:     "internal/config/",
			wantNoErr: true,
		},
		// TC-PV-07
		{
			name:      "TC-PV-07: accepts path with duplicate slashes",
			input:     "internal//config//file.go",
			wantNoErr: true,
		},
		// TC-PV-08
		{
			name:    "TC-PV-08: rejects empty string",
			input:   "",
			wantErr: pathutils.ErrPathIsEmpty,
		},
		// TC-PV-09
		{
			name:    "TC-PV-09: rejects absolute path with leading slash",
			input:   "/etc/passwd",
			wantErr: pathutils.ErrPathIsAbsolute,
		},
		// TC-PV-10
		{
			name:    "TC-PV-10: rejects absolute path with drive letter",
			input:   "C:/Windows/system32",
			wantErr: pathutils.ErrPathIsAbsolute,
		},
		// TC-PV-11
		{
			name:    "TC-PV-11: rejects backslash",
			input:   `internal\config\config.go`,
			wantErr: pathutils.ErrPathContainsBackslash,
		},
		// TC-PV-12
		{
			name:    "TC-PV-12: rejects simple traversal",
			input:   "../../etc/passwd",
			wantErr: pathutils.ErrDirectoryTraversal,
		},
		// TC-PV-13
		{
			name:    "TC-PV-13: rejects embedded traversal",
			input:   "internal/../../outside/file.go",
			wantErr: pathutils.ErrDirectoryTraversal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := pathutils.PathValidateCfs(tc.input)
			if tc.wantNoErr {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
				return
			}
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("expected error %v, got: %v", tc.wantErr, err)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// PathCfsToOs
// ---------------------------------------------------------------------------

// TC-CO-01: Converts valid path that exists.
func TestPathCfsToOs_ValidPathExists(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	subDir := filepath.Join(tempDir, "internal", "config")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "config.go"), []byte("package config"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfs := &pathutils.PathCfs{Value: "internal/config/config.go"}
	osPath, err := pathutils.PathCfsToOs(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(osPath.Value) {
		t.Errorf("expected absolute path, got: %s", osPath.Value)
	}
	wantSuffix := filepath.Join("internal", "config", "config.go")
	if !strings.HasSuffix(osPath.Value, wantSuffix) {
		t.Errorf("expected path to end with %q, got: %s", wantSuffix, osPath.Value)
	}
}

// TC-CO-02: Converts valid path that does not exist.
func TestPathCfsToOs_ValidPathNotExist(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfs := &pathutils.PathCfs{Value: "internal/newdir/newfile.go"}
	osPath, err := pathutils.PathCfsToOs(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(osPath.Value) {
		t.Errorf("expected absolute path, got: %s", osPath.Value)
	}
	wantSuffix := filepath.Join("internal", "newdir", "newfile.go")
	if !strings.HasSuffix(osPath.Value, wantSuffix) {
		t.Errorf("expected path to end with %q, got: %s", wantSuffix, osPath.Value)
	}
}

// TC-CO-03: Converts path with duplicate slashes.
func TestPathCfsToOs_DuplicateSlashes(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfs := &pathutils.PathCfs{Value: "internal//config.go"}
	osPath, err := pathutils.PathCfsToOs(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(osPath.Value, "//") || strings.Contains(osPath.Value, `\\`) {
		t.Errorf("expected normalized path without duplicate separators, got: %s", osPath.Value)
	}
}

// TC-CO-04: Rejects invalid CFS path (directory traversal).
func TestPathCfsToOs_RejectsInvalidPath(t *testing.T) {
	cfs := &pathutils.PathCfs{Value: "../../etc/passwd"}
	_, err := pathutils.PathCfsToOs(cfs)
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got: %v", err)
	}
}

// TC-CO-05: Rejects symlink escaping project root.
func TestPathCfsToOs_RejectsSymlinkEscapingRoot(t *testing.T) {
	// Create the project root temp dir and a separate "outside" dir.
	projectDir := t.TempDir()
	outsideDir := t.TempDir()
	testChdir(t, projectDir)

	// Create a symlink inside the project root pointing outside.
	symlinkName := "escape-link"
	symlinkPath := filepath.Join(projectDir, symlinkName)
	if err := os.Symlink(outsideDir, symlinkPath); err != nil {
		t.Skipf("cannot create symlink (may require elevated privileges): %v", err)
	}

	cfs := &pathutils.PathCfs{Value: symlinkName + "/sensitive-file"}
	_, err := pathutils.PathCfsToOs(cfs)
	if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
		t.Errorf("expected ErrResolvesOutsideRoot, got: %v", err)
	}
}

// TC-CO-06: Roundtrip — CfsToOs then OsToCfs.
func TestPathCfsToOs_Roundtrip(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	const cfsValue = "internal/config/config.go"
	cfs := &pathutils.PathCfs{Value: cfsValue}

	osPath, err := pathutils.PathCfsToOs(cfs)
	if err != nil {
		t.Fatalf("PathCfsToOs: %v", err)
	}

	cfsBack, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("PathOsToCfs: %v", err)
	}

	if cfsBack.Value != cfsValue {
		t.Errorf("roundtrip mismatch: got %q, want %q", cfsBack.Value, cfsValue)
	}
}

// ---------------------------------------------------------------------------
// PathOsToCfs
// ---------------------------------------------------------------------------

// TC-OC-01: Converts valid OS path that exists.
func TestPathOsToCfs_ValidPathExists(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	subDir := filepath.Join(tempDir, "mypackage")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	filePath := filepath.Join(subDir, "file.go")
	if err := os.WriteFile(filePath, []byte("package mypackage"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	osPath := &pathutils.PathOs{Value: filePath}
	cfs, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.HasPrefix(cfs.Value, "/") {
		t.Errorf("CFS path must not start with '/': %s", cfs.Value)
	}
	if len(cfs.Value) > 1 && cfs.Value[1] == ':' {
		t.Errorf("CFS path must not contain drive letter: %s", cfs.Value)
	}
	if cfs.Value != "mypackage/file.go" {
		t.Errorf("expected %q, got %q", "mypackage/file.go", cfs.Value)
	}
}

// TC-OC-02: Converts valid OS path that does not exist.
func TestPathOsToCfs_ValidPathNotExist(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	nonExistent := filepath.Join(tempDir, "ghost", "file.go")
	osPath := &pathutils.PathOs{Value: nonExistent}
	cfs, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfs.Value != "ghost/file.go" {
		t.Errorf("expected %q, got %q", "ghost/file.go", cfs.Value)
	}
}

// TC-OC-03: Result uses forward slashes.
func TestPathOsToCfs_ForwardSlashes(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	subPath := filepath.Join(tempDir, "a", "b", "c.go")
	osPath := &pathutils.PathOs{Value: subPath}
	cfs, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(cfs.Value, `\`) {
		t.Errorf("CFS path must not contain backslashes, got: %s", cfs.Value)
	}
}

// TC-OC-04: Symlink within root resolving within root.
func TestPathOsToCfs_SymlinkWithinRoot(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create a real target inside the project root.
	targetDir := filepath.Join(tempDir, "real")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Create a symlink inside the project root pointing to that target.
	symlinkPath := filepath.Join(tempDir, "link-to-real")
	if err := os.Symlink(targetDir, symlinkPath); err != nil {
		t.Skipf("cannot create symlink (may require elevated privileges): %v", err)
	}

	osPath := &pathutils.PathOs{Value: symlinkPath}
	cfs, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfs.Value == "" {
		t.Error("expected non-empty CFS path")
	}
}

// TC-OC-05: Rejects path outside project root.
func TestPathOsToCfs_RejectsPathOutsideRoot(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Use a path that is definitely outside the project root.
	// We pick the parent of the temp dir, which is outside tempDir.
	outsidePath := filepath.Dir(tempDir)
	outsideFile := filepath.Join(outsidePath, "outside-file.txt")

	osPath := &pathutils.PathOs{Value: outsideFile}
	_, err := pathutils.PathOsToCfs(osPath)
	if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
		t.Errorf("expected ErrResolvesOutsideRoot, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// PathGetProjectRoot
// ---------------------------------------------------------------------------

// TC-GR-01: Returns an absolute path.
func TestPathGetProjectRoot_ReturnsAbsolutePath(t *testing.T) {
	root, err := pathutils.PathGetProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root.Value == "" {
		t.Fatal("expected non-empty root path")
	}
	if !filepath.IsAbs(root.Value) {
		t.Errorf("expected absolute path, got: %s", root.Value)
	}
}

// TC-GR-02: Matches working directory.
func TestPathGetProjectRoot_MatchesWorkingDirectory(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}

	root, err := pathutils.PathGetProjectRoot()
	if err != nil {
		t.Fatalf("PathGetProjectRoot: %v", err)
	}

	// Evaluate symlinks on both to get canonical paths for comparison.
	canonicalCwd, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		canonicalCwd = cwd
	}
	canonicalRoot, err := filepath.EvalSymlinks(root.Value)
	if err != nil {
		canonicalRoot = root.Value
	}

	if canonicalRoot != canonicalCwd {
		t.Errorf("PathGetProjectRoot returned %q, want %q", canonicalRoot, canonicalCwd)
	}
}
