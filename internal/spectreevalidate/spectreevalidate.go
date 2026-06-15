// code-from-spec: SPEC/golang/implementation/spec_tree/validate@MAbawU8Ck4ym3u0e2wSAf9i7Ie4
package spectreevalidate

import (
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/textnormalization"
	"strings"
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
	var errors []*FormatError

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
				errors = append(errors, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "name_heading",
					Detail: "first heading " + entry.Node.NameSection.RawHeading + " does not match node logical name " + entry.LogicalName,
				})
			}
		}

		if hasChildren[entry.LogicalName] {
			if entry.Frontmatter != nil && len(entry.Frontmatter.DependsOn) > 0 {
				errors = append(errors, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "field depends_on is only permitted on leaf nodes",
				})
			}
			if entry.Frontmatter != nil && entry.Frontmatter.Input != "" {
				errors = append(errors, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "field input is only permitted on leaf nodes",
				})
			}
			if entry.Frontmatter != nil && entry.Frontmatter.Output != "" {
				errors = append(errors, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "field output is only permitted on leaf nodes",
				})
			}
		}

		if hasChildren[entry.LogicalName] {
			if entry.Node != nil && entry.Node.Agent != nil {
				errors = append(errors, &FormatError{
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
						errors = append(errors, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "dependency_targets",
							Detail: "depends_on target " + dep + " does not exist",
						})
					} else if bareName == entry.LogicalName {
						errors = append(errors, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "dependency_targets",
							Detail: "depends_on target " + dep + " refers to the node itself",
						})
					} else if strings.HasPrefix(entry.LogicalName, bareName+"/") {
						errors = append(errors, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "dependency_targets",
							Detail: "depends_on target " + dep + " is an ancestor of this node",
						})
					} else if strings.HasPrefix(bareName, entry.LogicalName+"/") {
						errors = append(errors, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "dependency_targets",
							Detail: "depends_on target " + dep + " is a descendant of this node",
						})
					}
				} else if logicalnames.LogicalNameIsArtifact(dep) {
					bareRef := logicalnames.LogicalNameStripQualifier(dep)
					if !knownNames[bareRef] {
						errors = append(errors, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "dependency_targets",
							Detail: "depends_on target " + dep + " does not exist",
						})
					}
				} else if logicalnames.LogicalNameIsExternal(dep) {
					cfsPath, err := logicalnames.LogicalNameExternalToPath(dep)
					if err != nil {
						errors = append(errors, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "dependency_targets",
							Detail: "depends_on external target " + dep + " is not readable",
						})
					} else {
						reader, err := filereader.FileOpen(*cfsPath)
						if err != nil {
							errors = append(errors, &FormatError{
								Node:   entry.LogicalName,
								Rule:   "dependency_targets",
								Detail: "depends_on external target " + dep + " is not readable",
							})
						} else {
							filereader.FileClose(reader)
						}
					}
				} else {
					errors = append(errors, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "dependency_targets",
						Detail: "depends_on entry " + dep + " has an unrecognized prefix",
					})
				}
			}
		}

		if entry.Frontmatter != nil && entry.Frontmatter.Input != "" {
			inp := entry.Frontmatter.Input
			if logicalnames.LogicalNameIsArtifact(inp) {
				bareRef := logicalnames.LogicalNameStripQualifier(inp)
				if !knownNames[bareRef] {
					errors = append(errors, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "input_target",
						Detail: "input target " + inp + " does not exist",
					})
				}
			} else if logicalnames.LogicalNameIsExternal(inp) {
				cfsPath, err := logicalnames.LogicalNameExternalToPath(inp)
				if err != nil {
					errors = append(errors, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "input_target",
						Detail: "input external target " + inp + " is not readable",
					})
				} else {
					reader, err := filereader.FileOpen(*cfsPath)
					if err != nil {
						errors = append(errors, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "input_target",
							Detail: "input external target " + inp + " is not readable",
						})
					} else {
						filereader.FileClose(reader)
					}
				}
			} else {
				errors = append(errors, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "input_target",
					Detail: "input field must start with ARTIFACT/ or EXTERNAL/",
				})
			}
		}

		if entry.Frontmatter != nil && entry.Frontmatter.Output != "" {
			if err := pathutils.PathValidateCfs(entry.Frontmatter.Output); err != nil {
				errors = append(errors, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "output_paths",
					Detail: "output path " + entry.Frontmatter.Output + " is invalid: " + err.Error(),
				})
			}
		}

		if entry.Node != nil && entry.Node.Public != nil {
			for _, line := range entry.Node.Public.Content {
				if strings.TrimSpace(line) != "" {
					errors = append(errors, &FormatError{
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
						errors = append(errors, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "duplicate_subsections",
							Detail: "duplicate ## subsection heading " + subsection.RawHeading + " in # Public",
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
			errors = append(errors, &FormatError{
				Node:   dirPath,
				Rule:   "missing_node_md",
				Detail: "subdirectory has no _node.md",
			})
		}
	}

	return errors
}
