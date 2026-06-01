// code-from-spec: ROOT/golang/tests/parsing/frontmatter@LZR8KX9GWFyhMvZom6gKCv3gbNg
package frontmatter_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// testChdir changes the working directory to dir and restores the
// original working directory when the test ends.
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

// testWriteFile writes content to a relative path within the current
// working directory, creating any intermediate directories as needed.
func testWriteFile(t *testing.T, relPath string, content string) {
	t.Helper()
	if err := os.MkdirAll(".", 0o755); err != nil {
		t.Fatalf("testWriteFile MkdirAll: %v", err)
	}
	if err := os.WriteFile(relPath, []byte(content), 0o644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

// testPath returns a PathCfs for the given relative path using forward slashes.
func testPath(relPath string) *pathutils.PathCfs {
	return &pathutils.PathCfs{Value: relPath}
}

// TestFrontmatterParse_TC_HP_01 tests parsing complete frontmatter with all fields.
func TestFrontmatterParse_TC_HP_01(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := `---
depends_on:
  - "dep/one"
  - "dep/two"
external:
  - path: "some/external/file.md"
    fragments:
      - description: "A fragment description"
        lines: "10-20"
        hash: "abc123"
input: "some/input/file.md"
outputs:
  - id: "output_one"
    path: "path/to/output_one.go"
  - id: "output_two"
    path: "path/to/output_two.go"
---
Body content here.
`
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(testPath("node.md"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// depends_on
	if len(fm.DependsOn) != 2 {
		t.Fatalf("expected 2 depends_on entries, got %d", len(fm.DependsOn))
	}
	if fm.DependsOn[0] != "dep/one" {
		t.Errorf("depends_on[0]: got %q, want %q", fm.DependsOn[0], "dep/one")
	}
	if fm.DependsOn[1] != "dep/two" {
		t.Errorf("depends_on[1]: got %q, want %q", fm.DependsOn[1], "dep/two")
	}

	// external
	if len(fm.External) != 1 {
		t.Fatalf("expected 1 external entry, got %d", len(fm.External))
	}
	ext := fm.External[0]
	if ext.Path != "some/external/file.md" {
		t.Errorf("external[0].path: got %q, want %q", ext.Path, "some/external/file.md")
	}
	if len(ext.Fragments) != 1 {
		t.Fatalf("expected 1 fragment, got %d", len(ext.Fragments))
	}
	frag := ext.Fragments[0]
	if frag.Description != "A fragment description" {
		t.Errorf("fragment.description: got %q, want %q", frag.Description, "A fragment description")
	}
	if frag.Lines != "10-20" {
		t.Errorf("fragment.lines: got %q, want %q", frag.Lines, "10-20")
	}
	if frag.Hash != "abc123" {
		t.Errorf("fragment.hash: got %q, want %q", frag.Hash, "abc123")
	}

	// input
	if fm.Input != "some/input/file.md" {
		t.Errorf("input: got %q, want %q", fm.Input, "some/input/file.md")
	}

	// outputs
	if len(fm.Outputs) != 2 {
		t.Fatalf("expected 2 output entries, got %d", len(fm.Outputs))
	}
	if fm.Outputs[0].ID != "output_one" {
		t.Errorf("outputs[0].id: got %q, want %q", fm.Outputs[0].ID, "output_one")
	}
	if fm.Outputs[0].Path != "path/to/output_one.go" {
		t.Errorf("outputs[0].path: got %q, want %q", fm.Outputs[0].Path, "path/to/output_one.go")
	}
	if fm.Outputs[1].ID != "output_two" {
		t.Errorf("outputs[1].id: got %q, want %q", fm.Outputs[1].ID, "output_two")
	}
	if fm.Outputs[1].Path != "path/to/output_two.go" {
		t.Errorf("outputs[1].path: got %q, want %q", fm.Outputs[1].Path, "path/to/output_two.go")
	}
}

// TestFrontmatterParse_TC_HP_02 tests parsing frontmatter with only outputs.
func TestFrontmatterParse_TC_HP_02(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := `---
outputs:
  - id: "my_output"
    path: "path/to/my_output.go"
---
Body content here.
`
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(testPath("node.md"))
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
	if len(fm.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(fm.Outputs))
	}
	if fm.Outputs[0].ID != "my_output" {
		t.Errorf("outputs[0].id: got %q, want %q", fm.Outputs[0].ID, "my_output")
	}
	if fm.Outputs[0].Path != "path/to/my_output.go" {
		t.Errorf("outputs[0].path: got %q, want %q", fm.Outputs[0].Path, "path/to/my_output.go")
	}
}

// TestFrontmatterParse_TC_HP_03 tests parsing frontmatter with only depends_on.
func TestFrontmatterParse_TC_HP_03(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := `---
depends_on:
  - "dep/alpha"
  - "dep/beta"
---
Body content here.
`
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(testPath("node.md"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 2 {
		t.Fatalf("expected 2 depends_on entries, got %d", len(fm.DependsOn))
	}
	if fm.DependsOn[0] != "dep/alpha" {
		t.Errorf("depends_on[0]: got %q, want %q", fm.DependsOn[0], "dep/alpha")
	}
	if fm.DependsOn[1] != "dep/beta" {
		t.Errorf("depends_on[1]: got %q, want %q", fm.DependsOn[1], "dep/beta")
	}
	if len(fm.External) != 0 {
		t.Errorf("expected empty external, got %v", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("expected empty input, got %q", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("expected empty outputs, got %v", fm.Outputs)
	}
}

// TestFrontmatterParse_TC_HP_04 tests parsing frontmatter with external and multiple fragments.
func TestFrontmatterParse_TC_HP_04(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := `---
external:
  - path: "docs/first.md"
    fragments:
      - description: "First fragment"
        lines: "1-5"
        hash: "hash001"
      - description: "Second fragment"
        lines: "10-15"
        hash: "hash002"
  - path: "docs/second.md"
---
Body content.
`
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(testPath("node.md"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.External) != 2 {
		t.Fatalf("expected 2 external entries, got %d", len(fm.External))
	}

	first := fm.External[0]
	if first.Path != "docs/first.md" {
		t.Errorf("external[0].path: got %q, want %q", first.Path, "docs/first.md")
	}
	if len(first.Fragments) != 2 {
		t.Fatalf("expected 2 fragments in first external, got %d", len(first.Fragments))
	}
	if first.Fragments[0].Description != "First fragment" {
		t.Errorf("fragments[0].description: got %q, want %q", first.Fragments[0].Description, "First fragment")
	}
	if first.Fragments[0].Lines != "1-5" {
		t.Errorf("fragments[0].lines: got %q, want %q", first.Fragments[0].Lines, "1-5")
	}
	if first.Fragments[0].Hash != "hash001" {
		t.Errorf("fragments[0].hash: got %q, want %q", first.Fragments[0].Hash, "hash001")
	}
	if first.Fragments[1].Description != "Second fragment" {
		t.Errorf("fragments[1].description: got %q, want %q", first.Fragments[1].Description, "Second fragment")
	}
	if first.Fragments[1].Lines != "10-15" {
		t.Errorf("fragments[1].lines: got %q, want %q", first.Fragments[1].Lines, "10-15")
	}
	if first.Fragments[1].Hash != "hash002" {
		t.Errorf("fragments[1].hash: got %q, want %q", first.Fragments[1].Hash, "hash002")
	}

	second := fm.External[1]
	if second.Path != "docs/second.md" {
		t.Errorf("external[1].path: got %q, want %q", second.Path, "docs/second.md")
	}
	if len(second.Fragments) != 0 {
		t.Errorf("expected empty fragments for second external, got %d", len(second.Fragments))
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("expected empty depends_on, got %v", fm.DependsOn)
	}
	if fm.Input != "" {
		t.Errorf("expected empty input, got %q", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("expected empty outputs, got %v", fm.Outputs)
	}
}

// TestFrontmatterParse_TC_HP_05 tests parsing frontmatter with only the input field.
func TestFrontmatterParse_TC_HP_05(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := `---
input: "path/to/input/source.md"
---
Body content.
`
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(testPath("node.md"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Input != "path/to/input/source.md" {
		t.Errorf("input: got %q, want %q", fm.Input, "path/to/input/source.md")
	}
	if len(fm.DependsOn) != 0 {
		t.Errorf("expected empty depends_on, got %v", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("expected empty external, got %v", fm.External)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("expected empty outputs, got %v", fm.Outputs)
	}
}

// TestFrontmatterParse_TC_HP_06 tests parsing a fragment without a description field.
func TestFrontmatterParse_TC_HP_06(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := `---
external:
  - path: "docs/nodesc.md"
    fragments:
      - lines: "5-10"
        hash: "hashXYZ"
---
`
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(testPath("node.md"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.External) != 1 {
		t.Fatalf("expected 1 external entry, got %d", len(fm.External))
	}
	ext := fm.External[0]
	if ext.Path != "docs/nodesc.md" {
		t.Errorf("external[0].path: got %q, want %q", ext.Path, "docs/nodesc.md")
	}
	if len(ext.Fragments) != 1 {
		t.Fatalf("expected 1 fragment, got %d", len(ext.Fragments))
	}
	frag := ext.Fragments[0]
	if frag.Description != "" {
		t.Errorf("fragment.description: expected empty, got %q", frag.Description)
	}
	if frag.Lines != "5-10" {
		t.Errorf("fragment.lines: got %q, want %q", frag.Lines, "5-10")
	}
	if frag.Hash != "hashXYZ" {
		t.Errorf("fragment.hash: got %q, want %q", frag.Hash, "hashXYZ")
	}
}

// TestFrontmatterParse_TC_HP_07 tests that unknown frontmatter fields are silently ignored.
func TestFrontmatterParse_TC_HP_07(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := `---
depends_on:
  - "dep/known"
custom_field: "some value"
another_unknown: 42
outputs:
  - id: "out1"
    path: "path/out1.go"
---
Body content.
`
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(testPath("node.md"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 1 || fm.DependsOn[0] != "dep/known" {
		t.Errorf("depends_on: got %v, want [dep/known]", fm.DependsOn)
	}
	if len(fm.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(fm.Outputs))
	}
	if fm.Outputs[0].ID != "out1" {
		t.Errorf("outputs[0].id: got %q, want %q", fm.Outputs[0].ID, "out1")
	}
	if fm.Outputs[0].Path != "path/out1.go" {
		t.Errorf("outputs[0].path: got %q, want %q", fm.Outputs[0].Path, "path/out1.go")
	}
}

// TestFrontmatterParse_TC_HP_08 tests that a file with no frontmatter returns empty result.
func TestFrontmatterParse_TC_HP_08(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := `This is just body content.
No frontmatter here.
`
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(testPath("node.md"))
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
	if len(fm.Outputs) != 0 {
		t.Errorf("expected empty outputs, got %v", fm.Outputs)
	}
}

// TestFrontmatterParse_TC_EC_01 tests parsing an empty frontmatter block.
func TestFrontmatterParse_TC_EC_01(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := `---
---
Some body content.
`
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(testPath("node.md"))
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
	if len(fm.Outputs) != 0 {
		t.Errorf("expected empty outputs, got %v", fm.Outputs)
	}
}

// TestFrontmatterParse_TC_EC_02 tests parsing a file with only a frontmatter block and no body.
func TestFrontmatterParse_TC_EC_02(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := `---
depends_on:
  - "dep/only"
---
`
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(testPath("node.md"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 1 || fm.DependsOn[0] != "dep/only" {
		t.Errorf("depends_on: got %v, want [dep/only]", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("expected empty external, got %v", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("expected empty input, got %q", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("expected empty outputs, got %v", fm.Outputs)
	}
}

// TestFrontmatterParse_TC_EC_03 tests that a delimiter with trailing whitespace is not recognized.
func TestFrontmatterParse_TC_EC_03(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	// Note: "---   " with trailing spaces is not a valid delimiter.
	content := "---   \nSome body content.\n"
	testWriteFile(t, "node.md", content)

	fm, err := frontmatter.FrontmatterParse(testPath("node.md"))
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
	if len(fm.Outputs) != 0 {
		t.Errorf("expected empty outputs, got %v", fm.Outputs)
	}
}

// TestFrontmatterParse_TC_FC_01 tests that a non-existent file returns ErrFileUnreadable.
func TestFrontmatterParse_TC_FC_01(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := frontmatter.FrontmatterParse(testPath("nonexistent/file.md"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

// TestFrontmatterParse_TC_FC_02 tests that a directory traversal path returns ErrDirectoryTraversal.
func TestFrontmatterParse_TC_FC_02(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := frontmatter.FrontmatterParse(testPath("../../outside"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got: %v", err)
	}
}

// TestFrontmatterParse_TC_FC_03 tests that malformed YAML returns ErrMalformedYAML.
func TestFrontmatterParse_TC_FC_03(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := `---
depends_on: [unclosed bracket
  - bad: yaml: here
---
Body content.
`
	testWriteFile(t, "node.md", content)

	_, err := frontmatter.FrontmatterParse(testPath("node.md"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got: %v", err)
	}
}

// TestFrontmatterParse_TC_FC_04 tests that an unclosed frontmatter block returns ErrMalformedYAML.
func TestFrontmatterParse_TC_FC_04(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := `---
depends_on:
  - "dep/one"
Body content with no closing delimiter.
`
	testWriteFile(t, "node.md", content)

	_, err := frontmatter.FrontmatterParse(testPath("node.md"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got: %v", err)
	}
}

// TestFrontmatterParse_TC_FC_05 tests that an external entry without path returns ErrMalformedYAML.
func TestFrontmatterParse_TC_FC_05(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := `---
external:
  - fragments:
      - lines: "1-5"
        hash: "abc"
---
`
	testWriteFile(t, "node.md", content)

	_, err := frontmatter.FrontmatterParse(testPath("node.md"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got: %v", err)
	}
}

// TestFrontmatterParse_TC_FC_06 tests that a fragment missing hash returns ErrMalformedYAML.
func TestFrontmatterParse_TC_FC_06(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := `---
external:
  - path: "docs/file.md"
    fragments:
      - description: "Some fragment"
        lines: "1-10"
---
`
	testWriteFile(t, "node.md", content)

	_, err := frontmatter.FrontmatterParse(testPath("node.md"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got: %v", err)
	}
}

// TestFrontmatterParse_TC_FC_07 tests that an output entry missing path returns ErrMalformedYAML.
func TestFrontmatterParse_TC_FC_07(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	content := `---
outputs:
  - id: "output_without_path"
---
`
	testWriteFile(t, "node.md", content)

	_, err := frontmatter.FrontmatterParse(testPath("node.md"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got: %v", err)
	}
}

// Ensure fmt is used (suppress unused import error if needed).
var _ = fmt.Sprintf
