// code-from-spec: SPEC/golang/implementation/mcp_tools/validate_specs@f3RPl6I1UGR28TACUYT6FM2WZlw
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

// StalenessEntry describes the staleness status of a single output artifact.
type StalenessEntry struct {
	Node         string
	ArtifactPath string
	Status       string
	Detail       string
	Rank         int
}

// ValidationReport holds the full result of a spec-tree validation run.
type ValidationReport struct {
	FormatErrors []*spectreevalidate.FormatError
	Cycles       []string
	Staleness    []*StalenessEntry
}

// MCPValidateSpecs scans the entire spec tree starting from code-from-spec/,
// validates all nodes against format rules, detects dependency cycles, and
// checks whether each output artifact is up to date. Always returns a report —
// never returns an error. Problems are collected in the report fields.
//
// StalenessEntry.Status is one of:
//   - "missing"       — file does not exist.
//   - "stale"         — hash mismatch.
//   - "malformed tag" — file exists but has no artifact tag or the tag cannot be parsed.
//
// Entries whose hash matches are not included in Staleness.
// Cycles contains logical names involved in non-convergence during ranking.
// StalenessEntry.Rank is the rank from NodeRankCompute; entries with equal rank
// have no dependency between them and can be processed in parallel.
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

	allDirs, err := collectSubdirs("code-from-spec")
	if err != nil {
		report.FormatErrors = append(report.FormatErrors, &spectreevalidate.FormatError{
			Node:   "",
			Rule:   "scan",
			Detail: err.Error(),
		})
		return report
	}

	type parsedEntry struct {
		logicalName string
		fm          *frontmatter.Frontmatter
		node        *parsenode.Node
	}

	parsed := make([]*parsedEntry, 0, len(nodes))
	parseFailed := make(map[string]bool)

	for _, n := range nodes {
		fm, fmErr := frontmatter.FrontmatterParse(&n.FilePath)
		if fmErr != nil {
			report.FormatErrors = append(report.FormatErrors, &spectreevalidate.FormatError{
				Node:   n.LogicalName,
				Rule:   "parse",
				Detail: fmErr.Error(),
			})
			parseFailed[n.LogicalName] = true
			continue
		}

		pn, pnErr := parsenode.NodeParse(n.LogicalName)
		if pnErr != nil {
			report.FormatErrors = append(report.FormatErrors, &spectreevalidate.FormatError{
				Node:   n.LogicalName,
				Rule:   "parse",
				Detail: pnErr.Error(),
			})
			parseFailed[n.LogicalName] = true
			continue
		}

		parsed = append(parsed, &parsedEntry{
			logicalName: n.LogicalName,
			fm:          fm,
			node:        pn,
		})
	}

	validateInputs := make([]*spectreevalidate.SpecTreeValidateInput, 0, len(parsed))
	for _, e := range parsed {
		validateInputs = append(validateInputs, &spectreevalidate.SpecTreeValidateInput{
			LogicalName: e.logicalName,
			Frontmatter: e.fm,
			Node:        e.node,
		})
	}

	formatErrs := spectreevalidate.SpecTreeValidate(validateInputs, allDirs)
	report.FormatErrors = append(report.FormatErrors, formatErrs...)

	rankedEntries := []*noderanking.NodeRankEntry{}
	cycles := []string{}

	if len(report.FormatErrors) == 0 {
		rankInputs := make([]*noderanking.NodeRankInput, 0, len(parsed))
		for _, e := range parsed {
			rankInputs = append(rankInputs, &noderanking.NodeRankInput{
				LogicalName: e.logicalName,
				Frontmatter: e.fm,
			})
		}

		ranked, detectedCycles, rankErr := noderanking.NodeRankCompute(rankInputs)
		if rankErr != nil {
			report.FormatErrors = append(report.FormatErrors, &spectreevalidate.FormatError{
				Node:   "",
				Rule:   "ranking",
				Detail: rankErr.Error(),
			})
		} else {
			rankedEntries = ranked
			cycles = detectedCycles
		}
	}

	report.Cycles = cycles

	rankByName := make(map[string]int, len(rankedEntries))
	for _, re := range rankedEntries {
		rankByName[re.LogicalName] = re.Rank
	}

	type orderedNode struct {
		logicalName string
		fm          *frontmatter.Frontmatter
		rank        int
	}

	var toCheck []*orderedNode
	for _, e := range parsed {
		if e.fm.Output == "" {
			continue
		}
		rank := 0
		if r, ok := rankByName[e.logicalName]; ok {
			rank = r
		}
		toCheck = append(toCheck, &orderedNode{
			logicalName: e.logicalName,
			fm:          e.fm,
			rank:        rank,
		})
	}

	if len(rankedEntries) > 0 {
		sort.Slice(toCheck, func(i, j int) bool {
			if toCheck[i].rank != toCheck[j].rank {
				return toCheck[i].rank < toCheck[j].rank
			}
			return toCheck[i].logicalName < toCheck[j].logicalName
		})
	} else {
		sort.Slice(toCheck, func(i, j int) bool {
			return toCheck[i].logicalName < toCheck[j].logicalName
		})
	}

	var stalenessEntries []*StalenessEntry

	for _, on := range toCheck {
		nodeRank := on.rank

		chain, chainErr := chainresolver.ChainResolve(on.logicalName)
		if chainErr != nil {
			stalenessEntries = append(stalenessEntries, &StalenessEntry{
				Node:         on.logicalName,
				ArtifactPath: on.fm.Output,
				Status:       "missing",
				Detail:       chainErr.Error(),
				Rank:         nodeRank,
			})
			continue
		}

		expectedHash, hashErr := chainhash.ChainHashCompute(chain)
		if hashErr != nil {
			stalenessEntries = append(stalenessEntries, &StalenessEntry{
				Node:         on.logicalName,
				ArtifactPath: on.fm.Output,
				Status:       "missing",
				Detail:       hashErr.Error(),
				Rank:         nodeRank,
			})
			continue
		}

		outputPath := &pathutils.PathCfs{Value: on.fm.Output}
		tag, tagErr := artifacttag.ArtifactTagExtract(outputPath)
		if tagErr != nil {
			if errors.Is(tagErr, artifacttag.ErrNoTagFound) || errors.Is(tagErr, artifacttag.ErrMalformedTag) {
				stalenessEntries = append(stalenessEntries, &StalenessEntry{
					Node:         on.logicalName,
					ArtifactPath: on.fm.Output,
					Status:       "malformed tag",
					Detail:       tagErr.Error(),
					Rank:         nodeRank,
				})
			} else {
				stalenessEntries = append(stalenessEntries, &StalenessEntry{
					Node:         on.logicalName,
					ArtifactPath: on.fm.Output,
					Status:       "missing",
					Detail:       tagErr.Error(),
					Rank:         nodeRank,
				})
			}
			continue
		}

		if tag.Hash != expectedHash {
			stalenessEntries = append(stalenessEntries, &StalenessEntry{
				Node:         on.logicalName,
				ArtifactPath: on.fm.Output,
				Status:       "stale",
				Detail:       fmt.Sprintf("file hash %s does not match expected hash %s", tag.Hash, expectedHash),
				Rank:         nodeRank,
			})
		}
	}

	sort.Slice(stalenessEntries, func(i, j int) bool {
		if stalenessEntries[i].Rank != stalenessEntries[j].Rank {
			return stalenessEntries[i].Rank < stalenessEntries[j].Rank
		}
		return stalenessEntries[i].Node < stalenessEntries[j].Node
	})

	report.Staleness = stalenessEntries

	_ = parseFailed

	return report
}

// collectSubdirs returns all subdirectory paths (as strings) under root,
// including root itself.
func collectSubdirs(root string) ([]string, error) {
	var dirs []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			dirs = append(dirs, filepath.ToSlash(path))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("collecting subdirs under %s: %w", root, err)
	}
	return dirs, nil
}
