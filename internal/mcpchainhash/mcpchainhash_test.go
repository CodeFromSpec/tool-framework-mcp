// code-from-spec: SPEC/golang/tests/mcp_tools/chain_hash@c1jTMlnV7Q7qxaPC-q3FzhLBF2c
package mcpchainhash_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcpchainhash"
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

func testWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	dir := path[:strings.LastIndex(path, "/")]
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

func testSetupBasicTree(t *testing.T) {
	t.Helper()
	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n\n## Context\n\nRoot node context.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: some/output/path.md\n---\n\n# SPEC/a\n\n## Context\n\nLeaf node content.\n")
}

func TestMCPChainHash_Returns27CharHash(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testSetupBasicTree(t)

	hash, err := mcpchainhash.MCPChainHash("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27-character hash, got %d characters: %q", len(hash), hash)
	}
}

func TestMCPChainHash_Deterministic(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testSetupBasicTree(t)

	hash1, err := mcpchainhash.MCPChainHash("SPEC/a")
	if err != nil {
		t.Fatalf("first call unexpected error: %v", err)
	}

	hash2, err := mcpchainhash.MCPChainHash("SPEC/a")
	if err != nil {
		t.Fatalf("second call unexpected error: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("hash is not deterministic: %q != %q", hash1, hash2)
	}
}

func TestMCPChainHash_MatchesLoadChainHash(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testSetupBasicTree(t)

	chainHashResult, err := mcpchainhash.MCPChainHash("SPEC/a")
	if err != nil {
		t.Fatalf("MCPChainHash unexpected error: %v", err)
	}

	loadChainResult, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("MCPLoadChain unexpected error: %v", err)
	}

	firstLine := strings.SplitN(loadChainResult, "\n", 2)[0]
	const prefix = "chain_hash: "
	if !strings.HasPrefix(firstLine, prefix) {
		t.Fatalf("unexpected first line format: %q", firstLine)
	}
	loadChainHash := strings.TrimPrefix(firstLine, prefix)

	if chainHashResult != loadChainHash {
		t.Errorf("hash mismatch: MCPChainHash=%q, MCPLoadChain=%q", chainHashResult, loadChainHash)
	}
}

func TestMCPChainHash_InvalidLogicalName(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	_, err := mcpchainhash.MCPChainHash("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("expected ErrUnsupportedReference, got: %v", err)
	}
}

func TestMCPChainHash_NonexistentNodeFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n\n## Context\n\nRoot node context.\n")

	_, err := mcpchainhash.MCPChainHash("SPEC/nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestMCPChainHash_NoOutputDeclared(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "code-from-spec/_node.md", "# SPEC\n\n## Context\n\nRoot node context.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# SPEC/a\n\n## Context\n\nLeaf node without output.\n")

	_, err := mcpchainhash.MCPChainHash("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpchainhash.ErrNoOutput) {
		t.Errorf("expected ErrNoOutput, got: %v", err)
	}
}
