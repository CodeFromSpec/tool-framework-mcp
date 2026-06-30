// code-from-spec: SPEC/golang/implementation/spec_tree/validate@z4eFkPhiMNC-FhF0MNiZ1LvIJBI
package spectreevalidate

import (
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"
)

type FormatError struct {
	Node   string
	Rule   string
	Detail string
}

func SpecTreeValidate(entries []parsing.Node, allDirs []string) []FormatError {
	var errs []FormatError

	knownNames := make(map[string]bool)
	for _, entry := range entries {
		knownNames[entry.Reference.LogicalName] = true
		if entry.Frontmatter != nil && entry.Frontmatter.Output != nil {
			suffix := strings.TrimPrefix(entry.Reference.LogicalName, "SPEC/")
			artifactName := "ARTIFACT/" + suffix
			knownNames[artifactName] = true
		}
	}

	hasChildren := make(map[string]bool)
	for _, entry := range entries {
		hasChildren[entry.Reference.LogicalName] = false
	}
	for _, a := range entries {
		for _, b := range entries {
			if a.Reference.LogicalName != b.Reference.LogicalName {
				if strings.HasPrefix(b.Reference.LogicalName, a.Reference.LogicalName+"/") {
					hasChildren[a.Reference.LogicalName] = true
				}
			}
		}
	}

	for _, entry := range entries {
		normalizedHeading := parsing.NormalizeText(entry.NameSection.Heading)
		normalizedName := parsing.NormalizeText(entry.Reference.LogicalName)
		if normalizedHeading != normalizedName {
			errs = append(errs, FormatError{
				Node:   entry.Reference.LogicalName,
				Rule:   "name_heading",
				Detail: "first heading does not match the node logical name",
			})
		}

		if hasChildren[entry.Reference.LogicalName] {
			if entry.Frontmatter != nil && len(entry.Frontmatter.DependsOn) > 0 {
				errs = append(errs, FormatError{
					Node:   entry.Reference.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "depends_on is only permitted on leaf nodes",
				})
			}
			if entry.Frontmatter != nil && entry.Frontmatter.Input != nil {
				errs = append(errs, FormatError{
					Node:   entry.Reference.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "input is only permitted on leaf nodes",
				})
			}
			if entry.Frontmatter != nil && entry.Frontmatter.Output != nil {
				errs = append(errs, FormatError{
					Node:   entry.Reference.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "output is only permitted on leaf nodes",
				})
			}
		}

		if hasChildren[entry.Reference.LogicalName] {
			if entry.Agent != nil {
				errs = append(errs, FormatError{
					Node:   entry.Reference.LogicalName,
					Rule:   "leaf_only_agent",
					Detail: "# Agent section is only permitted on leaf nodes",
				})
			}
		}

		if entry.Frontmatter != nil {
			for _, dep := range entry.Frontmatter.DependsOn {
				if strings.HasPrefix(dep, "SPEC/") {
					ref, err := parsing.CfsReferenceFromName(dep)
					if err != nil {
						errs = append(errs, FormatError{
							Node:   entry.Reference.LogicalName,
							Rule:   "dependency_targets",
							Detail: "depends_on entry cannot be parsed: " + dep,
						})
						continue
					}
					if !knownNames[ref.LogicalName] {
						errs = append(errs, FormatError{
							Node:   entry.Reference.LogicalName,
							Rule:   "dependency_targets",
							Detail: "depends_on references unknown SPEC node: " + dep,
						})
					} else if ref.LogicalName == entry.Reference.LogicalName {
						errs = append(errs, FormatError{
							Node:   entry.Reference.LogicalName,
							Rule:   "dependency_targets",
							Detail: "depends_on must not reference the node itself: " + dep,
						})
					} else if strings.HasPrefix(entry.Reference.LogicalName, ref.LogicalName+"/") {
						errs = append(errs, FormatError{
							Node:   entry.Reference.LogicalName,
							Rule:   "dependency_targets",
							Detail: "depends_on must not reference an ancestor: " + dep,
						})
					} else if strings.HasPrefix(ref.LogicalName, entry.Reference.LogicalName+"/") {
						errs = append(errs, FormatError{
							Node:   entry.Reference.LogicalName,
							Rule:   "dependency_targets",
							Detail: "depends_on must not reference a descendant: " + dep,
						})
					}
				} else if strings.HasPrefix(dep, "ARTIFACT/") {
					if !knownNames[dep] {
						errs = append(errs, FormatError{
							Node:   entry.Reference.LogicalName,
							Rule:   "dependency_targets",
							Detail: "depends_on references unknown ARTIFACT: " + dep,
						})
					}
				} else if strings.HasPrefix(dep, "EXTERNAL/") {
					relative := strings.TrimPrefix(dep, "EXTERNAL/")
					cfsPath := oslayer.CfsPath(relative)
					handle, err := oslayer.OpenFile(cfsPath, "read", 30000)
					if err != nil {
						errs = append(errs, FormatError{
							Node:   entry.Reference.LogicalName,
							Rule:   "dependency_targets",
							Detail: "depends_on references unreadable EXTERNAL file: " + dep,
						})
					} else {
						handle.Close()
					}
				} else {
					errs = append(errs, FormatError{
						Node:   entry.Reference.LogicalName,
						Rule:   "dependency_targets",
						Detail: "depends_on entry has unrecognized prefix: " + dep,
					})
				}
			}
		}

		if entry.Frontmatter != nil && entry.Frontmatter.Input != nil {
			inp := *entry.Frontmatter.Input
			if strings.HasPrefix(inp, "SPEC/") {
				ref, err := parsing.CfsReferenceFromName(inp)
				if err != nil {
					errs = append(errs, FormatError{
						Node:   entry.Reference.LogicalName,
						Rule:   "input_target",
						Detail: "input entry cannot be parsed: " + inp,
					})
				} else {
					if !knownNames[ref.LogicalName] {
						errs = append(errs, FormatError{
							Node:   entry.Reference.LogicalName,
							Rule:   "input_target",
							Detail: "input references unknown SPEC node: " + inp,
						})
					}
				}
			} else if strings.HasPrefix(inp, "ARTIFACT/") {
				if !knownNames[inp] {
					errs = append(errs, FormatError{
						Node:   entry.Reference.LogicalName,
						Rule:   "input_target",
						Detail: "input references unknown ARTIFACT: " + inp,
					})
				}
			} else if strings.HasPrefix(inp, "EXTERNAL/") {
				relative := strings.TrimPrefix(inp, "EXTERNAL/")
				cfsPath := oslayer.CfsPath(relative)
				handle, err := oslayer.OpenFile(cfsPath, "read", 30000)
				if err != nil {
					errs = append(errs, FormatError{
						Node:   entry.Reference.LogicalName,
						Rule:   "input_target",
						Detail: "input references unreadable EXTERNAL file: " + inp,
					})
				} else {
					handle.Close()
				}
			} else {
				errs = append(errs, FormatError{
					Node:   entry.Reference.LogicalName,
					Rule:   "input_target",
					Detail: "input must start with SPEC/, ARTIFACT/, or EXTERNAL/",
				})
			}
		}

		if entry.Frontmatter != nil && entry.Frontmatter.Output != nil {
			if err := oslayer.ValidateStringIsCfsPath(*entry.Frontmatter.Output); err != nil {
				errs = append(errs, FormatError{
					Node:   entry.Reference.LogicalName,
					Rule:   "output_paths",
					Detail: "output path is invalid: " + err.Error(),
				})
			}
		}

		if entry.Public != nil {
			for _, line := range entry.Public.Content {
				if strings.TrimSpace(line) != "" {
					errs = append(errs, FormatError{
						Node:   entry.Reference.LogicalName,
						Rule:   "public_subsection_required",
						Detail: "content in # Public must be under a ## subsection",
					})
					break
				}
			}

			if len(entry.Public.Subsections) > 0 {
				seenHeadings := make(map[string]bool)
				for _, subsection := range entry.Public.Subsections {
					norm := parsing.NormalizeText(subsection.Heading)
					if seenHeadings[norm] {
						errs = append(errs, FormatError{
							Node:   entry.Reference.LogicalName,
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

	knownLogicalNames := make(map[string]bool)
	for _, entry := range entries {
		knownLogicalNames[entry.Reference.LogicalName] = true
	}

	for _, dirPath := range allDirs {
		if dirPath == "code-from-spec" || dirPath == "code-from-spec/" {
			continue
		}
		relativePart := strings.TrimPrefix(dirPath, "code-from-spec/")
		if relativePart == dirPath || relativePart == "" {
			continue
		}
		segments := strings.Split(relativePart, "/")
		skip := false
		for _, seg := range segments {
			if strings.HasPrefix(seg, ".") {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		expectedLogicalName := "SPEC/" + relativePart
		if !knownLogicalNames[expectedLogicalName] {
			errs = append(errs, FormatError{
				Node:   dirPath,
				Rule:   "missing_node_md",
				Detail: "subdirectory has no _node.md",
			})
		}
	}

	return errs
}
