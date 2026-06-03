// code-from-spec: ROOT/golang/tests/parsing/artifact_tag@LIOqj1X7OUgc1lq6L65tp6DCxzM
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
	if err := os.WriteFile("file.go", []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	tag, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.LogicalName != "ROOT/golang/implementation/internal/foo/code(bar)" {
		t.Errorf("logical name = %q", tag.LogicalName)
	}
	if tag.Hash != "abcdefghijklmnopqrstuvwxyza" {
		t.Errorf("hash = %q", tag.Hash)
	}
}

func TestArtifactTagExtract_HashComment(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	content := "# code-from-spec: ROOT/some/node(id)@123456789012345678901234567\n"
	if err := os.WriteFile("file.py", []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	tag, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.py"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.LogicalName != "ROOT/some/node(id)" {
		t.Errorf("logical name = %q", tag.LogicalName)
	}
	if tag.Hash != "123456789012345678901234567" {
		t.Errorf("hash = %q", tag.Hash)
	}
}

func TestArtifactTagExtract_HTMLComment(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	content := "<!-- code-from-spec: ROOT/docs/readme@abcdefghijklmnopqrstuvwxyza -->\n"
	if err := os.WriteFile("file.md", []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	tag, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.LogicalName != "ROOT/docs/readme" {
		t.Errorf("logical name = %q", tag.LogicalName)
	}
	if tag.Hash != "abcdefghijklmnopqrstuvwxyza" {
		t.Errorf("hash = %q", tag.Hash)
	}
}

func TestArtifactTagExtract_StopsAtFirstMatch(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	content := "// code-from-spec: ROOT/first/node@abcdefghijklmnopqrstuvwxyza\n// code-from-spec: ROOT/second/node@zyxwvutsrqponmlkjihgfedcbaa\n"
	if err := os.WriteFile("file.go", []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	tag, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.LogicalName != "ROOT/first/node" {
		t.Errorf("logical name = %q", tag.LogicalName)
	}
	if tag.Hash != "abcdefghijklmnopqrstuvwxyza" {
		t.Errorf("hash = %q", tag.Hash)
	}
}

func TestArtifactTagExtract_TagOnNonFirstLine(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	content := "line one\nline two\n// code-from-spec: ROOT/some/node@abcdefghijklmnopqrstuvwxyza\n"
	if err := os.WriteFile("file.go", []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	tag, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.LogicalName != "ROOT/some/node" {
		t.Errorf("logical name = %q", tag.LogicalName)
	}
	if tag.Hash != "abcdefghijklmnopqrstuvwxyza" {
		t.Errorf("hash = %q", tag.Hash)
	}
}

func TestArtifactTagExtract_ExtraWhitespaceBeforeLogicalName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	content := "// code-from-spec:   ROOT/x(y)@abcdefghijklmnopqrstuvwxyza\n"
	if err := os.WriteFile("file.go", []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	tag, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.LogicalName != "ROOT/x(y)" {
		t.Errorf("logical name = %q", tag.LogicalName)
	}
	if tag.Hash != "abcdefghijklmnopqrstuvwxyza" {
		t.Errorf("hash = %q", tag.Hash)
	}
}

func TestArtifactTagExtract_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	if err := os.WriteFile("empty.go", []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "empty.go"})
	if !errors.Is(err, artifacttag.ErrNoTagFound) {
		t.Errorf("expected ErrNoTagFound, got %v", err)
	}
}

func TestArtifactTagExtract_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "nonexistent/file.txt"})
	if !errors.Is(err, artifacttag.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestArtifactTagExtract_PropagatesPathErrors(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "../../outside"})
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got %v", err)
	}
}

func TestArtifactTagExtract_NoTagInFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	content := "this file has no artifact tag\njust plain text\n"
	if err := os.WriteFile("file.go", []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrNoTagFound) {
		t.Errorf("expected ErrNoTagFound, got %v", err)
	}
}

func TestArtifactTagExtract_MalformedTagNoAtSeparator(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	content := "// code-from-spec: ROOT/foo/bar\n"
	if err := os.WriteFile("file.go", []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("expected ErrMalformedTag, got %v", err)
	}
}

func TestArtifactTagExtract_MalformedTagEmptyLogicalName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	content := "// code-from-spec: @abcdefghijklmnopqrstuvwxyza\n"
	if err := os.WriteFile("file.go", []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("expected ErrMalformedTag, got %v", err)
	}
}

func TestArtifactTagExtract_MalformedTagWrongHashLength(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)
	content := "// code-from-spec: ROOT/foo(bar)@short\n"
	if err := os.WriteFile("file.go", []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("expected ErrMalformedTag, got %v", err)
	}
}
