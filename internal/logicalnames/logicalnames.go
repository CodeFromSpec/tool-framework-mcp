// code-from-spec: ROOT/golang/internal/logical_names/code@PENDING
package logicalnames

import (
	"path/filepath"
	"strings"
)

const (
	rootPrefix     = "ROOT/"
	artifactPrefix = "ARTIFACT/"
	specDir        = "code-from-spec"
	nodeFile       = "_node.md"
)

// PathFromLogicalName resolves a ROOT/ logical name to a file path relative
// to the project root. Qualifiers are stripped before resolving. Returns
// ("", false) for ARTIFACT/ references or unrecognized input.
// Returned paths always use forward slashes.
func PathFromLogicalName(logicalName string) (string, bool) {
	if strings.HasPrefix(logicalName, artifactPrefix) {
		return "", false
	}

	base := stripQualifier(logicalName)

	if base == "ROOT" {
		return filepath.ToSlash(specDir + "/" + nodeFile), true
	}
	if !strings.HasPrefix(base, rootPrefix) {
		return "", false
	}

	relative := base[len(rootPrefix):]
	return filepath.ToSlash(specDir + "/" + relative + "/" + nodeFile), true
}

// LogicalNameFromPath derives the logical name from a file path relative to
// the project root. Only handles _node.md files under code-from-spec/.
// Returns ("", false) for paths that do not match.
func LogicalNameFromPath(filePath string) (string, bool) {
	// Normalize to forward slashes.
	p := filepath.ToSlash(filePath)

	prefix := specDir + "/"
	suffix := "/" + nodeFile

	if !strings.HasPrefix(p, prefix) {
		return "", false
	}

	inner := p[len(prefix):]

	// Exact root node: code-from-spec/_node.md
	if inner == nodeFile {
		return "ROOT", true
	}

	if !strings.HasSuffix(inner, suffix) {
		return "", false
	}

	// Strip the trailing /_node.md to get the relative segment.
	relative := inner[:len(inner)-len(suffix)]
	if relative == "" {
		return "", false
	}

	return "ROOT/" + relative, true
}

// HasParent reports whether the logical name has a parent node.
// Returns (hasParent, ok) where ok indicates valid input.
// ARTIFACT/ references return (false, false).
func HasParent(logicalName string) (bool, bool) {
	if strings.HasPrefix(logicalName, artifactPrefix) {
		return false, false
	}

	base := stripQualifier(logicalName)

	if base == "ROOT" {
		return false, true
	}
	if !strings.HasPrefix(base, rootPrefix) {
		return false, false
	}
	// Any ROOT/<path> has a parent.
	return true, true
}

// ParentLogicalName derives the parent's logical name. Qualifiers are
// stripped before deriving the parent. Returns ("", false) if there is
// no parent or the input is invalid.
func ParentLogicalName(logicalName string) (string, bool) {
	if strings.HasPrefix(logicalName, artifactPrefix) {
		return "", false
	}

	base := stripQualifier(logicalName)

	if base == "ROOT" {
		return "", false
	}
	if !strings.HasPrefix(base, rootPrefix) {
		return "", false
	}

	// Find last "/" to remove the last path segment.
	idx := strings.LastIndex(base, "/")
	if idx < 0 {
		return "", false
	}

	parent := base[:idx]
	return parent, true
}

// HasQualifier reports whether the logical name has a parenthetical qualifier.
// Returns (hasQualifier, ok) where ok indicates valid input.
func HasQualifier(logicalName string) (bool, bool) {
	if logicalName == "" {
		return false, false
	}
	// Accept both ROOT/ and ARTIFACT/ prefixed names.
	if logicalName != "ROOT" &&
		!strings.HasPrefix(logicalName, rootPrefix) &&
		!strings.HasPrefix(logicalName, artifactPrefix) {
		return false, false
	}

	if strings.Contains(logicalName, "(") && strings.HasSuffix(logicalName, ")") {
		return true, true
	}
	return false, true
}

// QualifierName extracts the qualifier text from a logical name.
// Returns ("", false) if there is no qualifier.
func QualifierName(logicalName string) (string, bool) {
	idx := strings.LastIndex(logicalName, "(")
	if idx < 0 {
		return "", false
	}
	if !strings.HasSuffix(logicalName, ")") {
		return "", false
	}
	qualifier := logicalName[idx+1 : len(logicalName)-1]
	return qualifier, true
}

// IsArtifactRef returns true if the logical name starts with "ARTIFACT/".
func IsArtifactRef(logicalName string) bool {
	return strings.HasPrefix(logicalName, artifactPrefix)
}

// ArtifactRefParts parses an ARTIFACT/ reference into the resolved node path
// and the artifact ID. The qualifier is required; returns ("", "", false) if
// the input is not an ARTIFACT/ reference or lacks a qualifier.
// The node path is returned as a forward-slash project-relative path.
func ArtifactRefParts(logicalName string) (string, string, bool) {
	if !strings.HasPrefix(logicalName, artifactPrefix) {
		return "", "", false
	}

	// Qualifier (artifact ID) is required.
	parenIdx := strings.LastIndex(logicalName, "(")
	if parenIdx < 0 || !strings.HasSuffix(logicalName, ")") {
		return "", "", false
	}

	artifactID := logicalName[parenIdx+1 : len(logicalName)-1]

	// The node portion is between "ARTIFACT/" and the last "(".
	nodeSegment := logicalName[len(artifactPrefix):parenIdx]

	// Build a ROOT/ logical name for the node, then resolve to a path.
	nodePath := filepath.ToSlash(specDir + "/" + nodeSegment + "/" + nodeFile)
	return nodePath, artifactID, true
}

// stripQualifier removes the parenthetical qualifier from a logical name,
// returning only the base name. If no qualifier is present, returns the
// input unchanged.
func stripQualifier(logicalName string) string {
	idx := strings.Index(logicalName, "(")
	if idx < 0 {
		return logicalName
	}
	return logicalName[:idx]
}
