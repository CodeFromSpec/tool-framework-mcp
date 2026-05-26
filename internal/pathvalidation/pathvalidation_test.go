// code-from-spec: ROOT/golang/internal/pathvalidation/tests@bCQpXvq-2DgFPeQIcx4m3glteUY

// Package pathvalidation contains tests for the ValidatePath function.
//
// Each test case uses t.TempDir() as the project root so the directory
// actually exists on disk (required by EvalSymlinks in the implementation).
// Tests are grouped by category: happy path, edge cases, and failure cases.
//
// All helper types and functions are prefixed with "test" per project convention
// to avoid collisions with unexported names in the package under test.
package pathvalidation

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// testCase describes a single ValidatePath scenario.
type testCase struct {
	// name is the human-readable test description shown by `go test -v`.
	name string

	// path is the caller-supplied path argument.
	path string

	// wantErr controls whether we expect an error from ValidatePath.
	wantErr bool

	// errContains is the substring that must appear in the error message.
	// Only checked when wantErr is true.
	errContains string
}

// testRunCases iterates over a slice of testCase values and exercises
// ValidatePath for each one using the provided projectRoot.
func testRunCases(t *testing.T, root string, cases []testCase) {
	t.Helper()
	for _, tc := range cases {
		tc := tc // capture loop variable for potential parallel use
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePath(tc.path, root)

			if tc.wantErr {
				// We expected an error — make sure we got one.
				if err == nil {
					t.Fatalf("ValidatePath(%q, root) = nil; want error containing %q",
						tc.path, tc.errContains)
				}
				// Verify the error message contains the expected substring.
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Fatalf("ValidatePath(%q, root) error = %q; want it to contain %q",
						tc.path, err.Error(), tc.errContains)
				}
			} else {
				// We expected success — any error is a failure.
				if err != nil {
					t.Fatalf("ValidatePath(%q, root) = %v; want nil", tc.path, err)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Happy-path tests — paths that should be accepted without error.
// ---------------------------------------------------------------------------

func TestValidatePath_HappyPath(t *testing.T) {
	root := t.TempDir()

	cases := []testCase{
		{
			// A typical Go source file nested two levels deep.
			name: "simple relative path",
			path: "internal/config/config.go",
		},
		{
			// Another nested path that must resolve inside root.
			name: "nested path",
			path: "cmd/framework-mcp/main.go",
		},
		{
			// A bare filename with no directory component.
			name: "single filename",
			path: "main.go",
		},
		{
			// A dot segment in the middle: internal/./config/config.go
			// filepath.Clean normalizes this to internal/config/config.go,
			// which is still safely inside the project root.
			name: "path with dot segment",
			path: "internal/./config/config.go",
		},
	}

	testRunCases(t, root, cases)
}

// ---------------------------------------------------------------------------
// Edge-case tests — unusual but valid paths that should still be accepted.
// ---------------------------------------------------------------------------

func TestValidatePath_EdgeCases(t *testing.T) {
	root := t.TempDir()

	cases := []testCase{
		{
			// A trailing slash is cleaned away by filepath.Clean; the result
			// resolves to a directory inside the project root, which is fine.
			name: "path with trailing slash",
			path: "internal/config/",
		},
		{
			// Duplicate separators are collapsed by filepath.Clean into a
			// single separator; the cleaned path is still inside the root.
			name: "path with duplicate separators",
			path: "internal//config//config.go",
		},
	}

	testRunCases(t, root, cases)
}

// ---------------------------------------------------------------------------
// Failure-case tests — paths that must be rejected with specific errors.
// ---------------------------------------------------------------------------

func TestValidatePath_FailureCases(t *testing.T) {
	root := t.TempDir()

	cases := []testCase{
		{
			// An empty string cannot name any file; must be caught first.
			name:        "empty path",
			path:        "",
			wantErr:     true,
			errContains: "path is empty",
		},
		{
			// A Unix absolute path must be rejected before symlink resolution.
			name:        "absolute path with leading slash",
			path:        "/etc/passwd",
			wantErr:     true,
			errContains: "path is absolute",
		},
		{
			// A Windows-style drive-letter path must be rejected on all
			// platforms (the check is intentionally platform-independent).
			name:        "absolute path with drive letter",
			path:        `C:\Windows\system32`,
			wantErr:     true,
			errContains: "path is absolute",
		},
		{
			// A leading ".." that stays at the top level after Clean.
			name:        "simple traversal",
			path:        "../../etc/passwd",
			wantErr:     true,
			errContains: "directory traversal",
		},
		{
			// A ".." embedded in an otherwise valid-looking path.
			// After Clean this becomes "../outside/file.go", which still
			// contains a ".." component.
			name:        "embedded traversal",
			path:        "internal/../../outside/file.go",
			wantErr:     true,
			errContains: "directory traversal",
		},
		{
			// A path that looks deeper but collapses to a traversal:
			// a/../../outside  →  Clean  →  ../outside
			name:        "traversal disguised with dot segments",
			path:        "a/../../outside",
			wantErr:     true,
			errContains: "directory traversal",
		},
	}

	testRunCases(t, root, cases)
}

// ---------------------------------------------------------------------------
// Symlink test — requires OS-level symlink creation.
// Skipped on platforms where os.Symlink is not available or requires
// elevated privileges (e.g., Windows without Developer Mode).
// ---------------------------------------------------------------------------

func TestValidatePath_SymlinkEscapesRoot(t *testing.T) {
	// Symlink creation can fail on Windows without elevated privileges or
	// Developer Mode enabled. Skip rather than fail in that environment.
	if runtime.GOOS == "windows" {
		t.Skip("skipping symlink test on Windows: may require elevated privileges")
	}

	root := t.TempDir()

	// "outside" is a real directory that lives outside the project root.
	// We use a separate TempDir so the OS guarantees it exists on disk
	// (EvalSymlinks requires the target to be reachable).
	outside := t.TempDir()

	// Create the symlink inside root/ pointing to the outside directory.
	// e.g.  <root>/escape -> <outside>/
	symlinkName := "escape"
	symlinkPath := filepath.Join(root, symlinkName)

	if err := os.Symlink(outside, symlinkPath); err != nil {
		t.Fatalf("failed to create symlink %q -> %q: %v", symlinkPath, outside, err)
	}

	// The path "escape/file.go" looks like a legitimate relative path, but
	// after EvalSymlinks it resolves to <outside>/file.go — outside root.
	err := ValidatePath(symlinkName+"/file.go", root)
	if err == nil {
		t.Fatalf("ValidatePath(%q, root) = nil; want error containing %q",
			symlinkName+"/file.go", "resolves outside project root")
	}
	if !strings.Contains(err.Error(), "resolves outside project root") {
		t.Fatalf("ValidatePath(%q, root) error = %q; want it to contain %q",
			symlinkName+"/file.go", err.Error(), "resolves outside project root")
	}
}
