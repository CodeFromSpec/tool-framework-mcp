// code-from-spec: ROOT/golang/implementation/utils/logical_names@MCaIYbF9rkAUqlt3-Oojno5cVUg

package logicalnames

import (
	"errors"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

var (
	// ErrUnsupportedReference is returned when a logical name does not
	// start with ROOT/.
	ErrUnsupportedReference = errors.New("unsupported reference")

	// ErrInvalidPath is returned when the path is not a _node.md file
	// under code-from-spec/.
	ErrInvalidPath = errors.New("invalid path")

	// ErrNoParent is returned when the logical name is ROOT itself.
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
	if logical_name != "ROOT" && !strings.HasPrefix(logical_name, "ROOT/") {
		return nil, ErrUnsupportedReference
	}

	stripped := LogicalNameStripQualifier(logical_name)

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

	if !strings.HasPrefix(path, "code-from-spec/") || !strings.HasSuffix(path, "/_node.md") {
		return "", ErrInvalidPath
	}

	if path == "code-from-spec/_node.md" {
		return "ROOT", nil
	}

	middle := strings.TrimPrefix(path, "code-from-spec/")
	middle = strings.TrimSuffix(middle, "/_node.md")

	if middle == "" {
		return "", ErrInvalidPath
	}

	return "ROOT/" + middle, nil
}

// LogicalNameGetParent returns the logical name of the parent node.
// Strips any qualifier before computing the parent. Only accepts
// ROOT/ references.
func LogicalNameGetParent(logical_name string) (string, error) {
	if logical_name != "ROOT" && !strings.HasPrefix(logical_name, "ROOT/") {
		return "", ErrNotARootReference
	}

	stripped := LogicalNameStripQualifier(logical_name)

	if stripped == "ROOT" {
		return "", ErrNoParent
	}

	lastSlash := strings.LastIndex(stripped, "/")
	parent := stripped[:lastSlash]

	if parent == "" {
		return "", ErrNoParent
	}

	return parent, nil
}

// LogicalNameGetQualifier extracts the parenthetical qualifier from a
// logical name. Returns an empty string and false if no qualifier is
// present. Works with both ROOT/ and ARTIFACT/ references.
func LogicalNameGetQualifier(logical_name string) (string, bool) {
	openPos := strings.Index(logical_name, "(")
	if openPos == -1 {
		return "", false
	}

	closePos := strings.Index(logical_name[openPos:], ")")
	if closePos == -1 {
		return "", false
	}
	closePos = openPos + closePos

	qualifier := logical_name[openPos+1 : closePos]
	if qualifier == "" {
		return "", false
	}

	return qualifier, true
}

// LogicalNameStripQualifier returns the logical name without the
// parenthetical qualifier. If no qualifier is present, returns the
// input unchanged. Works with both ROOT/ and ARTIFACT/ references.
func LogicalNameStripQualifier(logical_name string) string {
	openPos := strings.Index(logical_name, "(")
	if openPos == -1 {
		return logical_name
	}

	return logical_name[:openPos]
}

// LogicalNameHasParent returns true if the logical name is a ROOT/
// reference other than ROOT itself. Returns false for ROOT, ARTIFACT/
// references, and unrecognized prefixes.
func LogicalNameHasParent(logical_name string) bool {
	if !strings.HasPrefix(logical_name, "ROOT/") {
		return false
	}

	stripped := LogicalNameStripQualifier(logical_name)

	if stripped == "ROOT" {
		return false
	}

	return true
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
		return "", ErrNotAnArtifactReference
	}

	stripped := LogicalNameStripQualifier(logical_name)

	relative := strings.TrimPrefix(stripped, "ARTIFACT/")

	return "ROOT/" + relative, nil
}
