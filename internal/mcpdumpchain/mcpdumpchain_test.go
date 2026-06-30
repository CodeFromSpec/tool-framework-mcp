// code-from-spec: SPEC/golang/test/cases/mcp_tools/dump_chain@yrrZvhFhkmi-EWloiPEwSYhaPc4
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
	root.SetPublic("## Context\nsome context content")
	root.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.go")
	a.Write()

	result, err := mcpdumpchain.MCPDumpChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "wrote dump_chain.xml" {
		t.Errorf("expected 'wrote dump_chain.xml', got %q", result)
	}

	data, err := os.ReadFile("dump_chain.xml")
	if err != nil {
		t.Fatalf("dump_chain.xml not found: %v", err)
	}
	content := string(data)

	if !strings.HasPrefix(content, "chain_hash: ") {
		t.Errorf("content does not start with 'chain_hash: '")
	}

	lines := strings.SplitN(content, "\n", 2)
	hashLine := lines[0]
	hashPart := strings.TrimPrefix(hashLine, "chain_hash: ")
	hashPart = strings.TrimRight(hashPart, "\r")
	if len(hashPart) != 27 {
		t.Errorf("expected 27-character hash, got %d characters: %q", len(hashPart), hashPart)
	}

	if !strings.Contains(content, "<chain>") {
		t.Errorf("content does not contain '<chain>'")
	}
	if !strings.Contains(content, "</chain>") {
		t.Errorf("content does not contain '</chain>'")
	}
	if !strings.Contains(content, "<constraints>") {
		t.Errorf("content does not contain '<constraints>'")
	}
	if !strings.Contains(content, `<entry name="SPEC/root">`) {
		t.Errorf("content does not contain entry for SPEC/root")
	}
}

func TestMCPDumpChain_ContentMatchesMCPLoadChain(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.SetPublic("## Context\nsome context content")
	root.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.go")
	a.SetAgent("agent instructions here")
	a.Write()

	expected, err := mcploadchain.MCPLoadChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("MCPLoadChain failed: %v", err)
	}

	_, err = mcpdumpchain.MCPDumpChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("MCPDumpChain failed: %v", err)
	}

	data, err := os.ReadFile("dump_chain.xml")
	if err != nil {
		t.Fatalf("dump_chain.xml not found: %v", err)
	}

	if string(data) != expected {
		t.Errorf("file content does not match MCPLoadChain output\ngot:  %q\nwant: %q", string(data), expected)
	}
}

func TestMCPDumpChain_OverwritesExisting(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.SetPublic("## Context\nsome context content")
	root.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.go")
	a.Write()

	err := os.WriteFile("dump_chain.xml", []byte("old"), 0644)
	if err != nil {
		t.Fatalf("failed to write existing dump_chain.xml: %v", err)
	}

	_, err = mcpdumpchain.MCPDumpChain("SPEC/root/a")
	if err != nil {
		t.Fatalf("MCPDumpChain failed: %v", err)
	}

	data, err := os.ReadFile("dump_chain.xml")
	if err != nil {
		t.Fatalf("dump_chain.xml not found: %v", err)
	}

	if string(data) == "old" {
		t.Errorf("dump_chain.xml was not overwritten")
	}
	if !strings.Contains(string(data), "chain_hash: ") {
		t.Errorf("dump_chain.xml does not contain new chain content")
	}
}

func TestMCPDumpChain_NoOutput(t *testing.T) {
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
		t.Errorf("expected ErrNoOutput, got %v", err)
	}

	_, statErr := os.Stat("dump_chain.xml")
	if statErr == nil {
		t.Errorf("dump_chain.xml should not exist when error occurs")
	}
}

func TestMCPDumpChain_InvalidLogicalName(t *testing.T) {
	testutils.Chdir(t)

	_, err := mcpdumpchain.MCPDumpChain("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
