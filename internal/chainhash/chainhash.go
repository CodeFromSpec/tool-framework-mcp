// code-from-spec: ROOT/golang/implementation/chain/hash@8LgEtLs-PanMOv6ODVP6c0nz7k4
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

var ErrFileUnreadable = errors.New("file unreadable")
var ErrParseFailure = errors.New("parse failure")

var artifactTagPattern = regexp.MustCompile(`code-from-spec: [^\s@]+@(.{27})`)

func ChainHashCompute(chain *chainresolver.Chain) (string, error) {
	var contentHashes [][]byte

	for _, ancestor := range chain.Ancestors {
		node, err := parsenode.NodeParse(ancestor.LogicalName)
		if err != nil {
			return "", fmt.Errorf("%w: %s: %v", ErrParseFailure, ancestor.LogicalName, err)
		}
		h := hashFullSection(node.Public)
		if h != nil {
			contentHashes = append(contentHashes, h)
		}
	}

	for _, dep := range chain.Dependencies {
		if logicalnames.LogicalNameIsArtifact(dep.LogicalName) {
			h, err := hashArtifactFile(&dep.FilePath)
			if err != nil {
				return "", err
			}
			contentHashes = append(contentHashes, h)
		} else if dep.Qualifier == "" {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return "", fmt.Errorf("%w: %s: %v", ErrParseFailure, dep.LogicalName, err)
			}
			h := hashFullSection(node.Public)
			if h != nil {
				contentHashes = append(contentHashes, h)
			}
		} else {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return "", fmt.Errorf("%w: %s: %v", ErrParseFailure, dep.LogicalName, err)
			}
			if node.Public != nil {
				normalizedQualifier := textnormalization.NormalizeText(dep.Qualifier)
				for _, sub := range node.Public.Subsections {
					if sub.Heading == normalizedQualifier {
						h := hashSubsection(sub)
						contentHashes = append(contentHashes, h)
						break
					}
				}
			}
		}
	}

	for _, ext := range chain.External {
		h, err := hashExternalFile(ext.Path)
		if err != nil {
			return "", err
		}
		contentHashes = append(contentHashes, h)
	}

	targetNode, err := parsenode.NodeParse(chain.Target.LogicalName)
	if err != nil {
		return "", fmt.Errorf("%w: %s: %v", ErrParseFailure, chain.Target.LogicalName, err)
	}
	publicHash := hashFullSection(targetNode.Public)
	if publicHash != nil {
		contentHashes = append(contentHashes, publicHash)
	}
	agentHash := hashFullSection(targetNode.Agent)
	if agentHash != nil {
		contentHashes = append(contentHashes, agentHash)
	}

	if chain.Input != nil {
		h, err := hashArtifactFile(&chain.Input.FilePath)
		if err != nil {
			return "", err
		}
		contentHashes = append(contentHashes, h)
	}

	var combined []byte
	for _, h := range contentHashes {
		combined = append(combined, h...)
	}

	finalHash := sha1.Sum(combined)
	return base64.RawURLEncoding.EncodeToString(finalHash[:]), nil
}

func hashFullSection(section *parsenode.NodeSection) []byte {
	if section == nil {
		return nil
	}

	var sb strings.Builder
	sb.WriteString(section.RawHeading)
	sb.WriteString("\n")

	for _, line := range section.Content {
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	for _, sub := range section.Subsections {
		sb.WriteString(sub.RawHeading)
		sb.WriteString("\n")
		for _, line := range sub.Content {
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	h := sha1.Sum([]byte(sb.String()))
	return h[:]
}

func hashSubsection(sub *parsenode.NodeSubsection) []byte {
	var sb strings.Builder
	sb.WriteString(sub.RawHeading)
	sb.WriteString("\n")
	for _, line := range sub.Content {
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	h := sha1.Sum([]byte(sb.String()))
	return h[:]
}

func hashArtifactFile(filePath *pathutils.PathCfs) ([]byte, error) {
	reader, err := filereader.FileOpen(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrFileUnreadable, filePath.Value, err)
	}

	firstLine, err := filereader.FileReadLine(reader)
	if err != nil {
		if errors.Is(err, filereader.ErrEndOfFile) {
			filereader.FileClose(reader)
			h := sha1.Sum([]byte{})
			return h[:], nil
		}
		filereader.FileClose(reader)
		return nil, fmt.Errorf("%w: %s: %v", ErrFileUnreadable, filePath.Value, err)
	}

	var lines []string
	if firstLine == "---" {
		for {
			line, err := filereader.FileReadLine(reader)
			if err != nil {
				if errors.Is(err, filereader.ErrEndOfFile) {
					break
				}
				filereader.FileClose(reader)
				return nil, fmt.Errorf("%w: %s: %v", ErrFileUnreadable, filePath.Value, err)
			}
			if line == "---" {
				break
			}
		}
		for {
			line, err := filereader.FileReadLine(reader)
			if err != nil {
				if errors.Is(err, filereader.ErrEndOfFile) {
					break
				}
				filereader.FileClose(reader)
				return nil, fmt.Errorf("%w: %s: %v", ErrFileUnreadable, filePath.Value, err)
			}
			lines = append(lines, line)
		}
	} else {
		lines = append(lines, firstLine)
		for {
			line, err := filereader.FileReadLine(reader)
			if err != nil {
				if errors.Is(err, filereader.ErrEndOfFile) {
					break
				}
				filereader.FileClose(reader)
				return nil, fmt.Errorf("%w: %s: %v", ErrFileUnreadable, filePath.Value, err)
			}
			lines = append(lines, line)
		}
	}

	filereader.FileClose(reader)

	var sb strings.Builder
	for _, line := range lines {
		sb.WriteString(neutralizeArtifactTag(line))
		sb.WriteString("\n")
	}

	h := sha1.Sum([]byte(sb.String()))
	return h[:], nil
}

func hashExternalFile(pathString string) ([]byte, error) {
	cfsPath := &pathutils.PathCfs{Value: pathString}
	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrFileUnreadable, pathString, err)
	}

	var sb strings.Builder
	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				break
			}
			filereader.FileClose(reader)
			return nil, fmt.Errorf("%w: %s: %v", ErrFileUnreadable, pathString, err)
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	filereader.FileClose(reader)

	h := sha1.Sum([]byte(sb.String()))
	return h[:], nil
}

func neutralizeArtifactTag(line string) string {
	return artifactTagPattern.ReplaceAllStringFunc(line, func(match string) string {
		loc := artifactTagPattern.FindStringSubmatchIndex(match)
		if loc == nil {
			return match
		}
		start := loc[2]
		end := loc[3]
		return match[:start] + "---------------------------" + match[end:]
	})
}
