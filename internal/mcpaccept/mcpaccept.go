package mcpaccept

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/manifest"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/parsing"
)

var ErrInvalidPrefix = errors.New("logical name must start with ARTIFACT/ or VERDICT/")
var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")
var ErrNoOutput = errors.New("node has no type field")
var ErrAlreadyUpToDate = errors.New("entry is already up to date")

func MCPAccept(logicalName string) (string, error) {
	var specName string
	var isVerdict bool

	if strings.HasPrefix(logicalName, "ARTIFACT/") {
		specName = "SPEC/" + strings.TrimPrefix(logicalName, "ARTIFACT/")
		isVerdict = false
	} else if strings.HasPrefix(logicalName, "VERDICT/") {
		specName = "SPEC/" + strings.TrimPrefix(logicalName, "VERDICT/")
		isVerdict = true
	} else {
		return "", ErrInvalidPrefix
	}

	node, err := parsing.ParseNode(specName)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrUnreadableFrontmatter, err)
	}

	resolvedOutput := parsing.ResolvedOutput(node)
	if resolvedOutput == nil {
		return "", ErrNoOutput
	}

	artifactPath := oslayer.CfsPath(*resolvedOutput)
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

	chain, err := chainresolver.ChainResolve(specName)
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

	entry, exists := m.Entries[logicalName]
	if !exists {
		newEntry := manifest.ManifestEntry{
			Path:      *resolvedOutput,
			Checksum:  checksum,
			ChainHash: chainHash,
		}
		if isVerdict {
			newEntry.Result = "accepted"
		}
		m.Entries[logicalName] = newEntry

		if err := m.Save(); err != nil {
			return "", fmt.Errorf("saving manifest: %w", err)
		}

		return "accepted " + *resolvedOutput, nil
	}

	if entry.Checksum == checksum && entry.ChainHash == chainHash {
		if isVerdict && entry.Result != "accepted" {
			entry.Result = "accepted"
			m.Entries[logicalName] = entry

			if err := m.Save(); err != nil {
				return "", fmt.Errorf("saving manifest: %w", err)
			}

			return "accepted " + *resolvedOutput, nil
		}
		return "", ErrAlreadyUpToDate
	}

	entry.Checksum = checksum
	entry.ChainHash = chainHash
	if isVerdict {
		entry.Result = "accepted"
	}
	m.Entries[logicalName] = entry

	if err := m.Save(); err != nil {
		return "", fmt.Errorf("saving manifest: %w", err)
	}

	return "accepted " + *resolvedOutput, nil
}
