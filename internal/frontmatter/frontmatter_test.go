// code-from-spec: ROOT/golang/tests/parsing/frontmatter@uL9aJGzKm8FNTdmECm0BC5WEe14

package frontmatter_test

import (
	"errors"
	"os"
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

// testWriteFile creates any intermediate directories and writes content to
// a relative path within the current working directory.
func testWriteFile(t *testing.T, relPath string, content string) {
	t.Helper()
	if err := os.MkdirAll(".", 0755); err != nil {
		t.Fatalf("testWriteFile MkdirAll: %v", err)
	}
	if err := os.WriteFile(relPath, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile WriteFile %s: %v", relPath, err)
	}
}

// ---------------------------------------------------------------------------
// Happy Path
// ---------------------------------------------------------------------------

func TestFrontmatterParse_CompleteAllFields(t *testing.T) {
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

	// depends_on
	if len(fm.DependsOn) != 2 {
		t.Fatalf("expected 2 depends_on, got %d", len(fm.DependsOn))
	}
	if fm.DependsOn[0] != "dep-one" {
		t.Errorf("DependsOn[0] = %q, want %q", fm.DependsOn[0], "dep-one")
	}
	if fm.DependsOn[1] != "dep-two" {
		t.Errorf("DependsOn[1] = %q, want %q", fm.DependsOn[1], "dep-two")
	}

	// external
	if len(fm.External) != 1 {
		t.Fatalf("expected 1 external, got %d", len(fm.External))
	}
	ext := fm.External[0]
	if ext.Path != "some/external/file.md" {
		t.Errorf("External[0].Path = %q, want %q", ext.Path, "some/external/file.md")
	}
	if len(ext.Fragments) != 1 {
		t.Fatalf("expected 1 fragment, got %d", len(ext.Fragments))
	}
	frag := ext.Fragments[0]
	if frag.Description != "Fragment description" {
		t.Errorf("Fragment.Description = %q, want %q", frag.Description, "Fragment description")
	}
	if frag.Lines != "10-20" {
		t.Errorf("Fragment.Lines = %q, want %q", frag.Lines, "10-20")
	}
	if frag.Hash != "abc123" {
		t.Errorf("Fragment.Hash = %q, want %q", frag.Hash, "abc123")
	}

	// input
	if fm.Input != "some/input/file.md" {
		t.Errorf("Input = %q, want %q", fm.Input, "some/input/file.md")
	}

	// outputs
	if len(fm.Outputs) != 2 {
		t.Fatalf("expected 2 outputs, got %d", len(fm.Outputs))
	}
	if fm.Outputs[0].ID != "out-one" {
		t.Errorf("Outputs[0].ID = %q, want %q", fm.Outputs[0].ID, "out-one")
	}
	if fm.Outputs[0].Path != "path/to/out-one.go" {
		t.Errorf("Outputs[0].Path = %q, want %q", fm.Outputs[0].Path, "path/to/out-one.go")
	}
	if fm.Outputs[1].ID != "out-two" {
		t.Errorf("Outputs[1].ID = %q, want %q", fm.Outputs[1].ID, "out-two")
	}
	if fm.Outputs[1].Path != "path/to/out-two.go" {
		t.Errorf("Outputs[1].Path = %q, want %q", fm.Outputs[1].Path, "path/to/out-two.go")
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
		t.Errorf("expected empty DependsOn, got %v", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("expected empty External, got %v", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("expected empty Input, got %q", fm.Input)
	}
	if len(fm.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(fm.Outputs))
	}
	if fm.Outputs[0].ID != "result" {
		t.Errorf("Outputs[0].ID = %q, want %q", fm.Outputs[0].ID, "result")
	}
	if fm.Outputs[0].Path != "gen/result.go" {
		t.Errorf("Outputs[0].Path = %q, want %q", fm.Outputs[0].Path, "gen/result.go")
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

	if len(fm.DependsOn) != 2 {
		t.Fatalf("expected 2 depends_on, got %d", len(fm.DependsOn))
	}
	if fm.DependsOn[0] != "alpha" {
		t.Errorf("DependsOn[0] = %q, want %q", fm.DependsOn[0], "alpha")
	}
	if fm.DependsOn[1] != "beta" {
		t.Errorf("DependsOn[1] = %q, want %q", fm.DependsOn[1], "beta")
	}
	if len(fm.External) != 0 {
		t.Errorf("expected empty External, got %v", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("expected empty Input, got %q", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("expected empty Outputs, got %v", fm.Outputs)
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
		t.Fatalf("expected 2 externals, got %d", len(fm.External))
	}

	first := fm.External[0]
	if first.Path != "first/path.md" {
		t.Errorf("External[0].Path = %q, want %q", first.Path, "first/path.md")
	}
	if len(first.Fragments) != 2 {
		t.Fatalf("External[0]: expected 2 fragments, got %d", len(first.Fragments))
	}
	if first.Fragments[0].Description != "First fragment" {
		t.Errorf("Fragment[0].Description = %q, want %q", first.Fragments[0].Description, "First fragment")
	}
	if first.Fragments[0].Lines != "1-5" {
		t.Errorf("Fragment[0].Lines = %q, want %q", first.Fragments[0].Lines, "1-5")
	}
	if first.Fragments[0].Hash != "hash-one" {
		t.Errorf("Fragment[0].Hash = %q, want %q", first.Fragments[0].Hash, "hash-one")
	}
	if first.Fragments[1].Description != "Second fragment" {
		t.Errorf("Fragment[1].Description = %q, want %q", first.Fragments[1].Description, "Second fragment")
	}
	if first.Fragments[1].Lines != "7-9" {
		t.Errorf("Fragment[1].Lines = %q, want %q", first.Fragments[1].Lines, "7-9")
	}
	if first.Fragments[1].Hash != "hash-two" {
		t.Errorf("Fragment[1].Hash = %q, want %q", first.Fragments[1].Hash, "hash-two")
	}

	second := fm.External[1]
	if second.Path != "second/path.md" {
		t.Errorf("External[1].Path = %q, want %q", second.Path, "second/path.md")
	}
	if len(second.Fragments) != 0 {
		t.Errorf("External[1]: expected no fragments, got %d", len(second.Fragments))
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
		t.Errorf("Input = %q, want %q", fm.Input, "data/source.txt")
	}
	if len(fm.DependsOn) != 0 {
		t.Errorf("expected empty DependsOn, got %v", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("expected empty External, got %v", fm.External)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("expected empty Outputs, got %v", fm.Outputs)
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
		t.Fatalf("expected 1 external, got %d", len(fm.External))
	}
	if len(fm.External[0].Fragments) != 1 {
		t.Fatalf("expected 1 fragment, got %d", len(fm.External[0].Fragments))
	}
	frag := fm.External[0].Fragments[0]
	if frag.Description != "" {
		t.Errorf("Fragment.Description = %q, want empty", frag.Description)
	}
	if frag.Lines != "3-8" {
		t.Errorf("Fragment.Lines = %q, want %q", frag.Lines, "3-8")
	}
	if frag.Hash != "no-desc-hash" {
		t.Errorf("Fragment.Hash = %q, want %q", frag.Hash, "no-desc-hash")
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
		t.Errorf("DependsOn = %v, want [known-dep]", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("expected empty External, got %v", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("expected empty Input, got %q", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("expected empty Outputs, got %v", fm.Outputs)
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
		t.Errorf("expected empty DependsOn, got %v", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("expected empty External, got %v", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("expected empty Input, got %q", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("expected empty Outputs, got %v", fm.Outputs)
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
	testWriteFile(t, "spec.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("expected empty DependsOn, got %v", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("expected empty External, got %v", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("expected empty Input, got %q", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("expected empty Outputs, got %v", fm.Outputs)
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
		t.Errorf("DependsOn = %v, want [lonely-dep]", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("expected empty External, got %v", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("expected empty Input, got %q", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("expected empty Outputs, got %v", fm.Outputs)
	}
}

func TestFrontmatterParse_DelimiterWithTrailingWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// First line is "---   " (dashes with trailing spaces) — not a valid delimiter.
	content := "---   \nbody content here\n"
	testWriteFile(t, "spec.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("expected empty DependsOn, got %v", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("expected empty External, got %v", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("expected empty Input, got %q", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("expected empty Outputs, got %v", fm.Outputs)
	}
}

// ---------------------------------------------------------------------------
// Failure Cases
// ---------------------------------------------------------------------------

func TestFrontmatterParse_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "nonexistent/file.txt"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filereader.ErrFileUnreadable) {
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
depends_on: [unclosed bracket
---
`
	testWriteFile(t, "spec.md", content)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
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
		t.Errorf("expected ErrMalformedYAML, got %v", err)
	}
}

func TestFrontmatterParse_MissingRequiredFieldInExternal(t *testing.T) {
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
		t.Errorf("expected ErrMalformedYAML, got %v", err)
	}
}

func TestFrontmatterParse_MissingRequiredFieldInFragment(t *testing.T) {
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
		t.Errorf("expected ErrMalformedYAML, got %v", err)
	}
}

func TestFrontmatterParse_MissingRequiredFieldInOutput(t *testing.T) {
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
		t.Errorf("expected ErrMalformedYAML, got %v", err)
	}
}
