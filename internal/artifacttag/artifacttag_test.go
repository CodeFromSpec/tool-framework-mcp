// code-from-spec: ROOT/golang/internal/artifact_tag/tests@HQNUmFzfhKhEkENh00JXChy3xiU

package artifacttag

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// testWriteFile is a helper that writes content to a named file inside dir
// and returns the full path. It fails the test immediately on any error.
func testWriteFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
	return path
}

// ---------------------------------------------------------------------------
// Happy-path tests
// ---------------------------------------------------------------------------

// TestExtractArtifactTag_GoComment verifies that a tag embedded in a Go-style
// line comment is extracted correctly.
func TestExtractArtifactTag_GoComment(t *testing.T) {
	dir := t.TempDir()
	path := testWriteFile(t, dir, "gocomment.go",
		"// code-from-spec: ROOT/golang/internal/foo/code(bar)@abcdefghijklmnopqrstuvwxyza\n",
	)

	tag, err := ExtractArtifactTag(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.LogicalName != "ROOT/golang/internal/foo/code(bar)" {
		t.Errorf("LogicalName = %q, want %q", tag.LogicalName, "ROOT/golang/internal/foo/code(bar)")
	}
	if tag.Hash != "abcdefghijklmnopqrstuvwxyza" {
		t.Errorf("Hash = %q, want %q", tag.Hash, "abcdefghijklmnopqrstuvwxyza")
	}
}

// TestExtractArtifactTag_HashComment verifies that a tag embedded in a
// shell/hash-style comment is extracted correctly.
func TestExtractArtifactTag_HashComment(t *testing.T) {
	dir := t.TempDir()
	path := testWriteFile(t, dir, "hashcomment.sh",
		"# code-from-spec: ROOT/some/node(id)@123456789012345678901234567\n",
	)

	tag, err := ExtractArtifactTag(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.LogicalName != "ROOT/some/node(id)" {
		t.Errorf("LogicalName = %q, want %q", tag.LogicalName, "ROOT/some/node(id)")
	}
	if tag.Hash != "123456789012345678901234567" {
		t.Errorf("Hash = %q, want %q", tag.Hash, "123456789012345678901234567")
	}
}

// TestExtractArtifactTag_StopsAtFirstMatch verifies that when multiple
// code-from-spec lines are present only the very first one is returned.
func TestExtractArtifactTag_StopsAtFirstMatch(t *testing.T) {
	dir := t.TempDir()
	// Both lines are valid tags; the function must return the first one.
	content := "" +
		"// code-from-spec: ROOT/first/node@aaaaaaaaaaaaaaaaaaaaaaaaaa1\n" +
		"// code-from-spec: ROOT/second/node@bbbbbbbbbbbbbbbbbbbbbbbbbbb\n"
	path := testWriteFile(t, dir, "multi.go", content)

	tag, err := ExtractArtifactTag(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.LogicalName != "ROOT/first/node" {
		t.Errorf("LogicalName = %q, want first match %q", tag.LogicalName, "ROOT/first/node")
	}
	if tag.Hash != "aaaaaaaaaaaaaaaaaaaaaaaaaa1" {
		t.Errorf("Hash = %q, want first match hash", tag.Hash)
	}
}

// TestExtractArtifactTag_TagOnNonFirstLine verifies that the function finds a
// tag even when it does not appear on the very first line of the file.
func TestExtractArtifactTag_TagOnNonFirstLine(t *testing.T) {
	dir := t.TempDir()
	content := "" +
		"package main\n" +
		"\n" +
		"// code-from-spec: ROOT/some/deep/node@zzzzzzzzzzzzzzzzzzzzzzzzzzz\n"
	path := testWriteFile(t, dir, "nonFirst.go", content)

	tag, err := ExtractArtifactTag(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.LogicalName != "ROOT/some/deep/node" {
		t.Errorf("LogicalName = %q, want %q", tag.LogicalName, "ROOT/some/deep/node")
	}
	if tag.Hash != "zzzzzzzzzzzzzzzzzzzzzzzzzzz" {
		t.Errorf("Hash = %q, want %q", tag.Hash, "zzzzzzzzzzzzzzzzzzzzzzzzzzz")
	}
}

// ---------------------------------------------------------------------------
// Failure-case tests — table-driven where cases share the same assertion shape
// ---------------------------------------------------------------------------

// testCase describes a single failure scenario used in the table-driven test.
type testCase struct {
	name        string
	fileContent string // empty means "do not create a file"
	wantErr     error
}

func TestExtractArtifactTag_FailureCases(t *testing.T) {
	cases := []testCase{
		{
			// The tag is present but there is no '@' separator between the
			// logical name and the hash.
			name:        "malformed — no @ separator",
			fileContent: "// code-from-spec: ROOT/foo/bar\n",
			wantErr:     ErrMalformedTag,
		},
		{
			// The '@' separator is present but the logical name part is empty.
			name:        "malformed — empty logical name",
			fileContent: "// code-from-spec: @abcdefghijklmnopqrstuvwxyza\n",
			wantErr:     ErrMalformedTag,
		},
		{
			// The hash portion exists but is shorter than the required 27
			// characters.
			name:        "malformed — wrong hash length",
			fileContent: "// code-from-spec: ROOT/foo(bar)@short\n",
			wantErr:     ErrMalformedTag,
		},
		{
			// The file exists but contains no code-from-spec substring at all.
			name:        "no tag in file",
			fileContent: "package main\n\nfunc main() {}\n",
			wantErr:     ErrNoTagFound,
		},
	}

	for _, tc := range cases {
		tc := tc // capture loop variable
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := testWriteFile(t, dir, "subject.go", tc.fileContent)

			_, err := ExtractArtifactTag(path)
			if err == nil {
				t.Fatalf("expected an error wrapping %v, got nil", tc.wantErr)
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("errors.Is(err, %v) = false; err = %v", tc.wantErr, err)
			}
		})
	}
}

// TestExtractArtifactTag_FileDoesNotExist verifies that a missing file causes
// ErrFileUnreadable to be returned. This is kept separate from the table above
// because the file must NOT be created.
func TestExtractArtifactTag_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	// Deliberately reference a path that was never created.
	path := filepath.Join(dir, "does_not_exist.go")

	_, err := ExtractArtifactTag(path)
	if err == nil {
		t.Fatal("expected ErrFileUnreadable, got nil")
	}
	if !errors.Is(err, ErrFileUnreadable) {
		t.Errorf("errors.Is(err, ErrFileUnreadable) = false; err = %v", err)
	}
}
