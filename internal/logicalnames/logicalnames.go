// code-from-spec: ROOT/golang/implementation/utils/logical_names@P8ZwFDAeIylcxOCgQZ8PMWc7NxQ

package logicalnames

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ErrUnsupportedReference is returned when a logical name is not a ROOT/ reference.
var ErrUnsupportedReference = errors.New("unsupported reference: not a ROOT/ reference")

// ErrInvalidPath is returned when a path is not a _node.md file under code-from-spec/.
var ErrInvalidPath = errors.New("invalid path: not a _node.md file under code-from-spec/")

// ErrNoParent is returned when the logical name is ROOT itself and has no parent.
var ErrNoParent = errors.New("no parent: ROOT has no parent")

// ErrNotARootReference is returned when the logical name is not a ROOT/ reference.
var ErrNotARootReference = errors.New("not a ROOT/ reference")

// ErrNotAnArtifactReference is returned when the logical name does not start with ARTIFACT/.
var ErrNotAnArtifactReference = errors.New("not an ARTIFACT/ reference")

// LogicalNameToPath converts a ROOT/ logical name to the PathCfs of the
// corresponding _node.md file. Strips any qualifier before resolving.
// Only accepts ROOT/ references (including ROOT itself).
//
// Errors:
//   - ErrUnsupportedReference: the logical name is not a ROOT/ reference
//     (neither ROOT nor ROOT/...).
func LogicalNameToPath(logicalName string) (*pathutils.PathCfs, error) {
	// Step 1: must start with "ROOT" (either "ROOT" or "ROOT/...")
	if !strings.HasPrefix(logicalName, "ROOT") {
		return nil, fmt.Errorf("%w: logical name must be a ROOT/ reference", ErrUnsupportedReference)
	}

	// Step 2: strip any qualifier
	bareName := LogicalNameStripQualifier(logicalName)

	// Step 3: if exactly "ROOT", return the root node path
	if bareName == "ROOT" {
		return &pathutils.PathCfs{Value: "code-from-spec/_node.md"}, nil
	}

	// Step 4: verify bare name starts with "ROOT/"
	if !strings.HasPrefix(bareName, "ROOT/") {
		return nil, fmt.Errorf("%w: logical name must be a ROOT/ reference", ErrUnsupportedReference)
	}

	// Step 5: strip the "ROOT/" prefix to get the relative segment
	relativeSegment := strings.TrimPrefix(bareName, "ROOT/")

	// Step 6: return the PathCfs
	return &pathutils.PathCfs{Value: "code-from-spec/" + relativeSegment + "/_node.md"}, nil
}

// LogicalNameFromPath derives the ROOT/ logical name from a _node.md file
// path. The inverse of LogicalNameToPath. Always returns a ROOT/ reference.
//
// Errors:
//   - ErrInvalidPath: the path is not a _node.md file under code-from-spec/.
func LogicalNameFromPath(cfsPath *pathutils.PathCfs) (string, error) {
	pathValue := cfsPath.Value

	// Step 2: must start with "code-from-spec/"
	if !strings.HasPrefix(pathValue, "code-from-spec/") {
		return "", fmt.Errorf("%w: path is not under code-from-spec/", ErrInvalidPath)
	}

	// Step 3: must end with "/_node.md" or be exactly "code-from-spec/_node.md"
	if pathValue != "code-from-spec/_node.md" && !strings.HasSuffix(pathValue, "/_node.md") {
		return "", fmt.Errorf("%w: path is not a _node.md file", ErrInvalidPath)
	}

	// Step 4: root node
	if pathValue == "code-from-spec/_node.md" {
		return "ROOT", nil
	}

	// Step 5: strip leading "code-from-spec/" and trailing "/_node.md"
	withoutPrefix := strings.TrimPrefix(pathValue, "code-from-spec/")
	middleSegment := strings.TrimSuffix(withoutPrefix, "/_node.md")

	// Step 6: return "ROOT/<middle segment>"
	return "ROOT/" + middleSegment, nil
}

// LogicalNameGetParent returns the logical name of the parent node.
// Strips any qualifier before computing the parent.
// Only accepts ROOT/ references (including ROOT itself, which returns NoParent).
//
// Errors:
//   - ErrNoParent: the logical name is ROOT itself.
//   - ErrNotARootReference: the logical name is not a ROOT/ reference
//     (neither ROOT nor ROOT/...).
func LogicalNameGetParent(logicalName string) (string, error) {
	// Step 1: must start with "ROOT"
	if !strings.HasPrefix(logicalName, "ROOT") {
		return "", fmt.Errorf("%w: logical name must be a ROOT/ reference", ErrNotARootReference)
	}

	// Step 2: strip any qualifier
	bareName := LogicalNameStripQualifier(logicalName)

	// Step 3: if exactly "ROOT", no parent
	if bareName == "ROOT" {
		return "", fmt.Errorf("%w: ROOT has no parent", ErrNoParent)
	}

	// Step 4: must contain "/" after "ROOT"
	if !strings.Contains(bareName, "/") {
		return "", fmt.Errorf("%w: logical name must be a ROOT/ reference", ErrNotARootReference)
	}

	// Step 5: find the last "/"
	lastSlash := strings.LastIndex(bareName, "/")
	parent := bareName[:lastSlash]

	// Step 6 & 7: return parent (handles both "ROOT" and deeper paths)
	return parent, nil
}

// LogicalNameGetQualifier extracts the parenthetical qualifier from a logical
// name. Returns ("", false) if no qualifier is present. Works with both ROOT/
// and ARTIFACT/ references.
//
// Examples:
//   - "ROOT/x/y(z)"      → ("z", true)
//   - "ROOT/x/y"         → ("", false)
//   - "ARTIFACT/x/y(id)" → ("id", true)
func LogicalNameGetQualifier(logicalName string) (qualifier string, ok bool) {
	// Step 1: check for "(<qualifier>)" at the end
	if !strings.HasSuffix(logicalName, ")") {
		return "", false
	}

	// Step 2: find the last "(" to get the qualifier
	openParen := strings.LastIndex(logicalName, "(")
	if openParen == -1 {
		return "", false
	}

	// Extract the qualifier between "(" and ")"
	qualifier = logicalName[openParen+1 : len(logicalName)-1]
	if qualifier == "" {
		return "", false
	}

	return qualifier, true
}

// LogicalNameStripQualifier returns the logical name without the parenthetical
// qualifier. If no qualifier is present, returns the input unchanged.
// Works with both ROOT/ and ARTIFACT/ references.
//
// Examples:
//   - "ROOT/x/y(z)"       → "ROOT/x/y"
//   - "ARTIFACT/x/y(id)"  → "ARTIFACT/x/y"
//   - "ROOT/x/y"          → "ROOT/x/y"
func LogicalNameStripQualifier(logicalName string) string {
	// Step 1: check for "(<qualifier>)" at the end
	if !strings.HasSuffix(logicalName, ")") {
		return logicalName
	}

	// Step 2: find the last "(" and strip from there
	openParen := strings.LastIndex(logicalName, "(")
	if openParen == -1 {
		return logicalName
	}

	return logicalName[:openParen]
}

// LogicalNameHasParent returns true if the logical name is a ROOT/ reference
// other than ROOT itself. Returns false for ROOT, ARTIFACT/ references, and
// unrecognized prefixes.
func LogicalNameHasParent(logicalName string) bool {
	// Step 1: strip qualifier
	bareName := LogicalNameStripQualifier(logicalName)

	// Step 2: if exactly "ROOT", return false
	if bareName == "ROOT" {
		return false
	}

	// Step 3: if starts with "ROOT/", return true
	if strings.HasPrefix(bareName, "ROOT/") {
		return true
	}

	// Step 4: return false
	return false
}

// LogicalNameHasQualifier returns true if the logical name contains a
// parenthetical qualifier. Works with both ROOT/ and ARTIFACT/ references.
func LogicalNameHasQualifier(logicalName string) bool {
	_, ok := LogicalNameGetQualifier(logicalName)
	return ok
}

// LogicalNameIsArtifact returns true if the logical name starts with ARTIFACT/.
func LogicalNameIsArtifact(logicalName string) bool {
	return strings.HasPrefix(logicalName, "ARTIFACT/")
}

// LogicalNameGetArtifactGenerator returns the ROOT/ logical name of the node
// that generates the referenced artifact. Strips the ARTIFACT/ prefix and any
// qualifier.
//
// Examples:
//   - "ARTIFACT/x/y(id)" → "ROOT/x/y"
//   - "ARTIFACT/x/y"     → "ROOT/x/y"
//
// Errors:
//   - ErrNotAnArtifactReference: the logical name does not start with ARTIFACT/.
func LogicalNameGetArtifactGenerator(logicalName string) (string, error) {
	// Step 1: must start with "ARTIFACT/"
	if !strings.HasPrefix(logicalName, "ARTIFACT/") {
		return "", fmt.Errorf("%w: logical name does not start with ARTIFACT/", ErrNotAnArtifactReference)
	}

	// Step 2: strip any qualifier
	bareName := LogicalNameStripQualifier(logicalName)

	// Step 3: strip the "ARTIFACT/" prefix
	segment := strings.TrimPrefix(bareName, "ARTIFACT/")

	// Step 4: return "ROOT/<segment>"
	return "ROOT/" + segment, nil
}
