// code-from-spec: ROOT/golang/internal/tools/validate_specs/code@PENDING
package validate_specs

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/artifacttag"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/formatvalidation"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/nodediscovery"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/noderanking"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/parsenode"
	"github.com/goccy/go-yaml"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ValidateSpecsArgs defines the input parameters for the validate_specs tool.
// No input parameters; scans the entire spec tree.
type ValidateSpecsArgs struct{}

// FormatErrorEntry represents a single format error in the report.
type FormatErrorEntry struct {
	Node    string `yaml:"node"`
	Rule    string `yaml:"rule,omitempty"`
	Message string `yaml:"message,omitempty"`
	Detail  string `yaml:"detail,omitempty"`
}

// StalenessEntry represents a staleness finding for an artifact.
type StalenessEntry struct {
	Node         string `yaml:"node"`
	ArtifactPath string `yaml:"artifact_path"`
	Status       string `yaml:"status"`
}

// ValidationReport is the YAML-serializable validation result.
type ValidationReport struct {
	FormatErrors       []FormatErrorEntry `yaml:"format_errors"`
	CircularReferences [][]string         `yaml:"circular_references"`
	Staleness          []StalenessEntry   `yaml:"staleness"`
}

// HandleValidateSpecs discovers all spec nodes, validates format,
// detects circular references, and checks artifact staleness.
func HandleValidateSpecs(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args ValidateSpecsArgs,
) (*mcp.CallToolResult, any, error) {
	// Step 1: Discover nodes.
	nodes, err := nodediscovery.DiscoverNodes()
	if err != nil {
		return toolError(fmt.Sprintf("node discovery failure: %v", err)), nil, nil
	}

	// Step 2: Parse all nodes (cache frontmatter and parsed bodies).
	fmCache := make(map[string]*frontmatter.Frontmatter)
	parsedCache := make(map[string]*parsenode.ParsedNode)
	var parseErrors []FormatErrorEntry

	for _, node := range nodes {
		fm, err := frontmatter.ParseFrontmatter(node.FilePath)
		if err != nil {
			parseErrors = append(parseErrors, FormatErrorEntry{
				Node:    node.LogicalName,
				Message: fmt.Sprintf("unreadable file: %v", err),
			})
			continue
		}
		fmCache[node.LogicalName] = fm

		parsed, err := parsenode.ParseNode(node.LogicalName)
		if err != nil {
			parseErrors = append(parseErrors, FormatErrorEntry{
				Node:    node.LogicalName,
				Message: fmt.Sprintf("parse failure: %v", err),
			})
			continue
		}
		parsedCache[node.LogicalName] = parsed
	}

	// Step 3: Format validation.
	var formatErrors []FormatErrorEntry
	formatErrors = append(formatErrors, parseErrors...)

	fmtErrors, err := formatvalidation.ValidateFormat(nodes)
	if err != nil {
		// Non-fatal: include as a format error.
		formatErrors = append(formatErrors, FormatErrorEntry{
			Node:    "",
			Message: fmt.Sprintf("format validation error: %v", err),
		})
	}
	for _, fe := range fmtErrors {
		formatErrors = append(formatErrors, FormatErrorEntry{
			Node:   fe.Node,
			Rule:   fe.Rule,
			Detail: fe.Detail,
		})
	}

	// Step 4: Ranking and cycle detection.
	ranked, cycles, err := noderanking.DetectCycles(nodes)
	if err != nil && !errors.Is(err, noderanking.ErrUnresolvableRef) {
		// Non-fatal for unresolvable refs.
		formatErrors = append(formatErrors, FormatErrorEntry{
			Node:    "",
			Message: fmt.Sprintf("ranking error: %v", err),
		})
	}

	var circularRefs [][]string
	if len(cycles) > 0 {
		circularRefs = append(circularRefs, cycles)
	}

	// Build rank map for ordering.
	rankMap := make(map[string]int)
	for _, r := range ranked {
		rankMap[r.LogicalName] = r.Rank
	}

	// Step 5: Staleness detection.
	// Collect nodes with outputs, sorted by rank.
	type nodeWithRank struct {
		logicalName string
		fm          *frontmatter.Frontmatter
		rank        int
	}
	var outputNodes []nodeWithRank
	for _, node := range nodes {
		fm, ok := fmCache[node.LogicalName]
		if !ok || len(fm.Outputs) == 0 {
			continue
		}
		rank, ok := rankMap[node.LogicalName]
		if !ok {
			rank = 999999 // Unranked nodes go last.
		}
		outputNodes = append(outputNodes, nodeWithRank{
			logicalName: node.LogicalName,
			fm:          fm,
			rank:        rank,
		})
	}
	sort.Slice(outputNodes, func(i, j int) bool {
		return outputNodes[i].rank < outputNodes[j].rank
	})

	var staleness []StalenessEntry
	for _, on := range outputNodes {
		chainHash := computeChainHash(on.logicalName, on.fm, fmCache, parsedCache)

		for _, out := range on.fm.Outputs {
			tag, err := artifacttag.ExtractArtifactTag(out.Path)
			if err != nil {
				// File missing or no tag.
				staleness = append(staleness, StalenessEntry{
					Node:         on.logicalName,
					ArtifactPath: out.Path,
					Status:       "missing",
				})
				continue
			}
			if tag.Hash != chainHash {
				staleness = append(staleness, StalenessEntry{
					Node:         on.logicalName,
					ArtifactPath: out.Path,
					Status:       "stale",
				})
			}
		}
	}

	// Step 6: Build report.
	report := ValidationReport{
		FormatErrors:       formatErrors,
		CircularReferences: circularRefs,
		Staleness:          staleness,
	}

	yamlBytes, err := yaml.Marshal(report)
	if err != nil {
		return toolError(fmt.Sprintf("report serialization failure: %v", err)), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(yamlBytes)}},
	}, nil, nil
}

// computeChainHash computes the chain hash for a node using the same algorithm
// as load_chain. This is a simplified version that handles common cases.
func computeChainHash(
	logicalName string,
	fm *frontmatter.Frontmatter,
	fmCache map[string]*frontmatter.Frontmatter,
	parsedCache map[string]*parsenode.ParsedNode,
) string {
	var hashParts [][]byte

	// Ancestors: walk up from target to root.
	var ancestors []string
	current := logicalName
	for {
		parent, ok := logicalnames.ParentLogicalName(current)
		if !ok {
			break
		}
		ancestors = append([]string{parent}, ancestors...)
		current = parent
	}

	for _, ancestor := range ancestors {
		parsed, ok := parsedCache[ancestor]
		if !ok || parsed.Public == nil {
			continue
		}
		publicContent := sectionWithHeading(parsed.Public)
		if strings.TrimSpace(publicContent) == "" {
			continue
		}
		hash := sha1.Sum([]byte(publicContent))
		hashParts = append(hashParts, hash[:])
	}

	// Dependencies: sorted alphabetically.
	if len(fm.DependsOn) > 0 {
		deps := make([]string, len(fm.DependsOn))
		copy(deps, fm.DependsOn)
		sort.Strings(deps)

		for _, dep := range deps {
			if logicalnames.IsArtifactRef(dep) {
				content, err := readArtifactContent(dep)
				if err != nil {
					continue
				}
				hash := sha1.Sum([]byte(content))
				hashParts = append(hashParts, hash[:])
			} else {
				hasQual, _ := logicalnames.HasQualifier(dep)
				if hasQual {
					qualName, _ := logicalnames.QualifierName(dep)
					// Strip qualifier to get the base logical name for lookup.
					baseName := dep[:strings.Index(dep, "(")]
					parsed, ok := parsedCache[baseName]
					if !ok || parsed.Public == nil {
						continue
					}
					sub := findSubsection(parsed.Public, qualName)
					if sub == nil {
						continue
					}
					subContent := "## " + sub.Heading + "\n" + sub.Content
					hash := sha1.Sum([]byte(subContent))
					hashParts = append(hashParts, hash[:])
				} else {
					parsed, ok := parsedCache[dep]
					if !ok || parsed.Public == nil {
						continue
					}
					publicContent := sectionWithHeading(parsed.Public)
					hash := sha1.Sum([]byte(publicContent))
					hashParts = append(hashParts, hash[:])
				}
			}
		}
	}

	// External files: sorted alphabetically by path.
	if len(fm.External) > 0 {
		externals := make([]frontmatter.External, len(fm.External))
		copy(externals, fm.External)
		sort.Slice(externals, func(i, j int) bool {
			return externals[i].Path < externals[j].Path
		})

		for _, ext := range externals {
			content, err := readExternalContent(ext)
			if err != nil {
				continue
			}
			hash := sha1.Sum([]byte(content))
			hashParts = append(hashParts, hash[:])
		}
	}

	// Target # Public.
	parsed, ok := parsedCache[logicalName]
	if ok && parsed.Public != nil {
		publicContent := sectionWithHeading(parsed.Public)
		hash := sha1.Sum([]byte(publicContent))
		hashParts = append(hashParts, hash[:])
	}

	// Target # Agent.
	if ok && parsed.Agent != nil {
		agentContent := sectionWithHeading(parsed.Agent)
		hash := sha1.Sum([]byte(agentContent))
		hashParts = append(hashParts, hash[:])
	}

	// Input.
	if fm.Input != "" {
		content, err := readArtifactContent(fm.Input)
		if err == nil {
			hash := sha1.Sum([]byte(content))
			hashParts = append(hashParts, hash[:])
		}
	}

	// Final hash.
	if len(hashParts) == 0 {
		return ""
	}

	var concatenated []byte
	for _, h := range hashParts {
		concatenated = append(concatenated, h...)
	}
	finalHash := sha1.Sum(concatenated)
	encoded := base64.RawURLEncoding.EncodeToString(finalHash[:])
	if len(encoded) > 27 {
		encoded = encoded[:27]
	}
	return encoded
}

// sectionWithHeading returns the full section content including its heading.
func sectionWithHeading(s *parsenode.Section) string {
	var buf strings.Builder
	buf.WriteString("# ")
	buf.WriteString(s.Heading)
	buf.WriteString("\n")
	if s.Content != "" {
		buf.WriteString(s.Content)
	}
	for _, sub := range s.Subsections {
		buf.WriteString("\n## ")
		buf.WriteString(sub.Heading)
		buf.WriteString("\n")
		if sub.Content != "" {
			buf.WriteString(sub.Content)
		}
	}
	return buf.String()
}

// findSubsection finds a subsection within a section by name comparison.
func findSubsection(public *parsenode.Section, qualifier string) *parsenode.Subsection {
	if public == nil {
		return nil
	}
	for i := range public.Subsections {
		if strings.EqualFold(public.Subsections[i].Heading, qualifier) {
			return &public.Subsections[i]
		}
	}
	return nil
}

// readArtifactContent resolves an ARTIFACT/ reference to its artifact file
// and reads the content excluding frontmatter.
func readArtifactContent(artifactRef string) (string, error) {
	nodePath, artifactID, ok := logicalnames.ArtifactRefParts(artifactRef)
	if !ok {
		return "", fmt.Errorf("cannot resolve artifact reference: %s", artifactRef)
	}

	fm, err := frontmatter.ParseFrontmatter(nodePath)
	if err != nil {
		return "", fmt.Errorf("cannot read node %s: %w", nodePath, err)
	}

	var artifactPath string
	for _, out := range fm.Outputs {
		if out.ID == artifactID {
			artifactPath = out.Path
			break
		}
	}
	if artifactPath == "" {
		return "", fmt.Errorf("artifact ID %q not found in outputs of %s", artifactID, nodePath)
	}

	return readFileExcludingFrontmatter(artifactPath)
}

// readFileExcludingFrontmatter reads a file and strips YAML frontmatter.
func readFileExcludingFrontmatter(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	return stripFrontmatter(content), nil
}

// stripFrontmatter removes YAML frontmatter delimited by --- lines.
func stripFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---") {
		return content
	}
	idx := strings.Index(content[3:], "\n")
	if idx < 0 {
		return content
	}
	rest := content[3+idx+1:]
	closingIdx := strings.Index(rest, "---")
	if closingIdx < 0 {
		return content
	}
	afterClosing := rest[closingIdx+3:]
	nlIdx := strings.Index(afterClosing, "\n")
	if nlIdx < 0 {
		return ""
	}
	return afterClosing[nlIdx+1:]
}

// readExternalContent reads external file content with optional fragment extraction.
func readExternalContent(ext frontmatter.External) (string, error) {
	data, err := os.ReadFile(ext.Path)
	if err != nil {
		return "", err
	}
	content := strings.ReplaceAll(string(data), "\r\n", "\n")

	if len(ext.Fragments) == 0 {
		return content, nil
	}

	lines := strings.Split(content, "\n")
	var result strings.Builder
	for _, frag := range ext.Fragments {
		parts := strings.SplitN(frag.Lines, "-", 2)
		if len(parts) != 2 {
			continue
		}
		start := 0
		end := 0
		fmt.Sscanf(parts[0], "%d", &start)
		fmt.Sscanf(parts[1], "%d", &end)
		if start < 1 || end < start || end > len(lines) {
			continue
		}
		extracted := lines[start-1 : end]
		if result.Len() > 0 {
			result.WriteString("\n")
		}
		result.WriteString(strings.Join(extracted, "\n"))
	}

	return result.String(), nil
}

// toolError returns a CallToolResult with IsError set to true.
func toolError(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		IsError: true,
	}
}
