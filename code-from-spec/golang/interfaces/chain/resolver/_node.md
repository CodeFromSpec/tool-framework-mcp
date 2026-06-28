---
depends_on:
  - ARTIFACT/golang/interfaces/os/path_utils
  - ARTIFACT/golang/interfaces/parsing/frontmatter
output: code-from-spec/golang/interfaces/chain/resolver/output.md
---

# SPEC/golang/interfaces/chain/resolver

Resolves the ordered list of positions that form the
chain for a given target logical name.

# Public

## Package

`package chainresolver`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainresolver"`

## Interface

```go
type ChainItem struct {
	UnqualifiedLogicalName string
	FilePath               pathutils.PathCfs
	Qualifier              string // empty if absent
}

type Chain struct {
	Ancestors    []*ChainItem
	Dependencies []*ChainItem
	Target       *ChainItem
	Input        *ChainItem // nil if absent
}

func ChainResolve(targetLogicalName string) (*Chain, error)
```

### Chain assembly order

1. **Ancestors** — from root down to (but not including)
   the target node.
2. **Dependencies** — all entries from the target's
   `depends_on`, sorted alphabetically by logical name.
3. **Target** — the target node itself.
4. **Input** — the target's `input`, if present.

### Errors

- `ErrUnreadableFrontmatter`: a node's frontmatter
  cannot be parsed.
- `ErrUnresolvableArtifact`: an ARTIFACT/ reference
  cannot be resolved.
- Propagated errors from `logicalnames`, `frontmatter`
  packages.

# Agent

Generate an interface specification document listing
the package, import path, struct definitions, and
function signatures.
