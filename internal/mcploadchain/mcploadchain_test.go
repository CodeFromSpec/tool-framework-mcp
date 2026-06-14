// code-from-spec: ROOT/golang/tests/mcp_tools/load_chain@wtnkQz3h1lnV_RXpt8w6QN1-GH8
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
			t.Fatalf("testWriteFile mkdir %s: %v", dir, err)
		}
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile %s: %v", path, err)
	}
}

func TestMCPLoadChain_TC01_SimpleLeafNode(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC

# Public

## Context

Root context line
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: out/a.txt
---
# SPEC/a

# Public

## Interface

Interface line

# Agent

Agent content line
`)

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.SplitN(result, "\n", 2)
	if len(lines) < 1 {
		t.Fatal("result is empty")
	}
	firstLine := lines[0]
	if !strings.HasPrefix(firstLine, "chain_hash: ") {
		t.Errorf("first line does not start with 'chain_hash: ': %q", firstLine)
	}
	hash := strings.TrimPrefix(firstLine, "chain_hash: ")
	if len(hash) != 27 {
		t.Errorf("hash length = %d, want 27", len(hash))
	}

	if !strings.Contains(result, "--- context ---") {
		t.Error("result missing '--- context ---'")
	}
	if strings.Contains(result, "# Public") {
		t.Error("result should not contain '# Public' heading")
	}
	if !strings.Contains(result, "## Context") {
		t.Error("result missing '## Context' heading from SPEC")
	}
	if !strings.Contains(result, "Root context line") {
		t.Error("result missing root context content")
	}
	if !strings.Contains(result, "## Interface") {
		t.Error("result missing '## Interface' heading from SPEC/a")
	}
	if !strings.Contains(result, "Interface line") {
		t.Error("result missing interface content")
	}
	if !strings.Contains(result, "# Agent") {
		t.Error("result missing '# Agent' heading")
	}
	if !strings.Contains(result, "Agent content line") {
		t.Error("result missing agent content")
	}
	if strings.Contains(result, "--- input ---") {
		t.Error("result should not contain '--- input ---'")
	}
	if strings.Contains(result, "--- existing artifact ---") {
		t.Error("result should not contain '--- existing artifact ---'")
	}

	contextIdx := strings.Index(result, "--- context ---")
	contextSection := result[contextIdx:]
	if strings.Contains(contextSection, "depends_on") {
		t.Error("context section should not contain 'depends_on'")
	}
	if !strings.Contains(contextSection, "output") {
		t.Error("context section should contain 'output' field")
	}
}

func TestMCPLoadChain_TC02_AncestorPublicContentIncluded(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC

# Public

## Overview

Root overview line
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `# SPEC/a

# Public

## Description

A description line
`)

	testWriteFile(t, "code-from-spec/a/b/_node.md", `---
output: out/b.txt
---
# SPEC/a/b

# Public

## Details

B details
`)

	result, err := mcploadchain.MCPLoadChain("SPEC/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "## Overview") {
		t.Error("result missing SPEC '## Overview'")
	}
	if !strings.Contains(result, "Root overview line") {
		t.Error("result missing root overview content")
	}
	if !strings.Contains(result, "## Description") {
		t.Error("result missing SPEC/a '## Description'")
	}
	if !strings.Contains(result, "A description line") {
		t.Error("result missing SPEC/a description content")
	}
}

func TestMCPLoadChain_TC03_AncestorWithoutPublicSectionSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC

This is name section content only
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: out/a.txt
---
# SPEC/a

# Public

## Summary

Summary content
`)

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(result, "This is name section content only") {
		t.Error("result should not contain SPEC name section content")
	}
	if !strings.Contains(result, "## Summary") {
		t.Error("result missing '## Summary'")
	}
	if !strings.Contains(result, "Summary content") {
		t.Error("result missing summary content")
	}
}

func TestMCPLoadChain_TC04_AncestorWithEmptyPublicSectionSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC

# Public
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: out/a.txt
---
# SPEC/a

# Public

## Summary

Summary content
`)

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "## Summary") {
		t.Error("result missing '## Summary'")
	}
	if !strings.Contains(result, "Summary content") {
		t.Error("result missing summary content")
	}
}

func TestMCPLoadChain_TC05_DependencyWithoutQualifier(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: out/a.txt
depends_on:
  - SPEC/b
---
# SPEC/a
`)

	testWriteFile(t, "code-from-spec/b/_node.md", `# SPEC/b

# Public

## Interface

B interface content

## Constraints

B constraints content
`)

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "## Interface") {
		t.Error("result missing '## Interface' from SPEC/b")
	}
	if !strings.Contains(result, "B interface content") {
		t.Error("result missing B interface content")
	}
	if !strings.Contains(result, "## Constraints") {
		t.Error("result missing '## Constraints' from SPEC/b")
	}
	if !strings.Contains(result, "B constraints content") {
		t.Error("result missing B constraints content")
	}
}

func TestMCPLoadChain_TC06_DependencyWithQualifier(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: out/a.txt
depends_on:
  - SPEC/b(interface)
---
# SPEC/a
`)

	testWriteFile(t, "code-from-spec/b/_node.md", `# SPEC/b

# Public

## Interface

B interface content

## Constraints

B constraints content
`)

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "## Interface") {
		t.Error("result missing '## Interface' from SPEC/b")
	}
	if !strings.Contains(result, "B interface content") {
		t.Error("result missing B interface content")
	}
	if strings.Contains(result, "## Constraints") {
		t.Error("result should not contain '## Constraints' when qualifier filters to 'interface' subsection only")
	}
	if strings.Contains(result, "B constraints content") {
		t.Error("result should not contain B constraints content")
	}
}

func TestMCPLoadChain_TC07_ArtifactDependencyTagLineRemoved(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: out/a.txt
depends_on:
  - ARTIFACT/b
---
# SPEC/a
`)

	testWriteFile(t, "code-from-spec/b/_node.md", `---
output: out/b.go
---
# SPEC/b
`)

	testWriteFile(t, "out/b.go", `// code-from-spec: SPEC/b@AAAAAAAAAAAAAAAAAAAAAAAAAAA
package main

// body content line
`)

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "body content line") {
		t.Error("result missing artifact body content")
	}
	if strings.Contains(result, "code-from-spec: SPEC/b@") {
		t.Error("result should not contain the artifact tag line")
	}
}

func TestMCPLoadChain_TC08_ExternalDependencyFullContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: out/a.txt
depends_on:
  - EXTERNAL/data/config.yaml
---
# SPEC/a
`)

	testWriteFile(t, "data/config.yaml", `key: value
another: line
`)

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "key: value") {
		t.Error("result missing external file content")
	}
	if !strings.Contains(result, "another: line") {
		t.Error("result missing second line of external file")
	}
}

func TestMCPLoadChain_TC09_TargetFrontmatterOutputOnly(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: out/a.txt
depends_on:
  - SPEC/b
---
# SPEC/a
`)

	testWriteFile(t, "code-from-spec/b/_node.md", `# SPEC/b

# Public

## Info

Info content
`)

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(result, "depends_on") {
		t.Error("result should not contain 'depends_on' in the frontmatter block")
	}
	if !strings.Contains(result, "output") {
		t.Error("result should contain 'output' in the frontmatter block")
	}
}

func TestMCPLoadChain_TC10_TargetAgentSectionIncluded(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: out/a.txt
---
# SPEC/a

# Public

## Interface

Interface content

# Agent

Agent guidance here
`)

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "## Interface") {
		t.Error("result missing '## Interface'")
	}
	if !strings.Contains(result, "Interface content") {
		t.Error("result missing interface content")
	}
	if !strings.Contains(result, "# Agent") {
		t.Error("result missing '# Agent' heading")
	}
	if !strings.Contains(result, "Agent guidance here") {
		t.Error("result missing agent content")
	}
	if strings.Contains(result, "# Public") {
		t.Error("result should not contain '# Public' heading")
	}
}

func TestMCPLoadChain_TC11_TargetWithoutAgentSectionSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: out/a.txt
---
# SPEC/a

# Public

## Interface

Interface content
`)

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "## Interface") {
		t.Error("result missing '## Interface'")
	}
	if strings.Contains(result, "# Agent") {
		t.Error("result should not contain '# Agent' heading when no agent section exists")
	}
}

func TestMCPLoadChain_TC12_InputPresentInSeparateSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: out/a.txt
input: ARTIFACT/b
---
# SPEC/a
`)

	testWriteFile(t, "code-from-spec/b/_node.md", `---
output: out/data.json
---
# SPEC/b
`)

	testWriteFile(t, "out/data.json", `// code-from-spec: SPEC/b@AAAAAAAAAAAAAAAAAAAAAAAAAAA
{"key": "value"}
more json content
`)

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "--- input ---") {
		t.Error("result missing '--- input ---'")
	}

	inputIdx := strings.Index(result, "--- input ---")
	inputSection := result[inputIdx:]

	if !strings.Contains(inputSection, `{"key": "value"}`) {
		t.Error("input section missing json content")
	}
	if !strings.Contains(inputSection, "more json content") {
		t.Error("input section missing second content line")
	}
	if strings.Contains(inputSection, "code-from-spec: SPEC/b@") {
		t.Error("input section should not contain artifact tag line")
	}

	contextIdx := strings.Index(result, "--- context ---")
	contextSection := result[contextIdx:inputIdx]
	if strings.Contains(contextSection, `{"key": "value"}`) {
		t.Error("context section should not contain input content")
	}
}

func TestMCPLoadChain_TC13_ExternalInputFullContentInInputSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: out/a.txt
input: EXTERNAL/docs/vendor/spec.yaml
---
# SPEC/a
`)

	testWriteFile(t, "docs/vendor/spec.yaml", `openapi: "3.0"
info:
  title: Test
`)

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "--- input ---") {
		t.Error("result missing '--- input ---'")
	}

	inputIdx := strings.Index(result, "--- input ---")
	inputSection := result[inputIdx:]

	if !strings.Contains(inputSection, `openapi: "3.0"`) {
		t.Error("input section missing external file content")
	}
	if !strings.Contains(inputSection, "title: Test") {
		t.Error("input section missing title line")
	}
}

func TestMCPLoadChain_TC14_NoInputSectionAbsent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: out/a.txt
---
# SPEC/a
`)

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(result, "--- input ---") {
		t.Error("result should not contain '--- input ---'")
	}
}

func TestMCPLoadChain_TC15_ExistingArtifactPresent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: out/a.go
---
# SPEC/a
`)

	testWriteFile(t, "out/a.go", `package main

func main() {}
`)

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "--- existing artifact ---") {
		t.Error("result missing '--- existing artifact ---'")
	}

	artifactIdx := strings.Index(result, "--- existing artifact ---")
	artifactSection := result[artifactIdx:]

	if !strings.Contains(artifactSection, "package main") {
		t.Error("existing artifact section missing file content")
	}
	if !strings.Contains(artifactSection, "func main() {}") {
		t.Error("existing artifact section missing func main")
	}
}

func TestMCPLoadChain_TC16_ExistingArtifactAbsent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: out/a.go
---
# SPEC/a
`)

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(result, "--- existing artifact ---") {
		t.Error("result should not contain '--- existing artifact ---' when file does not exist")
	}
}

func TestMCPLoadChain_TC17_HashIsDeterministic(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: out/a.txt
---
# SPEC/a

# Public

## Notes

Fixed notes content line
`)

	result1, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("first call unexpected error: %v", err)
	}

	result2, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("second call unexpected error: %v", err)
	}

	hash1 := strings.SplitN(result1, "\n", 2)[0]
	hash2 := strings.SplitN(result2, "\n", 2)[0]

	if hash1 != hash2 {
		t.Errorf("hashes differ: %q != %q", hash1, hash2)
	}
}

func TestMCPLoadChain_TC18_InvalidLogicalName(t *testing.T) {
	_, err := mcploadchain.MCPLoadChain("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("expected ErrUnsupportedReference, got: %v", err)
	}
}

func TestMCPLoadChain_TC19_NonexistentNodeFile(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	_, err := mcploadchain.MCPLoadChain("SPEC/nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestMCPLoadChain_TC20_NoOutputDeclared(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `# SPEC/a
`)

	_, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcploadchain.ErrNoOutput) {
		t.Errorf("expected ErrNoOutput, got: %v", err)
	}
}

func TestMCPLoadChain_TC21_InvalidOutputPathTraversal(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: ../../etc/passwd
---
# SPEC/a
`)

	_, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcploadchain.ErrInvalidOutputPath) {
		t.Errorf("expected ErrInvalidOutputPath, got: %v", err)
	}
}

func TestMCPLoadChain_TC22_UnresolvableDependency(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC
`)

	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: out/a.txt
depends_on:
  - SPEC/missing
---
# SPEC/a
`)

	_, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
