// code-from-spec: ROOT/golang/tests/mcp_tools/chain_hash@rSsNU-IeOTc4OOiKANTvrT3fw9I
package mcpchainhash_test

import (
	"errors"
	"os"
	"path/filepath"
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

func testWriteFile(t *testing.T, name string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(name, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile write: %v", err)
	}
}

func testWriteNode(t *testing.T, logicalName string, content string) {
	t.Helper()
	path := filepath.Join("code-from-spec", filepath.FromSlash(logicalName[len("ROOT"):]), "_node.md")
	testWriteFile(t, path, content)
}

func TestMCPChainHash_Returns27CharacterHash(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "# ROOT\n\n# Public\n\nPublic content.\n")
	testWriteNode(t, "ROOT/a", "---\noutput: out/a.go\n---\n# ROOT/a\n\n# Public\n\nLeaf content.\n")

	hash, err := mcpchainhash.MCPChainHash("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27-character hash, got %d characters: %q", len(hash), hash)
	}
}

func TestMCPChainHash_IsDeterministic(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "# ROOT\n\n# Public\n\nKnown public content.\n")
	testWriteNode(t, "ROOT/a", "---\noutput: out/a.go\n---\n# ROOT/a\n\n# Public\n\nKnown leaf content.\n")

	hash1, err := mcpchainhash.MCPChainHash("ROOT/a")
	if err != nil {
		t.Fatalf("first call unexpected error: %v", err)
	}
	hash2, err := mcpchainhash.MCPChainHash("ROOT/a")
	if err != nil {
		t.Fatalf("second call unexpected error: %v", err)
	}
	if hash1 != hash2 {
		t.Errorf("expected deterministic hash, got %q and %q", hash1, hash2)
	}
}

func TestMCPChainHash_MatchesLoadChainHash(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "# ROOT\n\n# Public\n\nPublic content.\n")
	testWriteNode(t, "ROOT/a", "---\noutput: out/a.go\n---\n# ROOT/a\n\n# Public\n\nLeaf content.\n")

	hashA, err := mcpchainhash.MCPChainHash("ROOT/a")
	if err != nil {
		t.Fatalf("MCPChainHash unexpected error: %v", err)
	}

	result, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("MCPLoadChain unexpected error: %v", err)
	}

	if hashA != result.ChainHash {
		t.Errorf("hash mismatch: MCPChainHash=%q, MCPLoadChain.ChainHash=%q", hashA, result.ChainHash)
	}
}

func TestMCPChainHash_InvalidLogicalName(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := mcpchainhash.MCPChainHash("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("expected ErrUnsupportedReference, got %v", err)
	}
}

func TestMCPChainHash_NonexistentNodeFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := mcpchainhash.MCPChainHash("ROOT/nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrFileUnreadable) {
		t.Errorf("expected frontmatter.ErrFileUnreadable, got %v", err)
	}
}

func TestMCPChainHash_NoOutputDeclared(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteNode(t, "ROOT", "# ROOT\n\n# Public\n\nPublic content.\n")
	testWriteNode(t, "ROOT/a", "# ROOT/a\n\n# Public\n\nLeaf content.\n")

	_, err := mcpchainhash.MCPChainHash("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpchainhash.ErrNoOutput) {
		t.Errorf("expected ErrNoOutput, got %v", err)
	}
}
