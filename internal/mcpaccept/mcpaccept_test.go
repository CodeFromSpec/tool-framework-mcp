// code-from-spec: SPEC/golang/test/cases/mcp_tools/accept@a2mcJWcEKYVETgqQbTns6KZsvzc
package mcpaccept_test

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"strings"
	"testing"

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
	b.Write()

	ba := testutils.CreateSpecNode(t, "SPEC/root/a")
	ba.SetOutput("out/a.go")
	ba.Write()

	fileContent := "modified content"
	writeFile(t, "out/a.go", fileContent)

	staleChecksum := "AAAAAAAAAAAAAAAAAAAAAAAAAAA"
	chainHash := "BBBBBBBBBBBBBBBBBBBBBBBBBBB"
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

func TestMCPAccept_NoManifestEntry_NotModified(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/root")
	b.Write()

	ba := testutils.CreateSpecNode(t, "SPEC/root/a")
	ba.SetOutput("out/a.go")
	ba.Write()

	writeFile(t, "out/a.go", "some content")

	_, err := mcpaccept.MCPAccept("SPEC/root/a")
	if !errors.Is(err, mcpaccept.ErrNotModified) {
		t.Fatalf("expected ErrNotModified, got %v", err)
	}
}

func TestMCPAccept_ArtifactFileDoesNotExist(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/root")
	b.Write()

	ba := testutils.CreateSpecNode(t, "SPEC/root/a")
	ba.SetOutput("out/a.go")
	ba.Write()

	writeManifestEntry(t, "ARTIFACT/root/a", "AAAAAAAAAAAAAAAAAAAAAAAAAAA", "BBBBBBBBBBBBBBBBBBBBBBBBBBB")

	_, err := mcpaccept.MCPAccept("SPEC/root/a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errors.Is(err, mcpaccept.ErrNotModified) {
		t.Fatalf("expected oslayer error, got ErrNotModified")
	}
}

func TestMCPAccept_ChecksumAlreadyMatches_NotModified(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/root")
	b.Write()

	ba := testutils.CreateSpecNode(t, "SPEC/root/a")
	ba.SetOutput("out/a.go")
	ba.Write()

	fileContent := "same content"
	writeFile(t, "out/a.go", fileContent)

	matchingChecksum := checksumOf(fileContent)
	writeManifestEntry(t, "ARTIFACT/root/a", matchingChecksum, "BBBBBBBBBBBBBBBBBBBBBBBBBBB")

	_, err := mcpaccept.MCPAccept("SPEC/root/a")
	if !errors.Is(err, mcpaccept.ErrNotModified) {
		t.Fatalf("expected ErrNotModified, got %v", err)
	}
}
