// code-from-spec: SPEC/golang/implementation/chain/hash@eDH9kM0QSx-ZbbcnqRYYhKneQ9c
package chainhash

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"
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

func concatenateSubsections(subsections []*parsing.NodeSubsection) string {
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

func hashPublicSubsections(node *parsing.Node) []byte {
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

func hashQualifiedSubsection(node *parsing.Node, qualifier string) []byte {
	normalizedQualifier := parsing.NormalizeText(qualifier)
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

func hashAgentSection(node *parsing.Node) []byte {
	if node.Agent == nil {
		return nil
	}
	if extractBlock(node.Agent.Content) == "" && len(node.Agent.Subsections) == 0 {
		return nil
	}
	text := extractBlock(node.Agent.Content)
	for _, sub := range node.Agent.Subsections {
		subBlock := formatSection(sub.RawHeading, sub.Content)
		if text != "" && subBlock != "" {
			text += "\n"
		}
		text += subBlock
	}
	sum := sha1.Sum([]byte(text))
	return sum[:]
}

func hashFileContent(filePath oslayer.CfsPath) ([]byte, error) {
	handle, err := oslayer.OpenFile(filePath, "read", 30000)
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %w", filePath, err)
	}

	var sb strings.Builder
	for {
		line, err := handle.ReadLine()
		if errors.Is(err, oslayer.ErrEndOfFile) {
			break
		}
		if err != nil {
			handle.Close()
			return nil, fmt.Errorf("reading file %s: %w", filePath, err)
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	handle.Close()

	text := sb.String()
	sum := sha1.Sum([]byte(text))
	return sum[:], nil
}

func processSpecDep(ref parsing.CfsReference) ([]byte, error) {
	node, err := parsing.ParseNode(ref.LogicalName)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %s", ErrParseFailure, ref.LogicalName, err)
	}
	if ref.Qualifier == nil {
		return hashPublicSubsections(node), nil
	}
	return hashQualifiedSubsection(node, *ref.Qualifier), nil
}

func ChainHashCompute(chain chainresolver.Chain) (string, error) {
	var hashes [][]byte

	for _, ancestor := range chain.Ancestors {
		node, err := parsing.ParseNode(ancestor.LogicalName)
		if err != nil {
			return "", fmt.Errorf("%w: ancestor %s: %s", ErrParseFailure, ancestor.LogicalName, err)
		}
		h := hashPublicSubsections(node)
		if h != nil {
			hashes = append(hashes, h)
		}
	}

	for _, dep := range chain.Dependencies {
		if strings.HasPrefix(dep.LogicalName, "ARTIFACT/") {
			h, err := hashFileContent(oslayer.CfsPath(dep.Path))
			if err != nil {
				return "", err
			}
			hashes = append(hashes, h)
		} else if strings.HasPrefix(dep.LogicalName, "EXTERNAL/") {
			h, err := hashFileContent(oslayer.CfsPath(dep.Path))
			if err != nil {
				return "", err
			}
			hashes = append(hashes, h)
		} else if strings.HasPrefix(dep.LogicalName, "SPEC/") {
			h, err := processSpecDep(dep)
			if err != nil {
				return "", err
			}
			if h != nil {
				hashes = append(hashes, h)
			}
		}
	}

	targetNode, err := parsing.ParseNode(chain.Target.LogicalName)
	if err != nil {
		return "", fmt.Errorf("%w: target %s: %s", ErrParseFailure, chain.Target.LogicalName, err)
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
		hashes = append(hashes, []byte{0x49})
		input := chain.Input
		if strings.HasPrefix(input.LogicalName, "ARTIFACT/") {
			h, err := hashFileContent(oslayer.CfsPath(input.Path))
			if err != nil {
				return "", err
			}
			hashes = append(hashes, h)
		} else if strings.HasPrefix(input.LogicalName, "EXTERNAL/") {
			h, err := hashFileContent(oslayer.CfsPath(input.Path))
			if err != nil {
				return "", err
			}
			hashes = append(hashes, h)
		} else if strings.HasPrefix(input.LogicalName, "SPEC/") {
			h, err := processSpecDep(*input)
			if err != nil {
				return "", err
			}
			if h != nil {
				hashes = append(hashes, h)
			}
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
