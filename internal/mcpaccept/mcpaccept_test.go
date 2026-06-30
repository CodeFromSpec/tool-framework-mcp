package mcpaccept_test

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/manifest"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/mcpaccept"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"
)

func checksumOf(content string) string {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	if !strings.HasSuffix(normalized, "\n") {
		normalized += "\n"
	}
	h := sha1.Sum([]byte(normalized))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func currentChainHash(t *testing.T, logicalName string) string {
	t.Helper()
	chain, err := chainresolver.ChainResolve(logicalName)
	if err != nil {
		t.Fatalf("ChainResolve: %v", err)
	}
	hash, _, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute: %v", err)
	}
	return hash
}

func writeManifestEntry(t *testing.T, logicalName, checksum, chainHash string) {
	t.Helper()
	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("OpenManifest: %v", err)
	}
	defer func() { _ = m.Discard() }()
	m.Entries[logicalName] = manifest.ManifestEntry{
		Path:      "",
		Checksum:  checksum,
		ChainHash: chainHash,
	}
	if err := m.Save(); err != nil {
		t.Fatalf("Save manifest: %v", err)
	}
}

func writeEmptyManifest(t *testing.T) {
	t.Helper()
	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("OpenManifest: %v", err)
	}
	defer func() { _ = m.Discard() }()
	if err := m.Save(); err != nil {
		t.Fatalf("Save manifest: %v", err)
	}
}

func writeFile(t *testing.T, cfsPath oslayer.CfsPath, content string) {
	t.Helper()
	f, err := oslayer.OpenFile(cfsPath, "overwrite", 5000)
	if err != nil {
		t.Fatalf("OpenFile %s: %v", cfsPath, err)
	}
	defer f.Close()
	if err := f.Write(content); err != nil {
		t.Fatalf("Write %s: %v", cfsPath, err)
	}
}

func TestMCPAccept_AcceptsModifiedArtifact(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/root")
	b.SetPublic("## Context\ncontent")
	b.Write()

	ba := testutils.CreateSpecNode(t, "SPEC/root/a")
	ba.SetOutput("out/a.go")
	ba.Write()

	fileContent := "modified content"
	writeFile(t, "out/a.go", fileContent)

	chainHash := currentChainHash(t, "SPEC/root/a")
	staleChecksum := "AAAAAAAAAAAAAAAAAAAAAAAAAAA"
	writeManifestEntry(t, "ARTIFACT/root/a", staleChecksum, chainHash)

	result, err := mcpaccept.MCPAccept("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "accepted out/a.go" {
		t.Fatalf("expected 'accepted out/a.go', got %q", result)
	}

	m, err := manifest.OpenManifest(true)
	if err != nil {
		t.Fatalf("OpenManifest read: %v", err)
	}
	entry, ok := m.Entries["ARTIFACT/root/a"]
	if !ok {
		t.Fatal("manifest entry not found after accept")
	}
	expected := checksumOf(fileContent)
	if entry.Checksum != expected {
		t.Fatalf("expected checksum %q, got %q", expected, entry.Checksum)
	}
	if entry.ChainHash != chainHash {
		t.Fatalf("expected chain hash %q unchanged, got %q", chainHash, entry.ChainHash)
	}
}

func TestMCPAccept_AcceptsStaleArtifact(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/root")
	b.SetPublic("## Context\ncontent")
	b.Write()

	ba := testutils.CreateSpecNode(t, "SPEC/root/a")
	ba.SetOutput("out/a.go")
	ba.Write()

	fileContent := "artifact content"
	writeFile(t, "out/a.go", fileContent)

	correctChecksum := checksumOf(fileContent)
	staleChainHash := "AAAAAAAAAAAAAAAAAAAAAAAAAAA"
	writeManifestEntry(t, "ARTIFACT/root/a", correctChecksum, staleChainHash)

	result, err := mcpaccept.MCPAccept("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "accepted out/a.go" {
		t.Fatalf("expected 'accepted out/a.go', got %q", result)
	}

	m, err := manifest.OpenManifest(true)
	if err != nil {
		t.Fatalf("OpenManifest read: %v", err)
	}
	entry, ok := m.Entries["ARTIFACT/root/a"]
	if !ok {
		t.Fatal("manifest entry not found after accept")
	}
	if entry.Checksum != correctChecksum {
		t.Fatalf("expected checksum %q unchanged, got %q", correctChecksum, entry.Checksum)
	}
	expectedChainHash := currentChainHash(t, "SPEC/root/a")
	if entry.ChainHash != expectedChainHash {
		t.Fatalf("expected chain hash %q, got %q", expectedChainHash, entry.ChainHash)
	}
}

func TestMCPAccept_CreatesEntryWhenNoneExists(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/root")
	b.SetPublic("## Context\ncontent")
	b.Write()

	ba := testutils.CreateSpecNode(t, "SPEC/root/a")
	ba.SetOutput("out/a.go")
	ba.Write()

	fileContent := "new content"
	writeFile(t, "out/a.go", fileContent)

	writeEmptyManifest(t)

	result, err := mcpaccept.MCPAccept("SPEC/root/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "accepted out/a.go" {
		t.Fatalf("expected 'accepted out/a.go', got %q", result)
	}

	m, err := manifest.OpenManifest(true)
	if err != nil {
		t.Fatalf("OpenManifest read: %v", err)
	}
	entry, ok := m.Entries["ARTIFACT/root/a"]
	if !ok {
		t.Fatal("manifest entry not found after accept")
	}
	expectedChecksum := checksumOf(fileContent)
	if entry.Checksum != expectedChecksum {
		t.Fatalf("expected checksum %q, got %q", expectedChecksum, entry.Checksum)
	}
	expectedChainHash := currentChainHash(t, "SPEC/root/a")
	if entry.ChainHash != expectedChainHash {
		t.Fatalf("expected chain hash %q, got %q", expectedChainHash, entry.ChainHash)
	}
}

func TestMCPAccept_NotASpecReference(t *testing.T) {
	testutils.Chdir(t)

	_, err := mcpaccept.MCPAccept("ARTIFACT/root/a")
	if !errors.Is(err, mcpaccept.ErrNotASpecReference) {
		t.Fatalf("expected ErrNotASpecReference, got %v", err)
	}
}

func TestMCPAccept_NonexistentNodeFile(t *testing.T) {
	testutils.Chdir(t)

	_, err := mcpaccept.MCPAccept("SPEC/root/missing")
	if !errors.Is(err, mcpaccept.ErrUnreadableFrontmatter) {
		t.Fatalf("expected ErrUnreadableFrontmatter, got %v", err)
	}
}

func TestMCPAccept_NoOutputDeclared(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/root")
	b.Write()

	ba := testutils.CreateSpecNode(t, "SPEC/root/a")
	ba.Write()

	_, err := mcpaccept.MCPAccept("SPEC/root/a")
	if !errors.Is(err, mcpaccept.ErrNoOutput) {
		t.Fatalf("expected ErrNoOutput, got %v", err)
	}
}

func TestMCPAccept_ArtifactFileDoesNotExist(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/root")
	b.Write()

	ba := testutils.CreateSpecNode(t, "SPEC/root/a")
	ba.SetOutput("out/a.go")
	ba.Write()

	_, err := mcpaccept.MCPAccept("SPEC/root/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errors.Is(err, mcpaccept.ErrAlreadyUpToDate) {
		t.Fatalf("expected oslayer error, got ErrAlreadyUpToDate")
	}
}

func TestMCPAccept_AlreadyUpToDate(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/root")
	b.SetPublic("## Context\ncontent")
	b.Write()

	ba := testutils.CreateSpecNode(t, "SPEC/root/a")
	ba.SetOutput("out/a.go")
	ba.Write()

	fileContent := "same content"
	writeFile(t, "out/a.go", fileContent)

	matchingChecksum := checksumOf(fileContent)
	matchingChainHash := currentChainHash(t, "SPEC/root/a")
	writeManifestEntry(t, "ARTIFACT/root/a", matchingChecksum, matchingChainHash)

	_, err := mcpaccept.MCPAccept("SPEC/root/a")
	if !errors.Is(err, mcpaccept.ErrAlreadyUpToDate) {
		t.Fatalf("expected ErrAlreadyUpToDate, got %v", err)
	}
}
