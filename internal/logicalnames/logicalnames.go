// code-from-spec: ROOT/golang/implementation/utils/logical_names@zqUOKWLUbYJFQ4mK079YpI6YXak

// Package logicalnames provides functions for working with logical names —
// the ROOT/ and ARTIFACT/ reference strings used throughout the framework
// to identify spec nodes and generated artifacts.
package logicalnames

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ErrUnsupportedReference is returned when a logical name is not a
// ROOT/ reference (neither ROOT nor ROOT/...).
var ErrUnsupportedReference = errors.New("unsupported reference: expected a ROOT/ reference")

// ErrInvalidPath is returned when a path is not a _node.md file
// under code-from-spec/.
var ErrInvalidPath = errors.New("invalid path: not a _node.md file under code-from-spec/")

// ErrNoParent is returned when the logical name is ROOT itself and
// has no parent.
var ErrNoParent = errors.New("no parent: ROOT has no parent node")

// ErrNotARootReference is returned when the logical name is not a
// ROOT/ reference (neither ROOT nor ROOT/...).
var ErrNotARootReference = errors.New("not a ROOT/ reference")

// ErrNotAnArtifactReference is returned when the logical name does
// not start with ARTIFACT/.
var ErrNotAnArtifactReference = errors.New("not an ARTIFACT/ reference")

// LogicalNameToPath converts a ROOT/ logical name to the PathCfs of
// the corresponding _node.md file. Strips any qualifier before resolving.
// Only accepts ROOT/ references (including ROOT itself).
//
// Errors:
//   - ErrUnsupportedReference: the logical name is not a ROOT/ reference
//     (neither ROOT nor ROOT/...).
func LogicalNameToPath(logicalName string) (*pathutils.PathCfs, error) {
	// Step 1: must start with "ROOT" (either exactly "ROOT" or "ROOT/..." or "ROOT(...)")
	if logicalName != "ROOT" && !strings.HasPrefix(logicalName, "ROOT/") &&
		!strings.HasPrefix(logicalName, "ROOT(") {
		return nil, fmt.Errorf("%w: logical name must be a ROOT/ reference", ErrUnsupportedReference)
	}

	// Step 2: strip any qualifier
	bare := LogicalNameStripQualifier(logicalName)

	// Step 3: if bare is exactly "ROOT", return root node path
	if bare == "ROOT" {
		return &pathutils.PathCfs{Value: "code-from-spec/_node.md"}, nil
	}

	// Step 4: verify it starts with "ROOT/"
	if !strings.HasPrefix(bare, "ROOT/") {
		return nil, fmt.Errorf("%w: logical name must be a ROOT/ reference", ErrUnsupportedReference)
	}

	// Step 5: strip "ROOT/" prefix to get relative segment
	segment := strings.TrimPrefix(bare, "ROOT/")

	// Step 6: return the PathCfs
	return &pathutils.PathCfs{Value: "code-from-spec/" + segment + "/_node.md"}, nil
}

// LogicalNameFromPath derives the ROOT/ logical name from a _node.md
// file path. The inverse of LogicalNameToPath. Always returns a ROOT/
// reference.
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

	// Step 4: if exactly the root node path, return "ROOT"
	if pathValue == "code-from-spec/_node.md" {
		return "ROOT", nil
	}

	// Step 5: strip prefix "code-from-spec/" and suffix "/_node.md"
	middle := strings.TrimPrefix(pathValue, "code-from-spec/")
	middle = strings.TrimSuffix(middle, "/_node.md")

	// Step 6: return "ROOT/<middle>"
	return "ROOT/" + middle, nil
}

// LogicalNameGetParent returns the logical name of the parent node.
// Strips any qualifier before computing the parent. Only accepts ROOT/
// references (including ROOT itself, which returns ErrNoParent).
//
// Errors:
//   - ErrNoParent: the logical name is ROOT itself.
//   - ErrNotARootReference: the logical name is not a ROOT/ reference
//     (neither ROOT nor ROOT/...).
func LogicalNameGetParent(logicalName string) (string, error) {
	// Step 1: must start with "ROOT" (either exactly "ROOT" or "ROOT/..." or "ROOT(...)")
	if logicalName != "ROOT" && !strings.HasPrefix(logicalName, "ROOT/") &&
		!strings.HasPrefix(logicalName, "ROOT(") {
		return "", fmt.Errorf("%w: logical name must be a ROOT/ reference", ErrNotARootReference)
	}

	// Step 2: strip any qualifier
	bare := LogicalNameStripQualifier(logicalName)

	// Step 3: if bare is exactly "ROOT", it has no parent
	if bare == "ROOT" {
		return "", fmt.Errorf("%w: ROOT has no parent", ErrNoParent)
	}

	// Step 4: must contain "/" after "ROOT" (i.e. starts with "ROOT/")
	if !strings.HasPrefix(bare, "ROOT/") {
		return "", fmt.Errorf("%w: logical name must be a ROOT/ reference", ErrNotARootReference)
	}

	// Step 5: find the last "/" in bare name
	lastSlash := strings.LastIndex(bare, "/")
	parent := bare[:lastSlash]

	// Step 6: if parent is "ROOT", return "ROOT"
	if parent == "ROOT" {
		return "ROOT", nil
	}

	// Step 7: return parent
	return parent, nil
}

// LogicalNameGetQualifier extracts the parenthetical qualifier from a
// logical name. Returns ("", false) if no qualifier is present. Works
// with both ROOT/ and ARTIFACT/ references.
//
// Examples:
//   - "ROOT/x/y(z)"        → ("z", true)
//   - "ROOT/x/y"           → ("", false)
//   - "ARTIFACT/x/y(id)"   → ("id", true)
func LogicalNameGetQualifier(logicalName string) (qualifier string, present bool) {
	// Step 1: look for "(<qualifier>)" at end of string — closing ")" must be last char
	if !strings.HasSuffix(logicalName, ")") {
		return "", false
	}

	openParen := strings.LastIndex(logicalName, "(")
	if openParen < 0 {
		return "", false
	}

	// Step 2: extract content between parentheses
	qualifier = logicalName[openParen+1 : len(logicalName)-1]
	return qualifier, true
}

// LogicalNameStripQualifier returns the logical name without the
// parenthetical qualifier. If no qualifier is present, returns the
// input unchanged. Works with both ROOT/ and ARTIFACT/ references.
//
// Examples:
//   - "ROOT/x/y(z)"        → "ROOT/x/y"
//   - "ARTIFACT/x/y(id)"   → "ARTIFACT/x/y"
//   - "ROOT/x/y"           → "ROOT/x/y"
func LogicalNameStripQualifier(logicalName string) string {
	// Step 1: look for "(<qualifier>)" at end of string — closing ")" must be last char
	if !strings.HasSuffix(logicalName, ")") {
		return logicalName
	}

	openParen := strings.LastIndex(logicalName, "(")
	if openParen < 0 {
		return logicalName
	}

	// Step 2: return substring before the opening "("
	return logicalName[:openParen]
}

// LogicalNameHasParent returns true if the logical name is a ROOT/
// reference other than ROOT itself. Returns false for ROOT, ARTIFACT/
// references, and unrecognized prefixes.
func LogicalNameHasParent(logicalName string) bool {
	// Step 1: strip qualifier
	bare := LogicalNameStripQualifier(logicalName)

	// Step 2: if exactly "ROOT", return false
	if bare == "ROOT" {
		return false
	}

	// Step 3: if starts with "ROOT/", return true
	if strings.HasPrefix(bare, "ROOT/") {
		return true
	}

	// Step 4: all other cases, return false
	return false
}

// LogicalNameHasQualifier returns true if the logical name contains a
// parenthetical qualifier. Works with both ROOT/ and ARTIFACT/ references.
func LogicalNameHasQualifier(logicalName string) bool {
	_, present := LogicalNameGetQualifier(logicalName)
	return present
}

// LogicalNameIsArtifact returns true if the logical name starts with ARTIFACT/.
func LogicalNameIsArtifact(logicalName string) bool {
	return strings.HasPrefix(logicalName, "ARTIFACT/")
}

// LogicalNameGetArtifactGenerator returns the ROOT/ logical name of
// the node that generates the referenced artifact. Strips the
// ARTIFACT/ prefix and any qualifier.
//
// Examples:
//   - "ARTIFACT/x/y(id)"  → "ROOT/x/y"
//   - "ARTIFACT/x/y"      → "ROOT/x/y"
//
// Errors:
//   - ErrNotAnArtifactReference: the logical name does not start with ARTIFACT/.
func LogicalNameGetArtifactGenerator(logicalName string) (string, error) {
	// Step 1: must start with "ARTIFACT/"
	if !strings.HasPrefix(logicalName, "ARTIFACT/") {
		return "", fmt.Errorf("%w: logical name does not start with ARTIFACT/", ErrNotAnArtifactReference)
	}

	// Step 2: strip qualifier
	bare := LogicalNameStripQualifier(logicalName)

	// Step 3: strip "ARTIFACT/" prefix to get segment
	segment := strings.TrimPrefix(bare, "ARTIFACT/")

	// Step 4: return "ROOT/<segment>"
	return "ROOT/" + segment, nil
}
