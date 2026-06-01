// code-from-spec: ROOT/golang/implementation/chain/hash@hmwqbGb0JaWy28fSl6RosI93k-8
package chainhash

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
)

var ErrFileUnreadable = errors.New("file unreadable")
var ErrParseFailure = errors.New("parse failure")

func hashSpecSection(node *parsenode.Node, sectionType string) ([]byte, bool) {
	var section *parsenode.NodeSection
	if sectionType == "public" {
		section = node.Public
	} else {
		section = node.Agent
	}

	if section == nil {
		return nil, false
	}

	if len(section.Content) == 0 && len(section.Subsections) == 0 {
		return nil, false
	}

	var buf []byte
	buf = append(buf, []byte(section.RawHeading+"\n")...)

	for _, line := range section.Content {
		buf = append(buf, []byte(line+"\n")...)
	}

	for _, sub := range section.Subsections {
		buf = append(buf, []byte(sub.RawHeading+"\n")...)
		for _, line := range sub.Content {
			buf = append(buf, []byte(line+"\n")...)
		}
	}

	digest := sha1.Sum(buf)
	return digest[:], true
}

func hashSpecSubsection(node *parsenode.Node, qualifier string) ([]byte, bool) {
	normalizedQualifier := textnormalization.NormalizeText(qualifier)

	if node.Public == nil {
		return nil, false
	}

	var found *parsenode.NodeSubsection
	for _, sub := range node.Public.Subsections {
		if sub.Heading == normalizedQualifier {
			found = sub
			break
		}
	}

	if found == nil {
		return nil, false
	}

	var buf []byte
	buf = append(buf, []byte(found.RawHeading+"\n")...)

	for _, line := range found.Content {
		buf = append(buf, []byte(line+"\n")...)
	}

	digest := sha1.Sum(buf)
	return digest[:], true
}

func hashArtifactFile(filePath *pathutils.PathCfs) ([]byte, error) {
	reader, err := filereader.FileOpen(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	var buf []byte

	firstLine, err := filereader.FileReadLine(reader)
	if err != nil {
		if errors.Is(err, filereader.ErrEndOfFile) {
			filereader.FileClose(reader)
			digest := sha1.Sum(buf)
			return digest[:], nil
		}
		filereader.FileClose(reader)
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	if firstLine == "---" {
		for {
			line, err := filereader.FileReadLine(reader)
			if err != nil {
				if errors.Is(err, filereader.ErrEndOfFile) {
					filereader.FileClose(reader)
					digest := sha1.Sum(buf)
					return digest[:], nil
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
			buf = append(buf, []byte(line+"\n")...)
		}
	} else {
		buf = append(buf, []byte(firstLine+"\n")...)

		for {
			line, err := filereader.FileReadLine(reader)
			if err != nil {
				if errors.Is(err, filereader.ErrEndOfFile) {
					break
				}
				filereader.FileClose(reader)
				return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
			}
			buf = append(buf, []byte(line+"\n")...)
		}
	}

	filereader.FileClose(reader)
	digest := sha1.Sum(buf)
	return digest[:], nil
}

func hashExternalFile(pathString string) ([]byte, error) {
	cfsPath := &pathutils.PathCfs{Value: pathString}

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	var buf []byte

	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				break
			}
			filereader.FileClose(reader)
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		buf = append(buf, []byte(line+"\n")...)
	}

	filereader.FileClose(reader)
	digest := sha1.Sum(buf)
	return digest[:], nil
}

func ChainHashCompute(chain *chainresolver.Chain) (string, error) {
	var hashes [][]byte

	for _, ancestor := range chain.Ancestors {
		node, err := parsenode.NodeParse(ancestor.LogicalName)
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
		}

		if digest, ok := hashSpecSection(node, "public"); ok {
			hashes = append(hashes, digest)
		}
	}

	for _, dep := range chain.Dependencies {
		if logicalnames.LogicalNameIsArtifact(dep.LogicalName) {
			digest, err := hashArtifactFile(&dep.FilePath)
			if err != nil {
				return "", err
			}
			hashes = append(hashes, digest)
		} else if dep.Qualifier == nil {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
			}

			if digest, ok := hashSpecSection(node, "public"); ok {
				hashes = append(hashes, digest)
			}
		} else {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
			}

			if digest, ok := hashSpecSubsection(node, *dep.Qualifier); ok {
				hashes = append(hashes, digest)
			}
		}
	}

	for _, ext := range chain.External {
		digest, err := hashExternalFile(ext.Path)
		if err != nil {
			return "", err
		}
		hashes = append(hashes, digest)
	}

	targetNode, err := parsenode.NodeParse(chain.Target.LogicalName)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
	}

	if digest, ok := hashSpecSection(targetNode, "public"); ok {
		hashes = append(hashes, digest)
	}

	if digest, ok := hashSpecSection(targetNode, "agent"); ok {
		hashes = append(hashes, digest)
	}

	if chain.Input != nil {
		digest, err := hashArtifactFile(&chain.Input.FilePath)
		if err != nil {
			return "", err
		}
		hashes = append(hashes, digest)
	}

	var rawConcat []byte
	for _, h := range hashes {
		rawConcat = append(rawConcat, h...)
	}

	finalDigest := sha1.Sum(rawConcat)
	result := base64.RawURLEncoding.EncodeToString(finalDigest[:])
	return result, nil
}
