// code-from-spec: ROOT/golang/tests/parsing/artifact_tag@h6app5-C1QsB92x0JFBVx4zmstI
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
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "// code-from-spec: ROOT/golang/implementation/internal/foo/code(bar)@abcdefghijklmnopqrstuvwxyza\n"
	if err := os.WriteFile("file.go", []byte(content), 0600); err != nil {
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

func TestArtifactTagExtract_HashComment(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "# code-from-spec: ROOT/some/node(id)@123456789012345678901234567\n"
	if err := os.WriteFile("file.sh", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	tag, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.sh"})
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
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "<!-- code-from-spec: ROOT/docs/readme@abcdefghijklmnopqrstuvwxyza -->\n"
	if err := os.WriteFile("file.md", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

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
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "// code-from-spec: ROOT/first/node@abcdefghijklmnopqrstuvwxyza\n// code-from-spec: ROOT/second/node@zyxwvutsrqponmlkjihgfedcbaa\n"
	if err := os.WriteFile("file.go", []byte(content), 0600); err != nil {
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

func TestArtifactTagExtract_TagOnNonFirstLine(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "line one\nline two\n// code-from-spec: ROOT/some/node@abcdefghijklmnopqrstuvwxyza\n"
	if err := os.WriteFile("file.go", []byte(content), 0600); err != nil {
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

func TestArtifactTagExtract_ExtraWhitespaceBeforeLogicalName(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "// code-from-spec:   ROOT/x(y)@abcdefghijklmnopqrstuvwxyza\n"
	if err := os.WriteFile("file.go", []byte(content), 0600); err != nil {
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

func TestArtifactTagExtract_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	if err := os.WriteFile("empty.go", []byte{}, 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "empty.go"})
	if !errors.Is(err, artifacttag.ErrNoTagFound) {
		t.Errorf("expected ErrNoTagFound, got %v", err)
	}
}

func TestArtifactTagExtract_FileDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "nonexistent/file.go"})
	if !errors.Is(err, artifacttag.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestArtifactTagExtract_PropagatesPathErrors(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "../../outside"})
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got %v", err)
	}
}

func TestArtifactTagExtract_NoTagInFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "this file has no artifact tag\njust plain text\n"
	if err := os.WriteFile("file.go", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrNoTagFound) {
		t.Errorf("expected ErrNoTagFound, got %v", err)
	}
}

func TestArtifactTagExtract_MalformedTag_NoAtSeparator(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "// code-from-spec: ROOT/foo/bar\n"
	if err := os.WriteFile("file.go", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("expected ErrMalformedTag, got %v", err)
	}
}

func TestArtifactTagExtract_MalformedTag_EmptyLogicalName(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "// code-from-spec: @abcdefghijklmnopqrstuvwxyza\n"
	if err := os.WriteFile("file.go", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("expected ErrMalformedTag, got %v", err)
	}
}

func TestArtifactTagExtract_MalformedTag_WrongHashLength(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := "// code-from-spec: ROOT/foo(bar)@short\n"
	if err := os.WriteFile("file.go", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("expected ErrMalformedTag, got %v", err)
	}
}
