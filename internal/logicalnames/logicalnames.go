// code-from-spec: SPEC/golang/implementation/utils/logical_names@p4j1KR50rGghz-QEYEy8eIalKqw
package logicalnames

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

var ErrUnsupportedReference = errors.New("logical name is not a SPEC/ reference")
var ErrInvalidPath = errors.New("path is not a _node.md file under code-from-spec/")
var ErrNoParent = errors.New("logical name is SPEC itself")
var ErrNotASpecReference = errors.New("logical name is not a SPEC/ reference")
var ErrNotAnArtifactReference = errors.New("logical name does not start with ARTIFACT/")
var ErrNotAnExternalReference = errors.New("logical name does not start with EXTERNAL/")

func LogicalNameToPath(logicalName string) (pathutils.PathCfs, error) {
	stripped := LogicalNameStripQualifier(logicalName)

	if stripped != "SPEC" && !strings.HasPrefix(stripped, "SPEC/") {
		return pathutils.PathCfs{}, fmt.Errorf("%w: %s", ErrUnsupportedReference, logicalName)
	}

	if stripped == "SPEC" {
		return pathutils.PathCfs{Value: "code-from-spec/_node.md"}, nil
	}

	relativePath := strings.TrimPrefix(stripped, "SPEC/")
	return pathutils.PathCfs{Value: "code-from-spec/" + relativePath + "/_node.md"}, nil
}

func LogicalNameFromPath(cfsPath pathutils.PathCfs) (string, error) {
	pathValue := cfsPath.Value

	if pathValue != "code-from-spec/_node.md" && !strings.HasSuffix(pathValue, "/_node.md") {
		return "", fmt.Errorf("%w: %s", ErrInvalidPath, pathValue)
	}

	if !strings.HasPrefix(pathValue, "code-from-spec/") {
		return "", fmt.Errorf("%w: %s", ErrInvalidPath, pathValue)
	}

	if pathValue == "code-from-spec/_node.md" {
		return "SPEC", nil
	}

	relativePath := strings.TrimPrefix(pathValue, "code-from-spec/")
	relativePath = strings.TrimSuffix(relativePath, "/_node.md")
	return "SPEC/" + relativePath, nil
}

func LogicalNameGetParent(logicalName string) (string, error) {
	stripped := LogicalNameStripQualifier(logicalName)

	if stripped != "SPEC" && !strings.HasPrefix(stripped, "SPEC/") {
		return "", fmt.Errorf("%w: %s", ErrNotASpecReference, logicalName)
	}

	if stripped == "SPEC" {
		return "", fmt.Errorf("%w: %s", ErrNoParent, logicalName)
	}

	relativePath := strings.TrimPrefix(stripped, "SPEC/")

	lastSlash := strings.LastIndex(relativePath, "/")
	if lastSlash == -1 {
		return "SPEC", nil
	}

	parentRelative := relativePath[:lastSlash]
	return "SPEC/" + parentRelative, nil
}

func LogicalNameGetQualifier(logicalName string) (string, bool) {
	openIdx := strings.Index(logicalName, "(")
	if openIdx == -1 {
		return "", false
	}

	closeIdx := strings.Index(logicalName[openIdx:], ")")
	if closeIdx == -1 {
		return "", false
	}

	qualifier := logicalName[openIdx+1 : openIdx+closeIdx]
	return qualifier, true
}

func LogicalNameStripQualifier(logicalName string) string {
	openIdx := strings.Index(logicalName, "(")
	if openIdx == -1 {
		return logicalName
	}
	return logicalName[:openIdx]
}

func LogicalNameHasParent(logicalName string) bool {
	stripped := LogicalNameStripQualifier(logicalName)
	return strings.HasPrefix(stripped, "SPEC/")
}

func LogicalNameHasQualifier(logicalName string) bool {
	_, ok := LogicalNameGetQualifier(logicalName)
	return ok
}

func LogicalNameIsArtifact(logicalName string) bool {
	return strings.HasPrefix(logicalName, "ARTIFACT/")
}

func LogicalNameIsSpec(logicalName string) bool {
	if logicalName == "SPEC" {
		return true
	}
	return strings.HasPrefix(logicalName, "SPEC/")
}

func LogicalNameIsExternal(logicalName string) bool {
	return strings.HasPrefix(logicalName, "EXTERNAL/")
}

func LogicalNameGetArtifactGenerator(logicalName string) (string, error) {
	if !strings.HasPrefix(logicalName, "ARTIFACT/") {
		return "", fmt.Errorf("%w: %s", ErrNotAnArtifactReference, logicalName)
	}

	relativePath := strings.TrimPrefix(logicalName, "ARTIFACT/")
	return "SPEC/" + relativePath, nil
}

func LogicalNameExternalToPath(logicalName string) (pathutils.PathCfs, error) {
	if !strings.HasPrefix(logicalName, "EXTERNAL/") {
		return pathutils.PathCfs{}, fmt.Errorf("%w: %s", ErrNotAnExternalReference, logicalName)
	}

	relativePath := strings.TrimPrefix(logicalName, "EXTERNAL/")
	return pathutils.PathCfs{Value: relativePath}, nil
}
