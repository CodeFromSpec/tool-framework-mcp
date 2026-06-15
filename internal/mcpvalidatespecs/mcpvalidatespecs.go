// code-from-spec: SPEC/golang/implementation/mcp_tools/validate_specs@N2X9yAovHP5Y2NtkYC81q4lDJhE
package mcpvalidatespecs

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/artifacttag"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/noderanking"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/spectree"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/spectreevalidate"
)

type StalenessEntry struct {
	Node         string
	ArtifactPath string
	Status       string
	Detail       string
	Rank         int
}

type ValidationReport struct {
	FormatErrors []*spectreevalidate.FormatError
	Cycles       []string
	Staleness    []*StalenessEntry
}

type parsedNodeEntry struct {
	logicalName string
	fm          *frontmatter.Frontmatter
	node        *parsenode.Node
}

func MCPValidateSpecs() *ValidationReport {
	report := &ValidationReport{
		FormatErrors: []*spectreevalidate.FormatError{},
		Cycles:       []string{},
		Staleness:    []*StalenessEntry{},
	}

	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		report.FormatErrors = append(report.FormatErrors, &spectreevalidate.FormatError{
			Node:   "",
			Rule:   "scan",
			Detail: err.Error(),
		})
		return report
	}

	allDirs := discoverAllDirs()

	parsedEntries := make([]*parsedNodeEntry, 0, len(nodes))
	for _, specNode := range nodes {
		fm, fmErr := frontmatter.FrontmatterParse(&specNode.FilePath)
		if fmErr != nil {
			report.FormatErrors = append(report.FormatErrors, &spectreevalidate.FormatError{
				Node:   specNode.LogicalName,
				Rule:   "parse",
				Detail: fmErr.Error(),
			})
			continue
		}

		parsedNode, npErr := parsenode.NodeParse(specNode.LogicalName)
		if npErr != nil {
			report.FormatErrors = append(report.FormatErrors, &spectreevalidate.FormatError{
				Node:   specNode.LogicalName,
				Rule:   "parse",
				Detail: npErr.Error(),
			})
			continue
		}

		parsedEntries = append(parsedEntries, &parsedNodeEntry{
			logicalName: specNode.LogicalName,
			fm:          fm,
			node:        parsedNode,
		})
	}

	validateInputs := make([]*spectreevalidate.SpecTreeValidateInput, 0, len(parsedEntries))
	for _, entry := range parsedEntries {
		validateInputs = append(validateInputs, &spectreevalidate.SpecTreeValidateInput{
			LogicalName: entry.logicalName,
			Frontmatter: entry.fm,
			Node:        entry.node,
		})
	}
	formatErrs := spectreevalidate.SpecTreeValidate(validateInputs, allDirs)
	report.FormatErrors = append(report.FormatErrors, formatErrs...)

	rankMap := make(map[string]int)
	var cycles []string

	if len(report.FormatErrors) == 0 {
		rankInputs := make([]*noderanking.NodeRankInput, 0, len(parsedEntries))
		for _, entry := range parsedEntries {
			rankInputs = append(rankInputs, &noderanking.NodeRankInput{
				LogicalName: entry.logicalName,
				Frontmatter: entry.fm,
			})
		}

		rankedEntries, rankCycles, rankErr := noderanking.NodeRankCompute(rankInputs)
		if rankErr != nil {
			if errors.Is(rankErr, noderanking.ErrUnresolvableReference) {
				report.FormatErrors = append(report.FormatErrors, &spectreevalidate.FormatError{
					Node:   "",
					Rule:   "ranking",
					Detail: rankErr.Error(),
				})
			} else {
				report.FormatErrors = append(report.FormatErrors, &spectreevalidate.FormatError{
					Node:   "",
					Rule:   "ranking",
					Detail: rankErr.Error(),
				})
			}
		} else {
			for _, re := range rankedEntries {
				rankMap[re.LogicalName] = re.Rank
			}
			cycles = rankCycles
		}
	}

	report.Cycles = cycles

	orderedEntries := make([]*parsedNodeEntry, len(parsedEntries))
	copy(orderedEntries, parsedEntries)

	if len(rankMap) > 0 {
		sort.Slice(orderedEntries, func(i, j int) bool {
			ri := rankMap[orderedEntries[i].logicalName]
			rj := rankMap[orderedEntries[j].logicalName]
			if ri != rj {
				return ri < rj
			}
			return orderedEntries[i].logicalName < orderedEntries[j].logicalName
		})
	} else {
		sort.Slice(orderedEntries, func(i, j int) bool {
			return orderedEntries[i].logicalName < orderedEntries[j].logicalName
		})
	}

	for _, entry := range orderedEntries {
		if entry.fm.Output == "" {
			continue
		}

		nodeRank := rankMap[entry.logicalName]

		chain, resolveErr := chainresolver.ChainResolve(entry.logicalName)
		if resolveErr != nil {
			report.Staleness = append(report.Staleness, &StalenessEntry{
				Node:         entry.logicalName,
				ArtifactPath: entry.fm.Output,
				Status:       "missing",
				Detail:       resolveErr.Error(),
				Rank:         nodeRank,
			})
			continue
		}

		chainHash, hashErr := chainhash.ChainHashCompute(chain)
		if hashErr != nil {
			report.Staleness = append(report.Staleness, &StalenessEntry{
				Node:         entry.logicalName,
				ArtifactPath: entry.fm.Output,
				Status:       "missing",
				Detail:       hashErr.Error(),
				Rank:         nodeRank,
			})
			continue
		}

		artifactPath := &pathutils.PathCfs{Value: entry.fm.Output}
		tag, tagErr := artifacttag.ArtifactTagExtract(artifactPath)
		if tagErr != nil {
			if errors.Is(tagErr, artifacttag.ErrFileUnreadable) {
				report.Staleness = append(report.Staleness, &StalenessEntry{
					Node:         entry.logicalName,
					ArtifactPath: entry.fm.Output,
					Status:       "missing",
					Detail:       tagErr.Error(),
					Rank:         nodeRank,
				})
				continue
			}

			if errors.Is(tagErr, artifacttag.ErrMalformedTag) {
				report.Staleness = append(report.Staleness, &StalenessEntry{
					Node:         entry.logicalName,
					ArtifactPath: entry.fm.Output,
					Status:       "malformed tag",
					Detail:       tagErr.Error(),
					Rank:         nodeRank,
				})
				continue
			}

			report.Staleness = append(report.Staleness, &StalenessEntry{
				Node:         entry.logicalName,
				ArtifactPath: entry.fm.Output,
				Status:       "malformed tag",
				Detail:       tagErr.Error(),
				Rank:         nodeRank,
			})
			continue
		}

		if tag.Hash != chainHash {
			report.Staleness = append(report.Staleness, &StalenessEntry{
				Node:         entry.logicalName,
				ArtifactPath: entry.fm.Output,
				Status:       "stale",
				Detail:       fmt.Sprintf("file hash %s does not match expected hash %s", tag.Hash, chainHash),
				Rank:         nodeRank,
			})
		}
	}

	sort.Slice(report.Staleness, func(i, j int) bool {
		if report.Staleness[i].Rank != report.Staleness[j].Rank {
			return report.Staleness[i].Rank < report.Staleness[j].Rank
		}
		return report.Staleness[i].Node < report.Staleness[j].Node
	})

	return report
}

func discoverAllDirs() []string {
	root, err := pathutils.PathGetProjectRoot()
	if err != nil {
		return []string{}
	}

	baseDir := filepath.Join(root.Value, "code-from-spec")

	var dirs []string
	_ = filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if path == baseDir {
			return nil
		}
		rel, relErr := filepath.Rel(root.Value, path)
		if relErr != nil {
			return nil
		}
		cfsPath := filepath.ToSlash(rel)
		dirs = append(dirs, cfsPath)
		return nil
	})

	if _, statErr := os.Stat(baseDir); statErr != nil {
		return []string{}
	}

	return dirs
}
