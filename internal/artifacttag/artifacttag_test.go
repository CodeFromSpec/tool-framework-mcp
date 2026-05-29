// code-from-spec: ROOT/golang/tests/parsing/artifact_tag@Hzs-FdToojXop1GLxqXLXn42Mpc

package artifacttag_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/artifacttag"
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

// TC-01: Extracts tag from slash-slash comment
func TestArtifactTagExtract_SlashSlashComment(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "// code-from-spec: ROOT/golang/implementation/internal/foo/code(bar)@abcdefghijklmnopqrstuvwxyza\n"
	if err := os.WriteFile("file.go", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	tag, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.LogicalName != "ROOT/golang/implementation/internal/foo/code(bar)" {
		t.Errorf("LogicalName = %q, want %q", tag.LogicalName, "ROOT/golang/implementation/internal/foo/code(bar)")
	}
	if tag.Hash != "abcdefghijklmnopqrstuvwxyza" {
		t.Errorf("Hash = %q, want %q", tag.Hash, "abcdefghijklmnopqrstuvwxyza")
	}
}

// TC-02: Extracts tag from hash comment
func TestArtifactTagExtract_HashComment(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# code-from-spec: ROOT/some/node(id)@123456789012345678901234567\n"
	if err := os.WriteFile("file.py", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	tag, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.py"})
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

// TC-03: Extracts tag from HTML comment
func TestArtifactTagExtract_HTMLComment(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "<!-- code-from-spec: ROOT/docs/readme@abcdefghijklmnopqrstuvwxyza -->\n"
	if err := os.WriteFile("README.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	tag, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "README.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.LogicalName != "ROOT/docs/readme" {
		t.Errorf("LogicalName = %q, want %q", tag.LogicalName, "ROOT/docs/readme")
	}
	if tag.Hash != "abcdefghijklmnopqrstuvwxyza" {
		t.Errorf("Hash = %q, want %q", tag.Hash, "abcdefghijklmnopqrstuvwxyza")
	}
}

// TC-04: Stops reading at first match
func TestArtifactTagExtract_StopsAtFirstMatch(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "// code-from-spec: ROOT/first/node@abcdefghijklmnopqrstuvwxyza\n" +
		"// code-from-spec: ROOT/second/node@zyxwvutsrqponmlkjihgfedcbaz\n"
	if err := os.WriteFile("file.go", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	tag, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.LogicalName != "ROOT/first/node" {
		t.Errorf("LogicalName = %q, want %q", tag.LogicalName, "ROOT/first/node")
	}
	if tag.Hash != "abcdefghijklmnopqrstuvwxyza" {
		t.Errorf("Hash = %q, want %q", tag.Hash, "abcdefghijklmnopqrstuvwxyza")
	}
}

// TC-05: Tag on non-first line
func TestArtifactTagExtract_TagOnNonFirstLine(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "Some content here.\nMore content here.\n// code-from-spec: ROOT/docs/readme@abcdefghijklmnopqrstuvwxyza\n"
	if err := os.WriteFile("file.go", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	tag, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.LogicalName != "ROOT/docs/readme" {
		t.Errorf("LogicalName = %q, want %q", tag.LogicalName, "ROOT/docs/readme")
	}
	if tag.Hash != "abcdefghijklmnopqrstuvwxyza" {
		t.Errorf("Hash = %q, want %q", tag.Hash, "abcdefghijklmnopqrstuvwxyza")
	}
}

// TC-06: Extra whitespace before logical name
func TestArtifactTagExtract_ExtraWhitespaceBeforeLogicalName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "// code-from-spec:   ROOT/x(y)@abcdefghijklmnopqrstuvwxyza\n"
	if err := os.WriteFile("file.go", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	tag, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.LogicalName != "ROOT/x(y)" {
		t.Errorf("LogicalName = %q, want %q", tag.LogicalName, "ROOT/x(y)")
	}
	if tag.Hash != "abcdefghijklmnopqrstuvwxyza" {
		t.Errorf("Hash = %q, want %q", tag.Hash, "abcdefghijklmnopqrstuvwxyza")
	}
}

// TC-07: Empty file
func TestArtifactTagExtract_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("empty.go", []byte{}, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "empty.go"})
	if !errors.Is(err, artifacttag.ErrNoTagFound) {
		t.Errorf("err = %v, want ErrNoTagFound", err)
	}
}

// TC-08: File does not exist
func TestArtifactTagExtract_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "nonexistent.go"})
	if !errors.Is(err, artifacttag.ErrFileUnreadable) {
		t.Errorf("err = %v, want ErrFileUnreadable", err)
	}
}

// TC-09: Propagates path errors (directory traversal)
func TestArtifactTagExtract_PropagatesPathErrors(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "../../outside"})
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("err = %v, want ErrDirectoryTraversal", err)
	}
}

// TC-10: No tag in file
func TestArtifactTagExtract_NoTagInFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "This file has no artifact tag at all.\nJust regular content.\n"
	if err := os.WriteFile("file.go", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrNoTagFound) {
		t.Errorf("err = %v, want ErrNoTagFound", err)
	}
}

// TC-11: Malformed tag — no @ separator
func TestArtifactTagExtract_MalformedTag_NoAtSeparator(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "// code-from-spec: ROOT/foo/bar\n"
	if err := os.WriteFile("file.go", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("err = %v, want ErrMalformedTag", err)
	}
}

// TC-12: Malformed tag — empty logical name
func TestArtifactTagExtract_MalformedTag_EmptyLogicalName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "// code-from-spec: @abcdefghijklmnopqrstuvwxyza\n"
	if err := os.WriteFile("file.go", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("err = %v, want ErrMalformedTag", err)
	}
}

// TC-13: Malformed tag — wrong hash length
func TestArtifactTagExtract_MalformedTag_WrongHashLength(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "// code-from-spec: ROOT/foo(bar)@short\n"
	if err := os.WriteFile("file.go", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("err = %v, want ErrMalformedTag", err)
	}
}
