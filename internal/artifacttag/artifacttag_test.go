// code-from-spec: ROOT/golang/tests/parsing/artifact_tag@oiuulYpIbF4kdfEiJv_VHqAQi5w

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

// testWriteFile writes content to a file at path (relative to cwd), creating parent dirs.
func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

// --- Happy Path ---

func TestArtifactTagExtract_SlashSlashComment(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "file.go", "// code-from-spec: ROOT/golang/implementation/internal/foo/code(bar)@abcdefghijklmnopqrstuvwxyza\n")

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

func TestArtifactTagExtract_HashComment(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "file.py", "# code-from-spec: ROOT/some/node(id)@123456789012345678901234567\n")

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

func TestArtifactTagExtract_HTMLComment(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "file.md", "<!-- code-from-spec: ROOT/docs/readme@abcdefghijklmnopqrstuvwxyza -->\n")

	tag, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.md"})
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

func TestArtifactTagExtract_StopsAtFirstMatch(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "// code-from-spec: ROOT/first/node@aaaaaaaaaaaaaaaaaaaaaaaaaa1\n" +
		"// code-from-spec: ROOT/second/node@bbbbbbbbbbbbbbbbbbbbbbbbbbb\n"
	testWriteFile(t, "file.go", content)

	tag, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.LogicalName != "ROOT/first/node" {
		t.Errorf("LogicalName = %q, want %q", tag.LogicalName, "ROOT/first/node")
	}
	if tag.Hash != "aaaaaaaaaaaaaaaaaaaaaaaaaa1" {
		t.Errorf("Hash = %q, want %q", tag.Hash, "aaaaaaaaaaaaaaaaaaaaaaaaaa1")
	}
}

func TestArtifactTagExtract_TagOnNonFirstLine(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "Some preamble text\n" +
		"Another line of text\n" +
		"// code-from-spec: ROOT/some/node@abcdefghijklmnopqrstuvwxyza\n"
	testWriteFile(t, "file.go", content)

	tag, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.LogicalName != "ROOT/some/node" {
		t.Errorf("LogicalName = %q, want %q", tag.LogicalName, "ROOT/some/node")
	}
	if tag.Hash != "abcdefghijklmnopqrstuvwxyza" {
		t.Errorf("Hash = %q, want %q", tag.Hash, "abcdefghijklmnopqrstuvwxyza")
	}
}

func TestArtifactTagExtract_ExtraWhitespaceBeforeLogicalName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "file.go", "// code-from-spec:   ROOT/x(y)@abcdefghijklmnopqrstuvwxyza\n")

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

// --- Edge Cases ---

func TestArtifactTagExtract_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "empty.go", "")

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "empty.go"})
	if !errors.Is(err, artifacttag.ErrNoTagFound) {
		t.Errorf("err = %v, want ErrNoTagFound", err)
	}
}

// --- Failure Cases ---

func TestArtifactTagExtract_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "nonexistent/file.go"})
	if !errors.Is(err, artifacttag.ErrFileUnreadable) {
		t.Errorf("err = %v, want ErrFileUnreadable", err)
	}
}

func TestArtifactTagExtract_PropagatesPathErrors(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "../../outside"})
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("err = %v, want ErrDirectoryTraversal", err)
	}
}

func TestArtifactTagExtract_NoTagInFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "This file has no artifact tag.\nJust some regular content here.\n"
	testWriteFile(t, "file.go", content)

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrNoTagFound) {
		t.Errorf("err = %v, want ErrNoTagFound", err)
	}
}

func TestArtifactTagExtract_MalformedTag_NoAtSeparator(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "file.go", "// code-from-spec: ROOT/foo/bar\n")

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("err = %v, want ErrMalformedTag", err)
	}
}

func TestArtifactTagExtract_MalformedTag_EmptyLogicalName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "file.go", "// code-from-spec: @abcdefghijklmnopqrstuvwxyza\n")

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("err = %v, want ErrMalformedTag", err)
	}
}

func TestArtifactTagExtract_MalformedTag_WrongHashLength(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "file.go", "// code-from-spec: ROOT/foo(bar)@short\n")

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("err = %v, want ErrMalformedTag", err)
	}
}
