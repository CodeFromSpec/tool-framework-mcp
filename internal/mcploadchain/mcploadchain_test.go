// code-from-spec: ROOT/golang/tests/mcp_tools/load_chain@Kje4Mfor4MXVbMM_WfYvcOybFek
package mcploadchain_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcploadchain"
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

// testWriteFile creates the file at path (relative to cwd), making parent directories as needed.
func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	parts := strings.Split(path, "/")
	if len(parts) > 1 {
		dir := strings.Join(parts[:len(parts)-1], "/")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("testWriteFile MkdirAll(%q): %v", dir, err)
		}
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile WriteFile(%q): %v", path, err)
	}
}

// TC-01: Simple leaf node — context and hash
func TestMCPLoadChain_TC01_SimpleLeafNode(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
name: ROOT
---
# ROOT
# Public
Root public content line one.
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
outputs:
  - id: main
    path: out/a.txt
---
# ROOT/a
# Public
Leaf A public content.

# Agent
Leaf A agent guidance.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("MCPLoadChain: unexpected error: %v", err)
	}

	if len(result.ChainHash) != 27 {
		t.Errorf("ChainHash length = %d, want 27", len(result.ChainHash))
	}

	if !strings.Contains(result.Context, "Root public content line one.") {
		t.Errorf("context missing root public content")
	}

	if !strings.Contains(result.Context, "# Public") {
		t.Errorf("context missing # Public heading")
	}

	if !strings.Contains(result.Context, "outputs:") {
		t.Errorf("context missing outputs field in frontmatter")
	}

	if !strings.Contains(result.Context, "Leaf A public content.") {
		t.Errorf("context missing leaf A public content")
	}

	if !strings.Contains(result.Context, "# Agent") {
		t.Errorf("context missing # Agent heading")
	}

	if !strings.Contains(result.Context, "Leaf A agent guidance.") {
		t.Errorf("context missing leaf A agent guidance")
	}

	if result.Input != nil {
		t.Errorf("input should be absent, got %q", *result.Input)
	}
}

// TC-02: Ancestor public content included
func TestMCPLoadChain_TC02_AncestorPublicContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
name: ROOT
---
# ROOT
# Public
Root public content.
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
name: a
---
# ROOT/a
# Public
A public content.
`)

	testWriteFile(t, "code-from-spec/a/b/_node.md", `---
outputs:
  - id: main
    path: out/b.txt
---
# ROOT/a/b
# Public
B public content.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a/b")
	if err != nil {
		t.Fatalf("MCPLoadChain: unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "Root public content.") {
		t.Errorf("context missing root public content")
	}

	if !strings.Contains(result.Context, "A public content.") {
		t.Errorf("context missing A public content")
	}

	rootIdx := strings.Index(result.Context, "Root public content.")
	aIdx := strings.Index(result.Context, "A public content.")
	if rootIdx >= aIdx {
		t.Errorf("ROOT public content should appear before ROOT/a public content in context")
	}
}

// TC-03: Ancestor without public section skipped
func TestMCPLoadChain_TC03_AncestorWithoutPublicSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
name: ROOT
---
# ROOT
Root name section only.
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
outputs:
  - id: main
    path: out/a.txt
---
# ROOT/a
# Public
A public content.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("MCPLoadChain: unexpected error: %v", err)
	}

	if strings.Contains(result.Context, "Root name section only.") {
		t.Errorf("context should not contain ROOT name section content")
	}

	if !strings.Contains(result.Context, "A public content.") {
		t.Errorf("context missing A public content")
	}
}

// TC-04: Ancestor with empty public section skipped
func TestMCPLoadChain_TC04_AncestorWithEmptyPublicSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
name: ROOT
---
# ROOT
# Public
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
outputs:
  - id: main
    path: out/a.txt
---
# ROOT/a
# Public
A public content.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("MCPLoadChain: unexpected error: %v", err)
	}

	// Count occurrences of "# Public" - should only appear once (for ROOT/a)
	// ROOT's empty public section should be skipped
	publicHeadings := strings.Count(result.Context, "# Public\n")
	if publicHeadings > 1 {
		t.Errorf("context contains %d # Public headings, expected only ROOT/a's", publicHeadings)
	}

	if !strings.Contains(result.Context, "A public content.") {
		t.Errorf("context missing A public content")
	}
}

// TC-05: Dependency without qualifier — public included
func TestMCPLoadChain_TC05_DependencyWithoutQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
name: ROOT
---
# ROOT
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
depends_on:
  - ROOT/b
outputs:
  - id: main
    path: out/a.txt
---
# ROOT/a
# Public
A public content.
`)

	testWriteFile(t, "code-from-spec/b/_node.md", `---
name: b
---
# ROOT/b
# Public
B intro content.

## Interface
B interface details.

## Constraints
B constraint details.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("MCPLoadChain: unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "B intro content.") {
		t.Errorf("context missing B intro content")
	}

	if !strings.Contains(result.Context, "## Interface") {
		t.Errorf("context missing ## Interface heading")
	}

	if !strings.Contains(result.Context, "B interface details.") {
		t.Errorf("context missing B interface details")
	}

	if !strings.Contains(result.Context, "## Constraints") {
		t.Errorf("context missing ## Constraints heading")
	}

	if !strings.Contains(result.Context, "B constraint details.") {
		t.Errorf("context missing B constraint details")
	}
}

// TC-06: Dependency with qualifier — subsection only
func TestMCPLoadChain_TC06_DependencyWithQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
name: ROOT
---
# ROOT
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
depends_on:
  - ROOT/b(interface)
outputs:
  - id: main
    path: out/a.txt
---
# ROOT/a
# Public
A public content.
`)

	testWriteFile(t, "code-from-spec/b/_node.md", `---
name: b
---
# ROOT/b
# Public
B intro content.

## Interface
B interface details.

## Constraints
B constraint details.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("MCPLoadChain: unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "## Interface") {
		t.Errorf("context missing ## Interface heading")
	}

	if !strings.Contains(result.Context, "B interface details.") {
		t.Errorf("context missing B interface details")
	}

	if strings.Contains(result.Context, "## Constraints") {
		t.Errorf("context should not contain ## Constraints heading")
	}

	if strings.Contains(result.Context, "B constraint details.") {
		t.Errorf("context should not contain B constraint details")
	}

	if strings.Contains(result.Context, "B intro content.") {
		t.Errorf("context should not contain B intro content (only qualified subsection)")
	}
}

// TC-07: ARTIFACT dependency — content minus frontmatter
func TestMCPLoadChain_TC07_ArtifactDependency(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
name: ROOT
---
# ROOT
`)

	testWriteFile(t, "code-from-spec/b/_node.md", `---
outputs:
  - id: code
    path: out/b.go
---
# ROOT/b
`)

	testWriteFile(t, "out/b.go", `---
some: frontmatter
---
package main

// Body content of b.go
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
depends_on:
  - ARTIFACT/b(code)
outputs:
  - id: main
    path: out/a.txt
---
# ROOT/a
# Public
A public content.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("MCPLoadChain: unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "package main") {
		t.Errorf("context missing 'package main'")
	}

	if !strings.Contains(result.Context, "// Body content of b.go") {
		t.Errorf("context missing artifact body content")
	}

	if strings.Contains(result.Context, "some: frontmatter") {
		t.Errorf("context should not contain artifact frontmatter content")
	}
}

// TC-08: External file — full content
func TestMCPLoadChain_TC08_ExternalFileFullContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
name: ROOT
---
# ROOT
`)

	testWriteFile(t, "data/config.yaml", `key: value
setting: enabled
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
external:
  - path: data/config.yaml
outputs:
  - id: main
    path: out/a.txt
---
# ROOT/a
# Public
A public content.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("MCPLoadChain: unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "key: value") {
		t.Errorf("context missing 'key: value' from external file")
	}

	if !strings.Contains(result.Context, "setting: enabled") {
		t.Errorf("context missing 'setting: enabled' from external file")
	}
}

// TC-09: External file with fragments — line ranges only
func TestMCPLoadChain_TC09_ExternalFileWithFragments(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
name: ROOT
---
# ROOT
`)

	testWriteFile(t, "data/big.txt", `line 1
line 2
line 3
line 4
line 5
line 6
line 7
line 8
line 9
line 10
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
external:
  - path: data/big.txt
    fragments:
      - lines: "2-4"
        hash: ignored
outputs:
  - id: main
    path: out/a.txt
---
# ROOT/a
# Public
A public content.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("MCPLoadChain: unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "line 2") {
		t.Errorf("context missing 'line 2'")
	}

	if !strings.Contains(result.Context, "line 3") {
		t.Errorf("context missing 'line 3'")
	}

	if !strings.Contains(result.Context, "line 4") {
		t.Errorf("context missing 'line 4'")
	}

	for _, excluded := range []string{"line 1", "line 5", "line 6", "line 7", "line 8", "line 9", "line 10"} {
		if strings.Contains(result.Context, excluded+"\n") || strings.HasSuffix(result.Context, excluded) {
			t.Errorf("context should not contain %q", excluded)
		}
	}
}

// TC-10: Target has reduced frontmatter with outputs only
func TestMCPLoadChain_TC10_TargetReducedFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
name: ROOT
---
# ROOT
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
depends_on:
  - ROOT/b
outputs:
  - id: main
    path: out/a.txt
---
# ROOT/a
# Public
A public content.
`)

	testWriteFile(t, "code-from-spec/b/_node.md", `---
name: b
---
# ROOT/b
# Public
B public content.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("MCPLoadChain: unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "outputs:") {
		t.Errorf("context missing outputs field in target frontmatter")
	}

	if strings.Contains(result.Context, "depends_on:") {
		t.Errorf("context should not contain depends_on field in target frontmatter")
	}
}

// TC-11: Target agent section included
func TestMCPLoadChain_TC11_TargetAgentSection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
name: ROOT
---
# ROOT
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
outputs:
  - id: main
    path: out/a.txt
---
# ROOT/a
# Public
A public content.

# Agent
A agent guidance.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("MCPLoadChain: unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "# Public") {
		t.Errorf("context missing # Public heading")
	}

	if !strings.Contains(result.Context, "A public content.") {
		t.Errorf("context missing A public content")
	}

	if !strings.Contains(result.Context, "# Agent") {
		t.Errorf("context missing # Agent heading")
	}

	if !strings.Contains(result.Context, "A agent guidance.") {
		t.Errorf("context missing A agent guidance")
	}
}

// TC-12: Target without agent section — skipped
func TestMCPLoadChain_TC12_TargetWithoutAgentSection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
name: ROOT
---
# ROOT
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
outputs:
  - id: main
    path: out/a.txt
---
# ROOT/a
# Public
A public content.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("MCPLoadChain: unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "A public content.") {
		t.Errorf("context missing A public content")
	}

	if strings.Contains(result.Context, "# Agent") {
		t.Errorf("context should not contain # Agent heading when none declared")
	}
}

// TC-13: Input separated from context
func TestMCPLoadChain_TC13_InputSeparatedFromContext(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
name: ROOT
---
# ROOT
`)

	testWriteFile(t, "code-from-spec/b/_node.md", `---
outputs:
  - id: data
    path: out/data.json
---
# ROOT/b
`)

	testWriteFile(t, "out/data.json", `---
artifact: frontmatter
---
{"key": "value", "count": 42}
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
input: ARTIFACT/b(data)
outputs:
  - id: main
    path: out/a.txt
---
# ROOT/a
# Public
A public content.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("MCPLoadChain: unexpected error: %v", err)
	}

	if result.Input == nil {
		t.Fatalf("result.Input should not be nil")
	}

	if !strings.Contains(*result.Input, `{"key": "value", "count": 42}`) {
		t.Errorf("input missing expected JSON content, got: %q", *result.Input)
	}

	if strings.Contains(*result.Input, "artifact: frontmatter") {
		t.Errorf("input should not contain frontmatter content")
	}

	if strings.Contains(result.Context, `{"key": "value", "count": 42}`) {
		t.Errorf("context should not contain input file body content")
	}
}

// TC-14: No input — field absent
func TestMCPLoadChain_TC14_NoInput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
name: ROOT
---
# ROOT
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
outputs:
  - id: main
    path: out/a.txt
---
# ROOT/a
# Public
A public content.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("MCPLoadChain: unexpected error: %v", err)
	}

	if result.Input != nil {
		t.Errorf("result.Input should be nil, got %q", *result.Input)
	}
}

// TC-15: Hash is deterministic
func TestMCPLoadChain_TC15_HashIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
name: ROOT
---
# ROOT
# Public
Deterministic root content.
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
outputs:
  - id: main
    path: out/a.txt
---
# ROOT/a
# Public
Deterministic A content.
`)

	result1, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("first MCPLoadChain: unexpected error: %v", err)
	}

	result2, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("second MCPLoadChain: unexpected error: %v", err)
	}

	if result1.ChainHash != result2.ChainHash {
		t.Errorf("hash not deterministic: %q != %q", result1.ChainHash, result2.ChainHash)
	}
}

// TC-E01: Invalid logical name — not ROOT/
func TestMCPLoadChain_TCE01_InvalidLogicalName(t *testing.T) {
	_, err := mcploadchain.MCPLoadChain("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("expected ErrUnsupportedReference, got: %v", err)
	}
}

// TC-E02: Nonexistent node file
func TestMCPLoadChain_TCE02_NonexistentNodeFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// No _node.md created for ROOT/nonexistent.
	_, err := mcploadchain.MCPLoadChain("ROOT/nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, frontmatter.ErrFileUnreadable) {
		t.Errorf("expected frontmatter.ErrFileUnreadable, got: %v", err)
	}
}

// TC-E03: No outputs declared
func TestMCPLoadChain_TCE03_NoOutputsDeclared(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
name: ROOT
---
# ROOT
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
name: a
---
# ROOT/a
# Public
A public content.
`)

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, mcploadchain.ErrNoOutputs) {
		t.Errorf("expected ErrNoOutputs, got: %v", err)
	}
}

// TC-E04: Invalid output path — traversal
func TestMCPLoadChain_TCE04_InvalidOutputPath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
name: ROOT
---
# ROOT
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
outputs:
  - id: main
    path: ../../etc/passwd
---
# ROOT/a
# Public
A public content.
`)

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, mcploadchain.ErrInvalidOutputPath) {
		t.Errorf("expected ErrInvalidOutputPath, got: %v", err)
	}
}

// TC-E05: Unresolvable dependency
func TestMCPLoadChain_TCE05_UnresolvableDependency(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
name: ROOT
---
# ROOT
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
depends_on:
  - ROOT/missing
outputs:
  - id: main
    path: out/a.txt
---
# ROOT/a
# Public
A public content.
`)

	// Do not create code-from-spec/missing/_node.md

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// The error propagates from chain processing because the missing file cannot be read.
	// It may wrap chainresolver.ErrUnreadableFrontmatter or a filereader error.
	if !errors.Is(err, chainresolver.ErrUnreadableFrontmatter) &&
		!errors.Is(err, frontmatter.ErrFileUnreadable) {
		t.Errorf("expected an error related to unresolvable dependency, got: %v", err)
	}
}
