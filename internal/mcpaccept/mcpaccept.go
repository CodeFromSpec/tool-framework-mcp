// code-from-spec: SPEC/golang/implementation/mcp_tools/accept@ltZOJt_LUGsWDD3Pdb3fyJ_1dCo
package mcpaccept

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/manifest"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"
)

var ErrNotASpecReference = errors.New("not a SPEC/ reference")
var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")
var ErrNoOutput = errors.New("node has no output field")
var ErrNotModified = errors.New("artifact is not in modified status")

func MCPAccept(logicalName string) (string, error) {
	if !strings.HasPrefix(logicalName, "SPEC/") {
		return "", ErrNotASpecReference
	}

	node, err := parsing.ParseNode(logicalName)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrUnreadableFrontmatter, err)
	}

	if node.Frontmatter == nil || node.Frontmatter.Output == nil {
		return "", ErrNoOutput
	}

	artifactLogicalName := "ARTIFACT/" + strings.TrimPrefix(logicalName, "SPEC/")

	m, err := manifest.OpenManifest(false)
	if err != nil {
		return "", fmt.Errorf("opening manifest: %w", err)
	}

	entry, exists := m.Entries[artifactLogicalName]
	if !exists {
		if discardErr := m.Discard(); discardErr != nil {
			return "", fmt.Errorf("discarding manifest: %w", discardErr)
		}
		return "", ErrNotModified
	}

	artifactPath := oslayer.CfsPath(*node.Frontmatter.Output)
	handle, err := oslayer.OpenFile(artifactPath, "read", 30000)
	if err != nil {
		if discardErr := m.Discard(); discardErr != nil {
			return "", fmt.Errorf("discarding manifest: %w", discardErr)
		}
		return "", fmt.Errorf("opening artifact file: %w", err)
	}

	var contentBuilder strings.Builder
	for {
		line, readErr := handle.ReadLine()
		if readErr != nil {
			if errors.Is(readErr, oslayer.ErrEndOfFile) {
				break
			}
			handle.Close()
			if discardErr := m.Discard(); discardErr != nil {
				return "", fmt.Errorf("discarding manifest: %w", discardErr)
			}
			return "", fmt.Errorf("reading artifact file: %w", readErr)
		}
		contentBuilder.WriteString(line)
		contentBuilder.WriteString("\n")
	}
	handle.Close()

	normalized := contentBuilder.String()

	hasher := sha1.New()
	hasher.Write([]byte(normalized))
	sum := hasher.Sum(nil)
	checksum := base64.RawURLEncoding.EncodeToString(sum)[:27]

	if checksum == entry.Checksum {
		if discardErr := m.Discard(); discardErr != nil {
			return "", fmt.Errorf("discarding manifest: %w", discardErr)
		}
		return "", ErrNotModified
	}

	entry.Checksum = checksum
	m.Entries[artifactLogicalName] = entry

	if err := m.Save(); err != nil {
		return "", fmt.Errorf("saving manifest: %w", err)
	}

	return "accepted " + *node.Frontmatter.Output, nil
}
