// code-from-spec: ROOT/golang/implementation/spec_tree/validate@CbThEj8DzS37hMYf5Bje-cugDKs

package spectreevalidate

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/textnormalization"
)

// SpecTreeValidateInput represents a single discovered node with its parsed
// frontmatter and body, used as input to SpecTreeValidate.
type SpecTreeValidateInput struct {
	LogicalName string
	Frontmatter *frontmatter.Frontmatter
	Node        *parsenode.Node
}

// FormatError represents a single format violation found during validation.
// Node identifies the offending logical name, Rule identifies which
// validation rule was violated, and Detail provides a human-readable
// explanation.
type FormatError struct {
	Node   string
	Rule   string
	Detail string
}

// SpecTreeValidate validates the full set of discovered nodes against the
// spec tree format rules. It returns a list of FormatErrors describing every
// violation found. The returned slice is empty when all nodes are valid.
//
// A node is considered to have children if any other entry in the input list
// has a logical name that starts with the node's logical name followed by
// "/". A node is considered a leaf if no other entry starts with its logical
// name followed by "/".
func SpecTreeValidate(entries []*SpecTreeValidateInput) ([]*FormatError, error) {
	// Step 1 — Build the known logical names set.
	knownNames := make(map[string]struct{})
	for _, entry := range entries {
		knownNames[entry.LogicalName] = struct{}{}
	}
	for _, entry := range entries {
		if len(entry.Frontmatter.Outputs) > 0 {
			barePath := strings.TrimPrefix(entry.LogicalName, "ROOT/")
			for _, output := range entry.Frontmatter.Outputs {
				artifactName := "ARTIFACT/" + barePath + "(" + output.ID + ")"
				knownNames[artifactName] = struct{}{}
			}
		}
	}

	// Step 2 — Collect errors across all entries.
	var errs []*FormatError
	for _, entry := range entries {
		hasChildren := false
		prefix := entry.LogicalName + "/"
		for _, other := range entries {
			if other.LogicalName != entry.LogicalName && strings.HasPrefix(other.LogicalName, prefix) {
				hasChildren = true
				break
			}
		}

		errs = append(errs, ruleNameHeading(entry)...)
		errs = append(errs, ruleLeafOnlyFields(entry, hasChildren)...)
		errs = append(errs, ruleLeafOnlyAgent(entry, hasChildren)...)
		errs = append(errs, ruleDependencyTargets(entry, knownNames)...)
		errs = append(errs, ruleInputTarget(entry, knownNames)...)
		errs = append(errs, ruleExternalFiles(entry)...)
		errs = append(errs, ruleOutputPaths(entry)...)
		errs = append(errs, ruleDuplicateSubsections(entry)...)
	}

	return errs, nil
}

func ruleNameHeading(entry *SpecTreeValidateInput) []*FormatError {
	if entry.Node == nil || entry.Node.NameSection == nil {
		return nil
	}
	normalizedHeading := textnormalization.NormalizeText(entry.Node.NameSection.Heading)
	normalizedName := textnormalization.NormalizeText(entry.LogicalName)
	if normalizedHeading != normalizedName {
		return []*FormatError{
			{
				Node:   entry.LogicalName,
				Rule:   "name_heading",
				Detail: fmt.Sprintf("first heading %s does not match logical name %s", normalizedHeading, normalizedName),
			},
		}
	}
	return nil
}

func ruleLeafOnlyFields(entry *SpecTreeValidateInput, hasChildren bool) []*FormatError {
	if !hasChildren {
		return nil
	}
	var errs []*FormatError
	if len(entry.Frontmatter.DependsOn) > 0 {
		errs = append(errs, &FormatError{
			Node:   entry.LogicalName,
			Rule:   "leaf_only_fields",
			Detail: "depends_on is only permitted on leaf nodes",
		})
	}
	if len(entry.Frontmatter.External) > 0 {
		errs = append(errs, &FormatError{
			Node:   entry.LogicalName,
			Rule:   "leaf_only_fields",
			Detail: "external is only permitted on leaf nodes",
		})
	}
	if entry.Frontmatter.Input != "" {
		errs = append(errs, &FormatError{
			Node:   entry.LogicalName,
			Rule:   "leaf_only_fields",
			Detail: "input is only permitted on leaf nodes",
		})
	}
	if len(entry.Frontmatter.Outputs) > 0 {
		errs = append(errs, &FormatError{
			Node:   entry.LogicalName,
			Rule:   "leaf_only_fields",
			Detail: "outputs is only permitted on leaf nodes",
		})
	}
	return errs
}

func ruleLeafOnlyAgent(entry *SpecTreeValidateInput, hasChildren bool) []*FormatError {
	if hasChildren && entry.Node != nil && entry.Node.Agent != nil {
		return []*FormatError{
			{
				Node:   entry.LogicalName,
				Rule:   "leaf_only_agent",
				Detail: "# Agent section is only permitted on leaf nodes",
			},
		}
	}
	return nil
}

// logicalNameStripQualifier removes a parenthetical qualifier from a logical
// name, e.g. "ROOT/a/b(qualifier)" becomes "ROOT/a/b".
func logicalNameStripQualifier(name string) string {
	if idx := strings.Index(name, "("); idx != -1 {
		return name[:idx]
	}
	return name
}

func ruleDependencyTargets(entry *SpecTreeValidateInput, knownNames map[string]struct{}) []*FormatError {
	var errs []*FormatError
	for _, depPtr := range entry.Frontmatter.DependsOn {
		if depPtr == nil {
			continue
		}
		dep := *depPtr

		if strings.HasPrefix(dep, "ROOT/") {
			bareName := logicalNameStripQualifier(dep)
			if _, known := knownNames[bareName]; !known {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "dependency_targets",
					Detail: fmt.Sprintf("depends_on entry %s references unknown node %s", dep, bareName),
				})
				continue
			}
			if bareName == entry.LogicalName {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "dependency_targets",
					Detail: fmt.Sprintf("depends_on entry %s points to the node itself", dep),
				})
				continue
			}
			if strings.HasPrefix(entry.LogicalName, bareName+"/") {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "dependency_targets",
					Detail: fmt.Sprintf("depends_on entry %s points to an ancestor", dep),
				})
				continue
			}
			if strings.HasPrefix(bareName, entry.LogicalName+"/") {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "dependency_targets",
					Detail: fmt.Sprintf("depends_on entry %s points to a descendant", dep),
				})
				continue
			}
		} else if strings.HasPrefix(dep, "ARTIFACT/") {
			if _, known := knownNames[dep]; !known {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "dependency_targets",
					Detail: fmt.Sprintf("depends_on entry %s references unknown artifact", dep),
				})
			}
		} else {
			errs = append(errs, &FormatError{
				Node:   entry.LogicalName,
				Rule:   "dependency_targets",
				Detail: fmt.Sprintf("depends_on entry %s has unrecognized prefix", dep),
			})
		}
	}
	return errs
}

func ruleInputTarget(entry *SpecTreeValidateInput, knownNames map[string]struct{}) []*FormatError {
	if entry.Frontmatter.Input == "" {
		return nil
	}
	if !strings.HasPrefix(entry.Frontmatter.Input, "ARTIFACT/") {
		return []*FormatError{
			{
				Node:   entry.LogicalName,
				Rule:   "input_target",
				Detail: fmt.Sprintf("input must be an ARTIFACT/ reference, got %s", entry.Frontmatter.Input),
			},
		}
	}
	if _, known := knownNames[entry.Frontmatter.Input]; !known {
		return []*FormatError{
			{
				Node:   entry.LogicalName,
				Rule:   "input_target",
				Detail: fmt.Sprintf("input references unknown artifact %s", entry.Frontmatter.Input),
			},
		}
	}
	return nil
}

func ruleExternalFiles(entry *SpecTreeValidateInput) []*FormatError {
	var errs []*FormatError
	for _, ext := range entry.Frontmatter.External {
		pathCfs := &pathutils.PathCfs{Value: ext.Path}

		// Step 1 — Verify existence.
		reader, err := filereader.FileOpen(pathCfs)
		if err != nil {
			errs = append(errs, &FormatError{
				Node:   entry.LogicalName,
				Rule:   "external_files",
				Detail: fmt.Sprintf("external file %s cannot be opened: %s", ext.Path, err.Error()),
			})
			continue
		}
		filereader.FileClose(reader)

		// Step 2 — Verify fragments.
		if len(ext.Fragments) == 0 {
			continue
		}
		for _, fragment := range ext.Fragments {
			start, end, parseErr := parseFragmentLines(fragment.Lines)
			if parseErr != nil {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "external_files",
					Detail: fmt.Sprintf("external file %s fragment has invalid lines format: %s", ext.Path, fragment.Lines),
				})
				continue
			}
			if start < 1 {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "external_files",
					Detail: fmt.Sprintf("external file %s fragment start line must be >= 1, got %d", ext.Path, start),
				})
				continue
			}
			if start > end {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "external_files",
					Detail: fmt.Sprintf("external file %s fragment start %d exceeds end %d", ext.Path, start, end),
				})
				continue
			}

			fragReader, openErr := filereader.FileOpen(pathCfs)
			if openErr != nil {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "external_files",
					Detail: fmt.Sprintf("external file %s cannot be opened for fragment read: %s", ext.Path, openErr.Error()),
				})
				continue
			}

			filereader.FileSkipLines(fragReader, start-1)

			lineCount := end - start + 1
			readLines := make([]string, 0, lineCount)
			outOfRange := false
			for i := 0; i < lineCount; i++ {
				line, readErr := filereader.FileReadLine(fragReader)
				if readErr != nil {
					if errors.Is(readErr, filereader.ErrEndOfFile) {
						filereader.FileClose(fragReader)
						errs = append(errs, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "external_files",
							Detail: fmt.Sprintf("external file %s fragment %s is out of range", ext.Path, fragment.Lines),
						})
						outOfRange = true
						break
					}
					filereader.FileClose(fragReader)
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "external_files",
						Detail: fmt.Sprintf("external file %s fragment %s read error: %s", ext.Path, fragment.Lines, readErr.Error()),
					})
					outOfRange = true
					break
				}
				readLines = append(readLines, line)
			}
			if outOfRange {
				continue
			}

			filereader.FileClose(fragReader)

			content := strings.Join(readLines, "\n")
			digest := sha1.Sum([]byte(content))
			computedHash := base64.RawURLEncoding.EncodeToString(digest[:])

			if computedHash != fragment.Hash {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "external_files",
					Detail: fmt.Sprintf("external file %s fragment %s hash mismatch: expected %s, got %s", ext.Path, fragment.Lines, fragment.Hash, computedHash),
				})
			}
		}
	}
	return errs
}

// parseFragmentLines parses a fragment lines string of the form "<start>-<end>"
// and returns start and end as integers. Returns an error if the format is invalid.
func parseFragmentLines(lines string) (int, int, error) {
	parts := strings.SplitN(lines, "-", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid lines format: %s", lines)
	}
	start, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start in lines format: %s", lines)
	}
	end, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end in lines format: %s", lines)
	}
	return start, end, nil
}

func ruleOutputPaths(entry *SpecTreeValidateInput) []*FormatError {
	var errs []*FormatError
	for _, output := range entry.Frontmatter.Outputs {
		if err := pathutils.PathValidateCfs(output.Path); err != nil {
			errs = append(errs, &FormatError{
				Node:   entry.LogicalName,
				Rule:   "output_paths",
				Detail: fmt.Sprintf("output path %s is invalid: %s", output.Path, err.Error()),
			})
		}
	}
	return errs
}

func ruleDuplicateSubsections(entry *SpecTreeValidateInput) []*FormatError {
	if entry.Node == nil || entry.Node.Public == nil {
		return nil
	}
	if len(entry.Node.Public.Subsections) == 0 {
		return nil
	}
	var errs []*FormatError
	seenHeadings := make(map[string]struct{})
	for _, subsection := range entry.Node.Public.Subsections {
		normalizedHeading := textnormalization.NormalizeText(subsection.Heading)
		if _, seen := seenHeadings[normalizedHeading]; seen {
			errs = append(errs, &FormatError{
				Node:   entry.LogicalName,
				Rule:   "duplicate_subsections",
				Detail: fmt.Sprintf("duplicate ## subsection heading %q in # Public section", subsection.Heading),
			})
		} else {
			seenHeadings[normalizedHeading] = struct{}{}
		}
	}
	return errs
}
