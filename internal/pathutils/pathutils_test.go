// code-from-spec: ROOT/golang/tests/os/path_utils@eSclI-Oj27FW0NNJD1ieBH8TGg4

package pathutils

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPathValidateCfs covers all TC-PV-* test cases.
func TestPathValidateCfs(t *testing.T) {
	type testCase struct {
		name    string
		input   string
		wantErr error
	}

	cases := []testCase{
		// TC-PV-01
		{
			name:    "valid simple relative path",
			input:   "internal/config/config.go",
			wantErr: nil,
		},
		// TC-PV-02
		{
			name:    "valid nested path",
			input:   "cmd/framework-mcp/main.go",
			wantErr: nil,
		},
		// TC-PV-03
		{
			name:    "valid single filename",
			input:   "main.go",
			wantErr: nil,
		},
		// TC-PV-04
		{
			name:    "accepts path with dot segment",
			input:   "internal/./config/config.go",
			wantErr: nil,
		},
		// TC-PV-05
		{
			name:    "accepts traversal that resolves within root",
			input:   "a/b/../c",
			wantErr: nil,
		},
		// TC-PV-06
		{
			name:    "accepts path with trailing slash",
			input:   "internal/config/",
			wantErr: nil,
		},
		// TC-PV-07
		{
			name:    "accepts path with duplicate slashes",
			input:   "internal//config//file.go",
			wantErr: nil,
		},
		// TC-PV-08
		{
			name:    "rejects empty string",
			input:   "",
			wantErr: ErrPathEmpty,
		},
		// TC-PV-09
		{
			name:    "rejects absolute path with leading slash",
			input:   "/etc/passwd",
			wantErr: ErrPathAbsolute,
		},
		// TC-PV-10
		{
			name:    "rejects absolute path with drive letter",
			input:   "C:/Windows/system32",
			wantErr: ErrPathAbsolute,
		},
		// TC-PV-11
		{
			name:    "rejects backslash",
			input:   `internal\config\config.go`,
			wantErr: ErrPathContainsBackslash,
		},
		// TC-PV-12
		{
			name:    "rejects simple traversal",
			input:   "../../etc/passwd",
			wantErr: ErrDirectoryTraversal,
		},
		// TC-PV-13
		{
			name:    "rejects embedded traversal",
			input:   "internal/../../outside/file.go",
			wantErr: ErrDirectoryTraversal,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := PathValidateCfs(tc.input)
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

// TestPathCfsToOs_ValidExistingPath covers TC-CO-01.
func TestPathCfsToOs_ValidExistingPath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	subDir := filepath.Join(dir, "internal", "config")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(subDir, "config.go")
	if err := os.WriteFile(filePath, []byte("package config"), 0644); err != nil {
		t.Fatal(err)
	}

	cfs := &PathCfs{Value: "internal/config/config.go"}
	result, err := PathCfsToOs(cfs)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !filepath.IsAbs(result.Value) {
		t.Fatalf("expected absolute path, got: %s", result.Value)
	}
	wantSuffix := filepath.Join("internal", "config", "config.go")
	if !strings.HasSuffix(result.Value, wantSuffix) {
		t.Fatalf("expected path ending in %s, got: %s", wantSuffix, result.Value)
	}
}

// TestPathCfsToOs_ValidNonExistingPath covers TC-CO-02.
func TestPathCfsToOs_ValidNonExistingPath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	cfs := &PathCfs{Value: "internal/newdir/newfile.go"}
	result, err := PathCfsToOs(cfs)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !filepath.IsAbs(result.Value) {
		t.Fatalf("expected absolute path, got: %s", result.Value)
	}
	wantSuffix := filepath.Join("internal", "newdir", "newfile.go")
	if !strings.HasSuffix(result.Value, wantSuffix) {
		t.Fatalf("expected path ending in %s, got: %s", wantSuffix, result.Value)
	}
}

// TestPathCfsToOs_DuplicateSlashes covers TC-CO-03.
func TestPathCfsToOs_DuplicateSlashes(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	cfs := &PathCfs{Value: "internal//config.go"}
	result, err := PathCfsToOs(cfs)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !filepath.IsAbs(result.Value) {
		t.Fatalf("expected absolute path, got: %s", result.Value)
	}
}

// TestPathCfsToOs_RejectsDirectoryTraversal covers TC-CO-04.
func TestPathCfsToOs_RejectsDirectoryTraversal(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	cfs := &PathCfs{Value: "../../etc/passwd"}
	_, err := PathCfsToOs(cfs)
	if !errors.Is(err, ErrDirectoryTraversal) {
		t.Fatalf("expected ErrDirectoryTraversal, got: %v", err)
	}
}

// TestPathCfsToOs_RejectsSymlinkEscapingRoot covers TC-CO-05.
func TestPathCfsToOs_RejectsSymlinkEscapingRoot(t *testing.T) {
	dir := t.TempDir()
	outside := t.TempDir()
	testChdir(t, dir)

	// Create a symlink inside the project root pointing outside.
	symlinkName := "escaping-link"
	symlinkPath := filepath.Join(dir, symlinkName)
	if err := os.Symlink(outside, symlinkPath); err != nil {
		t.Skip("cannot create symlink on this platform:", err)
	}

	cfs := &PathCfs{Value: symlinkName + "/secret.txt"}
	_, err := PathCfsToOs(cfs)
	if !errors.Is(err, ErrResolvesOutsideRoot) {
		t.Fatalf("expected ErrResolvesOutsideRoot, got: %v", err)
	}
}

// TestPathCfsToOs_Roundtrip covers TC-CO-06.
func TestPathCfsToOs_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	const cfsValue = "internal/config/config.go"
	cfs := &PathCfs{Value: cfsValue}

	osPath, err := PathCfsToOs(cfs)
	if err != nil {
		t.Fatalf("PathCfsToOs error: %v", err)
	}

	cfsBack, err := PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("PathOsToCfs error: %v", err)
	}

	if cfsBack.Value != cfsValue {
		t.Fatalf("roundtrip mismatch: got %q, want %q", cfsBack.Value, cfsValue)
	}
}

// TestPathOsToCfs_ValidExistingPath covers TC-OC-01.
func TestPathOsToCfs_ValidExistingPath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	subDir := filepath.Join(dir, "some", "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(subDir, "file.go")
	if err := os.WriteFile(filePath, []byte("package sub"), 0644); err != nil {
		t.Fatal(err)
	}

	osPath := &PathOs{Value: filePath}
	result, err := PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Value == "" {
		t.Fatal("expected non-empty CFS value")
	}
	if !strings.HasPrefix(result.Value, "some/subdir") {
		t.Fatalf("unexpected CFS value: %s", result.Value)
	}
}

// TestPathOsToCfs_ValidNonExistingPath covers TC-OC-02.
func TestPathOsToCfs_ValidNonExistingPath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	absPath := filepath.Join(dir, "nonexistent", "file.go")
	osPath := &PathOs{Value: absPath}
	result, err := PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Value != "nonexistent/file.go" {
		t.Fatalf("unexpected CFS value: %s", result.Value)
	}
}

// TestPathOsToCfs_ForwardSlashes covers TC-OC-03.
func TestPathOsToCfs_ForwardSlashes(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// Build a native OS path inside the root.
	absPath := filepath.Join(dir, "a", "b", "c.go")
	osPath := &PathOs{Value: absPath}
	result, err := PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if strings.Contains(result.Value, `\`) {
		t.Fatalf("CFS value contains backslash: %s", result.Value)
	}
}

// TestPathOsToCfs_SymlinkWithinRoot covers TC-OC-04.
func TestPathOsToCfs_SymlinkWithinRoot(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// Create a real directory and a symlink to it, both inside the root.
	targetDir := filepath.Join(dir, "real-target")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatal(err)
	}
	symlinkPath := filepath.Join(dir, "sym-link")
	if err := os.Symlink(targetDir, symlinkPath); err != nil {
		t.Skip("cannot create symlink on this platform:", err)
	}

	osPath := &PathOs{Value: symlinkPath}
	result, err := PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Value == "" {
		t.Fatal("expected non-empty CFS value")
	}
}

// TestPathOsToCfs_RejectsPathOutsideRoot covers TC-OC-05.
func TestPathOsToCfs_RejectsPathOutsideRoot(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	outside := t.TempDir()
	osPath := &PathOs{Value: filepath.Join(outside, "secret.txt")}
	_, err := PathOsToCfs(osPath)
	if !errors.Is(err, ErrResolvesOutsideRoot) {
		t.Fatalf("expected ErrResolvesOutsideRoot, got: %v", err)
	}
}

// TestPathGetProjectRoot_ReturnsAbsolutePath covers TC-GR-01.
func TestPathGetProjectRoot_ReturnsAbsolutePath(t *testing.T) {
	result, err := PathGetProjectRoot()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Value == "" {
		t.Fatal("expected non-empty path")
	}
	if !filepath.IsAbs(result.Value) {
		t.Fatalf("expected absolute path, got: %s", result.Value)
	}
}

// TestPathGetProjectRoot_MatchesWorkingDirectory covers TC-GR-02.
func TestPathGetProjectRoot_MatchesWorkingDirectory(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	result, err := PathGetProjectRoot()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Resolve both to account for symlinks.
	resolvedWd, err := filepath.EvalSymlinks(wd)
	if err != nil {
		resolvedWd = wd
	}
	resolvedResult, err := filepath.EvalSymlinks(result.Value)
	if err != nil {
		resolvedResult = result.Value
	}

	if resolvedResult != resolvedWd {
		t.Fatalf("project root %q does not match working directory %q", resolvedResult, resolvedWd)
	}
}

// testChdir changes the working directory to dir for the duration of the
// test and restores it when the test finishes.
func testChdir(t *testing.T, dir string) {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	})
}
