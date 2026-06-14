// code-from-spec: ROOT/golang/implementation/mcp_tools/validate_specs@6kmZDT5CbyzvnOMB_IepkkJpZvg
package mcpvalidatespecs

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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

type parsedNodeEntry struct {
	logicalName string
	fm          *frontmatter.Frontmatter
	node        *parsenode.Node
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

	allDirs := collectAllDirs()

	var parseErrors []*spectreevalidate.FormatError
	var parsedEntries []*parsedNodeEntry

	for _, n := range nodes {
		fm, fmErr := frontmatter.FrontmatterParse(&n.FilePath)
		if fmErr != nil {
			parseErrors = append(parseErrors, &spectreevalidate.FormatError{
				Node:   n.LogicalName,
				Rule:   "parse",
				Detail: fmErr.Error(),
			})
			continue
		}

		parsed, npErr := parsenode.NodeParse(n.LogicalName)
		if npErr != nil {
			parseErrors = append(parseErrors, &spectreevalidate.FormatError{
				Node:   n.LogicalName,
				Rule:   "parse",
				Detail: npErr.Error(),
			})
			continue
		}

		parsedEntries = append(parsedEntries, &parsedNodeEntry{
			logicalName: n.LogicalName,
			fm:          fm,
			node:        parsed,
		})
	}

	validateInputs := make([]*spectreevalidate.SpecTreeValidateInput, 0, len(parsedEntries))
	for _, e := range parsedEntries {
		validateInputs = append(validateInputs, &spectreevalidate.SpecTreeValidateInput{
			LogicalName: e.logicalName,
			Frontmatter: e.fm,
			Node:        e.node,
		})
	}

	formatValidationErrors := spectreevalidate.SpecTreeValidate(validateInputs, allDirs)

	allFormatErrors := append(parseErrors, formatValidationErrors...)

	skipRanking := len(allFormatErrors) > 0
	var rankedEntries []*noderanking.NodeRankEntry
	var cycles []string
	rankingErrorPresent := false

	if !skipRanking {
		rankInputs := make([]*noderanking.NodeRankInput, 0, len(parsedEntries))
		for _, e := range parsedEntries {
			rankInputs = append(rankInputs, &noderanking.NodeRankInput{
				LogicalName: e.logicalName,
				Frontmatter: e.fm,
			})
		}

		ranked, cycleList, rankErr := noderanking.NodeRankCompute(rankInputs)
		if rankErr != nil {
			if errors.Is(rankErr, noderanking.ErrUnresolvableReference) {
				allFormatErrors = append(allFormatErrors, &spectreevalidate.FormatError{
					Node:   "",
					Rule:   "ranking",
					Detail: rankErr.Error(),
				})
				rankingErrorPresent = true
			} else {
				allFormatErrors = append(allFormatErrors, &spectreevalidate.FormatError{
					Node:   "",
					Rule:   "ranking",
					Detail: rankErr.Error(),
				})
				rankingErrorPresent = true
			}
		} else {
			rankedEntries = ranked
			cycles = cycleList
		}
	}

	rankMap := make(map[string]int)
	for _, r := range rankedEntries {
		rankMap[r.LogicalName] = r.Rank
	}

	type nodeWithRank struct {
		entry *parsedNodeEntry
		rank  int
	}

	var nodesWithOutput []nodeWithRank
	for _, e := range parsedEntries {
		if e.fm.Output == "" {
			continue
		}
		rank := 0
		if !rankingErrorPresent && !skipRanking {
			if r, ok := rankMap[e.logicalName]; ok {
				rank = r
			}
		}
		nodesWithOutput = append(nodesWithOutput, nodeWithRank{entry: e, rank: rank})
	}

	sort.Slice(nodesWithOutput, func(i, j int) bool {
		if nodesWithOutput[i].rank != nodesWithOutput[j].rank {
			return nodesWithOutput[i].rank < nodesWithOutput[j].rank
		}
		return nodesWithOutput[i].entry.logicalName < nodesWithOutput[j].entry.logicalName
	})

	var stalenessEntries []*StalenessEntry

	for _, nwr := range nodesWithOutput {
		e := nwr.entry
		rank := nwr.rank

		chain, resolveErr := chainresolver.ChainResolve(e.logicalName)
		if resolveErr != nil {
			stalenessEntries = append(stalenessEntries, &StalenessEntry{
				Node:         e.logicalName,
				ArtifactPath: e.fm.Output,
				Status:       "missing",
				Detail:       resolveErr.Error(),
				Rank:         rank,
			})
			continue
		}

		computedHash, hashErr := chainhash.ChainHashCompute(chain)
		if hashErr != nil {
			stalenessEntries = append(stalenessEntries, &StalenessEntry{
				Node:         e.logicalName,
				ArtifactPath: e.fm.Output,
				Status:       "missing",
				Detail:       hashErr.Error(),
				Rank:         rank,
			})
			continue
		}

		artifactPath := &pathutils.PathCfs{Value: e.fm.Output}
		tag, tagErr := artifacttag.ArtifactTagExtract(artifactPath)

		if tagErr != nil {
			if errors.Is(tagErr, artifacttag.ErrFileUnreadable) {
				stalenessEntries = append(stalenessEntries, &StalenessEntry{
					Node:         e.logicalName,
					ArtifactPath: e.fm.Output,
					Status:       "missing",
					Detail:       tagErr.Error(),
					Rank:         rank,
				})
			} else {
				stalenessEntries = append(stalenessEntries, &StalenessEntry{
					Node:         e.logicalName,
					ArtifactPath: e.fm.Output,
					Status:       "malformed tag",
					Detail:       tagErr.Error(),
					Rank:         rank,
				})
			}
			continue
		}

		if tag.Hash != computedHash {
			stalenessEntries = append(stalenessEntries, &StalenessEntry{
				Node:         e.logicalName,
				ArtifactPath: e.fm.Output,
				Status:       "stale",
				Detail:       fmt.Sprintf("file hash: %s, expected: %s", tag.Hash, computedHash),
				Rank:         rank,
			})
		}
	}

	sort.Slice(stalenessEntries, func(i, j int) bool {
		if stalenessEntries[i].Rank != stalenessEntries[j].Rank {
			return stalenessEntries[i].Rank < stalenessEntries[j].Rank
		}
		return stalenessEntries[i].Node < stalenessEntries[j].Node
	})

	if allFormatErrors == nil {
		allFormatErrors = []*spectreevalidate.FormatError{}
	}
	if cycles == nil {
		cycles = []string{}
	}
	if stalenessEntries == nil {
		stalenessEntries = []*StalenessEntry{}
	}

	return &ValidationReport{
		FormatErrors: allFormatErrors,
		Cycles:       cycles,
		Staleness:    stalenessEntries,
	}
}

func collectAllDirs() []string {
	root, err := pathutils.PathGetProjectRoot()
	if err != nil {
		return nil
	}

	cfsRoot := filepath.Join(root.Value, "code-from-spec")

	var dirs []string
	err = filepath.WalkDir(cfsRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() && path != cfsRoot {
			rel, relErr := filepath.Rel(root.Value, path)
			if relErr != nil {
				return nil
			}
			cfsPath := filepath.ToSlash(rel)
			dirs = append(dirs, cfsPath)
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
	}

	return dirs
}
