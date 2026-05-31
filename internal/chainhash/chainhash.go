// code-from-spec: ROOT/golang/implementation/chain/hash@GzAAyYn_3OWm-1WQ_iT-Tox2C3Q
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

// ErrFileUnreadable is returned when a file in the chain cannot be read or opened.
var ErrFileUnreadable = errors.New("file unreadable")

// ErrParseFailure is returned when a node file cannot be parsed.
var ErrParseFailure = errors.New("parse failure")

// ChainHashCompute receives a Chain (as returned by ChainResolve) and returns
// a 27-character base64url encoded SHA-1 hash.
//
// The function reads each position's content from disk, computes a content
// hash (SHA-1) for each, concatenates all content hashes as raw bytes in
// chain assembly order, and computes the final SHA-1 of the concatenation.
//
// Errors:
//   - ErrFileUnreadable: a file in the chain cannot be read or opened.
//   - ErrParseFailure: a node file cannot be parsed.
//   - (FileReader.*): propagated from FileOpen.
//   - (NodeParsing.*): propagated from NodeParse.
func ChainHashCompute(chain *chainresolver.Chain) (string, error) {
	var digests [][]byte

	// Step 1a — Ancestors
	for _, ancestor := range chain.Ancestors {
		node, err := parsenode.NodeParse(ancestor.LogicalName)
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
		}
		digest, err := hashFullSection(node.Public)
		if err != nil {
			return "", err
		}
		if digest != nil {
			digests = append(digests, digest)
		}
	}

	// Step 1b — Dependencies
	for _, dep := range chain.Dependencies {
		if logicalnames.LogicalNameIsArtifact(dep.LogicalName) {
			digest, err := hashArtifactFile(dep.FilePath)
			if err != nil {
				return "", err
			}
			digests = append(digests, digest)
		} else if dep.Qualifier == nil {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
			}
			digest, err := hashFullSection(node.Public)
			if err != nil {
				return "", err
			}
			if digest != nil {
				digests = append(digests, digest)
			}
		} else {
			node, err := parsenode.NodeParse(dep.LogicalName)
			if err != nil {
				return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
			}
			digest, err := hashSubsection(node, *dep.Qualifier)
			if err != nil {
				return "", err
			}
			if digest != nil {
				digests = append(digests, digest)
			}
		}
	}

	// Step 1c — External entries
	for _, ext := range chain.External {
		digest, err := hashExternalEntry(ext)
		if err != nil {
			return "", err
		}
		if digest != nil {
			digests = append(digests, digest)
		}
	}

	// Step 1d — Target # Public section
	targetNode, err := parsenode.NodeParse(chain.Target.LogicalName)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrParseFailure, err)
	}
	publicDigest, err := hashFullSection(targetNode.Public)
	if err != nil {
		return "", err
	}
	if publicDigest != nil {
		digests = append(digests, publicDigest)
	}

	// Step 1e — Target # Agent section
	agentDigest, err := hashFullSection(targetNode.Agent)
	if err != nil {
		return "", err
	}
	if agentDigest != nil {
		digests = append(digests, agentDigest)
	}

	// Step 1f — Input
	if chain.Input != nil {
		inputDigest, err := hashArtifactFile(chain.Input.FilePath)
		if err != nil {
			return "", err
		}
		digests = append(digests, inputDigest)
	}

	// Step 2 — Compute final hash
	var combined []byte
	for _, d := range digests {
		combined = append(combined, d...)
	}
	final := sha1.Sum(combined)
	return base64.RawURLEncoding.EncodeToString(final[:]), nil
}

// hashFullSection hashes a full NodeSection (e.g. # Public or # Agent).
// Returns nil if the section is absent or has no content and no subsections.
func hashFullSection(section *parsenode.NodeSection) ([]byte, error) {
	if section == nil {
		return nil, nil
	}
	// If there is no content and no subsections, skip.
	if len(section.Content) == 0 && len(section.Subsections) == 0 {
		return nil, nil
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
	return digest[:], nil
}

// hashSubsection hashes a ## <qualifier> subsection within the # Public section.
// Returns nil if the public section is absent or the subsection is not found.
func hashSubsection(node *parsenode.Node, qualifier string) ([]byte, error) {
	if node.Public == nil {
		return nil, nil
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
		return nil, nil
	}

	var buf []byte
	buf = append(buf, []byte(found.RawHeading+"\n")...)
	for _, line := range found.Content {
		buf = append(buf, []byte(line+"\n")...)
	}

	digest := sha1.Sum(buf)
	return digest[:], nil
}

// hashArtifactFile hashes an artifact file at the given PathCfs,
// skipping any YAML frontmatter block delimited by "---" at the top.
func hashArtifactFile(filePath *pathutils.PathCfs) ([]byte, error) {
	reader, err := filereader.FileOpen(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	firstLine, err := filereader.FileReadLine(reader)
	if errors.Is(err, filereader.ErrEndOfFile) {
		filereader.FileClose(reader)
		empty := sha1.Sum([]byte{})
		return empty[:], nil
	}
	if err != nil {
		filereader.FileClose(reader)
		return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
	}

	var buf []byte

	if firstLine == "---" {
		// Skip lines until the closing "---" delimiter.
		for {
			line, err := filereader.FileReadLine(reader)
			if errors.Is(err, filereader.ErrEndOfFile) {
				break
			}
			if err != nil {
				filereader.FileClose(reader)
				return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
			}
			if line == "---" {
				break
			}
		}
	} else {
		// First line is content — include it in the buffer.
		buf = append(buf, []byte(firstLine+"\n")...)
	}

	// Read all remaining lines.
	for {
		line, err := filereader.FileReadLine(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			break
		}
		if err != nil {
			filereader.FileClose(reader)
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}
		buf = append(buf, []byte(line+"\n")...)
	}

	filereader.FileClose(reader)

	digest := sha1.Sum(buf)
	return digest[:], nil
}

// hashExternalEntry hashes an external file reference, with or without fragments.
func hashExternalEntry(ext *frontmatter.FrontmatterExternal) ([]byte, error) {
	cfsPath := &pathutils.PathCfs{Value: ext.Path}

	if len(ext.Fragments) == 0 {
		// Hash the entire file.
		reader, err := filereader.FileOpen(cfsPath)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}

		var buf []byte
		for {
			line, err := filereader.FileReadLine(reader)
			if errors.Is(err, filereader.ErrEndOfFile) {
				break
			}
			if err != nil {
				filereader.FileClose(reader)
				return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
			}
			buf = append(buf, []byte(line+"\n")...)
		}
		filereader.FileClose(reader)

		digest := sha1.Sum(buf)
		return digest[:], nil
	}

	// Hash with fragments — concatenate all fragment line ranges into one buffer.
	var buf []byte
	for _, frag := range ext.Fragments {
		start, end, err := parseLineRange(frag.Lines)
		if err != nil {
			return nil, fmt.Errorf("invalid fragment line range %q: %w", frag.Lines, err)
		}

		reader, err := filereader.FileOpen(cfsPath)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
		}

		filereader.FileSkipLines(reader, start-1)

		count := end - start + 1
		for i := 0; i < count; i++ {
			line, err := filereader.FileReadLine(reader)
			if errors.Is(err, filereader.ErrEndOfFile) {
				break
			}
			if err != nil {
				filereader.FileClose(reader)
				return nil, fmt.Errorf("%w: %w", ErrFileUnreadable, err)
			}
			buf = append(buf, []byte(line+"\n")...)
		}
		filereader.FileClose(reader)
	}

	digest := sha1.Sum(buf)
	return digest[:], nil
}

// parseLineRange parses a "start-end" line range string into 1-based inclusive integers.
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
