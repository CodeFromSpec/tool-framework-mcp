// code-from-spec: ROOT/golang/tests/os/path_utils@NNfXd5UAeDrx_oFNn2MFclRRhuU
package pathutils_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// testChdir changes the working directory to dir for the duration of the test.
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
		name    string
		input   string
		wantErr error
	}

	tests := []testCase{
		// TC-PV-01: Valid simple relative path
		{
			name:    "TC-PV-01 valid simple relative path",
			input:   "internal/config/config.go",
			wantErr: nil,
		},
		// TC-PV-02: Valid nested path
		{
			name:    "TC-PV-02 valid nested path",
			input:   "cmd/framework-mcp/main.go",
			wantErr: nil,
		},
		// TC-PV-03: Valid single filename
		{
			name:    "TC-PV-03 valid single filename",
			input:   "main.go",
			wantErr: nil,
		},
		// TC-PV-04: Accepts path with dot segment
		{
			name:    "TC-PV-04 accepts path with dot segment",
			input:   "internal/./config/config.go",
			wantErr: nil,
		},
		// TC-PV-05: Accepts traversal that resolves within root
		{
			name:    "TC-PV-05 accepts traversal resolving within root",
			input:   "a/b/../c",
			wantErr: nil,
		},
		// TC-PV-06: Accepts path with trailing slash
		{
			name:    "TC-PV-06 accepts path with trailing slash",
			input:   "internal/config/",
			wantErr: nil,
		},
		// TC-PV-07: Accepts path with duplicate slashes
		{
			name:    "TC-PV-07 accepts path with duplicate slashes",
			input:   "internal//config//file.go",
			wantErr: nil,
		},
		// TC-PV-08: Rejects empty string
		{
			name:    "TC-PV-08 rejects empty string",
			input:   "",
			wantErr: pathutils.ErrPathEmpty,
		},
		// TC-PV-09: Rejects absolute path with leading slash
		{
			name:    "TC-PV-09 rejects absolute path with leading slash",
			input:   "/etc/passwd",
			wantErr: pathutils.ErrPathAbsolute,
		},
		// TC-PV-10: Rejects absolute path with drive letter
		{
			name:    "TC-PV-10 rejects absolute path with drive letter",
			input:   "C:/Windows/system32",
			wantErr: pathutils.ErrPathAbsolute,
		},
		// TC-PV-11: Rejects backslash
		{
			name:    "TC-PV-11 rejects backslash",
			input:   `internal\config\config.go`,
			wantErr: pathutils.ErrPathContainsBackslash,
		},
		// TC-PV-12: Rejects simple traversal
		{
			name:    "TC-PV-12 rejects simple traversal",
			input:   "../../etc/passwd",
			wantErr: pathutils.ErrDirectoryTraversal,
		},
		// TC-PV-13: Rejects embedded traversal
		{
			name:    "TC-PV-13 rejects embedded traversal",
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

// ---------------------------------------------------------------------------
// PathCfsToOs
// ---------------------------------------------------------------------------

// TC-CO-01: Converts valid path that exists.
// We use a file we know exists: the test file itself.
func TestPathCfsToOs_ExistingFile(t *testing.T) {
	// internal/pathutils/pathutils_test.go should exist since we're running it.
	cfsPath := &pathutils.PathCfs{Value: "internal/pathutils/pathutils_test.go"}
	osPath, err := pathutils.PathCfsToOs(cfsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(osPath.Value) {
		t.Errorf("expected absolute path, got: %s", osPath.Value)
	}
	if !strings.HasSuffix(osPath.Value, filepath.Join("internal", "pathutils", "pathutils_test.go")) {
		t.Errorf("path %q does not end with expected suffix", osPath.Value)
	}
}

// TC-CO-02: Converts valid path that does not exist.
func TestPathCfsToOs_NonExistingFile(t *testing.T) {
	cfsPath := &pathutils.PathCfs{Value: "internal/newdir/newfile.go"}
	osPath, err := pathutils.PathCfsToOs(cfsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(osPath.Value) {
		t.Errorf("expected absolute path, got: %s", osPath.Value)
	}
	if !strings.HasSuffix(osPath.Value, filepath.Join("internal", "newdir", "newfile.go")) {
		t.Errorf("path %q does not end with expected suffix", osPath.Value)
	}
}

// TC-CO-03: Converts path with duplicate slashes.
func TestPathCfsToOs_DuplicateSlashes(t *testing.T) {
	cfsPath := &pathutils.PathCfs{Value: "internal//config.go"}
	osPath, err := pathutils.PathCfsToOs(cfsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(osPath.Value, "//") || strings.Contains(osPath.Value, `\\`) {
		t.Errorf("normalized path should not contain duplicate separators: %s", osPath.Value)
	}
}

// TC-CO-04: Rejects invalid CFS path.
func TestPathCfsToOs_InvalidPath(t *testing.T) {
	cfsPath := &pathutils.PathCfs{Value: "../../etc/passwd"}
	_, err := pathutils.PathCfsToOs(cfsPath)
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got: %v", err)
	}
}

// TC-CO-05: Rejects symlink escaping project root.
func TestPathCfsToOs_SymlinkEscapesRoot(t *testing.T) {
	// We need a symlink inside the project root pointing outside.
	// Use t.TempDir() for the external target, then create a symlink
	// under the project root temp area.
	// We'll chdir to a temp dir to make it the project root.
	tempRoot := t.TempDir()
	outsideDir := t.TempDir()

	testChdir(t, tempRoot)

	// Create a symlink inside tempRoot pointing to outsideDir.
	symlinkName := "escape_link"
	err := os.Symlink(outsideDir, filepath.Join(tempRoot, symlinkName))
	if err != nil {
		t.Skipf("symlink creation not supported or insufficient permissions: %v", err)
	}

	// Create a real file in the outside directory so os.Stat succeeds
	// and the symlink resolution check runs.
	if err := os.WriteFile(filepath.Join(outsideDir, "file.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to create target file: %v", err)
	}

	cfsPath := &pathutils.PathCfs{Value: symlinkName + "/file.txt"}
	_, err = pathutils.PathCfsToOs(cfsPath)
	if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
		t.Errorf("expected ErrResolvesOutsideRoot, got: %v", err)
	}
}

// TC-CO-06: Roundtrip — CfsToOs then OsToCfs.
func TestPathCfsToOs_Roundtrip(t *testing.T) {
	cfsInput := &pathutils.PathCfs{Value: "internal/pathutils/pathutils_test.go"}
	osPath, err := pathutils.PathCfsToOs(cfsInput)
	if err != nil {
		t.Fatalf("PathCfsToOs unexpected error: %v", err)
	}

	cfsResult, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("PathOsToCfs unexpected error: %v", err)
	}

	if cfsResult.Value != "internal/pathutils/pathutils_test.go" {
		t.Errorf("roundtrip mismatch: got %q, want %q", cfsResult.Value, "internal/pathutils/pathutils_test.go")
	}
}

// ---------------------------------------------------------------------------
// PathOsToCfs
// ---------------------------------------------------------------------------

// TC-OC-01: Converts valid OS path that exists.
func TestPathOsToCfs_ExistingFile(t *testing.T) {
	tempRoot := t.TempDir()
	testChdir(t, tempRoot)

	relPath := "subdir/file.txt"
	absPath := filepath.Join(tempRoot, "subdir", "file.txt")
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(absPath, []byte("content"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	osPath := &pathutils.PathOs{Value: absPath}
	cfsPath, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(cfsPath.Value, `\`) {
		t.Errorf("CFS path contains backslash: %s", cfsPath.Value)
	}
	if cfsPath.Value != relPath {
		t.Errorf("got %q, want %q", cfsPath.Value, relPath)
	}
}

// TC-OC-02: Converts valid OS path that does not exist.
func TestPathOsToCfs_NonExistingFile(t *testing.T) {
	tempRoot := t.TempDir()
	testChdir(t, tempRoot)

	absPath := filepath.Join(tempRoot, "nonexistent", "file.go")
	osPath := &pathutils.PathOs{Value: absPath}
	cfsPath, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(cfsPath.Value, `\`) {
		t.Errorf("CFS path contains backslash: %s", cfsPath.Value)
	}
	if cfsPath.Value != "nonexistent/file.go" {
		t.Errorf("got %q, want %q", cfsPath.Value, "nonexistent/file.go")
	}
}

// TC-OC-03: Result uses forward slashes on any OS.
func TestPathOsToCfs_ForwardSlashes(t *testing.T) {
	tempRoot := t.TempDir()
	testChdir(t, tempRoot)

	absPath := filepath.Join(tempRoot, "a", "b", "c.txt")
	osPath := &pathutils.PathOs{Value: absPath}
	cfsPath, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(cfsPath.Value, `\`) {
		t.Errorf("CFS path must not contain backslash: %s", cfsPath.Value)
	}
}

// TC-OC-04: Symlink within root resolving within root.
func TestPathOsToCfs_SymlinkWithinRoot(t *testing.T) {
	tempRoot := t.TempDir()
	testChdir(t, tempRoot)

	// Create target inside root.
	targetPath := filepath.Join(tempRoot, "real_file.txt")
	if err := os.WriteFile(targetPath, []byte("data"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Create symlink inside root pointing to target inside root.
	symlinkPath := filepath.Join(tempRoot, "link_file.txt")
	err := os.Symlink(targetPath, symlinkPath)
	if err != nil {
		t.Skipf("symlink creation not supported or insufficient permissions: %v", err)
	}

	osPath := &pathutils.PathOs{Value: symlinkPath}
	cfsPath, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfsPath.Value == "" {
		t.Error("expected non-empty CFS path")
	}
}

// TC-OC-05: Rejects path outside project root.
func TestPathOsToCfs_OutsideRoot(t *testing.T) {
	// Use a system-level path that is definitely outside any project root.
	outsidePath := t.TempDir()

	// Ensure it's truly outside by using the original working directory.
	osPath := &pathutils.PathOs{Value: outsidePath}
	_, err := pathutils.PathOsToCfs(osPath)
	if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
		t.Errorf("expected ErrResolvesOutsideRoot, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// PathGetProjectRoot
// ---------------------------------------------------------------------------

// TC-PR-01: Returns an absolute path.
func TestPathGetProjectRoot_ReturnsAbsolutePath(t *testing.T) {
	root, err := pathutils.PathGetProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root.Value == "" {
		t.Error("expected non-empty root path")
	}
	if !filepath.IsAbs(root.Value) {
		t.Errorf("expected absolute path, got: %s", root.Value)
	}
}

// TC-PR-02: Matches working directory.
func TestPathGetProjectRoot_MatchesWorkingDirectory(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}

	root, err := pathutils.PathGetProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Normalize both for comparison (resolve symlinks, clean).
	cwdResolved, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		// If EvalSymlinks fails, fall back to direct comparison.
		cwdResolved = filepath.Clean(cwd)
	}
	rootResolved, err := filepath.EvalSymlinks(root.Value)
	if err != nil {
		rootResolved = filepath.Clean(root.Value)
	}

	if cwdResolved != rootResolved {
		t.Errorf("project root %q does not match working directory %q", root.Value, cwd)
	}
}
