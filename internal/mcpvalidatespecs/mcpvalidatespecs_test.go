// code-from-spec: SPEC/golang/test/cases/mcp_tools/validate_specs@79K1ibU9VbPZH2kYE6nQHEw7Jzo
package mcpvalidatespecs_test

import (
	"crypto/sha1"
	"encoding/base64"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/manifest"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/mcpvalidatespecs"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"
)

func fileChecksum(content string) string {
	h := sha1.New()
	h.Write([]byte(content))
	sum := h.Sum(nil)
	encoded := base64.RawURLEncoding.EncodeToString(sum)
	return encoded[:27]
}

func computeChainHash(t *testing.T, logicalName string) string {
	t.Helper()
	chain, err := chainresolver.ChainResolve(logicalName)
	if err != nil {
		t.Fatalf("ChainResolve(%q): %v", logicalName, err)
	}
	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute: %v", err)
	}
	return hash
}

func writeManifestEntry(t *testing.T, artifactLogicalName, path, checksum, chainHash string) {
	t.Helper()
	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("OpenManifest: %v", err)
	}
	defer func() { _ = m.Discard() }()
	m.Entries[artifactLogicalName] = manifest.ManifestEntry{
		Path:      path,
		Checksum:  checksum,
		ChainHash: chainHash,
	}
	if err := m.Save(); err != nil {
		t.Fatalf("manifest.Save: %v", err)
	}
}

func createRootNode(t *testing.T) {
	t.Helper()
	b := testutils.CreateSpecNode(t, "SPEC/root")
	b.SetPublic("## Context\nroot context content")
	b.Write()
}

func TestCleanTree(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	b := testutils.CreateSpecNode(t, "SPEC/root/a")
	b.SetOutput("out/a.go")
	b.Write()

	fileContent := "package a\n"
	if err := os.MkdirAll("out", 0755); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	if err := os.WriteFile("out/a.go", []byte(fileContent), 0644); err != nil {
		t.Fatalf("write out/a.go: %v", err)
	}

	chainHash := computeChainHash(t, "SPEC/root/a")
	checksum := fileChecksum(fileContent)

	writeManifestEntry(t, "ARTIFACT/root/a", "out/a.go", checksum, chainHash)

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) != 0 {
		t.Errorf("expected no format errors, got %v", report.FormatErrors)
	}
	if len(report.Cycles) != 0 {
		t.Errorf("expected no cycles, got %v", report.Cycles)
	}
	if len(report.Staleness) != 0 {
		t.Errorf("expected no staleness entries, got %v", report.Staleness)
	}
}

func TestStaleArtifact(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	b := testutils.CreateSpecNode(t, "SPEC/root/a")
	b.SetOutput("out/a.go")
	b.Write()

	fileContent := "package a\n"
	if err := os.MkdirAll("out", 0755); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	if err := os.WriteFile("out/a.go", []byte(fileContent), 0644); err != nil {
		t.Fatalf("write out/a.go: %v", err)
	}

	checksum := fileChecksum(fileContent)
	staleHash := "AAAAAAAAAAAAAAAAAAAAAAAAAAA"

	writeManifestEntry(t, "ARTIFACT/root/a", "out/a.go", checksum, staleHash)

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, s := range report.Staleness {
		if s.Node == "SPEC/root/a" && s.Status == "stale" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected staleness entry for SPEC/root/a with status 'stale', got %v", report.Staleness)
	}
}

func TestMissingArtifactNoManifestEntry(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	b := testutils.CreateSpecNode(t, "SPEC/root/a")
	b.SetOutput("out/a.go")
	b.Write()

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, s := range report.Staleness {
		if s.Node == "SPEC/root/a" && s.Status == "missing" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected staleness entry for SPEC/root/a with status 'missing', got %v", report.Staleness)
	}
}

func TestMissingArtifactFileDoesNotExist(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	b := testutils.CreateSpecNode(t, "SPEC/root/a")
	b.SetOutput("out/a.go")
	b.Write()

	chainHash := computeChainHash(t, "SPEC/root/a")
	placeholderChecksum := "AAAAAAAAAAAAAAAAAAAAAAAAAAA"

	writeManifestEntry(t, "ARTIFACT/root/a", "out/a.go", placeholderChecksum, chainHash)

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, s := range report.Staleness {
		if s.Node == "SPEC/root/a" && s.Status == "missing" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected staleness entry for SPEC/root/a with status 'missing', got %v", report.Staleness)
	}
}

func TestModifiedArtifact(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	b := testutils.CreateSpecNode(t, "SPEC/root/a")
	b.SetOutput("out/a.go")
	b.Write()

	originalContent := "package a // original\n"
	if err := os.MkdirAll("out", 0755); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	if err := os.WriteFile("out/a.go", []byte(originalContent), 0644); err != nil {
		t.Fatalf("write out/a.go: %v", err)
	}

	chainHash := computeChainHash(t, "SPEC/root/a")
	originalChecksum := fileChecksum(originalContent)

	writeManifestEntry(t, "ARTIFACT/root/a", "out/a.go", originalChecksum, chainHash)

	modifiedContent := "package a // modified\n"
	if err := os.WriteFile("out/a.go", []byte(modifiedContent), 0644); err != nil {
		t.Fatalf("overwrite out/a.go: %v", err)
	}

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, s := range report.Staleness {
		if s.Node == "SPEC/root/a" && s.Status == "modified" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected staleness entry for SPEC/root/a with status 'modified', got %v", report.Staleness)
	}
}

func TestOrphanManifestEntry(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/root")
	b.Write()

	writeManifestEntry(t, "ARTIFACT/root/deleted", "out/deleted.go", "AAAAAAAAAAAAAAAAAAAAAAAAAAA", "AAAAAAAAAAAAAAAAAAAAAAAAAAA")

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, s := range report.Staleness {
		if s.Status == "orphan" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected staleness entry with status 'orphan', got %v", report.Staleness)
	}
}

func TestStalenessEntriesIncludeRank(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	ba := testutils.CreateSpecNode(t, "SPEC/root/a")
	ba.SetOutput("out/a.go")
	ba.Write()

	bb := testutils.CreateSpecNode(t, "SPEC/root/b")
	bb.SetOutput("out/b.go")
	bb.AddDependsOn("SPEC/root/a")
	bb.Write()

	report := mcpvalidatespecs.MCPValidateSpecs()

	var rankA, rankB int
	foundA, foundB := false, false
	for _, s := range report.Staleness {
		if s.Node == "SPEC/root/a" {
			rankA = s.Rank
			foundA = true
		}
		if s.Node == "SPEC/root/b" {
			rankB = s.Rank
			foundB = true
		}
	}
	if !foundA {
		t.Fatal("expected staleness entry for SPEC/root/a")
	}
	if !foundB {
		t.Fatal("expected staleness entry for SPEC/root/b")
	}
	if rankA >= rankB {
		t.Errorf("expected rank of SPEC/root/a (%d) < rank of SPEC/root/b (%d)", rankA, rankB)
	}
}

func TestStalenessOrderedByRankThenName(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	bz := testutils.CreateSpecNode(t, "SPEC/root/z")
	bz.SetOutput("out/z.go")
	bz.Write()

	ba := testutils.CreateSpecNode(t, "SPEC/root/a")
	ba.SetOutput("out/a.go")
	ba.Write()

	report := mcpvalidatespecs.MCPValidateSpecs()

	idxA, idxZ := -1, -1
	for i, s := range report.Staleness {
		if s.Node == "SPEC/root/a" {
			idxA = i
		}
		if s.Node == "SPEC/root/z" {
			idxZ = i
		}
	}
	if idxA == -1 {
		t.Fatal("expected staleness entry for SPEC/root/a")
	}
	if idxZ == -1 {
		t.Fatal("expected staleness entry for SPEC/root/z")
	}
	if idxA >= idxZ {
		t.Errorf("expected SPEC/root/a (idx %d) before SPEC/root/z (idx %d)", idxA, idxZ)
	}
}

func TestFormatErrorInvalidDependsOn(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	b := testutils.CreateSpecNode(t, "SPEC/root/a")
	b.AddDependsOn("SPEC/root/missing")
	b.Write()

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, fe := range report.FormatErrors {
		if fe.Node == "SPEC/root/a" && fe.Rule == "dependency_targets" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected format error for SPEC/root/a with rule 'dependency_targets', got %v", report.FormatErrors)
	}
}

func TestFormatErrorParseFailure(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	testutils.WriteRawNode(t, "SPEC/root/a", "plain text before any heading\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, fe := range report.FormatErrors {
		if fe.Node == "SPEC/root/a" && fe.Rule == "parse" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected format error for SPEC/root/a with rule 'parse', got %v", report.FormatErrors)
	}
}

func TestContinuesAfterParseFailure(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	testutils.WriteRawNode(t, "SPEC/root/a", "plain text before any heading\n")

	bb := testutils.CreateSpecNode(t, "SPEC/root/b")
	bb.SetOutput("out/b.go")
	bb.Write()

	report := mcpvalidatespecs.MCPValidateSpecs()

	foundParseError := false
	for _, fe := range report.FormatErrors {
		if fe.Node == "SPEC/root/a" {
			foundParseError = true
			break
		}
	}
	if !foundParseError {
		t.Errorf("expected format error for SPEC/root/a, got %v", report.FormatErrors)
	}

	foundMissing := false
	for _, s := range report.Staleness {
		if s.Node == "SPEC/root/b" && s.Status == "missing" {
			foundMissing = true
			break
		}
	}
	if !foundMissing {
		t.Errorf("expected staleness entry for SPEC/root/b with status 'missing', got %v", report.Staleness)
	}
}

func TestSimpleCycleDetected(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	ba := testutils.CreateSpecNode(t, "SPEC/root/a")
	ba.AddDependsOn("SPEC/root/b")
	ba.Write()

	bb := testutils.CreateSpecNode(t, "SPEC/root/b")
	bb.AddDependsOn("SPEC/root/a")
	bb.Write()

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Cycles) == 0 {
		t.Fatal("expected cycles to be non-empty")
	}

	foundCycleMember := false
	for _, c := range report.Cycles {
		if c == "SPEC/root/a" || c == "SPEC/root/b" {
			foundCycleMember = true
			break
		}
	}
	if !foundCycleMember {
		t.Errorf("expected cycles to contain SPEC/root/a or SPEC/root/b, got %v", report.Cycles)
	}
}

func TestRankingSkippedWhenFormatErrorsExist(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	ba := testutils.CreateSpecNode(t, "SPEC/root/a")
	ba.AddDependsOn("SPEC/root/missing")
	ba.Write()

	bb := testutils.CreateSpecNode(t, "SPEC/root/b")
	bb.SetOutput("out/b.go")
	bb.Write()

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) == 0 {
		t.Fatal("expected format errors to be non-empty")
	}

	for _, s := range report.Staleness {
		if s.Node == "SPEC/root/b" {
			if s.Rank != 0 {
				t.Errorf("expected rank 0 for SPEC/root/b when format errors exist, got %d", s.Rank)
			}
		}
	}
}

func TestEmptySpecTreeScanFails(t *testing.T) {
	testutils.Chdir(t)

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, fe := range report.FormatErrors {
		if fe.Rule == "scan" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected format error with rule 'scan', got %v", report.FormatErrors)
	}
	if len(report.Cycles) != 0 {
		t.Errorf("expected no cycles, got %v", report.Cycles)
	}
	if len(report.Staleness) != 0 {
		t.Errorf("expected no staleness entries, got %v", report.Staleness)
	}
}

func TestNodeWithNoOutputNotInStaleness(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	b := testutils.CreateSpecNode(t, "SPEC/root/a")
	b.Write()

	report := mcpvalidatespecs.MCPValidateSpecs()

	for _, s := range report.Staleness {
		if s.Node == "SPEC/root/a" {
			t.Errorf("expected no staleness entry for SPEC/root/a (no output), got %v", s)
		}
	}
}

func TestNoManifestFileAllArtifactsMissing(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	b := testutils.CreateSpecNode(t, "SPEC/root/a")
	b.SetOutput("out/a.go")
	b.Write()

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, s := range report.Staleness {
		if s.Node == "SPEC/root/a" && s.Status == "missing" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected staleness entry for SPEC/root/a with status 'missing', got %v", report.Staleness)
	}
}
