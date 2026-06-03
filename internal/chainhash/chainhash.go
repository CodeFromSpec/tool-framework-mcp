// code-from-spec: ROOT/golang/implementation/chain/hash@xPtfCgKZo6dhvBdbe7pmfxdkshI
package chainhash

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
)

var ErrFileUnreadable = errors.New("file unreadable")
var ErrParseFailure = errors.New("parse failure")

var artifactTagPattern = regexp.MustCompile(`code-from-spec: [^@]+@[A-Za-z0-9_\-]{27}`)

func ChainHashCompute(chain *chainresolver.Chain) (string, error) {
	if chain == nil {
		return "", fmt.Errorf("%w: chain is nil", ErrFileUnreadable)
	}

	var contentHashes [][]byte

	for _, ancestor := range chain.Ancestors {
		if ancestor == nil {
			continue
		}
		node, err := parsenode.NodeParse(ancestor.LogicalName)
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
		}
		h := hashFullSection(node.Public)
		if h != nil {
			contentHashes = append(contentHashes, h)
		}
	}

	for _, dep := range chain.Dependencies {
		if dep == nil {
			continue
		}
		if logicalnames.LogicalNameIsArtifact(dep.LogicalName) {
			h, err := hashArtifactFile(&dep.FilePath)
			if err != nil {
				return "", err
			}
			contentHashes = append(contentHashes, h)
		} else if dep.Qualifier == "" {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
			}
			h := hashFullSection(node.Public)
			if h != nil {
				contentHashes = append(contentHashes, h)
			}
		} else {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
			}
			if node.Public != nil {
				normalizedQualifier := textnormalization.NormalizeText(dep.Qualifier)
				for _, sub := range node.Public.Subsections {
					if sub == nil {
						continue
					}
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
		if ext == nil {
			continue
		}
		h, err := hashExternalFile(ext.Path)
		if err != nil {
			return "", err
		}
		contentHashes = append(contentHashes, h)
	}

	if chain.Target == nil {
		return "", fmt.Errorf("%w: target is nil", ErrFileUnreadable)
	}
	targetNode, err := parsenode.NodeParse(chain.Target.LogicalName)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
	}

	h := hashFullSection(targetNode.Public)
	if h != nil {
		contentHashes = append(contentHashes, h)
	}

	h = hashFullSection(targetNode.Agent)
	if h != nil {
		contentHashes = append(contentHashes, h)
	}

	if chain.Input != nil {
		h, err := hashArtifactFile(&chain.Input.FilePath)
		if err != nil {
			return "", err
		}
		contentHashes = append(contentHashes, h)
	}

	var concatenated []byte
	for _, ch := range contentHashes {
		concatenated = append(concatenated, ch...)
	}

	finalHash := sha1.Sum(concatenated)
	return base64.RawURLEncoding.EncodeToString(finalHash[:]), nil
}

func hashFullSection(section *parsenode.NodeSection) []byte {
	if section == nil {
		return nil
	}

	var acc []byte
	acc = append(acc, []byte(section.RawHeading+"\n")...)

	for _, line := range section.Content {
		acc = append(acc, []byte(line+"\n")...)
	}

	for _, sub := range section.Subsections {
		if sub == nil {
			continue
		}
		acc = append(acc, []byte(sub.RawHeading+"\n")...)
		for _, line := range sub.Content {
			acc = append(acc, []byte(line+"\n")...)
		}
	}

	h := sha1.Sum(acc)
	result := make([]byte, 20)
	copy(result, h[:])
	return result
}

func hashSubsection(sub *parsenode.NodeSubsection) []byte {
	if sub == nil {
		return nil
	}

	var acc []byte
	acc = append(acc, []byte(sub.RawHeading+"\n")...)

	for _, line := range sub.Content {
		acc = append(acc, []byte(line+"\n")...)
	}

	h := sha1.Sum(acc)
	result := make([]byte, 20)
	copy(result, h[:])
	return result
}

func hashArtifactFile(filePath *pathutils.PathCfs) ([]byte, error) {
	if filePath == nil {
		return nil, fmt.Errorf("%w: file path is nil", ErrFileUnreadable)
	}

	reader, err := filereader.FileOpen(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	firstLine, err := filereader.FileReadLine(reader)
	if err != nil {
		if errors.Is(err, filereader.ErrEndOfFile) {
			filereader.FileClose(reader)
			h := sha1.Sum([]byte{})
			result := make([]byte, 20)
			copy(result, h[:])
			return result, nil
		}
		filereader.FileClose(reader)
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
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
				return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
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
				return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
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
				return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
			}
			lines = append(lines, line)
		}
	}

	var acc []byte
	for _, line := range lines {
		neutralized := neutralizeArtifactTag(line)
		acc = append(acc, []byte(neutralized+"\n")...)
	}

	filereader.FileClose(reader)

	h := sha1.Sum(acc)
	result := make([]byte, 20)
	copy(result, h[:])
	return result, nil
}

func hashExternalFile(pathString string) ([]byte, error) {
	cfsPath := &pathutils.PathCfs{Value: pathString}

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	var acc []byte
	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				break
			}
			filereader.FileClose(reader)
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		acc = append(acc, []byte(line+"\n")...)
	}

	filereader.FileClose(reader)

	h := sha1.Sum(acc)
	result := make([]byte, 20)
	copy(result, h[:])
	return result, nil
}

func neutralizeArtifactTag(line string) string {
	return artifactTagPattern.ReplaceAllStringFunc(line, func(match string) string {
		atIdx := len(match) - 28
		return match[:atIdx+1] + "---------------------------"
	})
}
