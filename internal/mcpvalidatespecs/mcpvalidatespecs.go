package mcpvalidatespecs

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"io"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/manifest"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/noderanking"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/parsing"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/spectree"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/spectreevalidate"
)

type StalenessEntry struct {
	Node         string
	ArtifactPath string
	Status       string
	Detail       string
	Rank         int
	Result       string
	Blocked      bool
	BlockedBy    string
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

	var knownSpecNodes []string
	for _, n := range successfulNodes {
		knownSpecNodes = append(knownSpecNodes, n.Reference.LogicalName)
	}

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
		if n.Frontmatter == nil || n.Frontmatter.Type == nil {
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

	blockedSet := make(map[string]string)

	for _, nwr := range nodesToProcess {
		n := nwr.node
		rank := nwr.rank
		resolvedOutput := parsing.ResolvedOutput(&n)
		artifactPath := ""
		if resolvedOutput != nil {
			artifactPath = *resolvedOutput
		}

		suffix := strings.TrimPrefix(n.Reference.LogicalName, "SPEC/")
		manifestKey := ""
		if n.Frontmatter != nil && n.Frontmatter.Type != nil && *n.Frontmatter.Type == "verdict" {
			manifestKey = "VERDICT/" + suffix
		} else {
			manifestKey = "ARTIFACT/" + suffix
		}

		var currentEntry *StalenessEntry

		chain, chainErr := chainresolver.ChainResolve(n.Reference.LogicalName, knownSpecNodes)
		if chainErr != nil {
			e := StalenessEntry{
				Node:         n.Reference.LogicalName,
				ArtifactPath: artifactPath,
				Status:       "missing",
				Detail:       chainErr.Error(),
				Rank:         rank,
			}
			currentEntry = &e
		} else {
			computedHash, _, hashErr := chainhash.ChainHashCompute(chain)
			if hashErr != nil {
				e := StalenessEntry{
					Node:         n.Reference.LogicalName,
					ArtifactPath: artifactPath,
					Status:       "missing",
					Detail:       hashErr.Error(),
					Rank:         rank,
				}
				currentEntry = &e
			} else {
				mEntry, entryExists := manifestEntries[manifestKey]
				if !entryExists {
					e := StalenessEntry{
						Node:         n.Reference.LogicalName,
						ArtifactPath: artifactPath,
						Status:       "missing",
						Detail:       "no manifest entry",
						Rank:         rank,
					}
					currentEntry = &e
				} else if mEntry.ChainHash != computedHash {
					e := StalenessEntry{
						Node:         n.Reference.LogicalName,
						ArtifactPath: artifactPath,
						Status:       "stale",
						Detail:       "manifest chain hash " + mEntry.ChainHash + " does not match expected hash " + computedHash,
						Rank:         rank,
						Result:       mEntry.Result,
					}
					currentEntry = &e
				} else {
					filePath := oslayer.CfsPath(artifactPath)
					handle, openErr := oslayer.OpenFile(filePath, "read", 30000)
					if openErr != nil {
						e := StalenessEntry{
							Node:         n.Reference.LogicalName,
							ArtifactPath: artifactPath,
							Status:       "missing",
							Detail:       openErr.Error(),
							Rank:         rank,
						}
						currentEntry = &e
					} else {
						fileChecksum, readErr := computeFileChecksum(handle)
						handle.Close()
						if readErr != nil {
							e := StalenessEntry{
								Node:         n.Reference.LogicalName,
								ArtifactPath: artifactPath,
								Status:       "missing",
								Detail:       readErr.Error(),
								Rank:         rank,
							}
							currentEntry = &e
						} else if fileChecksum != mEntry.Checksum {
							e := StalenessEntry{
								Node:         n.Reference.LogicalName,
								ArtifactPath: artifactPath,
								Status:       "modified",
								Detail:       "file checksum does not match manifest checksum",
								Rank:         rank,
								Result:       mEntry.Result,
							}
							currentEntry = &e
						}
					}
				}
			}
		}

		blocked, reason := computeBlocking(&n, manifestEntries, knownSpecNodes, blockedSet)

		if blocked {
			if currentEntry == nil {
				e := StalenessEntry{
					Node:         n.Reference.LogicalName,
					ArtifactPath: artifactPath,
					Status:       "",
					Rank:         rank,
					Blocked:      true,
					BlockedBy:    reason,
				}
				stalenessEntries = append(stalenessEntries, e)
			} else {
				currentEntry.Blocked = true
				currentEntry.BlockedBy = reason
				stalenessEntries = append(stalenessEntries, *currentEntry)
			}
			blockedSet[manifestKey] = reason
		} else if currentEntry != nil {
			stalenessEntries = append(stalenessEntries, *currentEntry)
		}
	}

	for artifactKey, entry := range manifestEntries {
		var specLogicalName string
		if strings.HasPrefix(artifactKey, "ARTIFACT/") {
			specLogicalName = "SPEC/" + strings.TrimPrefix(artifactKey, "ARTIFACT/")
		} else if strings.HasPrefix(artifactKey, "VERDICT/") {
			specLogicalName = "SPEC/" + strings.TrimPrefix(artifactKey, "VERDICT/")
		} else {
			continue
		}
		node, nodeExists := parsedNodes[specLogicalName]
		if !nodeExists || node.Frontmatter == nil || node.Frontmatter.Type == nil {
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

func computeBlocking(n *parsing.Node, manifestEntries map[string]manifest.ManifestEntry, knownSpecNodes []string, blockedSet map[string]string) (bool, string) {
	if n.Frontmatter == nil {
		return false, ""
	}

	declaringNode := n.Reference.LogicalName

	waitOnTargets := expandRefs(n.Frontmatter.WaitOn, knownSpecNodes, &declaringNode)
	for _, target := range waitOnTargets {
		if !isWaitOnTargetSatisfied(target, manifestEntries, knownSpecNodes) {
			return true, "wait_on target not satisfied: " + target
		}
	}

	allImportsInput := append(append([]string{}, n.Frontmatter.Imports...), n.Frontmatter.Input...)
	expandedDeps := expandRefs(allImportsInput, knownSpecNodes, &declaringNode)
	var artifactDeps []string
	for _, dep := range expandedDeps {
		if strings.HasPrefix(dep, "ARTIFACT/") {
			artifactDeps = append(artifactDeps, dep)
		}
	}
	for _, dep := range artifactDeps {
		entry, exists := manifestEntries[dep]
		if exists && !checksumMatchesFile(entry) {
			return true, "dependency artifact modified: " + dep
		}
	}

	for _, target := range waitOnTargets {
		if _, inBlockedSet := blockedSet[target]; inBlockedSet {
			return true, "dependency blocked: " + target
		}
	}
	for _, dep := range artifactDeps {
		if _, inBlockedSet := blockedSet[dep]; inBlockedSet {
			return true, "dependency blocked: " + dep
		}
	}

	return false, ""
}

func expandRefs(patterns []string, knownSpecNodes []string, declaringNode *string) []string {
	var result []string
	for _, pattern := range patterns {
		if strings.Contains(pattern, "*") {
			expanded, err := parsing.ExpandGlob(pattern, knownSpecNodes, declaringNode)
			if err == nil {
				result = append(result, expanded...)
			}
		} else {
			result = append(result, pattern)
		}
	}
	return result
}

func isWaitOnTargetSatisfied(target string, manifestEntries map[string]manifest.ManifestEntry, knownSpecNodes []string) bool {
	entry, exists := manifestEntries[target]
	if !exists {
		return false
	}

	var specName string
	if strings.HasPrefix(target, "ARTIFACT/") {
		specName = "SPEC/" + strings.TrimPrefix(target, "ARTIFACT/")
	} else if strings.HasPrefix(target, "VERDICT/") {
		specName = "SPEC/" + strings.TrimPrefix(target, "VERDICT/")
	} else {
		return false
	}

	chain, err := chainresolver.ChainResolve(specName, knownSpecNodes)
	if err != nil {
		return false
	}

	computedHash, _, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		return false
	}

	if entry.ChainHash != computedHash {
		return false
	}

	if !checksumMatchesFile(entry) {
		return false
	}

	if strings.HasPrefix(target, "VERDICT/") {
		if entry.Result != "pass" && entry.Result != "accepted" {
			return false
		}
	}

	return true
}

func checksumMatchesFile(entry manifest.ManifestEntry) bool {
	filePath := oslayer.CfsPath(entry.Path)
	handle, err := oslayer.OpenFile(filePath, "read", 30000)
	if err != nil {
		return true
	}
	checksum, err := computeFileChecksum(handle)
	handle.Close()
	if err != nil {
		return true
	}
	return checksum == entry.Checksum
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
