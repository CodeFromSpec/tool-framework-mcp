// code-from-spec: ROOT/golang/tests/os/path_utils@eiUAiacgPwJ6oeZ1hPCbLELFQdU

package pathutils

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

	cases := []testCase{
		// Valid paths
		{name: "TC-PV-01 valid simple relative path", input: "internal/config/config.go", wantErr: nil},
		{name: "TC-PV-02 valid nested path", input: "cmd/framework-mcp/main.go", wantErr: nil},
		{name: "TC-PV-03 valid single filename", input: "main.go", wantErr: nil},
		{name: "TC-PV-04 accepts path with dot segment", input: "internal/./config/config.go", wantErr: nil},
		{name: "TC-PV-05 accepts traversal that resolves within root", input: "a/b/../c", wantErr: nil},
		{name: "TC-PV-06 accepts path with trailing slash", input: "internal/config/", wantErr: nil},
		{name: "TC-PV-07 accepts path with duplicate slashes", input: "internal//config//file.go", wantErr: nil},
		// Invalid paths
		{name: "TC-PV-08 rejects empty string", input: "", wantErr: ErrPathEmpty},
		{name: "TC-PV-09 rejects absolute path with leading slash", input: "/etc/passwd", wantErr: ErrPathAbsolute},
		{name: "TC-PV-10 rejects absolute path with drive letter", input: "C:/Windows/system32", wantErr: ErrPathAbsolute},
		{name: "TC-PV-11 rejects backslash", input: `internal\config\config.go`, wantErr: ErrPathContainsBackslash},
		{name: "TC-PV-12 rejects simple traversal", input: "../../etc/passwd", wantErr: ErrDirectoryTraversal},
		{name: "TC-PV-13 rejects embedded traversal", input: "internal/../../outside/file.go", wantErr: ErrDirectoryTraversal},
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

// ---------------------------------------------------------------------------
// PathCfsToOs
// ---------------------------------------------------------------------------

// TC-CO-01: Converts valid path that exists.
// We use an actual file in the project root (the module file always exists).
func TestPathCfsToOs_ExistingFile(t *testing.T) {
	cfs := &PathCfs{Value: "go.mod"}
	result, err := PathCfsToOs(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(result.Value) {
		t.Fatalf("expected absolute path, got: %s", result.Value)
	}
	nativeSuffix := filepath.FromSlash("go.mod")
	if !strings.HasSuffix(result.Value, nativeSuffix) {
		t.Fatalf("expected path to end with %q, got: %s", nativeSuffix, result.Value)
	}
}

// TC-CO-02: Converts valid path that does not exist.
func TestPathCfsToOs_NonExistingFile(t *testing.T) {
	cfs := &PathCfs{Value: "internal/newdir/newfile.go"}
	result, err := PathCfsToOs(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(result.Value) {
		t.Fatalf("expected absolute path, got: %s", result.Value)
	}
	nativeSuffix := filepath.FromSlash("internal/newdir/newfile.go")
	if !strings.HasSuffix(result.Value, nativeSuffix) {
		t.Fatalf("expected path to end with %q, got: %s", nativeSuffix, result.Value)
	}
}

// TC-CO-03: Converts path with duplicate slashes.
func TestPathCfsToOs_DuplicateSlashes(t *testing.T) {
	cfs := &PathCfs{Value: "internal//config.go"}
	result, err := PathCfsToOs(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(result.Value) {
		t.Fatalf("expected absolute path, got: %s", result.Value)
	}
}

// TC-CO-04: Rejects invalid CfsPath — directory traversal.
func TestPathCfsToOs_DirectoryTraversal(t *testing.T) {
	cfs := &PathCfs{Value: "../../etc/passwd"}
	_, err := PathCfsToOs(cfs)
	if !errors.Is(err, ErrDirectoryTraversal) {
		t.Fatalf("expected ErrDirectoryTraversal, got: %v", err)
	}
}

// TC-CO-05: Rejects symlink escaping project root.
func TestPathCfsToOs_SymlinkEscapesRoot(t *testing.T) {
	// Create an outside directory that the symlink will point to.
	outsideDir := t.TempDir()

	// Create a secret file in the outside directory.
	secretFile := filepath.Join(outsideDir, "secret.txt")
	if err := os.WriteFile(secretFile, []byte("secret"), 0600); err != nil {
		t.Fatalf("failed to create secret file: %v", err)
	}

	// Create a temp dir to act as a fake project root so we can write a
	// symlink into it without polluting the real working directory.
	fakeRoot := t.TempDir()

	// Create the symlink inside fakeRoot pointing outside.
	linkName := filepath.Join(fakeRoot, "escape_link")
	if err := os.Symlink(outsideDir, linkName); err != nil {
		t.Skipf("symlinks not supported on this platform: %v", err)
	}

	// Build a PathOs that points through the symlink and call PathOsToCfs
	// to verify the containment check triggers (PathCfsToOs relies on the
	// working directory for root, so we exercise the check via PathOsToCfs
	// using the real resolved path which is outside the fakeRoot).
	targetViaLink := filepath.Join(linkName, "secret.txt")
	osPath := &PathOs{Value: targetViaLink}

	// Override: We need to test PathCfsToOs containment. Since PathCfsToOs
	// uses os.Getwd() for the root, we construct the scenario using a real
	// symlink inside the actual working directory.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get working directory: %v", err)
	}

	// Create the symlink inside the actual working directory.
	linkInWd := filepath.Join(wd, "test_escape_link_co05")
	// Clean up before and after.
	_ = os.Remove(linkInWd)
	t.Cleanup(func() { _ = os.Remove(linkInWd) })

	if err := os.Symlink(outsideDir, linkInWd); err != nil {
		t.Skipf("symlinks not supported or insufficient permissions: %v", err)
	}

	// Now call PathCfsToOs with a path that goes through the symlink.
	cfs := &PathCfs{Value: "test_escape_link_co05/secret.txt"}
	_, err = PathCfsToOs(cfs)
	if !errors.Is(err, ErrResolvesOutsideRoot) {
		t.Fatalf("expected ErrResolvesOutsideRoot, got: %v", err)
	}

	// Suppress unused variable warning.
	_ = osPath
}

// TC-CO-06: Roundtrip — CfsToOs then OsToCfs.
func TestPathCfsToOs_Roundtrip(t *testing.T) {
	original := "internal/config/config.go"
	cfs := &PathCfs{Value: original}

	osPath, err := PathCfsToOs(cfs)
	if err != nil {
		t.Fatalf("PathCfsToOs error: %v", err)
	}

	cfsBack, err := PathOsToCfs(osPath)
	if err != nil {
		t.Fatalf("PathOsToCfs error: %v", err)
	}

	if cfsBack.Value != original {
		t.Fatalf("expected %q, got %q", original, cfsBack.Value)
	}
}

// ---------------------------------------------------------------------------
// PathOsToCfs
// ---------------------------------------------------------------------------

// testMakeFileInRoot creates a temporary file inside the project working
// directory and returns its absolute OS path. The file is removed when the
// test ends.
func testMakeFileInRoot(t *testing.T, relPath string, content []byte) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get working directory: %v", err)
	}
	absPath := filepath.Join(wd, relPath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		t.Fatalf("failed to create directories: %v", err)
	}
	if err := os.WriteFile(absPath, content, 0600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(absPath) })
	return absPath
}

// TC-OC-01: Converts valid OS path that exists.
func TestPathOsToCfs_ExistingFile(t *testing.T) {
	absPath := testMakeFileInRoot(t, "test_oc01_file.txt", []byte("hello"))

	result, err := PathOsToCfs(&PathOs{Value: absPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Value == "" {
		t.Fatal("expected non-empty CFS path")
	}
	if strings.ContainsRune(result.Value, '\\') {
		t.Fatalf("CFS path contains backslash: %s", result.Value)
	}
	if !strings.HasSuffix(result.Value, "test_oc01_file.txt") {
		t.Fatalf("expected path to end with test_oc01_file.txt, got: %s", result.Value)
	}
}

// TC-OC-02: Converts valid OS path that does not exist.
func TestPathOsToCfs_NonExistingFile(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get working directory: %v", err)
	}
	absPath := filepath.Join(wd, "nonexistent_dir", "nonexistent_file.go")

	result, err := PathOsToCfs(&PathOs{Value: absPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Value == "" {
		t.Fatal("expected non-empty CFS path")
	}
	if strings.ContainsRune(result.Value, '\\') {
		t.Fatalf("CFS path contains backslash: %s", result.Value)
	}
}

// TC-OC-03: Result uses forward slashes.
func TestPathOsToCfs_ForwardSlashes(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get working directory: %v", err)
	}
	// Build a path using the OS-native separator.
	absPath := filepath.Join(wd, "some", "nested", "path.go")

	result, err := PathOsToCfs(&PathOs{Value: absPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.ContainsRune(result.Value, '\\') {
		t.Fatalf("CFS path contains backslash: %s", result.Value)
	}
}

// TC-OC-04: Symlink within root resolving within root.
func TestPathOsToCfs_SymlinkWithinRoot(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get working directory: %v", err)
	}

	// Create a real file inside the project root.
	targetName := "test_oc04_target.txt"
	targetPath := testMakeFileInRoot(t, targetName, []byte("target"))

	// Create a symlink inside the project root pointing to the target.
	linkName := filepath.Join(wd, "test_oc04_link.txt")
	_ = os.Remove(linkName)
	t.Cleanup(func() { _ = os.Remove(linkName) })

	if err := os.Symlink(targetPath, linkName); err != nil {
		t.Skipf("symlinks not supported or insufficient permissions: %v", err)
	}

	result, err := PathOsToCfs(&PathOs{Value: linkName})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Value == "" {
		t.Fatal("expected non-empty CFS path")
	}
}

// TC-OC-05: Rejects path outside project root.
func TestPathOsToCfs_OutsideRoot(t *testing.T) {
	outsideDir := t.TempDir()
	outsidePath := filepath.Join(outsideDir, "outside.txt")

	_, err := PathOsToCfs(&PathOs{Value: outsidePath})
	if !errors.Is(err, ErrResolvesOutsideRoot) {
		t.Fatalf("expected ErrResolvesOutsideRoot, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// PathGetProjectRoot
// ---------------------------------------------------------------------------

// TC-GR-01: Returns an absolute path.
func TestPathGetProjectRoot_ReturnsAbsolutePath(t *testing.T) {
	root, err := PathGetProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root.Value == "" {
		t.Fatal("expected non-empty root path")
	}
	if !filepath.IsAbs(root.Value) {
		t.Fatalf("expected absolute path, got: %s", root.Value)
	}
}

// TC-GR-02: Matches working directory.
func TestPathGetProjectRoot_MatchesWorkingDirectory(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get working directory: %v", err)
	}

	root, err := PathGetProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if root.Value != wd {
		t.Fatalf("expected %q, got %q", wd, root.Value)
	}
}
