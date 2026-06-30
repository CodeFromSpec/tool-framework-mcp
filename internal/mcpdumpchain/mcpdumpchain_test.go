package mcpdumpchain_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/mcpdumpchain"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/mcploadchain"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"
)

func TestMCPDumpChain_WritesDumpChainXml(t *testing.T) {
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
	if result != "wrote dump_chain.xml" {
		t.Fatalf("expected 'wrote dump_chain.xml', got %q", result)
	}

	data, err := os.ReadFile("dump_chain.xml")
	if err != nil {
		t.Fatalf("dump_chain.xml not found: %v", err)
	}
	content := string(data)
	if !strings.HasPrefix(content, "<chain>") {
		t.Errorf("content does not start with <chain>: %q", content[:min(len(content), 50)])
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

	expected, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("MCPLoadChain error: %v", err)
	}

	_, err = mcpdumpchain.MCPDumpChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("MCPDumpChain error: %v", err)
	}

	data, err := os.ReadFile("dump_chain.xml")
	if err != nil {
		t.Fatalf("dump_chain.xml not found: %v", err)
	}
	if string(data) != expected {
		t.Errorf("dump_chain.xml content does not match MCPLoadChain output\ngot:  %q\nwant: %q", string(data), expected)
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

	err := os.WriteFile("dump_chain.xml", []byte("old"), 0o644)
	if err != nil {
		t.Fatalf("failed to write old dump_chain.xml: %v", err)
	}

	_, err = mcpdumpchain.MCPDumpChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("MCPDumpChain error: %v", err)
	}

	data, err := os.ReadFile("dump_chain.xml")
	if err != nil {
		t.Fatalf("dump_chain.xml not found: %v", err)
	}
	if string(data) == "old" {
		t.Error("dump_chain.xml still contains old content")
	}
	if !strings.Contains(string(data), "<chain>") {
		t.Error("dump_chain.xml does not contain new chain content")
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

	_, statErr := os.Stat("dump_chain.xml")
	if statErr == nil {
		t.Error("dump_chain.xml should not exist after error")
	}
}

func TestMCPDumpChain_InvalidLogicalName(t *testing.T) {
	testutils.Chdir(t)

	_, err := mcpdumpchain.MCPDumpChain("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
