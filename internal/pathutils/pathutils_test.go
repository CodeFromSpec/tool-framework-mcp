// code-from-spec: ROOT/golang/tests/os/path_utils@dOvlZ3-8tTKckRgtOgPcX0aXPfY

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

	cases := []testCase{
		// TC-PV-01
		{name: "TC-PV-01 valid simple relative path", input: "internal/config/config.go", wantErr: nil},
		// TC-PV-02
		{name: "TC-PV-02 valid nested path", input: "cmd/framework-mcp/main.go", wantErr: nil},
		// TC-PV-03
		{name: "TC-PV-03 valid single filename", input: "main.go", wantErr: nil},
		// TC-PV-04
		{name: "TC-PV-04 accepts path with dot segment", input: "internal/./config/config.go", wantErr: nil},
		// TC-PV-05
		{name: "TC-PV-05 accepts traversal that resolves within root", input: "a/b/../c", wantErr: nil},
		// TC-PV-06
		{name: "TC-PV-06 accepts path with trailing slash", input: "internal/config/", wantErr: nil},
		// TC-PV-07
		{name: "TC-PV-07 accepts path with duplicate slashes", input: "internal//config//file.go", wantErr: nil},
		// TC-PV-08
		{name: "TC-PV-08 rejects empty string", input: "", wantErr: pathutils.ErrPathEmpty},
		// TC-PV-09
		{name: "TC-PV-09 rejects absolute path with leading slash", input: "/etc/passwd", wantErr: pathutils.ErrPathAbsolute},
		// TC-PV-10
		{name: "TC-PV-10 rejects absolute path with drive letter", input: "C:/Windows/system32", wantErr: pathutils.ErrPathAbsolute},
		// TC-PV-11
		{name: "TC-PV-11 rejects backslash", input: `internal\config\config.go`, wantErr: pathutils.ErrPathContainsBackslash},
		// TC-PV-12
		{name: "TC-PV-12 rejects simple traversal", input: "../../etc/passwd", wantErr: pathutils.ErrDirectoryTraversal},
		// TC-PV-13
		{name: "TC-PV-13 rejects embedded traversal", input: "internal/../../outside/file.go", wantErr: pathutils.ErrDirectoryTraversal},
	}

	for _, tc := range cases {
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
func TestPathCfsToOs_ValidPathExists(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create the file so it exists.
	if err := os.MkdirAll(filepath.Join(tempDir, "internal", "config"), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "internal", "config", "config.go"), []byte("package config"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfs := &pathutils.PathCfs{Value: "internal/config/config.go"}
	result, err := pathutils.PathCfsToOs(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(result.Value) {
		t.Errorf("expected absolute path, got: %q", result.Value)
	}
	wantSuffix := filepath.Join("internal", "config", "config.go")
	if !strings.HasSuffix(result.Value, wantSuffix) {
		t.Errorf("expected path ending in %q, got: %q", wantSuffix, result.Value)
	}
}

// TC-CO-02: Converts valid path that does not exist.
func TestPathCfsToOs_ValidPathNotExists(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfs := &pathutils.PathCfs{Value: "internal/newdir/newfile.go"}
	result, err := pathutils.PathCfsToOs(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(result.Value) {
		t.Errorf("expected absolute path, got: %q", result.Value)
	}
	wantSuffix := filepath.Join("internal", "newdir", "newfile.go")
	if !strings.HasSuffix(result.Value, wantSuffix) {
		t.Errorf("expected path ending in %q, got: %q", wantSuffix, result.Value)
	}
}

// TC-CO-03: Converts path with duplicate slashes.
func TestPathCfsToOs_DuplicateSlashes(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	cfs := &pathutils.PathCfs{Value: "internal//config.go"}
	result, err := pathutils.PathCfsToOs(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(result.Value) {
		t.Errorf("expected absolute path, got: %q", result.Value)
	}
}

// TC-CO-04: Rejects invalid CfsPath — directory traversal.
func TestPathCfsToOs_DirectoryTraversal(t *testing.T) {
	cfs := &pathutils.PathCfs{Value: "../../etc/passwd"}
	_, err := pathutils.PathCfsToOs(cfs)
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got: %v", err)
	}
}

// TC-CO-05: Rejects symlink escaping project root.
func TestPathCfsToOs_SymlinkEscapesRoot(t *testing.T) {
	// Create a temp dir to act as the project root.
	projectRoot := t.TempDir()
	// Create a separate dir outside the project root for the symlink target.
	outsideDir := t.TempDir()

	testChdir(t, projectRoot)

	// Create the symlink inside the project root pointing outside.
	symlinkName := "escape-link"
	symlinkPath := filepath.Join(projectRoot, symlinkName)
	if err := os.Symlink(outsideDir, symlinkPath); err != nil {
		t.Skipf("cannot create symlink (may require elevated privileges): %v", err)
	}

	cfs := &pathutils.PathCfs{Value: symlinkName + "/secret.txt"}
	_, err := pathutils.PathCfsToOs(cfs)
	if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
		t.Errorf("expected ErrResolvesOutsideRoot, got: %v", err)
	}
}

// TC-CO-06: Roundtrip — CfsToOs then OsToCfs.
func TestPathCfsToOs_Roundtrip(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	original := "internal/config/config.go"
	cfs := &pathutils.PathCfs{Value: original}

	osPath, err := pathutils.PathCfsToOs(cfs)
	if err != nil {
		t.Fatalf("PathCfsToOs: %v", err)
	}

	cfsBack, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("PathOsToCfs: %v", err)
	}

	if cfsBack.Value != original {
		t.Errorf("roundtrip mismatch: got %q, want %q", cfsBack.Value, original)
	}
}

// ---------------------------------------------------------------------------
// PathOsToCfs
// ---------------------------------------------------------------------------

// TC-OC-01: Converts valid OS path that exists.
func TestPathOsToCfs_ValidPathExists(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create a file inside the temp dir (project root).
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	filePath := filepath.Join(subDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	osPath := &pathutils.PathOs{Value: filePath}
	result, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result.Value, "\\") {
		t.Errorf("expected forward slashes only, got: %q", result.Value)
	}
	if !strings.HasPrefix(result.Value, "subdir/") {
		t.Errorf("expected path starting with subdir/, got: %q", result.Value)
	}
}

// TC-OC-02: Converts valid OS path that does not exist.
func TestPathOsToCfs_ValidPathNotExists(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	nonExistentPath := filepath.Join(tempDir, "nonexistent", "file.go")
	osPath := &pathutils.PathOs{Value: nonExistentPath}
	result, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result.Value, "\\") {
		t.Errorf("expected forward slashes only, got: %q", result.Value)
	}
	if result.Value == "" {
		t.Error("expected non-empty CFS path")
	}
}

// TC-OC-03: Result uses forward slashes.
func TestPathOsToCfs_ForwardSlashes(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Build a path using OS-native separators.
	nativePath := filepath.Join(tempDir, "a", "b", "c.txt")
	osPath := &pathutils.PathOs{Value: nativePath}
	result, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result.Value, "\\") {
		t.Errorf("result must not contain backslashes, got: %q", result.Value)
	}
}

// TC-OC-04: Symlink within root resolving within root.
func TestPathOsToCfs_SymlinkWithinRoot(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create target dir inside root.
	targetDir := filepath.Join(tempDir, "target")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Create symlink inside root pointing to target (also inside root).
	symlinkPath := filepath.Join(tempDir, "link-to-target")
	if err := os.Symlink(targetDir, symlinkPath); err != nil {
		t.Skipf("cannot create symlink (may require elevated privileges): %v", err)
	}

	osPath := &pathutils.PathOs{Value: symlinkPath}
	result, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Value == "" {
		t.Error("expected non-empty CFS path")
	}
}

// TC-OC-05: Rejects path outside project root.
func TestPathOsToCfs_OutsideRoot(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Use the OS temp dir itself as an outside path; it is guaranteed to
	// exist and be outside our tempDir project root.
	outsidePath := filepath.Dir(tempDir)
	osPath := &pathutils.PathOs{Value: outsidePath}
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
	result, err := pathutils.PathGetProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Value == "" {
		t.Error("expected non-empty path")
	}
	if !filepath.IsAbs(result.Value) {
		t.Errorf("expected absolute path, got: %q", result.Value)
	}
}

// TC-GR-02: Matches working directory.
func TestPathGetProjectRoot_MatchesWorkingDirectory(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}

	result, err := pathutils.PathGetProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Evaluate symlinks on both sides for a fair comparison.
	cwdReal, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		t.Fatalf("EvalSymlinks cwd: %v", err)
	}
	resultReal, err := filepath.EvalSymlinks(result.Value)
	if err != nil {
		t.Fatalf("EvalSymlinks result: %v", err)
	}

	if cwdReal != resultReal {
		t.Errorf("expected %q, got %q", cwdReal, resultReal)
	}
}
