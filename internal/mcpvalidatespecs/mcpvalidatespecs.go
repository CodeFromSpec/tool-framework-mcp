// code-from-spec: ROOT/golang/implementation/mcp_tools/validate_specs@WaOLA2A6WaSTCegfCwalnlQT6FY
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
	Rank         int
	Detail       string
}

type ValidationReport struct {
	FormatErrors []*spectreevalidate.FormatError
	Cycles       []string
	Staleness    []*StalenessEntry
}

type parsedNode struct {
	fm   *frontmatter.Frontmatter
	node *parsenode.Node
}

func MCPValidateSpecs() *ValidationReport {
	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		return &ValidationReport{
			FormatErrors: []*spectreevalidate.FormatError{
				{Node: "", Rule: "scan", Detail: err.Error()},
			},
			Cycles:    []string{},
			Staleness: []*StalenessEntry{},
		}
	}

	parsedNodes := make(map[string]*parsedNode)
	var parseErrors []*spectreevalidate.FormatError

	for _, n := range nodes {
		fm, err := frontmatter.FrontmatterParse(&n.FilePath)
		if err != nil {
			parseErrors = append(parseErrors, &spectreevalidate.FormatError{
				Node:   n.LogicalName,
				Rule:   "parse",
				Detail: err.Error(),
			})
			continue
		}

		parsed, err := parsenode.NodeParse(n.LogicalName)
		if err != nil {
			parseErrors = append(parseErrors, &spectreevalidate.FormatError{
				Node:   n.LogicalName,
				Rule:   "parse",
				Detail: err.Error(),
			})
			continue
		}

		parsedNodes[n.LogicalName] = &parsedNode{fm: fm, node: parsed}
	}

	var validateEntries []*spectreevalidate.SpecTreeValidateInput
	for logicalName, pn := range parsedNodes {
		validateEntries = append(validateEntries, &spectreevalidate.SpecTreeValidateInput{
			LogicalName: logicalName,
			Frontmatter: pn.fm,
			Node:        pn.node,
		})
	}

	validateErrors := spectreevalidate.SpecTreeValidate(validateEntries)

	allFormatErrors := append(parseErrors, validateErrors...)

	rankedEntries := make(map[string]int)
	var cycles []string
	rankingAvailable := false

	if len(allFormatErrors) == 0 {
		var rankInputs []*noderanking.NodeRankInput
		for logicalName, pn := range parsedNodes {
			rankInputs = append(rankInputs, &noderanking.NodeRankInput{
				LogicalName: logicalName,
				Frontmatter: pn.fm,
			})
		}

		ranked, cycleNames, err := noderanking.NodeRankCompute(rankInputs)
		if err != nil {
			if errors.Is(err, noderanking.ErrUnresolvableReference) {
				allFormatErrors = append(allFormatErrors, &spectreevalidate.FormatError{
					Node:   "",
					Rule:   "ranking",
					Detail: err.Error(),
				})
			} else {
				allFormatErrors = append(allFormatErrors, &spectreevalidate.FormatError{
					Node:   "",
					Rule:   "ranking",
					Detail: err.Error(),
				})
			}
			cycles = []string{}
		} else {
			for _, entry := range ranked {
				rankedEntries[entry.LogicalName] = entry.Rank
			}
			cycles = cycleNames
			rankingAvailable = true
		}
	} else {
		cycles = []string{}
	}

	type workItem struct {
		logicalName string
		rank        int
	}

	var workList []workItem
	for logicalName, pn := range parsedNodes {
		if pn.fm.Output == "" {
			continue
		}
		rank := 0
		if rankingAvailable {
			if r, ok := rankedEntries[logicalName]; ok {
				rank = r
			}
		}
		workList = append(workList, workItem{logicalName: logicalName, rank: rank})
	}

	if rankingAvailable {
		sort.Slice(workList, func(i, j int) bool {
			if workList[i].rank != workList[j].rank {
				return workList[i].rank < workList[j].rank
			}
			return workList[i].logicalName < workList[j].logicalName
		})
	} else {
		sort.Slice(workList, func(i, j int) bool {
			return workList[i].logicalName < workList[j].logicalName
		})
	}

	var staleness []*StalenessEntry

	for _, item := range workList {
		pn := parsedNodes[item.logicalName]
		rank := item.rank

		chain, err := chainresolver.ChainResolve(item.logicalName)
		if err != nil {
			staleness = append(staleness, &StalenessEntry{
				Node:         item.logicalName,
				ArtifactPath: pn.fm.Output,
				Status:       "missing",
				Detail:       err.Error(),
				Rank:         rank,
			})
			continue
		}

		chainHash, err := chainhash.ChainHashCompute(chain)
		if err != nil {
			staleness = append(staleness, &StalenessEntry{
				Node:         item.logicalName,
				ArtifactPath: pn.fm.Output,
				Status:       "missing",
				Detail:       err.Error(),
				Rank:         rank,
			})
			continue
		}

		cfsPath := &pathutils.PathCfs{Value: pn.fm.Output}
		tag, err := artifacttag.ArtifactTagExtract(cfsPath)
		if err != nil {
			status := "malformed tag"
			if errors.Is(err, artifacttag.ErrFileUnreadable) {
				status = "missing"
			}
			staleness = append(staleness, &StalenessEntry{
				Node:         item.logicalName,
				ArtifactPath: pn.fm.Output,
				Status:       status,
				Detail:       err.Error(),
				Rank:         rank,
			})
			continue
		}

		if tag.Hash != chainHash {
			staleness = append(staleness, &StalenessEntry{
				Node:         item.logicalName,
				ArtifactPath: pn.fm.Output,
				Status:       "stale",
				Detail:       fmt.Sprintf("file hash %s, expected %s", tag.Hash, chainHash),
				Rank:         rank,
			})
		}
	}

	sort.Slice(staleness, func(i, j int) bool {
		if staleness[i].Rank != staleness[j].Rank {
			return staleness[i].Rank < staleness[j].Rank
		}
		return staleness[i].Node < staleness[j].Node
	})

	if allFormatErrors == nil {
		allFormatErrors = []*spectreevalidate.FormatError{}
	}
	if cycles == nil {
		cycles = []string{}
	}
	if staleness == nil {
		staleness = []*StalenessEntry{}
	}

	return &ValidationReport{
		FormatErrors: allFormatErrors,
		Cycles:       cycles,
		Staleness:    staleness,
	}
}
