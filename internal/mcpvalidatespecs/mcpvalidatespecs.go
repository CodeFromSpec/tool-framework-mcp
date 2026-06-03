// code-from-spec: ROOT/golang/implementation/mcp_tools/validate_specs@wUzttLPR3F4IvzG6sgzwbAVOzdQ
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

	parsed := make([]*parsedNode, 0, len(nodes))
	for _, n := range nodes {
		fm, fmErr := frontmatter.FrontmatterParse(&n.FilePath)
		nd, ndErr := parsenode.NodeParse(n.LogicalName)
		if fmErr != nil {
			report.FormatErrors = append(report.FormatErrors, &spectreevalidate.FormatError{
				Node:   n.LogicalName,
				Rule:   "parse",
				Detail: fmErr.Error(),
			})
			continue
		}
		if ndErr != nil {
			report.FormatErrors = append(report.FormatErrors, &spectreevalidate.FormatError{
				Node:   n.LogicalName,
				Rule:   "parse",
				Detail: ndErr.Error(),
			})
			continue
		}
		parsed = append(parsed, &parsedNode{
			logicalName: n.LogicalName,
			fm:          fm,
			node:        nd,
		})
	}

	validateInputs := make([]*spectreevalidate.SpecTreeValidateInput, 0, len(parsed))
	for _, p := range parsed {
		validateInputs = append(validateInputs, &spectreevalidate.SpecTreeValidateInput{
			LogicalName: p.logicalName,
			Frontmatter: *p.fm,
			Node:        *p.node,
		})
	}
	formatErrs := spectreevalidate.SpecTreeValidate(validateInputs)
	report.FormatErrors = append(report.FormatErrors, formatErrs...)

	rankMap := make(map[string]int)
	skipRanking := len(report.FormatErrors) > 0

	if !skipRanking {
		rankInputs := make([]*noderanking.NodeRankInput, 0, len(parsed))
		for _, p := range parsed {
			rankInputs = append(rankInputs, &noderanking.NodeRankInput{
				LogicalName: p.logicalName,
				Frontmatter: p.fm,
			})
		}

		ranked, cycles, rankErr := noderanking.NodeRankCompute(rankInputs)
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
			report.Cycles = cycles
			for _, entry := range ranked {
				rankMap[entry.LogicalName] = entry.Rank
			}
		}
	}

	type stalenessCandidate struct {
		p    *parsedNode
		rank int
	}

	candidates := make([]stalenessCandidate, 0)
	for _, p := range parsed {
		if p.fm.Output == "" {
			continue
		}
		rank := 0
		if r, ok := rankMap[p.logicalName]; ok {
			rank = r
		}
		candidates = append(candidates, stalenessCandidate{p: p, rank: rank})
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].rank != candidates[j].rank {
			return candidates[i].rank < candidates[j].rank
		}
		return candidates[i].p.logicalName < candidates[j].p.logicalName
	})

	for _, c := range candidates {
		p := c.p
		rank := c.rank

		chain, chainErr := chainresolver.ChainResolve(p.logicalName)
		if chainErr != nil {
			report.Staleness = append(report.Staleness, &StalenessEntry{
				Node:         p.logicalName,
				ArtifactPath: p.fm.Output,
				Status:       "missing",
				Detail:       chainErr.Error(),
				Rank:         rank,
			})
			continue
		}

		chainHash, hashErr := chainhash.ChainHashCompute(chain)
		if hashErr != nil {
			report.Staleness = append(report.Staleness, &StalenessEntry{
				Node:         p.logicalName,
				ArtifactPath: p.fm.Output,
				Status:       "missing",
				Detail:       hashErr.Error(),
				Rank:         rank,
			})
			continue
		}

		artifactPath := &pathutils.PathCfs{Value: p.fm.Output}
		tag, tagErr := artifacttag.ArtifactTagExtract(artifactPath)
		if tagErr != nil {
			status := "malformed tag"
			if errors.Is(tagErr, artifacttag.ErrFileUnreadable) {
				status = "missing"
			}
			report.Staleness = append(report.Staleness, &StalenessEntry{
				Node:         p.logicalName,
				ArtifactPath: p.fm.Output,
				Status:       status,
				Detail:       tagErr.Error(),
				Rank:         rank,
			})
			continue
		}

		if tag.Hash != chainHash {
			report.Staleness = append(report.Staleness, &StalenessEntry{
				Node:         p.logicalName,
				ArtifactPath: p.fm.Output,
				Status:       "stale",
				Detail:       fmt.Sprintf("file hash %s does not match expected hash %s", tag.Hash, chainHash),
				Rank:         rank,
			})
		}
	}

	return report
}
