// code-from-spec: ROOT/golang/tests/os/path_utils@WCt93u3AQg1hQ0pkyHNG0EQuSnw
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
// function to restore the original directory.
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
		name        string
		input       string
		wantErr     error
		wantNoError bool
	}

	cases := []testCase{
		// TC-PV-01
		{
			name:        "TC-PV-01: valid simple relative path",
			input:       "internal/config/config.go",
			wantNoError: true,
		},
		// TC-PV-02
		{
			name:        "TC-PV-02: valid nested path",
			input:       "cmd/framework-mcp/main.go",
			wantNoError: true,
		},
		// TC-PV-03
		{
			name:        "TC-PV-03: valid single filename",
			input:       "main.go",
			wantNoError: true,
		},
		// TC-PV-04
		{
			name:        "TC-PV-04: accepts path with dot segment",
			input:       "internal/./config/config.go",
			wantNoError: true,
		},
		// TC-PV-05
		{
			name:        "TC-PV-05: accepts traversal that resolves within root",
			input:       "a/b/../c",
			wantNoError: true,
		},
		// TC-PV-06
		{
			name:        "TC-PV-06: accepts path with trailing slash",
			input:       "internal/config/",
			wantNoError: true,
		},
		// TC-PV-07
		{
			name:        "TC-PV-07: accepts path with duplicate slashes",
			input:       "internal//config//file.go",
			wantNoError: true,
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

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := pathutils.PathValidateCfs(tc.input)
			if tc.wantNoError {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error %v, got nil", tc.wantErr)
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("expected error %v, got %v", tc.wantErr, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// PathCfsToOs
// ---------------------------------------------------------------------------

// TC-CO-01: Converts valid path that exists
func TestPathCfsToOs_ValidPathExists(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create the file so the path exists.
	if err := os.MkdirAll("internal/config", 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile("internal/config/config.go", []byte("package config"), 0644); err != nil {
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
		t.Errorf("expected path ending with %q, got %q", wantSuffix, osPath.Value)
	}
}

// TC-CO-02: Converts valid path that does not exist
func TestPathCfsToOs_ValidPathNotExists(t *testing.T) {
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
		t.Errorf("expected path ending with %q, got %q", wantSuffix, osPath.Value)
	}
}

// TC-CO-03: Converts path with duplicate slashes
func TestPathCfsToOs_DuplicateSlashes(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfs := &pathutils.PathCfs{Value: "internal//config.go"}
	osPath, err := pathutils.PathCfsToOs(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(osPath.Value, string([]rune{filepath.Separator, filepath.Separator})) {
		t.Errorf("expected normalized path with no duplicate separators, got: %s", osPath.Value)
	}
}

// TC-CO-04: Rejects invalid CfsPath (directory traversal)
func TestPathCfsToOs_InvalidPath(t *testing.T) {
	cfs := &pathutils.PathCfs{Value: "../../etc/passwd"}
	_, err := pathutils.PathCfsToOs(cfs)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got: %v", err)
	}
}

// TC-CO-05: Rejects symlink escaping project root
func TestPathCfsToOs_SymlinkEscapesRoot(t *testing.T) {
	// outer is outside the project root; inner is the simulated project root.
	outer := t.TempDir()
	inner := t.TempDir()
	testChdir(t, inner)

	// Create a symlink inside inner that points to outer.
	symlinkName := "escape-link"
	symlinkPath := filepath.Join(inner, symlinkName)
	if err := os.Symlink(outer, symlinkPath); err != nil {
		t.Skipf("cannot create symlink (may require elevated privileges on Windows): %v", err)
	}

	cfs := &pathutils.PathCfs{Value: symlinkName + "/sensitive-file"}
	_, err := pathutils.PathCfsToOs(cfs)
	if err == nil {
		t.Fatal("expected error for symlink escaping root, got nil")
	}
	if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
		t.Errorf("expected ErrResolvesOutsideRoot, got: %v", err)
	}
}

// TC-CO-06: Roundtrip — CfsToOs then OsToCfs
func TestPathCfsToOs_Roundtrip(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfs := &pathutils.PathCfs{Value: "internal/config/config.go"}
	osPath, err := pathutils.PathCfsToOs(cfs)
	if err != nil {
		t.Fatalf("PathCfsToOs: %v", err)
	}

	cfsBack, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("PathOsToCfs: %v", err)
	}

	if cfsBack.Value != "internal/config/config.go" {
		t.Errorf("expected roundtrip value %q, got %q", "internal/config/config.go", cfsBack.Value)
	}
}

// ---------------------------------------------------------------------------
// PathOsToCfs
// ---------------------------------------------------------------------------

// TC-OC-01: Converts valid OS path that exists
func TestPathOsToCfs_ValidPathExists(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create a real file so the path exists.
	if err := os.MkdirAll(filepath.Join(tempDir, "internal", "config"), 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	absPath := filepath.Join(tempDir, "internal", "config", "config.go")
	if err := os.WriteFile(absPath, []byte("package config"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	osPath := &pathutils.PathOs{Value: absPath}
	cfsPath, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Must be relative, use forward slashes, and be correct.
	if strings.HasPrefix(cfsPath.Value, "/") {
		t.Errorf("expected relative CFS path, got: %s", cfsPath.Value)
	}
	if strings.Contains(cfsPath.Value, "\\") {
		t.Errorf("expected no backslashes, got: %s", cfsPath.Value)
	}
	if cfsPath.Value != "internal/config/config.go" {
		t.Errorf("expected %q, got %q", "internal/config/config.go", cfsPath.Value)
	}
}

// TC-OC-02: Converts valid OS path that does not exist
func TestPathOsToCfs_ValidPathNotExists(t *testing.T) {
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
}

// TC-OC-03: Result uses forward slashes
func TestPathOsToCfs_ForwardSlashes(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	absPath := filepath.Join(tempDir, "some", "nested", "file.go")
	osPath := &pathutils.PathOs{Value: absPath}
	cfsPath, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(cfsPath.Value, "\\") {
		t.Errorf("CFS path must not contain backslashes, got: %s", cfsPath.Value)
	}
}

// TC-OC-04: Symlink within root resolving within root
func TestPathOsToCfs_SymlinkWithinRoot(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create a real target directory inside tempDir.
	targetDir := filepath.Join(tempDir, "real-target")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("setup mkdir: %v", err)
	}

	// Create a symlink inside tempDir pointing to the target (also inside tempDir).
	symlinkPath := filepath.Join(tempDir, "link-to-target")
	if err := os.Symlink(targetDir, symlinkPath); err != nil {
		t.Skipf("cannot create symlink (may require elevated privileges on Windows): %v", err)
	}

	osPath := &pathutils.PathOs{Value: symlinkPath}
	_, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Errorf("expected no error for symlink within root, got: %v", err)
	}
}

// TC-OC-05: Rejects path outside project root
func TestPathOsToCfs_PathOutsideRoot(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Use a path that is clearly outside tempDir (the simulated project root).
	outsidePath := filepath.Join(os.TempDir(), "outside", "file.txt")
	osPath := &pathutils.PathOs{Value: outsidePath}
	_, err := pathutils.PathOsToCfs(osPath)
	if err == nil {
		t.Fatal("expected error for path outside root, got nil")
	}
	if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
		t.Errorf("expected ErrResolvesOutsideRoot, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// PathGetProjectRoot
// ---------------------------------------------------------------------------

// TC-GR-01: Returns an absolute path
func TestPathGetProjectRoot_ReturnsAbsolutePath(t *testing.T) {
	root, err := pathutils.PathGetProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root.Value == "" {
		t.Fatal("expected non-empty path")
	}
	if !filepath.IsAbs(root.Value) {
		t.Errorf("expected absolute path, got: %s", root.Value)
	}
}

// TC-GR-02: Matches working directory
func TestPathGetProjectRoot_MatchesWorkingDirectory(t *testing.T) {
	root, err := pathutils.PathGetProjectRoot()
	if err != nil {
		t.Fatalf("PathGetProjectRoot: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}

	// Evaluate symlinks on both sides for a fair comparison.
	rootResolved, err := filepath.EvalSymlinks(root.Value)
	if err != nil {
		// If path does not exist yet, fall back to clean comparison.
		rootResolved = filepath.Clean(root.Value)
	}
	cwdResolved, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		cwdResolved = filepath.Clean(cwd)
	}

	if rootResolved != cwdResolved {
		t.Errorf("PathGetProjectRoot returned %q, want %q (cwd)", rootResolved, cwdResolved)
	}
}
