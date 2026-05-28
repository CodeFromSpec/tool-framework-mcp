// code-from-spec: ROOT/golang/tests/parsing/frontmatter@jBjMz_m-gGxg81q2r83dkVFRgWM

package frontmatter_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
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

// testWriteFile writes content to filename inside the current working directory,
// creating any intermediate directories as needed.
func testWriteFile(t *testing.T, filename, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
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
  - dep/one
  - dep/two
external:
  - path: ext/path/one
    fragments:
      - description: Fragment description
        lines: "line one\nline two"
        hash: abc123
input: some input value
outputs:
  - id: out1
    path: out/path/one
  - id: out2
    path: out/path/two
---
Body content here.
`
	testWriteFile(t, "spec.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// depends_on
	if len(fm.DependsOn) != 2 {
		t.Fatalf("DependsOn: got %d items, want 2", len(fm.DependsOn))
	}
	if *fm.DependsOn[0] != "dep/one" {
		t.Errorf("DependsOn[0]: got %q, want %q", *fm.DependsOn[0], "dep/one")
	}
	if *fm.DependsOn[1] != "dep/two" {
		t.Errorf("DependsOn[1]: got %q, want %q", *fm.DependsOn[1], "dep/two")
	}

	// external
	if len(fm.External) != 1 {
		t.Fatalf("External: got %d items, want 1", len(fm.External))
	}
	if fm.External[0].Path != "ext/path/one" {
		t.Errorf("External[0].Path: got %q, want %q", fm.External[0].Path, "ext/path/one")
	}
	if len(fm.External[0].Fragments) != 1 {
		t.Fatalf("External[0].Fragments: got %d items, want 1", len(fm.External[0].Fragments))
	}
	frag := fm.External[0].Fragments[0]
	if frag.Description != "Fragment description" {
		t.Errorf("fragment Description: got %q, want %q", frag.Description, "Fragment description")
	}
	if frag.Lines != "line one\nline two" {
		t.Errorf("fragment Lines: got %q, want %q", frag.Lines, "line one\nline two")
	}
	if frag.Hash != "abc123" {
		t.Errorf("fragment Hash: got %q, want %q", frag.Hash, "abc123")
	}

	// input
	if fm.Input != "some input value" {
		t.Errorf("Input: got %q, want %q", fm.Input, "some input value")
	}

	// outputs
	if len(fm.Outputs) != 2 {
		t.Fatalf("Outputs: got %d items, want 2", len(fm.Outputs))
	}
	if fm.Outputs[0].ID != "out1" || fm.Outputs[0].Path != "out/path/one" {
		t.Errorf("Outputs[0]: got id=%q path=%q, want id=%q path=%q",
			fm.Outputs[0].ID, fm.Outputs[0].Path, "out1", "out/path/one")
	}
	if fm.Outputs[1].ID != "out2" || fm.Outputs[1].Path != "out/path/two" {
		t.Errorf("Outputs[1]: got id=%q path=%q, want id=%q path=%q",
			fm.Outputs[1].ID, fm.Outputs[1].Path, "out2", "out/path/two")
	}
}

func TestFrontmatterParse_OnlyOutputs(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
outputs:
  - id: myid
    path: my/path
---
`
	testWriteFile(t, "spec.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: got %d items, want 0", len(fm.DependsOn))
	}
	if len(fm.External) != 0 {
		t.Errorf("External: got %d items, want 0", len(fm.External))
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 1 {
		t.Fatalf("Outputs: got %d items, want 1", len(fm.Outputs))
	}
	if fm.Outputs[0].ID != "myid" || fm.Outputs[0].Path != "my/path" {
		t.Errorf("Outputs[0]: got id=%q path=%q, want id=%q path=%q",
			fm.Outputs[0].ID, fm.Outputs[0].Path, "myid", "my/path")
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
	testWriteFile(t, "spec.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 2 {
		t.Fatalf("DependsOn: got %d items, want 2", len(fm.DependsOn))
	}
	if *fm.DependsOn[0] != "first/dep" {
		t.Errorf("DependsOn[0]: got %q, want %q", *fm.DependsOn[0], "first/dep")
	}
	if *fm.DependsOn[1] != "second/dep" {
		t.Errorf("DependsOn[1]: got %q, want %q", *fm.DependsOn[1], "second/dep")
	}
	if len(fm.External) != 0 {
		t.Errorf("External: got %d items, want 0", len(fm.External))
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("Outputs: got %d items, want 0", len(fm.Outputs))
	}
}

func TestFrontmatterParse_ExternalWithMultipleFragments(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
external:
  - path: ext/first
    fragments:
      - description: Frag A
        lines: "line a"
        hash: hashA
      - description: Frag B
        lines: "line b"
        hash: hashB
  - path: ext/second
---
`
	testWriteFile(t, "spec.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.External) != 2 {
		t.Fatalf("External: got %d items, want 2", len(fm.External))
	}

	first := fm.External[0]
	if first.Path != "ext/first" {
		t.Errorf("External[0].Path: got %q, want %q", first.Path, "ext/first")
	}
	if len(first.Fragments) != 2 {
		t.Fatalf("External[0].Fragments: got %d items, want 2", len(first.Fragments))
	}
	if first.Fragments[0].Description != "Frag A" || first.Fragments[0].Lines != "line a" || first.Fragments[0].Hash != "hashA" {
		t.Errorf("External[0].Fragments[0]: got desc=%q lines=%q hash=%q",
			first.Fragments[0].Description, first.Fragments[0].Lines, first.Fragments[0].Hash)
	}
	if first.Fragments[1].Description != "Frag B" || first.Fragments[1].Lines != "line b" || first.Fragments[1].Hash != "hashB" {
		t.Errorf("External[0].Fragments[1]: got desc=%q lines=%q hash=%q",
			first.Fragments[1].Description, first.Fragments[1].Lines, first.Fragments[1].Hash)
	}

	second := fm.External[1]
	if second.Path != "ext/second" {
		t.Errorf("External[1].Path: got %q, want %q", second.Path, "ext/second")
	}
	if len(second.Fragments) != 0 {
		t.Errorf("External[1].Fragments: got %d items, want 0", len(second.Fragments))
	}
}

func TestFrontmatterParse_OnlyInput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
input: my input content
---
`
	testWriteFile(t, "spec.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Input != "my input content" {
		t.Errorf("Input: got %q, want %q", fm.Input, "my input content")
	}
	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: got %d items, want 0", len(fm.DependsOn))
	}
	if len(fm.External) != 0 {
		t.Errorf("External: got %d items, want 0", len(fm.External))
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("Outputs: got %d items, want 0", len(fm.Outputs))
	}
}

func TestFrontmatterParse_FragmentWithoutDescription(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
external:
  - path: ext/path
    fragments:
      - lines: "some lines"
        hash: hashval
---
`
	testWriteFile(t, "spec.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.External) != 1 {
		t.Fatalf("External: got %d items, want 1", len(fm.External))
	}
	if len(fm.External[0].Fragments) != 1 {
		t.Fatalf("External[0].Fragments: got %d items, want 1", len(fm.External[0].Fragments))
	}
	frag := fm.External[0].Fragments[0]
	if frag.Description != "" {
		t.Errorf("fragment Description: got %q, want empty", frag.Description)
	}
	if frag.Lines != "some lines" {
		t.Errorf("fragment Lines: got %q, want %q", frag.Lines, "some lines")
	}
	if frag.Hash != "hashval" {
		t.Errorf("fragment Hash: got %q, want %q", frag.Hash, "hashval")
	}
}

func TestFrontmatterParse_IgnoresUnknownFields(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
depends_on:
  - known/dep
outputs:
  - id: myid
    path: my/path
custom_field: some value
---
`
	testWriteFile(t, "spec.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 1 || *fm.DependsOn[0] != "known/dep" {
		t.Errorf("DependsOn: unexpected value")
	}
	if len(fm.Outputs) != 1 || fm.Outputs[0].ID != "myid" || fm.Outputs[0].Path != "my/path" {
		t.Errorf("Outputs: unexpected value")
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
		t.Errorf("DependsOn: got %d items, want 0", len(fm.DependsOn))
	}
	if len(fm.External) != 0 {
		t.Errorf("External: got %d items, want 0", len(fm.External))
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("Outputs: got %d items, want 0", len(fm.Outputs))
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
Body content.
`
	testWriteFile(t, "spec.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: got %d items, want 0", len(fm.DependsOn))
	}
	if len(fm.External) != 0 {
		t.Errorf("External: got %d items, want 0", len(fm.External))
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("Outputs: got %d items, want 0", len(fm.Outputs))
	}
}

func TestFrontmatterParse_OnlyFrontmatterNoBody(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
input: only frontmatter
---
`
	testWriteFile(t, "spec.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Input != "only frontmatter" {
		t.Errorf("Input: got %q, want %q", fm.Input, "only frontmatter")
	}
}

func TestFrontmatterParse_DelimiterWithTrailingWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// First line has trailing spaces — should NOT be recognized as a delimiter.
	content := "---   \nThis is body content.\n"
	testWriteFile(t, "spec.md", content)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: got %d items, want 0", len(fm.DependsOn))
	}
	if len(fm.External) != 0 {
		t.Errorf("External: got %d items, want 0", len(fm.External))
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("Outputs: got %d items, want 0", len(fm.Outputs))
	}
}

// ---------------------------------------------------------------------------
// Failure Cases
// ---------------------------------------------------------------------------

func TestFrontmatterParse_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "nonexistent.md"})
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
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got: %v", err)
	}
}

func TestFrontmatterParse_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
key: valid
  bad_indent: broken
    : also broken
---
`
	testWriteFile(t, "spec.md", content)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got: %v", err)
	}
}

func TestFrontmatterParse_UnclosedFrontmatterBlock(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
input: some value
depends_on:
  - dep/one
`
	testWriteFile(t, "spec.md", content)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got: %v", err)
	}
}

func TestFrontmatterParse_MissingPathInExternalEntry(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
external:
  - fragments:
      - lines: "some lines"
        hash: hashval
---
`
	testWriteFile(t, "spec.md", content)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got: %v", err)
	}
}

func TestFrontmatterParse_MissingHashInFragment(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
external:
  - path: ext/path
    fragments:
      - lines: "some lines"
---
`
	testWriteFile(t, "spec.md", content)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got: %v", err)
	}
}

func TestFrontmatterParse_MissingPathInOutputEntry(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	content := `---
outputs:
  - id: myid
---
`
	testWriteFile(t, "spec.md", content)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "spec.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got: %v", err)
	}
}
