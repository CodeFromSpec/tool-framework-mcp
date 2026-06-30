package chainhash_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"
)

func TestHashIsDeterministic(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/root/a")
	b.SetPublic("## Interface\nsome content")
	b.Write()

	chain := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
	}

	hash1, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	hash2, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if hash1 != hash2 {
		t.Errorf("expected deterministic hash, got %q and %q", hash1, hash2)
	}
}

func TestHashIs27Characters(t *testing.T) {
	testutils.Chdir(t)

	b := testutils.CreateSpecNode(t, "SPEC/root/a")
	b.SetPublic("## Interface\nsome content")
	b.Write()

	chain := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27 characters, got %d: %q", len(hash), hash)
	}
}

func TestHashChangesWhenAncestorContentChanges(t *testing.T) {
	testutils.Chdir(t)

	rootB := testutils.CreateSpecNode(t, "SPEC/root")
	rootB.SetPublic("## Context\ninitial content")
	rootB.Write()

	aB := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB.SetPublic("## Interface\nsome interface")
	aB.Write()

	chain := chainresolver.Chain{
		Ancestors: []parsing.CfsReference{
			{
				NodeType:    parsing.CfsNodeTypeSpec,
				LogicalName: "SPEC/root",
				Path:        "code-from-spec/root/_node.md",
			},
		},
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	rootB2 := testutils.CreateSpecNode(t, "SPEC/root")
	rootB2.SetPublic("## Context\nmodified content")
	rootB2.Write()

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change after ancestor content change")
	}
}

func TestHashChangesWhenDependencyContentChanges(t *testing.T) {
	testutils.Chdir(t)

	testutils.CreateSpecNode(t, "SPEC/root").Write()

	bB := testutils.CreateSpecNode(t, "SPEC/root/b")
	bB.SetPublic("## Interface\ninitial content")
	bB.Write()

	aB := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB.SetPublic("## Interface\nsome interface")
	aB.Write()

	chain := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
		Dependencies: []parsing.CfsReference{
			{
				NodeType:    parsing.CfsNodeTypeSpec,
				LogicalName: "SPEC/root/b",
				Path:        "code-from-spec/root/b/_node.md",
			},
		},
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	bB2 := testutils.CreateSpecNode(t, "SPEC/root/b")
	bB2.SetPublic("## Interface\nmodified content")
	bB2.Write()

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change after dependency content change")
	}
}

func TestHashChangesWhenTargetPublicChanges(t *testing.T) {
	testutils.Chdir(t)

	aB := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB.SetPublic("## Interface\ninitial content")
	aB.Write()

	chain := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	aB2 := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB2.SetPublic("## Interface\nmodified content")
	aB2.Write()

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change after target public content change")
	}
}

func TestHashChangesWhenTargetAgentChanges(t *testing.T) {
	testutils.Chdir(t)

	aB := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB.SetPublic("## Interface\nsome interface")
	aB.SetAgent("initial agent content")
	aB.Write()

	chain := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	aB2 := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB2.SetPublic("## Interface\nsome interface")
	aB2.SetAgent("modified agent content")
	aB2.Write()

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change after target agent change")
	}
}

func TestAncestorWithPublicSubsectionsContributesHash(t *testing.T) {
	testutils.Chdir(t)

	rootB := testutils.CreateSpecNode(t, "SPEC/root")
	rootB.SetPublic("## Context\nsome context")
	rootB.Write()

	aB := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB.SetPublic("## Interface\nsome interface")
	aB.Write()

	chain := chainresolver.Chain{
		Ancestors: []parsing.CfsReference{
			{
				NodeType:    parsing.CfsNodeTypeSpec,
				LogicalName: "SPEC/root",
				Path:        "code-from-spec/root/_node.md",
			},
		},
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27 characters, got %d", len(hash))
	}
}

func TestAncestorWithoutPublicSectionSkipped(t *testing.T) {
	testutils.Chdir(t)

	aB := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB.SetPublic("## Interface\nsome interface")
	aB.Write()

	rootB := testutils.CreateSpecNode(t, "SPEC/root")
	rootB.SetPublic("## Context\nsome context")
	rootB.Write()

	chain := chainresolver.Chain{
		Ancestors: []parsing.CfsReference{
			{
				NodeType:    parsing.CfsNodeTypeSpec,
				LogicalName: "SPEC/root",
				Path:        "code-from-spec/root/_node.md",
			},
		},
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
	}

	hashWithPublic, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	testutils.CreateSpecNode(t, "SPEC/root").Write()

	hashWithoutPublic, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if hashWithPublic == hashWithoutPublic {
		t.Error("expected hash to differ when ancestor has no public section")
	}
}

func TestMultipleAncestorsOrderMatters(t *testing.T) {
	testutils.Chdir(t)

	rootB := testutils.CreateSpecNode(t, "SPEC/root")
	rootB.SetPublic("## Context\nroot context")
	rootB.Write()

	aB := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB.SetPublic("## Context\na context")
	aB.Write()

	bB := testutils.CreateSpecNode(t, "SPEC/root/a/b")
	bB.SetPublic("## Interface\nsome interface")
	bB.Write()

	chainA := chainresolver.Chain{
		Ancestors: []parsing.CfsReference{
			{
				NodeType:    parsing.CfsNodeTypeSpec,
				LogicalName: "SPEC/root",
				Path:        "code-from-spec/root/_node.md",
			},
			{
				NodeType:    parsing.CfsNodeTypeSpec,
				LogicalName: "SPEC/root/a",
				Path:        "code-from-spec/root/a/_node.md",
			},
		},
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a/b",
			Path:        "code-from-spec/root/a/b/_node.md",
		},
	}

	chainB := chainresolver.Chain{
		Ancestors: []parsing.CfsReference{
			{
				NodeType:    parsing.CfsNodeTypeSpec,
				LogicalName: "SPEC/root/a",
				Path:        "code-from-spec/root/a/_node.md",
			},
			{
				NodeType:    parsing.CfsNodeTypeSpec,
				LogicalName: "SPEC/root",
				Path:        "code-from-spec/root/_node.md",
			},
		},
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a/b",
			Path:        "code-from-spec/root/a/b/_node.md",
		},
	}

	hashA, err := chainhash.ChainHashCompute(chainA)
	if err != nil {
		t.Fatalf("chain A failed: %v", err)
	}
	hashB, err := chainhash.ChainHashCompute(chainB)
	if err != nil {
		t.Fatalf("chain B failed: %v", err)
	}

	if hashA == hashB {
		t.Error("expected hash to differ when ancestor order changes")
	}
}

func TestSpecDependencyWithoutQualifierHashesPublicSubsections(t *testing.T) {
	testutils.Chdir(t)

	bB := testutils.CreateSpecNode(t, "SPEC/root/b")
	bB.SetPublic("## Interface\ninitial content")
	bB.Write()

	aB := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB.SetPublic("## Interface\nsome interface")
	aB.Write()

	chain := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
		Dependencies: []parsing.CfsReference{
			{
				NodeType:    parsing.CfsNodeTypeSpec,
				LogicalName: "SPEC/root/b",
				Path:        "code-from-spec/root/b/_node.md",
			},
		},
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	bB2 := testutils.CreateSpecNode(t, "SPEC/root/b")
	bB2.SetPublic("## Interface\nmodified content")
	bB2.Write()

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change after dependency content change")
	}
}

func TestSpecDependencyWithQualifierHashesSubsection(t *testing.T) {
	testutils.Chdir(t)

	bB := testutils.CreateSpecNode(t, "SPEC/root/b")
	bB.SetPublic("## Interface\ninitial content")
	bB.Write()

	aB := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB.SetPublic("## Interface\nsome interface")
	aB.Write()

	chain := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
		Dependencies: []parsing.CfsReference{
			{
				NodeType:    parsing.CfsNodeTypeSpec,
				LogicalName: "SPEC/root/b",
				Path:        "code-from-spec/root/b/_node.md",
				Qualifier:   testutils.Ptr("interface"),
			},
		},
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	bB2 := testutils.CreateSpecNode(t, "SPEC/root/b")
	bB2.SetPublic("## Interface\nmodified content")
	bB2.Write()

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change after qualified dependency content change")
	}
}

func TestQualifierCaseNormalization(t *testing.T) {
	testutils.Chdir(t)

	bB := testutils.CreateSpecNode(t, "SPEC/root/b")
	bB.SetPublic("## Interface\nsome content")
	bB.Write()

	aB := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB.SetPublic("## Interface\nsome interface")
	aB.Write()

	chain := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
		Dependencies: []parsing.CfsReference{
			{
				NodeType:    parsing.CfsNodeTypeSpec,
				LogicalName: "SPEC/root/b",
				Path:        "code-from-spec/root/b/_node.md",
				Qualifier:   testutils.Ptr("INTERFACE"),
			},
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestArtifactDependencyHashesFullFileContent(t *testing.T) {
	testutils.Chdir(t)

	if err := os.MkdirAll("internal/artifact", 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile("internal/artifact/out.go", []byte("initial content"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	aB := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB.SetPublic("## Interface\nsome interface")
	aB.Write()

	chain := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
		Dependencies: []parsing.CfsReference{
			{
				NodeType:    parsing.CfsNodeTypeArtifact,
				LogicalName: "ARTIFACT/artifact/out",
				Path:        "internal/artifact/out.go",
			},
		},
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	if err := os.WriteFile("internal/artifact/out.go", []byte("modified content"), 0644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change after artifact file modification")
	}
}

func TestExternalDependencyHashesAllContent(t *testing.T) {
	testutils.Chdir(t)

	if err := os.WriteFile("external.txt", []byte("initial content"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	aB := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB.SetPublic("## Interface\nsome interface")
	aB.Write()

	chain := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
		Dependencies: []parsing.CfsReference{
			{
				NodeType:    parsing.CfsNodeTypeExternal,
				LogicalName: "EXTERNAL/external.txt",
				Path:        "external.txt",
			},
		},
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	if err := os.WriteFile("external.txt", []byte("modified content"), 0644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change after external file modification")
	}
}

func TestLeadingBlankLinesRemovedFromSubsection(t *testing.T) {
	testutils.Chdir(t)

	aB := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB.SetPublic("## Interface\n\n\nsome content")
	aB.Write()

	chainA := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
	}

	hashA, err := chainhash.ChainHashCompute(chainA)
	if err != nil {
		t.Fatalf("chain A failed: %v", err)
	}

	bB := testutils.CreateSpecNode(t, "SPEC/root/b")
	bB.SetPublic("## Interface\nsome content")
	bB.Write()

	chainB := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/b",
			Path:        "code-from-spec/root/b/_node.md",
		},
	}

	hashB, err := chainhash.ChainHashCompute(chainB)
	if err != nil {
		t.Fatalf("chain B failed: %v", err)
	}

	if hashA != hashB {
		t.Errorf("expected equal hashes when leading blanks differ, got %q and %q", hashA, hashB)
	}
}

func TestTrailingBlankLinesRemovedFromSubsection(t *testing.T) {
	testutils.Chdir(t)

	aB := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB.SetPublic("## Interface\nsome content\n\n")
	aB.Write()

	chainA := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
	}

	hashA, err := chainhash.ChainHashCompute(chainA)
	if err != nil {
		t.Fatalf("chain A failed: %v", err)
	}

	bB := testutils.CreateSpecNode(t, "SPEC/root/b")
	bB.SetPublic("## Interface\nsome content")
	bB.Write()

	chainB := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/b",
			Path:        "code-from-spec/root/b/_node.md",
		},
	}

	hashB, err := chainhash.ChainHashCompute(chainB)
	if err != nil {
		t.Fatalf("chain B failed: %v", err)
	}

	if hashA != hashB {
		t.Errorf("expected equal hashes when trailing blanks differ, got %q and %q", hashA, hashB)
	}
}

func TestInteriorBlankLinesPreserved(t *testing.T) {
	testutils.Chdir(t)

	aB := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB.SetPublic("## Interface\nfirst line\n\nsecond line")
	aB.Write()

	chainA := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
	}

	hashA, err := chainhash.ChainHashCompute(chainA)
	if err != nil {
		t.Fatalf("chain A failed: %v", err)
	}

	bB := testutils.CreateSpecNode(t, "SPEC/root/b")
	bB.SetPublic("## Interface\nfirst line\nsecond line")
	bB.Write()

	chainB := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/b",
			Path:        "code-from-spec/root/b/_node.md",
		},
	}

	hashB, err := chainhash.ChainHashCompute(chainB)
	if err != nil {
		t.Fatalf("chain B failed: %v", err)
	}

	if hashA == hashB {
		t.Error("expected different hashes when interior blank lines differ")
	}
}

func TestTargetPublicAndAgentBothContribute(t *testing.T) {
	testutils.Chdir(t)

	aB := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB.SetPublic("## Interface\nsome interface")
	aB.SetAgent("some agent content")
	aB.Write()

	chain := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	aB2 := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB2.SetPublic("## Interface\nsome interface")
	aB2.Write()

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change when agent section removed")
	}
}

func TestTargetWithoutAgentIsSkipped(t *testing.T) {
	testutils.Chdir(t)

	aB := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB.SetPublic("## Interface\nsome interface")
	aB.Write()

	chain := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27 characters, got %d", len(hash))
	}
}

func TestInputHashesFullFileContent(t *testing.T) {
	testutils.Chdir(t)

	if err := os.MkdirAll("internal/artifact", 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile("internal/artifact/input.go", []byte("initial input"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	aB := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB.SetPublic("## Interface\nsome interface")
	aB.Write()

	inputRef := parsing.CfsReference{
		NodeType:    parsing.CfsNodeTypeArtifact,
		LogicalName: "ARTIFACT/artifact/input",
		Path:        "internal/artifact/input.go",
	}

	chain := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
		Input: &inputRef,
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	if err := os.WriteFile("internal/artifact/input.go", []byte("modified input"), 0644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change after input file modification")
	}
}

func TestSpecInputHashesPublicSubsections(t *testing.T) {
	testutils.Chdir(t)

	bB := testutils.CreateSpecNode(t, "SPEC/root/b")
	bB.SetPublic("## Interface\ninitial content")
	bB.Write()

	aB := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB.SetPublic("## Interface\nsome interface")
	aB.Write()

	inputRef := parsing.CfsReference{
		NodeType:    parsing.CfsNodeTypeSpec,
		LogicalName: "SPEC/root/b",
		Path:        "code-from-spec/root/b/_node.md",
	}

	chain := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
		Input: &inputRef,
	}

	hashBefore, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	bB2 := testutils.CreateSpecNode(t, "SPEC/root/b")
	bB2.SetPublic("## Interface\nmodified content")
	bB2.Write()

	hashAfter, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if hashBefore == hashAfter {
		t.Error("expected hash to change after spec input content change")
	}
}

func TestNoInputSkipped(t *testing.T) {
	testutils.Chdir(t)

	aB := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB.SetPublic("## Interface\nsome interface")
	aB.Write()

	chain := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) != 27 {
		t.Errorf("expected 27 characters, got %d", len(hash))
	}
}

func TestQualifierReferencesNonExistentSubsection(t *testing.T) {
	testutils.Chdir(t)

	bB := testutils.CreateSpecNode(t, "SPEC/root/b")
	bB.SetPublic("## Context\nonly context, no interface")
	bB.Write()

	aB := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB.SetPublic("## Interface\nsome interface")
	aB.Write()

	chain := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
		Dependencies: []parsing.CfsReference{
			{
				NodeType:    parsing.CfsNodeTypeSpec,
				LogicalName: "SPEC/root/b",
				Path:        "code-from-spec/root/b/_node.md",
				Qualifier:   testutils.Ptr("interface"),
			},
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		t.Errorf("expected no error when qualifier does not match any subsection, got: %v", err)
	}
}

func TestUnreadableSpecNodeFile(t *testing.T) {
	testutils.Chdir(t)

	chain := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/nonexistent",
			Path:        "code-from-spec/root/nonexistent/_node.md",
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for unreadable spec node file")
	}
	if !errors.Is(err, chainhash.ErrParseFailure) {
		t.Errorf("expected ErrParseFailure, got: %v", err)
	}
}

func TestUnreadableArtifactFile(t *testing.T) {
	testutils.Chdir(t)

	aB := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB.SetPublic("## Interface\nsome interface")
	aB.Write()

	chain := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
		Dependencies: []parsing.CfsReference{
			{
				NodeType:    parsing.CfsNodeTypeArtifact,
				LogicalName: "ARTIFACT/nonexistent/file",
				Path:        "nonexistent/file.go",
			},
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for unreadable artifact file")
	}
	if !errors.Is(err, oslayer.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestUnreadableExternalFile(t *testing.T) {
	testutils.Chdir(t)

	aB := testutils.CreateSpecNode(t, "SPEC/root/a")
	aB.SetPublic("## Interface\nsome interface")
	aB.Write()

	chain := chainresolver.Chain{
		Target: parsing.CfsReference{
			NodeType:    parsing.CfsNodeTypeSpec,
			LogicalName: "SPEC/root/a",
			Path:        "code-from-spec/root/a/_node.md",
		},
		Dependencies: []parsing.CfsReference{
			{
				NodeType:    parsing.CfsNodeTypeExternal,
				LogicalName: "EXTERNAL/nonexistent/file.txt",
				Path:        "nonexistent/file.txt",
			},
		},
	}

	_, err := chainhash.ChainHashCompute(chain)
	if err == nil {
		t.Fatal("expected error for unreadable external file")
	}
	if !errors.Is(err, oslayer.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}
