// code-from-spec: ROOT/golang/internal/chain_hash/code@PENDING
package chainhash

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/normalizename"
)

// ComputeChainHash computes the chain hash for a spec node by reading raw
// files from disk. The only normalization applied is CRLF to LF conversion.
func ComputeChainHash(logicalName string) (string, error) {
	var hashParts [][]byte

	// Step 1: Ancestors (root first, then down to target's parent).
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
		ancestorPath, ok := logicalnames.PathFromLogicalName(ancestor)
		if !ok {
			continue
		}
		raw, err := readRawFile(ancestorPath)
		if err != nil {
			continue
		}
		publicSection := extractSection(raw, "Public")
		if strings.TrimSpace(publicSection) == "" {
			continue
		}
		hash := sha1.Sum([]byte(publicSection))
		hashParts = append(hashParts, hash[:])
	}

	// Step 2: Dependencies (depends_on), sorted alphabetically.
	targetPath, ok := logicalnames.PathFromLogicalName(logicalName)
	if !ok {
		return "", fmt.Errorf("cannot resolve logical name: %s", logicalName)
	}
	fm, err := frontmatter.ParseFrontmatter(targetPath)
	if err != nil {
		return "", fmt.Errorf("cannot read frontmatter: %w", err)
	}

	if len(fm.DependsOn) > 0 {
		deps := make([]string, len(fm.DependsOn))
		copy(deps, fm.DependsOn)
		sort.Strings(deps)

		for _, dep := range deps {
			if logicalnames.IsArtifactRef(dep) {
				// ARTIFACT/x/y(id): resolve artifact path, read file, strip frontmatter, SHA-1.
				content, err := readArtifactContent(dep)
				if err != nil {
					return "", fmt.Errorf("cannot read artifact %s: %w", dep, err)
				}
				hash := sha1.Sum([]byte(content))
				hashParts = append(hashParts, hash[:])
			} else {
				hasQual, _ := logicalnames.HasQualifier(dep)
				if hasQual {
					// ROOT/x/y(z): extract ## z subsection within # Public.
					qualName, _ := logicalnames.QualifierName(dep)
					baseName := dep[:strings.Index(dep, "(")]
					basePath, ok := logicalnames.PathFromLogicalName(baseName)
					if !ok {
						return "", fmt.Errorf("cannot resolve dependency: %s", dep)
					}
					raw, err := readRawFile(basePath)
					if err != nil {
						return "", fmt.Errorf("cannot read dependency %s: %w", dep, err)
					}
					publicSection := extractSection(raw, "Public")
					subSection := extractSubsection(publicSection, qualName)
					hash := sha1.Sum([]byte(subSection))
					hashParts = append(hashParts, hash[:])
				} else {
					// ROOT/x/y: read raw file, extract # Public, SHA-1.
					depPath, ok := logicalnames.PathFromLogicalName(dep)
					if !ok {
						return "", fmt.Errorf("cannot resolve dependency: %s", dep)
					}
					raw, err := readRawFile(depPath)
					if err != nil {
						return "", fmt.Errorf("cannot read dependency %s: %w", dep, err)
					}
					publicSection := extractSection(raw, "Public")
					hash := sha1.Sum([]byte(publicSection))
					hashParts = append(hashParts, hash[:])
				}
			}
		}
	}

	// Step 3: External files, sorted alphabetically by path.
	if len(fm.External) > 0 {
		externals := make([]frontmatter.External, len(fm.External))
		copy(externals, fm.External)
		sort.Slice(externals, func(i, j int) bool {
			return externals[i].Path < externals[j].Path
		})

		for _, ext := range externals {
			content, err := readExternalContent(ext)
			if err != nil {
				return "", fmt.Errorf("cannot read external %s: %w", ext.Path, err)
			}
			hash := sha1.Sum([]byte(content))
			hashParts = append(hashParts, hash[:])
		}
	}

	// Step 4: Target # Public.
	targetRaw, err := readRawFile(targetPath)
	if err != nil {
		return "", fmt.Errorf("cannot read target: %w", err)
	}

	publicSection := extractSection(targetRaw, "Public")
	if publicSection != "" {
		hash := sha1.Sum([]byte(publicSection))
		hashParts = append(hashParts, hash[:])
	}

	// Step 5: Target # Agent.
	agentSection := extractSection(targetRaw, "Agent")
	if agentSection != "" {
		hash := sha1.Sum([]byte(agentSection))
		hashParts = append(hashParts, hash[:])
	}

	// Step 6: Input artifact.
	if fm.Input != "" {
		content, err := readArtifactContent(fm.Input)
		if err != nil {
			return "", fmt.Errorf("cannot read input artifact %s: %w", fm.Input, err)
		}
		hash := sha1.Sum([]byte(content))
		hashParts = append(hashParts, hash[:])
	}

	// Step 7: Final hash.
	if len(hashParts) == 0 {
		return "", nil
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
	return encoded, nil
}

// readRawFile reads a file from disk and normalizes CRLF to LF.
func readRawFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(string(data), "\r\n", "\n"), nil
}

// extractSection extracts a top-level section (# <heading>) from raw content.
// The match is case-insensitive using normalizename. Returns the full section
// from its heading line to the next # heading line or EOF.
// Returns empty string if the section is not found.
func extractSection(rawContent, sectionHeading string) string {
	normalizedTarget := normalizename.NormalizeName(sectionHeading)
	lines := strings.Split(rawContent, "\n")
	startIdx := -1

	for i, line := range lines {
		if strings.HasPrefix(line, "# ") {
			heading := line[2:]
			if normalizename.NormalizeName(heading) == normalizedTarget {
				startIdx = i
				break
			}
		}
	}

	if startIdx < 0 {
		return ""
	}

	// Find the end: next # heading or EOF.
	endIdx := len(lines)
	for i := startIdx + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "# ") {
			endIdx = i
			break
		}
	}

	return strings.Join(lines[startIdx:endIdx], "\n")
}

// extractSubsection extracts a ## subsection within a # Public section.
// The match is case-insensitive using normalizename. Returns from the
// ## heading line to the next ## line, next # line, or EOF.
// Returns empty string if not found.
func extractSubsection(publicSection, subsectionHeading string) string {
	normalizedTarget := normalizename.NormalizeName(subsectionHeading)
	lines := strings.Split(publicSection, "\n")
	startIdx := -1

	for i, line := range lines {
		if strings.HasPrefix(line, "## ") {
			heading := line[3:]
			if normalizename.NormalizeName(heading) == normalizedTarget {
				startIdx = i
				break
			}
		}
	}

	if startIdx < 0 {
		return ""
	}

	// Find the end: next ## heading, next # heading, or EOF.
	endIdx := len(lines)
	for i := startIdx + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") || strings.HasPrefix(lines[i], "# ") {
			endIdx = i
			break
		}
	}

	return strings.Join(lines[startIdx:endIdx], "\n")
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

// readFileExcludingFrontmatter reads a file and returns its content
// with YAML frontmatter stripped. CRLF is normalized to LF.
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
