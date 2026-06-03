// code-from-spec: ROOT/golang/tests/mcp_tools/chain_hash@DeP64QBrGm3Z7mSSo7AAkpGNMoA
package mcpchainhash_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpchainhash"
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
	dir := path[:strings.LastIndex(path, "/")]
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testWriteFile MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile WriteFile: %v", err)
	}
}

func testSetupBasicTree(t *testing.T) {
	t.Helper()
	testWriteFile(t, "code-from-spec/_node.md", "---\n---\n# ROOT\n\nPublic section.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: some/output.go\n---\n# ROOT/a\n\nLeaf node.\n")
}

func TestMCPChainHash_Returns27CharHash(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testSetupBasicTree(t)

	result, err := mcpchainhash.MCPChainHash("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 27 {
		t.Errorf("expected 27-character hash, got %d characters: %q", len(result), result)
	}
}

func TestMCPChainHash_IsDeterministic(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testSetupBasicTree(t)

	hash1, err := mcpchainhash.MCPChainHash("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error on first call: %v", err)
	}

	hash2, err := mcpchainhash.MCPChainHash("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("hash is not deterministic: %q != %q", hash1, hash2)
	}
}

func TestMCPChainHash_MatchesLoadChainHash(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testSetupBasicTree(t)

	chainHashResult, err := mcpchainhash.MCPChainHash("ROOT/a")
	if err != nil {
		t.Fatalf("MCPChainHash error: %v", err)
	}

	loadChainResult, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("MCPLoadChain error: %v", err)
	}

	if chainHashResult != loadChainResult.ChainHash {
		t.Errorf("hash mismatch: MCPChainHash=%q, MCPLoadChain.ChainHash=%q", chainHashResult, loadChainResult.ChainHash)
	}
}

func TestMCPChainHash_InvalidLogicalName(t *testing.T) {
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
	testWriteFile(t, "code-from-spec/_node.md", "---\n---\n# ROOT\n\nPublic section.\n")

	_, err := mcpchainhash.MCPChainHash("ROOT/nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestMCPChainHash_NoOutputDeclared(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)
	testWriteFile(t, "code-from-spec/_node.md", "---\n---\n# ROOT\n\nPublic section.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\n---\n# ROOT/a\n\nLeaf node without output.\n")

	_, err := mcpchainhash.MCPChainHash("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpchainhash.ErrNoOutput) {
		t.Errorf("expected ErrNoOutput, got: %v", err)
	}
}
