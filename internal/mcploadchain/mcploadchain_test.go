// code-from-spec: SPEC/golang/tests/mcp_tools/load_chain@LZqVKHj8aof9iQ3558TSdEh3F1k
package mcploadchain_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/file"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcploadchain"
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

func testWriteFile(t *testing.T, path, content string) {
	t.Helper()
	dir := path[:strings.LastIndex(path, "/")]
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testWriteFile MkdirAll %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile %s: %v", path, err)
	}
}

func testParseResult(result string) (firstLine, context, input, existingArtifact string) {
	lines := strings.Split(result, "\n")
	if len(lines) > 0 {
		firstLine = lines[0]
	}

	const (
		delimContext  = "--- context ---"
		delimInput    = "--- input ---"
		delimArtifact = "--- existing artifact ---"
	)

	var contextLines, inputLines, artifactLines []string
	section := ""
	for _, line := range lines[1:] {
		switch line {
		case delimContext:
			section = "context"
		case delimInput:
			section = "input"
		case delimArtifact:
			section = "artifact"
		default:
			switch section {
			case "context":
				contextLines = append(contextLines, line)
			case "input":
				inputLines = append(inputLines, line)
			case "artifact":
				artifactLines = append(artifactLines, line)
			}
		}
	}

	context = strings.Join(contextLines, "\n")
	input = strings.Join(inputLines, "\n")
	existingArtifact = strings.Join(artifactLines, "\n")
	return
}

func TestMCPLoadChain_TC01_SimpleLeafNode(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "---\n---\n# SPEC\n# Public\n## Context\nRoot context line.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.txt\n---\n# SPEC/a\n# Public\n## Interface\nInterface description.\n# Agent\nAgent guidance.\n")

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	firstLine, context, input, artifact := testParseResult(result)

	if !strings.HasPrefix(firstLine, "chain_hash: ") {
		t.Errorf("first line should start with 'chain_hash: ', got: %q", firstLine)
	}
	hashPart := strings.TrimSpace(strings.TrimPrefix(firstLine, "chain_hash: "))
	if len(hashPart) != 27 {
		t.Errorf("hash should be 27 characters, got %d: %q", len(hashPart), hashPart)
	}

	if !strings.Contains(context, "## Context") {
		t.Errorf("context should contain '## Context'")
	}
	if !strings.Contains(context, "Root context line.") {
		t.Errorf("context should contain 'Root context line.'")
	}
	if strings.Contains(context, "# Public") {
		t.Errorf("context should not contain '# Public'")
	}
	if !strings.Contains(context, "output: out/a.txt") {
		t.Errorf("context should contain 'output: out/a.txt'")
	}
	if !strings.Contains(context, "## Interface") {
		t.Errorf("context should contain '## Interface'")
	}
	if !strings.Contains(context, "Interface description.") {
		t.Errorf("context should contain 'Interface description.'")
	}
	if !strings.Contains(context, "# Agent") {
		t.Errorf("context should contain '# Agent'")
	}
	if !strings.Contains(context, "Agent guidance.") {
		t.Errorf("context should contain 'Agent guidance.'")
	}
	if input != "" {
		t.Errorf("input section should be absent, got: %q", input)
	}
	if artifact != "" {
		t.Errorf("existing artifact section should be absent, got: %q", artifact)
	}
}

func TestMCPLoadChain_TC02_AncestorPublicContentIncluded(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "---\n---\n# SPEC\n# Public\n## Overview\nRoot overview.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\n---\n# SPEC/a\n# Public\n## Details\nNode a details.\n")
	testWriteFile(t, "code-from-spec/a/b/_node.md", "---\noutput: out/b.txt\n---\n# SPEC/a/b\n")

	result, err := mcploadchain.MCPLoadChain("SPEC/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, context, _, _ := testParseResult(result)

	if !strings.Contains(context, "## Overview") {
		t.Errorf("context should contain '## Overview'")
	}
	if !strings.Contains(context, "Root overview.") {
		t.Errorf("context should contain 'Root overview.'")
	}
	if !strings.Contains(context, "## Details") {
		t.Errorf("context should contain '## Details'")
	}
	if !strings.Contains(context, "Node a details.") {
		t.Errorf("context should contain 'Node a details.'")
	}
	if strings.Contains(context, "# Public") {
		t.Errorf("context should not contain '# Public'")
	}
}

func TestMCPLoadChain_TC03_AncestorWithoutPublicSectionSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "---\n---\n# SPEC\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.txt\n---\n# SPEC/a\n# Public\n## Interface\nNode a interface.\n")

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, context, _, _ := testParseResult(result)

	if !strings.Contains(context, "## Interface") {
		t.Errorf("context should contain '## Interface'")
	}
	if !strings.Contains(context, "Node a interface.") {
		t.Errorf("context should contain 'Node a interface.'")
	}
}

func TestMCPLoadChain_TC04_AncestorWithEmptyPublicSectionSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "---\n---\n# SPEC\n# Public\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.txt\n---\n# SPEC/a\n# Public\n## Interface\nNode a interface.\n")

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, context, _, _ := testParseResult(result)

	if !strings.Contains(context, "## Interface") {
		t.Errorf("context should contain '## Interface'")
	}
	if !strings.Contains(context, "Node a interface.") {
		t.Errorf("context should contain 'Node a interface.'")
	}
}

func TestMCPLoadChain_TC05_DependencyWithoutQualifierFullPublicIncluded(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "---\n---\n# SPEC\n")
	testWriteFile(t, "code-from-spec/b/_node.md", "---\n---\n# SPEC/b\n# Public\n## Interface\nNode b interface.\n## Constraints\nNode b constraints.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.txt\ndepends_on:\n  - SPEC/b\n---\n# SPEC/a\n")

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, context, _, _ := testParseResult(result)

	if !strings.Contains(context, "## Interface") {
		t.Errorf("context should contain '## Interface'")
	}
	if !strings.Contains(context, "Node b interface.") {
		t.Errorf("context should contain 'Node b interface.'")
	}
	if !strings.Contains(context, "## Constraints") {
		t.Errorf("context should contain '## Constraints'")
	}
	if !strings.Contains(context, "Node b constraints.") {
		t.Errorf("context should contain 'Node b constraints.'")
	}
}

func TestMCPLoadChain_TC06_DependencyWithQualifierSubsectionOnly(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "---\n---\n# SPEC\n")
	testWriteFile(t, "code-from-spec/b/_node.md", "---\n---\n# SPEC/b\n# Public\n## Interface\nNode b interface.\n## Constraints\nNode b constraints.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.txt\ndepends_on:\n  - SPEC/b(interface)\n---\n# SPEC/a\n")

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, context, _, _ := testParseResult(result)

	if !strings.Contains(context, "## Interface") {
		t.Errorf("context should contain '## Interface'")
	}
	if !strings.Contains(context, "Node b interface.") {
		t.Errorf("context should contain 'Node b interface.'")
	}
	if strings.Contains(context, "## Constraints") {
		t.Errorf("context should not contain '## Constraints'")
	}
	if strings.Contains(context, "Node b constraints.") {
		t.Errorf("context should not contain 'Node b constraints.'")
	}
}

func TestMCPLoadChain_TC07_ArtifactDependencyTagLineRemoved(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "---\n---\n# SPEC\n")
	testWriteFile(t, "code-from-spec/b/_node.md", "---\noutput: out/b.go\n---\n# SPEC/b\n")
	testWriteFile(t, "out/b.go", "// code-from-spec: SPEC/b@AAAAAAAAAAAAAAAAAAAAAAAAAAA\npackage main\n// body content\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\ndepends_on:\n  - ARTIFACT/b\n---\n# SPEC/a\n")

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, context, _, _ := testParseResult(result)

	if !strings.Contains(context, "package main") {
		t.Errorf("context should contain 'package main'")
	}
	if !strings.Contains(context, "// body content") {
		t.Errorf("context should contain '// body content'")
	}
	if strings.Contains(context, "code-from-spec:") {
		t.Errorf("context should not contain the artifact tag line")
	}
}

func TestMCPLoadChain_TC08_ExternalDependencyFullContent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "---\n---\n# SPEC\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.txt\ndepends_on:\n  - EXTERNAL/data/config.yaml\n---\n# SPEC/a\n")
	testWriteFile(t, "data/config.yaml", "key: value\nsetting: enabled\n")

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, context, _, _ := testParseResult(result)

	if !strings.Contains(context, "key: value") {
		t.Errorf("context should contain 'key: value'")
	}
	if !strings.Contains(context, "setting: enabled") {
		t.Errorf("context should contain 'setting: enabled'")
	}
}

func TestMCPLoadChain_TC09_TargetFrontmatterOutputOnly(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "---\n---\n# SPEC\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.txt\ndepends_on:\n  - SPEC/b\n---\n# SPEC/a\n")
	testWriteFile(t, "code-from-spec/b/_node.md", "---\n---\n# SPEC/b\n")

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, context, _, _ := testParseResult(result)

	if !strings.Contains(context, "output: out/a.txt") {
		t.Errorf("context should contain 'output: out/a.txt'")
	}
	if strings.Contains(context, "depends_on") {
		t.Errorf("context frontmatter should not contain 'depends_on'")
	}
}

func TestMCPLoadChain_TC10_TargetAgentSectionIncluded(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "---\n---\n# SPEC\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.txt\n---\n# SPEC/a\n# Public\n## Interface\nPublic interface.\n# Agent\nAgent-specific guidance.\n")

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, context, _, _ := testParseResult(result)

	if !strings.Contains(context, "## Interface") {
		t.Errorf("context should contain '## Interface'")
	}
	if !strings.Contains(context, "Public interface.") {
		t.Errorf("context should contain 'Public interface.'")
	}
	if !strings.Contains(context, "# Agent") {
		t.Errorf("context should contain '# Agent'")
	}
	if !strings.Contains(context, "Agent-specific guidance.") {
		t.Errorf("context should contain 'Agent-specific guidance.'")
	}
	if strings.Contains(context, "# Public") {
		t.Errorf("context should not contain '# Public'")
	}
}

func TestMCPLoadChain_TC11_TargetWithoutAgentSectionSkipped(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "---\n---\n# SPEC\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.txt\n---\n# SPEC/a\n# Public\n## Interface\nPublic interface.\n")

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, context, _, _ := testParseResult(result)

	if !strings.Contains(context, "## Interface") {
		t.Errorf("context should contain '## Interface'")
	}
	if !strings.Contains(context, "Public interface.") {
		t.Errorf("context should contain 'Public interface.'")
	}
	if strings.Contains(context, "# Agent") {
		t.Errorf("context should not contain '# Agent'")
	}
}

func TestMCPLoadChain_TC12_InputPresentInSeparateSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "---\n---\n# SPEC\n")
	testWriteFile(t, "code-from-spec/b/_node.md", "---\noutput: out/data.json\n---\n# SPEC/b\n")
	testWriteFile(t, "out/data.json", "// code-from-spec: SPEC/b@AAAAAAAAAAAAAAAAAAAAAAAAAAA\n{\"key\": \"value\"}\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.txt\ninput: ARTIFACT/b\n---\n# SPEC/a\n")

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "--- input ---") {
		t.Errorf("result should contain '--- input ---'")
	}

	_, context, input, _ := testParseResult(result)

	if !strings.Contains(input, "{\"key\": \"value\"}") {
		t.Errorf("input section should contain '{\"key\": \"value\"}'")
	}
	if strings.Contains(input, "code-from-spec:") {
		t.Errorf("input section should not contain the artifact tag line")
	}
	if strings.Contains(context, "{\"key\": \"value\"}") {
		t.Errorf("input content should not appear in context section")
	}
}

func TestMCPLoadChain_TC13_ExternalInputFullContentInInputSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "---\n---\n# SPEC\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.txt\ninput: EXTERNAL/docs/vendor/spec.yaml\n---\n# SPEC/a\n")
	testWriteFile(t, "docs/vendor/spec.yaml", "version: 1\ntitle: Vendor spec\n")

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "--- input ---") {
		t.Errorf("result should contain '--- input ---'")
	}

	_, _, input, _ := testParseResult(result)

	if !strings.Contains(input, "version: 1") {
		t.Errorf("input section should contain 'version: 1'")
	}
	if !strings.Contains(input, "title: Vendor spec") {
		t.Errorf("input section should contain 'title: Vendor spec'")
	}
}

func TestMCPLoadChain_TC14_NoInputSectionAbsent(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "---\n---\n# SPEC\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.txt\n---\n# SPEC/a\n")

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(result, "--- input ---") {
		t.Errorf("result should not contain '--- input ---'")
	}
}

func TestMCPLoadChain_TC15_ExistingArtifactPresentInSeparateSection(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "---\n---\n# SPEC\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# SPEC/a\n")
	testWriteFile(t, "out/a.go", "package main\nfunc main() {}\n")

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "--- existing artifact ---") {
		t.Errorf("result should contain '--- existing artifact ---'")
	}

	_, _, _, artifact := testParseResult(result)

	if !strings.Contains(artifact, "package main") {
		t.Errorf("existing artifact section should contain 'package main'")
	}
	if !strings.Contains(artifact, "func main() {}") {
		t.Errorf("existing artifact section should contain 'func main() {}'")
	}
}

func TestMCPLoadChain_TC16_ExistingArtifactAbsentSectionOmitted(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "---\n---\n# SPEC\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.go\n---\n# SPEC/a\n")

	result, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(result, "--- existing artifact ---") {
		t.Errorf("result should not contain '--- existing artifact ---'")
	}
}

func TestMCPLoadChain_TC17_HashIsDeterministic(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "---\n---\n# SPEC\n# Public\n## Overview\nStable content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.txt\n---\n# SPEC/a\n")

	result1, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("first call unexpected error: %v", err)
	}
	result2, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("second call unexpected error: %v", err)
	}

	firstLine1 := strings.SplitN(result1, "\n", 2)[0]
	firstLine2 := strings.SplitN(result2, "\n", 2)[0]

	if firstLine1 != firstLine2 {
		t.Errorf("hash lines differ: %q vs %q", firstLine1, firstLine2)
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
	if !errors.Is(err, file.ErrFileUnreadable) {
		t.Errorf("expected file.ErrFileUnreadable, got: %v", err)
	}
}

func TestMCPLoadChain_TC20_NoOutputDeclared(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testWriteFile(t, "code-from-spec/_node.md", "---\n---\n# SPEC\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\n---\n# SPEC/a\n")

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

	testWriteFile(t, "code-from-spec/_node.md", "---\n---\n# SPEC\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: ../../etc/passwd\n---\n# SPEC/a\n")

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

	testWriteFile(t, "code-from-spec/_node.md", "---\n---\n# SPEC\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: out/a.txt\ndepends_on:\n  - SPEC/missing\n---\n# SPEC/a\n")

	_, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
