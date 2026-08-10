package mcpwriteartifact

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/cache"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/chainresolver"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/manifest"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/parsing"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/subagenttoken"
)

var (
	ErrUnreadableFrontmatter = errors.New("node frontmatter cannot be parsed")
	ErrNoOutput              = errors.New("node has no type field")
	ErrNotAnArtifact         = errors.New("node type is not artifact")
)

func MCPWriteArtifact(token, content string) (string, error) {
	logicalName, err := subagenttoken.SubagentTokenValidate(token)
	if err != nil {
		return "", err
	}

	node, err := parsing.ParseNode(logicalName)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrUnreadableFrontmatter, err)
	}

	if node.Frontmatter == nil || node.Frontmatter.Type == nil {
		return "", ErrNoOutput
	}

	if *node.Frontmatter.Type != "artifact" {
		return "", ErrNotAnArtifact
	}

	resolvedOutput := parsing.ResolvedOutput(node)
	if resolvedOutput == nil {
		return "", ErrNoOutput
	}

	path := *resolvedOutput

	if err := oslayer.ValidateStringIsCfsPath(path); err != nil {
		return "", err
	}

	cfsPath := oslayer.CfsPath(path)
	handle, err := oslayer.OpenFile(cfsPath, "overwrite", 30000)
	if err != nil {
		return "", err
	}

	if err := handle.Write(content); err != nil {
		handle.Close()
		return "", err
	}

	handle.Close()

	checksum := computeChecksum(content)

	chain, err := chainresolver.ChainResolve(logicalName)
	if err != nil {
		return "", err
	}

	chainHash, positions, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		return "", err
	}

	m, err := manifest.OpenManifest(false)
	if err != nil {
		return "", err
	}
	defer func() { _ = m.Discard() }()

	artifactName := "ARTIFACT/" + strings.TrimPrefix(logicalName, "SPEC/")
	m.Entries[artifactName] = manifest.ManifestEntry{
		Path:      path,
		Checksum:  checksum,
		ChainHash: chainHash,
	}

	if err := m.Save(); err != nil {
		return "", err
	}

	_ = cache.WriteChain(chainHash, positions)

	return "wrote " + path, nil
}

func computeChecksum(content string) string {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	if !strings.HasSuffix(normalized, "\n") {
		normalized += "\n"
	}
	h := sha1.Sum([]byte(normalized))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
