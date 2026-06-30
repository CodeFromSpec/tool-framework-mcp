// code-from-spec: SPEC/golang/test/cases/mcp_tools/load_chain@-S1D5etJqp6dDyhv_P20Ufcmc2k
package mcploadchain_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/mcploadchain"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"
)

func TestMCPLoadChain_SimpleLeafNode(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.SetPublic("## Context\nroot context content")
	root.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.txt")
	a.SetPublic("## Interface\ninterface content")
	a.SetAgent("agent instructions here")
	a.Write()

	result, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.SplitN(result, "\n", 2)
	if len(lines) < 2 {
		t.Fatalf("expected at least two lines, got: %q", result)
	}
	firstLine := lines[0]
	if !strings.HasPrefix(firstLine, "chain_hash: ") {
		t.Errorf("first line does not start with 'chain_hash: ': %q", firstLine)
	}
	hash := strings.TrimPrefix(firstLine, "chain_hash: ")
	hash = strings.TrimSpace(hash)
	if len(hash) != 27 {
		t.Errorf("expected 27-character hash, got %d characters: %q", len(hash), hash)
	}

	if !strings.Contains(result, "<chain>") {
		t.Error("expected <chain> root element")
	}
	if !strings.Contains(result, `<entry name="SPEC/root"`) {
		t.Error("expected entry for SPEC/root")
	}
	if !strings.Contains(result, "## Context") {
		t.Error("expected ## Context content in constraints")
	}
	if !strings.Contains(result, `<entry name="SPEC/root/a"`) {
		t.Error("expected entry for SPEC/root/a")
	}
	if !strings.Contains(result, "## Interface") {
		t.Error("expected ## Interface content in constraints")
	}
	if strings.Contains(result, "# Public") {
		t.Error("# Public heading should not appear in output")
	}
	if !strings.Contains(result, "<instructions>") {
		t.Error("expected <instructions> element")
	}
	if !strings.Contains(result, "agent instructions here") {
		t.Error("expected agent content in instructions")
	}
	if strings.Contains(result, "# Agent") {
		t.Error("# Agent heading should not appear in instructions")
	}
	if strings.Contains(result, "<existing_artifact>") {
		t.Error("expected no <existing_artifact> section")
	}
	if strings.Contains(result, "<input>") {
		t.Error("expected no <input> section")
	}
}

func TestMCPLoadChain_AncestorPublicContentIncluded(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.SetPublic("## Overview\nroot overview")
	root.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetPublic("## Details\na details")
	a.Write()

	b := testutils.CreateSpecNode(t, "SPEC/root/a/b")
	b.SetOutput("out/b.txt")
	b.SetPublic("## Contract\nb contract")
	b.Write()

	result, err := mcploadchain.MCPLoadChain("SPEC/root/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `<entry name="SPEC/root"`) {
		t.Error("expected entry for SPEC/root")
	}
	if !strings.Contains(result, "## Overview") {
		t.Error("expected ## Overview content")
	}
	if !strings.Contains(result, `<entry name="SPEC/root/a"`) {
		t.Error("expected entry for SPEC/root/a")
	}
	if !strings.Contains(result, "## Details") {
		t.Error("expected ## Details content")
	}
	if !strings.Contains(result, `<entry name="SPEC/root/a/b"`) {
		t.Error("expected entry for SPEC/root/a/b")
	}
	if !strings.Contains(result, "## Contract") {
		t.Error("expected ## Contract content")
	}
}

func TestMCPLoadChain_AncestorWithoutPublicSkipped(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.txt")
	a.SetPublic("## Interface\ninterface content")
	a.Write()

	result, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(result, `<entry name="SPEC/root"`) {
		t.Error("expected no entry for SPEC/root (no public section)")
	}
	if !strings.Contains(result, `<entry name="SPEC/root/a"`) {
		t.Error("expected entry for SPEC/root/a")
	}
	if !strings.Contains(result, "## Interface") {
		t.Error("expected ## Interface content")
	}
}

func TestMCPLoadChain_AncestorWithEmptyPublicSkipped(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.SetPublic("")
	root.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.txt")
	a.SetPublic("## Interface\ninterface content")
	a.Write()

	result, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(result, `<entry name="SPEC/root"`) {
		t.Error("expected no entry for SPEC/root (empty public section)")
	}
}

func TestMCPLoadChain_DependencyWithoutQualifier(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	b := testutils.CreateSpecNode(t, "SPEC/root/b")
	b.SetPublic("## Interface\nb interface\n## Constraints\nb constraints")
	b.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.txt")
	a.AddDependsOn("SPEC/root/b")
	a.Write()

	result, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `<entry name="SPEC/root/b"`) {
		t.Error("expected entry for SPEC/root/b")
	}
	if !strings.Contains(result, "## Interface") {
		t.Error("expected ## Interface content")
	}
	if !strings.Contains(result, "## Constraints") {
		t.Error("expected ## Constraints content")
	}
}

func TestMCPLoadChain_DependencyWithQualifier(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	b := testutils.CreateSpecNode(t, "SPEC/root/b")
	b.SetPublic("## Interface\nb interface\n## Constraints\nb constraints")
	b.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.txt")
	a.AddDependsOn("SPEC/root/b(interface)")
	a.Write()

	result, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `<entry name="SPEC/root/b(interface)"`) {
		t.Error("expected entry for SPEC/root/b(interface)")
	}
	if !strings.Contains(result, "## Interface") {
		t.Error("expected ## Interface content")
	}
	if strings.Contains(result, "## Constraints") {
		t.Error("expected ## Constraints content to be excluded when qualifier is 'interface'")
	}
}

func TestMCPLoadChain_ARTIFACTDependency(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	b := testutils.CreateSpecNode(t, "SPEC/root/b")
	b.SetOutput("out/b.go")
	b.Write()

	if err := os.MkdirAll("out", 0755); err != nil {
		t.Fatalf("failed to create out dir: %v", err)
	}
	if err := os.WriteFile("out/b.go", []byte("package b\n// artifact content"), 0644); err != nil {
		t.Fatalf("failed to write out/b.go: %v", err)
	}

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.go")
	a.AddDependsOn("ARTIFACT/root/b")
	a.Write()

	result, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `<entry name="ARTIFACT/root/b"`) {
		t.Error("expected entry for ARTIFACT/root/b")
	}
	if !strings.Contains(result, "artifact content") {
		t.Error("expected artifact file content in constraints")
	}
}

func TestMCPLoadChain_EXTERNALDependency(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	if err := os.MkdirAll("data", 0755); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}
	if err := os.WriteFile("data/config.yaml", []byte("key: value\nother: data"), 0644); err != nil {
		t.Fatalf("failed to write data/config.yaml: %v", err)
	}

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.txt")
	a.AddDependsOn("EXTERNAL/data/config.yaml")
	a.Write()

	result, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `<entry name="EXTERNAL/data/config.yaml"`) {
		t.Error("expected entry for EXTERNAL/data/config.yaml")
	}
	if !strings.Contains(result, "key: value") {
		t.Error("expected external file content in constraints")
	}
}

func TestMCPLoadChain_TargetAgentSectionInInstructions(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.txt")
	a.SetPublic("## Interface\nsome interface")
	a.SetAgent("do this specific thing")
	a.Write()

	result, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `<entry name="SPEC/root/a"`) {
		t.Error("expected entry for SPEC/root/a")
	}
	if !strings.Contains(result, "## Interface") {
		t.Error("expected ## Interface in constraints")
	}
	if !strings.Contains(result, "<instructions>") {
		t.Error("expected <instructions> element")
	}
	if !strings.Contains(result, "do this specific thing") {
		t.Error("expected agent content in instructions")
	}
}

func TestMCPLoadChain_TargetWithoutAgentSection(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.txt")
	a.SetPublic("## Interface\nsome interface")
	a.Write()

	result, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(result, "<instructions>") {
		t.Error("expected no <instructions> element when no agent section")
	}
}

func TestMCPLoadChain_InputPresentARTIFACT(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	b := testutils.CreateSpecNode(t, "SPEC/root/b")
	b.SetOutput("out/data.json")
	b.Write()

	if err := os.MkdirAll("out", 0755); err != nil {
		t.Fatalf("failed to create out dir: %v", err)
	}
	if err := os.WriteFile("out/data.json", []byte(`{"key":"value"}`), 0644); err != nil {
		t.Fatalf("failed to write out/data.json: %v", err)
	}

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.txt")
	a.SetInput("ARTIFACT/root/b")
	a.Write()

	result, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "<input>") {
		t.Error("expected <input> element")
	}
	if !strings.Contains(result, `{"key":"value"}`) {
		t.Error("expected artifact file content in input")
	}
	constraintsIdx := strings.Index(result, "<constraints>")
	inputContentIdx := strings.Index(result, `{"key":"value"}`)
	if constraintsIdx >= 0 && inputContentIdx >= 0 {
		constraintsEnd := strings.Index(result, "</constraints>")
		if constraintsEnd >= 0 && inputContentIdx < constraintsEnd {
			t.Error("input content should not appear inside <constraints>")
		}
	}
}

func TestMCPLoadChain_EXTERNALInput(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	if err := os.MkdirAll("docs/vendor", 0755); err != nil {
		t.Fatalf("failed to create docs/vendor dir: %v", err)
	}
	if err := os.WriteFile("docs/vendor/spec.yaml", []byte("spec: content\nversion: 1"), 0644); err != nil {
		t.Fatalf("failed to write docs/vendor/spec.yaml: %v", err)
	}

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.txt")
	a.SetInput("EXTERNAL/docs/vendor/spec.yaml")
	a.Write()

	result, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "<input>") {
		t.Error("expected <input> element")
	}
	if !strings.Contains(result, "spec: content") {
		t.Error("expected external file content in input")
	}
}

func TestMCPLoadChain_SPECInput(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	b := testutils.CreateSpecNode(t, "SPEC/root/b")
	b.SetPublic("## Acceptance tests\nacceptance test content")
	b.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.txt")
	a.SetInput("SPEC/root/b")
	a.Write()

	result, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "<input>") {
		t.Error("expected <input> element")
	}
	if !strings.Contains(result, "## Acceptance tests") {
		t.Error("expected ## Acceptance tests content in input")
	}
}

func TestMCPLoadChain_NoInput(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.txt")
	a.Write()

	result, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(result, "<input>") {
		t.Error("expected no <input> element when no input declared")
	}
}

func TestMCPLoadChain_ExistingArtifactPresent(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.go")
	a.Write()

	if err := os.MkdirAll("out", 0755); err != nil {
		t.Fatalf("failed to create out dir: %v", err)
	}
	if err := os.WriteFile("out/a.go", []byte("package a\n// existing content"), 0644); err != nil {
		t.Fatalf("failed to write out/a.go: %v", err)
	}

	result, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "<existing_artifact>") {
		t.Error("expected <existing_artifact> element")
	}
	if !strings.Contains(result, "existing content") {
		t.Error("expected artifact file content in existing_artifact")
	}
}

func TestMCPLoadChain_ExistingArtifactAbsent(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.go")
	a.Write()

	result, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(result, "<existing_artifact>") {
		t.Error("expected no <existing_artifact> element when file does not exist")
	}
}

func TestMCPLoadChain_HashIsDeterministic(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.SetPublic("## Overview\nstable overview content")
	root.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.txt")
	a.Write()

	result1, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("first call unexpected error: %v", err)
	}

	result2, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("second call unexpected error: %v", err)
	}

	hash1 := strings.SplitN(result1, "\n", 2)[0]
	hash2 := strings.SplitN(result2, "\n", 2)[0]
	if hash1 != hash2 {
		t.Errorf("hashes differ: %q vs %q", hash1, hash2)
	}
}

func TestMCPLoadChain_InvalidLogicalName(t *testing.T) {
	testutils.Chdir(t)

	_, err := mcploadchain.MCPLoadChain("INVALID/something")
	if err == nil {
		t.Fatal("expected error for invalid logical name")
	}
	if !errors.Is(err, parsing.ErrUnrecognizedPrefix) {
		t.Errorf("expected ErrUnrecognizedPrefix, got: %v", err)
	}
}

func TestMCPLoadChain_NonexistentNodeFile(t *testing.T) {
	testutils.Chdir(t)

	_, err := mcploadchain.MCPLoadChain("SPEC/root/nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent node")
	}
	if !errors.Is(err, oslayer.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestMCPLoadChain_NoOutputDeclared(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.Write()

	_, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err == nil {
		t.Fatal("expected error when no output declared")
	}
	if !errors.Is(err, mcploadchain.ErrNoOutput) {
		t.Errorf("expected ErrNoOutput, got: %v", err)
	}
}

func TestMCPLoadChain_InvalidOutputPathTraversal(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("../../etc/passwd")
	a.Write()

	_, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err == nil {
		t.Fatal("expected error for traversal output path")
	}
	if !errors.Is(err, mcploadchain.ErrInvalidOutputPath) {
		t.Errorf("expected ErrInvalidOutputPath, got: %v", err)
	}
}

func TestMCPLoadChain_ModifiedArtifactBlocked(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.go")
	a.Write()

	if err := os.MkdirAll("out", 0755); err != nil {
		t.Fatalf("failed to create out dir: %v", err)
	}
	if err := os.WriteFile("out/a.go", []byte("original"), 0644); err != nil {
		t.Fatalf("failed to write out/a.go: %v", err)
	}

	if err := os.MkdirAll("code-from-spec", 0755); err != nil {
		t.Fatalf("failed to create code-from-spec dir: %v", err)
	}
	manifestContent := "code-from-spec: v5\n" +
		"ARTIFACT/root/a;path:out/a.go;checksum:Kx9mP2vB7wY2tHsJ8dFak4Xz9pQ;chain:Jz3qR7nL5cW1gT4yK8mDfAx0vBe\n"
	if err := os.WriteFile("code-from-spec/.manifest", []byte(manifestContent), 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	if err := os.WriteFile("out/a.go", []byte("modified"), 0644); err != nil {
		t.Fatalf("failed to overwrite out/a.go: %v", err)
	}

	_, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err == nil {
		t.Fatal("expected error for modified artifact")
	}
	if !errors.Is(err, mcploadchain.ErrArtifactModified) {
		t.Errorf("expected ErrArtifactModified, got: %v", err)
	}
}

func TestMCPLoadChain_NoManifestModifiedCheckSkipped(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.go")
	a.Write()

	if err := os.MkdirAll("out", 0755); err != nil {
		t.Fatalf("failed to create out dir: %v", err)
	}
	if err := os.WriteFile("out/a.go", []byte("some content"), 0644); err != nil {
		t.Fatalf("failed to write out/a.go: %v", err)
	}

	_, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error when no manifest: %v", err)
	}
}

func TestMCPLoadChain_UnresolvableDependency(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.txt")
	a.AddDependsOn("SPEC/root/missing")
	a.Write()

	_, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err == nil {
		t.Fatal("expected error for unresolvable dependency")
	}
}
