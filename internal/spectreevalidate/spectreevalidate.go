// code-from-spec: ROOT/golang/implementation/spec_tree/validate@1zPU3hH103F43uW3egM28fvqw3g

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

// SpecTreeValidateInput holds the data for a single spec tree node
// that is to be validated.
type SpecTreeValidateInput struct {
	// LogicalName is the full logical name of the node (e.g. "ROOT/a/b").
	LogicalName string

	// Frontmatter holds the parsed frontmatter for this node.
	Frontmatter *frontmatter.Frontmatter

	// Node holds the parsed node body for this node.
	Node *parsenode.Node
}

// FormatError describes a single format rule violation found during
// validation of a spec tree node.
type FormatError struct {
	// Node is the logical name of the node that violated the rule.
	Node string

	// Rule is the name or identifier of the rule that was violated.
	Rule string

	// Detail provides a human-readable explanation of the violation.
	Detail string
}

// SpecTreeValidate validates the full set of discovered nodes.
//
// It takes the complete list of nodes with their parsed frontmatter and
// body, and returns a list of FormatError values describing any format
// violations found. The returned slice is empty when all nodes are valid.
//
// A node is considered to have children if any other entry in the input
// list has a logical name that starts with the node's logical name
// followed by "/". For example, given entries "ROOT/a" and "ROOT/a/b",
// "ROOT/a" has children. A node is a leaf if no other entry's logical
// name starts with its own logical name followed by "/".
func SpecTreeValidate(entries []*SpecTreeValidateInput) []*FormatError {
	// Step 1: Build the known logical names set.
	knownNames := make(map[string]bool)
	for _, entry := range entries {
		knownNames[entry.LogicalName] = true
		if len(entry.Frontmatter.Outputs) > 0 {
			remainder := strings.TrimPrefix(entry.LogicalName, "ROOT/")
			for _, output := range entry.Frontmatter.Outputs {
				artifactName := "ARTIFACT/" + remainder + "(" + output.ID + ")"
				knownNames[artifactName] = true
			}
		}
	}

	// Step 2: Initialize errors list.
	var formatErrors []*FormatError

	// Step 3: For each entry, run all validation rules.
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

		// Rule: name_heading
		normalizedHeading := textnormalization.NormalizeText(entry.Node.NameSection.Heading)
		normalizedName := textnormalization.NormalizeText(entry.LogicalName)
		if normalizedHeading != normalizedName {
			formatErrors = append(formatErrors, &FormatError{
				Node:   entry.LogicalName,
				Rule:   "name_heading",
				Detail: fmt.Sprintf("first section heading %s does not match logical name %s", normalizedHeading, normalizedName),
			})
		}

		// Rule: leaf_only_fields
		if hasChildren {
			if len(entry.Frontmatter.DependsOn) > 0 {
				formatErrors = append(formatErrors, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "depends_on is only permitted on leaf nodes",
				})
			}
			if len(entry.Frontmatter.External) > 0 {
				formatErrors = append(formatErrors, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "external is only permitted on leaf nodes",
				})
			}
			if entry.Frontmatter.Input != "" {
				formatErrors = append(formatErrors, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "input is only permitted on leaf nodes",
				})
			}
			if len(entry.Frontmatter.Outputs) > 0 {
				formatErrors = append(formatErrors, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "outputs is only permitted on leaf nodes",
				})
			}
		}

		// Rule: leaf_only_agent
		if hasChildren && entry.Node.Agent != nil {
			formatErrors = append(formatErrors, &FormatError{
				Node:   entry.LogicalName,
				Rule:   "leaf_only_agent",
				Detail: "# Agent section is only permitted on leaf nodes",
			})
		}

		// Rule: dependency_targets
		for _, dep := range entry.Frontmatter.DependsOn {
			if strings.HasPrefix(dep, "ROOT/") {
				bare := logicalNameStripQualifier(dep)
				if !knownNames[bare] {
					formatErrors = append(formatErrors, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "dependency_targets",
						Detail: fmt.Sprintf("depends_on references unknown node %s", dep),
					})
				} else if bare == entry.LogicalName {
					formatErrors = append(formatErrors, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "dependency_targets",
						Detail: fmt.Sprintf("depends_on references the node itself: %s", dep),
					})
				} else if strings.HasPrefix(entry.LogicalName, bare+"/") {
					formatErrors = append(formatErrors, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "dependency_targets",
						Detail: fmt.Sprintf("depends_on references an ancestor: %s", dep),
					})
				} else if strings.HasPrefix(bare, entry.LogicalName+"/") {
					formatErrors = append(formatErrors, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "dependency_targets",
						Detail: fmt.Sprintf("depends_on references a descendant: %s", dep),
					})
				}
			} else if strings.HasPrefix(dep, "ARTIFACT/") {
				if !knownNames[dep] {
					formatErrors = append(formatErrors, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "dependency_targets",
						Detail: fmt.Sprintf("depends_on references unknown artifact %s", dep),
					})
				}
			} else {
				formatErrors = append(formatErrors, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "dependency_targets",
					Detail: fmt.Sprintf("depends_on entry has unrecognized prefix: %s", dep),
				})
			}
		}

		// Rule: input_target
		if entry.Frontmatter.Input != "" {
			if !strings.HasPrefix(entry.Frontmatter.Input, "ARTIFACT/") {
				formatErrors = append(formatErrors, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "input_target",
					Detail: fmt.Sprintf("input must be an ARTIFACT/ reference, got: %s", entry.Frontmatter.Input),
				})
			} else {
				if !knownNames[entry.Frontmatter.Input] {
					formatErrors = append(formatErrors, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "input_target",
						Detail: fmt.Sprintf("input references unknown artifact: %s", entry.Frontmatter.Input),
					})
				}
			}
		}

		// Rule: external_files
		for _, ext := range entry.Frontmatter.External {
			cfsPath := &pathutils.PathCfs{Value: ext.Path}

			// Step 1 — Verify existence.
			reader, err := filereader.FileOpen(cfsPath)
			if err != nil {
				formatErrors = append(formatErrors, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "external_files",
					Detail: fmt.Sprintf("external file cannot be opened: %s", ext.Path),
				})
				continue
			}
			filereader.FileClose(reader)

			// Step 2 — Verify fragments.
			if len(ext.Fragments) > 0 {
				for _, fragment := range ext.Fragments {
					// Parse fragment.lines as "<start>-<end>".
					start, end, parseErr := parseLineRange(fragment.Lines)
					if parseErr != nil {
						formatErrors = append(formatErrors, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "external_files",
							Detail: fmt.Sprintf("invalid lines format in fragment for %s: %s", ext.Path, fragment.Lines),
						})
						continue
					}
					if start < 1 || start > end {
						formatErrors = append(formatErrors, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "external_files",
							Detail: fmt.Sprintf("invalid line range in fragment for %s: %s", ext.Path, fragment.Lines),
						})
						continue
					}

					fragReader, openErr := filereader.FileOpen(cfsPath)
					if openErr != nil {
						formatErrors = append(formatErrors, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "external_files",
							Detail: fmt.Sprintf("external file cannot be opened for fragment read: %s", ext.Path),
						})
						continue
					}

					filereader.FileSkipLines(fragReader, start-1)
					lineCount := end - start + 1
					var contentBuilder strings.Builder
					readOk := true
					for i := 0; i < lineCount; i++ {
						line, readErr := filereader.FileReadLine(fragReader)
						if readErr != nil {
							if errors.Is(readErr, filereader.ErrEndOfFile) {
								filereader.FileClose(fragReader)
								formatErrors = append(formatErrors, &FormatError{
									Node:   entry.LogicalName,
									Rule:   "external_files",
									Detail: fmt.Sprintf("fragment out of range for %s: %s", ext.Path, fragment.Lines),
								})
								readOk = false
								break
							}
							filereader.FileClose(fragReader)
							formatErrors = append(formatErrors, &FormatError{
								Node:   entry.LogicalName,
								Rule:   "external_files",
								Detail: fmt.Sprintf("error reading fragment for %s: %s", ext.Path, fragment.Lines),
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
					digest := sha1.Sum([]byte(content))
					computedHash := base64.RawURLEncoding.EncodeToString(digest[:])
					if computedHash != fragment.Hash {
						formatErrors = append(formatErrors, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "external_files",
							Detail: fmt.Sprintf("fragment hash mismatch for %s lines %s: expected %s, got %s", ext.Path, fragment.Lines, fragment.Hash, computedHash),
						})
					}
				}
			}
		}

		// Rule: output_paths
		for _, output := range entry.Frontmatter.Outputs {
			if err := pathutils.PathValidateCfs(output.Path); err != nil {
				formatErrors = append(formatErrors, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "output_paths",
					Detail: fmt.Sprintf("invalid output path %s: %s", output.Path, err.Error()),
				})
			}
		}

		// Rule: duplicate_subsections
		if entry.Node.Public != nil {
			seenHeadings := make(map[string]bool)
			for _, subsection := range entry.Node.Public.Subsections {
				normalized := textnormalization.NormalizeText(subsection.Heading)
				if seenHeadings[normalized] {
					formatErrors = append(formatErrors, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "duplicate_subsections",
						Detail: fmt.Sprintf("duplicate ## subsection heading in # Public: %s", subsection.RawHeading),
					})
				} else {
					seenHeadings[normalized] = true
				}
			}
		}
	}

	// Step 4: Return errors.
	return formatErrors
}

// logicalNameStripQualifier removes a parenthetical qualifier from the
// end of a logical name, if present. For example, "ROOT/a/b(v2)" becomes
// "ROOT/a/b".
func logicalNameStripQualifier(name string) string {
	idx := strings.LastIndex(name, "(")
	if idx == -1 {
		return name
	}
	// Ensure the opening paren has a corresponding closing paren at the end.
	if strings.HasSuffix(name, ")") {
		return name[:idx]
	}
	return name
}

// parseLineRange parses a string in the format "<start>-<end>" and returns
// the start and end line numbers as integers. Returns an error if the format
// is invalid.
func parseLineRange(lines string) (int, int, error) {
	parts := strings.SplitN(lines, "-", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid line range format: %s", lines)
	}
	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start line: %s", parts[0])
	}
	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end line: %s", parts[1])
	}
	return start, end, nil
}
