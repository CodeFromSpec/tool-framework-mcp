// code-from-spec: ROOT/golang/tests/utils/logical_names@YEe2OHRyUc8CEc2n9WddM0Mc6_0
package logicalnames_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func TestLogicalNameToPath_ROOTAlone(t *testing.T) {
	result, err := logicalnames.LogicalNameToPath("ROOT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Value != "code-from-spec/_node.md" {
		t.Errorf("Value = %q, want %q", result.Value, "code-from-spec/_node.md")
	}
}

func TestLogicalNameToPath_ROOTWithPath(t *testing.T) {
	result, err := logicalnames.LogicalNameToPath("ROOT/payments/processor")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Value != "code-from-spec/payments/processor/_node.md" {
		t.Errorf("Value = %q, want %q", result.Value, "code-from-spec/payments/processor/_node.md")
	}
}

func TestLogicalNameToPath_StripsQualifierBeforeResolving(t *testing.T) {
	result, err := logicalnames.LogicalNameToPath("ROOT/x/y(interface)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Value != "code-from-spec/x/y/_node.md" {
		t.Errorf("Value = %q, want %q", result.Value, "code-from-spec/x/y/_node.md")
	}
}

func TestLogicalNameToPath_RejectsARTIFACTReference(t *testing.T) {
	_, err := logicalnames.LogicalNameToPath("ARTIFACT/x")
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("error = %v, want ErrUnsupportedReference", err)
	}
}

func TestLogicalNameToPath_RejectsUnrecognizedPrefix(t *testing.T) {
	_, err := logicalnames.LogicalNameToPath("UNKNOWN/something")
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("error = %v, want ErrUnsupportedReference", err)
	}
}

func TestLogicalNameToPath_RejectsEmptyString(t *testing.T) {
	_, err := logicalnames.LogicalNameToPath("")
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("error = %v, want ErrUnsupportedReference", err)
	}
}

func TestLogicalNameFromPath_RootNode(t *testing.T) {
	result, err := logicalnames.LogicalNameFromPath(&pathutils.PathCfs{Value: "code-from-spec/_node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ROOT" {
		t.Errorf("result = %q, want %q", result, "ROOT")
	}
}

func TestLogicalNameFromPath_NestedNode(t *testing.T) {
	result, err := logicalnames.LogicalNameFromPath(&pathutils.PathCfs{Value: "code-from-spec/x/y/_node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ROOT/x/y" {
		t.Errorf("result = %q, want %q", result, "ROOT/x/y")
	}
}

func TestLogicalNameFromPath_RejectsNonNodePath(t *testing.T) {
	_, err := logicalnames.LogicalNameFromPath(&pathutils.PathCfs{Value: "internal/config/config.go"})
	if !errors.Is(err, logicalnames.ErrInvalidPath) {
		t.Errorf("error = %v, want ErrInvalidPath", err)
	}
}

func TestLogicalNameFromPath_RejectsPathWithoutNodeMd(t *testing.T) {
	_, err := logicalnames.LogicalNameFromPath(&pathutils.PathCfs{Value: "code-from-spec/x/y/output.md"})
	if !errors.Is(err, logicalnames.ErrInvalidPath) {
		t.Errorf("error = %v, want ErrInvalidPath", err)
	}
}

func TestLogicalNameGetParent_ROOTXParentIsROOT(t *testing.T) {
	result, err := logicalnames.LogicalNameGetParent("ROOT/domain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ROOT" {
		t.Errorf("result = %q, want %q", result, "ROOT")
	}
}

func TestLogicalNameGetParent_ROOTXYParentIsROOTX(t *testing.T) {
	result, err := logicalnames.LogicalNameGetParent("ROOT/domain/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ROOT/domain" {
		t.Errorf("result = %q, want %q", result, "ROOT/domain")
	}
}

func TestLogicalNameGetParent_StripsQualifierBeforeComputingParent(t *testing.T) {
	result, err := logicalnames.LogicalNameGetParent("ROOT/domain/config(interface)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ROOT/domain" {
		t.Errorf("result = %q, want %q", result, "ROOT/domain")
	}
}

func TestLogicalNameGetParent_ROOTHasNoParent(t *testing.T) {
	_, err := logicalnames.LogicalNameGetParent("ROOT")
	if !errors.Is(err, logicalnames.ErrNoParent) {
		t.Errorf("error = %v, want ErrNoParent", err)
	}
}

func TestLogicalNameGetParent_RejectsARTIFACTReference(t *testing.T) {
	_, err := logicalnames.LogicalNameGetParent("ARTIFACT/x")
	if !errors.Is(err, logicalnames.ErrNotARootReference) {
		t.Errorf("error = %v, want ErrNotARootReference", err)
	}
}

func TestLogicalNameGetQualifier_ExtractsQualifierFromROOTReference(t *testing.T) {
	qualifier, ok := logicalnames.LogicalNameGetQualifier("ROOT/x/y(interface)")
	if !ok {
		t.Fatalf("ok = false, want true")
	}
	if qualifier != "interface" {
		t.Errorf("qualifier = %q, want %q", qualifier, "interface")
	}
}

func TestLogicalNameGetQualifier_ARTIFACTWithoutQualifierReturnsAbsent(t *testing.T) {
	qualifier, ok := logicalnames.LogicalNameGetQualifier("ARTIFACT/x/y")
	if ok {
		t.Errorf("ok = true, want false")
	}
	if qualifier != "" {
		t.Errorf("qualifier = %q, want empty", qualifier)
	}
}

func TestLogicalNameGetQualifier_ReturnsAbsentWhenNoQualifier(t *testing.T) {
	qualifier, ok := logicalnames.LogicalNameGetQualifier("ROOT/x/y")
	if ok {
		t.Errorf("ok = true, want false")
	}
	if qualifier != "" {
		t.Errorf("qualifier = %q, want empty", qualifier)
	}
}

func TestLogicalNameGetQualifier_ReturnsAbsentForROOTAlone(t *testing.T) {
	qualifier, ok := logicalnames.LogicalNameGetQualifier("ROOT")
	if ok {
		t.Errorf("ok = true, want false")
	}
	if qualifier != "" {
		t.Errorf("qualifier = %q, want empty", qualifier)
	}
}

func TestLogicalNameStripQualifier_StripsQualifierFromROOTReference(t *testing.T) {
	result := logicalnames.LogicalNameStripQualifier("ROOT/x/y(interface)")
	if result != "ROOT/x/y" {
		t.Errorf("result = %q, want %q", result, "ROOT/x/y")
	}
}

func TestLogicalNameStripQualifier_ARTIFACTWithoutQualifierReturnsUnchanged(t *testing.T) {
	result := logicalnames.LogicalNameStripQualifier("ARTIFACT/x/y")
	if result != "ARTIFACT/x/y" {
		t.Errorf("result = %q, want %q", result, "ARTIFACT/x/y")
	}
}

func TestLogicalNameStripQualifier_NoQualifierReturnsUnchanged(t *testing.T) {
	result := logicalnames.LogicalNameStripQualifier("ROOT/x/y")
	if result != "ROOT/x/y" {
		t.Errorf("result = %q, want %q", result, "ROOT/x/y")
	}
}

func TestLogicalNameStripQualifier_ROOTAloneReturnsUnchanged(t *testing.T) {
	result := logicalnames.LogicalNameStripQualifier("ROOT")
	if result != "ROOT" {
		t.Errorf("result = %q, want %q", result, "ROOT")
	}
}

func TestLogicalNameStripQualifier_EmptyStringReturnsUnchanged(t *testing.T) {
	result := logicalnames.LogicalNameStripQualifier("")
	if result != "" {
		t.Errorf("result = %q, want empty", result)
	}
}

func TestLogicalNameHasParent_ROOTAlone(t *testing.T) {
	result := logicalnames.LogicalNameHasParent("ROOT")
	if result {
		t.Errorf("result = true, want false")
	}
}

func TestLogicalNameHasParent_ROOTWithPath(t *testing.T) {
	result := logicalnames.LogicalNameHasParent("ROOT/domain/config")
	if !result {
		t.Errorf("result = false, want true")
	}
}

func TestLogicalNameHasParent_ROOTWithQualifier(t *testing.T) {
	result := logicalnames.LogicalNameHasParent("ROOT/domain/config(interface)")
	if !result {
		t.Errorf("result = false, want true")
	}
}

func TestLogicalNameHasParent_ARTIFACTReference(t *testing.T) {
	result := logicalnames.LogicalNameHasParent("ARTIFACT/x")
	if result {
		t.Errorf("result = true, want false")
	}
}

func TestLogicalNameHasParent_EmptyString(t *testing.T) {
	result := logicalnames.LogicalNameHasParent("")
	if result {
		t.Errorf("result = true, want false")
	}
}

func TestLogicalNameHasQualifier_WithoutQualifier(t *testing.T) {
	result := logicalnames.LogicalNameHasQualifier("ROOT/x")
	if result {
		t.Errorf("result = true, want false")
	}
}

func TestLogicalNameHasQualifier_WithQualifier(t *testing.T) {
	result := logicalnames.LogicalNameHasQualifier("ROOT/x(y)")
	if !result {
		t.Errorf("result = false, want true")
	}
}

func TestLogicalNameHasQualifier_ARTIFACTWithoutQualifier(t *testing.T) {
	result := logicalnames.LogicalNameHasQualifier("ARTIFACT/x")
	if result {
		t.Errorf("result = true, want false")
	}
}

func TestLogicalNameHasQualifier_ROOTAlone(t *testing.T) {
	result := logicalnames.LogicalNameHasQualifier("ROOT")
	if result {
		t.Errorf("result = true, want false")
	}
}

func TestLogicalNameHasQualifier_EmptyString(t *testing.T) {
	result := logicalnames.LogicalNameHasQualifier("")
	if result {
		t.Errorf("result = true, want false")
	}
}

func TestLogicalNameIsArtifact_ARTIFACTReference(t *testing.T) {
	result := logicalnames.LogicalNameIsArtifact("ARTIFACT/x")
	if !result {
		t.Errorf("result = false, want true")
	}
}

func TestLogicalNameIsArtifact_ROOTReference(t *testing.T) {
	result := logicalnames.LogicalNameIsArtifact("ROOT/x(y)")
	if result {
		t.Errorf("result = true, want false")
	}
}

func TestLogicalNameIsArtifact_EmptyString(t *testing.T) {
	result := logicalnames.LogicalNameIsArtifact("")
	if result {
		t.Errorf("result = true, want false")
	}
}

func TestLogicalNameGetArtifactGenerator_SimpleArtifact(t *testing.T) {
	result, err := logicalnames.LogicalNameGetArtifactGenerator("ARTIFACT/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ROOT/x" {
		t.Errorf("result = %q, want %q", result, "ROOT/x")
	}
}

func TestLogicalNameGetArtifactGenerator_NestedArtifact(t *testing.T) {
	result, err := logicalnames.LogicalNameGetArtifactGenerator("ARTIFACT/x/y/z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ROOT/x/y/z" {
		t.Errorf("result = %q, want %q", result, "ROOT/x/y/z")
	}
}

func TestLogicalNameGetArtifactGenerator_RejectsROOTReference(t *testing.T) {
	_, err := logicalnames.LogicalNameGetArtifactGenerator("ROOT/x(y)")
	if !errors.Is(err, logicalnames.ErrNotAnArtifactReference) {
		t.Errorf("error = %v, want ErrNotAnArtifactReference", err)
	}
}
