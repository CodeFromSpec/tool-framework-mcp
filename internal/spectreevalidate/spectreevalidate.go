// code-from-spec: ROOT/golang/implementation/spec_tree/validate@-966vFr9gx1ObecjgbH6_wzX5Aw

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
func SpecTreeValidate(entries []*SpecTreeValidateInput) ([]*FormatError, error) {
	// Step 1 — Build the known names set.
	knownNames := make(map[string]struct{})

	for _, entry := range entries {
		knownNames[entry.LogicalName] = struct{}{}
	}

	for _, entry := range entries {
		if entry.Frontmatter == nil {
			continue
		}
		barePath := strings.TrimPrefix(entry.LogicalName, "ROOT/")
		for _, output := range entry.Frontmatter.Outputs {
			if output == nil {
				continue
			}
			artifactName := "ARTIFACT/" + barePath + "(" + output.ID + ")"
			knownNames[artifactName] = struct{}{}
		}
	}

	// Step 2 — Validate each entry.
	var formatErrors []*FormatError

	for _, entry := range entries {
		hasChildren := false
		prefix := entry.LogicalName + "/"
		for _, other := range entries {
			if other.LogicalName != entry.LogicalName && strings.HasPrefix(other.LogicalName, prefix) {
				hasChildren = true
				break
			}
		}

		formatErrors = append(formatErrors, validateNameHeading(entry)...)
		formatErrors = append(formatErrors, validateLeafOnlyFields(entry, hasChildren)...)
		formatErrors = append(formatErrors, validateLeafOnlyAgent(entry, hasChildren)...)
		formatErrors = append(formatErrors, validateDependencyTargets(entry, knownNames)...)
		formatErrors = append(formatErrors, validateInputTarget(entry, knownNames)...)
		formatErrors = append(formatErrors, validateExternalFiles(entry)...)
		formatErrors = append(formatErrors, validateOutputPaths(entry)...)
		formatErrors = append(formatErrors, validateDuplicateSubsections(entry)...)
	}

	return formatErrors, nil
}

func validateNameHeading(entry *SpecTreeValidateInput) []*FormatError {
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
				Detail: fmt.Sprintf("name section heading %q does not match logical name %q", entry.Node.NameSection.Heading, entry.LogicalName),
			},
		}
	}
	return nil
}

func validateLeafOnlyFields(entry *SpecTreeValidateInput, hasChildren bool) []*FormatError {
	if !hasChildren {
		return nil
	}
	if entry.Frontmatter == nil {
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

func validateLeafOnlyAgent(entry *SpecTreeValidateInput, hasChildren bool) []*FormatError {
	if !hasChildren {
		return nil
	}
	if entry.Node == nil {
		return nil
	}

	if entry.Node.Agent != nil {
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

func validateDependencyTargets(entry *SpecTreeValidateInput, knownNames map[string]struct{}) []*FormatError {
	if entry.Frontmatter == nil {
		return nil
	}

	var errs []*FormatError

	for _, refPtr := range entry.Frontmatter.DependsOn {
		if refPtr == nil {
			continue
		}
		ref := *refPtr

		if strings.HasPrefix(ref, "ROOT/") {
			bareName := ref
			if idx := strings.Index(ref, "("); idx != -1 {
				bareName = ref[:idx]
			}

			if _, exists := knownNames[bareName]; !exists {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "dependency_targets",
					Detail: fmt.Sprintf("depends_on refers to unknown node %q", ref),
				})
				continue
			}

			if bareName == entry.LogicalName {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "dependency_targets",
					Detail: fmt.Sprintf("depends_on refers to the node itself: %q", ref),
				})
				continue
			}

			if strings.HasPrefix(entry.LogicalName, bareName+"/") {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "dependency_targets",
					Detail: fmt.Sprintf("depends_on refers to an ancestor node: %q", ref),
				})
				continue
			}

			if strings.HasPrefix(bareName, entry.LogicalName+"/") {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "dependency_targets",
					Detail: fmt.Sprintf("depends_on refers to a descendant node: %q", ref),
				})
				continue
			}

		} else if strings.HasPrefix(ref, "ARTIFACT/") {
			if _, exists := knownNames[ref]; !exists {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "dependency_targets",
					Detail: fmt.Sprintf("depends_on refers to unknown artifact %q", ref),
				})
				continue
			}
		} else {
			errs = append(errs, &FormatError{
				Node:   entry.LogicalName,
				Rule:   "dependency_targets",
				Detail: fmt.Sprintf("depends_on entry has unrecognized prefix: %q", ref),
			})
		}
	}

	return errs
}

func validateInputTarget(entry *SpecTreeValidateInput, knownNames map[string]struct{}) []*FormatError {
	if entry.Frontmatter == nil || entry.Frontmatter.Input == "" {
		return nil
	}

	input := entry.Frontmatter.Input

	if !strings.HasPrefix(input, "ARTIFACT/") {
		return []*FormatError{
			{
				Node:   entry.LogicalName,
				Rule:   "input_target",
				Detail: fmt.Sprintf("input must start with ARTIFACT/, got %q", input),
			},
		}
	}

	if _, exists := knownNames[input]; !exists {
		return []*FormatError{
			{
				Node:   entry.LogicalName,
				Rule:   "input_target",
				Detail: fmt.Sprintf("input refers to unknown artifact %q", input),
			},
		}
	}

	return nil
}

func validateExternalFiles(entry *SpecTreeValidateInput) []*FormatError {
	if entry.Frontmatter == nil {
		return nil
	}

	var errs []*FormatError

	for _, ext := range entry.Frontmatter.External {
		if ext == nil {
			continue
		}

		pathCfs := &pathutils.PathCfs{Value: ext.Path}

		// Step 1 — Verify existence.
		reader, err := filereader.FileOpen(pathCfs)
		if err != nil {
			errs = append(errs, &FormatError{
				Node:   entry.LogicalName,
				Rule:   "external_files",
				Detail: fmt.Sprintf("external file cannot be opened: %q", ext.Path),
			})
			continue
		}
		filereader.FileClose(reader)

		// Step 2 — Verify fragments.
		if len(ext.Fragments) == 0 {
			continue
		}

		for _, fragment := range ext.Fragments {
			if fragment == nil {
				continue
			}

			start, end, parseErr := parseLineRange(fragment.Lines)
			if parseErr != nil || start < 1 || start > end {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "external_files",
					Detail: fmt.Sprintf("fragment has invalid lines range %q in %q", fragment.Lines, ext.Path),
				})
				continue
			}

			fragmentReader, openErr := filereader.FileOpen(pathCfs)
			if openErr != nil {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "external_files",
					Detail: fmt.Sprintf("external file cannot be opened for fragment read: %q", ext.Path),
				})
				continue
			}

			filereader.FileSkipLines(fragmentReader, start-1)

			lineCount := end - start + 1
			readLines := make([]string, 0, lineCount)
			outOfRange := false

			for i := 0; i < lineCount; i++ {
				line, readErr := filereader.FileReadLine(fragmentReader)
				if readErr != nil {
					if errors.Is(readErr, filereader.ErrEndOfFile) {
						filereader.FileClose(fragmentReader)
						errs = append(errs, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "external_files",
							Detail: fmt.Sprintf("fragment out of range: lines %d-%d in %q", start, end, ext.Path),
						})
						outOfRange = true
						break
					}
					filereader.FileClose(fragmentReader)
					outOfRange = true
					break
				}
				readLines = append(readLines, line)
			}

			if outOfRange {
				continue
			}

			filereader.FileClose(fragmentReader)

			content := strings.Join(readLines, "\n")
			digest := sha1.Sum([]byte(content))
			computedHash := base64.RawURLEncoding.EncodeToString(digest[:])

			if computedHash != fragment.Hash {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "external_files",
					Detail: fmt.Sprintf("fragment hash mismatch for lines %d-%d in %q: expected %q, got %q", start, end, ext.Path, fragment.Hash, computedHash),
				})
			}
		}
	}

	return errs
}

// parseLineRange parses a string of the form "<start>-<end>" and returns
// the two integers. Returns an error if the format is invalid.
func parseLineRange(s string) (int, int, error) {
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range format: %q", s)
	}
	start, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start in range %q: %w", s, err)
	}
	end, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end in range %q: %w", s, err)
	}
	return start, end, nil
}

func validateOutputPaths(entry *SpecTreeValidateInput) []*FormatError {
	if entry.Frontmatter == nil {
		return nil
	}

	var errs []*FormatError

	for _, output := range entry.Frontmatter.Outputs {
		if output == nil {
			continue
		}
		if err := pathutils.PathValidateCfs(output.Path); err != nil {
			errs = append(errs, &FormatError{
				Node:   entry.LogicalName,
				Rule:   "output_paths",
				Detail: fmt.Sprintf("output path %q is invalid: %s", output.Path, err.Error()),
			})
		}
	}

	return errs
}

func validateDuplicateSubsections(entry *SpecTreeValidateInput) []*FormatError {
	if entry.Node == nil || entry.Node.Public == nil {
		return nil
	}
	if len(entry.Node.Public.Subsections) == 0 {
		return nil
	}

	var errs []*FormatError
	seenHeadings := make(map[string]struct{})

	for _, subsection := range entry.Node.Public.Subsections {
		if subsection == nil {
			continue
		}
		normalized := textnormalization.NormalizeText(subsection.Heading)
		if _, exists := seenHeadings[normalized]; exists {
			errs = append(errs, &FormatError{
				Node:   entry.LogicalName,
				Rule:   "duplicate_subsections",
				Detail: fmt.Sprintf("duplicate public subsection heading %q", subsection.Heading),
			})
		} else {
			seenHeadings[normalized] = struct{}{}
		}
	}

	return errs
}
