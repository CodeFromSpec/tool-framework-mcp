package mcpdumpchain_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/mcpdumpchain"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/mcploadchain"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/subagenttoken"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/testutils"
)

func TestMCPDumpChain_WritesDumpFile(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.SetPublic("## Context\nsome content")
	root.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.go")
	a.Write()

	result, err := mcpdumpchain.MCPDumpChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "wrote code-from-spec/.dump/SPEC_root_a.xml" {
		t.Fatalf("expected 'wrote code-from-spec/.dump/SPEC_root_a.xml', got %q", result)
	}

	data, err := os.ReadFile("code-from-spec/.dump/SPEC_root_a.xml")
	if err != nil {
		t.Fatalf("code-from-spec/.dump/SPEC_root_a.xml not found: %v", err)
	}
	content := string(data)

	if !strings.HasPrefix(content, "<chain>") {
		t.Errorf("content does not start with <chain>: %q", content[:min(len(content), 60)])
	}
	if !strings.Contains(content, "</chain>") {
		t.Error("content does not contain </chain>")
	}
	if !strings.Contains(content, "<constraints>") {
		t.Error("content does not contain <constraints>")
	}
	if !strings.Contains(content, `<entry name="SPEC/root">`) {
		t.Error("content does not contain <entry name=\"SPEC/root\">")
	}
}

func TestMCPDumpChain_ContentMatchesMCPLoadChain(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.SetPublic("## Context\nsome content")
	root.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.go")
	a.SetAgent("agent instructions here")
	a.Write()

	token, err := subagenttoken.SubagentTokenGenerate("SPEC/root/a")
	if err != nil {
		t.Fatalf("SubagentTokenGenerate error: %v", err)
	}

	expected, err := mcploadchain.MCPLoadChain(token)
	if err != nil {
		t.Fatalf("MCPLoadChain error: %v", err)
	}

	_, err = mcpdumpchain.MCPDumpChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("MCPDumpChain error: %v", err)
	}

	data, err := os.ReadFile("code-from-spec/.dump/SPEC_root_a.xml")
	if err != nil {
		t.Fatalf("code-from-spec/.dump/SPEC_root_a.xml not found: %v", err)
	}
	if string(data) != expected {
		t.Errorf("dump file content does not match MCPLoadChain output\ngot:  %q\nwant: %q", string(data), expected)
	}
}

func TestMCPDumpChain_OverwritesExistingFile(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.SetPublic("## Context\nsome content")
	root.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.go")
	a.Write()

	err := os.MkdirAll("code-from-spec/.dump", 0o755)
	if err != nil {
		t.Fatalf("failed to create dump directory: %v", err)
	}
	err = os.WriteFile("code-from-spec/.dump/SPEC_root_a.xml", []byte("old"), 0o644)
	if err != nil {
		t.Fatalf("failed to write old dump file: %v", err)
	}

	_, err = mcpdumpchain.MCPDumpChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("MCPDumpChain error: %v", err)
	}

	data, err := os.ReadFile("code-from-spec/.dump/SPEC_root_a.xml")
	if err != nil {
		t.Fatalf("dump file not found: %v", err)
	}
	if string(data) == "old" {
		t.Error("dump file still contains old content")
	}
	if !strings.Contains(string(data), "<chain>") {
		t.Error("dump file does not contain new chain content")
	}
}

func TestMCPDumpChain_NoOutputDeclared(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.Write()

	_, err := mcpdumpchain.MCPDumpChain("SPEC/root/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcploadchain.ErrNoOutput) {
		t.Errorf("expected mcploadchain.ErrNoOutput, got %v", err)
	}

	_, statErr := os.Stat("code-from-spec/.dump/SPEC_root_a.xml")
	if statErr == nil {
		t.Error("dump file should not exist after error")
	}
}

func TestMCPDumpChain_InvalidLogicalName(t *testing.T) {
	testutils.Chdir(t)

	_, err := mcpdumpchain.MCPDumpChain("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
