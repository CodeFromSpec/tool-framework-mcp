// code-from-spec: ROOT/golang/tests/mcp_tools/chain_hash@7B6njHEKUmv1Kp5E1fZ9aKSQb1c
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
	dir := path[:strings.LastIndex(path, "/")]
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("testWriteFile MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile WriteFile: %v", err)
	}
}

func TestMCPChainHash_Returns27CharHash(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC

# Public

## Context

Some root content.
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: some/output/path.md
---
# SPEC/a

# Public

## Context

Some child content.
`)

	hash, err := mcpchainhash.MCPChainHash("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected hash length 27, got %d: %q", len(hash), hash)
	}
}

func TestMCPChainHash_Deterministic(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC

# Public

## Context

Fixed root content.
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: some/output/path.md
---
# SPEC/a

# Public

## Context

Fixed child content.
`)

	hash1, err := mcpchainhash.MCPChainHash("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error on first call: %v", err)
	}

	hash2, err := mcpchainhash.MCPChainHash("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("hash not deterministic: %q != %q", hash1, hash2)
	}
}

func TestMCPChainHash_MatchesLoadChainHash(t *testing.T) {
	tempDir := t.TempDir()
	testChdir(t, tempDir)

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC

# Public

## Context

Root content for comparison.
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `---
output: some/output/path.md
---
# SPEC/a

# Public

## Context

Child content for comparison.
`)

	chainHashResult, err := mcpchainhash.MCPChainHash("SPEC/a")
	if err != nil {
		t.Fatalf("MCPChainHash error: %v", err)
	}

	loadChainResult, err := mcploadchain.MCPLoadChain("SPEC/a")
	if err != nil {
		t.Fatalf("MCPLoadChain error: %v", err)
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

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC

# Public

## Context

Root content.
`)

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

	testWriteFile(t, "code-from-spec/_node.md", `# SPEC

# Public

## Context

Root content.
`)
	testWriteFile(t, "code-from-spec/a/_node.md", `# SPEC/a

# Public

## Context

Child content without output.
`)

	_, err := mcpchainhash.MCPChainHash("SPEC/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mcpchainhash.ErrNoOutput) {
		t.Errorf("expected ErrNoOutput, got: %v", err)
	}
}
