// code-from-spec: SPEC/golang/test/cases/mcp_tools/write_file@lJvRaEkM9z3c3SBnIX0LHqACZ5s
package mcpwritefile_test

import (
	"errors"
	"os"
	"regexp"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/manifest"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/mcpwritefile"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"
)

var base64urlPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{27}$`)

func TestMCPWriteFile_WritesFileSuccessfully(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	node := testutils.CreateSpecNode(t, "SPEC/root/a")
	node.SetOutput("output/file.go")
	node.Write()

	result, err := mcpwritefile.MCPWriteFile("SPEC/root/a", "package main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "wrote output/file.go" {
		t.Errorf("expected %q, got %q", "wrote output/file.go", result)
	}

	data, err := os.ReadFile("output/file.go")
	if err != nil {
		t.Fatalf("file not found: %v", err)
	}
	if string(data) != "package main" {
		t.Errorf("expected content %q, got %q", "package main", string(data))
	}
}

func TestMCPWriteFile_ManifestUpdatedAfterWrite(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	node := testutils.CreateSpecNode(t, "SPEC/root/a")
	node.SetOutput("output/file.go")
	node.Write()

	_, err := mcpwritefile.MCPWriteFile("SPEC/root/a", "package main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, err := manifest.OpenManifest(true)
	if err != nil {
		t.Fatalf("failed to open manifest: %v", err)
	}

	entry, ok := m.Entries["ARTIFACT/root/a"]
	if !ok {
		t.Fatal("manifest entry ARTIFACT/root/a not found")
	}
	if entry.Path != "output/file.go" {
		t.Errorf("expected path %q, got %q", "output/file.go", entry.Path)
	}
	if !base64urlPattern.MatchString(entry.Checksum) {
		t.Errorf("checksum %q is not a 27-char base64url string", entry.Checksum)
	}
	if !base64urlPattern.MatchString(entry.ChainHash) {
		t.Errorf("chain hash %q is not a 27-char base64url string", entry.ChainHash)
	}
}

func TestMCPWriteFile_CreatesIntermediateDirectories(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	node := testutils.CreateSpecNode(t, "SPEC/root/a")
	node.SetOutput("deep/nested/dir/file.go")
	node.Write()

	_, err := mcpwritefile.MCPWriteFile("SPEC/root/a", "package main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat("deep/nested/dir/file.go"); err != nil {
		t.Errorf("file not found: %v", err)
	}
}

func TestMCPWriteFile_OverwritesExistingFile(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	node := testutils.CreateSpecNode(t, "SPEC/root/a")
	node.SetOutput("output/file.go")
	node.Write()

	if err := os.MkdirAll("output", 0o755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}
	if err := os.WriteFile("output/file.go", []byte("old"), 0o644); err != nil {
		t.Fatalf("failed to write initial file: %v", err)
	}

	_, err := mcpwritefile.MCPWriteFile("SPEC/root/a", "new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile("output/file.go")
	if err != nil {
		t.Fatalf("file not found: %v", err)
	}
	if string(data) != "new" {
		t.Errorf("expected content %q, got %q", "new", string(data))
	}
}

func TestMCPWriteFile_ArtifactReference(t *testing.T) {
	testutils.Chdir(t)

	_, err := mcpwritefile.MCPWriteFile("ARTIFACT/x", "")
	if !errors.Is(err, mcpwritefile.ErrNotASpecReference) {
		t.Errorf("expected ErrNotASpecReference, got %v", err)
	}
}

func TestMCPWriteFile_WithQualifier(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	node := testutils.CreateSpecNode(t, "SPEC/root/a")
	node.SetOutput("out.go")
	node.Write()

	_, err := mcpwritefile.MCPWriteFile("SPEC/root/a(interface)", "")
	if !errors.Is(err, mcpwritefile.ErrQualifierNotAllowed) {
		t.Errorf("expected ErrQualifierNotAllowed, got %v", err)
	}
}

func TestMCPWriteFile_NonexistentNode(t *testing.T) {
	testutils.Chdir(t)

	_, err := mcpwritefile.MCPWriteFile("SPEC/missing", "")
	if !errors.Is(err, mcpwritefile.ErrUnreadableFrontmatter) {
		t.Errorf("expected ErrUnreadableFrontmatter, got %v", err)
	}
}

func TestMCPWriteFile_NoOutputDeclared(t *testing.T) {
	testutils.Chdir(t)

	root := testutils.CreateSpecNode(t, "SPEC/root")
	root.Write()

	node := testutils.CreateSpecNode(t, "SPEC/root/a")
	node.Write()

	_, err := mcpwritefile.MCPWriteFile("SPEC/root/a", "")
	if !errors.Is(err, mcpwritefile.ErrNoOutput) {
		t.Errorf("expected ErrNoOutput, got %v", err)
	}
}
