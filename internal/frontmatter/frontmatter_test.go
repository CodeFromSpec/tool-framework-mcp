// code-from-spec: ROOT/golang/tests/parsing/frontmatter@BNr95Vpxl9F8nD1z0NIdQ4_KaQ4
package frontmatter_test

import (
	"errors"
	"os"
	"testing"

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

// testWriteFile creates intermediate directories and writes content to a file
// relative to the current working directory.
func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("testWriteFile %q: %v", path, err)
	}
}

// --- Happy Path ---

func TestFrontmatterParse_AllFields(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
depends_on:
  - "dep-one"
  - "dep-two"
external:
  - path: "some/external/file.md"
    fragments:
      - description: "Fragment description"
        lines: "10-20"
        hash: "abc123"
input: "some/input/file.md"
outputs:
  - id: "out-one"
    path: "path/to/out-one.go"
  - id: "out-two"
    path: "path/to/out-two.go"
---
body content here
`
	testWriteFile(t, "spec.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 2 || fm.DependsOn[0] != "dep-one" || fm.DependsOn[1] != "dep-two" {
		t.Errorf("depends_on = %v, want [dep-one dep-two]", fm.DependsOn)
	}

	if len(fm.External) != 1 {
		t.Fatalf("external len = %d, want 1", len(fm.External))
	}
	ext := fm.External[0]
	if ext.Path != "some/external/file.md" {
		t.Errorf("external[0].path = %q, want %q", ext.Path, "some/external/file.md")
	}
	if len(ext.Fragments) != 1 {
		t.Fatalf("external[0].fragments len = %d, want 1", len(ext.Fragments))
	}
	frag := ext.Fragments[0]
	if frag.Description != "Fragment description" {
		t.Errorf("fragment.description = %q, want %q", frag.Description, "Fragment description")
	}
	if frag.Lines != "10-20" {
		t.Errorf("fragment.lines = %q, want %q", frag.Lines, "10-20")
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
	if fm.Outputs[0].ID != "out-one" || fm.Outputs[0].Path != "path/to/out-one.go" {
		t.Errorf("outputs[0] = {%q %q}, want {out-one path/to/out-one.go}", fm.Outputs[0].ID, fm.Outputs[0].Path)
	}
	if fm.Outputs[1].ID != "out-two" || fm.Outputs[1].Path != "path/to/out-two.go" {
		t.Errorf("outputs[1] = {%q %q}, want {out-two path/to/out-two.go}", fm.Outputs[1].ID, fm.Outputs[1].Path)
	}
}

func TestFrontmatterParse_OnlyOutputs(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
outputs:
  - id: "result"
    path: "gen/result.go"
---
`
	testWriteFile(t, "spec.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
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
	if fm.Outputs[0].ID != "result" || fm.Outputs[0].Path != "gen/result.go" {
		t.Errorf("outputs[0] = {%q %q}, want {result gen/result.go}", fm.Outputs[0].ID, fm.Outputs[0].Path)
	}
}

func TestFrontmatterParse_OnlyDependsOn(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
depends_on:
  - "alpha"
  - "beta"
---
`
	testWriteFile(t, "spec.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 2 || fm.DependsOn[0] != "alpha" || fm.DependsOn[1] != "beta" {
		t.Errorf("depends_on = %v, want [alpha beta]", fm.DependsOn)
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
  - path: "first/path.md"
    fragments:
      - description: "First fragment"
        lines: "1-5"
        hash: "hash-one"
      - description: "Second fragment"
        lines: "7-9"
        hash: "hash-two"
  - path: "second/path.md"
---
`
	testWriteFile(t, "spec.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.External) != 2 {
		t.Fatalf("external len = %d, want 2", len(fm.External))
	}

	first := fm.External[0]
	if first.Path != "first/path.md" {
		t.Errorf("external[0].path = %q, want %q", first.Path, "first/path.md")
	}
	if len(first.Fragments) != 2 {
		t.Fatalf("external[0].fragments len = %d, want 2", len(first.Fragments))
	}
	if first.Fragments[0].Description != "First fragment" || first.Fragments[0].Lines != "1-5" || first.Fragments[0].Hash != "hash-one" {
		t.Errorf("external[0].fragments[0] = %+v", first.Fragments[0])
	}
	if first.Fragments[1].Description != "Second fragment" || first.Fragments[1].Lines != "7-9" || first.Fragments[1].Hash != "hash-two" {
		t.Errorf("external[0].fragments[1] = %+v", first.Fragments[1])
	}

	second := fm.External[1]
	if second.Path != "second/path.md" {
		t.Errorf("external[1].path = %q, want %q", second.Path, "second/path.md")
	}
	if len(second.Fragments) != 0 {
		t.Errorf("external[1].fragments = %v, want empty", second.Fragments)
	}
}

func TestFrontmatterParse_OnlyInput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
input: "data/source.txt"
---
`
	testWriteFile(t, "spec.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Input != "data/source.txt" {
		t.Errorf("input = %q, want %q", fm.Input, "data/source.txt")
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
  - path: "some/file.md"
    fragments:
      - lines: "3-8"
        hash: "no-desc-hash"
---
`
	testWriteFile(t, "spec.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.External) != 1 {
		t.Fatalf("external len = %d, want 1", len(fm.External))
	}
	if len(fm.External[0].Fragments) != 1 {
		t.Fatalf("external[0].fragments len = %d, want 1", len(fm.External[0].Fragments))
	}
	frag := fm.External[0].Fragments[0]
	if frag.Description != "" {
		t.Errorf("fragment.description = %q, want empty", frag.Description)
	}
	if frag.Lines != "3-8" {
		t.Errorf("fragment.lines = %q, want %q", frag.Lines, "3-8")
	}
	if frag.Hash != "no-desc-hash" {
		t.Errorf("fragment.hash = %q, want %q", frag.Hash, "no-desc-hash")
	}
}

func TestFrontmatterParse_IgnoresUnknownFields(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
depends_on:
  - "known-dep"
custom_field: "ignored value"
another_unknown: 42
---
`
	testWriteFile(t, "spec.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 1 || fm.DependsOn[0] != "known-dep" {
		t.Errorf("depends_on = %v, want [known-dep]", fm.DependsOn)
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

func TestFrontmatterParse_NoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `This is just body content.
No frontmatter here.
`
	testWriteFile(t, "spec.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
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

// --- Edge Cases ---

func TestFrontmatterParse_EmptyFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
---
`
	testWriteFile(t, "spec.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
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
  - "lonely-dep"
---
`
	testWriteFile(t, "spec.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 1 || fm.DependsOn[0] != "lonely-dep" {
		t.Errorf("depends_on = %v, want [lonely-dep]", fm.DependsOn)
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

func TestFrontmatterParse_DelimiterWithTrailingWhitespaceNotRecognized(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// First line is "---   " (dashes + spaces), not a valid delimiter.
	content := "---   \ndepends_on:\n  - \"something\"\nbody content\n"
	testWriteFile(t, "spec.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
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

// --- Failure Cases ---

func TestFrontmatterParse_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "nonexistent.md"})
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

	content := `---
depends_on: [unclosed bracket
---
`
	testWriteFile(t, "spec.md", content)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
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
  - "something"
body content with no closing delimiter
`
	testWriteFile(t, "spec.md", content)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
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
  - fragments:
      - lines: "1-2"
        hash: "abc"
---
`
	testWriteFile(t, "spec.md", content)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("error = %v, want ErrMalformedYAML", err)
	}
}

func TestFrontmatterParse_MissingHashInFragment(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
external:
  - path: "some/file.md"
    fragments:
      - lines: "1-5"
---
`
	testWriteFile(t, "spec.md", content)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("error = %v, want ErrMalformedYAML", err)
	}
}

func TestFrontmatterParse_MissingPathInOutputEntry(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
outputs:
  - id: "out-only-id"
---
`
	testWriteFile(t, "spec.md", content)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("error = %v, want ErrMalformedYAML", err)
	}
}
