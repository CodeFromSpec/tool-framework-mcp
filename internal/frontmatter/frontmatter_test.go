// code-from-spec: ROOT/golang/tests/internal/frontmatter@VeZtECqMvJazrXXhakTGBqD1eT0

package frontmatter_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

// testMakeCfs creates a PathCfs from a string value.
func testMakeCfs(value string) *pathutils.PathCfs {
	return &pathutils.PathCfs{Value: value}
}

// testWriteFile writes content to a file inside dir with the given name and
// returns a PathCfs for it relative to the project root.
func testWriteFile(t *testing.T, dir, name, content string) *pathutils.PathCfs {
	t.Helper()
	fullPath := filepath.Join(dir, name)
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
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
	// Convert OS separator to forward slash.
	cfsVal := filepath.ToSlash(rel)
	return &pathutils.PathCfs{Value: cfsVal}
}

// testStringPtr returns a pointer to s.
func testStringPtr(s string) *string {
	return &s
}

// ---------------------------------------------------------------------------
// Happy Path
// ---------------------------------------------------------------------------

func TestFrontmatterParse_AllFields(t *testing.T) {
	dir := t.TempDir()
	content := `---
depends_on:
  - dep/one
  - dep/two
external:
  - path: some/external/file.go
    fragments:
      - description: A fragment
        lines: "1-10"
        hash: abc123
input: some input value
outputs:
  - id: out1
    path: out/path/one.go
  - id: out2
    path: out/path/two.go
---
body content here
`
	cfs := testWriteFile(t, dir, "allfields.md", content)

	fm, err := frontmatter.FrontmatterParse(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// depends_on
	if len(fm.DependsOn) != 2 {
		t.Fatalf("DependsOn: got len %d, want 2", len(fm.DependsOn))
	}
	if fm.DependsOn[0] != "dep/one" {
		t.Errorf("DependsOn[0]: got %q, want %q", fm.DependsOn[0], "dep/one")
	}
	if fm.DependsOn[1] != "dep/two" {
		t.Errorf("DependsOn[1]: got %q, want %q", fm.DependsOn[1], "dep/two")
	}

	// external
	if len(fm.External) != 1 {
		t.Fatalf("External: got len %d, want 1", len(fm.External))
	}
	ext := fm.External[0]
	if ext.Path != "some/external/file.go" {
		t.Errorf("External[0].Path: got %q, want %q", ext.Path, "some/external/file.go")
	}
	if ext.Fragments == nil {
		t.Fatalf("External[0].Fragments: got nil, want non-nil")
	}
	frags := *ext.Fragments
	if len(frags) != 1 {
		t.Fatalf("External[0].Fragments: got len %d, want 1", len(frags))
	}
	frag := frags[0]
	if frag.Description == nil || *frag.Description != "A fragment" {
		t.Errorf("fragment.Description: got %v, want \"A fragment\"", frag.Description)
	}
	if frag.Lines != "1-10" {
		t.Errorf("fragment.Lines: got %q, want %q", frag.Lines, "1-10")
	}
	if frag.Hash != "abc123" {
		t.Errorf("fragment.Hash: got %q, want %q", frag.Hash, "abc123")
	}

	// input
	if fm.Input != "some input value" {
		t.Errorf("Input: got %q, want %q", fm.Input, "some input value")
	}

	// outputs
	if len(fm.Outputs) != 2 {
		t.Fatalf("Outputs: got len %d, want 2", len(fm.Outputs))
	}
	if fm.Outputs[0].ID != "out1" || fm.Outputs[0].Path != "out/path/one.go" {
		t.Errorf("Outputs[0]: got {%q, %q}, want {\"out1\", \"out/path/one.go\"}", fm.Outputs[0].ID, fm.Outputs[0].Path)
	}
	if fm.Outputs[1].ID != "out2" || fm.Outputs[1].Path != "out/path/two.go" {
		t.Errorf("Outputs[1]: got {%q, %q}, want {\"out2\", \"out/path/two.go\"}", fm.Outputs[1].ID, fm.Outputs[1].Path)
	}
}

func TestFrontmatterParse_OnlyOutputs(t *testing.T) {
	dir := t.TempDir()
	content := `---
outputs:
  - id: myout
    path: some/output.go
---
`
	cfs := testWriteFile(t, dir, "onlyoutputs.md", content)

	fm, err := frontmatter.FrontmatterParse(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: got len %d, want 0", len(fm.DependsOn))
	}
	if len(fm.External) != 0 {
		t.Errorf("External: got len %d, want 0", len(fm.External))
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 1 {
		t.Fatalf("Outputs: got len %d, want 1", len(fm.Outputs))
	}
	if fm.Outputs[0].ID != "myout" || fm.Outputs[0].Path != "some/output.go" {
		t.Errorf("Outputs[0]: got {%q, %q}, want {\"myout\", \"some/output.go\"}", fm.Outputs[0].ID, fm.Outputs[0].Path)
	}
}

func TestFrontmatterParse_OnlyDependsOn(t *testing.T) {
	dir := t.TempDir()
	content := `---
depends_on:
  - alpha/one
  - beta/two
---
`
	cfs := testWriteFile(t, dir, "onlydependson.md", content)

	fm, err := frontmatter.FrontmatterParse(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 2 {
		t.Fatalf("DependsOn: got len %d, want 2", len(fm.DependsOn))
	}
	if fm.DependsOn[0] != "alpha/one" {
		t.Errorf("DependsOn[0]: got %q, want %q", fm.DependsOn[0], "alpha/one")
	}
	if fm.DependsOn[1] != "beta/two" {
		t.Errorf("DependsOn[1]: got %q, want %q", fm.DependsOn[1], "beta/two")
	}
	if len(fm.External) != 0 {
		t.Errorf("External: got len %d, want 0", len(fm.External))
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("Outputs: got len %d, want 0", len(fm.Outputs))
	}
}

func TestFrontmatterParse_ExternalWithFragments(t *testing.T) {
	dir := t.TempDir()
	content := `---
external:
  - path: first/file.go
    fragments:
      - description: First frag
        lines: "1-5"
        hash: hash1
      - description: Second frag
        lines: "10-20"
        hash: hash2
  - path: second/file.go
---
`
	cfs := testWriteFile(t, dir, "externalfrags.md", content)

	fm, err := frontmatter.FrontmatterParse(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.External) != 2 {
		t.Fatalf("External: got len %d, want 2", len(fm.External))
	}

	first := fm.External[0]
	if first.Path != "first/file.go" {
		t.Errorf("External[0].Path: got %q, want %q", first.Path, "first/file.go")
	}
	if first.Fragments == nil {
		t.Fatalf("External[0].Fragments: got nil, want non-nil")
	}
	frags0 := *first.Fragments
	if len(frags0) != 2 {
		t.Fatalf("External[0].Fragments: got len %d, want 2", len(frags0))
	}
	if frags0[0].Description == nil || *frags0[0].Description != "First frag" {
		t.Errorf("frags0[0].Description: got %v, want \"First frag\"", frags0[0].Description)
	}
	if frags0[0].Lines != "1-5" {
		t.Errorf("frags0[0].Lines: got %q, want %q", frags0[0].Lines, "1-5")
	}
	if frags0[0].Hash != "hash1" {
		t.Errorf("frags0[0].Hash: got %q, want %q", frags0[0].Hash, "hash1")
	}
	if frags0[1].Description == nil || *frags0[1].Description != "Second frag" {
		t.Errorf("frags0[1].Description: got %v, want \"Second frag\"", frags0[1].Description)
	}
	if frags0[1].Lines != "10-20" {
		t.Errorf("frags0[1].Lines: got %q, want %q", frags0[1].Lines, "10-20")
	}
	if frags0[1].Hash != "hash2" {
		t.Errorf("frags0[1].Hash: got %q, want %q", frags0[1].Hash, "hash2")
	}

	second := fm.External[1]
	if second.Path != "second/file.go" {
		t.Errorf("External[1].Path: got %q, want %q", second.Path, "second/file.go")
	}
	if second.Fragments != nil {
		t.Errorf("External[1].Fragments: got non-nil, want nil")
	}
}

func TestFrontmatterParse_OnlyInput(t *testing.T) {
	dir := t.TempDir()
	content := `---
input: hello world input
---
`
	cfs := testWriteFile(t, dir, "onlyinput.md", content)

	fm, err := frontmatter.FrontmatterParse(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Input != "hello world input" {
		t.Errorf("Input: got %q, want %q", fm.Input, "hello world input")
	}
	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: got len %d, want 0", len(fm.DependsOn))
	}
	if len(fm.External) != 0 {
		t.Errorf("External: got len %d, want 0", len(fm.External))
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("Outputs: got len %d, want 0", len(fm.Outputs))
	}
}

func TestFrontmatterParse_FragmentWithoutDescription(t *testing.T) {
	dir := t.TempDir()
	content := `---
external:
  - path: some/path.go
    fragments:
      - lines: "5-15"
        hash: nodesc99
---
`
	cfs := testWriteFile(t, dir, "nodesc.md", content)

	fm, err := frontmatter.FrontmatterParse(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.External) != 1 {
		t.Fatalf("External: got len %d, want 1", len(fm.External))
	}
	if fm.External[0].Fragments == nil {
		t.Fatalf("External[0].Fragments: got nil, want non-nil")
	}
	frags := *fm.External[0].Fragments
	if len(frags) != 1 {
		t.Fatalf("fragments: got len %d, want 1", len(frags))
	}
	if frags[0].Description != nil {
		t.Errorf("fragment.Description: got %v, want nil", frags[0].Description)
	}
	if frags[0].Lines != "5-15" {
		t.Errorf("fragment.Lines: got %q, want %q", frags[0].Lines, "5-15")
	}
	if frags[0].Hash != "nodesc99" {
		t.Errorf("fragment.Hash: got %q, want %q", frags[0].Hash, "nodesc99")
	}
}

func TestFrontmatterParse_IgnoresUnknownFields(t *testing.T) {
	dir := t.TempDir()
	content := `---
depends_on:
  - known/dep
outputs:
  - id: outid
    path: out/path.go
custom_field: some value
---
`
	cfs := testWriteFile(t, dir, "unknownfields.md", content)

	fm, err := frontmatter.FrontmatterParse(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 1 || fm.DependsOn[0] != "known/dep" {
		t.Errorf("DependsOn: got %v, want [\"known/dep\"]", fm.DependsOn)
	}
	if len(fm.Outputs) != 1 || fm.Outputs[0].ID != "outid" || fm.Outputs[0].Path != "out/path.go" {
		t.Errorf("Outputs: got %v", fm.Outputs)
	}
}

func TestFrontmatterParse_NoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	content := "just some body content\nno frontmatter here\n"
	cfs := testWriteFile(t, dir, "nobody.md", content)

	fm, err := frontmatter.FrontmatterParse(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: got len %d, want 0", len(fm.DependsOn))
	}
	if len(fm.External) != 0 {
		t.Errorf("External: got len %d, want 0", len(fm.External))
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("Outputs: got len %d, want 0", len(fm.Outputs))
	}
}

// ---------------------------------------------------------------------------
// Edge Cases
// ---------------------------------------------------------------------------

func TestFrontmatterParse_EmptyFrontmatter(t *testing.T) {
	dir := t.TempDir()
	content := "---\n---\nbody\n"
	cfs := testWriteFile(t, dir, "emptyfm.md", content)

	fm, err := frontmatter.FrontmatterParse(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: got len %d, want 0", len(fm.DependsOn))
	}
	if len(fm.External) != 0 {
		t.Errorf("External: got len %d, want 0", len(fm.External))
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("Outputs: got len %d, want 0", len(fm.Outputs))
	}
}

func TestFrontmatterParse_FrontmatterOnlyNoBody(t *testing.T) {
	dir := t.TempDir()
	content := "---\ndepends_on:\n  - a/dep\n---\n"
	cfs := testWriteFile(t, dir, "fmonly.md", content)

	fm, err := frontmatter.FrontmatterParse(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 1 || fm.DependsOn[0] != "a/dep" {
		t.Errorf("DependsOn: got %v, want [\"a/dep\"]", fm.DependsOn)
	}
}

func TestFrontmatterParse_DelimiterWithTrailingWhitespaceNotRecognized(t *testing.T) {
	dir := t.TempDir()
	// First line has trailing spaces — must NOT be recognized as a delimiter.
	content := "---   \nbody content here\n"
	cfs := testWriteFile(t, dir, "trailingws.md", content)

	fm, err := frontmatter.FrontmatterParse(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: got len %d, want 0", len(fm.DependsOn))
	}
	if len(fm.External) != 0 {
		t.Errorf("External: got len %d, want 0", len(fm.External))
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("Outputs: got len %d, want 0", len(fm.Outputs))
	}
}

// ---------------------------------------------------------------------------
// Failure Cases
// ---------------------------------------------------------------------------

func TestFrontmatterParse_FileDoesNotExist(t *testing.T) {
	cfs := testMakeCfs("internal/frontmatter/nonexistent_file_xyz.md")

	_, err := frontmatter.FrontmatterParse(cfs)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestFrontmatterParse_PropagatesPathErrors(t *testing.T) {
	cfs := testMakeCfs("../../outside")

	_, err := frontmatter.FrontmatterParse(cfs)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got: %v", err)
	}
}

func TestFrontmatterParse_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	// Invalid YAML: broken indentation / structure.
	content := "---\nkey: :\n  - bad: [unclosed\n---\n"
	cfs := testWriteFile(t, dir, "malformed.md", content)

	_, err := frontmatter.FrontmatterParse(cfs)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got: %v", err)
	}
}

func TestFrontmatterParse_UnclosedFrontmatterBlock(t *testing.T) {
	dir := t.TempDir()
	content := "---\ndepends_on:\n  - some/dep\nno closing delimiter\n"
	cfs := testWriteFile(t, dir, "unclosed.md", content)

	_, err := frontmatter.FrontmatterParse(cfs)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got: %v", err)
	}
}

func TestFrontmatterParse_MissingRequiredFieldInExternal(t *testing.T) {
	dir := t.TempDir()
	// External entry with no path field.
	content := "---\nexternal:\n  - fragments:\n      - lines: \"1-5\"\n        hash: abc\n---\n"
	cfs := testWriteFile(t, dir, "missingpath.md", content)

	_, err := frontmatter.FrontmatterParse(cfs)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got: %v", err)
	}
}

func TestFrontmatterParse_MissingRequiredFieldInFragment(t *testing.T) {
	dir := t.TempDir()
	// Fragment has lines but no hash.
	content := "---\nexternal:\n  - path: some/path.go\n    fragments:\n      - lines: \"1-5\"\n---\n"
	cfs := testWriteFile(t, dir, "missinghash.md", content)

	_, err := frontmatter.FrontmatterParse(cfs)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got: %v", err)
	}
}

func TestFrontmatterParse_MissingRequiredFieldInOutput(t *testing.T) {
	dir := t.TempDir()
	// Output entry has id but no path.
	content := "---\noutputs:\n  - id: myout\n---\n"
	cfs := testWriteFile(t, dir, "missingoutpath.md", content)

	_, err := frontmatter.FrontmatterParse(cfs)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got: %v", err)
	}
}
