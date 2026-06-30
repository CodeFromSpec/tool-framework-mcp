package mcpaccept

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/manifest"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"
)

var ErrNotASpecReference = errors.New("not a SPEC/ reference")
var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")
var ErrNoOutput = errors.New("node has no output field")
var ErrAlreadyUpToDate = errors.New("artifact is already up to date")

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

	artifactPath := oslayer.CfsPath(*node.Frontmatter.Output)
	handle, err := oslayer.OpenFile(artifactPath, "read", 30000)
	if err != nil {
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

	chain, err := chainresolver.ChainResolve(logicalName)
	if err != nil {
		return "", fmt.Errorf("resolving chain: %w", err)
	}

	chainHash, _, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		return "", fmt.Errorf("computing chain hash: %w", err)
	}

	m, err := manifest.OpenManifest(false)
	if err != nil {
		return "", fmt.Errorf("opening manifest: %w", err)
	}
	defer func() { _ = m.Discard() }()

	entry, exists := m.Entries[artifactLogicalName]
	if !exists {
		m.Entries[artifactLogicalName] = manifest.ManifestEntry{
			Path:      *node.Frontmatter.Output,
			Checksum:  checksum,
			ChainHash: chainHash,
		}

		if err := m.Save(); err != nil {
			return "", fmt.Errorf("saving manifest: %w", err)
		}

		return "accepted " + *node.Frontmatter.Output, nil
	}

	if entry.Checksum == checksum && entry.ChainHash == chainHash {
		return "", ErrAlreadyUpToDate
	}

	entry.Checksum = checksum
	entry.ChainHash = chainHash
	m.Entries[artifactLogicalName] = entry

	if err := m.Save(); err != nil {
		return "", fmt.Errorf("saving manifest: %w", err)
	}

	return "accepted " + *node.Frontmatter.Output, nil
}
