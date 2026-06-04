// code-from-spec: ROOT/golang/tests/mcp_tools/chain_hash@-W-guKsO5kHez10QCOKiAiifhsk
package mcpchainhash_test

import (
	"errors"
	"os"
	"strings"
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
	if err := os.MkdirAll(path[:strings.LastIndex(path, "/")], 0755); err != nil {
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

func TestReturns27CharacterHash(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n## Public\n\nSome public content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: some/output/file.go\n---\n# ROOT/a\n\nLeaf node content.\n")

	result, err := mcpchainhash.MCPChainHash("ROOT/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 27 {
		t.Errorf("expected 27-character hash, got %d characters: %q", len(result), result)
	}
}

func TestHashIsDeterministic(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n## Public\n\nSome public content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: some/output/file.go\n---\n# ROOT/a\n\nLeaf node content.\n")

	firstHash, err := mcpchainhash.MCPChainHash("ROOT/a")
	if err != nil {
		t.Fatalf("first call unexpected error: %v", err)
	}

	secondHash, err := mcpchainhash.MCPChainHash("ROOT/a")
	if err != nil {
		t.Fatalf("second call unexpected error: %v", err)
	}

	if firstHash != secondHash {
		t.Errorf("hash is not deterministic: first=%q second=%q", firstHash, secondHash)
	}
}

func TestHashMatchesLoadChainHash(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n## Public\n\nSome public content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "---\noutput: some/output/file.go\n---\n# ROOT/a\n\nLeaf node content.\n")

	chainHashResult, err := mcpchainhash.MCPChainHash("ROOT/a")
	if err != nil {
		t.Fatalf("MCPChainHash unexpected error: %v", err)
	}

	loadChainResult, err := mcploadchain.MCPLoadChain("ROOT/a")
	if err != nil {
		t.Fatalf("MCPLoadChain unexpected error: %v", err)
	}

	var loadChainHash string
	for _, line := range strings.Split(loadChainResult, "\n") {
		if strings.HasPrefix(line, "chain_hash: ") {
			loadChainHash = strings.TrimPrefix(line, "chain_hash: ")
			break
		}
	}

	if loadChainHash == "" {
		t.Fatalf("could not extract chain_hash from MCPLoadChain result")
	}

	if chainHashResult != loadChainHash {
		t.Errorf("hash mismatch: MCPChainHash=%q load_chain chain_hash=%q", chainHashResult, loadChainHash)
	}
}

func TestInvalidLogicalNameNotRoot(t *testing.T) {
	_, err := mcpchainhash.MCPChainHash("INVALID/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("expected ErrUnsupportedReference, got: %v", err)
	}
}

func TestNonexistentNodeFile(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n## Public\n\nSome public content.\n")

	_, err := mcpchainhash.MCPChainHash("ROOT/nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestNoOutputDeclared(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", "# ROOT\n\n## Public\n\nSome public content.\n")
	testWriteFile(t, "code-from-spec/a/_node.md", "# ROOT/a\n\nLeaf node without output field.\n")

	_, err := mcpchainhash.MCPChainHash("ROOT/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpchainhash.ErrNoOutput) {
		t.Errorf("expected ErrNoOutput, got: %v", err)
	}
}
