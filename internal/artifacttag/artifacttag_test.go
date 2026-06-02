// code-from-spec: ROOT/golang/tests/parsing/artifact_tag@4RBjx-_GCCHGtpN4DTN_3AH1wzA
package artifacttag_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/artifacttag"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

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
		t.Errorf("logical name = %q, want %q", tag.LogicalName, "ROOT/golang/implementation/internal/foo/code(bar)")
	}
	if tag.Hash != "abcdefghijklmnopqrstuvwxyza" {
		t.Errorf("hash = %q, want %q", tag.Hash, "abcdefghijklmnopqrstuvwxyza")
	}
}

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
		t.Errorf("logical name = %q, want %q", tag.LogicalName, "ROOT/some/node(id)")
	}
	if tag.Hash != "123456789012345678901234567" {
		t.Errorf("hash = %q, want %q", tag.Hash, "123456789012345678901234567")
	}
}

func TestArtifactTagExtract_HTMLComment(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "<!-- code-from-spec: ROOT/docs/readme@abcdefghijklmnopqrstuvwxyza -->\n"
	if err := os.WriteFile("file.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	tag, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.LogicalName != "ROOT/docs/readme" {
		t.Errorf("logical name = %q, want %q", tag.LogicalName, "ROOT/docs/readme")
	}
	if tag.Hash != "abcdefghijklmnopqrstuvwxyza" {
		t.Errorf("hash = %q, want %q", tag.Hash, "abcdefghijklmnopqrstuvwxyza")
	}
}

func TestArtifactTagExtract_StopsAtFirstMatch(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "// code-from-spec: ROOT/first/node@abcdefghijklmnopqrstuvwxyza\n// code-from-spec: ROOT/second/node@bcdefghijklmnopqrstuvwxyzab\n"
	if err := os.WriteFile("file.go", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	tag, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.LogicalName != "ROOT/first/node" {
		t.Errorf("logical name = %q, want %q", tag.LogicalName, "ROOT/first/node")
	}
	if tag.Hash != "abcdefghijklmnopqrstuvwxyza" {
		t.Errorf("hash = %q, want %q", tag.Hash, "abcdefghijklmnopqrstuvwxyza")
	}
}

func TestArtifactTagExtract_TagOnNonFirstLine(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "line one\nline two\n// code-from-spec: ROOT/x/y@abcdefghijklmnopqrstuvwxyza\n"
	if err := os.WriteFile("file.go", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	tag, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.LogicalName != "ROOT/x/y" {
		t.Errorf("logical name = %q, want %q", tag.LogicalName, "ROOT/x/y")
	}
	if tag.Hash != "abcdefghijklmnopqrstuvwxyza" {
		t.Errorf("hash = %q, want %q", tag.Hash, "abcdefghijklmnopqrstuvwxyza")
	}
}

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
		t.Errorf("logical name = %q, want %q", tag.LogicalName, "ROOT/x(y)")
	}
	if tag.Hash != "abcdefghijklmnopqrstuvwxyza" {
		t.Errorf("hash = %q, want %q", tag.Hash, "abcdefghijklmnopqrstuvwxyza")
	}
}

func TestArtifactTagExtract_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("empty.go", []byte(""), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "empty.go"})
	if !errors.Is(err, artifacttag.ErrNoTagFound) {
		t.Errorf("error = %v, want ErrNoTagFound", err)
	}
}

func TestArtifactTagExtract_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "nonexistent/file.go"})
	if !errors.Is(err, artifacttag.ErrFileUnreadable) {
		t.Errorf("error = %v, want ErrFileUnreadable", err)
	}
}

func TestArtifactTagExtract_PropagatesPathErrors(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "../../outside"})
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("error = %v, want ErrDirectoryTraversal", err)
	}
}

func TestArtifactTagExtract_NoTagInFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "This is just a regular file.\nNo artifact tag here.\n"
	if err := os.WriteFile("file.go", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrNoTagFound) {
		t.Errorf("error = %v, want ErrNoTagFound", err)
	}
}

func TestArtifactTagExtract_MalformedTag_NoAtSeparator(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "// code-from-spec: ROOT/foo/bar\n"
	if err := os.WriteFile("file.go", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("error = %v, want ErrMalformedTag", err)
	}
}

func TestArtifactTagExtract_MalformedTag_EmptyLogicalName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "// code-from-spec: @abcdefghijklmnopqrstuvwxyza\n"
	if err := os.WriteFile("file.go", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("error = %v, want ErrMalformedTag", err)
	}
}

func TestArtifactTagExtract_MalformedTag_WrongHashLength(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "// code-from-spec: ROOT/foo(bar)@short\n"
	if err := os.WriteFile("file.go", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("error = %v, want ErrMalformedTag", err)
	}
}
