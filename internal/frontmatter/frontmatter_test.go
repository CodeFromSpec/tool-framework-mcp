// code-from-spec: ROOT/golang/tests/parsing/frontmatter@Vmwqa_Tg8KdEnaFZBSCELwolu9A
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
  - dep/one
  - dep/two
external:
  - path: ext/path/one.md
  - path: ext/path/two.md
input: some-input-value
output: some/output/path.go
---
Body content here.
`
	if err := os.WriteFile("node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 2 || fm.DependsOn[0] != "dep/one" || fm.DependsOn[1] != "dep/two" {
		t.Errorf("DependsOn = %v, want [dep/one dep/two]", fm.DependsOn)
	}
	if len(fm.External) != 2 || fm.External[0].Path != "ext/path/one.md" || fm.External[1].Path != "ext/path/two.md" {
		t.Errorf("External = %v", fm.External)
	}
	if fm.Input != "some-input-value" {
		t.Errorf("Input = %q, want %q", fm.Input, "some-input-value")
	}
	if fm.Output != "some/output/path.go" {
		t.Errorf("Output = %q, want %q", fm.Output, "some/output/path.go")
	}
}

func TestFrontmatterParse_OnlyOutput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
output: only/output.go
---
`
	if err := os.WriteFile("node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

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
	if fm.Output != "only/output.go" {
		t.Errorf("Output = %q, want %q", fm.Output, "only/output.go")
	}
}

func TestFrontmatterParse_OnlyDependsOn(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
depends_on:
  - first/dep
  - second/dep
---
`
	if err := os.WriteFile("node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 2 || fm.DependsOn[0] != "first/dep" || fm.DependsOn[1] != "second/dep" {
		t.Errorf("DependsOn = %v, want [first/dep second/dep]", fm.DependsOn)
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

	content := `---
external:
  - path: a/b/c.md
  - path: d/e/f.md
---
`
	if err := os.WriteFile("node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.External) != 2 || fm.External[0].Path != "a/b/c.md" || fm.External[1].Path != "d/e/f.md" {
		t.Errorf("External = %v", fm.External)
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

	content := `---
input: my-input-value
---
`
	if err := os.WriteFile("node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Input != "my-input-value" {
		t.Errorf("Input = %q, want %q", fm.Input, "my-input-value")
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

	content := `---
output: known/output.go
depends_on:
  - known/dep
custom_field: some value
---
`
	if err := os.WriteFile("node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Output != "known/output.go" {
		t.Errorf("Output = %q, want %q", fm.Output, "known/output.go")
	}
	if len(fm.DependsOn) != 1 || fm.DependsOn[0] != "known/dep" {
		t.Errorf("DependsOn = %v, want [known/dep]", fm.DependsOn)
	}
}

func TestFrontmatterParse_NoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "Just body content without any frontmatter delimiters.\n"
	if err := os.WriteFile("node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

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

	content := "---\n---\n"
	if err := os.WriteFile("node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

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

	content := "---\noutput: no/body/output.go\n---\n"
	if err := os.WriteFile("node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Output != "no/body/output.go" {
		t.Errorf("Output = %q, want %q", fm.Output, "no/body/output.go")
	}
}

func TestFrontmatterParse_DelimiterWithTrailingWhitespaceNotRecognized(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "---   \noutput: some/path.go\n---\n"
	if err := os.WriteFile("node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

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

func TestFrontmatterParse_FileNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "nonexistent/file.txt"})
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
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("error = %v, want ErrDirectoryTraversal", err)
	}
}

func TestFrontmatterParse_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := "---\n: invalid: yaml: content\n  bad indentation:\n---\n"
	if err := os.WriteFile("node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

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

	content := "---\noutput: some/path.go\nno closing delimiter follows\n"
	if err := os.WriteFile("node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("error = %v, want ErrMalformedYAML", err)
	}
}

func TestFrontmatterParse_MissingPathInExternalEntry(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
external:
  - unknown_key: some value
---
`
	if err := os.WriteFile("node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("error = %v, want ErrMalformedYAML", err)
	}
}

var _ = filereader.ErrFileUnreadable
