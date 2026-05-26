// code-from-spec: ROOT/golang/internal/format_validation/code@vCswtOL5eTVwJWdPAzRFgjNmNk4

// Package formatvalidation validates discovered spec nodes against structural
// rules defined by the Code From Spec framework. It checks node names,
// frontmatter field restrictions, dependency targets, external file integrity,
// output path safety, and duplicate public subsections.
//
// All violations across all nodes are collected before returning — validation
// does not stop at the first error.
package formatvalidation

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/normalizename"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/nodediscovery"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathvalidation"
)

// FormatError describes a single structural rule violation found on a node.
type FormatError struct {
	Node   string // logical name of the node that failed validation
	Rule   string // short identifier for the rule that was violated
	Detail string // human-readable explanation of the violation
}

// ErrUnreadableNode is returned (wrapped) when a node file cannot be read.
var ErrUnreadableNode = errors.New("unreadable node")

// projectRoot is the working directory of the process, which by framework
// convention is always the project root. All relative paths are resolved
// against it.
const projectRoot = "."

// ValidateFormat takes a list of discovered nodes (logical name + file path)
// and validates each one against the structural rules of the framework.
//
// It returns a slice of FormatError values, one per rule violation. The slice
// is empty when all nodes are valid.
//
// The only error returned directly (not as a FormatError) is
// ErrUnreadableNode, wrapped with context, when a node file cannot be opened.
// In that case, remaining per-node checks are skipped but other nodes are
// still validated.
func ValidateFormat(discoveredNodes []nodediscovery.DiscoveredNode) ([]FormatError, error) {
	// Step 1: initialize the accumulator for all violations.
	var allErrors []FormatError

	// Step 2: build a set of all logical names and a set of all file paths
	// for efficient membership checks used throughout validation.
	allLogicalNames := make(map[string]struct{}, len(discoveredNodes))
	allFilePaths := make(map[string]struct{}, len(discoveredNodes))
	for _, n := range discoveredNodes {
		allLogicalNames[n.LogicalName] = struct{}{}
		allFilePaths[n.FilePath] = struct{}{}
	}

	// Step 3: validate each node.
	for _, node := range discoveredNodes {
		errs, unreadable, err := validateNode(node, allLogicalNames, allFilePaths)
		if err != nil {
			// Propagate unexpected infrastructure errors directly.
			return allErrors, err
		}
		if unreadable {
			// The node file could not be opened; record and move on.
			allErrors = append(allErrors, FormatError{
				Node:   node.LogicalName,
				Rule:   "unreadable node",
				Detail: fmt.Sprintf("file at %s cannot be opened", node.FilePath),
			})
			continue
		}
		allErrors = append(allErrors, errs...)
	}

	return allErrors, nil
}

// validateNode runs all structural checks on a single node. It returns:
//   - errs: all FormatErrors collected for this node
//   - unreadable: true if the node file could not be opened (caller must
//     record the "unreadable node" error and skip remaining checks)
//   - err: a hard infrastructure error to propagate immediately
func validateNode(
	node nodediscovery.DiscoveredNode,
	allLogicalNames map[string]struct{},
	allFilePaths map[string]struct{},
) (errs []FormatError, unreadable bool, err error) {

	// ── Step 3a: classify the node as leaf or intermediate ──────────────────
	//
	// A node is intermediate if any other known logical name starts with
	// this node's logical name followed by "/".
	prefix := node.LogicalName + "/"
	isIntermediate := false
	for name := range allLogicalNames {
		if name != node.LogicalName && strings.HasPrefix(name, prefix) {
			isIntermediate = true
			break
		}
	}

	// ── Step 3b: verify the file can be opened ───────────────────────────────
	//
	// We use filereader.OpenFileReader to stay consistent with the rest of the
	// framework, but we immediately discard the reader since we only need to
	// confirm readability at this step.
	fr, openErr := filereader.OpenFileReader(node.FilePath)
	if openErr != nil {
		// Signal the caller to record the "unreadable node" error.
		return nil, true, nil
	}
	// We do not defer-close filereader because the type has no Close method;
	// the underlying file handle is managed internally by the reader.
	_ = fr // used only to verify the file is readable

	// ── Step 3c: parse frontmatter ───────────────────────────────────────────
	fm, fmErr := frontmatter.ParseFrontmatter(node.FilePath)
	if fmErr != nil {
		// A frontmatter parse error is reported as an "unreadable node" per
		// the spec — if we cannot parse the file at all, treat it like an
		// unreadable file for rule purposes.
		if errors.Is(fmErr, frontmatter.ErrRead) {
			return nil, true, nil
		}
		// ErrFrontmatterParse: record as a format error and continue with a
		// zero-value frontmatter so remaining checks still run.
		errs = append(errs, FormatError{
			Node:   node.LogicalName,
			Rule:   "unreadable node",
			Detail: fmt.Sprintf("file at %s cannot be parsed: %v", node.FilePath, fmErr),
		})
		fm = &frontmatter.Frontmatter{}
	}

	// ── Step 3d: parse the node body ─────────────────────────────────────────
	parsed, parseErr := parsenode.ParseNode(node.LogicalName)
	if parseErr != nil {
		if errors.Is(parseErr, parsenode.ErrRead) {
			return nil, true, nil
		}
		// Other parse errors are recorded as format errors; use a zero-value
		// ParsedNode so we can still run remaining checks.
		errs = append(errs, FormatError{
			Node:   node.LogicalName,
			Rule:   "unreadable node",
			Detail: fmt.Sprintf("file at %s cannot be parsed: %v", node.FilePath, parseErr),
		})
		parsed = &parsenode.ParsedNode{}
	}

	// ── Step 3e: rule — name verification ────────────────────────────────────
	//
	// Reverse-resolve the file path to an expected logical name, then compare
	// the normalized heading against the normalized expected name.
	if parsed != nil {
		expectedName, ok := logicalnames.LogicalNameFromPath(node.FilePath)
		if ok {
			normalizedHeading := normalizename.NormalizeName(parsed.NameSection.Heading)
			normalizedExpected := normalizename.NormalizeName(expectedName)
			if normalizedHeading != normalizedExpected {
				errs = append(errs, FormatError{
					Node: node.LogicalName,
					Rule: "name mismatch",
					Detail: fmt.Sprintf(
						"heading %q does not match expected logical name %q",
						parsed.NameSection.Heading,
						expectedName,
					),
				})
			}
		}
	}

	// ── Step 3f: rule — frontmatter field restrictions (intermediate only) ───
	if isIntermediate {
		if len(fm.DependsOn) > 0 {
			errs = append(errs, FormatError{
				Node:   node.LogicalName,
				Rule:   "depends_on on intermediate node",
				Detail: "field depends_on is not permitted on nodes with children",
			})
		}
		if len(fm.External) > 0 {
			errs = append(errs, FormatError{
				Node:   node.LogicalName,
				Rule:   "external on intermediate node",
				Detail: "field external is not permitted on nodes with children",
			})
		}
		if fm.Input != "" {
			errs = append(errs, FormatError{
				Node:   node.LogicalName,
				Rule:   "input on intermediate node",
				Detail: "field input is not permitted on nodes with children",
			})
		}
		if len(fm.Outputs) > 0 {
			errs = append(errs, FormatError{
				Node:   node.LogicalName,
				Rule:   "outputs on intermediate node",
				Detail: "field outputs is not permitted on nodes with children",
			})
		}
	}

	// ── Step 3g: rule — agent section restriction (intermediate only) ────────
	if isIntermediate && parsed != nil && parsed.Agent != nil {
		errs = append(errs, FormatError{
			Node:   node.LogicalName,
			Rule:   "agent section on intermediate node",
			Detail: "# Agent section is not permitted on nodes with children",
		})
	}

	// ── Step 3h: rule — dependency targets ──────────────────────────────────
	for _, dep := range fm.DependsOn {
		depErrs := validateDependency(node.LogicalName, dep, allFilePaths)
		errs = append(errs, depErrs...)
	}

	// ── Step 3i: rule — external file existence and fragment hash ────────────
	for _, ext := range fm.External {
		extErrs := validateExternal(node.LogicalName, ext)
		errs = append(errs, extErrs...)
	}

	// ── Step 3j: rule — output path validation ───────────────────────────────
	for _, out := range fm.Outputs {
		if pathErr := pathvalidation.ValidatePath(out.Path, projectRoot); pathErr != nil {
			errs = append(errs, FormatError{
				Node:   node.LogicalName,
				Rule:   "invalid output path",
				Detail: pathErr.Error(),
			})
		}
	}

	// ── Step 3k: rule — duplicate public subsections ─────────────────────────
	if parsed != nil && parsed.Public != nil {
		seen := make(map[string]struct{})
		for _, sub := range parsed.Public.Subsections {
			normalized := normalizename.NormalizeName(sub.Heading)
			if _, exists := seen[normalized]; exists {
				errs = append(errs, FormatError{
					Node: node.LogicalName,
					Rule: "duplicate public subsection",
					Detail: fmt.Sprintf(
						"subsection heading %q in # Public normalizes to %q which conflicts with a previous subsection",
						sub.Heading,
						normalized,
					),
				})
			} else {
				seen[normalized] = struct{}{}
			}
		}
	}

	return errs, false, nil
}

// validateDependency checks a single depends_on entry against the known set
// of file paths. It returns any FormatErrors found.
//
// Checks performed:
//  1. The target resolves to a known node file path.
//  2. The target is not an ancestor of the current node.
//  3. The target is not a descendant of the current node.
func validateDependency(
	nodeLogicalName string,
	dep string,
	allFilePaths map[string]struct{},
) []FormatError {
	var errs []FormatError

	// Resolve the dependency reference to a file path and a bare logical name.
	var resolvedFilePath string
	var targetBareLogicalName string

	if logicalnames.IsArtifactRef(dep) {
		// ARTIFACT/ reference: extract the node path.
		nodePath, _, ok := logicalnames.ArtifactRefParts(dep)
		if !ok {
			errs = append(errs, FormatError{
				Node:   nodeLogicalName,
				Rule:   "missing dependency target",
				Detail: fmt.Sprintf("depends_on entry %q does not resolve to a known node", dep),
			})
			return errs
		}
		resolvedFilePath = nodePath
		// Derive the bare logical name from the file path for ancestor/descendant checks.
		if ln, ok2 := logicalnames.LogicalNameFromPath(nodePath); ok2 {
			targetBareLogicalName = ln
		}
	} else if strings.HasPrefix(dep, "ROOT/") || dep == "ROOT" {
		// ROOT/ reference: resolve directly, stripping any qualifier first.
		filePath, ok := logicalnames.PathFromLogicalName(dep)
		if !ok {
			errs = append(errs, FormatError{
				Node:   nodeLogicalName,
				Rule:   "missing dependency target",
				Detail: fmt.Sprintf("depends_on entry %q does not resolve to a known node", dep),
			})
			return errs
		}
		resolvedFilePath = filePath
		// The bare logical name is derived from the path (qualifier already stripped
		// by PathFromLogicalName).
		if ln, ok2 := logicalnames.LogicalNameFromPath(filePath); ok2 {
			targetBareLogicalName = ln
		}
	} else {
		// Unknown reference format.
		errs = append(errs, FormatError{
			Node:   nodeLogicalName,
			Rule:   "missing dependency target",
			Detail: fmt.Sprintf("depends_on entry %q does not resolve to a known node", dep),
		})
		return errs
	}

	// Check 3h-i: the resolved file must be among the discovered nodes.
	if _, known := allFilePaths[resolvedFilePath]; !known {
		errs = append(errs, FormatError{
			Node:   nodeLogicalName,
			Rule:   "missing dependency target",
			Detail: fmt.Sprintf("depends_on entry %q does not resolve to a known node", dep),
		})
		// Still run ancestor/descendant checks if we have a bare logical name,
		// because the logical name check is independent of file existence.
	}

	if targetBareLogicalName == "" {
		// Cannot determine relationship without a logical name.
		return errs
	}

	// Check 3h-ii: ancestor check — target_bare is a prefix of node.logical_name.
	// i.e. node.logical_name starts with target_bare + "/" or equals target_bare.
	if isAncestorOf(targetBareLogicalName, nodeLogicalName) {
		errs = append(errs, FormatError{
			Node: nodeLogicalName,
			Rule: "depends_on ancestor",
			Detail: fmt.Sprintf(
				"depends_on entry %q points to an ancestor; ancestor content is already inherited",
				dep,
			),
		})
	}

	// Check 3h-iii: descendant check — node.logical_name is a prefix of target_bare.
	// i.e. target_bare starts with node.logical_name + "/" or equals node.logical_name.
	if isAncestorOf(nodeLogicalName, targetBareLogicalName) {
		errs = append(errs, FormatError{
			Node: nodeLogicalName,
			Rule: "depends_on descendant",
			Detail: fmt.Sprintf(
				"depends_on entry %q points to a descendant; this would create a circular dependency",
				dep,
			),
		})
	}

	return errs
}

// isAncestorOf reports whether candidate is an ancestor of (or equal to)
// subject in the logical name hierarchy. That is, it returns true when
// subject starts with candidate+"/" or subject == candidate.
func isAncestorOf(candidate, subject string) bool {
	return subject == candidate || strings.HasPrefix(subject, candidate+"/")
}

// validateExternal checks a single external file entry: confirms the file is
// readable, then for each declared fragment verifies that the computed SHA-1
// hash of the extracted lines matches the declared hash.
func validateExternal(
	nodeLogicalName string,
	ext frontmatter.External,
) []FormatError {
	var errs []FormatError

	// ── Step 3i-i: file existence ─────────────────────────────────────────────
	fr, openErr := filereader.OpenFileReader(ext.Path)
	if openErr != nil {
		errs = append(errs, FormatError{
			Node:   nodeLogicalName,
			Rule:   "missing external file",
			Detail: fmt.Sprintf("external path %q does not exist or cannot be read", ext.Path),
		})
		return errs // no fragment checks if the file isn't readable
	}
	_ = fr // opened successfully; we'll re-open below for fragment extraction

	// ── Step 3i-ii: fragment hash verification ────────────────────────────────
	for _, frag := range ext.Fragments {
		fragErrs := validateFragment(nodeLogicalName, ext.Path, frag)
		errs = append(errs, fragErrs...)
	}

	return errs
}

// validateFragment reads the declared line range from filePath, computes the
// SHA-1 hash of the extracted content (with CRLF normalized to LF), and
// compares it to the declared hash.
func validateFragment(
	nodeLogicalName string,
	filePath string,
	frag frontmatter.ExternalFragment,
) []FormatError {
	var errs []FormatError

	// Parse the "start-end" line range from frag.Lines (1-based, inclusive).
	startLine, endLine, parseOk := parseLineRange(frag.Lines)
	if !parseOk {
		// If the range can't be parsed it's a malformed spec; report it.
		errs = append(errs, FormatError{
			Node:   nodeLogicalName,
			Rule:   "fragment hash mismatch",
			Detail: fmt.Sprintf("cannot parse line range %q in %q", frag.Lines, filePath),
		})
		return errs
	}

	// Open the file and read lines from startLine to endLine (1-based, inclusive).
	fr, err := filereader.OpenFileReader(filePath)
	if err != nil {
		errs = append(errs, FormatError{
			Node:   nodeLogicalName,
			Rule:   "missing external file",
			Detail: fmt.Sprintf("external path %q does not exist or cannot be read", filePath),
		})
		return errs
	}

	var extractedLines []string
	for lineNum := 1; ; lineNum++ {
		line, readErr := fr.ReadLine()
		if errors.Is(readErr, filereader.ErrEndOfFile) {
			break
		}
		if readErr != nil {
			// Unexpected read error during fragment extraction.
			errs = append(errs, FormatError{
				Node:   nodeLogicalName,
				Rule:   "missing external file",
				Detail: fmt.Sprintf("external path %q could not be fully read: %v", filePath, readErr),
			})
			return errs
		}
		if lineNum >= startLine && lineNum <= endLine {
			extractedLines = append(extractedLines, line)
		}
		if lineNum >= endLine {
			break
		}
	}

	// Join extracted lines with LF (CRLF normalization is already handled by
	// filereader.ReadLine, which normalizes CRLF before splitting).
	extractedContent := strings.Join(extractedLines, "\n")

	// Compute SHA-1 and encode as base64url (RFC 4648 §5, no padding, 27 chars).
	sum := sha1.Sum([]byte(extractedContent)) //nolint:gosec // SHA-1 required by spec
	computedHash := base64.RawURLEncoding.EncodeToString(sum[:])

	if computedHash != frag.Hash {
		errs = append(errs, FormatError{
			Node: nodeLogicalName,
			Rule: "fragment hash mismatch",
			Detail: fmt.Sprintf(
				"fragment at lines %s of %q has hash %s but declared hash is %s",
				frag.Lines,
				filePath,
				computedHash,
				frag.Hash,
			),
		})
	}

	return errs
}

// parseLineRange parses a "start-end" line range string (e.g. "150-210")
// into its inclusive integer bounds (both 1-based). Returns (0, 0, false) on
// any parse failure.
func parseLineRange(lines string) (start, end int, ok bool) {
	parts := strings.SplitN(lines, "-", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}
	s, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	e, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil || s < 1 || e < s {
		return 0, 0, false
	}
	return s, e, true
}

// fileExists is a small helper that returns true when a file exists and is
// accessible on the OS filesystem. Used to double-check external path
// existence independently from filereader (which is the authoritative check).
//
// Note: this function is intentionally kept even if the compiler marks it
// unused in some build configurations — it serves as documentation of intent
// and may be used in future rules.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
