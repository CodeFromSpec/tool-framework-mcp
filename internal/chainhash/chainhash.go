// code-from-spec: SPEC/golang/implementation/chain/hash@2ocQDHyeDy3KJjNkWay66tM7kS8
package chainhash

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/file"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/textnormalization"
)

var ErrParseFailure = errors.New("parse failure")

func extractBlock(content []string) string {
	start := 0
	for start < len(content) {
		if strings.TrimSpace(content[start]) != "" {
			break
		}
		start++
	}

	end := len(content) - 1
	for end >= start {
		if strings.TrimSpace(content[end]) != "" {
			break
		}
		end--
	}

	if start > end {
		return ""
	}

	return strings.Join(content[start:end+1], "\n") + "\n"
}

func formatSection(rawHeading string, content []string) string {
	head := strings.TrimRight(rawHeading, " \t") + "\n"
	body := extractBlock(content)
	return head + body
}

func concatenateSubsections(subsections []*parsenode.NodeSubsection) string {
	result := ""
	for _, sub := range subsections {
		block := formatSection(sub.RawHeading, sub.Content)
		if result != "" && block != "" {
			result += "\n"
		}
		result += block
	}
	return result
}

func hashPublicSubsections(node *parsenode.Node) []byte {
	if node.Public == nil {
		return nil
	}
	if len(node.Public.Subsections) == 0 {
		return nil
	}
	text := concatenateSubsections(node.Public.Subsections)
	sum := sha1.Sum([]byte(text))
	return sum[:]
}

func hashQualifiedSubsection(node *parsenode.Node, qualifier string) []byte {
	normalizedQualifier := textnormalization.NormalizeText(qualifier)
	if node.Public == nil {
		return nil
	}
	for _, sub := range node.Public.Subsections {
		if sub.Heading == normalizedQualifier {
			text := formatSection(sub.RawHeading, sub.Content)
			sum := sha1.Sum([]byte(text))
			return sum[:]
		}
	}
	return nil
}

func hashAgentSection(node *parsenode.Node) []byte {
	if node.Agent == nil {
		return nil
	}
	if extractBlock(node.Agent.Content) == "" && len(node.Agent.Subsections) == 0 {
		return nil
	}
	text := formatSection(node.Agent.RawHeading, node.Agent.Content)
	for _, sub := range node.Agent.Subsections {
		subBlock := formatSection(sub.RawHeading, sub.Content)
		if subBlock != "" {
			text += "\n" + subBlock
		}
	}
	sum := sha1.Sum([]byte(text))
	return sum[:]
}

var artifactTagPattern = "code-from-spec: "

func neutralizeLine(line string) string {
	idx := strings.Index(line, artifactTagPattern)
	if idx == -1 {
		return line
	}
	afterTag := line[idx+len(artifactTagPattern):]
	atSign := strings.Index(afterTag, "@")
	if atSign == -1 {
		return line
	}
	hashStart := atSign + 1
	if hashStart+27 > len(afterTag) {
		return line
	}
	prefix := line[:idx+len(artifactTagPattern)]
	logicalName := afterTag[:atSign]
	suffix := afterTag[hashStart+27:]
	return prefix + logicalName + "@" + "---------------------------" + suffix
}

func hashFileContent(filePath pathutils.PathCfs, neutralizeArtifactTag bool) ([]byte, error) {
	handle, err := file.FileOpen(filePath, "read", 30000)
	if errors.Is(err, file.ErrFileUnreadable) {
		return nil, fmt.Errorf("file unreadable: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %w", filePath.Value, err)
	}

	var sb strings.Builder
	for {
		line, err := file.FileReadLine(handle)
		if errors.Is(err, file.ErrEndOfFile) {
			break
		}
		if err != nil {
			file.FileClose(handle)
			return nil, fmt.Errorf("reading file %s: %w", filePath.Value, err)
		}
		if neutralizeArtifactTag {
			line = neutralizeLine(line)
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	file.FileClose(handle)

	text := sb.String()
	sum := sha1.Sum([]byte(text))
	return sum[:], nil
}

func ChainHashCompute(chain *chainresolver.Chain) (string, error) {
	var hashes [][]byte

	for _, ancestor := range chain.Ancestors {
		if ancestor == nil {
			continue
		}
		node, err := parsenode.NodeParse(ancestor.UnqualifiedLogicalName)
		if err != nil {
			return "", fmt.Errorf("%w: ancestor %s: %s", ErrParseFailure, ancestor.UnqualifiedLogicalName, err)
		}
		h := hashPublicSubsections(node)
		if h != nil {
			hashes = append(hashes, h)
		}
	}

	for _, dep := range chain.Dependencies {
		if dep == nil {
			continue
		}
		if strings.HasPrefix(dep.UnqualifiedLogicalName, "ARTIFACT/") {
			h, err := hashFileContent(dep.FilePath, true)
			if err != nil {
				return "", err
			}
			hashes = append(hashes, h)
		} else if strings.HasPrefix(dep.UnqualifiedLogicalName, "EXTERNAL/") {
			h, err := hashFileContent(dep.FilePath, false)
			if err != nil {
				return "", err
			}
			hashes = append(hashes, h)
		} else if strings.HasPrefix(dep.UnqualifiedLogicalName, "SPEC/") || dep.UnqualifiedLogicalName == "SPEC" {
			node, err := parsenode.NodeParse(dep.UnqualifiedLogicalName)
			if err != nil {
				return "", fmt.Errorf("%w: dependency %s: %s", ErrParseFailure, dep.UnqualifiedLogicalName, err)
			}
			if dep.Qualifier == "" {
				h := hashPublicSubsections(node)
				if h != nil {
					hashes = append(hashes, h)
				}
			} else {
				h := hashQualifiedSubsection(node, dep.Qualifier)
				if h != nil {
					hashes = append(hashes, h)
				}
			}
		}
	}

	if chain.Target == nil {
		return "", fmt.Errorf("%w: chain target is nil", ErrParseFailure)
	}
	targetNode, err := parsenode.NodeParse(chain.Target.UnqualifiedLogicalName)
	if err != nil {
		return "", fmt.Errorf("%w: target %s: %s", ErrParseFailure, chain.Target.UnqualifiedLogicalName, err)
	}
	h := hashPublicSubsections(targetNode)
	if h != nil {
		hashes = append(hashes, h)
	}

	agentHash := hashAgentSection(targetNode)
	if agentHash != nil {
		hashes = append(hashes, agentHash)
	}

	if chain.Input != nil {
		input := chain.Input
		if strings.HasPrefix(input.UnqualifiedLogicalName, "ARTIFACT/") {
			h, err := hashFileContent(input.FilePath, true)
			if err != nil {
				return "", err
			}
			hashes = append(hashes, h)
		} else if strings.HasPrefix(input.UnqualifiedLogicalName, "EXTERNAL/") {
			h, err := hashFileContent(input.FilePath, false)
			if err != nil {
				return "", err
			}
			hashes = append(hashes, h)
		}
	}

	var concatenated []byte
	for _, h := range hashes {
		concatenated = append(concatenated, h...)
	}

	finalSum := sha1.Sum(concatenated)
	encoded := base64.RawURLEncoding.EncodeToString(finalSum[:])
	return encoded, nil
}
