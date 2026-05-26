// code-from-spec: ROOT/golang/internal/format_validation/code@Rmdxg4i-zcRo8Nr9z5wX7ix6dmY

// Package formatvalidation validates spec nodes against structural rules.
// It takes a list of discovered nodes and checks each one for conformance
// with the framework's format requirements (frontmatter restrictions,
// section rules, dependency targets, external file existence, output paths,
// etc.). All errors across all nodes are collected before returning —
// validation never stops at the first error.
package formatvalidation

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/normalizename"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathvalidation"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/nodediscovery"
)

// FormatError records a single format rule violation found in a spec node.
type FormatError struct {
	Node   string // logical name of the node where the error was found
	Rule   string // name of the rule that was violated
	Detail string // human-readable description of the specific violation
}

// ErrUnreadableNode is the sentinel returned when a node file cannot be read.
var ErrUnreadableNode = errors.New("unreadable node")

// ValidateFormat takes a list of discovered nodes (logical name + file path)
// and validates each one against all structural rules. Returns a slice of
// FormatErrors (empty if all nodes are valid). All errors from all nodes are
// collected before returning.
func ValidateFormat(discoveredNodes []nodediscovery.DiscoveredNode) ([]FormatError, error) {
	var errs []FormatError

	// Build a set of file paths for quick ancestor/child lookup.
	// Also build a map from file path to logical name for dependency checks.
	nodeFilePathSet := make(map[string]bool, len(discoveredNodes))
	for _, n := range discoveredNodes {
		nodeFilePathSet[n.FilePath] = true
	}

	for _, node := range discoveredNodes {
		// Step 2a: Open the file to verify it is readable.
		reader, openErr := filereader.OpenFileReader(node.FilePath)
		if openErr != nil {
			errs = append(errs, FormatError{
				Node:   node.LogicalName,
				Rule:   "unreadable node",
				Detail: fmt.Sprintf("cannot open file at %s", node.FilePath),
			})
			// Skip all remaining steps for this node.
			continue
		}

		// Step 2b: Parse frontmatter.
		fm, fmErr := frontmatter.ParseFrontmatter(node.FilePath)
		if fmErr != nil {
			errs = append(errs, FormatError{
				Node:   node.LogicalName,
				Rule:   "unreadable node",
				Detail: fmt.Sprintf("frontmatter parse error: %s", fmErr.Error()),
			})
			reader.Close()
			continue
		}

		// Step 2c: Parse the body.
		parsed, parseErr := parsenode.ParseNode(node.LogicalName)
		if parseErr != nil {
			rule := "unreadable node"
			switch {
			case errors.Is(parseErr, parsenode.ErrInvalidNodeName):
				rule = "name_verification"
			case errors.Is(parseErr, parsenode.ErrDuplicatePublic):
				rule = "duplicate_public_section"
			case errors.Is(parseErr, parsenode.ErrDuplicateSubsection):
				rule = "duplicate_public_subsections"
			case errors.Is(parseErr, parsenode.ErrUnexpectedContent):
				rule = "unexpected_content"
			}
			errs = append(errs, FormatError{
				Node:   node.LogicalName,
				Rule:   rule,
				Detail: parseErr.Error(),
			})
			reader.Close()
			continue
		}

		// Step 2d: Close the reader — we have all the data we need.
		reader.Close()

		// Step 2e: Determine leaf vs. intermediate.
		// A node is intermediate if any other discovered node's logical name
		// starts with this node's logical name followed by "/".
		isIntermediate := false
		prefix := node.LogicalName + "/"
		for _, other := range discoveredNodes {
			if other.LogicalName == node.LogicalName {
				continue
			}
			if strings.HasPrefix(other.LogicalName, prefix) {
				isIntermediate = true
				break
			}
		}

		// Step 2f: Run all validation rules, collecting errors.

		// Rule: name_verification — applied to all nodes.
		errs = append(errs, ruleNameVerification(node, parsed)...)

		if isIntermediate {
			// Rule: frontmatter_field_restrictions — intermediate nodes only.
			errs = append(errs, ruleFrontmatterFieldRestrictions(node, fm)...)

			// Rule: agent_section_restrictions — intermediate nodes only.
			errs = append(errs, ruleAgentSectionRestrictions(node, parsed)...)
		} else {
			// Leaf-only rules.

			// Rule: dependency_targets — leaf nodes with depends_on.
			if len(fm.DependsOn) > 0 {
				errs = append(errs, ruleDependencyTargets(node, fm, nodeFilePathSet)...)
			}

			// Rule: external_file_existence — leaf nodes with external.
			if len(fm.External) > 0 {
				errs = append(errs, ruleExternalFileExistence(node, fm)...)
			}

			// Rule: output_path_validation — leaf nodes with outputs.
			if len(fm.Outputs) > 0 {
				errs = append(errs, ruleOutputPathValidation(node, fm)...)
			}
		}

		// Rule: duplicate_public_subsections — all nodes with a # Public section.
		if parsed.Public != nil {
			errs = append(errs, ruleDuplicatePublicSubsections(node, parsed)...)
		}
	}

	return errs, nil
}

// ---------------------------------------------------------------------------
// Rule implementations
// ---------------------------------------------------------------------------

// ruleNameVerification verifies that the first heading in the node file
// matches the logical name derived from the file path.
func ruleNameVerification(node nodediscovery.DiscoveredNode, parsed *parsenode.ParsedNode) []FormatError {
	var errs []FormatError

	// Derive the expected logical name from the file path.
	expected, ok := logicalnames.LogicalNameFromPath(node.FilePath)
	if !ok {
		errs = append(errs, FormatError{
			Node:   node.LogicalName,
			Rule:   "name_verification",
			Detail: fmt.Sprintf("cannot derive logical name from path: %s", node.FilePath),
		})
		return errs
	}

	// Get the actual first heading from the parsed node.
	actual := parsed.NameSection.Heading

	// Normalize both values before comparing.
	normalizedExpected := normalizename.NormalizeName(expected)
	normalizedActual := normalizename.NormalizeName(actual)

	if normalizedExpected != normalizedActual {
		errs = append(errs, FormatError{
			Node:   node.LogicalName,
			Rule:   "name_verification",
			Detail: fmt.Sprintf("heading %q does not match expected logical name %q", actual, expected),
		})
	}

	return errs
}

// ruleFrontmatterFieldRestrictions checks that intermediate nodes do not
// declare frontmatter fields that are only permitted on leaf nodes.
func ruleFrontmatterFieldRestrictions(node nodediscovery.DiscoveredNode, fm *frontmatter.Frontmatter) []FormatError {
	var errs []FormatError

	if len(fm.DependsOn) > 0 {
		errs = append(errs, FormatError{
			Node:   node.LogicalName,
			Rule:   "frontmatter_field_restrictions",
			Detail: `intermediate node must not have "depends_on"`,
		})
	}

	if len(fm.External) > 0 {
		errs = append(errs, FormatError{
			Node:   node.LogicalName,
			Rule:   "frontmatter_field_restrictions",
			Detail: `intermediate node must not have "external"`,
		})
	}

	if fm.Input != "" {
		errs = append(errs, FormatError{
			Node:   node.LogicalName,
			Rule:   "frontmatter_field_restrictions",
			Detail: `intermediate node must not have "input"`,
		})
	}

	if len(fm.Outputs) > 0 {
		errs = append(errs, FormatError{
			Node:   node.LogicalName,
			Rule:   "frontmatter_field_restrictions",
			Detail: `intermediate node must not have "outputs"`,
		})
	}

	return errs
}

// ruleAgentSectionRestrictions checks that intermediate nodes do not have
// an # Agent section.
func ruleAgentSectionRestrictions(node nodediscovery.DiscoveredNode, parsed *parsenode.ParsedNode) []FormatError {
	var errs []FormatError

	if parsed.Agent != nil {
		errs = append(errs, FormatError{
			Node:   node.LogicalName,
			Rule:   "agent_section_restrictions",
			Detail: `intermediate node must not have a "# Agent" section`,
		})
	}

	return errs
}

// ruleDependencyTargets validates each depends_on entry on a leaf node:
// - the reference must resolve to a known path
// - ROOT/ targets must exist in discoveredNodes
// - ROOT/ targets must not be ancestors or descendants of the node
func ruleDependencyTargets(node nodediscovery.DiscoveredNode, fm *frontmatter.Frontmatter, nodeFilePathSet map[string]bool) []FormatError {
	var errs []FormatError

	for _, dep := range fm.DependsOn {
		if logicalnames.IsArtifactRef(dep) {
			// ARTIFACT/ reference — only resolution is checked (existence of
			// the artifact file is not validated here per the spec pseudocode
			// step 2 for ROOT/ references; ARTIFACT/ refs continue after step 1).
			_, _, ok := logicalnames.ArtifactRefParts(dep)
			if !ok {
				errs = append(errs, FormatError{
					Node:   node.LogicalName,
					Rule:   "dependency_targets",
					Detail: fmt.Sprintf("cannot resolve depends_on entry %q: not a valid ARTIFACT/ reference", dep),
				})
			}
			// No further checks for ARTIFACT/ references per the spec.
			continue
		}

		// ROOT/ reference.
		resolvedPath, ok := logicalnames.PathFromLogicalName(dep)
		if !ok {
			errs = append(errs, FormatError{
				Node:   node.LogicalName,
				Rule:   "dependency_targets",
				Detail: fmt.Sprintf("cannot resolve depends_on entry %q: not a valid ROOT/ reference", dep),
			})
			continue
		}

		// Step 2: Verify the resolved file exists in discoveredNodes.
		if !nodeFilePathSet[resolvedPath] {
			errs = append(errs, FormatError{
				Node:   node.LogicalName,
				Rule:   "dependency_targets",
				Detail: fmt.Sprintf("depends_on target %q does not exist", dep),
			})
			continue
		}

		// Strip qualifier to get the bare logical name for relationship checks.
		bareDep := dep
		if _, hasQual := logicalnames.QualifierName(dep); hasQual {
			// Strip the qualifier: remove the "(qualifier)" suffix.
			// PathFromLogicalName already strips it, but we need the bare name
			// for string prefix checks.
			if idx := strings.Index(dep, "("); idx >= 0 {
				bareDep = dep[:idx]
			}
		}

		// Step 3: Check ancestor relationship.
		// If node.LogicalName starts with bareDep + "/" then dep is an ancestor.
		if strings.HasPrefix(node.LogicalName, bareDep+"/") {
			errs = append(errs, FormatError{
				Node:   node.LogicalName,
				Rule:   "dependency_targets",
				Detail: fmt.Sprintf("depends_on %q points to an ancestor (already inherited)", dep),
			})
		}

		// Step 4: Check descendant relationship.
		// If bareDep starts with node.LogicalName + "/" then dep is a descendant.
		if strings.HasPrefix(bareDep, node.LogicalName+"/") {
			errs = append(errs, FormatError{
				Node:   node.LogicalName,
				Rule:   "dependency_targets",
				Detail: fmt.Sprintf("depends_on %q points to a descendant (creates circular dependency)", dep),
			})
		}
	}

	return errs
}

// ruleExternalFileExistence validates each external entry on a leaf node:
// - path must pass ValidatePath
// - the file must be openable
// - if fragments are declared, each fragment's hash must match
func ruleExternalFileExistence(node nodediscovery.DiscoveredNode, fm *frontmatter.Frontmatter) []FormatError {
	var errs []FormatError

	projectRoot, wdErr := os.Getwd()
	if wdErr != nil {
		return errs
	}

	for _, ext := range fm.External {
		if err := pathvalidation.ValidatePath(ext.Path, projectRoot); err != nil {
			errs = append(errs, FormatError{
				Node:   node.LogicalName,
				Rule:   "external_file_existence",
				Detail: fmt.Sprintf("invalid path %q: %s", ext.Path, err.Error()),
			})
			continue
		}

		// Step 2: Verify the file can be opened.
		r, openErr := filereader.OpenFileReader(ext.Path)
		if openErr != nil {
			errs = append(errs, FormatError{
				Node:   node.LogicalName,
				Rule:   "external_file_existence",
				Detail: fmt.Sprintf("external file not found: %q", ext.Path),
			})
			continue
		}

		// Step 3: If no fragments, close and continue.
		if len(ext.Fragments) == 0 {
			r.Close()
			continue
		}

		// Step 4: Validate each fragment's hash.
		// Close the reader opened above — we reopen for each fragment to avoid
		// reader state issues (as specified in the pseudocode).
		r.Close()

		for _, frag := range ext.Fragments {
			fragErrs := validateFragment(node.LogicalName, ext.Path, frag)
			errs = append(errs, fragErrs...)
		}
	}

	return errs
}

// validateFragment opens the external file, reads the declared line range,
// computes its SHA-1 hash (base64url, no padding), and compares with the
// declared hash. Returns a FormatError slice (empty on success).
func validateFragment(nodeName, filePath string, frag frontmatter.ExternalFragment) []FormatError {
	// Parse "start-end" from frag.Lines.
	start, end, parseErr := parseLineRange(frag.Lines)
	if parseErr != nil {
		return []FormatError{{
			Node:   nodeName,
			Rule:   "external_file_existence",
			Detail: fmt.Sprintf("invalid fragment line range %q for %q: %s", frag.Lines, filePath, parseErr.Error()),
		}}
	}

	// Reopen the file for this fragment.
	r, openErr := filereader.OpenFileReader(filePath)
	if openErr != nil {
		return []FormatError{{
			Node:   nodeName,
			Rule:   "external_file_existence",
			Detail: fmt.Sprintf("external file not found: %q", filePath),
		}}
	}
	defer r.Close()

	// Skip to (start - 1) lines to position at the first line of the range.
	// Lines are 1-based; skip (start-1) lines to reach line `start`.
	if start > 1 {
		r.SkipLines(start - 1)
	}

	// Read (end - start + 1) lines.
	count := end - start + 1
	var sb strings.Builder
	for i := 0; i < count; i++ {
		line, readErr := r.ReadLine()
		if readErr != nil {
			if errors.Is(readErr, filereader.ErrEndOfFile) {
				// Fewer lines than expected — still hash what we have.
				break
			}
			// Unexpected read error.
			return []FormatError{{
				Node:   nodeName,
				Rule:   "external_file_existence",
				Detail: fmt.Sprintf("error reading fragment from %q lines %s: %s", filePath, frag.Lines, readErr.Error()),
			}}
		}
		sb.WriteString(line)
		if i < count-1 {
			sb.WriteString("\n")
		}
	}

	// Normalize CRLF to LF (ReadLine already strips terminators, but the
	// content of lines themselves may contain embedded \r — normalize just in
	// case).
	content := strings.ReplaceAll(sb.String(), "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	// Compute SHA-1 and encode as base64url (no padding, 27 chars).
	h := sha1.New()
	_, _ = io.WriteString(h, content)
	computed := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if computed != frag.Hash {
		return []FormatError{{
			Node:   nodeName,
			Rule:   "external_file_existence",
			Detail: fmt.Sprintf("fragment hash mismatch for %q lines %s: expected %q, got %q", filePath, frag.Lines, frag.Hash, computed),
		}}
	}

	return nil
}

// parseLineRange parses a "start-end" string into two integers.
// Both start and end are 1-based and inclusive.
func parseLineRange(s string) (start, end int, err error) {
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected format \"start-end\", got %q", s)
	}
	start, err = strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start line %q: %w", parts[0], err)
	}
	end, err = strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end line %q: %w", parts[1], err)
	}
	if start < 1 || end < start {
		return 0, 0, fmt.Errorf("invalid range %q: start must be >= 1 and end must be >= start", s)
	}
	return start, end, nil
}

// ruleOutputPathValidation validates each output path on a leaf node using
// ValidatePath.
func ruleOutputPathValidation(node nodediscovery.DiscoveredNode, fm *frontmatter.Frontmatter) []FormatError {
	var errs []FormatError

	projectRoot, wdErr := os.Getwd()
	if wdErr != nil {
		return errs
	}

	for _, out := range fm.Outputs {
		if err := pathvalidation.ValidatePath(out.Path, projectRoot); err != nil {
			errs = append(errs, FormatError{
				Node:   node.LogicalName,
				Rule:   "output_path_validation",
				Detail: fmt.Sprintf("invalid output path %q: %s", out.Path, err.Error()),
			})
		}
	}

	return errs
}

// ruleDuplicatePublicSubsections checks that no two ## headings within
// # Public normalize to the same value.
func ruleDuplicatePublicSubsections(node nodediscovery.DiscoveredNode, parsed *parsenode.ParsedNode) []FormatError {
	var errs []FormatError

	// parsed.Public is guaranteed non-nil by the caller.
	seen := make(map[string]bool)

	for _, sub := range parsed.Public.Subsections {
		normalized := normalizename.NormalizeName(sub.Heading)
		if seen[normalized] {
			errs = append(errs, FormatError{
				Node:   node.LogicalName,
				Rule:   "duplicate_public_subsections",
				Detail: fmt.Sprintf("duplicate \"##\" heading in \"# Public\": %q (normalized: %q)", sub.Heading, normalized),
			})
		} else {
			seen[normalized] = true
		}
	}

	return errs
}
