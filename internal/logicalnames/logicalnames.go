// code-from-spec: ROOT/golang/implementation/utils/logical_names@taVqVCT3ahewU7h2YL1TE_colWg
package logicalnames

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

var ErrUnsupportedReference = errors.New("unsupported reference")
var ErrInvalidPath = errors.New("invalid path")
var ErrNoParent = errors.New("no parent")
var ErrNotARootReference = errors.New("not a ROOT/ reference")
var ErrNotAnArtifactReference = errors.New("not an ARTIFACT/ reference")

func LogicalNameToPath(logical_name string) (*pathutils.PathCfs, error) {
	stripped := LogicalNameStripQualifier(logical_name)

	if stripped != "ROOT" && !strings.HasPrefix(stripped, "ROOT/") {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedReference, logical_name)
	}

	if stripped == "ROOT" {
		return &pathutils.PathCfs{Value: "code-from-spec/_node.md"}, nil
	}

	relPath := strings.TrimPrefix(stripped, "ROOT/")
	return &pathutils.PathCfs{Value: "code-from-spec/" + relPath + "/_node.md"}, nil
}

func LogicalNameFromPath(cfs_path *pathutils.PathCfs) (string, error) {
	if cfs_path == nil {
		return "", fmt.Errorf("%w: nil path", ErrInvalidPath)
	}

	pathValue := cfs_path.Value

	if !strings.HasPrefix(pathValue, "code-from-spec/") {
		return "", fmt.Errorf("%w: %s", ErrInvalidPath, pathValue)
	}

	if pathValue == "code-from-spec/_node.md" {
		return "ROOT", nil
	}

	if !strings.HasSuffix(pathValue, "/_node.md") {
		return "", fmt.Errorf("%w: %s", ErrInvalidPath, pathValue)
	}

	relPath := strings.TrimPrefix(pathValue, "code-from-spec/")
	relPath = strings.TrimSuffix(relPath, "/_node.md")
	return "ROOT/" + relPath, nil
}

func LogicalNameGetParent(logical_name string) (string, error) {
	stripped := LogicalNameStripQualifier(logical_name)

	if stripped != "ROOT" && !strings.HasPrefix(stripped, "ROOT/") {
		return "", fmt.Errorf("%w: %s", ErrNotARootReference, logical_name)
	}

	if stripped == "ROOT" {
		return "", fmt.Errorf("%w: %s", ErrNoParent, logical_name)
	}

	relPath := strings.TrimPrefix(stripped, "ROOT/")

	lastSlash := strings.LastIndex(relPath, "/")
	if lastSlash == -1 {
		return "ROOT", nil
	}

	parentRel := relPath[:lastSlash]
	return "ROOT/" + parentRel, nil
}

func LogicalNameGetQualifier(logical_name string) (string, bool) {
	openIdx := strings.LastIndex(logical_name, "(")
	if openIdx == -1 {
		return "", false
	}

	closeIdx := strings.Index(logical_name[openIdx:], ")")
	if closeIdx == -1 {
		return "", false
	}

	qualifier := logical_name[openIdx+1 : openIdx+closeIdx]
	if qualifier == "" {
		return "", false
	}

	return qualifier, true
}

func LogicalNameStripQualifier(logical_name string) string {
	openIdx := strings.LastIndex(logical_name, "(")
	if openIdx == -1 {
		return logical_name
	}

	closeIdx := strings.Index(logical_name[openIdx:], ")")
	if closeIdx == -1 {
		return logical_name
	}

	return logical_name[:openIdx]
}

func LogicalNameHasParent(logical_name string) bool {
	stripped := LogicalNameStripQualifier(logical_name)
	return strings.HasPrefix(stripped, "ROOT/")
}

func LogicalNameHasQualifier(logical_name string) bool {
	_, ok := LogicalNameGetQualifier(logical_name)
	return ok
}

func LogicalNameIsArtifact(logical_name string) bool {
	return strings.HasPrefix(logical_name, "ARTIFACT/")
}

func LogicalNameGetArtifactGenerator(logical_name string) (string, error) {
	if !strings.HasPrefix(logical_name, "ARTIFACT/") {
		return "", fmt.Errorf("%w: %s", ErrNotAnArtifactReference, logical_name)
	}

	relPath := strings.TrimPrefix(logical_name, "ARTIFACT/")
	return "ROOT/" + relPath, nil
}
