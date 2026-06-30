// code-from-spec: SPEC/golang/implementation/mcp_tools/write_file@rb68Dz0BYCUIJSP9F_hpCB43jcI
package mcpwritefile

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

var (
	ErrNotASpecReference     = errors.New("logical name is not a SPEC/ reference")
	ErrQualifierNotAllowed   = errors.New("logical name must not contain a qualifier")
	ErrUnreadableFrontmatter = errors.New("node frontmatter cannot be parsed")
	ErrNoOutput              = errors.New("node has no output field")
	ErrPathNotInOutput       = errors.New("path is not declared in the node's output")
)

func MCPWriteFile(logicalName, path, content string) (string, error) {
	if !strings.HasPrefix(logicalName, "SPEC/") {
		return "", ErrNotASpecReference
	}

	if strings.Contains(logicalName, "(") {
		return "", ErrQualifierNotAllowed
	}

	node, err := parsing.ParseNode(logicalName)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrUnreadableFrontmatter, err)
	}

	if node.Frontmatter == nil || node.Frontmatter.Output == nil {
		return "", ErrNoOutput
	}

	if err := oslayer.ValidateStringIsCfsPath(path); err != nil {
		return "", err
	}

	if path != *node.Frontmatter.Output {
		return "", ErrPathNotInOutput
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

	chainHash, err := chainhash.ChainHashCompute(chain)
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
