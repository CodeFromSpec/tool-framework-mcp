// code-from-spec: ROOT/golang/tests/mcp_tools/validate_specs@UlyNwYylLthC2a-Y9PQjVWGC3Aw
package mcpvalidatespecs_test

import (
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpvalidatespecs"
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
	if err := os.MkdirAll(parentDir(path), 0755); err != nil {
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

func parentDir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return "."
}

func testRootNode() string {
	return `---
---
# ROOT

## Public

Root public section.
`
}

func testLeafNode(logicalName string, frontmatter string) string {
	return frontmatter + "\n# " + logicalName + "\n"
}

func testCreateRoot(t *testing.T) {
	t.Helper()
	testWriteFile(t, "code-from-spec/_node.md", testRootNode())
}

func testCreateLeaf(t *testing.T, logicalName string, frontmatter string) {
	t.Helper()
	dirPath := "code-from-spec"
	suffix := logicalName[len("ROOT"):]
	for _, ch := range suffix {
		_ = ch
	}
	dirPath = "code-from-spec" + logicalNameToDir(logicalName[len("ROOT"):])
	testWriteFile(t, dirPath+"/_node.md", testLeafNode(logicalName, frontmatter))
}

func logicalNameToDir(suffix string) string {
	result := ""
	for _, ch := range suffix {
		if ch == '/' {
			result += "/"
		} else {
			result += string(ch)
		}
	}
	return result
}

func TestMCPValidateSpecs_HP1_CleanTree(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testCreateRoot(t)
	testCreateLeaf(t, "ROOT/a", "---\noutput: out/a.go\n---")

	report1 := mcpvalidatespecs.MCPValidateSpecs()

	var chainHash string
	for _, s := range report1.Staleness {
		if s.Node == "ROOT/a" {
			chainHash = s.Detail
			_ = chainHash
		}
	}

	report2 := mcpvalidatespecs.MCPValidateSpecs()
	var currentHash string
	_ = currentHash

	for _, s := range report2.Staleness {
		if s.Node == "ROOT/a" && s.Status == "missing" {
			currentHash = ""
		}
	}

	report3 := mcpvalidatespecs.MCPValidateSpecs()
	for _, s := range report3.Staleness {
		if s.Node == "ROOT/a" {
			_ = s.Detail
		}
	}

	staleReport := mcpvalidatespecs.MCPValidateSpecs()
	var foundHash string
	for _, s := range staleReport.Staleness {
		if s.Node == "ROOT/a" {
			foundHash = s.Detail
		}
	}

	testWriteFile(t, "out/a.go", "// code-from-spec: ROOT/a@"+foundHash+"\npackage main\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) != 0 {
		t.Errorf("expected no format errors, got %d", len(report.FormatErrors))
	}
	if len(report.Cycles) != 0 {
		t.Errorf("expected no cycles, got %v", report.Cycles)
	}
	if len(report.Staleness) != 0 {
		t.Errorf("expected no staleness entries, got %d", len(report.Staleness))
	}
}

func TestMCPValidateSpecs_HP2_StaleArtifact(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testCreateRoot(t)
	testCreateLeaf(t, "ROOT/a", "---\noutput: out/a.go\n---")
	testWriteFile(t, "out/a.go", "// code-from-spec: ROOT/a@outdatedhashXXXXXXXXXXX\npackage main\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, s := range report.Staleness {
		if s.Node == "ROOT/a" && s.Status == "stale" {
			found = true
			_ = s.Rank
		}
	}
	if !found {
		t.Errorf("expected staleness entry for ROOT/a with status=stale, got %+v", report.Staleness)
	}
	if len(report.Staleness) != 1 {
		t.Errorf("expected exactly 1 staleness entry, got %d", len(report.Staleness))
	}
}

func TestMCPValidateSpecs_HP3_MissingArtifact(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testCreateRoot(t)
	testCreateLeaf(t, "ROOT/a", "---\noutput: out/a.go\n---")

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, s := range report.Staleness {
		if s.Node == "ROOT/a" && s.Status == "missing" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected staleness entry for ROOT/a with status=missing, got %+v", report.Staleness)
	}
	if len(report.Staleness) != 1 {
		t.Errorf("expected exactly 1 staleness entry, got %d", len(report.Staleness))
	}
}

func TestMCPValidateSpecs_HP4_MalformedTag(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testCreateRoot(t)
	testCreateLeaf(t, "ROOT/a", "---\noutput: out/a.go\n---")
	testWriteFile(t, "out/a.go", "package main\n// no artifact tag here\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, s := range report.Staleness {
		if s.Node == "ROOT/a" && s.Status == "malformed tag" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected staleness entry for ROOT/a with status=malformed tag, got %+v", report.Staleness)
	}
	if len(report.Staleness) != 1 {
		t.Errorf("expected exactly 1 staleness entry, got %d", len(report.Staleness))
	}
}

func TestMCPValidateSpecs_HP5_StalenessRank(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testCreateRoot(t)
	testCreateLeaf(t, "ROOT/a", "---\noutput: out/a.go\n---")
	testCreateLeaf(t, "ROOT/b", "---\noutput: out/b.go\ndepends_on:\n  - ROOT/a\n---")
	testWriteFile(t, "out/a.go", "// code-from-spec: ROOT/a@outdatedhashXXXXXXXXXXX\npackage main\n")
	testWriteFile(t, "out/b.go", "// code-from-spec: ROOT/b@outdatedhashXXXXXXXXXXX\npackage main\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	var rankA, rankB int
	foundA, foundB := false, false
	for _, s := range report.Staleness {
		if s.Node == "ROOT/a" {
			rankA = s.Rank
			foundA = true
		}
		if s.Node == "ROOT/b" {
			rankB = s.Rank
			foundB = true
		}
	}

	if !foundA {
		t.Errorf("expected staleness entry for ROOT/a")
	}
	if !foundB {
		t.Errorf("expected staleness entry for ROOT/b")
	}
	if foundA && foundB && rankA >= rankB {
		t.Errorf("expected ROOT/a rank (%d) < ROOT/b rank (%d)", rankA, rankB)
	}
}

func TestMCPValidateSpecs_HP6_StalenessOrderedByRankThenName(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testCreateRoot(t)
	testCreateLeaf(t, "ROOT/a", "---\noutput: out/a.go\n---")
	testCreateLeaf(t, "ROOT/z", "---\noutput: out/z.go\n---")
	testWriteFile(t, "out/a.go", "// code-from-spec: ROOT/a@outdatedhashXXXXXXXXXXX\npackage main\n")
	testWriteFile(t, "out/z.go", "// code-from-spec: ROOT/z@outdatedhashXXXXXXXXXXX\npackage main\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Staleness) != 2 {
		t.Fatalf("expected 2 staleness entries, got %d", len(report.Staleness))
	}

	if report.Staleness[0].Node != "ROOT/a" {
		t.Errorf("expected ROOT/a first, got %s", report.Staleness[0].Node)
	}
	if report.Staleness[1].Node != "ROOT/z" {
		t.Errorf("expected ROOT/z second, got %s", report.Staleness[1].Node)
	}
}

func TestMCPValidateSpecs_FE1_InvalidDependsOn(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testCreateRoot(t)
	testCreateLeaf(t, "ROOT/a", "---\ndepends_on:\n  - ROOT/missing\n---")

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, e := range report.FormatErrors {
		if e.Node == "ROOT/a" && e.Rule == "dependency_targets" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected format error for ROOT/a with rule=dependency_targets, got %+v", report.FormatErrors)
	}
}

func TestMCPValidateSpecs_FE2_ParseFailure(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testCreateRoot(t)

	if err := os.MkdirAll("code-from-spec/a", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/a/_node.md", []byte("invalid content before heading\n# ROOT/a\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, e := range report.FormatErrors {
		if e.Node == "ROOT/a" && e.Rule == "parse" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected format error for ROOT/a with rule=parse, got %+v", report.FormatErrors)
	}
}

func TestMCPValidateSpecs_FE3_ContinuesAfterParseFailure(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testCreateRoot(t)

	if err := os.MkdirAll("code-from-spec/a", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/a/_node.md", []byte("invalid content before heading\n# ROOT/a\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	testCreateLeaf(t, "ROOT/b", "---\noutput: out/b.go\n---")
	testWriteFile(t, "out/b.go", "// code-from-spec: ROOT/b@outdatedhashXXXXXXXXXXX\npackage main\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	foundFormatErr := false
	for _, e := range report.FormatErrors {
		if e.Node == "ROOT/a" && e.Rule == "parse" {
			foundFormatErr = true
		}
	}
	if !foundFormatErr {
		t.Errorf("expected format error for ROOT/a with rule=parse, got %+v", report.FormatErrors)
	}

	foundStaleness := false
	for _, s := range report.Staleness {
		if s.Node == "ROOT/b" {
			foundStaleness = true
		}
	}
	if !foundStaleness {
		t.Errorf("expected staleness entry for ROOT/b, got %+v", report.Staleness)
	}
}

func TestMCPValidateSpecs_CD1_SimpleCycle(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testCreateRoot(t)
	testCreateLeaf(t, "ROOT/a", "---\ndepends_on:\n  - ROOT/b\n---")
	testCreateLeaf(t, "ROOT/b", "---\ndepends_on:\n  - ROOT/a\n---")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.Cycles) == 0 {
		t.Errorf("expected cycles to be non-empty")
	}

	foundA := false
	foundB := false
	for _, c := range report.Cycles {
		if c == "ROOT/a" {
			foundA = true
		}
		if c == "ROOT/b" {
			foundB = true
		}
	}
	if !foundA && !foundB {
		t.Errorf("expected cycles to contain ROOT/a or ROOT/b, got %v", report.Cycles)
	}
}

func TestMCPValidateSpecs_CD2_RankingSkippedOnFormatErrors(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testCreateRoot(t)
	testCreateLeaf(t, "ROOT/a", "---\ndepends_on:\n  - ROOT/missing\n---")
	testCreateLeaf(t, "ROOT/b", "---\noutput: out/b.go\n---")
	testWriteFile(t, "out/b.go", "// code-from-spec: ROOT/b@outdatedhashXXXXXXXXXXX\npackage main\n")

	report := mcpvalidatespecs.MCPValidateSpecs()

	if len(report.FormatErrors) == 0 {
		t.Errorf("expected format errors to be non-empty")
	}

	for _, s := range report.Staleness {
		if s.Node == "ROOT/b" {
			if s.Rank != 0 {
				t.Errorf("expected ROOT/b rank=0 when ranking skipped, got %d", s.Rank)
			}
		}
	}
}

func TestMCPValidateSpecs_EC1_EmptySpecTree(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	report := mcpvalidatespecs.MCPValidateSpecs()

	found := false
	for _, e := range report.FormatErrors {
		if e.Rule == "scan" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected format error with rule=scan, got %+v", report.FormatErrors)
	}
	if len(report.Cycles) != 0 {
		t.Errorf("expected no cycles, got %v", report.Cycles)
	}
	if len(report.Staleness) != 0 {
		t.Errorf("expected no staleness entries, got %+v", report.Staleness)
	}
}

func TestMCPValidateSpecs_EC2_NodeWithNoOutput(t *testing.T) {
	tmp := t.TempDir()
	testChdir(t, tmp)

	testCreateRoot(t)
	testCreateLeaf(t, "ROOT/a", "---\n---")

	report := mcpvalidatespecs.MCPValidateSpecs()

	for _, s := range report.Staleness {
		if s.Node == "ROOT/a" {
			t.Errorf("expected no staleness entry for ROOT/a (no output declared), got %+v", s)
		}
	}
}
