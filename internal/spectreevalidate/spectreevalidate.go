// code-from-spec: ROOT/golang/implementation/spec_tree/validate@DaBNkk9IuTfnRo4i_2QNWFNQSnk
package spectreevalidate

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
)

// SpecTreeValidateInput holds a single discovered node with its parsed
// frontmatter and parsed node structure, as input to SpecTreeValidate.
type SpecTreeValidateInput struct {
	// LogicalName is the logical name of the node (e.g. "ROOT/foo/bar").
	LogicalName string

	// Frontmatter is the parsed frontmatter of the node file.
	Frontmatter *frontmatter.Frontmatter

	// Node is the parsed node structure.
	Node *parsenode.Node
}

// FormatError describes a single validation failure for a node.
type FormatError struct {
	// Node is the logical name of the node that failed validation.
	Node string

	// Rule is the name of the rule that was violated.
	Rule string

	// Detail provides additional context about the violation.
	Detail string
}

// SpecTreeValidate validates the full set of discovered nodes.
//
// A node has children if any other entry in the input list has a logical name
// that starts with the node's logical name followed by "/". For example, given
// entries "ROOT/a" and "ROOT/a/b", "ROOT/a" has children. "ROOT/a/b" is a leaf
// if no entry starts with "ROOT/a/b/".
//
// Returns a list of FormatErrors describing all violations found. Returns an
// empty slice when all nodes are valid.
func SpecTreeValidate(entries []*SpecTreeValidateInput) []*FormatError {
	// Step 1: Build the known logical names set.
	knownNames := make(map[string]struct{})
	for _, entry := range entries {
		knownNames[entry.LogicalName] = struct{}{}
		if len(entry.Frontmatter.Outputs) > 0 {
			suffix := strings.TrimPrefix(entry.LogicalName, "ROOT/")
			for _, output := range entry.Frontmatter.Outputs {
				artifactName := "ARTIFACT/" + suffix + "(" + output.ID + ")"
				knownNames[artifactName] = struct{}{}
			}
		}
	}

	// Step 2: Initialize errors as an empty list.
	var errors []*FormatError

	// Step 3: For each entry, determine children and run all validation rules.
	for _, entry := range entries {
		hasChildren := false
		for _, other := range entries {
			if other.LogicalName != entry.LogicalName &&
				strings.HasPrefix(other.LogicalName, entry.LogicalName+"/") {
				hasChildren = true
				break
			}
		}

		validateNameHeading(entry, &errors)
		validateLeafOnlyFields(entry, hasChildren, &errors)
		validateLeafOnlyAgent(entry, hasChildren, &errors)
		validateDependencyTargets(entry, knownNames, &errors)
		validateInputTarget(entry, knownNames, &errors)
		validateExternalFiles(entry, &errors)
		validateOutputPaths(entry, &errors)
		validateDuplicateSubsections(entry, &errors)
	}

	// Step 4: Return errors.
	return errors
}

// logicalNameStripQualifier strips a parenthetical qualifier from a logical name.
// e.g. "ROOT/foo/bar(baz)" -> "ROOT/foo/bar"
func logicalNameStripQualifier(name string) string {
	if idx := strings.Index(name, "("); idx >= 0 {
		return name[:idx]
	}
	return name
}

func validateNameHeading(entry *SpecTreeValidateInput, errors *[]*FormatError) {
	normalizedHeading := textnormalization.NormalizeText(entry.Node.NameSection.Heading)
	normalizedName := textnormalization.NormalizeText(entry.LogicalName)
	if normalizedHeading != normalizedName {
		*errors = append(*errors, &FormatError{
			Node:   entry.LogicalName,
			Rule:   "name_heading",
			Detail: fmt.Sprintf("name section heading %s does not match logical name %s", normalizedHeading, normalizedName),
		})
	}
}

func validateLeafOnlyFields(entry *SpecTreeValidateInput, hasChildren bool, errors *[]*FormatError) {
	if !hasChildren {
		return
	}

	if len(entry.Frontmatter.DependsOn) > 0 {
		*errors = append(*errors, &FormatError{
			Node:   entry.LogicalName,
			Rule:   "leaf_only_fields",
			Detail: "non-leaf node has depends_on",
		})
	}

	if len(entry.Frontmatter.External) > 0 {
		*errors = append(*errors, &FormatError{
			Node:   entry.LogicalName,
			Rule:   "leaf_only_fields",
			Detail: "non-leaf node has external",
		})
	}

	if entry.Frontmatter.Input != "" {
		*errors = append(*errors, &FormatError{
			Node:   entry.LogicalName,
			Rule:   "leaf_only_fields",
			Detail: "non-leaf node has input",
		})
	}

	if len(entry.Frontmatter.Outputs) > 0 {
		*errors = append(*errors, &FormatError{
			Node:   entry.LogicalName,
			Rule:   "leaf_only_fields",
			Detail: "non-leaf node has outputs",
		})
	}
}

func validateLeafOnlyAgent(entry *SpecTreeValidateInput, hasChildren bool, errors *[]*FormatError) {
	if !hasChildren {
		return
	}

	if entry.Node.Agent != nil {
		*errors = append(*errors, &FormatError{
			Node:   entry.LogicalName,
			Rule:   "leaf_only_agent",
			Detail: "non-leaf node has an Agent section",
		})
	}
}

func validateDependencyTargets(entry *SpecTreeValidateInput, knownNames map[string]struct{}, errors *[]*FormatError) {
	for _, dep := range entry.Frontmatter.DependsOn {
		if strings.HasPrefix(dep, "ROOT/") {
			bareName := logicalNameStripQualifier(dep)

			if _, exists := knownNames[bareName]; !exists {
				*errors = append(*errors, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "dependency_targets",
					Detail: fmt.Sprintf("depends_on target %s does not exist", dep),
				})
				continue
			}

			if bareName == entry.LogicalName {
				*errors = append(*errors, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "dependency_targets",
					Detail: fmt.Sprintf("depends_on target %s points to the node itself", dep),
				})
				continue
			}

			if strings.HasPrefix(entry.LogicalName, bareName+"/") {
				*errors = append(*errors, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "dependency_targets",
					Detail: fmt.Sprintf("depends_on target %s is an ancestor of this node", dep),
				})
				continue
			}

			if strings.HasPrefix(bareName, entry.LogicalName+"/") {
				*errors = append(*errors, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "dependency_targets",
					Detail: fmt.Sprintf("depends_on target %s is a descendant of this node", dep),
				})
				continue
			}
		} else if strings.HasPrefix(dep, "ARTIFACT/") {
			if _, exists := knownNames[dep]; !exists {
				*errors = append(*errors, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "dependency_targets",
					Detail: fmt.Sprintf("depends_on target %s does not exist", dep),
				})
			}
		} else {
			*errors = append(*errors, &FormatError{
				Node:   entry.LogicalName,
				Rule:   "dependency_targets",
				Detail: fmt.Sprintf("depends_on entry %s has unrecognized prefix (expected ROOT/ or ARTIFACT/)", dep),
			})
		}
	}
}

func validateInputTarget(entry *SpecTreeValidateInput, knownNames map[string]struct{}, errors *[]*FormatError) {
	if entry.Frontmatter.Input == "" {
		return
	}

	if !strings.HasPrefix(entry.Frontmatter.Input, "ARTIFACT/") {
		*errors = append(*errors, &FormatError{
			Node:   entry.LogicalName,
			Rule:   "input_target",
			Detail: fmt.Sprintf("input %s must start with ARTIFACT/", entry.Frontmatter.Input),
		})
		return
	}

	if _, exists := knownNames[entry.Frontmatter.Input]; !exists {
		*errors = append(*errors, &FormatError{
			Node:   entry.LogicalName,
			Rule:   "input_target",
			Detail: fmt.Sprintf("input target %s does not exist", entry.Frontmatter.Input),
		})
	}
}

func validateExternalFiles(entry *SpecTreeValidateInput, errors *[]*FormatError) {
	for _, ext := range entry.Frontmatter.External {
		cfsPath := &pathutils.PathCfs{Value: ext.Path}

		// Step 1: Verify existence by opening the file.
		reader, err := filereader.FileOpen(cfsPath)
		if err != nil {
			*errors = append(*errors, &FormatError{
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
			// Parse fragment.lines as "start-end".
			parts := strings.SplitN(fragment.Lines, "-", 2)
			if len(parts) != 2 {
				*errors = append(*errors, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "external_files",
					Detail: fmt.Sprintf("external file %s fragment has invalid lines range %s", ext.Path, fragment.Lines),
				})
				continue
			}

			start, errStart := strconv.Atoi(parts[0])
			end, errEnd := strconv.Atoi(parts[1])
			if errStart != nil || errEnd != nil || start < 1 || start > end {
				*errors = append(*errors, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "external_files",
					Detail: fmt.Sprintf("external file %s fragment has invalid lines range %s", ext.Path, fragment.Lines),
				})
				continue
			}

			// Open the file again for fragment verification.
			fragReader, err := filereader.FileOpen(cfsPath)
			if err != nil {
				*errors = append(*errors, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "external_files",
					Detail: fmt.Sprintf("external file %s cannot be opened for fragment verification: %v", ext.Path, err),
				})
				continue
			}

			// Skip the first start-1 lines.
			filereader.FileSkipLines(fragReader, start-1)

			// Read end - start + 1 lines.
			lineCount := end - start + 1
			var contentBuilder strings.Builder
			outOfRange := false
			for i := 0; i < lineCount; i++ {
				line, err := filereader.FileReadLine(fragReader)
				if err != nil {
					filereader.FileClose(fragReader)
					*errors = append(*errors, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "external_files",
						Detail: fmt.Sprintf("external file %s fragment lines %s out of range", ext.Path, fragment.Lines),
					})
					outOfRange = true
					break
				}
				contentBuilder.WriteString(line)
				contentBuilder.WriteString("\n")
			}

			if outOfRange {
				continue
			}

			filereader.FileClose(fragReader)

			content := contentBuilder.String()

			// Compute SHA-1 of the content string.
			hash := sha1.Sum([]byte(content))

			// Encode as base64url with no padding.
			computed := base64.RawURLEncoding.EncodeToString(hash[:])

			if computed != fragment.Hash {
				*errors = append(*errors, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "external_files",
					Detail: fmt.Sprintf("external file %s fragment %s hash mismatch: expected %s, got %s", ext.Path, fragment.Lines, fragment.Hash, computed),
				})
			}
		}
	}
}

func validateOutputPaths(entry *SpecTreeValidateInput, errors *[]*FormatError) {
	for _, output := range entry.Frontmatter.Outputs {
		if err := pathutils.PathValidateCfs(output.Path); err != nil {
			*errors = append(*errors, &FormatError{
				Node:   entry.LogicalName,
				Rule:   "output_paths",
				Detail: fmt.Sprintf("output path %s is invalid: %v", output.Path, err),
			})
		}
	}
}

func validateDuplicateSubsections(entry *SpecTreeValidateInput, errors *[]*FormatError) {
	if entry.Node.Public == nil {
		return
	}

	if len(entry.Node.Public.Subsections) == 0 {
		return
	}

	seenHeadings := make(map[string]struct{})
	for _, subsection := range entry.Node.Public.Subsections {
		normalized := textnormalization.NormalizeText(subsection.Heading)
		if _, seen := seenHeadings[normalized]; seen {
			*errors = append(*errors, &FormatError{
				Node:   entry.LogicalName,
				Rule:   "duplicate_subsections",
				Detail: fmt.Sprintf("duplicate public subsection heading %s", subsection.RawHeading),
			})
		} else {
			seenHeadings[normalized] = struct{}{}
		}
	}
}
