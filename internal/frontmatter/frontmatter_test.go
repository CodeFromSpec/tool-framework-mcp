// code-from-spec: ROOT/golang/tests/parsing/frontmatter@2BdpE1kZBsS8nYjMJRknO-HYoY0
package frontmatter_test

import (
	"errors"
	"os"
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

func TestFrontmatterParse_AllFields(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
depends_on:
  - dep/one
  - dep/two
external:
  - path: ext/first.md
  - path: ext/second.md
input: some-input-value
output: some/output/path.go
---
# Body content
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
	if len(fm.External) != 2 || fm.External[0].Path != "ext/first.md" || fm.External[1].Path != "ext/second.md" {
		t.Errorf("External = %v, want [{ext/first.md} {ext/second.md}]", fm.External)
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
output: internal/foo/bar.go
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
	if fm.Output != "internal/foo/bar.go" {
		t.Errorf("Output = %q, want %q", fm.Output, "internal/foo/bar.go")
	}
}

func TestFrontmatterParse_OnlyDependsOn(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
depends_on:
  - alpha/beta
  - gamma/delta
---
`
	if err := os.WriteFile("node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm.DependsOn) != 2 || fm.DependsOn[0] != "alpha/beta" || fm.DependsOn[1] != "gamma/delta" {
		t.Errorf("DependsOn = %v, want [alpha/beta gamma/delta]", fm.DependsOn)
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
  - path: docs/one.md
  - path: docs/two.md
---
`
	if err := os.WriteFile("node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm.External) != 2 || fm.External[0].Path != "docs/one.md" || fm.External[1].Path != "docs/two.md" {
		t.Errorf("External = %v, want [{docs/one.md} {docs/two.md}]", fm.External)
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
input: my-input-id
---
`
	if err := os.WriteFile("node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Input != "my-input-id" {
		t.Errorf("Input = %q, want %q", fm.Input, "my-input-id")
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
output: internal/pkg/file.go
depends_on:
  - some/dep
custom_field: value
---
`
	if err := os.WriteFile("node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Output != "internal/pkg/file.go" {
		t.Errorf("Output = %q, want %q", fm.Output, "internal/pkg/file.go")
	}
	if len(fm.DependsOn) != 1 || fm.DependsOn[0] != "some/dep" {
		t.Errorf("DependsOn = %v, want [some/dep]", fm.DependsOn)
	}
}

func TestFrontmatterParse_NoFrontmatterReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `# Just a heading

Some body content without frontmatter.
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
	if fm.Output != "" {
		t.Errorf("Output = %q, want empty", fm.Output)
	}
}

func TestFrontmatterParse_EmptyFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
---
# body
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
	if fm.Output != "" {
		t.Errorf("Output = %q, want empty", fm.Output)
	}
}

func TestFrontmatterParse_OnlyFrontmatterNoBody(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
output: internal/result.go
---`
	if err := os.WriteFile("node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Output != "internal/result.go" {
		t.Errorf("Output = %q, want %q", fm.Output, "internal/result.go")
	}
}

func TestFrontmatterParse_DelimiterWithTrailingWhitespaceNotRecognized(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---   
output: internal/result.go
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
	if fm.Output != "" {
		t.Errorf("Output = %q, want empty", fm.Output)
	}
}

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

	content := `---
: invalid: yaml: content: [
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
		t.Errorf("expected ErrMalformedYAML, got %v", err)
	}
}

func TestFrontmatterParse_UnclosedFrontmatterBlock(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
output: internal/result.go
just body content without closing delimiter
`
	if err := os.WriteFile("node.md", []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

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

	content := `---
external:
  - unknown_key: value
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
		t.Errorf("expected ErrMalformedYAML, got %v", err)
	}
}
