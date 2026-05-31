// code-from-spec: ROOT/golang/tests/os/path_utils@GadQMAG8AC353PsFNtsyKrl48HY
package pathutils_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// testChdir changes the working directory to dir and restores it on cleanup.
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
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		// TC-PV-01
		{name: "TC-PV-01 valid simple relative path", input: "internal/config/config.go", wantErr: nil},
		// TC-PV-02
		{name: "TC-PV-02 valid nested path", input: "cmd/framework-mcp/main.go", wantErr: nil},
		// TC-PV-03
		{name: "TC-PV-03 valid single filename", input: "main.go", wantErr: nil},
		// TC-PV-04
		{name: "TC-PV-04 accepts path with dot segment", input: "internal/./config/config.go", wantErr: nil},
		// TC-PV-05
		{name: "TC-PV-05 accepts traversal resolving within root", input: "a/b/../c", wantErr: nil},
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
func TestPathCfsToOs_ValidExistingPath(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create the file structure.
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
		t.Errorf("expected absolute path, got: %s", result.Value)
	}
	wantSuffix := string(filepath.Separator) + filepath.Join("internal", "config", "config.go")
	if !strings.HasSuffix(result.Value, wantSuffix) {
		t.Errorf("expected path to end with %q, got: %s", wantSuffix, result.Value)
	}
}

// TC-CO-02: Converts valid path that does not exist.
func TestPathCfsToOs_ValidNonExistingPath(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	result, err := pathutils.PathCfsToOs(&pathutils.PathCfs{Value: "internal/newdir/newfile.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(result.Value) {
		t.Errorf("expected absolute path, got: %s", result.Value)
	}
	wantSuffix := string(filepath.Separator) + filepath.Join("internal", "newdir", "newfile.go")
	if !strings.HasSuffix(result.Value, wantSuffix) {
		t.Errorf("expected path to end with %q, got: %s", wantSuffix, result.Value)
	}
}

// TC-CO-03: Converts path with duplicate slashes.
func TestPathCfsToOs_DuplicateSlashes(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	result, err := pathutils.PathCfsToOs(&pathutils.PathCfs{Value: "internal//config.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result.Value, "//") || strings.Contains(result.Value, `\\`) {
		t.Errorf("expected normalized path without duplicate separators, got: %s", result.Value)
	}
}

// TC-CO-04: Rejects invalid CfsPath.
func TestPathCfsToOs_RejectsTraversal(t *testing.T) {
	_, err := pathutils.PathCfsToOs(&pathutils.PathCfs{Value: "../../etc/passwd"})
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got: %v", err)
	}
}

// TC-CO-05: Rejects symlink escaping project root.
func TestPathCfsToOs_RejectsSymlinkEscapingRoot(t *testing.T) {
	// We need two temp dirs: one as our "root" and one as the "outside" target.
	outsideDir := t.TempDir()
	projectDir := t.TempDir()
	testChdir(t, projectDir)

	// Create a symlink inside the project root pointing outside.
	symlinkPath := filepath.Join(projectDir, "escape_link")
	if err := os.Symlink(outsideDir, symlinkPath); err != nil {
		t.Skipf("cannot create symlink (may require elevated privileges on Windows): %v", err)
	}

	_, err := pathutils.PathCfsToOs(&pathutils.PathCfs{Value: "escape_link/some_file.txt"})
	if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
		t.Errorf("expected ErrResolvesOutsideRoot, got: %v", err)
	}
}

// TC-CO-06: Roundtrip — CfsToOs then OsToCfs.
func TestPathCfsToOs_Roundtrip(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	original := &pathutils.PathCfs{Value: "internal/config/config.go"}

	osPath, err := pathutils.PathCfsToOs(original)
	if err != nil {
		t.Fatalf("PathCfsToOs: %v", err)
	}

	cfsBack, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("PathOsToCfs: %v", err)
	}

	if cfsBack.Value != original.Value {
		t.Errorf("roundtrip mismatch: got %q, want %q", cfsBack.Value, original.Value)
	}
}

// ---------------------------------------------------------------------------
// PathOsToCfs
// ---------------------------------------------------------------------------

// TC-OC-01: Converts valid OS path that exists.
func TestPathOsToCfs_ValidExistingPath(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.MkdirAll("subdir", 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile("subdir/file.txt", []byte("hello"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	absPath := filepath.Join(tempDir, "subdir", "file.txt")
	result, err := pathutils.PathOsToCfs(&pathutils.PathOs{Value: absPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result.Value, `\`) {
		t.Errorf("PathCfs must use forward slashes, got: %s", result.Value)
	}
	if !strings.HasPrefix(result.Value, "subdir/") {
		t.Errorf("expected CFS path relative to project root, got: %s", result.Value)
	}
}

// TC-OC-02: Converts valid OS path that does not exist.
func TestPathOsToCfs_ValidNonExistingPath(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	absPath := filepath.Join(tempDir, "nonexistent", "file.go")
	result, err := pathutils.PathOsToCfs(&pathutils.PathOs{Value: absPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result.Value, `\`) {
		t.Errorf("PathCfs must use forward slashes, got: %s", result.Value)
	}
	if result.Value != "nonexistent/file.go" {
		t.Errorf("expected %q, got: %q", "nonexistent/file.go", result.Value)
	}
}

// TC-OC-03: Result uses forward slashes.
func TestPathOsToCfs_UsesForwardSlashes(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	absPath := filepath.Join(tempDir, "a", "b", "c.txt")
	result, err := pathutils.PathOsToCfs(&pathutils.PathOs{Value: absPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result.Value, `\`) {
		t.Errorf("PathCfs contains backslash on result: %s", result.Value)
	}
}

// TC-OC-04: Symlink within root resolving within root.
func TestPathOsToCfs_SymlinkWithinRoot(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Create a real file and a symlink to it, both inside the project root.
	if err := os.WriteFile(filepath.Join(tempDir, "real.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	symlinkPath := filepath.Join(tempDir, "link.txt")
	if err := os.Symlink(filepath.Join(tempDir, "real.txt"), symlinkPath); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	result, err := pathutils.PathOsToCfs(&pathutils.PathOs{Value: symlinkPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Value == "" {
		t.Error("expected non-empty PathCfs value")
	}
}

// TC-OC-05: Rejects path outside project root.
func TestPathOsToCfs_RejectsPathOutsideRoot(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Use the OS temp directory, which is outside our project root (tempDir).
	outsidePath := t.TempDir()

	_, err := pathutils.PathOsToCfs(&pathutils.PathOs{Value: outsidePath})
	if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
		t.Errorf("expected ErrResolvesOutsideRoot, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// PathGetProjectRoot
// ---------------------------------------------------------------------------

// TC-PR-01: Returns an absolute path.
func TestPathGetProjectRoot_ReturnsAbsolutePath(t *testing.T) {
	result, err := pathutils.PathGetProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Value == "" {
		t.Fatal("expected non-empty path")
	}
	if !filepath.IsAbs(result.Value) {
		t.Errorf("expected absolute path, got: %s", result.Value)
	}
}

// TC-PR-02: Matches working directory.
func TestPathGetProjectRoot_MatchesWorkingDirectory(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}

	result, err := pathutils.PathGetProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Normalize both paths for comparison (resolve symlinks if needed).
	cwdEval, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		cwdEval = cwd
	}
	resultEval, err := filepath.EvalSymlinks(result.Value)
	if err != nil {
		resultEval = result.Value
	}

	if cwdEval != resultEval {
		t.Errorf("expected root %q to match working directory %q", resultEval, cwdEval)
	}
}
