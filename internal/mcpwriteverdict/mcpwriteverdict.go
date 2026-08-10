package mcpwriteverdict

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
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/subagenttoken"
)

var (
	ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")
	ErrNoOutput              = errors.New("no output")
	ErrNotAVerdict           = errors.New("not a verdict")
)

func MCPWriteVerdict(token string, passed bool, content string) (string, error) {
	logicalName, err := subagenttoken.SubagentTokenValidate(token)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	node, err := parsing.ParseNode(logicalName)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrUnreadableFrontmatter, err)
	}

	if node.Frontmatter == nil || node.Frontmatter.Type == nil {
		return "", ErrNoOutput
	}
	if *node.Frontmatter.Type != "verdict" {
		return "", ErrNotAVerdict
	}

	resolvedOutput := parsing.ResolvedOutput(node)
	if resolvedOutput == nil {
		return "", ErrNoOutput
	}
	path := *resolvedOutput

	if err := oslayer.ValidateStringIsCfsPath(path); err != nil {
		return "", fmt.Errorf("%w", err)
	}

	handle, err := oslayer.OpenFile(oslayer.CfsPath(path), "overwrite", 30000)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	if err := handle.Write(content); err != nil {
		handle.Close()
		return "", fmt.Errorf("%w", err)
	}

	handle.Close()

	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	if !strings.HasSuffix(normalized, "\n") {
		normalized += "\n"
	}
	sum := sha1.Sum([]byte(normalized))
	checksum := base64.RawURLEncoding.EncodeToString(sum[:])

	chain, err := chainresolver.ChainResolve(logicalName)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	chainHash, _, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	m, err := manifest.OpenManifest(false)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	defer func() { _ = m.Discard() }()

	verdictName := "VERDICT/" + strings.TrimPrefix(logicalName, "SPEC/")
	resultValue := "fail"
	if passed {
		resultValue = "pass"
	}
	m.Entries[verdictName] = manifest.ManifestEntry{
		Path:      path,
		Checksum:  checksum,
		ChainHash: chainHash,
		Result:    resultValue,
	}

	if err := m.Save(); err != nil {
		return "", fmt.Errorf("%w", err)
	}

	return "wrote " + path, nil
}
