// code-from-spec: ROOT/golang/implementation/utils/logical_names@wMlzHRjDWVKhad89PggIRLzgMso

package logicalnames

import (
	"errors"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

// ErrUnsupportedReference is returned when a logical name does not start with ROOT/.
var ErrUnsupportedReference = errors.New("unsupported reference: logical name must start with ROOT/")

// ErrInvalidPath is returned when a path is not a _node.md file under code-from-spec/.
var ErrInvalidPath = errors.New("invalid path: not a _node.md file under code-from-spec/")

// ErrNoParent is returned when the logical name is ROOT itself and has no parent.
var ErrNoParent = errors.New("no parent: ROOT has no parent node")

// ErrNotARootReference is returned when the logical name does not start with ROOT/.
var ErrNotARootReference = errors.New("not a ROOT/ reference")

// ErrNotAnArtifactReference is returned when the logical name does not start with ARTIFACT/.
var ErrNotAnArtifactReference = errors.New("not an ARTIFACT/ reference")

// LogicalNameToPath converts a ROOT/ logical name to the PathCfs of the
// corresponding _node.md file. Strips any qualifier before resolving.
// Only accepts ROOT/ references.
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

// LogicalNameFromPath derives the ROOT/ logical name from a _node.md file path.
// The inverse of LogicalNameToPath. Always returns a ROOT/ reference.
func LogicalNameFromPath(cfs_path *pathutils.PathCfs) (string, error) {
	path := cfs_path.Value

	if !strings.HasPrefix(path, "code-from-spec/") {
		return "", ErrInvalidPath
	}

	if path == "code-from-spec/_node.md" {
		return "ROOT", nil
	}

	if !strings.HasSuffix(path, "/_node.md") {
		return "", ErrInvalidPath
	}

	middle := strings.TrimPrefix(path, "code-from-spec/")
	middle = strings.TrimSuffix(middle, "/_node.md")

	return "ROOT/" + middle, nil
}

// LogicalNameGetParent returns the logical name of the parent node.
// Strips any qualifier before computing the parent. Only accepts ROOT/ references.
func LogicalNameGetParent(logical_name string) (string, error) {
	if !strings.HasPrefix(logical_name, "ROOT/") {
		return "", ErrNotARootReference
	}

	stripped := LogicalNameStripQualifier(logical_name)

	if stripped == "ROOT" {
		return "", ErrNoParent
	}

	segment := strings.TrimPrefix(stripped, "ROOT/")

	lastSlash := strings.LastIndex(segment, "/")
	if lastSlash == -1 {
		return "ROOT", nil
	}

	parentSegment := segment[:lastSlash]
	return "ROOT/" + parentSegment, nil
}

// LogicalNameGetQualifier extracts the parenthetical qualifier from a logical name.
// Returns an empty string and false if no qualifier is present.
func LogicalNameGetQualifier(logical_name string) (qualifier string, ok bool) {
	openIdx := strings.Index(logical_name, "(")
	if openIdx == -1 {
		return "", false
	}

	start := openIdx + 1
	closeIdx := strings.Index(logical_name[start:], ")")
	if closeIdx == -1 {
		return "", false
	}

	extracted := logical_name[start : start+closeIdx]
	if extracted == "" {
		return "", false
	}

	return extracted, true
}

// LogicalNameStripQualifier returns the logical name without the parenthetical
// qualifier. If no qualifier is present, returns the input unchanged.
func LogicalNameStripQualifier(logical_name string) string {
	openIdx := strings.Index(logical_name, "(")
	if openIdx == -1 {
		return logical_name
	}

	start := openIdx + 1
	closeIdx := strings.Index(logical_name[start:], ")")
	if closeIdx == -1 {
		return logical_name
	}

	return logical_name[:openIdx] + logical_name[start+closeIdx+1:]
}

// LogicalNameHasParent returns true if the logical name is a ROOT/ reference
// other than ROOT itself.
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
// parenthetical qualifier.
func LogicalNameHasQualifier(logical_name string) bool {
	_, ok := LogicalNameGetQualifier(logical_name)
	return ok
}

// LogicalNameIsArtifact returns true if the logical name starts with ARTIFACT/.
func LogicalNameIsArtifact(logical_name string) bool {
	return strings.HasPrefix(logical_name, "ARTIFACT/")
}

// LogicalNameGetArtifactGenerator returns the ROOT/ logical name of the node
// that generates the referenced artifact.
func LogicalNameGetArtifactGenerator(logical_name string) (string, error) {
	if !strings.HasPrefix(logical_name, "ARTIFACT/") {
		return "", ErrNotAnArtifactReference
	}

	stripped := LogicalNameStripQualifier(logical_name)

	pathSegment := strings.TrimPrefix(stripped, "ARTIFACT/")
	return "ROOT/" + pathSegment, nil
}
