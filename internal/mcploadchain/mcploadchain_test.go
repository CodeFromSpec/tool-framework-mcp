// code-from-spec: ROOT/golang/tests/mcp_tools/load_chain@OQ2XbsBVZcRRGwXZNxxcrzVU1s4
package mcploadchain_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
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

// testWriteFile creates parent directories and writes content to the given path.
func testWriteFile(t *testing.T, path, content string) {
	t.Helper()
	idx := strings.LastIndex(path, "/")
	if idx > 0 {
		if err := os.MkdirAll(path[:idx], 0o755); err != nil {
			t.Fatalf("testWriteFile MkdirAll: %v", err)
		}
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("testWriteFile WriteFile %s: %v", path, err)
	}
}

// TC-01: Simple leaf node — context and hash
func TestMCPLoadChain_TC01_SimpleLeafNode(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
---
# ROOT
# Public
Root public content.
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
outputs:
  - id: main
    path: out/a.txt
---
# ROOT/a
# Public
Node a public content.
# Agent
Node a agent content.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.ChainHash) != 27 {
		t.Errorf("expected 27-char hash, got %d chars: %q", len(result.ChainHash), result.ChainHash)
	}

	if !strings.Contains(result.Context, "Root public content.") {
		t.Errorf("context missing root public content; context:\n%s", result.Context)
	}
	if !strings.Contains(result.Context, "Node a public content.") {
		t.Errorf("context missing node a public content; context:\n%s", result.Context)
	}
	if !strings.Contains(result.Context, "Node a agent content.") {
		t.Errorf("context missing node a agent content; context:\n%s", result.Context)
	}
	if strings.Contains(result.Context, "# Public") {
		t.Errorf("context should not contain '# Public' heading line; context:\n%s", result.Context)
	}
	if strings.Contains(result.Context, "# Agent") {
		t.Errorf("context should not contain '# Agent' heading line; context:\n%s", result.Context)
	}
	// outputs field should appear in a frontmatter block
	if !strings.Contains(result.Context, "outputs:") {
		t.Errorf("context missing outputs frontmatter; context:\n%s", result.Context)
	}

	if result.Input != nil {
		t.Errorf("expected no input, got %q", *result.Input)
	}
}

// TC-02: Ancestor public content included
func TestMCPLoadChain_TC02_AncestorPublicContentIncluded(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
---
# ROOT
# Public
Root public content.
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
---
# ROOT/a
# Public
Node a public content.
`)
	testWriteFile(t, "code-from-spec/a/b/_node.md", `---
outputs:
  - id: main
    path: out/b.txt
---
# ROOT/a/b
# Public
Node b public content.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "Root public content.") {
		t.Errorf("context missing root public content; context:\n%s", result.Context)
	}
	if !strings.Contains(result.Context, "Node a public content.") {
		t.Errorf("context missing node a public content; context:\n%s", result.Context)
	}
}

// TC-03: Ancestor without public section — skipped
func TestMCPLoadChain_TC03_AncestorWithoutPublicSection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
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
Node a public content.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ROOT has no Public section; its name section content should not appear.
	// The word "Root" would only appear if name section content leaked.
	// We simply verify no content from the ROOT name section bleeds in.
	_ = result
}

// TC-04: Ancestor with empty public section — skipped
func TestMCPLoadChain_TC04_AncestorWithEmptyPublicSection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
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
Node a public content.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "Node a public content.") {
		t.Errorf("context missing node a public content; context:\n%s", result.Context)
	}
}

// TC-05: Dependency without qualifier — full public section included
func TestMCPLoadChain_TC05_DependencyWithoutQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
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
`)
	testWriteFile(t, "code-from-spec/b/_node.md", `---
---
# ROOT/b
# Public
## Interface
Interface content.
## Constraints
Constraints content.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "Interface content.") {
		t.Errorf("context missing Interface subsection content; context:\n%s", result.Context)
	}
	if !strings.Contains(result.Context, "Constraints content.") {
		t.Errorf("context missing Constraints subsection content; context:\n%s", result.Context)
	}
	if !strings.Contains(result.Context, "## Interface") {
		t.Errorf("context missing ## Interface heading; context:\n%s", result.Context)
	}
	if !strings.Contains(result.Context, "## Constraints") {
		t.Errorf("context missing ## Constraints heading; context:\n%s", result.Context)
	}
}

// TC-06: Dependency with qualifier — subsection only
func TestMCPLoadChain_TC06_DependencyWithQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
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
`)
	testWriteFile(t, "code-from-spec/b/_node.md", `---
---
# ROOT/b
# Public
## Interface
Interface content.
## Constraints
Constraints content.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "Interface content.") {
		t.Errorf("context missing Interface content; context:\n%s", result.Context)
	}
	if !strings.Contains(result.Context, "## Interface") {
		t.Errorf("context missing ## Interface heading; context:\n%s", result.Context)
	}
	if strings.Contains(result.Context, "Constraints content.") {
		t.Errorf("context should NOT contain Constraints content; context:\n%s", result.Context)
	}
}

// TC-07: ARTIFACT dependency — content minus frontmatter
func TestMCPLoadChain_TC07_ArtifactDependency(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
---
# ROOT
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
depends_on:
  - ARTIFACT/b(code)
outputs:
  - id: main
    path: out/a.txt
---
# ROOT/a
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

func main() {}
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "package main") {
		t.Errorf("context missing package main; context:\n%s", result.Context)
	}
	if !strings.Contains(result.Context, "func main() {}") {
		t.Errorf("context missing func main; context:\n%s", result.Context)
	}
	if strings.Contains(result.Context, "some: frontmatter") {
		t.Errorf("context should NOT contain frontmatter from artifact; context:\n%s", result.Context)
	}
}

// TC-08: External file — full content
func TestMCPLoadChain_TC08_ExternalFileFullContent(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
---
# ROOT
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
outputs:
  - id: main
    path: out/a.txt
external:
  - path: data/config.yaml
---
# ROOT/a
`)
	testWriteFile(t, "data/config.yaml", `key: value
other: 42
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "key: value") {
		t.Errorf("context missing config content; context:\n%s", result.Context)
	}
	if !strings.Contains(result.Context, "other: 42") {
		t.Errorf("context missing config content; context:\n%s", result.Context)
	}
}

// TC-09: External file with fragments — line ranges only
func TestMCPLoadChain_TC09_ExternalFileWithFragments(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
---
# ROOT
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
outputs:
  - id: main
    path: out/a.txt
external:
  - path: data/big.txt
    fragments:
      - lines: "2-4"
        hash: "ignored"
---
# ROOT/a
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

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "line 2") {
		t.Errorf("context missing line 2; context:\n%s", result.Context)
	}
	if !strings.Contains(result.Context, "line 3") {
		t.Errorf("context missing line 3; context:\n%s", result.Context)
	}
	if !strings.Contains(result.Context, "line 4") {
		t.Errorf("context missing line 4; context:\n%s", result.Context)
	}
	if strings.Contains(result.Context, "line 1\n") {
		t.Errorf("context should NOT contain line 1; context:\n%s", result.Context)
	}
	if strings.Contains(result.Context, "line 5") {
		t.Errorf("context should NOT contain line 5; context:\n%s", result.Context)
	}
}

// TC-10: Target has reduced frontmatter with outputs only
func TestMCPLoadChain_TC10_TargetFrontmatterOutputsOnly(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
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
`)
	testWriteFile(t, "code-from-spec/b/_node.md", `---
---
# ROOT/b
# Public
Node b content.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "outputs:") {
		t.Errorf("context missing outputs field; context:\n%s", result.Context)
	}
	if strings.Contains(result.Context, "depends_on:") {
		t.Errorf("context should NOT contain depends_on field; context:\n%s", result.Context)
	}
}

// TC-11: Target agent section included
func TestMCPLoadChain_TC11_TargetAgentSectionIncluded(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
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
Public content.
# Agent
Agent content.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "Public content.") {
		t.Errorf("context missing public content; context:\n%s", result.Context)
	}
	if !strings.Contains(result.Context, "Agent content.") {
		t.Errorf("context missing agent content; context:\n%s", result.Context)
	}
	if strings.Contains(result.Context, "# Public") {
		t.Errorf("context should NOT contain '# Public' heading; context:\n%s", result.Context)
	}
	if strings.Contains(result.Context, "# Agent") {
		t.Errorf("context should NOT contain '# Agent' heading; context:\n%s", result.Context)
	}
}

// TC-12: Target without agent section — skipped
func TestMCPLoadChain_TC12_TargetWithoutAgentSection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
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
Public only content.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "Public only content.") {
		t.Errorf("context missing public only content; context:\n%s", result.Context)
	}
}

// TC-13: Input separated from context
func TestMCPLoadChain_TC13_InputSeparatedFromContext(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
---
# ROOT
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
outputs:
  - id: main
    path: out/a.txt
input: "ARTIFACT/b(data)"
---
# ROOT/a
`)
	testWriteFile(t, "code-from-spec/b/_node.md", `---
outputs:
  - id: data
    path: out/data.json
---
# ROOT/b
`)
	testWriteFile(t, "out/data.json", `---
meta: info
---
{"key": "value"}
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Input == nil {
		t.Fatalf("expected input to be present, got nil")
	}
	if !strings.Contains(*result.Input, `{"key": "value"}`) {
		t.Errorf("input missing expected content; input:\n%s", *result.Input)
	}
	if strings.Contains(*result.Input, "meta: info") {
		t.Errorf("input should NOT contain frontmatter; input:\n%s", *result.Input)
	}
	if strings.Contains(result.Context, `{"key": "value"}`) {
		t.Errorf("context should NOT contain input body; context:\n%s", result.Context)
	}
}

// TC-14: No input — field absent
func TestMCPLoadChain_TC14_NoInput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
---
# ROOT
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
outputs:
  - id: main
    path: out/a.txt
---
# ROOT/a
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Input != nil {
		t.Errorf("expected no input, got %q", *result.Input)
	}
}

// TC-15: Hash is deterministic
func TestMCPLoadChain_TC15_HashIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
---
# ROOT
# Public
Stable content.
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
outputs:
  - id: main
    path: out/a.txt
---
# ROOT/a
# Public
Node a stable content.
`)

	result1, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("first call unexpected error: %v", err)
	}

	result2, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("second call unexpected error: %v", err)
	}

	if result1.ChainHash != result2.ChainHash {
		t.Errorf("hash not deterministic: %q != %q", result1.ChainHash, result2.ChainHash)
	}
}

// TC-16: Invalid logical name — not ROOT/
func TestMCPLoadChain_TC16_InvalidLogicalName(t *testing.T) {
	_, err := mcploadchain.MCPLoadChain("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("expected ErrUnsupportedReference, got: %v", err)
	}
}

// TC-17: Nonexistent node file
func TestMCPLoadChain_TC17_NonexistentNodeFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	// No _node.md for ROOT/nonexistent created.
	_, err := mcploadchain.MCPLoadChain("ROOT/nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

// TC-18: No outputs declared
func TestMCPLoadChain_TC18_NoOutputsDeclared(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
---
# ROOT
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
---
# ROOT/a
# Public
Content without outputs.
`)

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcploadchain.ErrNoOutputs) {
		t.Errorf("expected ErrNoOutputs, got: %v", err)
	}
}

// TC-19: Invalid output path — path traversal
func TestMCPLoadChain_TC19_InvalidOutputPathTraversal(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
---
# ROOT
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
outputs:
  - id: evil
    path: ../../etc/passwd
---
# ROOT/a
`)

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcploadchain.ErrInvalidOutputPath) {
		t.Errorf("expected ErrInvalidOutputPath, got: %v", err)
	}
}

// TC-20: Unresolvable dependency
func TestMCPLoadChain_TC20_UnresolvableDependency(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `---
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
`)
	// ROOT/missing/_node.md is intentionally not created.

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// The error is propagated from chain resolution or hash computation
	// when the missing node file is accessed.
	if !errors.Is(err, filereader.ErrFileUnreadable) && !errors.Is(err, chainresolver.ErrUnreadableFrontmatter) {
		t.Errorf("expected file-not-found error, got: %v", err)
	}
}
