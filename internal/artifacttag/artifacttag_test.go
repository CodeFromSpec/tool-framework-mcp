// code-from-spec: SPEC/golang/tests/parsing/artifact_tag@qHdX3qEnN57hIXa7TgKeWuq14C4
package artifacttag_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/artifacttag"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
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

func TestArtifactTagExtract_TC01_SlashSlashComment(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

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

func TestArtifactTagExtract_TC02_HashComment(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

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

func TestArtifactTagExtract_TC03_HTMLComment(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "<!-- code-from-spec: ROOT/docs/readme@abcdefghijklmnopqrstuvwxyza -->\n"
	if err := os.WriteFile("file.html", []byte(content), 0644); err != nil {
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

func TestArtifactTagExtract_TC04_StopsAtFirstMatch(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

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

func TestArtifactTagExtract_TC05_TagOnNonFirstLine(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "package foo\n" +
		"// no tag here\n" +
		"// code-from-spec: ROOT/some/node@abcdefghijklmnopqrstuvwxyza\n"
	if err := os.WriteFile("file.go", []byte(content), 0644); err != nil {
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

func TestArtifactTagExtract_TC06_ExtraWhitespaceBeforeLogicalName(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

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

func TestArtifactTagExtract_TC07_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	if err := os.WriteFile("empty.go", []byte{}, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "empty.go"})
	if !errors.Is(err, artifacttag.ErrNoTagFound) {
		t.Errorf("err = %v, want ErrNoTagFound", err)
	}
}

func TestArtifactTagExtract_TC08_FileDoesNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "nonexistent/file.go"})
	if !errors.Is(err, artifacttag.ErrFileUnreadable) {
		t.Errorf("err = %v, want ErrFileUnreadable", err)
	}
}

func TestArtifactTagExtract_TC09_PropagatesPathErrors(t *testing.T) {
	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "../../outside"})
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("err = %v, want ErrDirectoryTraversal", err)
	}
}

func TestArtifactTagExtract_TC10_NoTagInFile(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "package foo\n\nfunc Hello() string {\n\treturn \"hello\"\n}\n"
	if err := os.WriteFile("file.go", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrNoTagFound) {
		t.Errorf("err = %v, want ErrNoTagFound", err)
	}
}

func TestArtifactTagExtract_TC11_MalformedTag_NoAtSeparator(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "// code-from-spec: ROOT/foo/bar\n"
	if err := os.WriteFile("file.go", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("err = %v, want ErrMalformedTag", err)
	}
}

func TestArtifactTagExtract_TC12_MalformedTag_EmptyLogicalName(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "// code-from-spec: @abcdefghijklmnopqrstuvwxyza\n"
	if err := os.WriteFile("file.go", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("err = %v, want ErrMalformedTag", err)
	}
}

func TestArtifactTagExtract_TC13_MalformedTag_WrongHashLength(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "// code-from-spec: ROOT/foo(bar)@short\n"
	if err := os.WriteFile("file.go", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: "file.go"})
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("err = %v, want ErrMalformedTag", err)
	}
}
