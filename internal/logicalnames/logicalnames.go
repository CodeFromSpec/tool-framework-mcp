// code-from-spec: ROOT/golang/internal/logical_names/code@9_tkZQOVQrZr6f_N4t0gF8eTG_A

// Package logicalnames provides pure functions for working with Code From Spec
// logical names — the identifiers used to reference nodes and artifacts in the
// spec tree.
//
// Logical names take two forms:
//
//   - ROOT/x/y        — references a spec node (maps to code-from-spec/x/y/_node.md)
//   - ARTIFACT/x/y(id) — references a generated artifact by node path and artifact id
//
// Both forms may carry a parenthetical qualifier: ROOT/x/y(z) or ARTIFACT/x/y(id).
// For ROOT/ names the qualifier identifies a public subsection; for ARTIFACT/ names
// it is required and identifies the artifact id.
//
// All functions are pure — no I/O, no side effects.
// All returned file paths use forward slashes regardless of the host OS.
package logicalnames

import (
	"path/filepath"
	"strings"
)

// ─── Constants ────────────────────────────────────────────────────────────────

const (
	// rootPrefix is the prefix used for spec node logical names.
	rootPrefix = "ROOT"

	// artifactPrefix is the prefix used for artifact reference logical names.
	artifactPrefix = "ARTIFACT/"

	// specDir is the directory under the project root where spec nodes live.
	specDir = "code-from-spec"

	// nodeFile is the filename of every spec node.
	nodeFile = "_node.md"
)

// ─── PathFromLogicalName ──────────────────────────────────────────────────────

// PathFromLogicalName resolves a ROOT/ logical name to the corresponding
// _node.md file path, relative to the project root. The path always uses
// forward slashes as separators.
//
// If the logical name carries a parenthetical qualifier it is stripped before
// resolution, so "ROOT/x/y(z)" resolves to the same path as "ROOT/x/y".
//
// Only ROOT/ references are handled. ARTIFACT/ references require frontmatter
// lookup to locate the actual file, so this function returns ("", false) for
// them — use IsArtifactRef and ArtifactRefParts instead.
//
// Returns ("", false) for any input that is not a valid ROOT/ reference.
//
// Examples:
//
//	PathFromLogicalName("ROOT")        → ("code-from-spec/_node.md", true)
//	PathFromLogicalName("ROOT/x/y")    → ("code-from-spec/x/y/_node.md", true)
//	PathFromLogicalName("ROOT/x/y(z)") → ("code-from-spec/x/y/_node.md", true)
//	PathFromLogicalName("ARTIFACT/x")  → ("", false)
//	PathFromLogicalName("")            → ("", false)
func PathFromLogicalName(logicalName string) (string, bool) {
	// ARTIFACT/ references cannot be statically resolved to a path.
	if strings.HasPrefix(logicalName, artifactPrefix) {
		return "", false
	}

	// Must start with "ROOT".
	if !strings.HasPrefix(logicalName, rootPrefix) {
		return "", false
	}

	// Strip any qualifier before resolving.
	bare := stripQualifier(logicalName)

	// "ROOT" itself → the root node file.
	if bare == rootPrefix {
		return specDir + "/" + nodeFile, true
	}

	// "ROOT/" followed by a path segment.
	prefix := rootPrefix + "/"
	if !strings.HasPrefix(bare, prefix) {
		// Something like "ROOTfoo" — not a valid name.
		return "", false
	}

	segmentPath := bare[len(prefix):]
	if segmentPath == "" {
		return "", false
	}

	// Build the path using forward slashes.
	// filepath.ToSlash ensures we are safe on Windows even if any OS-specific
	// join were used; here we build manually with "/".
	result := specDir + "/" + segmentPath + "/" + nodeFile
	return filepath.ToSlash(result), true
}

// ─── HasParent ────────────────────────────────────────────────────────────────

// HasParent reports whether the logical name has a parent node.
//
// Returns (hasParent, ok) where ok indicates whether the input is a valid ROOT/
// logical name. ARTIFACT/ references and empty strings return (false, false).
//
// Examples:
//
//	HasParent("ROOT")     → (false, true)   ROOT has no parent
//	HasParent("ROOT/x")   → (true,  true)
//	HasParent("ROOT/x(y)")→ (true,  true)
//	HasParent("")         → (false, false)
func HasParent(logicalName string) (hasParent, ok bool) {
	if logicalName == "" {
		return false, false
	}

	// ARTIFACT/ names are not valid for this function.
	if strings.HasPrefix(logicalName, artifactPrefix) {
		return false, false
	}

	// Must be "ROOT" or "ROOT/<something>".
	if logicalName == rootPrefix {
		return false, true
	}

	prefix := rootPrefix + "/"
	if strings.HasPrefix(logicalName, prefix) {
		return true, true
	}

	// Anything else is not a recognized logical name.
	return false, false
}

// ─── ParentLogicalName ────────────────────────────────────────────────────────

// ParentLogicalName derives the logical name of the parent node.
//
// The qualifier is stripped before computing the parent, so
// "ROOT/x/y(z)" yields parent "ROOT/x" (not "ROOT/x/y(z)").
//
// Returns ("", false) when:
//   - the input is "ROOT" (no parent)
//   - the input is not a valid ROOT/ name
//
// Examples:
//
//	ParentLogicalName("ROOT/x")     → ("ROOT",  true)
//	ParentLogicalName("ROOT/x/y")   → ("ROOT/x", true)
//	ParentLogicalName("ROOT/x/y(z)")→ ("ROOT/x", true)
//	ParentLogicalName("ROOT")       → ("", false)
func ParentLogicalName(logicalName string) (string, bool) {
	// ARTIFACT/ is not handled.
	if strings.HasPrefix(logicalName, artifactPrefix) {
		return "", false
	}

	// Must start with "ROOT".
	if !strings.HasPrefix(logicalName, rootPrefix) {
		return "", false
	}

	// Strip qualifier first.
	bare := stripQualifier(logicalName)

	// ROOT has no parent.
	if bare == rootPrefix {
		return "", false
	}

	// Find the last "/" and cut everything from it onward.
	idx := strings.LastIndex(bare, "/")
	if idx < 0 {
		// Should not happen given prefix check above, but be safe.
		return "", false
	}

	parent := bare[:idx]
	if parent == "" {
		return "", false
	}

	return parent, true
}

// ─── HasQualifier ─────────────────────────────────────────────────────────────

// HasQualifier reports whether the logical name carries a parenthetical
// qualifier.
//
// Returns (hasQualifier, ok) where ok indicates whether the input is a
// recognized logical name (ROOT/ or ARTIFACT/). Empty strings return
// (false, false).
//
// Examples:
//
//	HasQualifier("ROOT")          → (false, true)
//	HasQualifier("ROOT/x")        → (false, true)
//	HasQualifier("ROOT/x(y)")     → (true,  true)
//	HasQualifier("ARTIFACT/x(y)") → (true,  true)
//	HasQualifier("")              → (false, false)
func HasQualifier(logicalName string) (hasQualifier, ok bool) {
	if logicalName == "" {
		return false, false
	}

	// Both ROOT/ and ARTIFACT/ are valid for this function.
	isRoot := logicalName == rootPrefix || strings.HasPrefix(logicalName, rootPrefix+"/")
	isArtifact := strings.HasPrefix(logicalName, artifactPrefix)

	if !isRoot && !isArtifact {
		return false, false
	}

	qualifier, found := extractQualifier(logicalName)
	if !found || qualifier == "" {
		return false, true
	}

	return true, true
}

// ─── QualifierName ────────────────────────────────────────────────────────────

// QualifierName extracts the parenthetical qualifier from a logical name.
//
// Returns ("", false) when no qualifier is present or the qualifier is empty.
// Works for both ROOT/ and ARTIFACT/ names.
//
// Examples:
//
//	QualifierName("ROOT/x(y)")     → ("y", true)
//	QualifierName("ROOT/x/y(z)")   → ("z", true)
//	QualifierName("ARTIFACT/x(y)") → ("y", true)
//	QualifierName("ROOT/x")        → ("", false)
//	QualifierName("ROOT")          → ("", false)
func QualifierName(logicalName string) (string, bool) {
	return extractQualifier(logicalName)
}

// ─── IsArtifactRef ────────────────────────────────────────────────────────────

// IsArtifactRef reports whether the logical name is an ARTIFACT/ reference.
//
// Example:
//
//	IsArtifactRef("ARTIFACT/x/y(id)") → true
//	IsArtifactRef("ROOT/x")           → false
func IsArtifactRef(logicalName string) bool {
	return strings.HasPrefix(logicalName, artifactPrefix)
}

// ─── ArtifactRefParts ─────────────────────────────────────────────────────────

// ArtifactRefParts parses an ARTIFACT/ logical name into its node path and
// artifact id.
//
// The qualifier (artifact id) is mandatory for ARTIFACT/ references. Returns
// ("", "", false) when:
//   - the input is not an ARTIFACT/ reference
//   - the input has no qualifier
//
// Returned node paths use forward slashes.
//
// Examples:
//
//	ArtifactRefParts("ARTIFACT/x(y)")    → ("code-from-spec/x/_node.md",   "y",  true)
//	ArtifactRefParts("ARTIFACT/x/y(z)")  → ("code-from-spec/x/y/_node.md", "z",  true)
//	ArtifactRefParts("ARTIFACT/x")       → ("", "", false)   — no qualifier
//	ArtifactRefParts("ROOT/x(y)")        → ("", "", false)   — not ARTIFACT/
func ArtifactRefParts(logicalName string) (nodePath string, artifactID string, ok bool) {
	if !strings.HasPrefix(logicalName, artifactPrefix) {
		return "", "", false
	}

	// Qualifier is required for ARTIFACT/ references.
	qualifier, found := extractQualifier(logicalName)
	if !found || qualifier == "" {
		return "", "", false
	}

	// Strip the qualifier to get the bare reference.
	bare := stripQualifier(logicalName)

	// Take the portion after "ARTIFACT/" and build the node path.
	segmentPath := bare[len(artifactPrefix):]
	if segmentPath == "" {
		return "", "", false
	}

	path := filepath.ToSlash(specDir + "/" + segmentPath + "/" + nodeFile)
	return path, qualifier, true
}

// ─── LogicalNameFromPath ──────────────────────────────────────────────────────

// LogicalNameFromPath derives the logical name from a _node.md file path
// relative to the project root. This is the inverse of PathFromLogicalName.
//
// Only handles _node.md files under code-from-spec/. Returns ("", false) for
// any path that does not match.
//
// Examples:
//
//	LogicalNameFromPath("code-from-spec/_node.md")       → ("ROOT",    true)
//	LogicalNameFromPath("code-from-spec/x/_node.md")     → ("ROOT/x",  true)
//	LogicalNameFromPath("code-from-spec/x/y/_node.md")   → ("ROOT/x/y", true)
//	LogicalNameFromPath("internal/foo.go")               → ("", false)
func LogicalNameFromPath(filePath string) (string, bool) {
	// Normalize to forward slashes to handle Windows paths transparently.
	normalized := filepath.ToSlash(filePath)

	specPrefix := specDir + "/"
	nodeSuffix := "/" + nodeFile

	// Must reside under code-from-spec/.
	if !strings.HasPrefix(normalized, specPrefix) {
		return "", false
	}

	// Special case: the root node file itself.
	if normalized == specDir+"/"+nodeFile {
		return rootPrefix, true
	}

	// Must end with /_node.md.
	if !strings.HasSuffix(normalized, nodeSuffix) {
		return "", false
	}

	// Strip the leading "code-from-spec/" and trailing "/_node.md".
	inner := normalized[len(specPrefix) : len(normalized)-len(nodeSuffix)]
	if inner == "" {
		return "", false
	}

	return rootPrefix + "/" + inner, true
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

// extractQualifier extracts the parenthetical qualifier from a logical name.
//
// Returns (qualifier, true) when a valid non-empty qualifier is found.
// Returns ("", false) when no qualifier is present, the qualifier is empty,
// or the closing parenthesis is not the last character.
//
// This implements the ExtractQualifier algorithm from the functional spec:
//  1. Find the last "(" in the string.
//  2. Find the ")" after it.
//  3. ")" must be the very last character.
//  4. The text between "(" and ")" must be non-empty.
func extractQualifier(logicalName string) (string, bool) {
	// Step 1: last "(" in the string.
	openIdx := strings.LastIndex(logicalName, "(")
	if openIdx < 0 {
		return "", false
	}

	// Step 2: ")" after the "(".
	closeIdx := strings.Index(logicalName[openIdx:], ")")
	if closeIdx < 0 {
		return "", false
	}
	// closeIdx is relative to openIdx.
	closeIdx += openIdx

	// Step 3: ")" must be the last character.
	if closeIdx != len(logicalName)-1 {
		return "", false
	}

	// Step 4 & 5: Extract the content between the delimiters; must be non-empty.
	qualifier := logicalName[openIdx+1 : closeIdx]
	if qualifier == "" {
		return "", false
	}

	return qualifier, true
}

// stripQualifier removes the parenthetical qualifier from a logical name,
// returning the bare name. If there is no qualifier the original string is
// returned unchanged.
//
// Examples:
//
//	stripQualifier("ROOT/x/y(z)") → "ROOT/x/y"
//	stripQualifier("ROOT/x/y")    → "ROOT/x/y"
func stripQualifier(logicalName string) string {
	openIdx := strings.LastIndex(logicalName, "(")
	if openIdx < 0 {
		return logicalName
	}

	// Verify there is a closing paren at the very end.
	if !strings.HasSuffix(logicalName, ")") {
		return logicalName
	}

	return logicalName[:openIdx]
}
