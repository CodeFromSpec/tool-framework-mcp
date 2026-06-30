package mcpvalidatespecs

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"io"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/manifest"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/noderanking"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/spectree"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/spectreevalidate"
)

type StalenessEntry struct {
	Node         string
	ArtifactPath string
	Status       string
	Detail       string
	Rank         int
}

type ValidationReport struct {
	FormatErrors []spectreevalidate.FormatError
	Cycles       []string
	Staleness    []StalenessEntry
}

func MCPValidateSpecs() ValidationReport {
	var formatErrors []spectreevalidate.FormatError
	cycles := []string{}
	var stalenessEntries []StalenessEntry

	refs, err := spectree.SpecTreeScan()
	if err != nil {
		return ValidationReport{
			FormatErrors: []spectreevalidate.FormatError{
				{Node: "", Rule: "scan", Detail: err.Error()},
			},
			Cycles:    []string{},
			Staleness: []StalenessEntry{},
		}
	}

	allDirs := collectAllDirs()

	parsedNodes := make(map[string]parsing.Node)

	for _, ref := range refs {
		node, parseErr := parsing.ParseNode(ref.LogicalName)
		if parseErr != nil {
			formatErrors = append(formatErrors, spectreevalidate.FormatError{
				Node:   ref.LogicalName,
				Rule:   "parse",
				Detail: parseErr.Error(),
			})
			continue
		}
		parsedNodes[node.Reference.LogicalName] = *node
	}

	var successfulNodes []parsing.Node
	for _, n := range parsedNodes {
		successfulNodes = append(successfulNodes, n)
	}
	sort.Slice(successfulNodes, func(i, j int) bool {
		return successfulNodes[i].Reference.LogicalName < successfulNodes[j].Reference.LogicalName
	})

	validateErrors := spectreevalidate.SpecTreeValidate(successfulNodes, allDirs)
	formatErrors = append(formatErrors, validateErrors...)

	var rankedEntries []noderanking.NodeRankEntry
	rankMap := make(map[string]int)

	if len(formatErrors) == 0 {
		ranked, rankCycles, rankErr := noderanking.NodeRankCompute(successfulNodes)
		if rankErr != nil {
			formatErrors = append(formatErrors, spectreevalidate.FormatError{
				Node:   "",
				Rule:   "ranking",
				Detail: rankErr.Error(),
			})
		} else {
			rankedEntries = ranked
			cycles = rankCycles
			if cycles == nil {
				cycles = []string{}
			}
			for _, entry := range rankedEntries {
				rankMap[entry.Reference.LogicalName] = entry.Rank
			}
		}
	}

	manifestEntries := make(map[string]manifest.ManifestEntry)
	m, manifestErr := manifest.OpenManifest(true)
	if manifestErr == nil && m != nil {
		for k, v := range m.Entries {
			manifestEntries[k] = v
		}
	}

	type nodeWithRank struct {
		node parsing.Node
		rank int
	}

	var nodesToProcess []nodeWithRank
	for _, n := range successfulNodes {
		if n.Frontmatter == nil || n.Frontmatter.Output == nil || *n.Frontmatter.Output == "" {
			continue
		}
		rank := 0
		if r, ok := rankMap[n.Reference.LogicalName]; ok {
			rank = r
		}
		nodesToProcess = append(nodesToProcess, nodeWithRank{node: n, rank: rank})
	}

	if len(rankedEntries) > 0 {
		sort.Slice(nodesToProcess, func(i, j int) bool {
			if nodesToProcess[i].rank != nodesToProcess[j].rank {
				return nodesToProcess[i].rank < nodesToProcess[j].rank
			}
			return nodesToProcess[i].node.Reference.LogicalName < nodesToProcess[j].node.Reference.LogicalName
		})
	} else {
		sort.Slice(nodesToProcess, func(i, j int) bool {
			return nodesToProcess[i].node.Reference.LogicalName < nodesToProcess[j].node.Reference.LogicalName
		})
	}

	for _, nwr := range nodesToProcess {
		n := nwr.node
		rank := nwr.rank
		outputPath := *n.Frontmatter.Output
		artifactLogicalName := "ARTIFACT/" + strings.TrimPrefix(n.Reference.LogicalName, "SPEC/")

		chain, chainErr := chainresolver.ChainResolve(n.Reference.LogicalName)
		if chainErr != nil {
			stalenessEntries = append(stalenessEntries, StalenessEntry{
				Node:         n.Reference.LogicalName,
				ArtifactPath: outputPath,
				Status:       "missing",
				Detail:       chainErr.Error(),
				Rank:         rank,
			})
			continue
		}

		computedHash, hashErr := chainhash.ChainHashCompute(chain)
		if hashErr != nil {
			stalenessEntries = append(stalenessEntries, StalenessEntry{
				Node:         n.Reference.LogicalName,
				ArtifactPath: outputPath,
				Status:       "missing",
				Detail:       hashErr.Error(),
				Rank:         rank,
			})
			continue
		}

		entry, entryExists := manifestEntries[artifactLogicalName]
		if !entryExists {
			stalenessEntries = append(stalenessEntries, StalenessEntry{
				Node:         n.Reference.LogicalName,
				ArtifactPath: outputPath,
				Status:       "missing",
				Detail:       "no manifest entry",
				Rank:         rank,
			})
			continue
		}

		if entry.ChainHash != computedHash {
			stalenessEntries = append(stalenessEntries, StalenessEntry{
				Node:         n.Reference.LogicalName,
				ArtifactPath: outputPath,
				Status:       "stale",
				Detail:       "manifest chain hash " + entry.ChainHash + " does not match expected hash " + computedHash,
				Rank:         rank,
			})
			continue
		}

		filePath := oslayer.CfsPath(outputPath)
		handle, openErr := oslayer.OpenFile(filePath, "read", 30000)
		if openErr != nil {
			stalenessEntries = append(stalenessEntries, StalenessEntry{
				Node:         n.Reference.LogicalName,
				ArtifactPath: outputPath,
				Status:       "missing",
				Detail:       openErr.Error(),
				Rank:         rank,
			})
			continue
		}

		fileChecksum, readErr := computeFileChecksum(handle)
		handle.Close()
		if readErr != nil {
			stalenessEntries = append(stalenessEntries, StalenessEntry{
				Node:         n.Reference.LogicalName,
				ArtifactPath: outputPath,
				Status:       "missing",
				Detail:       readErr.Error(),
				Rank:         rank,
			})
			continue
		}

		if fileChecksum != entry.Checksum {
			stalenessEntries = append(stalenessEntries, StalenessEntry{
				Node:         n.Reference.LogicalName,
				ArtifactPath: outputPath,
				Status:       "modified",
				Detail:       "file checksum does not match manifest checksum",
				Rank:         rank,
			})
		}
	}

	for artifactKey, entry := range manifestEntries {
		if !strings.HasPrefix(artifactKey, "ARTIFACT/") {
			continue
		}
		specLogicalName := "SPEC/" + strings.TrimPrefix(artifactKey, "ARTIFACT/")
		node, nodeExists := parsedNodes[specLogicalName]
		if !nodeExists || node.Frontmatter == nil || node.Frontmatter.Output == nil || *node.Frontmatter.Output == "" {
			stalenessEntries = append(stalenessEntries, StalenessEntry{
				Node:         artifactKey,
				ArtifactPath: entry.Path,
				Status:       "orphan",
				Detail:       "manifest entry has no corresponding spec node",
				Rank:         0,
			})
		}
	}

	sort.Slice(stalenessEntries, func(i, j int) bool {
		if stalenessEntries[i].Rank != stalenessEntries[j].Rank {
			return stalenessEntries[i].Rank < stalenessEntries[j].Rank
		}
		return stalenessEntries[i].Node < stalenessEntries[j].Node
	})

	if formatErrors == nil {
		formatErrors = []spectreevalidate.FormatError{}
	}
	if stalenessEntries == nil {
		stalenessEntries = []StalenessEntry{}
	}

	return ValidationReport{
		FormatErrors: formatErrors,
		Cycles:       cycles,
		Staleness:    stalenessEntries,
	}
}

func computeFileChecksum(handle *oslayer.File) (string, error) {
	hasher := sha1.New()
	for {
		line, readErr := handle.ReadLine()
		if readErr != nil {
			if errors.Is(readErr, oslayer.ErrEndOfFile) {
				break
			}
			return "", readErr
		}
		io.WriteString(hasher, line+"\n")
	}
	return base64.RawURLEncoding.EncodeToString(hasher.Sum(nil)), nil
}

func collectAllDirs() []string {
	files, err := oslayer.ListAllFiles("code-from-spec")
	if err != nil {
		return nil
	}

	dirSet := make(map[string]bool)
	for _, f := range files {
		path := string(f)
		for i, c := range path {
			if c == '/' {
				dirSet[path[:i]] = true
			}
		}
	}

	var dirs []string
	for d := range dirSet {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	return dirs
}
