// code-from-spec: ROOT/golang/tests/utils/logical_names@RNP2gU7p5EOVBGI6pMPGV8ovsWc

package logicalnames_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// --- LogicalNameToPath ---

func TestLogicalNameToPath_TC01_RootAlone(t *testing.T) {
	result, err := logicalnames.LogicalNameToPath("ROOT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "code-from-spec/_node.md"
	if result.Value != want {
		t.Errorf("got %q, want %q", result.Value, want)
	}
}

func TestLogicalNameToPath_TC02_RootWithPath(t *testing.T) {
	result, err := logicalnames.LogicalNameToPath("ROOT/payments/processor")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "code-from-spec/payments/processor/_node.md"
	if result.Value != want {
		t.Errorf("got %q, want %q", result.Value, want)
	}
}

func TestLogicalNameToPath_TC03_StripsQualifierBeforeResolving(t *testing.T) {
	result, err := logicalnames.LogicalNameToPath("ROOT/x/y(interface)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "code-from-spec/x/y/_node.md"
	if result.Value != want {
		t.Errorf("got %q, want %q", result.Value, want)
	}
}

func TestLogicalNameToPath_TC04_RejectsArtifactReference(t *testing.T) {
	_, err := logicalnames.LogicalNameToPath("ARTIFACT/x(y)")
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("expected ErrUnsupportedReference, got %v", err)
	}
}

func TestLogicalNameToPath_TC05_RejectsUnrecognizedPrefix(t *testing.T) {
	_, err := logicalnames.LogicalNameToPath("UNKNOWN/something")
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("expected ErrUnsupportedReference, got %v", err)
	}
}

func TestLogicalNameToPath_TC06_RejectsEmptyString(t *testing.T) {
	_, err := logicalnames.LogicalNameToPath("")
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("expected ErrUnsupportedReference, got %v", err)
	}
}

// --- LogicalNameFromPath ---

func TestLogicalNameFromPath_TC07_RootNode(t *testing.T) {
	path := &pathutils.PathCfs{Value: "code-from-spec/_node.md"}
	result, err := logicalnames.LogicalNameFromPath(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "ROOT"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestLogicalNameFromPath_TC08_NestedNode(t *testing.T) {
	path := &pathutils.PathCfs{Value: "code-from-spec/x/y/_node.md"}
	result, err := logicalnames.LogicalNameFromPath(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "ROOT/x/y"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestLogicalNameFromPath_TC09_RejectsNonNodePath(t *testing.T) {
	path := &pathutils.PathCfs{Value: "internal/config/config.go"}
	_, err := logicalnames.LogicalNameFromPath(path)
	if !errors.Is(err, logicalnames.ErrInvalidPath) {
		t.Errorf("expected ErrInvalidPath, got %v", err)
	}
}

func TestLogicalNameFromPath_TC10_RejectsPathWithoutNodeMd(t *testing.T) {
	path := &pathutils.PathCfs{Value: "code-from-spec/x/y/output.md"}
	_, err := logicalnames.LogicalNameFromPath(path)
	if !errors.Is(err, logicalnames.ErrInvalidPath) {
		t.Errorf("expected ErrInvalidPath, got %v", err)
	}
}

// --- LogicalNameGetParent ---

func TestLogicalNameGetParent_TC11_RootXParentIsRoot(t *testing.T) {
	result, err := logicalnames.LogicalNameGetParent("ROOT/domain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "ROOT"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestLogicalNameGetParent_TC12_RootXYParentIsRootX(t *testing.T) {
	result, err := logicalnames.LogicalNameGetParent("ROOT/domain/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "ROOT/domain"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestLogicalNameGetParent_TC13_StripsQualifierBeforeComputingParent(t *testing.T) {
	result, err := logicalnames.LogicalNameGetParent("ROOT/domain/config(interface)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "ROOT/domain"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestLogicalNameGetParent_TC14_RootHasNoParent(t *testing.T) {
	_, err := logicalnames.LogicalNameGetParent("ROOT")
	if !errors.Is(err, logicalnames.ErrNoParent) {
		t.Errorf("expected ErrNoParent, got %v", err)
	}
}

func TestLogicalNameGetParent_TC15_RejectsArtifactReference(t *testing.T) {
	_, err := logicalnames.LogicalNameGetParent("ARTIFACT/x(y)")
	if !errors.Is(err, logicalnames.ErrNotARootReference) {
		t.Errorf("expected ErrNotARootReference, got %v", err)
	}
}

// --- LogicalNameGetQualifier ---

func TestLogicalNameGetQualifier_TC16_ExtractsQualifierFromRootReference(t *testing.T) {
	qualifier, ok := logicalnames.LogicalNameGetQualifier("ROOT/x/y(interface)")
	if !ok {
		t.Fatal("expected ok=true, got false")
	}
	want := "interface"
	if qualifier != want {
		t.Errorf("got %q, want %q", qualifier, want)
	}
}

func TestLogicalNameGetQualifier_TC17_ExtractsQualifierFromArtifactReference(t *testing.T) {
	qualifier, ok := logicalnames.LogicalNameGetQualifier("ARTIFACT/x/y(id)")
	if !ok {
		t.Fatal("expected ok=true, got false")
	}
	want := "id"
	if qualifier != want {
		t.Errorf("got %q, want %q", qualifier, want)
	}
}

func TestLogicalNameGetQualifier_TC18_ReturnsAbsentWhenNoQualifier(t *testing.T) {
	qualifier, ok := logicalnames.LogicalNameGetQualifier("ROOT/x/y")
	if ok {
		t.Errorf("expected ok=false, got true with qualifier %q", qualifier)
	}
	if qualifier != "" {
		t.Errorf("expected empty qualifier, got %q", qualifier)
	}
}

func TestLogicalNameGetQualifier_TC19_ReturnsAbsentForRootAlone(t *testing.T) {
	qualifier, ok := logicalnames.LogicalNameGetQualifier("ROOT")
	if ok {
		t.Errorf("expected ok=false, got true with qualifier %q", qualifier)
	}
	if qualifier != "" {
		t.Errorf("expected empty qualifier, got %q", qualifier)
	}
}

// --- LogicalNameStripQualifier ---

func TestLogicalNameStripQualifier_TC20_StripsQualifierFromRootReference(t *testing.T) {
	result := logicalnames.LogicalNameStripQualifier("ROOT/x/y(interface)")
	want := "ROOT/x/y"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestLogicalNameStripQualifier_TC21_StripsQualifierFromArtifactReference(t *testing.T) {
	result := logicalnames.LogicalNameStripQualifier("ARTIFACT/x/y(id)")
	want := "ARTIFACT/x/y"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestLogicalNameStripQualifier_TC22_NoQualifierReturnsUnchanged(t *testing.T) {
	result := logicalnames.LogicalNameStripQualifier("ROOT/x/y")
	want := "ROOT/x/y"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestLogicalNameStripQualifier_TC23_RootAloneReturnsUnchanged(t *testing.T) {
	result := logicalnames.LogicalNameStripQualifier("ROOT")
	want := "ROOT"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestLogicalNameStripQualifier_TC24_EmptyStringReturnsUnchanged(t *testing.T) {
	result := logicalnames.LogicalNameStripQualifier("")
	want := ""
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

// --- LogicalNameHasParent ---

func TestLogicalNameHasParent_TC25_RootAlone(t *testing.T) {
	result := logicalnames.LogicalNameHasParent("ROOT")
	if result {
		t.Error("expected false for ROOT, got true")
	}
}

func TestLogicalNameHasParent_TC26_RootWithPath(t *testing.T) {
	result := logicalnames.LogicalNameHasParent("ROOT/domain/config")
	if !result {
		t.Error("expected true for ROOT/domain/config, got false")
	}
}

func TestLogicalNameHasParent_TC27_RootWithQualifier(t *testing.T) {
	result := logicalnames.LogicalNameHasParent("ROOT/domain/config(interface)")
	if !result {
		t.Error("expected true for ROOT/domain/config(interface), got false")
	}
}

func TestLogicalNameHasParent_TC28_ArtifactReference(t *testing.T) {
	result := logicalnames.LogicalNameHasParent("ARTIFACT/x(y)")
	if result {
		t.Error("expected false for ARTIFACT/x(y), got true")
	}
}

func TestLogicalNameHasParent_TC29_EmptyString(t *testing.T) {
	result := logicalnames.LogicalNameHasParent("")
	if result {
		t.Error("expected false for empty string, got true")
	}
}

// --- LogicalNameHasQualifier ---

func TestLogicalNameHasQualifier_TC30_WithoutQualifier(t *testing.T) {
	result := logicalnames.LogicalNameHasQualifier("ROOT/x")
	if result {
		t.Error("expected false for ROOT/x, got true")
	}
}

func TestLogicalNameHasQualifier_TC31_WithQualifier(t *testing.T) {
	result := logicalnames.LogicalNameHasQualifier("ROOT/x(y)")
	if !result {
		t.Error("expected true for ROOT/x(y), got false")
	}
}

func TestLogicalNameHasQualifier_TC32_ArtifactWithQualifier(t *testing.T) {
	result := logicalnames.LogicalNameHasQualifier("ARTIFACT/x(y)")
	if !result {
		t.Error("expected true for ARTIFACT/x(y), got false")
	}
}

func TestLogicalNameHasQualifier_TC33_RootAlone(t *testing.T) {
	result := logicalnames.LogicalNameHasQualifier("ROOT")
	if result {
		t.Error("expected false for ROOT, got true")
	}
}

func TestLogicalNameHasQualifier_TC34_EmptyString(t *testing.T) {
	result := logicalnames.LogicalNameHasQualifier("")
	if result {
		t.Error("expected false for empty string, got true")
	}
}

// --- LogicalNameIsArtifact ---

func TestLogicalNameIsArtifact_TC35_ArtifactReference(t *testing.T) {
	result := logicalnames.LogicalNameIsArtifact("ARTIFACT/x(y)")
	if !result {
		t.Error("expected true for ARTIFACT/x(y), got false")
	}
}

func TestLogicalNameIsArtifact_TC36_RootReference(t *testing.T) {
	result := logicalnames.LogicalNameIsArtifact("ROOT/x(y)")
	if result {
		t.Error("expected false for ROOT/x(y), got true")
	}
}

func TestLogicalNameIsArtifact_TC37_EmptyString(t *testing.T) {
	result := logicalnames.LogicalNameIsArtifact("")
	if result {
		t.Error("expected false for empty string, got true")
	}
}

// --- LogicalNameGetArtifactGenerator ---

func TestLogicalNameGetArtifactGenerator_TC38_SimpleArtifact(t *testing.T) {
	result, err := logicalnames.LogicalNameGetArtifactGenerator("ARTIFACT/x(y)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "ROOT/x"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestLogicalNameGetArtifactGenerator_TC39_NestedArtifact(t *testing.T) {
	result, err := logicalnames.LogicalNameGetArtifactGenerator("ARTIFACT/x/y/z(id)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "ROOT/x/y/z"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestLogicalNameGetArtifactGenerator_TC40_RejectsRootReference(t *testing.T) {
	_, err := logicalnames.LogicalNameGetArtifactGenerator("ROOT/x(y)")
	if !errors.Is(err, logicalnames.ErrNotAnArtifactReference) {
		t.Errorf("expected ErrNotAnArtifactReference, got %v", err)
	}
}

func TestLogicalNameGetArtifactGenerator_TC41_ArtifactWithoutQualifier(t *testing.T) {
	result, err := logicalnames.LogicalNameGetArtifactGenerator("ARTIFACT/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "ROOT/x"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}
