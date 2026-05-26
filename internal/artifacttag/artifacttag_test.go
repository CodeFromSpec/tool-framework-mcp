// code-from-spec: ROOT/golang/internal/artifact_tag/tests@WwYRrnAlpDp1Z0Jk6M0cr6o2-SA

// Package artifacttag contains tests for ExtractArtifactTag.
//
// These tests live in the same package as the implementation (internal test
// file) so they can reference unexported constants such as hashLength if
// needed — though all assertions go through the public API.
//
// File-handle hygiene note
// ========================
// ExtractArtifactTag is responsible for opening AND closing the underlying
// filereader.FileReader. The tests never open a FileReader directly, so
// there are no extra file handles to close from the test side. If temp-file
// cleanup fails on Windows (e.g. "The process cannot access the file because
// it is being used by another process"), that indicates a handle leak inside
// ExtractArtifactTag — the implementation must call defer r.Close() after
// opening the reader.
package artifacttag

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// Test helper types and functions
// ---------------------------------------------------------------------------

// testCase describes a single table-driven test entry.
type testCase struct {
	name string
	// content is written verbatim to the temp file.
	// If empty and filePath is set, filePath is used directly (non-existent file).
	content     string
	useRealFile bool // when false, filePath is a non-existent path

	wantErr         bool
	wantSentinel    error  // checked with errors.Is when wantErr is true
	wantLogicalName string // checked when wantErr is false
	wantHash        string // checked when wantErr is false
}

// testWriteFile writes content into a new file inside dir and returns its
// absolute path. It calls t.Fatal on any write failure.
func testWriteFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("testWriteFile: could not write %s: %v", path, err)
	}
	return path
}

// ---------------------------------------------------------------------------
// Happy-path tests
// ---------------------------------------------------------------------------

// TestExtractArtifactTag_HappyPath covers all successful extraction cases.
func TestExtractArtifactTag_HappyPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		content         string
		wantLogicalName string
		wantHash        string
	}{
		{
			// Spec: "Extracts tag from Go comment"
			// The tag appears on the first line in a // comment.
			name:            "go_line_comment",
			content:         "// code-from-spec: ROOT/golang/internal/foo/code(bar)@abcdefghijklmnopqrstuvwxyza\n",
			wantLogicalName: "ROOT/golang/internal/foo/code(bar)",
			wantHash:        "abcdefghijklmnopqrstuvwxyza",
		},
		{
			// Spec: "Extracts tag from hash comment"
			// The tag appears in a # comment (shell/YAML style).
			name:            "hash_comment",
			content:         "# code-from-spec: ROOT/some/node(id)@123456789012345678901234567\n",
			wantLogicalName: "ROOT/some/node(id)",
			wantHash:        "123456789012345678901234567",
		},
		{
			// Spec: "Stops reading at first match"
			// The file has two valid code-from-spec lines; only the first is returned.
			name: "first_match_wins",
			content: "// code-from-spec: ROOT/first/node@aaaaaaaaaaaaaaaaaaaaaaaaaa1\n" +
				"// code-from-spec: ROOT/second/node@bbbbbbbbbbbbbbbbbbbbbbbbbbb\n",
			wantLogicalName: "ROOT/first/node",
			wantHash:        "aaaaaaaaaaaaaaaaaaaaaaaaaa1",
		},
		{
			// Spec: "Tag on non-first line"
			// Lines 1–2 are plain text; the tag is on line 3.
			name: "tag_on_third_line",
			content: "package main\n" +
				"\n" +
				"// code-from-spec: ROOT/golang/server@ccccccccccccccccccccccccccc\n",
			wantLogicalName: "ROOT/golang/server",
			wantHash:        "ccccccccccccccccccccccccccc",
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			path := testWriteFile(t, dir, "input.go", tc.content)

			// ExtractArtifactTag opens and closes the file reader internally.
			// No extra handle is held here.
			got, err := ExtractArtifactTag(path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.LogicalName != tc.wantLogicalName {
				t.Errorf("LogicalName: got %q, want %q", got.LogicalName, tc.wantLogicalName)
			}
			if got.Hash != tc.wantHash {
				t.Errorf("Hash: got %q, want %q", got.Hash, tc.wantHash)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Failure-case tests
// ---------------------------------------------------------------------------

// TestExtractArtifactTag_FileDoesNotExist verifies that a missing file
// causes ErrFileUnreadable to be returned (wrapped).
func TestExtractArtifactTag_FileDoesNotExist(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Deliberately point at a path that was never created.
	nonExistent := filepath.Join(dir, "does_not_exist.go")

	_, err := ExtractArtifactTag(nonExistent)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, ErrFileUnreadable) {
		t.Errorf("errors.Is(err, ErrFileUnreadable) = false; err = %v", err)
	}
}

// TestExtractArtifactTag_NoTag verifies that a file without any
// "code-from-spec:" substring causes ErrNoTagFound to be returned.
func TestExtractArtifactTag_NoTag(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	content := "package main\n\nfunc main() {}\n"
	path := testWriteFile(t, dir, "no_tag.go", content)

	_, err := ExtractArtifactTag(path)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, ErrNoTagFound) {
		t.Errorf("errors.Is(err, ErrNoTagFound) = false; err = %v", err)
	}
}

// TestExtractArtifactTag_MalformedTag covers all malformed-tag variants
// using a table-driven approach.
func TestExtractArtifactTag_MalformedTag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
	}{
		{
			// Spec: "Malformed tag — no @ separator"
			// The tag is present but has no "@" separating name from hash.
			name:    "no_at_separator",
			content: "// code-from-spec: ROOT/foo/bar\n",
		},
		{
			// Spec: "Malformed tag — empty logical name"
			// The "@" is present but nothing precedes it.
			name:    "empty_logical_name",
			content: "// code-from-spec: @abcdefghijklmnopqrstuvwxyza\n",
		},
		{
			// Spec: "Malformed tag — wrong hash length"
			// The logical name is valid but the hash is too short (5 chars, not 27).
			name:    "wrong_hash_length",
			content: "// code-from-spec: ROOT/foo(bar)@short\n",
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			path := testWriteFile(t, dir, "malformed.go", tc.content)

			_, err := ExtractArtifactTag(path)
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
			if !errors.Is(err, ErrMalformedTag) {
				t.Errorf("errors.Is(err, ErrMalformedTag) = false; err = %v", err)
			}
		})
	}
}
