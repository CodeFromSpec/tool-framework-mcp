// code-from-spec: ROOT/golang/tests/parsing/frontmatter@dnYATPGwktyUnU4ofuxFOXbGwzA
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

func TestFrontmatterParse_AllFields(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
depends_on:
  - ROOT/dep/a
  - ROOT/dep/b
external:
  - path: some/external/file.md
  - path: another/external/file.md
input: some-input-value
output: some/output/path.go
---
# Body content
`
	if err := os.WriteFile("node.md", []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 2 {
		t.Errorf("DependsOn length: got %d, want 2", len(fm.DependsOn))
	} else {
		if fm.DependsOn[0] != "ROOT/dep/a" {
			t.Errorf("DependsOn[0]: got %q, want %q", fm.DependsOn[0], "ROOT/dep/a")
		}
		if fm.DependsOn[1] != "ROOT/dep/b" {
			t.Errorf("DependsOn[1]: got %q, want %q", fm.DependsOn[1], "ROOT/dep/b")
		}
	}

	if len(fm.External) != 2 {
		t.Errorf("External length: got %d, want 2", len(fm.External))
	} else {
		if fm.External[0].Path != "some/external/file.md" {
			t.Errorf("External[0].Path: got %q, want %q", fm.External[0].Path, "some/external/file.md")
		}
		if fm.External[1].Path != "another/external/file.md" {
			t.Errorf("External[1].Path: got %q, want %q", fm.External[1].Path, "another/external/file.md")
		}
	}

	if fm.Input != "some-input-value" {
		t.Errorf("Input: got %q, want %q", fm.Input, "some-input-value")
	}

	if fm.Output != "some/output/path.go" {
		t.Errorf("Output: got %q, want %q", fm.Output, "some/output/path.go")
	}
}

func TestFrontmatterParse_OnlyOutput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
output: internal/pkg/file.go
---
# Body
`
	if err := os.WriteFile("node.md", []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: got %v, want empty", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("External: got %v, want empty", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if fm.Output != "internal/pkg/file.go" {
		t.Errorf("Output: got %q, want %q", fm.Output, "internal/pkg/file.go")
	}
}

func TestFrontmatterParse_OnlyDependsOn(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
depends_on:
  - ROOT/a
  - ROOT/b
---
`
	if err := os.WriteFile("node.md", []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 2 {
		t.Errorf("DependsOn length: got %d, want 2", len(fm.DependsOn))
	} else {
		if fm.DependsOn[0] != "ROOT/a" {
			t.Errorf("DependsOn[0]: got %q, want %q", fm.DependsOn[0], "ROOT/a")
		}
		if fm.DependsOn[1] != "ROOT/b" {
			t.Errorf("DependsOn[1]: got %q, want %q", fm.DependsOn[1], "ROOT/b")
		}
	}
	if len(fm.External) != 0 {
		t.Errorf("External: got %v, want empty", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if fm.Output != "" {
		t.Errorf("Output: got %q, want empty", fm.Output)
	}
}

func TestFrontmatterParse_OnlyExternal(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
external:
  - path: first/path.md
  - path: second/path.md
---
`
	if err := os.WriteFile("node.md", []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.External) != 2 {
		t.Errorf("External length: got %d, want 2", len(fm.External))
	} else {
		if fm.External[0].Path != "first/path.md" {
			t.Errorf("External[0].Path: got %q, want %q", fm.External[0].Path, "first/path.md")
		}
		if fm.External[1].Path != "second/path.md" {
			t.Errorf("External[1].Path: got %q, want %q", fm.External[1].Path, "second/path.md")
		}
	}
	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: got %v, want empty", fm.DependsOn)
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if fm.Output != "" {
		t.Errorf("Output: got %q, want empty", fm.Output)
	}
}

func TestFrontmatterParse_OnlyInput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
input: my-input-artifact
---
`
	if err := os.WriteFile("node.md", []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Input != "my-input-artifact" {
		t.Errorf("Input: got %q, want %q", fm.Input, "my-input-artifact")
	}
	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: got %v, want empty", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("External: got %v, want empty", fm.External)
	}
	if fm.Output != "" {
		t.Errorf("Output: got %q, want empty", fm.Output)
	}
}

func TestFrontmatterParse_IgnoresUnknownFields(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
output: internal/out.go
depends_on:
  - ROOT/dep
custom_field: some-value
---
`
	if err := os.WriteFile("node.md", []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Output != "internal/out.go" {
		t.Errorf("Output: got %q, want %q", fm.Output, "internal/out.go")
	}
	if len(fm.DependsOn) != 1 || fm.DependsOn[0] != "ROOT/dep" {
		t.Errorf("DependsOn: got %v, want [ROOT/dep]", fm.DependsOn)
	}
}

func TestFrontmatterParse_NoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# Just a heading

Some body content without any frontmatter.
`
	if err := os.WriteFile("node.md", []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: got %v, want empty", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("External: got %v, want empty", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if fm.Output != "" {
		t.Errorf("Output: got %q, want empty", fm.Output)
	}
}

func TestFrontmatterParse_EmptyFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
---
# Body
`
	if err := os.WriteFile("node.md", []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: got %v, want empty", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("External: got %v, want empty", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if fm.Output != "" {
		t.Errorf("Output: got %q, want empty", fm.Output)
	}
}

func TestFrontmatterParse_OnlyFrontmatterNoBody(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
output: internal/result.go
---`
	if err := os.WriteFile("node.md", []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Output != "internal/result.go" {
		t.Errorf("Output: got %q, want %q", fm.Output, "internal/result.go")
	}
}

func TestFrontmatterParse_DelimiterWithTrailingWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "---   \noutput: internal/out.go\n---\n"
	if err := os.WriteFile("node.md", []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: got %v, want empty", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("External: got %v, want empty", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if fm.Output != "" {
		t.Errorf("Output: got %q, want empty", fm.Output)
	}
}

func TestFrontmatterParse_FileNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "nonexistent/file.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestFrontmatterParse_PropagatesPathErrors(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "../../outside"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) && !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrDirectoryTraversal or ErrFileUnreadable propagated, got: %v", err)
	}
}

func TestFrontmatterParse_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
: invalid: yaml: content
  bad indentation:
    - [unclosed
---
`
	if err := os.WriteFile("node.md", []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got: %v", err)
	}
}

func TestFrontmatterParse_UnclosedFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
output: internal/out.go
depends_on:
  - ROOT/dep
# No closing ---
Some body content here.
`
	if err := os.WriteFile("node.md", []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got: %v", err)
	}
}

func TestFrontmatterParse_ExternalMissingPath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
external:
  - unknown_key: some-value
---
`
	if err := os.WriteFile("node.md", []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got: %v", err)
	}
}
