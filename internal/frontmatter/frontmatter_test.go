// code-from-spec: ROOT/golang/internal/frontmatter/tests@DejKmDhy2F2QkRnUOCGDk_Z9tM4

// Package frontmatter — test file.
//
// All tests use t.TempDir() for isolation. Test helper functions and types
// are prefixed with "test" to avoid collisions with unexported identifiers
// in the package under test (this file is in the same package).
package frontmatter

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// testWriteFile creates a file with the given content inside dir and returns
// its absolute path.
func testWriteFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
	return path
}

// ---------------------------------------------------------------------------
// Happy-path tests
// ---------------------------------------------------------------------------

// TestParseFrontmatter_CompleteAllFields verifies that every field is parsed
// correctly when a file contains all supported frontmatter keys.
func TestParseFrontmatter_CompleteAllFields(t *testing.T) {
	dir := t.TempDir()
	content := `---
depends_on:
  - ROOT/other
  - ROOT/architecture/backend
external:
  - path: CODE_FROM_SPEC.md
    fragments:
      - description: v3 format
        lines: "10-25"
        hash: abc123
input: ARTIFACT/some/artifact(id)
outputs:
  - id: config
    path: internal/config/config.go
  - id: config_test
    path: internal/config/config_test.go
---
`
	path := testWriteFile(t, dir, "node.md", content)

	fm, err := ParseFrontmatter(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// DependsOn
	wantDeps := []string{"ROOT/other", "ROOT/architecture/backend"}
	if len(fm.DependsOn) != len(wantDeps) {
		t.Fatalf("DependsOn: got %d entries, want %d", len(fm.DependsOn), len(wantDeps))
	}
	for i, dep := range wantDeps {
		if fm.DependsOn[i] != dep {
			t.Errorf("DependsOn[%d]: got %q, want %q", i, fm.DependsOn[i], dep)
		}
	}

	// External — one entry
	if len(fm.External) != 1 {
		t.Fatalf("External: got %d entries, want 1", len(fm.External))
	}
	ext := fm.External[0]
	if ext.Path != "CODE_FROM_SPEC.md" {
		t.Errorf("External[0].Path: got %q, want %q", ext.Path, "CODE_FROM_SPEC.md")
	}
	if len(ext.Fragments) != 1 {
		t.Fatalf("External[0].Fragments: got %d entries, want 1", len(ext.Fragments))
	}
	frag := ext.Fragments[0]
	if frag.Description != "v3 format" {
		t.Errorf("fragment Description: got %q, want %q", frag.Description, "v3 format")
	}
	if frag.Lines != "10-25" {
		t.Errorf("fragment Lines: got %q, want %q", frag.Lines, "10-25")
	}
	if frag.Hash != "abc123" {
		t.Errorf("fragment Hash: got %q, want %q", frag.Hash, "abc123")
	}

	// Input
	if fm.Input != "ARTIFACT/some/artifact(id)" {
		t.Errorf("Input: got %q, want %q", fm.Input, "ARTIFACT/some/artifact(id)")
	}

	// Outputs — two entries
	if len(fm.Outputs) != 2 {
		t.Fatalf("Outputs: got %d entries, want 2", len(fm.Outputs))
	}
	if fm.Outputs[0].ID != "config" {
		t.Errorf("Outputs[0].ID: got %q, want %q", fm.Outputs[0].ID, "config")
	}
	if fm.Outputs[0].Path != "internal/config/config.go" {
		t.Errorf("Outputs[0].Path: got %q, want %q", fm.Outputs[0].Path, "internal/config/config.go")
	}
	if fm.Outputs[1].ID != "config_test" {
		t.Errorf("Outputs[1].ID: got %q, want %q", fm.Outputs[1].ID, "config_test")
	}
	if fm.Outputs[1].Path != "internal/config/config_test.go" {
		t.Errorf("Outputs[1].Path: got %q, want %q", fm.Outputs[1].Path, "internal/config/config_test.go")
	}
}

// TestParseFrontmatter_OnlyOutputs verifies that a file with only the outputs
// key leaves all other fields nil/empty.
func TestParseFrontmatter_OnlyOutputs(t *testing.T) {
	dir := t.TempDir()
	content := `---
outputs:
  - id: main
    path: cmd/main.go
---
`
	path := testWriteFile(t, dir, "node.md", content)

	fm, err := ParseFrontmatter(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// DependsOn must be nil or empty (spec says nil; the implementation
	// normalises to an empty slice — both are acceptable empty states).
	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: expected empty, got %v", fm.DependsOn)
	}

	// External must be nil or empty.
	if len(fm.External) != 0 {
		t.Errorf("External: expected empty, got %v", fm.External)
	}

	// Input must be empty string.
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}

	// Outputs — one entry.
	if len(fm.Outputs) != 1 {
		t.Fatalf("Outputs: got %d entries, want 1", len(fm.Outputs))
	}
	if fm.Outputs[0].ID != "main" {
		t.Errorf("Outputs[0].ID: got %q, want %q", fm.Outputs[0].ID, "main")
	}
	if fm.Outputs[0].Path != "cmd/main.go" {
		t.Errorf("Outputs[0].Path: got %q, want %q", fm.Outputs[0].Path, "cmd/main.go")
	}
}

// TestParseFrontmatter_OnlyDependsOn verifies that a file with only the
// depends_on key leaves outputs and external nil/empty.
func TestParseFrontmatter_OnlyDependsOn(t *testing.T) {
	dir := t.TempDir()
	content := `---
depends_on:
  - ROOT/other/node
  - ROOT/another/node
---
`
	path := testWriteFile(t, dir, "node.md", content)

	fm, err := ParseFrontmatter(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantDeps := []string{"ROOT/other/node", "ROOT/another/node"}
	if len(fm.DependsOn) != len(wantDeps) {
		t.Fatalf("DependsOn: got %d entries, want %d", len(fm.DependsOn), len(wantDeps))
	}
	for i, dep := range wantDeps {
		if fm.DependsOn[i] != dep {
			t.Errorf("DependsOn[%d]: got %q, want %q", i, fm.DependsOn[i], dep)
		}
	}

	if len(fm.External) != 0 {
		t.Errorf("External: expected empty, got %v", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("Outputs: expected empty, got %v", fm.Outputs)
	}
}

// TestParseFrontmatter_ExternalWithFragments verifies that multiple external
// entries and their fragments are parsed correctly, including entries without
// any fragments.
func TestParseFrontmatter_ExternalWithFragments(t *testing.T) {
	dir := t.TempDir()
	content := `---
external:
  - path: CODE_FROM_SPEC.md
    fragments:
      - description: frontmatter format
        lines: "10-20"
        hash: def456
      - description: tree structure
        lines: "30-50"
        hash: ghi789
  - path: README.md
---
`
	path := testWriteFile(t, dir, "node.md", content)

	fm, err := ParseFrontmatter(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.External) != 2 {
		t.Fatalf("External: got %d entries, want 2", len(fm.External))
	}

	// First external entry — CODE_FROM_SPEC.md with two fragments.
	first := fm.External[0]
	if first.Path != "CODE_FROM_SPEC.md" {
		t.Errorf("External[0].Path: got %q, want %q", first.Path, "CODE_FROM_SPEC.md")
	}
	if len(first.Fragments) != 2 {
		t.Fatalf("External[0].Fragments: got %d, want 2", len(first.Fragments))
	}
	// Fragment 0
	if first.Fragments[0].Description != "frontmatter format" {
		t.Errorf("Fragments[0].Description: got %q, want %q", first.Fragments[0].Description, "frontmatter format")
	}
	if first.Fragments[0].Lines != "10-20" {
		t.Errorf("Fragments[0].Lines: got %q, want %q", first.Fragments[0].Lines, "10-20")
	}
	if first.Fragments[0].Hash != "def456" {
		t.Errorf("Fragments[0].Hash: got %q, want %q", first.Fragments[0].Hash, "def456")
	}
	// Fragment 1
	if first.Fragments[1].Description != "tree structure" {
		t.Errorf("Fragments[1].Description: got %q, want %q", first.Fragments[1].Description, "tree structure")
	}
	if first.Fragments[1].Lines != "30-50" {
		t.Errorf("Fragments[1].Lines: got %q, want %q", first.Fragments[1].Lines, "30-50")
	}
	if first.Fragments[1].Hash != "ghi789" {
		t.Errorf("Fragments[1].Hash: got %q, want %q", first.Fragments[1].Hash, "ghi789")
	}

	// Second external entry — README.md with no fragments.
	second := fm.External[1]
	if second.Path != "README.md" {
		t.Errorf("External[1].Path: got %q, want %q", second.Path, "README.md")
	}
	if len(second.Fragments) != 0 {
		t.Errorf("External[1].Fragments: expected empty, got %v", second.Fragments)
	}
}

// TestParseFrontmatter_InputField verifies that the input field is parsed and
// all other fields remain nil/empty.
func TestParseFrontmatter_InputField(t *testing.T) {
	dir := t.TempDir()
	content := `---
input: ARTIFACT/golang/internal/frontmatter(frontmatter)
---
`
	path := testWriteFile(t, dir, "node.md", content)

	fm, err := ParseFrontmatter(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Input != "ARTIFACT/golang/internal/frontmatter(frontmatter)" {
		t.Errorf("Input: got %q, want %q", fm.Input, "ARTIFACT/golang/internal/frontmatter(frontmatter)")
	}
	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: expected empty, got %v", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("External: expected empty, got %v", fm.External)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("Outputs: expected empty, got %v", fm.Outputs)
	}
}

// TestParseFrontmatter_UnknownFields verifies that extra/future fields in the
// frontmatter do not cause an error and known fields are still parsed.
func TestParseFrontmatter_UnknownFields(t *testing.T) {
	dir := t.TempDir()
	content := `---
depends_on:
  - ROOT/other
some_future_field: hello
another: 42
---
`
	path := testWriteFile(t, dir, "node.md", content)

	fm, err := ParseFrontmatter(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Known fields are still parsed correctly.
	if len(fm.DependsOn) != 1 || fm.DependsOn[0] != "ROOT/other" {
		t.Errorf("DependsOn: got %v, want [ROOT/other]", fm.DependsOn)
	}
}

// TestParseFrontmatter_NoFrontmatter verifies that a file with no "---"
// delimiters returns an empty Frontmatter (not an error).
func TestParseFrontmatter_NoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	content := "Just some text without frontmatter.\n"
	path := testWriteFile(t, dir, "node.md", content)

	fm, err := ParseFrontmatter(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All fields must be nil/zero.
	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: expected empty, got %v", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("External: expected empty, got %v", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("Outputs: expected empty, got %v", fm.Outputs)
	}
}

// ---------------------------------------------------------------------------
// Edge-case tests
// ---------------------------------------------------------------------------

// TestParseFrontmatter_EmptyFrontmatter verifies that a file with an opening
// and closing "---" but nothing in between returns an empty Frontmatter.
func TestParseFrontmatter_EmptyFrontmatter(t *testing.T) {
	dir := t.TempDir()
	// File contains only "---\n---\n" — frontmatter delimiters with no content.
	content := "---\n---\n"
	path := testWriteFile(t, dir, "node.md", content)

	fm, err := ParseFrontmatter(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: expected empty, got %v", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("External: expected empty, got %v", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("Outputs: expected empty, got %v", fm.Outputs)
	}
}

// TestParseFrontmatter_NothingAfterFrontmatter verifies that a file with valid
// frontmatter and no body content is parsed without error.
func TestParseFrontmatter_NothingAfterFrontmatter(t *testing.T) {
	dir := t.TempDir()
	// File ends immediately after the closing delimiter.
	content := "---\ndepends_on:\n  - ROOT/other\n---\n"
	path := testWriteFile(t, dir, "node.md", content)

	fm, err := ParseFrontmatter(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 1 || fm.DependsOn[0] != "ROOT/other" {
		t.Errorf("DependsOn: got %v, want [ROOT/other]", fm.DependsOn)
	}
}

// ---------------------------------------------------------------------------
// Failure-case tests
// ---------------------------------------------------------------------------

// TestParseFrontmatter_FileDoesNotExist verifies that a non-existent file
// returns an error that wraps ErrRead.
func TestParseFrontmatter_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	// Intentionally point at a path that does not exist.
	nonExistent := filepath.Join(dir, "does_not_exist.md")

	_, err := ParseFrontmatter(nonExistent)
	if err == nil {
		t.Fatal("expected an error but got nil")
	}
	if !errors.Is(err, ErrRead) {
		t.Errorf("expected errors.Is(err, ErrRead) to be true; err = %v", err)
	}
}

// TestParseFrontmatter_MalformedYAML verifies that invalid YAML between the
// frontmatter delimiters returns an error that wraps ErrFrontmatterParse.
func TestParseFrontmatter_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	// The YAML value for depends_on is an unclosed list — invalid syntax.
	content := "---\ndepends_on: [invalid\n---\n"
	path := testWriteFile(t, dir, "node.md", content)

	_, err := ParseFrontmatter(path)
	if err == nil {
		t.Fatal("expected an error but got nil")
	}
	if !errors.Is(err, ErrFrontmatterParse) {
		t.Errorf("expected errors.Is(err, ErrFrontmatterParse) to be true; err = %v", err)
	}
}
