// code-from-spec: ROOT/golang/implementation/chain/hash@3MWAvktrv6laaHccoWONrxHop3s
package chainhash

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/textnormalization"
)

// ChainHashCompute receives a Chain (as returned by chainresolver.ChainResolve)
// and returns a 27-character base64url-encoded SHA-1 hash.
func ChainHashCompute(chain *chainresolver.Chain) (string, error) {
	var contentHashes [][]byte

	// 2. Ancestors
	for _, ancestor := range chain.Ancestors {
		node, err := parsenode.NodeParse(ancestor.LogicalName)
		if err != nil {
			return "", fmt.Errorf("parse failure for ancestor %s: %w", ancestor.LogicalName, err)
		}
		h := hashSection(node.Public)
		if h != nil {
			contentHashes = append(contentHashes, h)
		}
	}

	// 3. Dependencies
	for _, dep := range chain.Dependencies {
		if logicalnames.LogicalNameIsArtifact(dep.LogicalName) {
			h, err := hashArtifactFile(dep.FilePath)
			if err != nil {
				return "", fmt.Errorf("file unreadable for dependency artifact %s: %w", dep.LogicalName, err)
			}
			contentHashes = append(contentHashes, h)
		} else if dep.Qualifier == nil {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return "", fmt.Errorf("parse failure for dependency %s: %w", dep.LogicalName, err)
			}
			h := hashSection(node.Public)
			if h != nil {
				contentHashes = append(contentHashes, h)
			}
		} else {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return "", fmt.Errorf("parse failure for dependency %s: %w", dep.LogicalName, err)
			}
			h := hashSubsection(node.Public, *dep.Qualifier)
			if h != nil {
				contentHashes = append(contentHashes, h)
			}
		}
	}

	// 4. External
	for _, ext := range chain.External {
		if len(ext.Fragments) == 0 {
			h, err := hashExternalFile(ext.Path)
			if err != nil {
				return "", fmt.Errorf("file unreadable for external %s: %w", ext.Path, err)
			}
			contentHashes = append(contentHashes, h)
		} else {
			h, err := hashExternalFileFragments(ext.Path, ext.Fragments)
			if err != nil {
				return "", fmt.Errorf("file unreadable for external fragments %s: %w", ext.Path, err)
			}
			contentHashes = append(contentHashes, h)
		}
	}

	// 5. Target # Public and 6. Target # Agent
	targetNode, err := parsenode.NodeParse(chain.Target.LogicalName)
	if err != nil {
		return "", fmt.Errorf("parse failure for target %s: %w", chain.Target.LogicalName, err)
	}

	h := hashSection(targetNode.Public)
	if h != nil {
		contentHashes = append(contentHashes, h)
	}

	h = hashSection(targetNode.Agent)
	if h != nil {
		contentHashes = append(contentHashes, h)
	}

	// 7. Input
	if chain.Input != nil {
		h, err := hashArtifactFile(chain.Input.FilePath)
		if err != nil {
			return "", fmt.Errorf("file unreadable for input %s: %w", chain.Input.LogicalName, err)
		}
		contentHashes = append(contentHashes, h)
	}

	// 8. Final hash
	var combined bytes.Buffer
	for _, ch := range contentHashes {
		combined.Write(ch)
	}

	finalHash := sha1.Sum(combined.Bytes())
	encoded := base64.RawURLEncoding.EncodeToString(finalHash[:])

	return encoded[:27], nil
}

// hashSection computes a SHA-1 content hash for a full section.
// Returns nil if the section is absent or contributes no content.
func hashSection(section *parsenode.NodeSection) []byte {
	if section == nil {
		return nil
	}
	if len(section.Content) == 0 && len(section.Subsections) == 0 {
		return nil
	}

	var buf bytes.Buffer

	buf.WriteString(section.RawHeading)
	buf.WriteByte('\n')

	for _, line := range section.Content {
		buf.WriteString(line)
		buf.WriteByte('\n')
	}

	for _, sub := range section.Subsections {
		buf.WriteString(sub.RawHeading)
		buf.WriteByte('\n')
		for _, line := range sub.Content {
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
	}

	h := sha1.Sum(buf.Bytes())
	return h[:]
}

// hashSubsection computes a SHA-1 content hash for a specific ## subsection.
// Returns nil if the section is absent or the subsection is not found.
func hashSubsection(section *parsenode.NodeSection, qualifier string) []byte {
	if section == nil {
		return nil
	}

	targetHeading := textnormalization.NormalizeText(qualifier)

	var found *parsenode.NodeSubsection
	for _, sub := range section.Subsections {
		if sub.Heading == targetHeading {
			found = sub
			break
		}
	}
	if found == nil {
		return nil
	}

	var buf bytes.Buffer

	buf.WriteString(found.RawHeading)
	buf.WriteByte('\n')

	for _, line := range found.Content {
		buf.WriteString(line)
		buf.WriteByte('\n')
	}

	h := sha1.Sum(buf.Bytes())
	return h[:]
}

// hashArtifactFile computes a SHA-1 hash of an artifact file's content,
// stripping frontmatter if present.
func hashArtifactFile(filePath *pathutils.PathCfs) ([]byte, error) {
	reader, err := filereader.FileOpen(filePath)
	if err != nil {
		return nil, fmt.Errorf("file unreadable %s: %w", filePath.Value, err)
	}

	firstLine, err := filereader.FileReadLine(reader)
	if err != nil {
		if errors.Is(err, filereader.ErrEndOfFile) {
			filereader.FileClose(reader)
			h := sha1.Sum([]byte{})
			return h[:], nil
		}
		filereader.FileClose(reader)
		return nil, fmt.Errorf("file unreadable %s: %w", filePath.Value, err)
	}

	var buf bytes.Buffer
	var firstContentLine string
	hasFrontmatter := firstLine == "---"

	if hasFrontmatter {
		// Skip until closing "---"
		for {
			line, err := filereader.FileReadLine(reader)
			if err != nil {
				if errors.Is(err, filereader.ErrEndOfFile) {
					filereader.FileClose(reader)
					h := sha1.Sum([]byte{})
					return h[:], nil
				}
				filereader.FileClose(reader)
				return nil, fmt.Errorf("file unreadable %s: %w", filePath.Value, err)
			}
			if line == "---" {
				break
			}
		}
	} else {
		firstContentLine = firstLine
	}

	if firstContentLine != "" {
		buf.WriteString(firstContentLine)
		buf.WriteByte('\n')
	}

	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				break
			}
			filereader.FileClose(reader)
			return nil, fmt.Errorf("file unreadable %s: %w", filePath.Value, err)
		}
		buf.WriteString(line)
		buf.WriteByte('\n')
	}

	filereader.FileClose(reader)

	h := sha1.Sum(buf.Bytes())
	return h[:], nil
}

// hashExternalFile computes a SHA-1 hash of a full external file's content.
func hashExternalFile(path string) ([]byte, error) {
	cfsPath := &pathutils.PathCfs{Value: path}

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		return nil, fmt.Errorf("file unreadable %s: %w", path, err)
	}

	var buf bytes.Buffer

	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				break
			}
			filereader.FileClose(reader)
			return nil, fmt.Errorf("file unreadable %s: %w", path, err)
		}
		buf.WriteString(line)
		buf.WriteByte('\n')
	}

	filereader.FileClose(reader)

	h := sha1.Sum(buf.Bytes())
	return h[:], nil
}

// hashExternalFileFragments computes a SHA-1 hash of selected line ranges
// from an external file, concatenated in declaration order.
func hashExternalFileFragments(path string, fragments []*frontmatter.FrontmatterExternalFragment) ([]byte, error) {
	var buf bytes.Buffer

	for _, fragment := range fragments {
		start, end, err := parseLineRange(fragment.Lines)
		if err != nil {
			return nil, fmt.Errorf("invalid fragment line range %q: %w", fragment.Lines, err)
		}

		cfsPath := &pathutils.PathCfs{Value: path}
		reader, err := filereader.FileOpen(cfsPath)
		if err != nil {
			return nil, fmt.Errorf("file unreadable %s: %w", path, err)
		}

		filereader.FileSkipLines(reader, start-1)

		count := end - start + 1
		for i := 0; i < count; i++ {
			line, err := filereader.FileReadLine(reader)
			if err != nil {
				if errors.Is(err, filereader.ErrEndOfFile) {
					filereader.FileClose(reader)
					break
				}
				filereader.FileClose(reader)
				return nil, fmt.Errorf("file unreadable %s: %w", path, err)
			}
			buf.WriteString(line)
			buf.WriteByte('\n')
		}

		filereader.FileClose(reader)
	}

	h := sha1.Sum(buf.Bytes())
	return h[:], nil
}

// parseLineRange parses a "<start>-<end>" line range string.
func parseLineRange(lines string) (start, end int, err error) {
	parts := strings.SplitN(lines, "-", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected format <start>-<end>, got %q", lines)
	}

	start, err = strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start line %q: %w", parts[0], err)
	}

	end, err = strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end line %q: %w", parts[1], err)
	}

	return start, end, nil
}
