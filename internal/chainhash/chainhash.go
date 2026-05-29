// code-from-spec: ROOT/golang/implementation/chain/hash@dTEBIlnr1kVZmdQukk2ZSgJSORA

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

var (
	// ErrUnreadableFile is returned when a file in the chain cannot be
	// read or opened.
	ErrUnreadableFile = errors.New("file unreadable")

	// ErrParseFailure is returned when a node file cannot be parsed.
	ErrParseFailure = errors.New("parse failure")
)

// contentHash holds a SHA-1 digest as raw bytes.
type contentHash struct {
	rawBytes [20]byte
}

// hashSection computes SHA-1 for a full section (# Public or # Agent).
// Returns nil if the section is absent or has no content and no subsections.
func hashSection(section *parsenode.NodeSection) *contentHash {
	if section == nil {
		return nil
	}
	if section.Content == "" && len(section.Subsections) == 0 {
		return nil
	}

	var buf strings.Builder

	buf.WriteString(section.Heading + "\n")

	if section.Content != "" {
		for _, line := range strings.Split(section.Content, "\n") {
			buf.WriteString(line + "\n")
		}
	}

	for _, sub := range section.Subsections {
		buf.WriteString(sub.Heading + "\n")
		if sub.Content != "" {
			for _, line := range strings.Split(sub.Content, "\n") {
				buf.WriteString(line + "\n")
			}
		}
	}

	digest := sha1.Sum([]byte(buf.String()))
	return &contentHash{rawBytes: digest}
}

// hashSubsection computes SHA-1 for a specific ## subsection within a section.
// Returns nil if the section is absent or the qualifier does not match any subsection.
func hashSubsection(section *parsenode.NodeSection, qualifier string) *contentHash {
	if section == nil {
		return nil
	}

	normalizedQualifier := textnormalization.NormalizeText(qualifier)

	var matchedSub *parsenode.NodeSubsection
	for _, sub := range section.Subsections {
		if sub.Heading == normalizedQualifier {
			matchedSub = sub
			break
		}
	}
	if matchedSub == nil {
		return nil
	}

	var buf strings.Builder

	buf.WriteString(matchedSub.Heading + "\n")

	if matchedSub.Content != "" {
		for _, line := range strings.Split(matchedSub.Content, "\n") {
			buf.WriteString(line + "\n")
		}
	}

	digest := sha1.Sum([]byte(buf.String()))
	return &contentHash{rawBytes: digest}
}

// hashFileStrippingFrontmatter reads the file at file_path, strips the YAML
// frontmatter block if present, and returns the SHA-1 of the remaining content.
func hashFileStrippingFrontmatter(filePath *pathutils.PathCfs) (*contentHash, error) {
	reader, err := filereader.FileOpen(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnreadableFile, err)
	}

	firstLine, err := filereader.FileReadLine(reader)
	if err != nil {
		// end of file — empty file
		filereader.FileClose(reader)
		digest := sha1.Sum([]byte{})
		return &contentHash{rawBytes: digest}, nil
	}

	var buf strings.Builder

	if firstLine == "---" {
		// Skip frontmatter until closing "---"
		for {
			line, err := filereader.FileReadLine(reader)
			if err != nil {
				// end of file before closing "---"
				filereader.FileClose(reader)
				return nil, fmt.Errorf("%w: frontmatter not closed", ErrUnreadableFile)
			}
			if line == "---" {
				break
			}
		}
	} else {
		// First line is content
		buf.WriteString(firstLine + "\n")
	}

	// Read remaining lines
	for {
		line, err := filereader.FileReadLine(reader)
		if err != nil {
			break
		}
		buf.WriteString(line + "\n")
	}

	filereader.FileClose(reader)

	digest := sha1.Sum([]byte(buf.String()))
	return &contentHash{rawBytes: digest}, nil
}

// hashExternalFile computes SHA-1 for an external file entry.
// If the entry has no fragments, the entire file is hashed.
// If it has fragments, only the specified line ranges are hashed.
func hashExternalFile(external *frontmatter.FrontmatterExternal) (*contentHash, error) {
	filePath := &pathutils.PathCfs{Value: external.Path}

	if len(external.Fragments) == 0 {
		reader, err := filereader.FileOpen(filePath)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrUnreadableFile, err)
		}

		var buf strings.Builder
		for {
			line, err := filereader.FileReadLine(reader)
			if err != nil {
				break
			}
			buf.WriteString(line + "\n")
		}
		filereader.FileClose(reader)

		digest := sha1.Sum([]byte(buf.String()))
		return &contentHash{rawBytes: digest}, nil
	}

	// Has fragments
	var buf strings.Builder
	for _, fragment := range external.Fragments {
		start, end, err := parseLineRange(fragment.Lines)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid fragment lines %q: %v", ErrUnreadableFile, fragment.Lines, err)
		}

		reader, err := filereader.FileOpen(filePath)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrUnreadableFile, err)
		}

		filereader.FileSkipLines(reader, start-1)

		count := end - start + 1
		for i := 0; i < count; i++ {
			line, err := filereader.FileReadLine(reader)
			if err != nil {
				filereader.FileClose(reader)
				return nil, fmt.Errorf("%w: unexpected end of file reading fragment", ErrUnreadableFile)
			}
			buf.WriteString(line + "\n")
		}

		filereader.FileClose(reader)
	}

	digest := sha1.Sum([]byte(buf.String()))
	return &contentHash{rawBytes: digest}, nil
}

// parseLineRange parses a "<start>-<end>" string into two integers.
func parseLineRange(s string) (int, int, error) {
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected format <start>-<end>, got %q", s)
	}
	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start: %w", err)
	}
	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end: %w", err)
	}
	return start, end, nil
}

// ChainHashCompute receives a Chain (as returned by ChainResolve) and
// returns a 27-character base64url encoded SHA-1 hash.
//
// The function reads each position's content from disk, computes a
// content hash (SHA-1) for each, concatenates all content hashes as
// raw bytes in chain assembly order, and computes the final SHA-1 of
// the concatenation.
//
// Chain assembly order:
//  1. Ancestors — from root down to (but not including) the target node.
//  2. Dependencies — from the target's depends_on, sorted alphabetically
//     by file path then by qualifier.
//  3. External — from the target's external, sorted alphabetically by path.
//  4. Target — the target node itself.
//  5. Input — the target's input artifact, if present.
//
// Possible errors:
//   - ErrUnreadableFile — a file in the chain cannot be read or opened.
//   - ErrParseFailure — a node file cannot be parsed.
func ChainHashCompute(chain *chainresolver.Chain) (string, error) {
	var contentHashes []*contentHash

	// 1a. Ancestors
	for _, ancestor := range chain.Ancestors {
		node, err := parsenode.NodeParse(ancestor.LogicalName)
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
		}
		h := hashSection(node.Public)
		if h != nil {
			contentHashes = append(contentHashes, h)
		}
	}

	// 1b. Dependencies
	for _, dep := range chain.Dependencies {
		if logicalnames.LogicalNameIsArtifact(dep.LogicalName) {
			h, err := hashFileStrippingFrontmatter(dep.FilePath)
			if err != nil {
				return "", err
			}
			contentHashes = append(contentHashes, h)
		} else if dep.Qualifier == nil {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
			}
			h := hashSection(node.Public)
			if h != nil {
				contentHashes = append(contentHashes, h)
			}
		} else {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
			}
			h := hashSubsection(node.Public, *dep.Qualifier)
			if h != nil {
				contentHashes = append(contentHashes, h)
			}
		}
	}

	// 1c. External files
	for _, external := range chain.External {
		h, err := hashExternalFile(external)
		if err != nil {
			return "", err
		}
		contentHashes = append(contentHashes, h)
	}

	// 1d. Target # Public and 1e. Target # Agent
	targetNode, err := parsenode.NodeParse(chain.Target.LogicalName)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
	}

	h := hashSection(targetNode.Public)
	if h != nil {
		contentHashes = append(contentHashes, h)
	}

	h = hashSection(targetNode.Agent)
	if h != nil {
		contentHashes = append(contentHashes, h)
	}

	// 1f. Input
	if chain.Input != nil {
		h, err := hashFileStrippingFrontmatter(chain.Input.FilePath)
		if err != nil {
			return "", err
		}
		contentHashes = append(contentHashes, h)
	}

	// Step 2 — Compute final hash
	var combinedBuf []byte
	for _, ch := range contentHashes {
		combinedBuf = append(combinedBuf, ch.rawBytes[:]...)
	}

	finalDigest := sha1.Sum(combinedBuf)
	encoded := base64.RawURLEncoding.EncodeToString(finalDigest[:])

	return encoded, nil
}
