// code-from-spec: ROOT/golang/implementation/utils/logical_names@tGu6tqY64B-Fi47kndUeL4jpwHE

package logicalnames

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

var (
	// ErrUnsupportedReference is returned when a logical name does not
	// start with ROOT/.
	ErrUnsupportedReference = errors.New("unsupported reference")

	// ErrInvalidPath is returned when a path is not a _node.md file
	// under code-from-spec/.
	ErrInvalidPath = errors.New("invalid path")

	// ErrNoParent is returned when the logical name is ROOT itself and
	// has no parent.
	ErrNoParent = errors.New("no parent")

	// ErrNotARootReference is returned when the logical name does not
	// start with ROOT/.
	ErrNotARootReference = errors.New("not a ROOT reference")

	// ErrNotAnArtifactReference is returned when the logical name does
	// not start with ARTIFACT/.
	ErrNotAnArtifactReference = errors.New("not an artifact reference")
)

// stripQualifier removes any parenthetical qualifier suffix from a logical
// name segment. For example, "ROOT/x/y(z)" becomes "ROOT/x/y".
func stripQualifier(logicalName string) string {
	idx := strings.LastIndex(logicalName, "(")
	if idx == -1 {
		return logicalName
	}
	return logicalName[:idx]
}

// isRootReference returns true if the logical name starts with "ROOT/" or
// is exactly "ROOT".
func isRootReference(logicalName string) bool {
	return logicalName == "ROOT" || strings.HasPrefix(logicalName, "ROOT/")
}

// LogicalNameToPath converts a ROOT/ logical name to the PathCfs of
// the corresponding _node.md file. Strips any qualifier before
// resolving. Only accepts ROOT/ references.
//
// Possible errors:
//   - ErrUnsupportedReference: the logical name does not start with ROOT/.
func LogicalNameToPath(logicalName string) (*pathutils.PathCfs, error) {
	if !isRootReference(logicalName) {
		return nil, fmt.Errorf("%w", ErrUnsupportedReference)
	}

	logicalName = stripQualifier(logicalName)

	if logicalName == "ROOT" {
		return &pathutils.PathCfs{Value: "code-from-spec/_node.md"}, nil
	}

	relative := strings.TrimPrefix(logicalName, "ROOT/")
	return &pathutils.PathCfs{Value: "code-from-spec/" + relative + "/_node.md"}, nil
}

// LogicalNameFromPath derives the ROOT/ logical name from a _node.md
// file path. The inverse of LogicalNameToPath. Always returns a ROOT/
// reference.
//
// Possible errors:
//   - ErrInvalidPath: the path is not a _node.md file under code-from-spec/.
func LogicalNameFromPath(cfsPath *pathutils.PathCfs) (string, error) {
	p := cfsPath.Value

	if !strings.HasPrefix(p, "code-from-spec/") {
		return "", fmt.Errorf("%w", ErrInvalidPath)
	}

	if p != "code-from-spec/_node.md" && !strings.HasSuffix(p, "/_node.md") {
		return "", fmt.Errorf("%w", ErrInvalidPath)
	}

	if p == "code-from-spec/_node.md" {
		return "ROOT", nil
	}

	relative := strings.TrimPrefix(p, "code-from-spec/")
	relative = strings.TrimSuffix(relative, "/_node.md")
	return "ROOT/" + relative, nil
}

// LogicalNameGetParent returns the logical name of the parent node.
// Strips any qualifier before computing the parent. Only accepts ROOT/
// references.
//
// Possible errors:
//   - ErrNoParent: the logical name is ROOT itself.
//   - ErrNotARootReference: the logical name does not start with ROOT/.
func LogicalNameGetParent(logicalName string) (string, error) {
	if !isRootReference(logicalName) {
		return "", fmt.Errorf("%w", ErrNotARootReference)
	}

	logicalName = stripQualifier(logicalName)

	if logicalName == "ROOT" {
		return "", fmt.Errorf("%w", ErrNoParent)
	}

	idx := strings.LastIndex(logicalName, "/")
	return logicalName[:idx], nil
}

// LogicalNameGetQualifier extracts the parenthetical qualifier from a
// logical name. Returns ("", false) if no qualifier is present. Works
// with both ROOT/ and ARTIFACT/ references.
func LogicalNameGetQualifier(logicalName string) (string, bool) {
	openIdx := strings.LastIndex(logicalName, "(")
	if openIdx == -1 {
		return "", false
	}

	closeIdx := strings.Index(logicalName[openIdx:], ")")
	if closeIdx == -1 {
		return "", false
	}

	qualifier := logicalName[openIdx+1 : openIdx+closeIdx]
	if qualifier == "" {
		return "", false
	}

	return qualifier, true
}

// LogicalNameHasParent returns true if the logical name is a ROOT/
// reference other than ROOT itself. Returns false for ROOT, ARTIFACT/
// references, and unrecognized prefixes.
func LogicalNameHasParent(logicalName string) bool {
	return strings.HasPrefix(logicalName, "ROOT/")
}

// LogicalNameHasQualifier returns true if the logical name contains a
// parenthetical qualifier. Works with both ROOT/ and ARTIFACT/
// references.
func LogicalNameHasQualifier(logicalName string) bool {
	_, ok := LogicalNameGetQualifier(logicalName)
	return ok
}

// LogicalNameIsArtifact returns true if the logical name starts with
// ARTIFACT/.
func LogicalNameIsArtifact(logicalName string) bool {
	return strings.HasPrefix(logicalName, "ARTIFACT/")
}

// LogicalNameGetArtifactGenerator returns the ROOT/ logical name of
// the node that generates the referenced artifact. Strips the ARTIFACT/
// prefix and any qualifier.
//
// Possible errors:
//   - ErrNotAnArtifactReference: the logical name does not start with ARTIFACT/.
func LogicalNameGetArtifactGenerator(logicalName string) (string, error) {
	if !strings.HasPrefix(logicalName, "ARTIFACT/") {
		return "", fmt.Errorf("%w", ErrNotAnArtifactReference)
	}

	relative := strings.TrimPrefix(logicalName, "ARTIFACT/")
	relative = stripQualifier(relative)
	return "ROOT/" + relative, nil
}
