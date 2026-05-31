// code-from-spec: ROOT/golang/implementation/chain/hash@PHFdvHjeWO1xGvPVtyoTIhuEApw
package chainhash

import (
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

// ErrFileUnreadable is returned when a file in the chain cannot be
// read or opened.
var ErrFileUnreadable = errors.New("file unreadable")

// ErrParseFailure is returned when a node file cannot be parsed.
var ErrParseFailure = errors.New("parse failure")

// ChainHashCompute receives a Chain (as returned by ChainResolve) and
// returns a 27-character base64url encoded SHA-1 hash.
//
// The function reads each position's content from disk, computes a
// content hash (SHA-1) for each file, concatenates all content hashes
// as raw bytes in chain assembly order, and computes the final SHA-1
// of the concatenation.
//
// Errors:
//   - ErrFileUnreadable: a file in the chain cannot be read or opened.
//   - ErrParseFailure: a node file cannot be parsed.
//   - (FileReader.*): propagated from FileOpen.
//   - (NodeParsing.*): propagated from NodeParse.
func ChainHashCompute(chain *chainresolver.Chain) (string, error) {
	var contentHashes [][]byte

	// Step 2: Process ancestors.
	for _, ancestor := range chain.Ancestors {
		node, err := parsenode.NodeParse(ancestor.LogicalName)
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
		}
		hashBytes, err := hashFullSection(node.Public)
		if err != nil {
			return "", err
		}
		if hashBytes != nil {
			contentHashes = append(contentHashes, hashBytes)
		}
	}

	// Step 3: Process dependencies.
	for _, dep := range chain.Dependencies {
		if logicalnames.LogicalNameIsArtifact(dep.LogicalName) {
			hashBytes, err := hashArtifactFile(dep.FilePath)
			if err != nil {
				return "", err
			}
			if hashBytes != nil {
				contentHashes = append(contentHashes, hashBytes)
			}
		} else if dep.Qualifier == "" {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
			}
			hashBytes, err := hashFullSection(node.Public)
			if err != nil {
				return "", err
			}
			if hashBytes != nil {
				contentHashes = append(contentHashes, hashBytes)
			}
		} else {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
			}
			hashBytes, err := hashSubsection(node.Public, dep.Qualifier)
			if err != nil {
				return "", err
			}
			if hashBytes != nil {
				contentHashes = append(contentHashes, hashBytes)
			}
		}
	}

	// Step 4: Process external entries.
	for _, ext := range chain.External {
		if len(ext.Fragments) == 0 {
			hashBytes, err := hashExternalFile(ext.Path)
			if err != nil {
				return "", err
			}
			if hashBytes != nil {
				contentHashes = append(contentHashes, hashBytes)
			}
		} else {
			hashBytes, err := hashExternalFragments(ext.Path, ext.Fragments)
			if err != nil {
				return "", err
			}
			if hashBytes != nil {
				contentHashes = append(contentHashes, hashBytes)
			}
		}
	}

	// Step 5: Process target.
	targetNode, err := parsenode.NodeParse(chain.Target.LogicalName)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
	}
	publicHash, err := hashFullSection(targetNode.Public)
	if err != nil {
		return "", err
	}
	if publicHash != nil {
		contentHashes = append(contentHashes, publicHash)
	}
	agentHash, err := hashFullSection(targetNode.Agent)
	if err != nil {
		return "", err
	}
	if agentHash != nil {
		contentHashes = append(contentHashes, agentHash)
	}

	// Step 6: Process input.
	if chain.Input != nil {
		hashBytes, err := hashArtifactFile(chain.Input.FilePath)
		if err != nil {
			return "", err
		}
		if hashBytes != nil {
			contentHashes = append(contentHashes, hashBytes)
		}
	}

	// Step 7: Compute final hash.
	var combined []byte
	for _, h := range contentHashes {
		combined = append(combined, h...)
	}
	finalHash := sha1.Sum(combined)
	encoded := base64.RawURLEncoding.EncodeToString(finalHash[:])
	return encoded, nil
}

// hashFullSection hashes a full spec section (# Public or # Agent).
// Returns nil when section is absent or has no content.
func hashFullSection(section *parsenode.NodeSection) ([]byte, error) {
	if section == nil {
		return nil, nil
	}
	if len(section.Content) == 0 && len(section.Subsections) == 0 {
		return nil, nil
	}

	var buf strings.Builder
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

	hash := sha1.Sum([]byte(buf.String()))
	return hash[:], nil
}

// hashSubsection hashes a specific ## subsection within # Public,
// matched by normalized heading.
// Returns nil when public_section is absent or the subsection is not found.
func hashSubsection(publicSection *parsenode.NodeSection, qualifier string) ([]byte, error) {
	if publicSection == nil {
		return nil, nil
	}

	targetHeading := textnormalization.NormalizeText(qualifier)

	var found *parsenode.NodeSubsection
	for _, sub := range publicSection.Subsections {
		if sub.Heading == targetHeading {
			found = sub
			break
		}
	}
	if found == nil {
		return nil, nil
	}

	var buf strings.Builder
	buf.WriteString(found.RawHeading)
	buf.WriteByte('\n')
	for _, line := range found.Content {
		buf.WriteString(line)
		buf.WriteByte('\n')
	}

	hash := sha1.Sum([]byte(buf.String()))
	return hash[:], nil
}

// hashArtifactFile hashes an artifact file, stripping any frontmatter block at the top.
// Returns nil when the file is empty or has no content after frontmatter.
func hashArtifactFile(filePath pathutils.PathCfs) ([]byte, error) {
	reader, err := filereader.FileOpen(&filePath)
	if err != nil {
		if errors.Is(err, filereader.ErrFileUnreadable) {
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	firstLine, err := filereader.FileReadLine(reader)
	if err != nil {
		if errors.Is(err, filereader.ErrEndOfFile) {
			filereader.FileClose(reader)
			return nil, nil
		}
		filereader.FileClose(reader)
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	var buf strings.Builder

	if firstLine == "---" {
		// Skip frontmatter lines until closing "---" or EOF.
		foundClose := false
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
				foundClose = true
				break
			}
		}
		if !foundClose {
			filereader.FileClose(reader)
			return nil, nil
		}
		// Read remaining lines after frontmatter.
		for {
			line, err := filereader.FileReadLine(reader)
			if err != nil {
				if errors.Is(err, filereader.ErrEndOfFile) {
					break
				}
				filereader.FileClose(reader)
				return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
			}
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
	} else {
		// No frontmatter; include first line and remaining.
		buf.WriteString(firstLine)
		buf.WriteByte('\n')
		for {
			line, err := filereader.FileReadLine(reader)
			if err != nil {
				if errors.Is(err, filereader.ErrEndOfFile) {
					break
				}
				filereader.FileClose(reader)
				return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
			}
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
	}

	filereader.FileClose(reader)

	if buf.Len() == 0 {
		return nil, nil
	}

	hash := sha1.Sum([]byte(buf.String()))
	return hash[:], nil
}

// hashExternalFile hashes the full content of an external file.
// Returns nil when the file is empty.
func hashExternalFile(path string) ([]byte, error) {
	cfsPath := &pathutils.PathCfs{Value: path}

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	var buf strings.Builder
	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			if errors.Is(err, filereader.ErrEndOfFile) {
				break
			}
			filereader.FileClose(reader)
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		buf.WriteString(line)
		buf.WriteByte('\n')
	}

	filereader.FileClose(reader)

	if buf.Len() == 0 {
		return nil, nil
	}

	hash := sha1.Sum([]byte(buf.String()))
	return hash[:], nil
}

// hashExternalFragments hashes selected line ranges from an external file,
// concatenated in declaration order.
// Returns nil when the combined content is empty.
func hashExternalFragments(path string, fragments []*frontmatter.FrontmatterExternalFragment) ([]byte, error) {
	cfsPath := &pathutils.PathCfs{Value: path}

	var buf strings.Builder

	for _, fragment := range fragments {
		// Parse "<start>-<end>" from fragment.lines.
		parts := strings.SplitN(fragment.Lines, "-", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("%w: invalid fragment lines format %q", ErrFileUnreadable, fragment.Lines)
		}
		start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, fmt.Errorf("%w: invalid fragment start line %q: %v", ErrFileUnreadable, parts[0], err)
		}
		end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("%w: invalid fragment end line %q: %v", ErrFileUnreadable, parts[1], err)
		}

		reader, err := filereader.FileOpen(cfsPath)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}

		// Skip lines before the range (start is 1-based).
		filereader.FileSkipLines(reader, start-1)

		// Read end - start + 1 lines.
		count := end - start + 1
		for i := 0; i < count; i++ {
			line, err := filereader.FileReadLine(reader)
			if err != nil {
				if errors.Is(err, filereader.ErrEndOfFile) {
					break
				}
				filereader.FileClose(reader)
				return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
			}
			buf.WriteString(line)
			buf.WriteByte('\n')
		}

		filereader.FileClose(reader)
	}

	if buf.Len() == 0 {
		return nil, nil
	}

	hash := sha1.Sum([]byte(buf.String()))
	return hash[:], nil
}
