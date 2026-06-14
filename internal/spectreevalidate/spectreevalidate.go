// code-from-spec: ROOT/golang/implementation/spec_tree/validate@hFtVtHn0C9rma0H6ASDOFsZScpg
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
			barePath := strings.TrimPrefix(entry.LogicalName, "SPEC/")
			artifactName := "ARTIFACT/" + barePath
			knownNames[artifactName] = true
		}
	}

	for _, entry := range entries {
		hasChildren := false
		for _, other := range entries {
			if strings.HasPrefix(other.LogicalName, entry.LogicalName+"/") {
				hasChildren = true
				break
			}
		}

		if entry.Node != nil && entry.Node.NameSection != nil {
			normalizedHeading := textnormalization.NormalizeText(entry.Node.NameSection.Heading)
			normalizedName := textnormalization.NormalizeText(entry.LogicalName)
			if normalizedHeading != normalizedName {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "name_heading",
					Detail: "name section heading does not match logical name",
				})
			}
		}

		if hasChildren && entry.Frontmatter != nil {
			if len(entry.Frontmatter.DependsOn) > 0 {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "depends_on is only permitted on leaf nodes",
				})
			}
			if entry.Frontmatter.Input != "" {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "input is only permitted on leaf nodes",
				})
			}
			if entry.Frontmatter.Output != "" {
				errs = append(errs, &FormatError{
					Node:   entry.LogicalName,
					Rule:   "leaf_only_fields",
					Detail: "output is only permitted on leaf nodes",
				})
			}
		}

		if hasChildren && entry.Node != nil && entry.Node.Agent != nil {
			errs = append(errs, &FormatError{
				Node:   entry.LogicalName,
				Rule:   "leaf_only_agent",
				Detail: "# Agent section is only permitted on leaf nodes",
			})
		}

		if entry.Frontmatter != nil {
			for _, dep := range entry.Frontmatter.DependsOn {
				if logicalnames.LogicalNameIsSpec(dep) {
					bare := logicalnames.LogicalNameStripQualifier(dep)
					if !knownNames[bare] {
						errs = append(errs, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "dependency_targets",
							Detail: fmt.Sprintf("depends_on target %s does not exist", dep),
						})
						continue
					}
					if bare == entry.LogicalName {
						errs = append(errs, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "dependency_targets",
							Detail: fmt.Sprintf("depends_on target %s points to the node itself", dep),
						})
						continue
					}
					if strings.HasPrefix(entry.LogicalName, bare+"/") {
						errs = append(errs, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "dependency_targets",
							Detail: fmt.Sprintf("depends_on target %s points to an ancestor", dep),
						})
						continue
					}
					if strings.HasPrefix(bare, entry.LogicalName+"/") {
						errs = append(errs, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "dependency_targets",
							Detail: fmt.Sprintf("depends_on target %s points to a descendant", dep),
						})
						continue
					}
				} else if logicalnames.LogicalNameIsArtifact(dep) {
					bare := logicalnames.LogicalNameStripQualifier(dep)
					if !knownNames[bare] {
						errs = append(errs, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "dependency_targets",
							Detail: fmt.Sprintf("depends_on target %s does not exist", dep),
						})
					}
				} else if logicalnames.LogicalNameIsExternal(dep) {
					extPathCfs, err := logicalnames.LogicalNameExternalToPath(dep)
					if err != nil {
						errs = append(errs, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "dependency_targets",
							Detail: fmt.Sprintf("depends_on external target %s is not readable", dep),
						})
					} else {
						reader, err := filereader.FileOpen(extPathCfs)
						if err != nil {
							errs = append(errs, &FormatError{
								Node:   entry.LogicalName,
								Rule:   "dependency_targets",
								Detail: fmt.Sprintf("depends_on external target %s is not readable", dep),
							})
						} else {
							filereader.FileClose(reader)
						}
					}
				} else {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "dependency_targets",
						Detail: fmt.Sprintf("depends_on entry %s has an unrecognized prefix", dep),
					})
				}
			}

			inp := entry.Frontmatter.Input
			if inp != "" {
				if logicalnames.LogicalNameIsArtifact(inp) {
					bare := logicalnames.LogicalNameStripQualifier(inp)
					if !knownNames[bare] {
						errs = append(errs, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "input_target",
							Detail: fmt.Sprintf("input target %s does not exist", inp),
						})
					}
				} else if logicalnames.LogicalNameIsExternal(inp) {
					extPathCfs, err := logicalnames.LogicalNameExternalToPath(inp)
					if err != nil {
						errs = append(errs, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "input_target",
							Detail: fmt.Sprintf("input external target %s is not readable", inp),
						})
					} else {
						reader, err := filereader.FileOpen(extPathCfs)
						if err != nil {
							errs = append(errs, &FormatError{
								Node:   entry.LogicalName,
								Rule:   "input_target",
								Detail: fmt.Sprintf("input external target %s is not readable", inp),
							})
						} else {
							filereader.FileClose(reader)
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

			if entry.Frontmatter.Output != "" {
				err := pathutils.PathValidateCfs(entry.Frontmatter.Output)
				if err != nil {
					errs = append(errs, &FormatError{
						Node:   entry.LogicalName,
						Rule:   "output_paths",
						Detail: fmt.Sprintf("output path is invalid: %s", err.Error()),
					})
				}
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
					normalized := textnormalization.NormalizeText(subsection.Heading)
					if seenHeadings[normalized] {
						errs = append(errs, &FormatError{
							Node:   entry.LogicalName,
							Rule:   "duplicate_subsections",
							Detail: fmt.Sprintf("duplicate ## subsection heading %s in # Public", subsection.RawHeading),
						})
					} else {
						seenHeadings[normalized] = true
					}
				}
			}
		}
	}

	for _, dirPath := range allDirs {
		if dirPath == "code-from-spec/" {
			continue
		}

		remainder := strings.TrimPrefix(dirPath, "code-from-spec/")
		firstSegment := remainder
		if idx := strings.Index(remainder, "/"); idx >= 0 {
			firstSegment = remainder[:idx]
		}
		if strings.HasPrefix(firstSegment, "_") {
			continue
		}

		expectedFile := dirPath
		if !strings.HasSuffix(expectedFile, "/") {
			expectedFile += "/"
		}
		expectedFile += "_node.md"

		found := false
		for _, entry := range entries {
			nodePath, err := logicalnames.LogicalNameToPath(entry.LogicalName)
			if err != nil {
				continue
			}
			if nodePath.Value == expectedFile {
				found = true
				break
			}
		}

		if !found {
			errs = append(errs, &FormatError{
				Node:   dirPath,
				Rule:   "missing_node_md",
				Detail: "subdirectory has no _node.md",
			})
		}
	}

	return errs
}
