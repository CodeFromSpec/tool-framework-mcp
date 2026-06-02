// code-from-spec: ROOT/golang/tests/parsing/frontmatter@-KTUTITsRW8bJC6iQZmQyAqdi94
package frontmatter_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
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

func testWriteFile(t *testing.T, name string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(name, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile write: %v", err)
	}
}

// --- Happy Path ---

func TestFrontmatterParse_AllFields(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "---\ndepends_on:\n  - ROOT/a\n  - ROOT/b\nexternal:\n  - path: some/external/file.md\n  - path: another/external/file.md\ninput: path/to/input.md\noutput: path/to/output.md\n---\nBody content here.\n"
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm.DependsOn) != 2 || fm.DependsOn[0] != "ROOT/a" || fm.DependsOn[1] != "ROOT/b" {
		t.Errorf("unexpected depends_on: %v", fm.DependsOn)
	}
	if len(fm.External) != 2 {
		t.Fatalf("expected 2 external entries, got %d", len(fm.External))
	}
	if fm.External[0].Path != "some/external/file.md" {
		t.Errorf("expected first external some/external/file.md, got %q", fm.External[0].Path)
	}
	if fm.External[1].Path != "another/external/file.md" {
		t.Errorf("expected second external another/external/file.md, got %q", fm.External[1].Path)
	}
	if fm.Input != "path/to/input.md" {
		t.Errorf("expected input path/to/input.md, got %q", fm.Input)
	}
	if fm.Output != "path/to/output.md" {
		t.Errorf("expected output path/to/output.md, got %q", fm.Output)
	}
}

func TestFrontmatterParse_OnlyOutput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "---\noutput: path/to/output.md\n---\nBody content here.\n"
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm.DependsOn) != 0 {
		t.Errorf("expected empty depends_on, got %v", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("expected empty external, got %v", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("expected empty input, got %q", fm.Input)
	}
	if fm.Output != "path/to/output.md" {
		t.Errorf("expected output path/to/output.md, got %q", fm.Output)
	}
}

func TestFrontmatterParse_OnlyDependsOn(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "---\ndepends_on:\n  - ROOT/x\n  - ROOT/y\n---\nBody content here.\n"
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm.DependsOn) != 2 || fm.DependsOn[0] != "ROOT/x" || fm.DependsOn[1] != "ROOT/y" {
		t.Errorf("unexpected depends_on: %v", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("expected empty external, got %v", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("expected empty input, got %q", fm.Input)
	}
	if fm.Output != "" {
		t.Errorf("expected empty output, got %q", fm.Output)
	}
}

func TestFrontmatterParse_ExternalEntries(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "---\nexternal:\n  - path: docs/reference.md\n  - path: docs/guide.md\n---\nBody content here.\n"
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm.External) != 2 {
		t.Fatalf("expected 2 external entries, got %d", len(fm.External))
	}
	if fm.External[0].Path != "docs/reference.md" {
		t.Errorf("expected first external docs/reference.md, got %q", fm.External[0].Path)
	}
	if fm.External[1].Path != "docs/guide.md" {
		t.Errorf("expected second external docs/guide.md, got %q", fm.External[1].Path)
	}
	if len(fm.DependsOn) != 0 || fm.Input != "" || fm.Output != "" {
		t.Errorf("expected all other fields empty")
	}
}

func TestFrontmatterParse_OnlyInput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "---\ninput: path/to/input.md\n---\nBody content here.\n"
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Input != "path/to/input.md" {
		t.Errorf("expected input path/to/input.md, got %q", fm.Input)
	}
	if len(fm.DependsOn) != 0 || len(fm.External) != 0 || fm.Output != "" {
		t.Errorf("expected all other fields empty")
	}
}

func TestFrontmatterParse_IgnoresUnknownFields(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "---\noutput: path/to/output.md\ncustom_field: some value\nanother_unknown: 42\n---\nBody content here.\n"
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Output != "path/to/output.md" {
		t.Errorf("expected output path/to/output.md, got %q", fm.Output)
	}
	if len(fm.DependsOn) != 0 || len(fm.External) != 0 || fm.Input != "" {
		t.Errorf("expected all other fields empty")
	}
}

func TestFrontmatterParse_NoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "This is just body content.\nNo frontmatter here.\n"
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm.DependsOn) != 0 || len(fm.External) != 0 || fm.Input != "" || fm.Output != "" {
		t.Errorf("expected all fields empty, got %+v", fm)
	}
}

// --- Edge Cases ---

func TestFrontmatterParse_EmptyFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "---\n---\nBody content here.\n"
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm.DependsOn) != 0 || len(fm.External) != 0 || fm.Input != "" || fm.Output != "" {
		t.Errorf("expected all fields empty, got %+v", fm)
	}
}

func TestFrontmatterParse_OnlyFrontmatterNoBody(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "---\noutput: path/to/output.md\n---\n"
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Output != "path/to/output.md" {
		t.Errorf("expected output path/to/output.md, got %q", fm.Output)
	}
	if len(fm.DependsOn) != 0 || len(fm.External) != 0 || fm.Input != "" {
		t.Errorf("expected all other fields empty")
	}
}

func TestFrontmatterParse_DelimiterWithTrailingWhitespaceNotRecognized(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "---   \noutput: path/to/output.md\n---\n"
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm.DependsOn) != 0 || len(fm.External) != 0 || fm.Input != "" || fm.Output != "" {
		t.Errorf("expected all fields empty (trailing whitespace delimiter not recognized), got %+v", fm)
	}
}

// --- Failure Cases ---

func TestFrontmatterParse_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "nonexistent/file.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestFrontmatterParse_PropagatesPathErrors(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "../../outside"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got %v", err)
	}
}

func TestFrontmatterParse_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "---\nkey: [unclosed bracket\nanother: : invalid\n---\nBody content.\n"
	testWriteFile(t, "node.md", content)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got %v", err)
	}
}

func TestFrontmatterParse_UnclosedFrontmatterBlock(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "---\noutput: path/to/output.md\nBody content with no closing delimiter.\n"
	testWriteFile(t, "node.md", content)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got %v", err)
	}
}

func TestFrontmatterParse_MissingPathInExternalEntry(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "---\nexternal:\n  - name: some-name\n---\nBody content.\n"
	testWriteFile(t, "node.md", content)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got %v", err)
	}
}
