package mcpprunecache_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/cache"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/manifest"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/mcpprunecache"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"
)

func saveManifestWithEntries(t *testing.T, entries map[string]manifest.ManifestEntry) {
	t.Helper()
	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("OpenManifest failed: %v", err)
	}
	defer func() { _ = m.Discard() }()
	for key, entry := range entries {
		m.Entries[key] = entry
	}
	if err := m.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
}

func TestMCPPruneCache_DeletesUnreferencedChainFile(t *testing.T) {
	testutils.Chdir(t)

	chainHashA := "aaaaaaaaaaaaaaaaaaaaaaaaaa1"
	chainHashB := "bbbbbbbbbbbbbbbbbbbbbbbbbbb"
	contentHash1 := "ccccccccccccccccccccccccccc"
	contentHash2 := "ddddddddddddddddddddddddddd"

	if err := cache.WriteContent(contentHash1, "content1"); err != nil {
		t.Fatalf("WriteContent failed: %v", err)
	}
	if err := cache.WriteContent(contentHash2, "content2"); err != nil {
		t.Fatalf("WriteContent failed: %v", err)
	}
	if err := cache.WriteChain(chainHashA, []chainhash.ContentHash{
		{Label: "SPEC/a", Hash: contentHash1},
	}); err != nil {
		t.Fatalf("WriteChain failed: %v", err)
	}
	if err := cache.WriteChain(chainHashB, []chainhash.ContentHash{
		{Label: "SPEC/b", Hash: contentHash2},
	}); err != nil {
		t.Fatalf("WriteChain failed: %v", err)
	}

	saveManifestWithEntries(t, map[string]manifest.ManifestEntry{
		"ARTIFACT/example": {
			Path:      "internal/example.go",
			Checksum:  "checksum1",
			ChainHash: chainHashA,
		},
	})

	summary, err := mcpprunecache.MCPPruneCache()
	if err != nil {
		t.Fatalf("MCPPruneCache failed: %v", err)
	}

	if !strings.Contains(summary, "1 chain files deleted") {
		t.Errorf("expected summary to contain '1 chain files deleted', got %q", summary)
	}

	if _, err := cache.ReadChain(chainHashB); !errors.Is(err, cache.ErrNotFound) {
		t.Errorf("expected ErrNotFound for chain %s, got %v", chainHashB, err)
	}

	if _, err := cache.ReadChain(chainHashA); err != nil {
		t.Errorf("expected chain %s to still exist, got %v", chainHashA, err)
	}
}

func TestMCPPruneCache_DeletesUnreferencedContentFile(t *testing.T) {
	testutils.Chdir(t)

	chainHashA := "aaaaaaaaaaaaaaaaaaaaaaaaaa1"
	contentHashC := "ccccccccccccccccccccccccccc"
	contentHashD := "ddddddddddddddddddddddddddd"

	if err := cache.WriteContent(contentHashC, "referenced content"); err != nil {
		t.Fatalf("WriteContent failed: %v", err)
	}
	if err := cache.WriteContent(contentHashD, "unreferenced content"); err != nil {
		t.Fatalf("WriteContent failed: %v", err)
	}
	if err := cache.WriteChain(chainHashA, []chainhash.ContentHash{
		{Label: "SPEC/a", Hash: contentHashC},
	}); err != nil {
		t.Fatalf("WriteChain failed: %v", err)
	}

	saveManifestWithEntries(t, map[string]manifest.ManifestEntry{
		"ARTIFACT/example": {
			Path:      "internal/example.go",
			Checksum:  "checksum1",
			ChainHash: chainHashA,
		},
	})

	summary, err := mcpprunecache.MCPPruneCache()
	if err != nil {
		t.Fatalf("MCPPruneCache failed: %v", err)
	}

	if !strings.Contains(summary, "1 content files deleted") {
		t.Errorf("expected summary to contain '1 content files deleted', got %q", summary)
	}

	if _, err := cache.ReadContent(contentHashD); !errors.Is(err, cache.ErrNotFound) {
		t.Errorf("expected ErrNotFound for content %s, got %v", contentHashD, err)
	}

	if _, err := cache.ReadContent(contentHashC); err != nil {
		t.Errorf("expected content %s to still exist, got %v", contentHashC, err)
	}
}

func TestMCPPruneCache_NothingToPrune(t *testing.T) {
	testutils.Chdir(t)

	chainHashA := "aaaaaaaaaaaaaaaaaaaaaaaaaa1"
	contentHashC := "ccccccccccccccccccccccccccc"

	if err := cache.WriteContent(contentHashC, "referenced content"); err != nil {
		t.Fatalf("WriteContent failed: %v", err)
	}
	if err := cache.WriteChain(chainHashA, []chainhash.ContentHash{
		{Label: "SPEC/a", Hash: contentHashC},
	}); err != nil {
		t.Fatalf("WriteChain failed: %v", err)
	}

	saveManifestWithEntries(t, map[string]manifest.ManifestEntry{
		"ARTIFACT/example": {
			Path:      "internal/example.go",
			Checksum:  "checksum1",
			ChainHash: chainHashA,
		},
	})

	summary, err := mcpprunecache.MCPPruneCache()
	if err != nil {
		t.Fatalf("MCPPruneCache failed: %v", err)
	}

	if !strings.Contains(summary, "0 chain files deleted, 0 content files deleted") {
		t.Errorf("expected summary to indicate nothing deleted, got %q", summary)
	}
}

func TestMCPPruneCache_EmptyCache(t *testing.T) {
	testutils.Chdir(t)

	chainHashA := "aaaaaaaaaaaaaaaaaaaaaaaaaa1"

	saveManifestWithEntries(t, map[string]manifest.ManifestEntry{
		"ARTIFACT/example": {
			Path:      "internal/example.go",
			Checksum:  "checksum1",
			ChainHash: chainHashA,
		},
	})

	summary, err := mcpprunecache.MCPPruneCache()
	if err != nil {
		t.Fatalf("MCPPruneCache failed: %v", err)
	}

	if !strings.Contains(summary, "0 chain files deleted, 0 content files deleted") {
		t.Errorf("expected summary to indicate nothing deleted, got %q", summary)
	}
}

func TestMCPPruneCache_EmptyManifestPrunesAll(t *testing.T) {
	testutils.Chdir(t)

	chainHashA := "aaaaaaaaaaaaaaaaaaaaaaaaaa1"
	contentHashC := "ccccccccccccccccccccccccccc"

	if err := cache.WriteContent(contentHashC, "content"); err != nil {
		t.Fatalf("WriteContent failed: %v", err)
	}
	if err := cache.WriteChain(chainHashA, []chainhash.ContentHash{
		{Label: "SPEC/a", Hash: contentHashC},
	}); err != nil {
		t.Fatalf("WriteChain failed: %v", err)
	}

	saveManifestWithEntries(t, map[string]manifest.ManifestEntry{})

	summary, err := mcpprunecache.MCPPruneCache()
	if err != nil {
		t.Fatalf("MCPPruneCache failed: %v", err)
	}

	if !strings.Contains(summary, "1 chain files deleted") {
		t.Errorf("expected summary to contain '1 chain files deleted', got %q", summary)
	}
	if !strings.Contains(summary, "1 content files deleted") {
		t.Errorf("expected summary to contain '1 content files deleted', got %q", summary)
	}

	if _, err := cache.ReadChain(chainHashA); !errors.Is(err, cache.ErrNotFound) {
		t.Errorf("expected ErrNotFound for chain %s, got %v", chainHashA, err)
	}
	if _, err := cache.ReadContent(contentHashC); !errors.Is(err, cache.ErrNotFound) {
		t.Errorf("expected ErrNotFound for content %s, got %v", contentHashC, err)
	}
}
