// code-from-spec: ROOT/golang/tests/parsing/frontmatter@QpR4q175Bjb5u5eNXRfURXJnDzY
package frontmatter_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// testChdir changes the working directory to dir for the duration of the test.
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

// testWriteFile creates any necessary directories and writes content to path.
func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("testWriteFile MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile WriteFile: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Happy Path
// ---------------------------------------------------------------------------

func TestFrontmatterParse_AllFields(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
depends_on:
  - dep-a
  - dep-b
external:
  - path: ext/file.md
    fragments:
      - description: "My fragment"
        lines: "1-10"
        hash: "abc123"
input: some/input/file.md
outputs:
  - id: out-1
    path: some/path/file.go
  - id: out-2
    path: other/path/file.go
---
Body content here.
`
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 2 || fm.DependsOn[0] != "dep-a" || fm.DependsOn[1] != "dep-b" {
		t.Errorf("depends_on = %v, want [dep-a dep-b]", fm.DependsOn)
	}

	if len(fm.External) != 1 {
		t.Fatalf("external len = %d, want 1", len(fm.External))
	}
	if fm.External[0].Path != "ext/file.md" {
		t.Errorf("external[0].path = %q, want %q", fm.External[0].Path, "ext/file.md")
	}
	if len(fm.External[0].Fragments) != 1 {
		t.Fatalf("external[0].fragments len = %d, want 1", len(fm.External[0].Fragments))
	}
	frag := fm.External[0].Fragments[0]
	if frag.Description != "My fragment" {
		t.Errorf("fragment.description = %q, want %q", frag.Description, "My fragment")
	}
	if frag.Lines != "1-10" {
		t.Errorf("fragment.lines = %q, want %q", frag.Lines, "1-10")
	}
	if frag.Hash != "abc123" {
		t.Errorf("fragment.hash = %q, want %q", frag.Hash, "abc123")
	}

	if fm.Input != "some/input/file.md" {
		t.Errorf("input = %q, want %q", fm.Input, "some/input/file.md")
	}

	if len(fm.Outputs) != 2 {
		t.Fatalf("outputs len = %d, want 2", len(fm.Outputs))
	}
	if fm.Outputs[0].ID != "out-1" || fm.Outputs[0].Path != "some/path/file.go" {
		t.Errorf("outputs[0] = {%q, %q}, want {out-1, some/path/file.go}", fm.Outputs[0].ID, fm.Outputs[0].Path)
	}
	if fm.Outputs[1].ID != "out-2" || fm.Outputs[1].Path != "other/path/file.go" {
		t.Errorf("outputs[1] = {%q, %q}, want {out-2, other/path/file.go}", fm.Outputs[1].ID, fm.Outputs[1].Path)
	}
}

func TestFrontmatterParse_OnlyOutputs(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
outputs:
  - id: out-1
    path: some/path/file.go
---
`
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("depends_on = %v, want empty", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("external = %v, want empty", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("input = %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 1 {
		t.Fatalf("outputs len = %d, want 1", len(fm.Outputs))
	}
	if fm.Outputs[0].ID != "out-1" || fm.Outputs[0].Path != "some/path/file.go" {
		t.Errorf("outputs[0] = {%q, %q}, want {out-1, some/path/file.go}", fm.Outputs[0].ID, fm.Outputs[0].Path)
	}
}

func TestFrontmatterParse_OnlyDependsOn(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
depends_on:
  - dep-x
  - dep-y
  - dep-z
---
`
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 3 || fm.DependsOn[0] != "dep-x" || fm.DependsOn[1] != "dep-y" || fm.DependsOn[2] != "dep-z" {
		t.Errorf("depends_on = %v, want [dep-x dep-y dep-z]", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("external = %v, want empty", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("input = %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("outputs = %v, want empty", fm.Outputs)
	}
}

func TestFrontmatterParse_ExternalWithFragments(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
external:
  - path: first/ext.md
    fragments:
      - description: "Frag one"
        lines: "1-5"
        hash: "hash1"
      - description: "Frag two"
        lines: "10-20"
        hash: "hash2"
  - path: second/ext.md
---
`
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.External) != 2 {
		t.Fatalf("external len = %d, want 2", len(fm.External))
	}

	first := fm.External[0]
	if first.Path != "first/ext.md" {
		t.Errorf("external[0].path = %q, want %q", first.Path, "first/ext.md")
	}
	if len(first.Fragments) != 2 {
		t.Fatalf("external[0].fragments len = %d, want 2", len(first.Fragments))
	}
	if first.Fragments[0].Description != "Frag one" || first.Fragments[0].Lines != "1-5" || first.Fragments[0].Hash != "hash1" {
		t.Errorf("external[0].fragments[0] = %+v", first.Fragments[0])
	}
	if first.Fragments[1].Description != "Frag two" || first.Fragments[1].Lines != "10-20" || first.Fragments[1].Hash != "hash2" {
		t.Errorf("external[0].fragments[1] = %+v", first.Fragments[1])
	}

	second := fm.External[1]
	if second.Path != "second/ext.md" {
		t.Errorf("external[1].path = %q, want %q", second.Path, "second/ext.md")
	}
	if len(second.Fragments) != 0 {
		t.Errorf("external[1].fragments = %v, want empty", second.Fragments)
	}
}

func TestFrontmatterParse_OnlyInput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
input: path/to/input.md
---
`
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Input != "path/to/input.md" {
		t.Errorf("input = %q, want %q", fm.Input, "path/to/input.md")
	}
	if len(fm.DependsOn) != 0 {
		t.Errorf("depends_on = %v, want empty", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("external = %v, want empty", fm.External)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("outputs = %v, want empty", fm.Outputs)
	}
}

func TestFrontmatterParse_FragmentWithoutDescription(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
external:
  - path: some/file.md
    fragments:
      - lines: "5-15"
        hash: "hashABC"
---
`
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.External) != 1 || len(fm.External[0].Fragments) != 1 {
		t.Fatalf("unexpected external structure: %+v", fm.External)
	}
	frag := fm.External[0].Fragments[0]
	if frag.Description != "" {
		t.Errorf("fragment.description = %q, want empty", frag.Description)
	}
	if frag.Lines != "5-15" {
		t.Errorf("fragment.lines = %q, want %q", frag.Lines, "5-15")
	}
	if frag.Hash != "hashABC" {
		t.Errorf("fragment.hash = %q, want %q", frag.Hash, "hashABC")
	}
}

func TestFrontmatterParse_IgnoresUnknownFields(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
depends_on:
  - dep-a
custom_field: some_value
another_field: 42
---
`
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 1 || fm.DependsOn[0] != "dep-a" {
		t.Errorf("depends_on = %v, want [dep-a]", fm.DependsOn)
	}
}

func TestFrontmatterParse_NoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `This file has no frontmatter at all.
Just plain body text.
`
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("depends_on = %v, want empty", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("external = %v, want empty", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("input = %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("outputs = %v, want empty", fm.Outputs)
	}
}

// ---------------------------------------------------------------------------
// Edge Cases
// ---------------------------------------------------------------------------

func TestFrontmatterParse_EmptyFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
---
`
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("depends_on = %v, want empty", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("external = %v, want empty", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("input = %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("outputs = %v, want empty", fm.Outputs)
	}
}

func TestFrontmatterParse_OnlyFrontmatterNoBody(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
depends_on:
  - dep-only
---`
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 1 || fm.DependsOn[0] != "dep-only" {
		t.Errorf("depends_on = %v, want [dep-only]", fm.DependsOn)
	}
}

func TestFrontmatterParse_DelimiterWithTrailingWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// The first line has trailing spaces — should NOT be treated as a frontmatter delimiter.
	content := "---   \ndepends_on:\n  - dep-a\n"
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("depends_on = %v, want empty (delimiter with trailing spaces must not be recognized)", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("external = %v, want empty", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("input = %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("outputs = %v, want empty", fm.Outputs)
	}
}

// ---------------------------------------------------------------------------
// Failure Cases
// ---------------------------------------------------------------------------

func TestFrontmatterParse_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "nonexistent/file.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("error = %v, want ErrFileUnreadable", err)
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
		t.Errorf("error = %v, want ErrDirectoryTraversal", err)
	}
}

func TestFrontmatterParse_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
depends_on: [unclosed bracket
---
`
	testWriteFile(t, "node.md", content)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("error = %v, want ErrMalformedYAML", err)
	}
}

func TestFrontmatterParse_UnclosedFrontmatterBlock(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
depends_on:
  - dep-a
No closing delimiter ever appears.
`
	testWriteFile(t, "node.md", content)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("error = %v, want ErrMalformedYAML", err)
	}
}

func TestFrontmatterParse_ExternalEntryMissingPath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
external:
  - fragments:
      - lines: "1-5"
        hash: "abc"
---
`
	testWriteFile(t, "node.md", content)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("error = %v, want ErrMalformedYAML", err)
	}
}

func TestFrontmatterParse_FragmentMissingHash(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
external:
  - path: some/file.md
    fragments:
      - lines: "1-5"
---
`
	testWriteFile(t, "node.md", content)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("error = %v, want ErrMalformedYAML", err)
	}
}

func TestFrontmatterParse_OutputEntryMissingPath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
outputs:
  - id: out-1
---
`
	testWriteFile(t, "node.md", content)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("error = %v, want ErrMalformedYAML", err)
	}
}
