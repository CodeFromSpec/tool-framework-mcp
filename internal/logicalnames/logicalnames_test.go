// code-from-spec: ROOT/golang/tests/utils/logical_names@x6P9abAWzs6y3LuOxjmDy-HNacM

package logicalnames_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

// ----------------------------------------------------------------------------
// LogicalNameToPath
// ----------------------------------------------------------------------------

func TestLogicalNameToPath_TC01_ROOTAlone(t *testing.T) {
	got, err := logicalnames.LogicalNameToPath("ROOT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "code-from-spec/_node.md"
	if got.Value != want {
		t.Errorf("got %q, want %q", got.Value, want)
	}
}

func TestLogicalNameToPath_TC02_ROOTWithPath(t *testing.T) {
	got, err := logicalnames.LogicalNameToPath("ROOT/payments/processor")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "code-from-spec/payments/processor/_node.md"
	if got.Value != want {
		t.Errorf("got %q, want %q", got.Value, want)
	}
}

func TestLogicalNameToPath_TC03_StripsQualifier(t *testing.T) {
	got, err := logicalnames.LogicalNameToPath("ROOT/x/y(interface)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "code-from-spec/x/y/_node.md"
	if got.Value != want {
		t.Errorf("got %q, want %q", got.Value, want)
	}
}

func TestLogicalNameToPath_TC04_RejectsArtifactReference(t *testing.T) {
	_, err := logicalnames.LogicalNameToPath("ARTIFACT/x(y)")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("got error %v, want ErrUnsupportedReference", err)
	}
}

func TestLogicalNameToPath_TC05_RejectsUnrecognizedPrefix(t *testing.T) {
	_, err := logicalnames.LogicalNameToPath("UNKNOWN/something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("got error %v, want ErrUnsupportedReference", err)
	}
}

func TestLogicalNameToPath_TC06_RejectsEmptyString(t *testing.T) {
	_, err := logicalnames.LogicalNameToPath("")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("got error %v, want ErrUnsupportedReference", err)
	}
}

// ----------------------------------------------------------------------------
// LogicalNameFromPath
// ----------------------------------------------------------------------------

func TestLogicalNameFromPath_TC07_RootNode(t *testing.T) {
	cfs := &pathutils.PathCfs{Value: "code-from-spec/_node.md"}
	got, err := logicalnames.LogicalNameFromPath(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "ROOT"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLogicalNameFromPath_TC08_NestedNode(t *testing.T) {
	cfs := &pathutils.PathCfs{Value: "code-from-spec/x/y/_node.md"}
	got, err := logicalnames.LogicalNameFromPath(cfs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "ROOT/x/y"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLogicalNameFromPath_TC09_RejectsNonNodePath(t *testing.T) {
	cfs := &pathutils.PathCfs{Value: "internal/config/config.go"}
	_, err := logicalnames.LogicalNameFromPath(cfs)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrInvalidPath) {
		t.Errorf("got error %v, want ErrInvalidPath", err)
	}
}

func TestLogicalNameFromPath_TC10_RejectsPathWithoutNodeMd(t *testing.T) {
	cfs := &pathutils.PathCfs{Value: "code-from-spec/x/y/output.md"}
	_, err := logicalnames.LogicalNameFromPath(cfs)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrInvalidPath) {
		t.Errorf("got error %v, want ErrInvalidPath", err)
	}
}

// ----------------------------------------------------------------------------
// LogicalNameGetParent
// ----------------------------------------------------------------------------

func TestLogicalNameGetParent_TC11_OneLevelDeep(t *testing.T) {
	got, err := logicalnames.LogicalNameGetParent("ROOT/domain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "ROOT"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLogicalNameGetParent_TC12_TwoLevelsDeep(t *testing.T) {
	got, err := logicalnames.LogicalNameGetParent("ROOT/domain/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "ROOT/domain"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLogicalNameGetParent_TC13_StripsQualifier(t *testing.T) {
	got, err := logicalnames.LogicalNameGetParent("ROOT/domain/config(interface)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "ROOT/domain"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLogicalNameGetParent_TC14_ROOTHasNoParent(t *testing.T) {
	_, err := logicalnames.LogicalNameGetParent("ROOT")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrNoParent) {
		t.Errorf("got error %v, want ErrNoParent", err)
	}
}

func TestLogicalNameGetParent_TC15_RejectsArtifactReference(t *testing.T) {
	_, err := logicalnames.LogicalNameGetParent("ARTIFACT/x(y)")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrNotARootReference) {
		t.Errorf("got error %v, want ErrNotARootReference", err)
	}
}

// ----------------------------------------------------------------------------
// LogicalNameGetQualifier
// ----------------------------------------------------------------------------

func TestLogicalNameGetQualifier_TC16_ROOTReference(t *testing.T) {
	got, ok := logicalnames.LogicalNameGetQualifier("ROOT/x/y(interface)")
	if !ok {
		t.Fatal("expected qualifier to be present")
	}
	want := "interface"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLogicalNameGetQualifier_TC17_ARTIFACTReference(t *testing.T) {
	got, ok := logicalnames.LogicalNameGetQualifier("ARTIFACT/x/y(id)")
	if !ok {
		t.Fatal("expected qualifier to be present")
	}
	want := "id"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLogicalNameGetQualifier_TC18_NoQualifier(t *testing.T) {
	got, ok := logicalnames.LogicalNameGetQualifier("ROOT/x/y")
	if ok {
		t.Errorf("expected no qualifier, but got %q", got)
	}
}

func TestLogicalNameGetQualifier_TC19_ROOTAlone(t *testing.T) {
	got, ok := logicalnames.LogicalNameGetQualifier("ROOT")
	if ok {
		t.Errorf("expected no qualifier, but got %q", got)
	}
}

// ----------------------------------------------------------------------------
// LogicalNameHasParent
// ----------------------------------------------------------------------------

func TestLogicalNameHasParent_TC20_ROOTAlone(t *testing.T) {
	if logicalnames.LogicalNameHasParent("ROOT") {
		t.Error("expected false for ROOT alone")
	}
}

func TestLogicalNameHasParent_TC21_ROOTWithPath(t *testing.T) {
	if !logicalnames.LogicalNameHasParent("ROOT/domain/config") {
		t.Error("expected true for ROOT/domain/config")
	}
}

func TestLogicalNameHasParent_TC22_ROOTWithQualifier(t *testing.T) {
	if !logicalnames.LogicalNameHasParent("ROOT/domain/config(interface)") {
		t.Error("expected true for ROOT/domain/config(interface)")
	}
}

func TestLogicalNameHasParent_TC23_ARTIFACTReference(t *testing.T) {
	if logicalnames.LogicalNameHasParent("ARTIFACT/x(y)") {
		t.Error("expected false for ARTIFACT reference")
	}
}

func TestLogicalNameHasParent_TC24_EmptyString(t *testing.T) {
	if logicalnames.LogicalNameHasParent("") {
		t.Error("expected false for empty string")
	}
}

// ----------------------------------------------------------------------------
// LogicalNameHasQualifier
// ----------------------------------------------------------------------------

func TestLogicalNameHasQualifier_TC25_WithoutQualifier(t *testing.T) {
	if logicalnames.LogicalNameHasQualifier("ROOT/x") {
		t.Error("expected false for ROOT/x")
	}
}

func TestLogicalNameHasQualifier_TC26_WithQualifier(t *testing.T) {
	if !logicalnames.LogicalNameHasQualifier("ROOT/x(y)") {
		t.Error("expected true for ROOT/x(y)")
	}
}

func TestLogicalNameHasQualifier_TC27_ARTIFACTWithQualifier(t *testing.T) {
	if !logicalnames.LogicalNameHasQualifier("ARTIFACT/x(y)") {
		t.Error("expected true for ARTIFACT/x(y)")
	}
}

func TestLogicalNameHasQualifier_TC28_ROOTAlone(t *testing.T) {
	if logicalnames.LogicalNameHasQualifier("ROOT") {
		t.Error("expected false for ROOT alone")
	}
}

func TestLogicalNameHasQualifier_TC29_EmptyString(t *testing.T) {
	if logicalnames.LogicalNameHasQualifier("") {
		t.Error("expected false for empty string")
	}
}

// ----------------------------------------------------------------------------
// LogicalNameIsArtifact
// ----------------------------------------------------------------------------

func TestLogicalNameIsArtifact_TC30_ARTIFACTReference(t *testing.T) {
	if !logicalnames.LogicalNameIsArtifact("ARTIFACT/x(y)") {
		t.Error("expected true for ARTIFACT/x(y)")
	}
}

func TestLogicalNameIsArtifact_TC31_ROOTReference(t *testing.T) {
	if logicalnames.LogicalNameIsArtifact("ROOT/x(y)") {
		t.Error("expected false for ROOT/x(y)")
	}
}

func TestLogicalNameIsArtifact_TC32_EmptyString(t *testing.T) {
	if logicalnames.LogicalNameIsArtifact("") {
		t.Error("expected false for empty string")
	}
}

// ----------------------------------------------------------------------------
// LogicalNameGetArtifactGenerator
// ----------------------------------------------------------------------------

func TestLogicalNameGetArtifactGenerator_TC33_SimpleArtifact(t *testing.T) {
	got, err := logicalnames.LogicalNameGetArtifactGenerator("ARTIFACT/x(y)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "ROOT/x"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLogicalNameGetArtifactGenerator_TC34_NestedArtifact(t *testing.T) {
	got, err := logicalnames.LogicalNameGetArtifactGenerator("ARTIFACT/x/y/z(id)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "ROOT/x/y/z"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLogicalNameGetArtifactGenerator_TC35_RejectsROOTReference(t *testing.T) {
	_, err := logicalnames.LogicalNameGetArtifactGenerator("ROOT/x(y)")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, logicalnames.ErrNotAnArtifactReference) {
		t.Errorf("got error %v, want ErrNotAnArtifactReference", err)
	}
}

func TestLogicalNameGetArtifactGenerator_TC36_ArtifactWithoutQualifier(t *testing.T) {
	got, err := logicalnames.LogicalNameGetArtifactGenerator("ARTIFACT/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "ROOT/x"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
