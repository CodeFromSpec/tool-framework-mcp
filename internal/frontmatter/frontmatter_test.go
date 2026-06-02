// code-from-spec: ROOT/golang/tests/parsing/frontmatter@NBKpsI01r2EKCFL0SI2OzsNgJ5U
package frontmatter_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
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
	if err := os.WriteFile(name, []byte(content), 0600); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

func TestFrontmatterParse_CompleteFields(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
depends_on:
  - ROOT/a
  - ROOT/b
external:
  - path: some/external/file.md
  - path: another/external/file.md
input: path/to/input.md
output: path/to/output.md
---
Body content here.
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 2 || fm.DependsOn[0] != "ROOT/a" || fm.DependsOn[1] != "ROOT/b" {
		t.Errorf("DependsOn = %v, want [ROOT/a ROOT/b]", fm.DependsOn)
	}
	if len(fm.External) != 2 {
		t.Fatalf("len(External) = %d, want 2", len(fm.External))
	}
	if fm.External[0].Path != "some/external/file.md" {
		t.Errorf("External[0].Path = %q, want %q", fm.External[0].Path, "some/external/file.md")
	}
	if fm.External[1].Path != "another/external/file.md" {
		t.Errorf("External[1].Path = %q, want %q", fm.External[1].Path, "another/external/file.md")
	}
	if fm.Input != "path/to/input.md" {
		t.Errorf("Input = %q, want %q", fm.Input, "path/to/input.md")
	}
	if fm.Output != "path/to/output.md" {
		t.Errorf("Output = %q, want %q", fm.Output, "path/to/output.md")
	}
}

func TestFrontmatterParse_OnlyOutput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
output: path/to/output.md
---
Body content here.
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn = %v, want empty", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("External = %v, want empty", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("Input = %q, want empty", fm.Input)
	}
	if fm.Output != "path/to/output.md" {
		t.Errorf("Output = %q, want %q", fm.Output, "path/to/output.md")
	}
}

func TestFrontmatterParse_OnlyDependsOn(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
depends_on:
  - ROOT/x
  - ROOT/y
---
Body content here.
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 2 || fm.DependsOn[0] != "ROOT/x" || fm.DependsOn[1] != "ROOT/y" {
		t.Errorf("DependsOn = %v, want [ROOT/x ROOT/y]", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("External = %v, want empty", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("Input = %q, want empty", fm.Input)
	}
	if fm.Output != "" {
		t.Errorf("Output = %q, want empty", fm.Output)
	}
}

func TestFrontmatterParse_OnlyExternal(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
external:
  - path: docs/reference.md
  - path: docs/guide.md
---
Body content here.
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.External) != 2 {
		t.Fatalf("len(External) = %d, want 2", len(fm.External))
	}
	if fm.External[0].Path != "docs/reference.md" {
		t.Errorf("External[0].Path = %q, want %q", fm.External[0].Path, "docs/reference.md")
	}
	if fm.External[1].Path != "docs/guide.md" {
		t.Errorf("External[1].Path = %q, want %q", fm.External[1].Path, "docs/guide.md")
	}
	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn = %v, want empty", fm.DependsOn)
	}
	if fm.Input != "" {
		t.Errorf("Input = %q, want empty", fm.Input)
	}
	if fm.Output != "" {
		t.Errorf("Output = %q, want empty", fm.Output)
	}
}

func TestFrontmatterParse_OnlyInput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
input: path/to/input.md
---
Body content here.
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Input != "path/to/input.md" {
		t.Errorf("Input = %q, want %q", fm.Input, "path/to/input.md")
	}
	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn = %v, want empty", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("External = %v, want empty", fm.External)
	}
	if fm.Output != "" {
		t.Errorf("Output = %q, want empty", fm.Output)
	}
}

func TestFrontmatterParse_IgnoresUnknownFields(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
output: path/to/output.md
custom_field: some value
another_unknown: 42
---
Body content here.
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Output != "path/to/output.md" {
		t.Errorf("Output = %q, want %q", fm.Output, "path/to/output.md")
	}
	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn = %v, want empty", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("External = %v, want empty", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("Input = %q, want empty", fm.Input)
	}
}

func TestFrontmatterParse_NoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `This is just body content.
No frontmatter here.
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn = %v, want empty", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("External = %v, want empty", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("Input = %q, want empty", fm.Input)
	}
	if fm.Output != "" {
		t.Errorf("Output = %q, want empty", fm.Output)
	}
}

func TestFrontmatterParse_EmptyFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
---
Body content here.
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn = %v, want empty", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("External = %v, want empty", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("Input = %q, want empty", fm.Input)
	}
	if fm.Output != "" {
		t.Errorf("Output = %q, want empty", fm.Output)
	}
}

func TestFrontmatterParse_OnlyFrontmatterNoBody(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
output: path/to/output.md
---
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Output != "path/to/output.md" {
		t.Errorf("Output = %q, want %q", fm.Output, "path/to/output.md")
	}
}

func TestFrontmatterParse_DelimiterWithTrailingWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", "---   \noutput: path/to/output.md\n---\n")

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn = %v, want empty", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("External = %v, want empty", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("Input = %q, want empty", fm.Input)
	}
	if fm.Output != "" {
		t.Errorf("Output = %q, want empty string (delimiter with trailing whitespace not recognized)", fm.Output)
	}
}

func TestFrontmatterParse_FileNotExist(t *testing.T) {
	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "nonexistent/file.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrFileUnreadable) {
		t.Errorf("error = %v, want ErrFileUnreadable", err)
	}
}

func TestFrontmatterParse_PropagatesPathErrors(t *testing.T) {
	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "../../outside"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) && !errors.Is(err, filereader.ErrFileUnreadable) && !errors.Is(err, frontmatter.ErrFileUnreadable) {
		t.Errorf("error = %v, want directory traversal or file unreadable error", err)
	}
}

func TestFrontmatterParse_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
key: [unclosed bracket
another: : invalid
---
Body content.
`)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("error = %v, want ErrMalformedYAML", err)
	}
}

func TestFrontmatterParse_UnclosedFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
output: path/to/output.md
Body content with no closing delimiter.
`)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("error = %v, want ErrMalformedYAML", err)
	}
}

func TestFrontmatterParse_ExternalMissingPath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
external:
  - name: some-name
---
Body content.
`)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("error = %v, want ErrMalformedYAML", err)
	}
}
