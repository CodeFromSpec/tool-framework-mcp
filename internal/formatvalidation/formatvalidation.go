// code-from-spec: ROOT/golang/internal/format_validation/code@PENDING
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
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/nodediscovery"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/normalizename"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathvalidation"
)

// FormatError describes a single validation failure.
type FormatError struct {
	Node   string
	Rule   string
	Detail string
}

// ErrUnreadableNode is returned when a node file cannot be read.
var ErrUnreadableNode = errors.New("unreadable node")

// ValidateFormat takes a list of discovered nodes and validates each one
// against structural rules. Returns a slice of format errors (empty if all
// nodes are valid).
func ValidateFormat(discoveredNodes []nodediscovery.DiscoveredNode) ([]FormatError, error) {
	var errs []FormatError

	// Build set of all known logical names.
	knownNames := make(map[string]bool, len(discoveredNodes))
	for _, n := range discoveredNodes {
		knownNames[n.LogicalName] = true
	}

	// Determine which nodes have children.
	hasChildren := make(map[string]bool, len(discoveredNodes))
	for _, n := range discoveredNodes {
		for _, other := range discoveredNodes {
			if strings.HasPrefix(other.LogicalName, n.LogicalName+"/") {
				hasChildren[n.LogicalName] = true
				break
			}
		}
	}

	for _, node := range discoveredNodes {
		// Parse frontmatter.
		fm, fmErr := frontmatter.ParseFrontmatter(node.FilePath)
		if fmErr != nil {
			errs = append(errs, FormatError{
				Node:   node.LogicalName,
				Rule:   "unreadable node",
				Detail: fmErr.Error(),
			})
			continue
		}

		// Parse body.
		parsed, parseErr := parsenode.ParseNode(node.LogicalName)
		if parseErr != nil {
			errs = append(errs, FormatError{
				Node:   node.LogicalName,
				Rule:   "unreadable node",
				Detail: parseErr.Error(),
			})
			continue
		}

		isParent := hasChildren[node.LogicalName]

		// Rule: Name verification.
		headingNorm := normalizename.NormalizeName(parsed.NameSection.Heading)
		nameNorm := normalizename.NormalizeName(node.LogicalName)
		if headingNorm != nameNorm {
			errs = append(errs, FormatError{
				Node:   node.LogicalName,
				Rule:   "name verification",
				Detail: fmt.Sprintf("heading %q does not match logical name %q", parsed.NameSection.Heading, node.LogicalName),
			})
		}

		// Rule: Frontmatter field restrictions.
		if isParent {
			if len(fm.DependsOn) > 0 {
				errs = append(errs, FormatError{
					Node:   node.LogicalName,
					Rule:   "frontmatter field restrictions",
					Detail: "depends_on not allowed on nodes with children",
				})
			}
			if len(fm.External) > 0 {
				errs = append(errs, FormatError{
					Node:   node.LogicalName,
					Rule:   "frontmatter field restrictions",
					Detail: "external not allowed on nodes with children",
				})
			}
			if fm.Input != "" {
				errs = append(errs, FormatError{
					Node:   node.LogicalName,
					Rule:   "frontmatter field restrictions",
					Detail: "input not allowed on nodes with children",
				})
			}
			if len(fm.Outputs) > 0 {
				errs = append(errs, FormatError{
					Node:   node.LogicalName,
					Rule:   "frontmatter field restrictions",
					Detail: "outputs not allowed on nodes with children",
				})
			}
		}

		// Rule: Agent section restrictions.
		if isParent && parsed.Agent != nil {
			errs = append(errs, FormatError{
				Node:   node.LogicalName,
				Rule:   "agent section restrictions",
				Detail: "agent section not allowed on nodes with children",
			})
		}

		// Rule: Dependency targets.
		for _, dep := range fm.DependsOn {
			// Check if the target can be resolved.
			_, resolveOk := logicalnames.PathFromLogicalName(dep)
			if !resolveOk || !knownNames[dep] {
				errs = append(errs, FormatError{
					Node:   node.LogicalName,
					Rule:   "dependency targets",
					Detail: fmt.Sprintf("unresolvable: %s", dep),
				})
				continue
			}
			// Check ancestor dependency.
			if strings.HasPrefix(node.LogicalName, dep+"/") {
				errs = append(errs, FormatError{
					Node:   node.LogicalName,
					Rule:   "dependency targets",
					Detail: fmt.Sprintf("ancestor dependency: %s", dep),
				})
			}
			// Check descendant dependency.
			if strings.HasPrefix(dep, node.LogicalName+"/") {
				errs = append(errs, FormatError{
					Node:   node.LogicalName,
					Rule:   "dependency targets",
					Detail: fmt.Sprintf("descendant dependency: %s", dep),
				})
			}
		}

		// Rule: External file existence.
		for _, ext := range fm.External {
			if _, err := os.Stat(ext.Path); os.IsNotExist(err) {
				errs = append(errs, FormatError{
					Node:   node.LogicalName,
					Rule:   "external file existence",
					Detail: fmt.Sprintf("file not found: %s", ext.Path),
				})
				continue
			}
			// Check fragment hashes.
			for _, frag := range ext.Fragments {
				if frag.Hash == "" || frag.Lines == "" {
					continue
				}
				hashErr := checkFragmentHash(ext.Path, frag.Lines, frag.Hash)
				if hashErr != nil {
					errs = append(errs, FormatError{
						Node:   node.LogicalName,
						Rule:   "external file existence",
						Detail: fmt.Sprintf("hash mismatch: %s %s", ext.Path, hashErr),
					})
				}
			}
		}

		// Rule: Output path validation.
		for _, out := range fm.Outputs {
			if err := pathvalidation.ValidatePath(out.Path, "."); err != nil {
				errs = append(errs, FormatError{
					Node:   node.LogicalName,
					Rule:   "output path validation",
					Detail: err.Error(),
				})
			}
		}

		// Rule: Duplicate public subsections.
		if parsed.Public != nil {
			seen := make(map[string]bool)
			for _, sub := range parsed.Public.Subsections {
				norm := normalizename.NormalizeName(sub.Heading)
				if seen[norm] {
					errs = append(errs, FormatError{
						Node:   node.LogicalName,
						Rule:   "duplicate public subsections",
						Detail: fmt.Sprintf("duplicate heading: %s", sub.Heading),
					})
				}
				seen[norm] = true
			}
		}
	}

	return errs, nil
}

// checkFragmentHash reads the specified line range from a file and compares
// the SHA-1 hash (base64url, 27 chars) with the expected hash.
func checkFragmentHash(filePath, lines, expectedHash string) error {
	reader, err := filereader.OpenFileReader(filePath)
	if err != nil {
		return fmt.Errorf("cannot read: %s", err)
	}

	// Parse line range "start-end".
	parts := strings.SplitN(lines, "-", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid line range: %s", lines)
	}
	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return fmt.Errorf("invalid start line: %s", parts[0])
	}
	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return fmt.Errorf("invalid end line: %s", parts[1])
	}

	// Skip to start line.
	reader.SkipLines(start - 1)

	// Read lines from start to end.
	var extracted []string
	for i := start; i <= end; i++ {
		line, readErr := reader.ReadLine()
		if errors.Is(readErr, filereader.ErrEndOfFile) {
			break
		}
		extracted = append(extracted, line)
	}

	content := strings.Join(extracted, "\n")
	hash := sha1.Sum([]byte(content))
	encoded := base64.RawURLEncoding.EncodeToString(hash[:])
	if len(encoded) > 27 {
		encoded = encoded[:27]
	}

	if encoded != expectedHash {
		return fmt.Errorf("expected %s, got %s", expectedHash, encoded)
	}
	return nil
}
