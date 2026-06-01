// code-from-spec: ROOT/golang/implementation/utils/logical_names@rj6CzH2ieye_Q1zJFBbrIaYJncE
package logicalnames

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

var ErrUnsupportedReference = errors.New("unsupported reference: expected a ROOT/ reference")
var ErrInvalidPath = errors.New("invalid path: not a _node.md file under code-from-spec/")
var ErrNoParent = errors.New("no parent: ROOT has no parent node")
var ErrNotARootReference = errors.New("not a ROOT/ reference")
var ErrNotAnArtifactReference = errors.New("not an ARTIFACT/ reference")

func LogicalNameToPath(logicalName string) (*pathutils.PathCfs, error) {
	if logicalName != "ROOT" && !strings.HasPrefix(logicalName, "ROOT/") {
		return nil, fmt.Errorf("%w", ErrUnsupportedReference)
	}

	bare := LogicalNameStripQualifier(logicalName)

	if bare == "ROOT" {
		return &pathutils.PathCfs{Value: "code-from-spec/_node.md"}, nil
	}

	if !strings.HasPrefix(bare, "ROOT/") {
		return nil, fmt.Errorf("%w", ErrUnsupportedReference)
	}

	segment := strings.TrimPrefix(bare, "ROOT/")
	return &pathutils.PathCfs{Value: "code-from-spec/" + segment + "/_node.md"}, nil
}

func LogicalNameFromPath(cfsPath *pathutils.PathCfs) (string, error) {
	value := cfsPath.Value

	if !strings.HasPrefix(value, "code-from-spec/") {
		return "", fmt.Errorf("%w", ErrInvalidPath)
	}

	if value != "code-from-spec/_node.md" && !strings.HasSuffix(value, "/_node.md") {
		return "", fmt.Errorf("%w", ErrInvalidPath)
	}

	if value == "code-from-spec/_node.md" {
		return "ROOT", nil
	}

	middle := strings.TrimPrefix(value, "code-from-spec/")
	middle = strings.TrimSuffix(middle, "/_node.md")
	return "ROOT/" + middle, nil
}

func LogicalNameGetParent(logicalName string) (string, error) {
	if logicalName != "ROOT" && !strings.HasPrefix(logicalName, "ROOT/") {
		return "", fmt.Errorf("%w", ErrNotARootReference)
	}

	bare := LogicalNameStripQualifier(logicalName)

	if bare == "ROOT" {
		return "", fmt.Errorf("%w", ErrNoParent)
	}

	if !strings.Contains(bare[4:], "/") {
		return "", fmt.Errorf("%w", ErrNotARootReference)
	}

	lastSlash := strings.LastIndex(bare, "/")
	parent := bare[:lastSlash]

	if parent == "ROOT" {
		return "ROOT", nil
	}

	return parent, nil
}

func LogicalNameGetQualifier(logicalName string) (string, bool) {
	if !strings.HasSuffix(logicalName, ")") {
		return "", false
	}

	openParen := strings.LastIndex(logicalName, "(")
	if openParen == -1 {
		return "", false
	}

	qualifier := logicalName[openParen+1 : len(logicalName)-1]
	return qualifier, true
}

func LogicalNameStripQualifier(logicalName string) string {
	if !strings.HasSuffix(logicalName, ")") {
		return logicalName
	}

	openParen := strings.LastIndex(logicalName, "(")
	if openParen == -1 {
		return logicalName
	}

	return logicalName[:openParen]
}

func LogicalNameHasParent(logicalName string) bool {
	bare := LogicalNameStripQualifier(logicalName)

	if bare == "ROOT" {
		return false
	}

	if strings.HasPrefix(bare, "ROOT/") {
		return true
	}

	return false
}

func LogicalNameHasQualifier(logicalName string) bool {
	_, present := LogicalNameGetQualifier(logicalName)
	return present
}

func LogicalNameIsArtifact(logicalName string) bool {
	return strings.HasPrefix(logicalName, "ARTIFACT/")
}

func LogicalNameGetArtifactGenerator(logicalName string) (string, error) {
	if !strings.HasPrefix(logicalName, "ARTIFACT/") {
		return "", fmt.Errorf("%w", ErrNotAnArtifactReference)
	}

	bare := LogicalNameStripQualifier(logicalName)
	segment := strings.TrimPrefix(bare, "ARTIFACT/")
	return "ROOT/" + segment, nil
}
