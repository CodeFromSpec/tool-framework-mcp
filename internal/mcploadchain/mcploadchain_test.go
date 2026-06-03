// code-from-spec: ROOT/golang/tests/mcp_tools/load_chain@3ArU-Xu_HPXw2HWspT9PECMe_ZM
package mcploadchain_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcploadchain"
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

func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	parts := strings.Split(path, "/")
	if len(parts) > 1 {
		dir := strings.Join(parts[:len(parts)-1], "/")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("testWriteFile MkdirAll: %v", err)
		}
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

func TestMCPLoadChain_SimpleLeafNode(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `# ROOT

# Public

Root public content line.
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: some/output.go
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
		t.Errorf("expected 27-char hash, got %q (len %d)", result.ChainHash, len(result.ChainHash))
	}

	if !strings.Contains(result.Context, "# Public") {
		t.Errorf("context missing # Public heading")
	}
	if !strings.Contains(result.Context, "Root public content line.") {
		t.Errorf("context missing ROOT public content")
	}
	if !strings.Contains(result.Context, "output: some/output.go") {
		t.Errorf("context missing output frontmatter field")
	}
	if !strings.Contains(result.Context, "Node a public content.") {
		t.Errorf("context missing ROOT/a public content")
	}
	if !strings.Contains(result.Context, "# Agent") {
		t.Errorf("context missing # Agent heading")
	}
	if !strings.Contains(result.Context, "Node a agent content.") {
		t.Errorf("context missing ROOT/a agent content")
	}

	rootIdx := strings.Index(result.Context, "Root public content line.")
	fmIdx := strings.Index(result.Context, "output: some/output.go")
	aPublicIdx := strings.Index(result.Context, "Node a public content.")
	aAgentIdx := strings.Index(result.Context, "Node a agent content.")

	if rootIdx >= fmIdx || fmIdx >= aPublicIdx || aPublicIdx >= aAgentIdx {
		t.Errorf("context content out of order")
	}

	if result.Input != nil {
		t.Errorf("expected no input, got %q", *result.Input)
	}
}

func TestMCPLoadChain_AncestorPublicContentIncluded(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `# ROOT

# Public

Root public content.
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `# ROOT/a

# Public

Node a public content.
`)
	testWriteFile(t, "code-from-spec/a/b/_node.md", `---
output: some/output.go
---
# ROOT/a/b
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "Root public content.") {
		t.Errorf("context missing ROOT public content")
	}
	if !strings.Contains(result.Context, "Node a public content.") {
		t.Errorf("context missing ROOT/a public content")
	}

	rootIdx := strings.Index(result.Context, "Root public content.")
	aIdx := strings.Index(result.Context, "Node a public content.")
	if rootIdx >= aIdx {
		t.Errorf("ROOT content should appear before ROOT/a content")
	}
}

func TestMCPLoadChain_AncestorWithoutPublicSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `# ROOT

# Name

Root name section only.
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: some/output.go
---
# ROOT/a

# Public

Node a public content.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(result.Context, "Root name section only.") {
		t.Errorf("context should not contain ROOT content when no public section")
	}
}

func TestMCPLoadChain_AncestorWithEmptyPublicSkipped(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `# ROOT

# Public
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: some/output.go
---
# ROOT/a

# Public

Node a public content.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pubCount := strings.Count(result.Context, "# Public")
	if pubCount > 1 {
		t.Errorf("expected only one # Public heading (ROOT/a), got %d", pubCount)
	}
}

func TestMCPLoadChain_DependencyWithoutQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `# ROOT
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: some/output.go
depends_on:
  - ROOT/b
---
# ROOT/a
`)
	testWriteFile(t, "code-from-spec/b/_node.md", `# ROOT/b

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

	if !strings.Contains(result.Context, "## Interface") {
		t.Errorf("context missing ## Interface subsection")
	}
	if !strings.Contains(result.Context, "Interface content.") {
		t.Errorf("context missing Interface content")
	}
	if !strings.Contains(result.Context, "## Constraints") {
		t.Errorf("context missing ## Constraints subsection")
	}
	if !strings.Contains(result.Context, "Constraints content.") {
		t.Errorf("context missing Constraints content")
	}
}

func TestMCPLoadChain_DependencyWithQualifier(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `# ROOT
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: some/output.go
depends_on:
  - ROOT/b(interface)
---
# ROOT/a
`)
	testWriteFile(t, "code-from-spec/b/_node.md", `# ROOT/b

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

	if !strings.Contains(result.Context, "## Interface") {
		t.Errorf("context missing ## Interface subsection")
	}
	if !strings.Contains(result.Context, "Interface content.") {
		t.Errorf("context missing Interface content")
	}
	if strings.Contains(result.Context, "## Constraints") {
		t.Errorf("context should not contain ## Constraints subsection")
	}
	if strings.Contains(result.Context, "Constraints content.") {
		t.Errorf("context should not contain Constraints content")
	}
}

func TestMCPLoadChain_ArtifactDependency(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `# ROOT
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: some/output.go
depends_on:
  - ARTIFACT/b
---
# ROOT/a
`)
	testWriteFile(t, "code-from-spec/b/_node.md", `---
output: out/b.go
---
# ROOT/b
`)
	testWriteFile(t, "out/b.go", `---
some: frontmatter
---
artifact body content here
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "artifact body content here") {
		t.Errorf("context missing artifact body content")
	}
	if strings.Contains(result.Context, "some: frontmatter") {
		t.Errorf("context should not contain artifact frontmatter")
	}
}

func TestMCPLoadChain_ExternalFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `# ROOT
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: some/output.go
external:
  - path: data/config.yaml
---
# ROOT/a
`)
	testWriteFile(t, "data/config.yaml", `key: value
another: setting
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "key: value") {
		t.Errorf("context missing external file content")
	}
	if !strings.Contains(result.Context, "another: setting") {
		t.Errorf("context missing external file content")
	}
}

func TestMCPLoadChain_TargetReducedFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `# ROOT
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: some/output.go
depends_on:
  - ROOT/b
---
# ROOT/a
`)
	testWriteFile(t, "code-from-spec/b/_node.md", `# ROOT/b

# Public

B public content.
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Context, "---") {
		t.Errorf("context missing frontmatter delimiters")
	}
	if !strings.Contains(result.Context, "output: some/output.go") {
		t.Errorf("context missing output field in frontmatter")
	}
	if strings.Contains(result.Context, "depends_on") {
		t.Errorf("context should not contain depends_on field in frontmatter")
	}
}

func TestMCPLoadChain_TargetAgentSectionIncluded(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `# ROOT
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: some/output.go
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

	if !strings.Contains(result.Context, "Node a public content.") {
		t.Errorf("context missing public content")
	}
	if !strings.Contains(result.Context, "# Agent") {
		t.Errorf("context missing # Agent heading")
	}
	if !strings.Contains(result.Context, "Node a agent content.") {
		t.Errorf("context missing agent content")
	}
}

func TestMCPLoadChain_TargetWithoutAgentSection(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `# ROOT
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: some/output.go
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
		t.Errorf("context missing public content")
	}
	if strings.Contains(result.Context, "# Agent") {
		t.Errorf("context should not contain # Agent heading")
	}
}

func TestMCPLoadChain_InputSeparatedFromContext(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `# ROOT
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: some/output.go
input: ARTIFACT/b
---
# ROOT/a
`)
	testWriteFile(t, "code-from-spec/b/_node.md", `---
output: out/data.json
---
# ROOT/b
`)
	testWriteFile(t, "out/data.json", `---
fm: yes
---
input body content here
`)

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Input == nil {
		t.Fatal("expected input to be present")
	}
	if !strings.Contains(*result.Input, "input body content here") {
		t.Errorf("input missing body content, got %q", *result.Input)
	}
	if strings.Contains(*result.Input, "fm: yes") {
		t.Errorf("input should not contain frontmatter")
	}
	if strings.Contains(result.Context, "input body content here") {
		t.Errorf("context should not contain input body content")
	}
}

func TestMCPLoadChain_NoInput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `# ROOT
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: some/output.go
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

func TestMCPLoadChain_HashIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `# ROOT

# Public

Root public content.
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: some/output.go
---
# ROOT/a

# Public

Node a public content.
`)

	result1, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}

	result2, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}

	if result1.ChainHash != result2.ChainHash {
		t.Errorf("hash not deterministic: %q != %q", result1.ChainHash, result2.ChainHash)
	}
}

func TestMCPLoadChain_InvalidLogicalName(t *testing.T) {
	_, err := mcploadchain.MCPLoadChain("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("expected ErrUnsupportedReference, got %v", err)
	}
}

func TestMCPLoadChain_NonexistentNodeFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := mcploadchain.MCPLoadChain("ROOT/nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestMCPLoadChain_NoOutputDeclared(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `# ROOT
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `# ROOT/a

# Public

Node a public content.
`)

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcploadchain.ErrNoOutput) {
		t.Errorf("expected ErrNoOutput, got %v", err)
	}
}

func TestMCPLoadChain_InvalidOutputPath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `# ROOT
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: "../../etc/passwd"
---
# ROOT/a
`)

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcploadchain.ErrInvalidOutputPath) {
		t.Errorf("expected ErrInvalidOutputPath, got %v", err)
	}
}

func TestMCPLoadChain_UnresolvableDependency(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", `# ROOT
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: some/output.go
depends_on:
  - ROOT/missing
---
# ROOT/a
`)

	_, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err == nil {
		t.Fatal("expected error for unresolvable dependency, got nil")
	}
}
