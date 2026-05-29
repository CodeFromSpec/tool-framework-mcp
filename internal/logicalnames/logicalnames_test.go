// code-from-spec: ROOT/golang/tests/utils/logical_names@nD77pyBGpFYClluuF-FHCg-0Ssg

package logicalnames_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

// --- LogicalNameToPath ---

func TestLogicalNameToPath_RootAlone(t *testing.T) {
	// TC-01
	got, err := logicalnames.LogicalNameToPath("ROOT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Value != "code-from-spec/_node.md" {
		t.Errorf("got %q, want %q", got.Value, "code-from-spec/_node.md")
	}
}

func TestLogicalNameToPath_RootWithPath(t *testing.T) {
	// TC-02
	got, err := logicalnames.LogicalNameToPath("ROOT/payments/processor")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Value != "code-from-spec/payments/processor/_node.md" {
		t.Errorf("got %q, want %q", got.Value, "code-from-spec/payments/processor/_node.md")
	}
}

func TestLogicalNameToPath_StripsQualifierBeforeResolving(t *testing.T) {
	// TC-03
	got, err := logicalnames.LogicalNameToPath("ROOT/x/y(interface)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Value != "code-from-spec/x/y/_node.md" {
		t.Errorf("got %q, want %q", got.Value, "code-from-spec/x/y/_node.md")
	}
}

func TestLogicalNameToPath_RejectsArtifactReference(t *testing.T) {
	// TC-04
	_, err := logicalnames.LogicalNameToPath("ARTIFACT/x(y)")
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("got error %v, want ErrUnsupportedReference", err)
	}
}

func TestLogicalNameToPath_RejectsUnrecognizedPrefix(t *testing.T) {
	// TC-05
	_, err := logicalnames.LogicalNameToPath("UNKNOWN/something")
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("got error %v, want ErrUnsupportedReference", err)
	}
}

func TestLogicalNameToPath_RejectsEmptyString(t *testing.T) {
	// TC-06
	_, err := logicalnames.LogicalNameToPath("")
	if !errors.Is(err, logicalnames.ErrUnsupportedReference) {
		t.Errorf("got error %v, want ErrUnsupportedReference", err)
	}
}

// --- LogicalNameFromPath ---

func TestLogicalNameFromPath_RootNode(t *testing.T) {
	// TC-07
	path := &pathutils.PathCfs{Value: "code-from-spec/_node.md"}
	got, err := logicalnames.LogicalNameFromPath(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ROOT" {
		t.Errorf("got %q, want %q", got, "ROOT")
	}
}

func TestLogicalNameFromPath_NestedNode(t *testing.T) {
	// TC-08
	path := &pathutils.PathCfs{Value: "code-from-spec/x/y/_node.md"}
	got, err := logicalnames.LogicalNameFromPath(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ROOT/x/y" {
		t.Errorf("got %q, want %q", got, "ROOT/x/y")
	}
}

func TestLogicalNameFromPath_RejectsNonNodePath(t *testing.T) {
	// TC-09
	path := &pathutils.PathCfs{Value: "internal/config/config.go"}
	_, err := logicalnames.LogicalNameFromPath(path)
	if !errors.Is(err, logicalnames.ErrInvalidPath) {
		t.Errorf("got error %v, want ErrInvalidPath", err)
	}
}

func TestLogicalNameFromPath_RejectsPathWithoutNodeMd(t *testing.T) {
	// TC-10
	path := &pathutils.PathCfs{Value: "code-from-spec/x/y/output.md"}
	_, err := logicalnames.LogicalNameFromPath(path)
	if !errors.Is(err, logicalnames.ErrInvalidPath) {
		t.Errorf("got error %v, want ErrInvalidPath", err)
	}
}

// --- LogicalNameGetParent ---

func TestLogicalNameGetParent_SingleSegment(t *testing.T) {
	// TC-11
	got, err := logicalnames.LogicalNameGetParent("ROOT/domain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ROOT" {
		t.Errorf("got %q, want %q", got, "ROOT")
	}
}

func TestLogicalNameGetParent_TwoSegments(t *testing.T) {
	// TC-12
	got, err := logicalnames.LogicalNameGetParent("ROOT/domain/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ROOT/domain" {
		t.Errorf("got %q, want %q", got, "ROOT/domain")
	}
}

func TestLogicalNameGetParent_StripsQualifierBeforeComputing(t *testing.T) {
	// TC-13
	got, err := logicalnames.LogicalNameGetParent("ROOT/domain/config(interface)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ROOT/domain" {
		t.Errorf("got %q, want %q", got, "ROOT/domain")
	}
}

func TestLogicalNameGetParent_RootHasNoParent(t *testing.T) {
	// TC-14
	_, err := logicalnames.LogicalNameGetParent("ROOT")
	if !errors.Is(err, logicalnames.ErrNoParent) {
		t.Errorf("got error %v, want ErrNoParent", err)
	}
}

func TestLogicalNameGetParent_RejectsArtifactReference(t *testing.T) {
	// TC-15
	_, err := logicalnames.LogicalNameGetParent("ARTIFACT/x(y)")
	if !errors.Is(err, logicalnames.ErrNotARootReference) {
		t.Errorf("got error %v, want ErrNotARootReference", err)
	}
}

// --- LogicalNameGetQualifier ---

func TestLogicalNameGetQualifier_RootReference(t *testing.T) {
	// TC-16
	got, ok := logicalnames.LogicalNameGetQualifier("ROOT/x/y(interface)")
	if !ok {
		t.Fatal("expected qualifier to be present")
	}
	if got != "interface" {
		t.Errorf("got %q, want %q", got, "interface")
	}
}

func TestLogicalNameGetQualifier_ArtifactReference(t *testing.T) {
	// TC-17
	got, ok := logicalnames.LogicalNameGetQualifier("ARTIFACT/x/y(id)")
	if !ok {
		t.Fatal("expected qualifier to be present")
	}
	if got != "id" {
		t.Errorf("got %q, want %q", got, "id")
	}
}

func TestLogicalNameGetQualifier_AbsentWhenNoQualifier(t *testing.T) {
	// TC-18
	_, ok := logicalnames.LogicalNameGetQualifier("ROOT/x/y")
	if ok {
		t.Error("expected no qualifier, but got one")
	}
}

func TestLogicalNameGetQualifier_AbsentForRootAlone(t *testing.T) {
	// TC-19
	_, ok := logicalnames.LogicalNameGetQualifier("ROOT")
	if ok {
		t.Error("expected no qualifier, but got one")
	}
}

// --- LogicalNameStripQualifier ---

func TestLogicalNameStripQualifier_RootReference(t *testing.T) {
	// TC-20
	got := logicalnames.LogicalNameStripQualifier("ROOT/x/y(interface)")
	if got != "ROOT/x/y" {
		t.Errorf("got %q, want %q", got, "ROOT/x/y")
	}
}

func TestLogicalNameStripQualifier_ArtifactReference(t *testing.T) {
	// TC-21
	got := logicalnames.LogicalNameStripQualifier("ARTIFACT/x/y(id)")
	if got != "ARTIFACT/x/y" {
		t.Errorf("got %q, want %q", got, "ARTIFACT/x/y")
	}
}

func TestLogicalNameStripQualifier_NoQualifier(t *testing.T) {
	// TC-22
	got := logicalnames.LogicalNameStripQualifier("ROOT/x/y")
	if got != "ROOT/x/y" {
		t.Errorf("got %q, want %q", got, "ROOT/x/y")
	}
}

func TestLogicalNameStripQualifier_RootAlone(t *testing.T) {
	// TC-23
	got := logicalnames.LogicalNameStripQualifier("ROOT")
	if got != "ROOT" {
		t.Errorf("got %q, want %q", got, "ROOT")
	}
}

func TestLogicalNameStripQualifier_EmptyString(t *testing.T) {
	// TC-24
	got := logicalnames.LogicalNameStripQualifier("")
	if got != "" {
		t.Errorf("got %q, want %q", got, "")
	}
}

// --- LogicalNameHasParent ---

func TestLogicalNameHasParent_RootAlone(t *testing.T) {
	// TC-25
	if logicalnames.LogicalNameHasParent("ROOT") {
		t.Error("expected false for ROOT alone")
	}
}

func TestLogicalNameHasParent_RootWithPath(t *testing.T) {
	// TC-26
	if !logicalnames.LogicalNameHasParent("ROOT/domain/config") {
		t.Error("expected true for ROOT/domain/config")
	}
}

func TestLogicalNameHasParent_RootWithQualifier(t *testing.T) {
	// TC-27
	if !logicalnames.LogicalNameHasParent("ROOT/domain/config(interface)") {
		t.Error("expected true for ROOT/domain/config(interface)")
	}
}

func TestLogicalNameHasParent_ArtifactReference(t *testing.T) {
	// TC-28
	if logicalnames.LogicalNameHasParent("ARTIFACT/x(y)") {
		t.Error("expected false for ARTIFACT reference")
	}
}

func TestLogicalNameHasParent_EmptyString(t *testing.T) {
	// TC-29
	if logicalnames.LogicalNameHasParent("") {
		t.Error("expected false for empty string")
	}
}

// --- LogicalNameHasQualifier ---

func TestLogicalNameHasQualifier_WithoutQualifier(t *testing.T) {
	// TC-30
	if logicalnames.LogicalNameHasQualifier("ROOT/x") {
		t.Error("expected false for ROOT/x")
	}
}

func TestLogicalNameHasQualifier_WithQualifier(t *testing.T) {
	// TC-31
	if !logicalnames.LogicalNameHasQualifier("ROOT/x(y)") {
		t.Error("expected true for ROOT/x(y)")
	}
}

func TestLogicalNameHasQualifier_ArtifactWithQualifier(t *testing.T) {
	// TC-32
	if !logicalnames.LogicalNameHasQualifier("ARTIFACT/x(y)") {
		t.Error("expected true for ARTIFACT/x(y)")
	}
}

func TestLogicalNameHasQualifier_RootAlone(t *testing.T) {
	// TC-33
	if logicalnames.LogicalNameHasQualifier("ROOT") {
		t.Error("expected false for ROOT alone")
	}
}

func TestLogicalNameHasQualifier_EmptyString(t *testing.T) {
	// TC-34
	if logicalnames.LogicalNameHasQualifier("") {
		t.Error("expected false for empty string")
	}
}

// --- LogicalNameIsArtifact ---

func TestLogicalNameIsArtifact_ArtifactReference(t *testing.T) {
	// TC-35
	if !logicalnames.LogicalNameIsArtifact("ARTIFACT/x(y)") {
		t.Error("expected true for ARTIFACT/x(y)")
	}
}

func TestLogicalNameIsArtifact_RootReference(t *testing.T) {
	// TC-36
	if logicalnames.LogicalNameIsArtifact("ROOT/x(y)") {
		t.Error("expected false for ROOT/x(y)")
	}
}

func TestLogicalNameIsArtifact_EmptyString(t *testing.T) {
	// TC-37
	if logicalnames.LogicalNameIsArtifact("") {
		t.Error("expected false for empty string")
	}
}

// --- LogicalNameGetArtifactGenerator ---

func TestLogicalNameGetArtifactGenerator_SimpleArtifact(t *testing.T) {
	// TC-38
	got, err := logicalnames.LogicalNameGetArtifactGenerator("ARTIFACT/x(y)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ROOT/x" {
		t.Errorf("got %q, want %q", got, "ROOT/x")
	}
}

func TestLogicalNameGetArtifactGenerator_NestedArtifact(t *testing.T) {
	// TC-39
	got, err := logicalnames.LogicalNameGetArtifactGenerator("ARTIFACT/x/y/z(id)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ROOT/x/y/z" {
		t.Errorf("got %q, want %q", got, "ROOT/x/y/z")
	}
}

func TestLogicalNameGetArtifactGenerator_RejectsRootReference(t *testing.T) {
	// TC-40
	_, err := logicalnames.LogicalNameGetArtifactGenerator("ROOT/x(y)")
	if !errors.Is(err, logicalnames.ErrNotAnArtifactReference) {
		t.Errorf("got error %v, want ErrNotAnArtifactReference", err)
	}
}

func TestLogicalNameGetArtifactGenerator_WithoutQualifier(t *testing.T) {
	// TC-41
	got, err := logicalnames.LogicalNameGetArtifactGenerator("ARTIFACT/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ROOT/x" {
		t.Errorf("got %q, want %q", got, "ROOT/x")
	}
}
