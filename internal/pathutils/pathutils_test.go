// code-from-spec: ROOT/golang/tests/os/path_utils@LHwFwFhVX2lgLKE7OFFimgoJVt8
package pathutils_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

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
		// TC-PV-01
		{name: "TC-PV-01 valid simple relative path", input: "internal/config/config.go", wantErr: nil},
		// TC-PV-02
		{name: "TC-PV-02 valid nested path", input: "cmd/framework-mcp/main.go", wantErr: nil},
		// TC-PV-03
		{name: "TC-PV-03 valid single filename", input: "main.go", wantErr: nil},
		// TC-PV-04
		{name: "TC-PV-04 accepts dot segment", input: "internal/./config/config.go", wantErr: nil},
		// TC-PV-05
		{name: "TC-PV-05 accepts traversal that resolves within root", input: "a/b/../c", wantErr: nil},
		// TC-PV-06
		{name: "TC-PV-06 accepts trailing slash", input: "internal/config/", wantErr: nil},
		// TC-PV-07
		{name: "TC-PV-07 accepts duplicate slashes", input: "internal//config//file.go", wantErr: nil},
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
					t.Fatalf("expected no error, got: %v", err)
				}
			} else {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got: %v", tc.wantErr, err)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// PathCfsToOs
// ---------------------------------------------------------------------------

// testChdir changes the working directory and restores it on cleanup.
func testChdir(t *testing.T, dir string) {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChdir: could not get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChdir: could not chdir to %q: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Errorf("testChdir cleanup: could not restore working directory: %v", err)
		}
	})
}

// TC-CO-01: Converts valid path that exists.
func TestPathCfsToOs_ExistingFile(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	// Create the file that the path will point to.
	dir := filepath.Join(tmp, "internal", "config")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.go"), []byte("package config"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfs := &pathutils.PathCfs{Value: "internal/config/config.go"}
	osPath, err := pathutils.PathCfsToOs(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(osPath.Value) {
		t.Fatalf("expected absolute path, got: %q", osPath.Value)
	}
	if !strings.HasSuffix(osPath.Value, filepath.Join("internal", "config", "config.go")) {
		t.Fatalf("path tail mismatch: %q", osPath.Value)
	}
}

// TC-CO-02: Converts valid path that does not exist.
func TestPathCfsToOs_NonExistentFile(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	cfs := &pathutils.PathCfs{Value: "internal/newdir/newfile.go"}
	osPath, err := pathutils.PathCfsToOs(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(osPath.Value) {
		t.Fatalf("expected absolute path, got: %q", osPath.Value)
	}
	if !strings.HasSuffix(osPath.Value, filepath.Join("internal", "newdir", "newfile.go")) {
		t.Fatalf("path tail mismatch: %q", osPath.Value)
	}
}

// TC-CO-03: Converts path with duplicate slashes.
func TestPathCfsToOs_DuplicateSlashes(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	cfs := &pathutils.PathCfs{Value: "internal//config.go"}
	osPath, err := pathutils.PathCfsToOs(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(osPath.Value) {
		t.Fatalf("expected absolute path, got: %q", osPath.Value)
	}
}

// TC-CO-04: Rejects invalid CfsPath — directory traversal.
func TestPathCfsToOs_DirectoryTraversal(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	cfs := &pathutils.PathCfs{Value: "../../etc/passwd"}
	_, err := pathutils.PathCfsToOs(cfs)
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Fatalf("expected ErrDirectoryTraversal, got: %v", err)
	}
}

// TC-CO-05: Rejects symlink escaping project root.
func TestPathCfsToOs_SymlinkEscapesRoot(t *testing.T) {
	tmp := t.TempDir()
	outside := t.TempDir()
	testChdir(t, tmp)

	symlinkName := "escape_link"
	symlinkPath := filepath.Join(tmp, symlinkName)
	if err := os.Symlink(outside, symlinkPath); err != nil {
		t.Skipf("cannot create symlink (may require elevated privileges): %v", err)
	}

	cfs := &pathutils.PathCfs{Value: symlinkName + "/secret.txt"}
	_, err := pathutils.PathCfsToOs(cfs)
	if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
		t.Fatalf("expected ErrResolvesOutsideRoot, got: %v", err)
	}
}

// TC-CO-06: Roundtrip — CfsToOs then OsToCfs.
func TestPathCfsToOs_Roundtrip(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	original := "internal/config/config.go"
	cfs := &pathutils.PathCfs{Value: original}

	osPath, err := pathutils.PathCfsToOs(cfs)
	if err != nil {
		t.Fatalf("PathCfsToOs: unexpected error: %v", err)
	}

	cfsBack, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("PathOsToCfs: unexpected error: %v", err)
	}

	if cfsBack.Value != original {
		t.Fatalf("roundtrip mismatch: got %q, want %q", cfsBack.Value, original)
	}
}

// ---------------------------------------------------------------------------
// PathOsToCfs
// ---------------------------------------------------------------------------

// TC-OC-01: Converts valid OS path that exists.
func TestPathOsToCfs_ExistingFile(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	filePath := filepath.Join(tmp, "somefile.go")
	if err := os.WriteFile(filePath, []byte("package main"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	osPath := &pathutils.PathOs{Value: filePath}
	cfsPath, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(cfsPath.Value, "somefile.go") {
		t.Fatalf("unexpected CFS path: %q", cfsPath.Value)
	}
	if strings.Contains(cfsPath.Value, string(os.PathSeparator)) && os.PathSeparator != '/' {
		t.Fatalf("CFS path contains OS separator: %q", cfsPath.Value)
	}
}

// TC-OC-02: Converts valid OS path that does not exist.
func TestPathOsToCfs_NonExistentFile(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	osPath := &pathutils.PathOs{Value: filepath.Join(tmp, "nonexistent", "file.go")}
	cfsPath, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(cfsPath.Value, "nonexistent/file.go") {
		t.Fatalf("unexpected CFS path: %q", cfsPath.Value)
	}
}

// TC-OC-03: Result uses forward slashes.
func TestPathOsToCfs_ForwardSlashes(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	// Build a nested path using OS separator to emphasise platform differences.
	nestedDir := filepath.Join(tmp, "a", "b", "c")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	filePath := filepath.Join(nestedDir, "file.go")
	if err := os.WriteFile(filePath, []byte(""), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	osPath := &pathutils.PathOs{Value: filePath}
	cfsPath, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(cfsPath.Value, `\`) {
		t.Fatalf("CFS path contains backslash: %q", cfsPath.Value)
	}
}

// TC-OC-04: Symlink within root resolving within root.
func TestPathOsToCfs_SymlinkWithinRoot(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	// Create target inside root.
	targetDir := filepath.Join(tmp, "target")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Create symlink inside root pointing to the target inside root.
	symlinkPath := filepath.Join(tmp, "link_to_target")
	if err := os.Symlink(targetDir, symlinkPath); err != nil {
		t.Skipf("cannot create symlink (may require elevated privileges): %v", err)
	}

	osPath := &pathutils.PathOs{Value: symlinkPath}
	cfsPath, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfsPath.Value == "" {
		t.Fatal("expected non-empty CFS path")
	}
}

// TC-OC-05: Rejects path outside project root.
func TestPathOsToCfs_OutsideRoot(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	outside := t.TempDir()
	osPath := &pathutils.PathOs{Value: filepath.Join(outside, "secret.txt")}
	_, err := pathutils.PathOsToCfs(osPath)
	if !errors.Is(err, pathutils.ErrResolvesOutsideRoot) {
		t.Fatalf("expected ErrResolvesOutsideRoot, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// PathGetProjectRoot
// ---------------------------------------------------------------------------

// TC-GR-01: Returns an absolute path.
func TestPathGetProjectRoot_IsAbsolute(t *testing.T) {
	root, err := pathutils.PathGetProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root.Value == "" {
		t.Fatal("expected non-empty root path")
	}
	if !filepath.IsAbs(root.Value) {
		t.Fatalf("expected absolute path, got: %q", root.Value)
	}
}

// TC-GR-02: Matches working directory.
func TestPathGetProjectRoot_MatchesWorkingDirectory(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	root, err := pathutils.PathGetProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}

	// Evaluate symlinks on both sides so comparison is stable.
	wdResolved, err := filepath.EvalSymlinks(wd)
	if err != nil {
		t.Fatalf("EvalSymlinks(wd): %v", err)
	}
	rootResolved, err := filepath.EvalSymlinks(root.Value)
	if err != nil {
		t.Fatalf("EvalSymlinks(root): %v", err)
	}

	if rootResolved != wdResolved {
		t.Fatalf("root %q does not match working directory %q", rootResolved, wdResolved)
	}
}
