// code-from-spec: ROOT/golang/implementation/spec_tree/validate@WpyfKHJM6sNWgb3RHIMvuhvfknY
package spectreevalidate

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
)

// SpecTreeValidateInput holds a single discovered node's logical name along
// with its parsed frontmatter and parsed node body, as produced by the
// frontmatter and parsenode packages respectively.
type SpecTreeValidateInput struct {
	LogicalName string
	Frontmatter *frontmatter.Frontmatter
	Node        *parsenode.Node
}

// FormatError describes a single format violation found in a node.
// Node is the logical name of the offending node. Rule identifies
// which validation rule was violated. Detail provides a human-readable
// explanation of the specific violation.
type FormatError struct {
	Node   string
	Rule   string
	Detail string
}

// logicalNameStripQualifier removes a parenthetical qualifier from a logical
// name, e.g. "ROOT/foo(bar)" -> "ROOT/foo".
func logicalNameStripQualifier(name string) string {
	if idx := strings.Index(name, "("); idx != -1 {
		return name[:idx]
	}
	return name
}

// SpecTreeValidate validates the full set of discovered nodes against the
// spec tree format rules. It accepts all entries at once so that cross-node
// rules (such as parent/child/leaf relationships) can be evaluated.
//
// A node is considered to have children if any other entry in the input list
// has a logical name that starts with the node's logical name followed by "/".
// A node is a leaf if no entry has a logical name that starts with the node's
// logical name followed by "/".
//
// Returns a list of FormatError values describing every violation found across
// all entries. Returns an empty list if all nodes are valid.
func SpecTreeValidate(entries []*SpecTreeValidateInput) []*FormatError {
	var errs []*FormatError

	// Step 1: Build the known names set.
	knownNames := make(map[string]bool)
	for _, entry := range entries {
		knownNames[entry.LogicalName] = true
		if len(entry.Frontmatter.Outputs) > 0 {
			bareSuffix := strings.TrimPrefix(entry.LogicalName, "ROOT/")
			for _, output := range entry.Frontmatter.Outputs {
				artifactName := "ARTIFACT/" + bareSuffix + "(" + output.ID + ")"
				knownNames[artifactName] = true
			}
		}
	}

	// Step 3: Validate each entry.
	for _, entry := range entries {
		// Step 3a: Determine if the entry has children.
		hasChildren := false
		prefix := entry.LogicalName + "/"
		for _, other := range entries {
			if strings.HasPrefix(other.LogicalName, prefix) {
				hasChildren = true
				break
			}
		}

		// Step 3b: Rule name_heading.
		normalizedLogicalName := textnormalization.NormalizeText(entry.LogicalName)
		if normalizedLogicalName != entry.Node.NameSection.Heading {
			errs = append(errs, &FormatError{
				Node:   entry.LogicalName,
				Rule:   "name_heading",
				Detail: fmt.Sprintf("heading %s does not match logical name %s", entry.Node.NameSection.Heading, entry.LogicalName),
			})
		}

		// Step 3c: Rule leaf_only_fields.
		if hasChildren {
			if len(entry.Frontmatter.DependsOn) > 0 {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "field depends_on is not permitted on non-leaf nodes",
				})
			}
			if len(entry.Frontmatter.External) > 0 {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "field external is not permitted on non-leaf nodes",
				})
			}
			if entry.Frontmatter.Input != "" {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "field input is not permitted on non-leaf nodes",
				})
			}
			if len(entry.Frontmatter.Outputs) > 0 {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "field outputs is not permitted on non-leaf nodes",
				})
			}
		}

		// Step 3d: Rule leaf_only_agent.
		if hasChildren && entry.Node.Agent != nil {
			errs = append(errs, &FormatError{
				Node:   entry.LogicalName,
				Rule:   "leaf_only_agent",
				Detail: "# Agent section is not permitted on non-leaf nodes",
			})
		}

		// Step 3e: Rule dependency_targets.
		for _, dep := range entry.Frontmatter.DependsOn {
			if strings.HasPrefix(dep, "ROOT/") {
				bare := logicalNameStripQualifier(dep)
				if !knownNames[bare] {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "dependency_targets",
						Detail: fmt.Sprintf("depends_on target %s does not exist", dep),
					})
				} else if bare == entry.LogicalName {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "dependency_targets",
						Detail: fmt.Sprintf("depends_on target %s refers to the node itself", dep),
					})
				} else if strings.HasPrefix(entry.LogicalName, bare+"/") {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "dependency_targets",
						Detail: fmt.Sprintf("depends_on target %s is an ancestor of this node", dep),
					})
				} else if strings.HasPrefix(bare, entry.LogicalName+"/") {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "dependency_targets",
						Detail: fmt.Sprintf("depends_on target %s is a descendant of this node", dep),
					})
				}
			} else if strings.HasPrefix(dep, "ARTIFACT/") {
				if !knownNames[dep] {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "dependency_targets",
						Detail: fmt.Sprintf("depends_on target %s does not exist", dep),
					})
				}
			} else {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "dependency_targets",
					Detail: fmt.Sprintf("depends_on entry %s is not a valid ROOT/ or ARTIFACT/ reference", dep),
				})
			}
		}

		// Step 3f: Rule input_target.
		if entry.Frontmatter.Input != "" {
			if !strings.HasPrefix(entry.Frontmatter.Input, "ARTIFACT/") {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "input_target",
					Detail: fmt.Sprintf("input %s must be an ARTIFACT/ reference", entry.Frontmatter.Input),
				})
			} else if !knownNames[entry.Frontmatter.Input] {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "input_target",
					Detail: fmt.Sprintf("input target %s does not exist", entry.Frontmatter.Input),
				})
			}
		}

		// Step 3g: Rule external_files.
		for _, ext := range entry.Frontmatter.External {
			cfsPath := &pathutils.PathCfs{Value: ext.Path}

			// Step 1: Verify existence.
			reader, err := filereader.FileOpen(cfsPath)
			if err != nil {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "external_files",
					Detail: fmt.Sprintf("external file %s cannot be opened: %v", ext.Path, err),
				})
				continue
			}
			filereader.FileClose(reader)

			// Step 2: Verify fragments.
			if len(ext.Fragments) == 0 {
				continue
			}

			for _, fragment := range ext.Fragments {
				// Parse "start-end" line range.
				start, end, parseErr := parseLineRange(fragment.Lines)
				if parseErr != nil || start < 1 || start > end {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "external_files",
						Detail: fmt.Sprintf("external file %s fragment has invalid lines field: %s", ext.Path, fragment.Lines),
					})
					continue
				}

				fragReader, openErr := filereader.FileOpen(cfsPath)
				if openErr != nil {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "external_files",
						Detail: fmt.Sprintf("external file %s cannot be opened for fragment verification", ext.Path),
					})
					continue
				}

				filereader.FileSkipLines(fragReader, start-1)

				var contentBuilder strings.Builder
				linesToRead := end - start + 1
				readOk := true

				for i := 0; i < linesToRead; i++ {
					line, readErr := filereader.FileReadLine(fragReader)
					if readErr != nil {
						if errors.Is(readErr, filereader.ErrEndOfFile) {
							filereader.FileClose(fragReader)
							errs = append(errs, &FormatError{
								Node:   entry.LogicalName,
								Rule:   "external_files",
								Detail: fmt.Sprintf("external file %s fragment lines %s is out of range", ext.Path, fragment.Lines),
							})
							readOk = false
							break
						}
						filereader.FileClose(fragReader)
						errs = append(errs, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "external_files",
							Detail: fmt.Sprintf("external file %s fragment lines %s could not be read: %v", ext.Path, fragment.Lines, readErr),
						})
						readOk = false
						break
					}
					contentBuilder.WriteString(line)
					contentBuilder.WriteString("\n")
				}

				if !readOk {
					continue
				}

				filereader.FileClose(fragReader)

				content := contentBuilder.String()
				hash := sha1.Sum([]byte(content))
				computed := base64.RawURLEncoding.EncodeToString(hash[:])

				if computed != fragment.Hash {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "external_files",
						Detail: fmt.Sprintf("external file %s fragment lines %s hash mismatch: expected %s, got %s", ext.Path, fragment.Lines, fragment.Hash, computed),
					})
				}
			}
		}

		// Step 3h: Rule output_paths.
		for _, output := range entry.Frontmatter.Outputs {
			if err := pathutils.PathValidateCfs(output.Path); err != nil {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "output_paths",
					Detail: fmt.Sprintf("output path %s is invalid: %v", output.Path, err),
				})
			}
		}

		// Step 3i: Rule duplicate_subsections.
		if entry.Node.Public == nil || len(entry.Node.Public.Subsections) == 0 {
			continue
		}

		seenHeadings := make(map[string]bool)
		for _, subsection := range entry.Node.Public.Subsections {
			if seenHeadings[subsection.Heading] {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "duplicate_subsections",
					Detail: fmt.Sprintf("duplicate subsection heading %s in # Public", subsection.Heading),
				})
			} else {
				seenHeadings[subsection.Heading] = true
			}
		}
	}

	if errs == nil {
		return []*FormatError{}
	}
	return errs
}

// parseLineRange parses a "start-end" string into two integers.
// Returns an error if the format is invalid.
func parseLineRange(lines string) (int, int, error) {
	parts := strings.SplitN(lines, "-", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid line range format: %q", lines)
	}
	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start line: %w", err)
	}
	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end line: %w", err)
	}
	return start, end, nil
}
