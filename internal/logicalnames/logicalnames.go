// code-from-spec: ROOT/golang/implementation/utils/logical_names@z-V3tDrkWaua4riPv-wND8vh1jM
package logicalnames

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

var ErrUnsupportedReference = errors.New("unsupported reference: not a SPEC/ reference")
var ErrInvalidPath = errors.New("invalid path: not a _node.md file under code-from-spec/")
var ErrNoParent = errors.New("no parent: logical name is SPEC itself")
var ErrNotASpecReference = errors.New("not a SPEC/ reference")
var ErrNotAnArtifactReference = errors.New("not an ARTIFACT/ reference")
var ErrNotAnExternalReference = errors.New("not an EXTERNAL/ reference")

func LogicalNameToPath(logical_name string) (*pathutils.PathCfs, error) {
	stripped := LogicalNameStripQualifier(logical_name)

	if stripped == "SPEC" {
		return &pathutils.PathCfs{Value: "code-from-spec/_node.md"}, nil
	}

	if strings.HasPrefix(stripped, "SPEC/") {
		relative := strings.TrimPrefix(stripped, "SPEC/")
		return &pathutils.PathCfs{Value: "code-from-spec/" + relative + "/_node.md"}, nil
	}

	return nil, fmt.Errorf("%w", ErrUnsupportedReference)
}

func LogicalNameFromPath(cfs_path *pathutils.PathCfs) (string, error) {
	if cfs_path == nil {
		return "", fmt.Errorf("%w", ErrInvalidPath)
	}

	pathValue := cfs_path.Value

	if !strings.HasSuffix(pathValue, "_node.md") {
		return "", fmt.Errorf("%w", ErrInvalidPath)
	}

	if pathValue == "code-from-spec/_node.md" {
		return "SPEC", nil
	}

	if strings.HasPrefix(pathValue, "code-from-spec/") && strings.HasSuffix(pathValue, "/_node.md") {
		relative := strings.TrimPrefix(pathValue, "code-from-spec/")
		relative = strings.TrimSuffix(relative, "/_node.md")
		return "SPEC/" + relative, nil
	}

	return "", fmt.Errorf("%w", ErrInvalidPath)
}

func LogicalNameGetParent(logical_name string) (string, error) {
	stripped := LogicalNameStripQualifier(logical_name)

	if stripped != "SPEC" && !strings.HasPrefix(stripped, "SPEC/") {
		return "", fmt.Errorf("%w", ErrNotASpecReference)
	}

	if stripped == "SPEC" {
		return "", fmt.Errorf("%w", ErrNoParent)
	}

	relative := strings.TrimPrefix(stripped, "SPEC/")

	lastSlash := strings.LastIndex(relative, "/")
	if lastSlash == -1 {
		return "SPEC", nil
	}

	parentRelative := relative[:lastSlash]
	return "SPEC/" + parentRelative, nil
}

func LogicalNameGetQualifier(logical_name string) (string, bool) {
	lastOpen := strings.LastIndex(logical_name, "(")
	if lastOpen == -1 {
		return "", false
	}

	closeIdx := strings.Index(logical_name[lastOpen:], ")")
	if closeIdx == -1 {
		return "", false
	}

	closeIdx = lastOpen + closeIdx

	if closeIdx != len(logical_name)-1 {
		return "", false
	}

	qualifier := logical_name[lastOpen+1 : closeIdx]
	return qualifier, true
}

func LogicalNameStripQualifier(logical_name string) string {
	lastOpen := strings.LastIndex(logical_name, "(")
	if lastOpen == -1 {
		return logical_name
	}

	if !strings.HasSuffix(logical_name, ")") {
		return logical_name
	}

	return logical_name[:lastOpen]
}

func LogicalNameHasParent(logical_name string) bool {
	stripped := LogicalNameStripQualifier(logical_name)

	if strings.HasPrefix(stripped, "SPEC/") {
		relative := strings.TrimPrefix(stripped, "SPEC/")
		return relative != ""
	}

	return false
}

func LogicalNameHasQualifier(logical_name string) bool {
	_, ok := LogicalNameGetQualifier(logical_name)
	return ok
}

func LogicalNameIsArtifact(logical_name string) bool {
	return strings.HasPrefix(logical_name, "ARTIFACT/")
}

func LogicalNameIsSpec(logical_name string) bool {
	return logical_name == "SPEC" || strings.HasPrefix(logical_name, "SPEC/")
}

func LogicalNameIsExternal(logical_name string) bool {
	return strings.HasPrefix(logical_name, "EXTERNAL/")
}

func LogicalNameGetArtifactGenerator(logical_name string) (string, error) {
	if !strings.HasPrefix(logical_name, "ARTIFACT/") {
		return "", fmt.Errorf("%w", ErrNotAnArtifactReference)
	}

	relative := strings.TrimPrefix(logical_name, "ARTIFACT/")
	return "SPEC/" + relative, nil
}

func LogicalNameExternalToPath(logical_name string) (*pathutils.PathCfs, error) {
	if !strings.HasPrefix(logical_name, "EXTERNAL/") {
		return nil, fmt.Errorf("%w", ErrNotAnExternalReference)
	}

	relative := strings.TrimPrefix(logical_name, "EXTERNAL/")
	return &pathutils.PathCfs{Value: relative}, nil
}
