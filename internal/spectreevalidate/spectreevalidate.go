// code-from-spec: SPEC/golang/implementation/spec_tree/validate@x9_kBP0fs4z32pTRFgOArMrFEgA
package spectreevalidate

import (
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/file"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/textnormalization"
)

type SpecTreeValidateInput struct {
	LogicalName string
	Frontmatter *frontmatter.Frontmatter
	Node        *parsenode.Node
}

type FormatError struct {
	Node   string
	Rule   string
	Detail string
}

func SpecTreeValidate(entries []*SpecTreeValidateInput, allDirs []string) []*FormatError {
	var errs []*FormatError

	knownNames := make(map[string]bool)
	for _, entry := range entries {
		knownNames[entry.LogicalName] = true
		if entry.Frontmatter != nil && entry.Frontmatter.Output != "" {
			suffix := strings.TrimPrefix(entry.LogicalName, "SPEC/")
			artifactName := "ARTIFACT/" + suffix
			knownNames[artifactName] = true
		}
	}

	hasChildren := make(map[string]bool)
	for _, entry := range entries {
		hasChildren[entry.LogicalName] = false
	}
	for _, a := range entries {
		for _, b := range entries {
			if a.LogicalName != b.LogicalName {
				if strings.HasPrefix(b.LogicalName, a.LogicalName+"/") {
					hasChildren[a.LogicalName] = true
				}
			}
		}
	}

	for _, entry := range entries {
		if entry.Node != nil && entry.Node.NameSection != nil {
			normalizedHeading := textnormalization.NormalizeText(entry.Node.NameSection.Heading)
			normalizedName := textnormalization.NormalizeText(entry.LogicalName)
			if normalizedHeading != normalizedName {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "name_heading",
					Detail: "first heading does not match the node logical name",
				})
			}
		}

		if hasChildren[entry.LogicalName] {
			if entry.Frontmatter != nil && len(entry.Frontmatter.DependsOn) > 0 {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "depends_on is only permitted on leaf nodes",
				})
			}
			if entry.Frontmatter != nil && entry.Frontmatter.Input != "" {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "input is only permitted on leaf nodes",
				})
			}
			if entry.Frontmatter != nil && entry.Frontmatter.Output != "" {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "output is only permitted on leaf nodes",
				})
			}
		}

		if hasChildren[entry.LogicalName] {
			if entry.Node != nil && entry.Node.Agent != nil {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_agent",
					Detail: "# Agent section is only permitted on leaf nodes",
				})
			}
		}

		if entry.Frontmatter != nil {
			for _, dep := range entry.Frontmatter.DependsOn {
				if logicalnames.LogicalNameIsSpec(dep) {
					bareName := logicalnames.LogicalNameStripQualifier(dep)
					if !knownNames[bareName] {
						errs = append(errs, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "dependency_targets",
							Detail: "depends_on references unknown SPEC node: " + dep,
						})
					} else if bareName == entry.LogicalName {
						errs = append(errs, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "dependency_targets",
							Detail: "depends_on must not reference the node itself: " + dep,
						})
					} else if strings.HasPrefix(entry.LogicalName, bareName+"/") {
						errs = append(errs, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "dependency_targets",
							Detail: "depends_on must not reference an ancestor: " + dep,
						})
					} else if strings.HasPrefix(bareName, entry.LogicalName+"/") {
						errs = append(errs, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "dependency_targets",
							Detail: "depends_on must not reference a descendant: " + dep,
						})
					}
				} else if logicalnames.LogicalNameIsArtifact(dep) {
					bareRef := logicalnames.LogicalNameStripQualifier(dep)
					if !knownNames[bareRef] {
						errs = append(errs, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "dependency_targets",
							Detail: "depends_on references unknown ARTIFACT: " + dep,
						})
					}
				} else if logicalnames.LogicalNameIsExternal(dep) {
					cfsPath, err := logicalnames.LogicalNameExternalToPath(dep)
					if err != nil {
						errs = append(errs, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "dependency_targets",
							Detail: "depends_on references unreadable EXTERNAL file: " + dep,
						})
					} else {
						handle, err := file.FileOpen(cfsPath, "read", 30000)
						if err != nil {
							errs = append(errs, &FormatError{
								Node:   entry.LogicalName,
								Rule:   "dependency_targets",
								Detail: "depends_on references unreadable EXTERNAL file: " + dep,
							})
						} else {
							file.FileClose(handle)
						}
					}
				} else {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "dependency_targets",
						Detail: "depends_on entry has unrecognized prefix: " + dep,
					})
				}
			}
		}

		if entry.Frontmatter != nil && entry.Frontmatter.Input != "" {
			inp := entry.Frontmatter.Input
			if logicalnames.LogicalNameIsArtifact(inp) {
				bareRef := logicalnames.LogicalNameStripQualifier(inp)
				if !knownNames[bareRef] {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "input_target",
						Detail: "input references unknown ARTIFACT: " + inp,
					})
				}
			} else if logicalnames.LogicalNameIsExternal(inp) {
				cfsPath, err := logicalnames.LogicalNameExternalToPath(inp)
				if err != nil {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "input_target",
						Detail: "input references unreadable EXTERNAL file: " + inp,
					})
				} else {
					handle, err := file.FileOpen(cfsPath, "read", 30000)
					if err != nil {
						errs = append(errs, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "input_target",
							Detail: "input references unreadable EXTERNAL file: " + inp,
						})
					} else {
						file.FileClose(handle)
					}
				}
			} else {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "input_target",
					Detail: "input must start with ARTIFACT/ or EXTERNAL/",
				})
			}
		}

		if entry.Frontmatter != nil && entry.Frontmatter.Output != "" {
			if err := pathutils.PathValidateCfs(entry.Frontmatter.Output); err != nil {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "output_paths",
					Detail: "output path is invalid: " + err.Error(),
				})
			}
		}

		if entry.Node != nil && entry.Node.Public != nil {
			for _, line := range entry.Node.Public.Content {
				if strings.TrimSpace(line) != "" {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "public_subsection_required",
						Detail: "content in # Public must be under a ## subsection",
					})
					break
				}
			}

			if len(entry.Node.Public.Subsections) > 0 {
				seenHeadings := make(map[string]bool)
				for _, subsection := range entry.Node.Public.Subsections {
					norm := textnormalization.NormalizeText(subsection.Heading)
					if seenHeadings[norm] {
						errs = append(errs, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "duplicate_subsections",
							Detail: "duplicate ## subsection heading in # Public: " + subsection.RawHeading,
						})
					} else {
						seenHeadings[norm] = true
					}
				}
			}
		}
	}

	knownPaths := make(map[string]bool)
	for _, entry := range entries {
		nodePath, err := logicalnames.LogicalNameToPath(entry.LogicalName)
		if err == nil && nodePath != nil {
			knownPaths[nodePath.Value] = true
		}
	}

	for _, dirPath := range allDirs {
		if dirPath == "code-from-spec" || dirPath == "code-from-spec/" {
			continue
		}
		relativePart := strings.TrimPrefix(dirPath, "code-from-spec/")
		if relativePart == dirPath || relativePart == "" {
			continue
		}
		firstSegment := relativePart
		if idx := strings.Index(relativePart, "/"); idx >= 0 {
			firstSegment = relativePart[:idx]
		}
		if strings.HasPrefix(firstSegment, "_") {
			continue
		}
		candidatePath := dirPath + "/_node.md"
		if !knownPaths[candidatePath] {
			errs = append(errs, &FormatError{
				Node:   dirPath,
				Rule:   "missing_node_md",
				Detail: "subdirectory has no _node.md",
			})
		}
	}

	return errs
}
