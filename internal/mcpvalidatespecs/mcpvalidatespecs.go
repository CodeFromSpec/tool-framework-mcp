// code-from-spec: ROOT/golang/implementation/mcp_tools/validate_specs@sjrVLbThfNAMIkkgKmKMtzza7JE
package mcpvalidatespecs

import (
	"errors"
	"fmt"
	"sort"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/artifacttag"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/noderanking"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectree"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectreevalidate"
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

type parsedNode struct {
	logicalName string
	filePath    pathutils.PathCfs
	fm          frontmatter.Frontmatter
	node        parsenode.Node
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

	var parsed []*parsedNode
	for _, n := range nodes {
		fp := n.FilePath
		fm, fmErr := frontmatter.FrontmatterParse(&fp)
		node, nodeErr := parsenode.NodeParse(n.LogicalName)

		if fmErr != nil || nodeErr != nil {
			detail := ""
			if fmErr != nil {
				detail = fmErr.Error()
			} else {
				detail = nodeErr.Error()
			}
			report.FormatErrors = append(report.FormatErrors, &spectreevalidate.FormatError{
				Node:   n.LogicalName,
				Rule:   "parse",
				Detail: detail,
			})
			continue
		}

		parsed = append(parsed, &parsedNode{
			logicalName: n.LogicalName,
			filePath:    n.FilePath,
			fm:          *fm,
			node:        *node,
		})
	}

	validateInputs := make([]*spectreevalidate.SpecTreeValidateInput, 0, len(parsed))
	for _, p := range parsed {
		validateInputs = append(validateInputs, &spectreevalidate.SpecTreeValidateInput{
			LogicalName: p.logicalName,
			Frontmatter: p.fm,
			Node:        p.node,
		})
	}

	formatErrs := spectreevalidate.SpecTreeValidate(validateInputs)
	report.FormatErrors = append(report.FormatErrors, formatErrs...)

	rankByName := map[string]int{}
	rankingAvailable := false

	if len(report.FormatErrors) == 0 {
		rankInputs := make([]*noderanking.NodeRankInput, 0, len(parsed))
		for _, p := range parsed {
			rankInputs = append(rankInputs, &noderanking.NodeRankInput{
				LogicalName: p.logicalName,
				Frontmatter: p.fm,
			})
		}

		ranked, cycles, rankErr := noderanking.NodeRankCompute(rankInputs)
		if rankErr != nil {
			report.FormatErrors = append(report.FormatErrors, &spectreevalidate.FormatError{
				Node:   "",
				Rule:   "ranking",
				Detail: rankErr.Error(),
			})
		} else {
			report.Cycles = cycles
			for _, r := range ranked {
				rankByName[r.LogicalName] = r.Rank
			}
			rankingAvailable = true
		}
	}

	type stalenessCandidate struct {
		logicalName  string
		outputPath   string
		rank         int
	}

	var candidates []stalenessCandidate
	for _, p := range parsed {
		if p.fm.Output == "" {
			continue
		}
		rank := 0
		if rankingAvailable {
			rank = rankByName[p.logicalName]
		}
		candidates = append(candidates, stalenessCandidate{
			logicalName: p.logicalName,
			outputPath:  p.fm.Output,
			rank:        rank,
		})
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].rank != candidates[j].rank {
			return candidates[i].rank < candidates[j].rank
		}
		return candidates[i].logicalName < candidates[j].logicalName
	})

	for _, c := range candidates {
		chain, resolveErr := chainresolver.ChainResolve(c.logicalName)
		if resolveErr != nil {
			report.Staleness = append(report.Staleness, &StalenessEntry{
				Node:         c.logicalName,
				ArtifactPath: c.outputPath,
				Status:       "missing",
				Detail:       resolveErr.Error(),
				Rank:         c.rank,
			})
			continue
		}

		chainHash, hashErr := chainhash.ChainHashCompute(chain)
		if hashErr != nil {
			report.Staleness = append(report.Staleness, &StalenessEntry{
				Node:         c.logicalName,
				ArtifactPath: c.outputPath,
				Status:       "missing",
				Detail:       hashErr.Error(),
				Rank:         c.rank,
			})
			continue
		}

		artifactPath := &pathutils.PathCfs{Value: c.outputPath}
		tag, tagErr := artifacttag.ArtifactTagExtract(artifactPath)

		if tagErr != nil {
			if errors.Is(tagErr, artifacttag.ErrFileUnreadable) {
				report.Staleness = append(report.Staleness, &StalenessEntry{
					Node:         c.logicalName,
					ArtifactPath: c.outputPath,
					Status:       "missing",
					Detail:       tagErr.Error(),
					Rank:         c.rank,
				})
			} else {
				report.Staleness = append(report.Staleness, &StalenessEntry{
					Node:         c.logicalName,
					ArtifactPath: c.outputPath,
					Status:       "malformed tag",
					Detail:       tagErr.Error(),
					Rank:         c.rank,
				})
			}
			continue
		}

		if tag.Hash != chainHash {
			report.Staleness = append(report.Staleness, &StalenessEntry{
				Node:         c.logicalName,
				ArtifactPath: c.outputPath,
				Status:       "stale",
				Detail:       fmt.Sprintf("file hash %s does not match expected hash %s", tag.Hash, chainHash),
				Rank:         c.rank,
			})
		}
	}

	return report
}
