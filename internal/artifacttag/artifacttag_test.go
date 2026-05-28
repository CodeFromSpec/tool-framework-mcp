// code-from-spec: ROOT/golang/tests/internal/artifact_tag@LfL3cGomWwgouUybxo4rVHvc1pk

package artifacttag_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/artifacttag"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

// testWriteFile writes content to a file inside dir with the given name.
// It returns a *pathutils.PathCfs relative to the project root.
func testWriteFile(t *testing.T, dir, name, content string) *pathutils.PathCfs {
	t.Helper()
	fullPath := filepath.Join(dir, name)
	if err := os.WriteFile(fullPath, []byte(content), 0600); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
	root, err := pathutils.PathGetProjectRoot()
	if err != nil {
		t.Fatalf("testWriteFile: PathGetProjectRoot: %v", err)
	}
	rel, err := filepath.Rel(root.Value, fullPath)
	if err != nil {
		t.Fatalf("testWriteFile: filepath.Rel: %v", err)
	}
	// Convert OS separator to forward slashes.
	cfsValue := filepath.ToSlash(rel)
	return &pathutils.PathCfs{Value: cfsValue}
}

// testTempSubdir creates a temp directory that is a subdirectory of the
// project root so that PathCfsToOs can resolve it without
// ErrResolvesOutsideRoot. It returns the absolute path of the created dir.
func testTempSubdir(t *testing.T) string {
	t.Helper()
	root, err := pathutils.PathGetProjectRoot()
	if err != nil {
		t.Fatalf("testTempSubdir: PathGetProjectRoot: %v", err)
	}
	dir, err := os.MkdirTemp(root.Value, "artifacttag_test_*")
	if err != nil {
		t.Fatalf("testTempSubdir: MkdirTemp: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

// ---------------------------------------------------------------------------
// Happy path tests
// ---------------------------------------------------------------------------

func TestArtifactTagExtract_SlashSlashComment(t *testing.T) {
	dir := testTempSubdir(t)
	content := "// code-from-spec: ROOT/golang/implementation/internal/foo/code(bar)@abcdefghijklmnopqrstuvwxyza\n"
	path := testWriteFile(t, dir, "file.go", content)

	tag, err := artifacttag.ArtifactTagExtract(path)
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
	dir := testTempSubdir(t)
	content := "# code-from-spec: ROOT/some/node(id)@123456789012345678901234567\n"
	path := testWriteFile(t, dir, "file.py", content)

	tag, err := artifacttag.ArtifactTagExtract(path)
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
	dir := testTempSubdir(t)
	content := "<!-- code-from-spec: ROOT/docs/readme@abcdefghijklmnopqrstuvwxyza -->\n"
	path := testWriteFile(t, dir, "readme.md", content)

	tag, err := artifacttag.ArtifactTagExtract(path)
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
	dir := testTempSubdir(t)
	content := "// code-from-spec: ROOT/first/node@abcdefghijklmnopqrstuvwxyza\n" +
		"// code-from-spec: ROOT/second/node@zyxwvutsrqponmlkjihgfedcbaa\n"
	path := testWriteFile(t, dir, "file.go", content)

	tag, err := artifacttag.ArtifactTagExtract(path)
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
	dir := testTempSubdir(t)
	content := "line one content\nline two content\n// code-from-spec: ROOT/some/node@abcdefghijklmnopqrstuvwxyza\n"
	path := testWriteFile(t, dir, "file.go", content)

	tag, err := artifacttag.ArtifactTagExtract(path)
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
	dir := testTempSubdir(t)
	content := "// code-from-spec:   ROOT/x(y)@abcdefghijklmnopqrstuvwxyza\n"
	path := testWriteFile(t, dir, "file.go", content)

	tag, err := artifacttag.ArtifactTagExtract(path)
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

// ---------------------------------------------------------------------------
// Edge case tests
// ---------------------------------------------------------------------------

func TestArtifactTagExtract_EmptyFile(t *testing.T) {
	dir := testTempSubdir(t)
	path := testWriteFile(t, dir, "empty.go", "")

	_, err := artifacttag.ArtifactTagExtract(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, artifacttag.ErrNoTagFound) {
		t.Errorf("error = %v, want wrapping ErrNoTagFound", err)
	}
}

// ---------------------------------------------------------------------------
// Failure case tests
// ---------------------------------------------------------------------------

func TestArtifactTagExtract_FileDoesNotExist(t *testing.T) {
	dir := testTempSubdir(t)
	root, err := pathutils.PathGetProjectRoot()
	if err != nil {
		t.Fatalf("PathGetProjectRoot: %v", err)
	}
	rel, err := filepath.Rel(root.Value, filepath.Join(dir, "nonexistent.go"))
	if err != nil {
		t.Fatalf("filepath.Rel: %v", err)
	}
	path := &pathutils.PathCfs{Value: filepath.ToSlash(rel)}

	_, err = artifacttag.ArtifactTagExtract(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("error = %v, want wrapping ErrFileUnreadable", err)
	}
}

func TestArtifactTagExtract_PropagatesPathErrors(t *testing.T) {
	path := &pathutils.PathCfs{Value: "../../outside"}

	_, err := artifacttag.ArtifactTagExtract(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("error = %v, want wrapping ErrDirectoryTraversal", err)
	}
}

func TestArtifactTagExtract_NoTagInFile(t *testing.T) {
	dir := testTempSubdir(t)
	content := "This file has no artifact tag at all.\nJust some regular text.\n"
	path := testWriteFile(t, dir, "file.go", content)

	_, err := artifacttag.ArtifactTagExtract(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, artifacttag.ErrNoTagFound) {
		t.Errorf("error = %v, want wrapping ErrNoTagFound", err)
	}
}

func TestArtifactTagExtract_MalformedTagNoAtSeparator(t *testing.T) {
	dir := testTempSubdir(t)
	content := "// code-from-spec: ROOT/foo/bar\n"
	path := testWriteFile(t, dir, "file.go", content)

	_, err := artifacttag.ArtifactTagExtract(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("error = %v, want wrapping ErrMalformedTag", err)
	}
}

func TestArtifactTagExtract_MalformedTagEmptyLogicalName(t *testing.T) {
	dir := testTempSubdir(t)
	content := "// code-from-spec: @abcdefghijklmnopqrstuvwxyza\n"
	path := testWriteFile(t, dir, "file.go", content)

	_, err := artifacttag.ArtifactTagExtract(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("error = %v, want wrapping ErrMalformedTag", err)
	}
}

func TestArtifactTagExtract_MalformedTagWrongHashLength(t *testing.T) {
	dir := testTempSubdir(t)
	content := "// code-from-spec: ROOT/foo(bar)@short\n"
	path := testWriteFile(t, dir, "file.go", content)

	_, err := artifacttag.ArtifactTagExtract(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("error = %v, want wrapping ErrMalformedTag", err)
	}
}
