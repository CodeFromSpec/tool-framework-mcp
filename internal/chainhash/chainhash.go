// code-from-spec: ROOT/golang/internal/chain_hash/code@6x3Y9ghWPeebVe3sV7bUyC3VnNk

// Package chainhash computes the chain hash for a given spec node.
//
// The chain hash is a 27-character base64url-encoded SHA-1 digest used for
// artifact staleness detection. It is built by collecting SHA-1 digests from
// each position in the chain (ancestors, depends_on, external files, the
// target itself, and its input), then hashing the concatenation of those raw
// digests.
//
// See the chain-hash spec document for the full algorithm description.
package chainhash

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
)

// ComputeChainHash returns the 27-character base64url chain hash for the
// spec node identified by logicalName (must start with "ROOT/").
//
// Errors:
//   - "invalid logical name: <detail>" — cannot resolve the logical name
//   - "unreadable file: <path>"        — a required file cannot be read
func ComputeChainHash(logicalName string) (string, error) {
	// ----------------------------------------------------------------
	// Preparation
	// ----------------------------------------------------------------

	// Step 1 — only ROOT/ names are supported.
	if !strings.HasPrefix(logicalName, "ROOT/") && logicalName != "ROOT" {
		return "", fmt.Errorf("invalid logical name: only ROOT/ names are supported")
	}

	// Step 2 — read and normalize the target node file.
	targetPath, ok := logicalnames.PathFromLogicalName(logicalName)
	if !ok {
		return "", fmt.Errorf("invalid logical name: cannot resolve path for %s", logicalName)
	}

	targetRaw, err := os.ReadFile(targetPath)
	if err != nil {
		return "", fmt.Errorf("unreadable file: %s", targetPath)
	}
	targetContent := normalizeCRLF(string(targetRaw))

	// Step 3 — parse the target node's frontmatter.
	targetFM, err := frontmatter.ParseFrontmatter(targetPath)
	if err != nil {
		return "", fmt.Errorf("unreadable file: %s", targetPath)
	}

	// ----------------------------------------------------------------
	// Accumulator — a list of raw 20-byte SHA-1 digests.
	// ----------------------------------------------------------------
	var digestList [][]byte // each entry is exactly 20 bytes

	// ----------------------------------------------------------------
	// Step 1 — Ancestor # Public hashes (root-first, not including target)
	// ----------------------------------------------------------------

	// Build the ancestor chain by walking up from the target's parent to ROOT.
	ancestors, err := buildAncestorChain(logicalName)
	if err != nil {
		return "", err
	}

	for _, ancestor := range ancestors {
		ancestorPath, ok := logicalnames.PathFromLogicalName(ancestor)
		if !ok {
			return "", fmt.Errorf("invalid logical name: cannot resolve path for %s", ancestor)
		}
		content, err := readAndNormalize(ancestorPath)
		if err != nil {
			return "", err
		}
		section := extractSection(content, "# Public")
		if section == "" {
			// No # Public section — skip this ancestor.
			continue
		}
		digestList = append(digestList, sha1Digest(section))
	}

	// ----------------------------------------------------------------
	// Step 2 — depends_on hashes
	// ----------------------------------------------------------------

	// Sort depends_on entries alphabetically by their logical name string.
	dependsOn := make([]string, len(targetFM.DependsOn))
	copy(dependsOn, targetFM.DependsOn)
	sort.Strings(dependsOn)

	for _, dep := range dependsOn {
		digests, err := dependsOnDigests(dep)
		if err != nil {
			return "", err
		}
		digestList = append(digestList, digests...)
	}

	// ----------------------------------------------------------------
	// Step 3 — external file hashes
	// ----------------------------------------------------------------

	// Sort external entries alphabetically by path.
	externals := make([]frontmatter.External, len(targetFM.External))
	copy(externals, targetFM.External)
	sort.Slice(externals, func(i, j int) bool {
		return externals[i].Path < externals[j].Path
	})

	for _, ext := range externals {
		digest, err := externalDigest(ext)
		if err != nil {
			return "", err
		}
		digestList = append(digestList, digest)
	}

	// ----------------------------------------------------------------
	// Step 4 — Target # Public hash
	// ----------------------------------------------------------------

	if section := extractSection(targetContent, "# Public"); section != "" {
		digestList = append(digestList, sha1Digest(section))
	}

	// ----------------------------------------------------------------
	// Step 5 — Target # Agent hash
	// ----------------------------------------------------------------

	if section := extractSection(targetContent, "# Agent"); section != "" {
		digestList = append(digestList, sha1Digest(section))
	}

	// ----------------------------------------------------------------
	// Step 6 — Input hash
	// ----------------------------------------------------------------

	if targetFM.Input != "" {
		filePath, err := resolveArtifactFilePath(targetFM.Input)
		if err != nil {
			return "", err
		}
		content, err := readAndNormalize(filePath)
		if err != nil {
			return "", err
		}
		stripped := stripFrontmatter(content)
		digestList = append(digestList, sha1Digest(stripped))
	}

	// ----------------------------------------------------------------
	// Step 7 — Final hash
	// ----------------------------------------------------------------

	// Concatenate all raw 20-byte digests.
	var combined []byte
	for _, d := range digestList {
		combined = append(combined, d...)
	}

	// Hash the concatenation to produce the final digest.
	final := sha1Digest(string(combined))

	// Encode as base64url without padding (RFC 4648 §5, no padding).
	// The result is always exactly 27 characters for a 20-byte input.
	encoded := base64.RawURLEncoding.EncodeToString(final)
	return encoded, nil
}

// ----------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------

// normalizeCRLF replaces all CRLF sequences with LF.
func normalizeCRLF(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

// sha1Digest returns the raw 20-byte SHA-1 digest of the given string.
func sha1Digest(content string) []byte {
	h := sha1.New()
	// sha1.Write never returns an error.
	h.Write([]byte(content)) //nolint:errcheck
	return h.Sum(nil)
}

// readAndNormalize reads a file from disk and normalizes CRLF → LF.
// Returns "unreadable file: <path>" on error.
func readAndNormalize(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("unreadable file: %s", path)
	}
	return normalizeCRLF(string(raw)), nil
}

// extractSection returns the text of the section that starts with heading
// (inclusive) up to the next heading at the same or higher level (exclusive),
// or end of file. Returns empty string if heading is not found.
//
// The heading level is determined by counting leading '#' characters.
// A terminating line is one that starts with 1..level '#' characters
// followed by a space.
func extractSection(fileContent, heading string) string {
	lines := strings.Split(fileContent, "\n")

	// Find the heading line.
	startIdx := -1
	for i, line := range lines {
		if line == heading {
			startIdx = i
			break
		}
	}
	if startIdx == -1 {
		return ""
	}

	// Determine the heading level.
	level := 0
	for _, ch := range heading {
		if ch == '#' {
			level++
		} else {
			break
		}
	}

	// Collect lines from the heading until a heading at the same or higher
	// level is encountered.
	var collected []string
	for i := startIdx; i < len(lines); i++ {
		if i > startIdx {
			// Check if this line is a heading at level 1..level.
			if isHeadingAtOrAboveLevel(lines[i], level) {
				break
			}
		}
		collected = append(collected, lines[i])
	}

	return strings.Join(collected, "\n")
}

// isHeadingAtOrAboveLevel returns true if line is a Markdown heading of
// level 1 through maxLevel (i.e., it starts with 1..maxLevel '#' characters
// followed by a space).
func isHeadingAtOrAboveLevel(line string, maxLevel int) bool {
	if len(line) == 0 || line[0] != '#' {
		return false
	}
	count := 0
	for _, ch := range line {
		if ch == '#' {
			count++
		} else {
			break
		}
	}
	if count > maxLevel {
		return false
	}
	// Must be followed by a space.
	if len(line) <= count || line[count] != ' ' {
		return false
	}
	return true
}

// extractSubsection returns the text of subsectionHeading within the parent
// section identified by parentHeading. Returns empty string if either heading
// is not found.
func extractSubsection(fileContent, parentHeading, subsectionHeading string) string {
	parentSection := extractSection(fileContent, parentHeading)
	if parentSection == "" {
		return ""
	}
	return extractSection(parentSection, subsectionHeading)
}

// stripFrontmatter removes the leading YAML frontmatter block (delimited by
// "---" lines) from fileContent. Returns content unchanged if no frontmatter
// is detected.
func stripFrontmatter(fileContent string) string {
	lines := strings.Split(fileContent, "\n")
	if len(lines) == 0 || lines[0] != "---" {
		return fileContent
	}
	// Find the closing "---".
	closingIdx := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			closingIdx = i
			break
		}
	}
	if closingIdx == -1 {
		// No closing delimiter found — treat as no frontmatter.
		return fileContent
	}
	// Return everything after the closing "---".
	rest := lines[closingIdx+1:]
	return strings.Join(rest, "\n")
}

// buildAncestorChain returns the list of ancestor logical names for
// logicalName, ordered from ROOT down to the immediate parent (not including
// the target itself).
func buildAncestorChain(logicalName string) ([]string, error) {
	var ancestors []string
	current := logicalName
	for {
		hasParent, ok := logicalnames.HasParent(current)
		if !ok {
			return nil, fmt.Errorf("invalid logical name: cannot determine parent of %s", current)
		}
		if !hasParent {
			break
		}
		parent, ok := logicalnames.ParentLogicalName(current)
		if !ok {
			return nil, fmt.Errorf("invalid logical name: cannot derive parent of %s", current)
		}
		// Prepend so the final slice is root-first.
		ancestors = append([]string{parent}, ancestors...)
		current = parent
	}
	return ancestors, nil
}

// dependsOnDigests returns the SHA-1 digest(s) for a single depends_on entry.
// Returns a slice because each entry contributes exactly one digest (or zero
// if the relevant section is empty).
func dependsOnDigests(dep string) ([][]byte, error) {
	switch {
	case logicalnames.IsArtifactRef(dep):
		// ARTIFACT/ reference — hash full artifact content (minus frontmatter).
		filePath, err := resolveArtifactFilePath(dep)
		if err != nil {
			return nil, err
		}
		content, err := readAndNormalize(filePath)
		if err != nil {
			return nil, err
		}
		stripped := stripFrontmatter(content)
		return [][]byte{sha1Digest(stripped)}, nil

	default:
		// ROOT/ reference — may have a qualifier.
		hasQualifier, ok := logicalnames.HasQualifier(dep)
		if !ok {
			return nil, fmt.Errorf("invalid logical name: %s", dep)
		}

		depPath, ok := logicalnames.PathFromLogicalName(dep)
		if !ok {
			return nil, fmt.Errorf("invalid logical name: cannot resolve path for %s", dep)
		}
		content, err := readAndNormalize(depPath)
		if err != nil {
			return nil, err
		}

		if hasQualifier {
			// ROOT/x/y(z) — hash the ## z subsection within # Public.
			qualifier, ok := logicalnames.QualifierName(dep)
			if !ok {
				return nil, fmt.Errorf("invalid logical name: cannot extract qualifier from %s", dep)
			}
			subsection := extractSubsection(content, "# Public", "## "+qualifier)
			if subsection == "" {
				return nil, nil // skip
			}
			return [][]byte{sha1Digest(subsection)}, nil
		}

		// ROOT/x/y — hash the entire # Public section.
		section := extractSection(content, "# Public")
		if section == "" {
			return nil, nil // skip
		}
		return [][]byte{sha1Digest(section)}, nil
	}
}

// externalDigest returns the SHA-1 digest for a single external entry.
// If the entry has no fragments, the full file content is hashed.
// If it has fragments, the concatenation of the specified line ranges is hashed.
func externalDigest(ext frontmatter.External) ([]byte, error) {
	content, err := readAndNormalize(ext.Path)
	if err != nil {
		return nil, err
	}

	if len(ext.Fragments) == 0 {
		// No fragments — hash the full file content.
		return sha1Digest(content), nil
	}

	// Hash the concatenation of each fragment's line range.
	lines := strings.Split(content, "\n")
	var buf strings.Builder

	for _, frag := range ext.Fragments {
		// Parse the line range "<start>-<end>" (1-based, inclusive).
		start, end, err := parseLineRange(frag.Lines)
		if err != nil {
			return nil, fmt.Errorf("invalid fragment line range %q in %s: %w", frag.Lines, ext.Path, err)
		}
		// Clamp to valid indices (lines slice is 0-based).
		if start < 1 {
			start = 1
		}
		if end > len(lines) {
			end = len(lines)
		}
		// Extract lines[start-1 .. end] (inclusive both ends).
		extracted := lines[start-1 : end]
		// Join with LF and append a trailing LF.
		buf.WriteString(strings.Join(extracted, "\n"))
		buf.WriteByte('\n')
	}

	return sha1Digest(buf.String()), nil
}

// parseLineRange parses a "<start>-<end>" string (1-based, inclusive).
func parseLineRange(s string) (int, int, error) {
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected format <start>-<end>, got %q", s)
	}
	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start line %q: %w", parts[0], err)
	}
	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end line %q: %w", parts[1], err)
	}
	return start, end, nil
}

// resolveArtifactFilePath resolves an "ARTIFACT/x/y(id)" logical name to the
// file path of the referenced artifact on disk.
//
// Steps:
//  1. Parse the ARTIFACT/ reference to get node path and artifact ID.
//  2. Parse the node's frontmatter to find the output with that ID.
//  3. Return the output's path.
func resolveArtifactFilePath(artifactLogicalName string) (string, error) {
	nodePath, artifactID, ok := logicalnames.ArtifactRefParts(artifactLogicalName)
	if !ok {
		return "", fmt.Errorf("invalid logical name: %s", artifactLogicalName)
	}

	fm, err := frontmatter.ParseFrontmatter(nodePath)
	if err != nil {
		return "", fmt.Errorf("unreadable file: %s", nodePath)
	}

	for _, out := range fm.Outputs {
		if out.ID == artifactID {
			return out.Path, nil
		}
	}

	return "", fmt.Errorf("invalid logical name: artifact id %s not found in %s", artifactID, nodePath)
}
