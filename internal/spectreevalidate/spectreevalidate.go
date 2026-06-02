// code-from-spec: ROOT/golang/implementation/spec_tree/validate@pVBJ9GuurHXIp3cL1L2xTj6uLjU
package spectreevalidate

import (
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
)

type SpecTreeValidateInput struct {
	LogicalName string
	Frontmatter frontmatter.Frontmatter
	Node        parsenode.Node
}

type FormatError struct {
	Node   string
	Rule   string
	Detail string
}

func SpecTreeValidate(entries []*SpecTreeValidateInput) []*FormatError {
	knownNames := buildKnownNames(entries)
	hasChildren := buildHasChildren(entries)

	var errs []*FormatError

	for _, entry := range entries {
		errs = append(errs, validateNameHeading(entry)...)
		errs = append(errs, validateLeafOnlyFields(entry, hasChildren[entry.LogicalName])...)
		errs = append(errs, validateLeafOnlyAgent(entry, hasChildren[entry.LogicalName])...)
		errs = append(errs, validateDependencyTargets(entry, knownNames)...)
		errs = append(errs, validateInputTarget(entry, knownNames)...)
		errs = append(errs, validateExternalFiles(entry)...)
		errs = append(errs, validateOutputPaths(entry)...)
		errs = append(errs, validateDuplicateSubsections(entry)...)
	}

	return errs
}

func buildKnownNames(entries []*SpecTreeValidateInput) map[string]bool {
	known := make(map[string]bool)
	for _, entry := range entries {
		known[entry.LogicalName] = true
		if entry.Frontmatter.Output != "" {
			artifactName := "ARTIFACT/" + strings.TrimPrefix(entry.LogicalName, "ROOT/")
			known[artifactName] = true
		}
	}
	return known
}

func buildHasChildren(entries []*SpecTreeValidateInput) map[string]bool {
	hasChildren := make(map[string]bool)
	for _, entry := range entries {
		for _, other := range entries {
			if other.LogicalName == entry.LogicalName {
				continue
			}
			if strings.HasPrefix(other.LogicalName, entry.LogicalName+"/") {
				hasChildren[entry.LogicalName] = true
				break
			}
		}
	}
	return hasChildren
}

func validateNameHeading(entry *SpecTreeValidateInput) []*FormatError {
	if entry.Node.NameSection == nil {
		return nil
	}
	normalizedHeading := textnormalization.NormalizeText(entry.Node.NameSection.Heading)
	normalizedName := textnormalization.NormalizeText(entry.LogicalName)
	if normalizedHeading != normalizedName {
		return []*FormatError{{
			Node:   entry.LogicalName,
			Rule:   "name_heading",
			Detail: fmt.Sprintf("heading %q does not match logical name %q", entry.Node.NameSection.Heading, entry.LogicalName),
		}}
	}
	return nil
}

func validateLeafOnlyFields(entry *SpecTreeValidateInput, hasChildren bool) []*FormatError {
	if !hasChildren {
		return nil
	}
	var errs []*FormatError
	if len(entry.Frontmatter.DependsOn) > 0 {
		errs = append(errs, &FormatError{
			Node:   entry.LogicalName,
			Rule:   "leaf_only_fields",
			Detail: "non-leaf node has depends_on field",
		})
	}
	if len(entry.Frontmatter.External) > 0 {
		errs = append(errs, &FormatError{
			Node:   entry.LogicalName,
			Rule:   "leaf_only_fields",
			Detail: "non-leaf node has external field",
		})
	}
	if entry.Frontmatter.Input != "" {
		errs = append(errs, &FormatError{
			Node:   entry.LogicalName,
			Rule:   "leaf_only_fields",
			Detail: "non-leaf node has input field",
		})
	}
	if entry.Frontmatter.Output != "" {
		errs = append(errs, &FormatError{
			Node:   entry.LogicalName,
			Rule:   "leaf_only_fields",
			Detail: "non-leaf node has output field",
		})
	}
	return errs
}

func validateLeafOnlyAgent(entry *SpecTreeValidateInput, hasChildren bool) []*FormatError {
	if !hasChildren {
		return nil
	}
	if entry.Node.Agent == nil {
		return nil
	}
	return []*FormatError{{
		Node:   entry.LogicalName,
		Rule:   "leaf_only_agent",
		Detail: "only leaf nodes may have an Agent section",
	}}
}

func validateDependencyTargets(entry *SpecTreeValidateInput, knownNames map[string]bool) []*FormatError {
	var errs []*FormatError
	for _, dep := range entry.Frontmatter.DependsOn {
		if dep == "ROOT" || strings.HasPrefix(dep, "ROOT/") {
			bare := logicalnames.LogicalNameStripQualifier(dep)
			if !knownNames[bare] {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "dependency_targets",
					Detail: fmt.Sprintf("unknown reference %q", bare),
				})
				continue
			}
			if bare == entry.LogicalName {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "dependency_targets",
					Detail: fmt.Sprintf("self-reference %q", bare),
				})
				continue
			}
			if strings.HasPrefix(entry.LogicalName, bare+"/") {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "dependency_targets",
					Detail: fmt.Sprintf("ancestor reference %q", bare),
				})
				continue
			}
			if strings.HasPrefix(bare, entry.LogicalName+"/") {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "dependency_targets",
					Detail: fmt.Sprintf("descendant reference %q", bare),
				})
			}
			continue
		}

		if strings.HasPrefix(dep, "ARTIFACT/") {
			bare := logicalnames.LogicalNameStripQualifier(dep)
			if !knownNames[bare] {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "dependency_targets",
					Detail: fmt.Sprintf("unknown artifact reference %q", bare),
				})
			}
		}
	}
	return errs
}

func validateInputTarget(entry *SpecTreeValidateInput, knownNames map[string]bool) []*FormatError {
	if entry.Frontmatter.Input == "" {
		return nil
	}
	if !strings.HasPrefix(entry.Frontmatter.Input, "ARTIFACT/") {
		return []*FormatError{{
			Node:   entry.LogicalName,
			Rule:   "input_target",
			Detail: fmt.Sprintf("input must be an ARTIFACT/ reference, got %q", entry.Frontmatter.Input),
		}}
	}
	bare := logicalnames.LogicalNameStripQualifier(entry.Frontmatter.Input)
	if !knownNames[bare] {
		return []*FormatError{{
			Node:   entry.LogicalName,
			Rule:   "input_target",
			Detail: fmt.Sprintf("unknown artifact reference %q", bare),
		}}
	}
	return nil
}

func validateExternalFiles(entry *SpecTreeValidateInput) []*FormatError {
	var errs []*FormatError
	for _, ext := range entry.Frontmatter.External {
		cfsPath := &pathutils.PathCfs{Value: ext.Path}
		r, err := filereader.FileOpen(cfsPath)
		if err != nil {
			errs = append(errs, &FormatError{
				Node:   entry.LogicalName,
				Rule:   "external_files",
				Detail: fmt.Sprintf("path %q: %v", ext.Path, err),
			})
			continue
		}
		filereader.FileClose(r)
	}
	return errs
}

func validateOutputPaths(entry *SpecTreeValidateInput) []*FormatError {
	if entry.Frontmatter.Output == "" {
		return nil
	}
	if err := pathutils.PathValidateCfs(entry.Frontmatter.Output); err != nil {
		return []*FormatError{{
			Node:   entry.LogicalName,
			Rule:   "output_paths",
			Detail: fmt.Sprintf("output path %q: %v", entry.Frontmatter.Output, err),
		}}
	}
	return nil
}

func validateDuplicateSubsections(entry *SpecTreeValidateInput) []*FormatError {
	if entry.Node.Public == nil {
		return nil
	}
	var errs []*FormatError
	seen := make(map[string]bool)
	for _, sub := range entry.Node.Public.Subsections {
		normalized := textnormalization.NormalizeText(sub.Heading)
		if seen[normalized] {
			errs = append(errs, &FormatError{
				Node:   entry.LogicalName,
				Rule:   "duplicate_subsections",
				Detail: fmt.Sprintf("duplicate subsection heading %q", sub.Heading),
			})
		} else {
			seen[normalized] = true
		}
	}
	return errs
}
