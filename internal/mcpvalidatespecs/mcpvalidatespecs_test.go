package mcpvalidatespecs_test

import (
	"crypto/sha1"
	"encoding/base64"
	"os"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/manifest"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/mcpvalidatespecs"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/spectree"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/testutils"
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
	refs, err := spectree.SpecTreeScan()
	if err != nil {
		t.Fatalf("SpecTreeScan: %v", err)
	}
	knownNodes := make([]string, len(refs))
	for i, ref := range refs {
		knownNodes[i] = ref.LogicalName
	}
	chain, err := chainresolver.ChainResolve(logicalName, knownNodes)
	if err != nil {
		t.Fatalf("ChainResolve(%q): %v", logicalName, err)
	}
	hash, _, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("ChainHashCompute: %v", err)
	}
	return hash
}

func writeManifestEntry(t *testing.T, artifactLogicalName, path, checksum, chainHash, result string) {
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
		Result:    result,
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
	b.SetType("artifact")
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

	writeManifestEntry(t, "ARTIFACT/root/a", "out/a.go", checksum, chainHash, "")

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
	b.SetType("artifact")
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

	writeManifestEntry(t, "ARTIFACT/root/a", "out/a.go", checksum, staleHash, "")

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
	b.SetType("artifact")
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
	b.SetType("artifact")
	b.SetOutput("out/a.go")
	b.Write()

	chainHash := computeChainHash(t, "SPEC/root/a")
	placeholderChecksum := "AAAAAAAAAAAAAAAAAAAAAAAAAAA"

	writeManifestEntry(t, "ARTIFACT/root/a", "out/a.go", placeholderChecksum, chainHash, "")

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
	b.SetType("artifact")
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

	writeManifestEntry(t, "ARTIFACT/root/a", "out/a.go", originalChecksum, chainHash, "")

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

	writeManifestEntry(t, "ARTIFACT/root/deleted", "out/deleted.go", "AAAAAAAAAAAAAAAAAAAAAAAAAAA", "AAAAAAAAAAAAAAAAAAAAAAAAAAA", "")

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
	ba.SetType("artifact")
	ba.SetOutput("out/a.go")
	ba.Write()

	bb := testutils.CreateSpecNode(t, "SPEC/root/b")
	bb.SetType("artifact")
	bb.SetOutput("out/b.go")
	bb.AddImport("SPEC/root/a")
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
	bz.SetType("artifact")
	bz.SetOutput("out/z.go")
	bz.Write()

	ba := testutils.CreateSpecNode(t, "SPEC/root/a")
	ba.SetType("artifact")
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

func TestBlockedByUnsatisfiedWaitOnArtifact(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	ba := testutils.CreateSpecNode(t, "SPEC/root/a")
	ba.SetType("artifact")
	ba.SetOutput("out/a.go")
	ba.Write()

	bb := testutils.CreateSpecNode(t, "SPEC/root/b")
	bb.SetType("verdict")
	bb.AddWaitOn("ARTIFACT/root/a")
	bb.Write()

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, s := range report.Staleness {
		if s.Node == "SPEC/root/b" {
			if !s.Blocked {
				t.Errorf("expected SPEC/root/b to be blocked, got Blocked=false")
			}
			if !strings.Contains(s.BlockedBy, "ARTIFACT/root/a") {
				t.Errorf("expected BlockedBy to contain ARTIFACT/root/a, got %q", s.BlockedBy)
			}
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected staleness entry for SPEC/root/b, got %v", report.Staleness)
	}
}

func TestBlockedByUnsatisfiedWaitOnVerdictNotPassed(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	bv := testutils.CreateSpecNode(t, "SPEC/root/v")
	bv.SetType("verdict")
	bv.SetOutput("code-from-spec/root/v/verdict.md")
	bv.Write()

	verdictContent := "verdict content\n"
	if err := os.MkdirAll("code-from-spec/root/v", 0755); err != nil {
		t.Fatalf("mkdir code-from-spec/root/v: %v", err)
	}
	if err := os.WriteFile("code-from-spec/root/v/verdict.md", []byte(verdictContent), 0644); err != nil {
		t.Fatalf("write verdict.md: %v", err)
	}

	chainHash := computeChainHash(t, "SPEC/root/v")
	checksum := fileChecksum(verdictContent)

	writeManifestEntry(t, "VERDICT/root/v", "code-from-spec/root/v/verdict.md", checksum, chainHash, "fail")

	bb := testutils.CreateSpecNode(t, "SPEC/root/b")
	bb.SetType("verdict")
	bb.AddWaitOn("VERDICT/root/v")
	bb.Write()

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, s := range report.Staleness {
		if s.Node == "SPEC/root/b" {
			if !s.Blocked {
				t.Errorf("expected SPEC/root/b to be blocked, got Blocked=false")
			}
			if !strings.Contains(s.BlockedBy, "VERDICT/root/v") {
				t.Errorf("expected BlockedBy to contain VERDICT/root/v, got %q", s.BlockedBy)
			}
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected staleness entry for SPEC/root/b, got %v", report.Staleness)
	}
}

func TestNotBlockedWhenWaitOnVerdictPassed(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	bv := testutils.CreateSpecNode(t, "SPEC/root/v")
	bv.SetType("verdict")
	bv.SetOutput("code-from-spec/root/v/verdict.md")
	bv.Write()

	verdictContent := "verdict content\n"
	if err := os.MkdirAll("code-from-spec/root/v", 0755); err != nil {
		t.Fatalf("mkdir code-from-spec/root/v: %v", err)
	}
	if err := os.WriteFile("code-from-spec/root/v/verdict.md", []byte(verdictContent), 0644); err != nil {
		t.Fatalf("write verdict.md: %v", err)
	}

	chainHash := computeChainHash(t, "SPEC/root/v")
	checksum := fileChecksum(verdictContent)

	writeManifestEntry(t, "VERDICT/root/v", "code-from-spec/root/v/verdict.md", checksum, chainHash, "pass")

	bb := testutils.CreateSpecNode(t, "SPEC/root/b")
	bb.SetType("verdict")
	bb.AddWaitOn("VERDICT/root/v")
	bb.Write()

	report := mcpvalidatespecs.MCPValidateSpecs()

	for _, s := range report.Staleness {
		if s.Node == "SPEC/root/b" {
			if s.Blocked {
				t.Errorf("expected SPEC/root/b to not be blocked, got Blocked=true (BlockedBy=%q)", s.BlockedBy)
			}
			return
		}
	}
}

func TestNotBlockedWhenWaitOnVerdictAccepted(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	bv := testutils.CreateSpecNode(t, "SPEC/root/v")
	bv.SetType("verdict")
	bv.SetOutput("code-from-spec/root/v/verdict.md")
	bv.Write()

	verdictContent := "verdict content\n"
	if err := os.MkdirAll("code-from-spec/root/v", 0755); err != nil {
		t.Fatalf("mkdir code-from-spec/root/v: %v", err)
	}
	if err := os.WriteFile("code-from-spec/root/v/verdict.md", []byte(verdictContent), 0644); err != nil {
		t.Fatalf("write verdict.md: %v", err)
	}

	chainHash := computeChainHash(t, "SPEC/root/v")
	checksum := fileChecksum(verdictContent)

	writeManifestEntry(t, "VERDICT/root/v", "code-from-spec/root/v/verdict.md", checksum, chainHash, "accepted")

	bb := testutils.CreateSpecNode(t, "SPEC/root/b")
	bb.SetType("verdict")
	bb.AddWaitOn("VERDICT/root/v")
	bb.Write()

	report := mcpvalidatespecs.MCPValidateSpecs()

	for _, s := range report.Staleness {
		if s.Node == "SPEC/root/b" {
			if s.Blocked {
				t.Errorf("expected SPEC/root/b to not be blocked, got Blocked=true (BlockedBy=%q)", s.BlockedBy)
			}
			return
		}
	}
}

func TestBlockedByModifiedArtifactDependency(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	ba := testutils.CreateSpecNode(t, "SPEC/root/a")
	ba.SetType("artifact")
	ba.SetOutput("out/a.go")
	ba.Write()

	originalContent := "package a // original\n"
	if err := os.MkdirAll("out", 0755); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	if err := os.WriteFile("out/a.go", []byte(originalContent), 0644); err != nil {
		t.Fatalf("write out/a.go: %v", err)
	}

	chainHashA := computeChainHash(t, "SPEC/root/a")
	originalChecksum := fileChecksum(originalContent)

	writeManifestEntry(t, "ARTIFACT/root/a", "out/a.go", originalChecksum, chainHashA, "")

	if err := os.WriteFile("out/a.go", []byte("package a // modified\n"), 0644); err != nil {
		t.Fatalf("overwrite out/a.go: %v", err)
	}

	bb := testutils.CreateSpecNode(t, "SPEC/root/b")
	bb.SetType("artifact")
	bb.SetOutput("out/b.go")
	bb.AddImport("ARTIFACT/root/a")
	bb.Write()

	report := mcpvalidatespecs.MCPValidateSpecs()

	foundModified := false
	for _, s := range report.Staleness {
		if s.Node == "SPEC/root/a" && s.Status == "modified" {
			foundModified = true
			break
		}
	}
	if !foundModified {
		t.Errorf("expected staleness entry for SPEC/root/a with status 'modified', got %v", report.Staleness)
	}

	foundBlocked := false
	for _, s := range report.Staleness {
		if s.Node == "SPEC/root/b" {
			if !s.Blocked {
				t.Errorf("expected SPEC/root/b to be blocked, got Blocked=false")
			}
			if !strings.Contains(s.BlockedBy, "ARTIFACT/root/a") {
				t.Errorf("expected BlockedBy to contain ARTIFACT/root/a, got %q", s.BlockedBy)
			}
			foundBlocked = true
			break
		}
	}
	if !foundBlocked {
		t.Errorf("expected staleness entry for SPEC/root/b, got %v", report.Staleness)
	}
}

func TestTransitiveBlocking(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	ba := testutils.CreateSpecNode(t, "SPEC/root/a")
	ba.SetType("artifact")
	ba.SetOutput("out/a.go")
	ba.Write()

	bb := testutils.CreateSpecNode(t, "SPEC/root/b")
	bb.SetType("verdict")
	bb.AddWaitOn("ARTIFACT/root/a")
	bb.Write()

	bc := testutils.CreateSpecNode(t, "SPEC/root/c")
	bc.SetType("verdict")
	bc.AddWaitOn("VERDICT/root/b")
	bc.Write()

	report := mcpvalidatespecs.MCPValidateSpecs()

	foundB := false
	for _, s := range report.Staleness {
		if s.Node == "SPEC/root/b" {
			if !s.Blocked {
				t.Errorf("expected SPEC/root/b to be blocked, got Blocked=false")
			}
			foundB = true
			break
		}
	}
	if !foundB {
		t.Errorf("expected staleness entry for SPEC/root/b, got %v", report.Staleness)
	}

	foundC := false
	for _, s := range report.Staleness {
		if s.Node == "SPEC/root/c" {
			if !s.Blocked {
				t.Errorf("expected SPEC/root/c to be blocked, got Blocked=false")
			}
			if !strings.Contains(s.BlockedBy, "VERDICT/root/b") {
				t.Errorf("expected BlockedBy to contain VERDICT/root/b, got %q", s.BlockedBy)
			}
			foundC = true
			break
		}
	}
	if !foundC {
		t.Errorf("expected staleness entry for SPEC/root/c, got %v", report.Staleness)
	}
}

func TestBlockedEntryRetainsUnderlyingStatus(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	ba := testutils.CreateSpecNode(t, "SPEC/root/a")
	ba.SetType("artifact")
	ba.SetOutput("out/a.go")
	ba.Write()

	fileContent := "package a\n"
	if err := os.MkdirAll("out", 0755); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	if err := os.WriteFile("out/a.go", []byte(fileContent), 0644); err != nil {
		t.Fatalf("write out/a.go: %v", err)
	}

	checksum := fileChecksum(fileContent)
	staleHash := "AAAAAAAAAAAAAAAAAAAAAAAAAAA"

	writeManifestEntry(t, "ARTIFACT/root/a", "out/a.go", checksum, staleHash, "")

	bb := testutils.CreateSpecNode(t, "SPEC/root/b")
	bb.SetType("verdict")
	bb.AddWaitOn("ARTIFACT/root/a")
	bb.Write()

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, s := range report.Staleness {
		if s.Node == "SPEC/root/b" {
			if s.Status != "missing" {
				t.Errorf("expected SPEC/root/b status 'missing', got %q", s.Status)
			}
			if !s.Blocked {
				t.Errorf("expected SPEC/root/b to be blocked, got Blocked=false")
			}
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected staleness entry for SPEC/root/b, got %v", report.Staleness)
	}
}

func TestFormatErrorInvalidImports(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	b := testutils.CreateSpecNode(t, "SPEC/root/a")
	b.SetType("artifact")
	b.AddImport("SPEC/root/missing")
	b.Write()

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, fe := range report.FormatErrors {
		if fe.Node == "SPEC/root/a" && fe.Rule == "import_targets" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected format error for SPEC/root/a with rule 'import_targets', got %v", report.FormatErrors)
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
	bb.SetType("artifact")
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
	ba.SetType("artifact")
	ba.AddImport("SPEC/root/b")
	ba.Write()

	bb := testutils.CreateSpecNode(t, "SPEC/root/b")
	bb.SetType("artifact")
	bb.AddImport("SPEC/root/a")
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
	ba.SetType("artifact")
	ba.AddImport("SPEC/root/missing")
	ba.Write()

	bb := testutils.CreateSpecNode(t, "SPEC/root/b")
	bb.SetType("artifact")
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

func TestNodeWithNoTypeNotInStaleness(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	b := testutils.CreateSpecNode(t, "SPEC/root/a")
	b.Write()

	report := mcpvalidatespecs.MCPValidateSpecs()

	for _, s := range report.Staleness {
		if s.Node == "SPEC/root/a" {
			t.Errorf("expected no staleness entry for SPEC/root/a (no type), got %v", s)
		}
	}
}

func TestVerdictNodeStalenessDetectedWithResult(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	bv := testutils.CreateSpecNode(t, "SPEC/root/v")
	bv.SetType("verdict")
	bv.SetOutput("code-from-spec/root/v/verdict.md")
	bv.Write()

	verdictContent := "verdict content\n"
	if err := os.MkdirAll("code-from-spec/root/v", 0755); err != nil {
		t.Fatalf("mkdir code-from-spec/root/v: %v", err)
	}
	if err := os.WriteFile("code-from-spec/root/v/verdict.md", []byte(verdictContent), 0644); err != nil {
		t.Fatalf("write verdict.md: %v", err)
	}

	checksum := fileChecksum(verdictContent)
	staleHash := "AAAAAAAAAAAAAAAAAAAAAAAAAAA"

	writeManifestEntry(t, "VERDICT/root/v", "code-from-spec/root/v/verdict.md", checksum, staleHash, "pass")

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, s := range report.Staleness {
		if s.Node == "SPEC/root/v" && s.Status == "stale" && s.Result == "pass" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected staleness entry for SPEC/root/v with status 'stale' and result 'pass', got %v", report.Staleness)
	}
}

func TestVerdictNodeUpToDate(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	bv := testutils.CreateSpecNode(t, "SPEC/root/v")
	bv.SetType("verdict")
	bv.SetOutput("code-from-spec/root/v/verdict.md")
	bv.Write()

	verdictContent := "verdict content\n"
	if err := os.MkdirAll("code-from-spec/root/v", 0755); err != nil {
		t.Fatalf("mkdir code-from-spec/root/v: %v", err)
	}
	if err := os.WriteFile("code-from-spec/root/v/verdict.md", []byte(verdictContent), 0644); err != nil {
		t.Fatalf("write verdict.md: %v", err)
	}

	chainHash := computeChainHash(t, "SPEC/root/v")
	checksum := fileChecksum(verdictContent)

	writeManifestEntry(t, "VERDICT/root/v", "code-from-spec/root/v/verdict.md", checksum, chainHash, "pass")

	report := mcpvalidatespecs.MCPValidateSpecs()

	for _, s := range report.Staleness {
		if s.Node == "SPEC/root/v" {
			t.Errorf("expected no staleness entry for SPEC/root/v (up to date), got %v", s)
		}
	}
}

func TestOrphanVerdictEntry(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/root")
	b.Write()

	writeManifestEntry(t, "VERDICT/root/deleted", "code-from-spec/root/deleted/verdict.md", "AAAAAAAAAAAAAAAAAAAAAAAAAAA", "AAAAAAAAAAAAAAAAAAAAAAAAAAA", "")

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, s := range report.Staleness {
		if s.Status == "orphan" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected staleness entry with status 'orphan' for VERDICT entry, got %v", report.Staleness)
	}
}

func TestNoManifestFileAllArtifactsMissing(t *testing.T) {
	testutils.Chdir(t)

	createRootNode(t)

	b := testutils.CreateSpecNode(t, "SPEC/root/a")
	b.SetType("artifact")
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
