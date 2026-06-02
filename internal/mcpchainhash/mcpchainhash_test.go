// code-from-spec: ROOT/golang/tests/mcp_tools/chain_hash@_jhTR0X2udtCZ-CbTr6HLKORo4c
package mcpchainhash_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
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
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

func TestMCPChainHash_Returns27CharHash(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n## Public\n\nRoot node.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: internal/a/a.go\n---\n# ROOT/a\n\nLeaf node.\n")

	hash, err := mcpchainhash.MCPChainHash("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27-character hash, got %d characters: %q", len(hash), hash)
	}
}

func TestMCPChainHash_Deterministic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n## Public\n\nKnown root content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: internal/a/a.go\n---\n# ROOT/a\n\nKnown leaf content.\n")

	hash1, err := mcpchainhash.MCPChainHash("ROOT/a")
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}

	hash2, err := mcpchainhash.MCPChainHash("ROOT/a")
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("hash not deterministic: %q != %q", hash1, hash2)
	}
}

func TestMCPChainHash_MatchesLoadChainHash(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n## Public\n\nRoot node.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: internal/a/a.go\n---\n# ROOT/a\n\nLeaf node.\n")

	hashA, err := mcpchainhash.MCPChainHash("ROOT/a")
	if err != nil {
		t.Fatalf("MCPChainHash error: %v", err)
	}

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("MCPLoadChain error: %v", err)
	}

	if hashA != result.ChainHash {
		t.Errorf("hash mismatch: MCPChainHash=%q, MCPLoadChain.ChainHash=%q", hashA, result.ChainHash)
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
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := mcpchainhash.MCPChainHash("ROOT/nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestMCPChainHash_NoOutputDeclared(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n## Public\n\nRoot node.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\nLeaf node with no output.\n")

	_, err := mcpchainhash.MCPChainHash("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpchainhash.ErrNoOutput) {
		t.Errorf("expected ErrNoOutput, got: %v", err)
	}
}
