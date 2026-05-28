// code-from-spec: ROOT/golang/implementation/internal/logical_names/code@Wmlf4neUll37H0HmMw8ppv1F-M4

package logicalnames

import (
	"fmt"
	"strings"
)

// LogicalNameToPath converts a ROOT/ logical name to the PathCfs of its _node.md file.
// Strips any qualifier before resolving.
func LogicalNameToPath(logicalName string) (string, error) {
	if logicalName != "ROOT" && !strings.HasPrefix(logicalName, "ROOT/") {
		return "", fmt.Errorf("unsupported reference")
	}

	// Strip any qualifier suffix by finding the last '('
	if idx := strings.LastIndex(logicalName, "("); idx != -1 {
		logicalName = logicalName[:idx]
	}

	if logicalName == "ROOT" {
		return "code-from-spec/_node.md", nil
	}

	// Remove the leading "ROOT/" prefix to get the relative segment
	relativeSegment := strings.TrimPrefix(logicalName, "ROOT/")

	return "code-from-spec/" + relativeSegment + "/_node.md", nil
}

// LogicalNameFromPath derives the ROOT/ logical name from a _node.md file path.
// The inverse of LogicalNameToPath.
func LogicalNameFromPath(cfsPath string) (string, error) {
	if !strings.HasPrefix(cfsPath, "code-from-spec/") {
		return "", fmt.Errorf("invalid path")
	}

	if cfsPath != "code-from-spec/_node.md" && !strings.HasSuffix(cfsPath, "/_node.md") {
		return "", fmt.Errorf("invalid path")
	}

	if cfsPath == "code-from-spec/_node.md" {
		return "ROOT", nil
	}

	// Remove the leading "code-from-spec/" prefix and trailing "/_node.md" suffix
	relativeSegment := strings.TrimPrefix(cfsPath, "code-from-spec/")
	relativeSegment = strings.TrimSuffix(relativeSegment, "/_node.md")

	return "ROOT/" + relativeSegment, nil
}

// LogicalNameGetParent returns the logical name of the parent node.
// Strips any qualifier before computing the parent.
func LogicalNameGetParent(logicalName string) (string, error) {
	if logicalName != "ROOT" && !strings.HasPrefix(logicalName, "ROOT/") {
		return "", fmt.Errorf("not a ROOT reference")
	}

	// Strip any qualifier suffix by finding the last '('
	if idx := strings.LastIndex(logicalName, "("); idx != -1 {
		logicalName = logicalName[:idx]
	}

	if logicalName == "ROOT" {
		return "", fmt.Errorf("no parent")
	}

	// Find the last '/' and take the substring before it
	idx := strings.LastIndex(logicalName, "/")
	return logicalName[:idx], nil
}

// LogicalNameGetQualifier extracts the parenthetical qualifier from a logical name.
// Works with both ROOT/ and ARTIFACT/ references.
// Returns the qualifier and true if found, or empty string and false if absent.
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

// LogicalNameHasParent returns true if the logical name is a ROOT/ reference
// other than ROOT itself.
func LogicalNameHasParent(logicalName string) bool {
	return strings.HasPrefix(logicalName, "ROOT/")
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
// that generates the referenced artifact. Strips the ARTIFACT/ prefix and any qualifier.
func LogicalNameGetArtifactGenerator(logicalName string) (string, error) {
	if !strings.HasPrefix(logicalName, "ARTIFACT/") {
		return "", fmt.Errorf("not an artifact reference")
	}

	// Remove the leading "ARTIFACT/" prefix to get the relative segment
	relativeSegment := strings.TrimPrefix(logicalName, "ARTIFACT/")

	// Strip any qualifier suffix by finding the last '('
	if idx := strings.LastIndex(relativeSegment, "("); idx != -1 {
		relativeSegment = relativeSegment[:idx]
	}

	return "ROOT/" + relativeSegment, nil
}
