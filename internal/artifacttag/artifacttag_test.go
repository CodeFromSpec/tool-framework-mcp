// code-from-spec: ROOT/golang/tests/parsing/artifact_tag@mA17Z2nlrKVwQ9nLV6C8FrCKR2M
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

func TestArtifactTagExtract_TC1_SlashSlashComment(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "// code-from-spec: ROOT/golang/implementation/internal/foo/code(bar)@abcdefghijklmnopqrstuvwxyza\n"
	if err := os.WriteFile("file.go", []byte(content), 0o644); err != nil {
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

func TestArtifactTagExtract_TC2_HashComment(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "# code-from-spec: ROOT/some/node(id)@123456789012345678901234567\n"
	if err := os.WriteFile("file.py", []byte(content), 0o644); err != nil {
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

func TestArtifactTagExtract_TC3_HTMLComment(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "<!-- code-from-spec: ROOT/docs/readme@abcdefghijklmnopqrstuvwxyza -->\n"
	if err := os.WriteFile("file.html", []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	tag, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.html"})
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

func TestArtifactTagExtract_TC4_StopsAtFirstMatch(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "// code-from-spec: ROOT/first/node@aaaaaaaaaaaaaaaaaaaaaaaaaa1\n" +
		"// code-from-spec: ROOT/second/node@bbbbbbbbbbbbbbbbbbbbbbbbbbb\n" +
		"// code-from-spec: ROOT/third/node@ccccccccccccccccccccccccccc\n"
	if err := os.WriteFile("file.go", []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

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

func TestArtifactTagExtract_TC5_TagOnNonFirstLine(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "package foo\n" +
		"\n" +
		"// code-from-spec: ROOT/some/node@abcdefghijklmnopqrstuvwxyza\n"
	if err := os.WriteFile("file.go", []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

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

func TestArtifactTagExtract_TC6_ExtraWhitespaceBeforeLogicalName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "// code-from-spec:   ROOT/x(y)@abcdefghijklmnopqrstuvwxyza\n"
	if err := os.WriteFile("file.go", []byte(content), 0o644); err != nil {
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

func TestArtifactTagExtract_TC7_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("empty.go", []byte{}, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "empty.go"})
	if !errors.Is(err, artifacttag.ErrNoTagFound) {
		t.Errorf("expected ErrNoTagFound, got %v", err)
	}
}

func TestArtifactTagExtract_TC8_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "nonexistent/file.go"})
	if !errors.Is(err, artifacttag.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestArtifactTagExtract_TC9_PropagatesPathErrors(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "../../outside"})
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got %v", err)
	}
}

func TestArtifactTagExtract_TC10_NoTagInFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "package foo\n\nfunc main() {}\n"
	if err := os.WriteFile("file.go", []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrNoTagFound) {
		t.Errorf("expected ErrNoTagFound, got %v", err)
	}
}

func TestArtifactTagExtract_TC11_MalformedTag_NoAtSeparator(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "// code-from-spec: ROOT/some/nodeabcdefghijklmnopqrstuvwxyza\n"
	if err := os.WriteFile("file.go", []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("expected ErrMalformedTag, got %v", err)
	}
}

func TestArtifactTagExtract_TC12_MalformedTag_EmptyLogicalName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "// code-from-spec: @abcdefghijklmnopqrstuvwxyza\n"
	if err := os.WriteFile("file.go", []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("expected ErrMalformedTag, got %v", err)
	}
}

func TestArtifactTagExtract_TC13_MalformedTag_WrongHashLength(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "// code-from-spec: ROOT/some/node@tooshort\n"
	if err := os.WriteFile("file.go", []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("expected ErrMalformedTag, got %v", err)
	}
}
