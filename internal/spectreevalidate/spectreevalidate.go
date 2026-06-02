// code-from-spec: ROOT/golang/implementation/spec_tree/validate@wjwNmX46oaJ-_nl5M-_Eqmwcv0w
package spectreevalidate

import (
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
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

func logicalNameStripQualifier(name string) string {
	idx := strings.Index(name, "(")
	if idx == -1 {
		return name
	}
	return strings.TrimSpace(name[:idx])
}

func SpecTreeValidate(entries []*SpecTreeValidateInput) []*FormatError {
	knownNames := make(map[string]bool)
	for _, entry := range entries {
		knownNames[entry.LogicalName] = true
		if entry.Frontmatter.Output != "" {
			bare := strings.TrimPrefix(entry.LogicalName, "ROOT/")
			artifactName := "ARTIFACT/" + bare
			knownNames[artifactName] = true
		}
	}

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

	var errs []*FormatError

	for _, entry := range entries {
		normalizedHeading := ""
		if entry.Node.NameSection != nil {
			normalizedHeading = textnormalization.NormalizeText(entry.Node.NameSection.Heading)
		}
		normalizedLogical := textnormalization.NormalizeText(entry.LogicalName)
		if normalizedHeading != normalizedLogical {
			errs = append(errs, &FormatError{
				Node:   entry.LogicalName,
				Rule:   "name_heading",
				Detail: fmt.Sprintf("heading %q does not match logical name %q after normalization", normalizedHeading, normalizedLogical),
			})
		}

		if hasChildren[entry.LogicalName] {
			if len(entry.Frontmatter.DependsOn) > 0 {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "depends_on is only allowed on leaf nodes",
				})
			}
			if len(entry.Frontmatter.External) > 0 {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "external is only allowed on leaf nodes",
				})
			}
			if entry.Frontmatter.Input != "" {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "input is only allowed on leaf nodes",
				})
			}
			if entry.Frontmatter.Output != "" {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "output is only allowed on leaf nodes",
				})
			}
		}

		if hasChildren[entry.LogicalName] && entry.Node.Agent != nil {
			errs = append(errs, &FormatError{
				Node:   entry.LogicalName,
				Rule:   "leaf_only_agent",
				Detail: "only leaf nodes may have an Agent section",
			})
		}

		for _, dep := range entry.Frontmatter.DependsOn {
			if dep == "ROOT" || strings.HasPrefix(dep, "ROOT/") {
				bareName := logicalNameStripQualifier(dep)
				if !knownNames[bareName] {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "dependency_targets",
						Detail: fmt.Sprintf("unknown dependency reference %q", bareName),
					})
				} else if bareName == entry.LogicalName {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "dependency_targets",
						Detail: fmt.Sprintf("dependency %q is a self-reference", bareName),
					})
				} else if strings.HasPrefix(entry.LogicalName, bareName+"/") {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "dependency_targets",
						Detail: fmt.Sprintf("dependency %q is an ancestor of this node", bareName),
					})
				} else if strings.HasPrefix(bareName, entry.LogicalName+"/") {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "dependency_targets",
						Detail: fmt.Sprintf("dependency %q is a descendant of this node", bareName),
					})
				}
			} else if strings.HasPrefix(dep, "ARTIFACT/") {
				bareRef := logicalNameStripQualifier(dep)
				if !knownNames[bareRef] {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "dependency_targets",
						Detail: fmt.Sprintf("unknown artifact reference %q", bareRef),
					})
				}
			}
		}

		if entry.Frontmatter.Input != "" {
			if !strings.HasPrefix(entry.Frontmatter.Input, "ARTIFACT/") {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "input_target",
					Detail: "input must be an ARTIFACT/ reference",
				})
			} else {
				bareRef := logicalNameStripQualifier(entry.Frontmatter.Input)
				if !knownNames[bareRef] {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "input_target",
						Detail: fmt.Sprintf("unknown artifact reference %q", bareRef),
					})
				}
			}
		}

		for _, ext := range entry.Frontmatter.External {
			cfsPath := &pathutils.PathCfs{Value: ext.Path}
			reader, err := filereader.FileOpen(cfsPath)
			if err != nil {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "external_files",
					Detail: fmt.Sprintf("cannot open external file %q: %s", ext.Path, err.Error()),
				})
				continue
			}
			filereader.FileClose(reader)
		}

		if entry.Frontmatter.Output != "" {
			if err := pathutils.PathValidateCfs(entry.Frontmatter.Output); err != nil {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "output_paths",
					Detail: fmt.Sprintf("invalid output path %q: %s", entry.Frontmatter.Output, err.Error()),
				})
			}
		}

		if entry.Node.Public != nil && len(entry.Node.Public.Subsections) > 0 {
			seen := make(map[string]bool)
			for _, sub := range entry.Node.Public.Subsections {
				normalized := textnormalization.NormalizeText(sub.Heading)
				if seen[normalized] {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "duplicate_subsections",
						Detail: fmt.Sprintf("duplicate subsection heading %q in Public section", normalized),
					})
				} else {
					seen[normalized] = true
				}
			}
		}
	}

	if errs == nil {
		return []*FormatError{}
	}
	return errs
}
