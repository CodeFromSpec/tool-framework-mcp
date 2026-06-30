package mcpreconstructcache_test

import (
	"crypto/sha1"
	"encoding/base64"
	"os"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/cache"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/manifest"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/mcpreconstructcache"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"
)

func computeChecksum(content string) string {
	sum := sha1.Sum([]byte(content))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func setupSingleEntryManifest(t *testing.T) (string, []chainhash.ContentHash, string) {
	t.Helper()

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.SetPublic("## Context\ncontext content")
	root.Write()

	a := testutils.CreateSpecNode(t, "SPEC/root/a")
	a.SetOutput("out/a.go")
	a.SetPublic("## Interface\ninterface content")
	a.Write()

	artifactContent := "package out\n"
	if err := os.MkdirAll("out", 0o755); err != nil {
		t.Fatalf("failed to create out directory: %v", err)
	}
	if err := os.WriteFile("out/a.go", []byte(artifactContent), 0o644); err != nil {
		t.Fatalf("failed to write artifact file: %v", err)
	}

	chain, err := chainresolver.ChainResolve("SPEC/root/a")
	if err != nil {
		t.Fatalf("ChainResolve failed: %v", err)
	}
	chainHash, positions, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute failed: %v", err)
	}

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("OpenManifest failed: %v", err)
	}
	defer func() { _ = m.Discard() }()
	m.Entries["ARTIFACT/root/a"] = manifest.ManifestEntry{
		Path:      "out/a.go",
		Checksum:  computeChecksum(artifactContent),
		ChainHash: chainHash,
	}
	if err := m.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	return chainHash, positions, artifactContent
}

func TestMCPReconstructCache_PopulatesCacheForSingleEntry(t *testing.T) {
	testutils.Chdir(t)

	chainHash, positions, _ := setupSingleEntryManifest(t)

	summary, err := mcpreconstructcache.MCPReconstructCache()
	if err != nil {
		t.Fatalf("MCPReconstructCache failed: %v", err)
	}
	if !strings.Contains(summary, "1 entries processed") {
		t.Fatalf("expected summary to contain '1 entries processed', got %q", summary)
	}

	if _, err := cache.ReadChain(chainHash); err != nil {
		t.Fatalf("ReadChain failed: %v", err)
	}
	if len(positions) == 0 {
		t.Fatalf("expected at least one content hash position")
	}
	if _, err := cache.ReadContent(positions[0].Hash); err != nil {
		t.Fatalf("ReadContent failed: %v", err)
	}
}

func TestMCPReconstructCache_IdempotentSkipsExistingCacheFiles(t *testing.T) {
	testutils.Chdir(t)

	setupSingleEntryManifest(t)

	if _, err := mcpreconstructcache.MCPReconstructCache(); err != nil {
		t.Fatalf("first MCPReconstructCache failed: %v", err)
	}

	summary, err := mcpreconstructcache.MCPReconstructCache()
	if err != nil {
		t.Fatalf("second MCPReconstructCache failed: %v", err)
	}
	if !strings.Contains(summary, "0 content files written") {
		t.Fatalf("expected summary to contain '0 content files written', got %q", summary)
	}
	if !strings.Contains(summary, "0 chain files written") {
		t.Fatalf("expected summary to contain '0 chain files written', got %q", summary)
	}
}

func TestMCPReconstructCache_SkipsDeletedNodesGracefully(t *testing.T) {
	testutils.Chdir(t)

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("OpenManifest failed: %v", err)
	}
	defer func() { _ = m.Discard() }()
	m.Entries["ARTIFACT/root/deleted"] = manifest.ManifestEntry{
		Path:      "out/deleted.go",
		Checksum:  "AAAAAAAAAAAAAAAAAAAAAAAAAAA",
		ChainHash: "AAAAAAAAAAAAAAAAAAAAAAAAAAA",
	}
	if err := m.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if _, err := mcpreconstructcache.MCPReconstructCache(); err != nil {
		t.Fatalf("MCPReconstructCache failed: %v", err)
	}
}

func TestMCPReconstructCache_NoManifestReturnsError(t *testing.T) {
	testutils.Chdir(t)

	if _, err := mcpreconstructcache.MCPReconstructCache(); err == nil {
		t.Fatalf("expected error when no manifest exists")
	}
}

func TestMCPReconstructCache_EmptyManifestZeroEntriesProcessed(t *testing.T) {
	testutils.Chdir(t)

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("OpenManifest failed: %v", err)
	}
	defer func() { _ = m.Discard() }()
	if err := m.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	summary, err := mcpreconstructcache.MCPReconstructCache()
	if err != nil {
		t.Fatalf("MCPReconstructCache failed: %v", err)
	}
	if !strings.Contains(summary, "0 entries processed") {
		t.Fatalf("expected summary to contain '0 entries processed', got %q", summary)
	}
}
