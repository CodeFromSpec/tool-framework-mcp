// code-from-spec: ROOT/golang/implementation/utils/logical_names@3a57KXzMve_w7FjXp_qth_CR_XE

package logicalnames

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
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

// LogicalNameToPath converts a ROOT/ logical name to the PathCfs of
// the corresponding _node.md file. Strips any qualifier before
// resolving. Only accepts ROOT/ references.
func LogicalNameToPath(logical_name string) (*pathutils.PathCfs, error) {
	stripped := LogicalNameStripQualifier(logical_name)

	if stripped != "ROOT" && !strings.HasPrefix(stripped, "ROOT/") {
		return nil, fmt.Errorf("%w", ErrUnsupportedReference)
	}

	if stripped == "ROOT" {
		return &pathutils.PathCfs{Value: "code-from-spec/_node.md"}, nil
	}

	relative := strings.TrimPrefix(stripped, "ROOT/")

	return &pathutils.PathCfs{Value: "code-from-spec/" + relative + "/_node.md"}, nil
}

// LogicalNameFromPath derives the ROOT/ logical name from a _node.md
// file path. The inverse of LogicalNameToPath. Always returns a ROOT/
// reference.
func LogicalNameFromPath(cfs_path *pathutils.PathCfs) (string, error) {
	path := cfs_path.Value

	if !strings.HasPrefix(path, "code-from-spec/") {
		return "", fmt.Errorf("%w", ErrInvalidPath)
	}

	if !strings.HasSuffix(path, "_node.md") {
		return "", fmt.Errorf("%w", ErrInvalidPath)
	}

	if path == "code-from-spec/_node.md" {
		return "ROOT", nil
	}

	middle := strings.TrimPrefix(path, "code-from-spec/")
	relative := strings.TrimSuffix(middle, "/_node.md")

	if relative == "" {
		return "", fmt.Errorf("%w", ErrInvalidPath)
	}

	return "ROOT/" + relative, nil
}

// LogicalNameGetParent returns the logical name of the parent node.
// Strips any qualifier before computing the parent. Only accepts
// ROOT/ references.
func LogicalNameGetParent(logical_name string) (string, error) {
	stripped := LogicalNameStripQualifier(logical_name)

	if stripped != "ROOT" && !strings.HasPrefix(stripped, "ROOT/") {
		return "", fmt.Errorf("%w", ErrNotARootReference)
	}

	if stripped == "ROOT" {
		return "", fmt.Errorf("%w", ErrNoParent)
	}

	relative := strings.TrimPrefix(stripped, "ROOT/")

	lastSlash := strings.LastIndex(relative, "/")
	if lastSlash == -1 {
		return "ROOT", nil
	}

	parentRelative := relative[:lastSlash]
	return "ROOT/" + parentRelative, nil
}

// LogicalNameGetQualifier extracts the parenthetical qualifier from a
// logical name. Returns the empty string and false if no qualifier is
// present. Works with both ROOT/ and ARTIFACT/ references.
func LogicalNameGetQualifier(logical_name string) (qualifier string, ok bool) {
	openPos := strings.Index(logical_name, "(")
	if openPos == -1 {
		return "", false
	}

	closePos := strings.LastIndex(logical_name, ")")
	if closePos == -1 || closePos < openPos {
		return "", false
	}

	if closePos != len(logical_name)-1 {
		return "", false
	}

	q := logical_name[openPos+1 : closePos]
	if q == "" {
		return "", false
	}

	return q, true
}

// LogicalNameStripQualifier returns the logical name without the
// parenthetical qualifier. If no qualifier is present, returns the
// input unchanged. Works with both ROOT/ and ARTIFACT/ references.
func LogicalNameStripQualifier(logical_name string) string {
	openPos := strings.Index(logical_name, "(")
	if openPos == -1 {
		return logical_name
	}

	closePos := strings.LastIndex(logical_name, ")")
	if closePos == -1 || closePos < openPos {
		return logical_name
	}

	if closePos != len(logical_name)-1 {
		return logical_name
	}

	return logical_name[:openPos]
}

// LogicalNameHasParent returns true if the logical name is a ROOT/
// reference other than ROOT itself. Returns false for ROOT,
// ARTIFACT/ references, and unrecognized prefixes.
func LogicalNameHasParent(logical_name string) bool {
	stripped := LogicalNameStripQualifier(logical_name)

	if stripped == "ROOT" {
		return false
	}

	if strings.HasPrefix(stripped, "ROOT/") {
		return true
	}

	return false
}

// LogicalNameHasQualifier returns true if the logical name contains a
// parenthetical qualifier. Works with both ROOT/ and ARTIFACT/
// references.
func LogicalNameHasQualifier(logical_name string) bool {
	_, ok := LogicalNameGetQualifier(logical_name)
	return ok
}

// LogicalNameIsArtifact returns true if the logical name starts with
// ARTIFACT/.
func LogicalNameIsArtifact(logical_name string) bool {
	return strings.HasPrefix(logical_name, "ARTIFACT/")
}

// LogicalNameGetArtifactGenerator returns the ROOT/ logical name of
// the node that generates the referenced artifact. Strips the
// ARTIFACT/ prefix and any qualifier.
func LogicalNameGetArtifactGenerator(logical_name string) (string, error) {
	if !strings.HasPrefix(logical_name, "ARTIFACT/") {
		return "", fmt.Errorf("%w", ErrNotAnArtifactReference)
	}

	stripped := LogicalNameStripQualifier(logical_name)
	relative := strings.TrimPrefix(stripped, "ARTIFACT/")

	return "ROOT/" + relative, nil
}
