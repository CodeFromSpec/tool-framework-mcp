// code-from-spec: ROOT/golang/implementation/spec_tree/validate@78R7RyyfYQYIoPMCdpufrJIzxGc
package spectreevalidate

import (
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
)

// SpecTreeValidateInput holds a single discovered node with its parsed
// frontmatter and body, ready for validation.
type SpecTreeValidateInput struct {
	LogicalName string
	Frontmatter *frontmatter.Frontmatter
	Node        *parsenode.Node
}

// FormatError describes a single format rule violation found in a node.
type FormatError struct {
	Node   string
	Rule   string
	Detail string
}

func logicalNameStripQualifier(name string) string {
	i := strings.Index(name, "(")
	if i == -1 {
		return name
	}
	return name[:i]
}

// SpecTreeValidate validates all entries in the input list against the
// spec tree format rules. It returns a list of FormatError values
// describing every violation found. An empty slice means all nodes are
// valid.
func SpecTreeValidate(entries []*SpecTreeValidateInput) []*FormatError {
	knownNames := make(map[string]bool)
	for _, entry := range entries {
		knownNames[entry.LogicalName] = true
		if entry.Frontmatter.Output != "" {
			suffix := strings.TrimPrefix(entry.LogicalName, "ROOT/")
			artifactName := "ARTIFACT/" + suffix
			knownNames[artifactName] = true
		}
	}

	var errs []*FormatError

	for _, entry := range entries {
		hasChildren := false
		prefix := entry.LogicalName + "/"
		for _, other := range entries {
			if strings.HasPrefix(other.LogicalName, prefix) {
				hasChildren = true
				break
			}
		}

		// Rule: name_heading
		expected := textnormalization.NormalizeText(entry.LogicalName)
		actual := entry.Node.NameSection.Heading
		if actual != expected {
			errs = append(errs, &FormatError{
				Node:   entry.LogicalName,
				Rule:   "name_heading",
				Detail: "name section heading " + actual + " does not match logical name " + entry.LogicalName,
			})
		}

		// Rule: leaf_only_fields
		if hasChildren {
			if len(entry.Frontmatter.DependsOn) > 0 {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "non-leaf node has depends_on",
				})
			}
			if len(entry.Frontmatter.External) > 0 {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "non-leaf node has external",
				})
			}
			if entry.Frontmatter.Input != "" {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "non-leaf node has input",
				})
			}
			if entry.Frontmatter.Output != "" {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "non-leaf node has output",
				})
			}
		}

		// Rule: leaf_only_agent
		if hasChildren && entry.Node.Agent != nil {
			errs = append(errs, &FormatError{
				Node:   entry.LogicalName,
				Rule:   "leaf_only_agent",
				Detail: "non-leaf node has an Agent section",
			})
		}

		// Rule: dependency_targets
		for _, dep := range entry.Frontmatter.DependsOn {
			if strings.HasPrefix(dep, "ROOT/") {
				bare := logicalNameStripQualifier(dep)
				if !knownNames[bare] {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "dependency_targets",
						Detail: "depends_on target " + dep + " does not exist",
					})
					continue
				}
				if bare == entry.LogicalName {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "dependency_targets",
						Detail: "depends_on target " + dep + " refers to the node itself",
					})
					continue
				}
				if strings.HasPrefix(entry.LogicalName, bare+"/") {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "dependency_targets",
						Detail: "depends_on target " + dep + " is an ancestor of the node",
					})
					continue
				}
				if strings.HasPrefix(bare, entry.LogicalName+"/") {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "dependency_targets",
						Detail: "depends_on target " + dep + " is a descendant of the node",
					})
					continue
				}
			} else if strings.HasPrefix(dep, "ARTIFACT/") {
				if !knownNames[dep] {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "dependency_targets",
						Detail: "depends_on target " + dep + " does not exist",
					})
				}
			} else {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "dependency_targets",
					Detail: "depends_on entry " + dep + " is not a ROOT/ or ARTIFACT/ reference",
				})
			}
		}

		// Rule: input_target
		if entry.Frontmatter.Input != "" {
			if !strings.HasPrefix(entry.Frontmatter.Input, "ARTIFACT/") {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "input_target",
					Detail: "input " + entry.Frontmatter.Input + " is not an ARTIFACT/ reference",
				})
			} else if !knownNames[entry.Frontmatter.Input] {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "input_target",
					Detail: "input target " + entry.Frontmatter.Input + " does not exist",
				})
			}
		}

		// Rule: external_files
		for _, ext := range entry.Frontmatter.External {
			cfsPath := &pathutils.PathCfs{Value: ext.Path}
			reader, err := filereader.FileOpen(cfsPath)
			if err != nil {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "external_files",
					Detail: "external file " + ext.Path + " cannot be opened",
				})
				continue
			}
			filereader.FileClose(reader)
		}

		// Rule: output_paths
		if entry.Frontmatter.Output != "" {
			if err := pathutils.PathValidateCfs(entry.Frontmatter.Output); err != nil {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "output_paths",
					Detail: "output path " + entry.Frontmatter.Output + " is invalid: " + err.Error(),
				})
			}
		}

		// Rule: duplicate_subsections
		if entry.Node.Public != nil && len(entry.Node.Public.Subsections) > 0 {
			seenHeadings := make(map[string]bool)
			for _, subsection := range entry.Node.Public.Subsections {
				normalized := textnormalization.NormalizeText(subsection.Heading)
				if seenHeadings[normalized] {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "duplicate_subsections",
						Detail: "duplicate Public subsection heading " + subsection.RawHeading,
					})
				} else {
					seenHeadings[normalized] = true
				}
			}
		}
	}

	return errs
}
