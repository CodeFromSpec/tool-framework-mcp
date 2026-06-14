// code-from-spec: ROOT/golang/implementation/chain/hash@xyJ2Wt92BZRK8unOV47TsaxCCjI
package chainhash

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
)

var ErrParseFailure = errors.New("a node file cannot be parsed")

var artifactTagPattern = regexp.MustCompile(`[A-Za-z0-9_-]{27}`)

func extractBlock(content []string) string {
	start := 0
	for start < len(content) {
		trimmed := strings.TrimLeft(content[start], " \t")
		if trimmed != "" {
			break
		}
		start++
	}

	end := len(content)
	for end > start {
		trimmed := strings.TrimLeft(content[end-1], " \t")
		if trimmed != "" {
			break
		}
		end--
	}

	if start >= end {
		return ""
	}

	return strings.Join(content[start:end], "\n") + "\n"
}

func concatenateSubsections(subsections []*parsenode.NodeSubsection) string {
	result := ""
	for _, sub := range subsections {
		headingLine := strings.TrimRight(sub.RawHeading, " \t") + "\n"
		body := extractBlock(sub.Content)

		if result != "" {
			result += "\n"
		}
		result += headingLine
		result += body
	}
	return result
}

func extractAgentSection(section *parsenode.NodeSection) string {
	headingLine := strings.TrimRight(section.RawHeading, " \t") + "\n"
	body := extractBlock(section.Content)
	result := headingLine + body

	for _, sub := range section.Subsections {
		subHeadingLine := strings.TrimRight(sub.RawHeading, " \t") + "\n"
		subBody := extractBlock(sub.Content)
		result += "\n"
		result += subHeadingLine
		result += subBody
	}

	return result
}

func neutralizeArtifactTag(line string) string {
	const marker = "code-from-spec: "
	idx := strings.Index(line, marker)
	if idx == -1 {
		return line
	}

	afterMarker := line[idx+len(marker):]
	loc := artifactTagPattern.FindStringIndex(afterMarker)
	if loc == nil {
		return line
	}

	hashStart := loc[0]
	hashEnd := loc[1]
	if hashEnd-hashStart != 27 {
		return line
	}

	neutralized := line[:idx+len(marker)] +
		afterMarker[:hashStart] +
		"---------------------------" +
		afterMarker[hashEnd:]

	return neutralized
}

func hashPublicSubsections(node *parsenode.Node) []byte {
	if node.Public == nil {
		return nil
	}
	if len(node.Public.Subsections) == 0 {
		return nil
	}
	content := concatenateSubsections(node.Public.Subsections)
	digest := sha1.Sum([]byte(content))
	return digest[:]
}

func hashQualifiedSubsection(node *parsenode.Node, qualifier string) []byte {
	if node.Public == nil {
		return nil
	}
	normalizedQualifier := textnormalization.NormalizeText(qualifier)

	var found *parsenode.NodeSubsection
	for _, sub := range node.Public.Subsections {
		if sub.Heading == normalizedQualifier {
			found = sub
			break
		}
	}

	if found == nil {
		return nil
	}

	headingLine := strings.TrimRight(found.RawHeading, " \t") + "\n"
	body := extractBlock(found.Content)
	content := headingLine + body

	digest := sha1.Sum([]byte(content))
	return digest[:]
}

func hashAgentSection(node *parsenode.Node) []byte {
	if node.Agent == nil {
		return nil
	}
	if len(node.Agent.Content) == 0 && len(node.Agent.Subsections) == 0 {
		return nil
	}
	content := extractAgentSection(node.Agent)
	digest := sha1.Sum([]byte(content))
	return digest[:]
}

func hashArtifactFile(filePath *pathutils.PathCfs) ([]byte, error) {
	reader, err := filereader.FileOpen(filePath)
	if err != nil {
		return nil, fmt.Errorf("file unreadable: %w", err)
	}

	content := ""
	for {
		line, err := filereader.FileReadLine(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			break
		}
		if err != nil {
			filereader.FileClose(reader)
			return nil, fmt.Errorf("reading artifact file: %w", err)
		}
		neutralized := neutralizeArtifactTag(line)
		content += neutralized + "\n"
	}

	filereader.FileClose(reader)

	digest := sha1.Sum([]byte(content))
	return digest[:], nil
}

func hashExternalFile(filePath *pathutils.PathCfs) ([]byte, error) {
	reader, err := filereader.FileOpen(filePath)
	if err != nil {
		return nil, fmt.Errorf("file unreadable: %w", err)
	}

	content := ""
	for {
		line, err := filereader.FileReadLine(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			break
		}
		if err != nil {
			filereader.FileClose(reader)
			return nil, fmt.Errorf("reading external file: %w", err)
		}
		content += line + "\n"
	}

	filereader.FileClose(reader)

	digest := sha1.Sum([]byte(content))
	return digest[:], nil
}

func ChainHashCompute(chain *chainresolver.Chain) (string, error) {
	if chain == nil {
		return "", fmt.Errorf("%w: chain is nil", ErrParseFailure)
	}

	var rawHashes [][]byte

	for _, ancestor := range chain.Ancestors {
		if ancestor == nil {
			continue
		}
		node, err := parsenode.NodeParse(ancestor.UnqualifiedLogicalName)
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
		}
		h := hashPublicSubsections(node)
		if h != nil {
			rawHashes = append(rawHashes, h)
		}
	}

	for _, dep := range chain.Dependencies {
		if dep == nil {
			continue
		}
		if logicalnames.LogicalNameIsArtifact(dep.UnqualifiedLogicalName) {
			h, err := hashArtifactFile(&dep.FilePath)
			if err != nil {
				return "", err
			}
			rawHashes = append(rawHashes, h)
		} else if logicalnames.LogicalNameIsExternal(dep.UnqualifiedLogicalName) {
			h, err := hashExternalFile(&dep.FilePath)
			if err != nil {
				return "", err
			}
			rawHashes = append(rawHashes, h)
		} else if logicalnames.LogicalNameIsSpec(dep.UnqualifiedLogicalName) && dep.Qualifier == nil {
			node, err := parsenode.NodeParse(dep.UnqualifiedLogicalName)
			if err != nil {
				return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
			}
			h := hashPublicSubsections(node)
			if h != nil {
				rawHashes = append(rawHashes, h)
			}
		} else if logicalnames.LogicalNameIsSpec(dep.UnqualifiedLogicalName) && dep.Qualifier != nil {
			node, err := parsenode.NodeParse(dep.UnqualifiedLogicalName)
			if err != nil {
				return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
			}
			h := hashQualifiedSubsection(node, *dep.Qualifier)
			if h != nil {
				rawHashes = append(rawHashes, h)
			}
		}
	}

	if chain.Target == nil {
		return "", fmt.Errorf("%w: target is nil", ErrParseFailure)
	}

	targetNode, err := parsenode.NodeParse(chain.Target.UnqualifiedLogicalName)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
	}

	hPublic := hashPublicSubsections(targetNode)
	if hPublic != nil {
		rawHashes = append(rawHashes, hPublic)
	}

	hAgent := hashAgentSection(targetNode)
	if hAgent != nil {
		rawHashes = append(rawHashes, hAgent)
	}

	if chain.Input != nil {
		input := chain.Input
		if logicalnames.LogicalNameIsArtifact(input.UnqualifiedLogicalName) {
			h, err := hashArtifactFile(&input.FilePath)
			if err != nil {
				return "", err
			}
			rawHashes = append(rawHashes, h)
		} else if logicalnames.LogicalNameIsExternal(input.UnqualifiedLogicalName) {
			h, err := hashExternalFile(&input.FilePath)
			if err != nil {
				return "", err
			}
			rawHashes = append(rawHashes, h)
		}
	}

	var concatenated []byte
	for _, h := range rawHashes {
		concatenated = append(concatenated, h...)
	}

	finalDigest := sha1.Sum(concatenated)
	encoded := base64.RawURLEncoding.EncodeToString(finalDigest[:])

	return encoded, nil
}
