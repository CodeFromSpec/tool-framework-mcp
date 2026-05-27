# ROOT/golang/implementation/internal/chain_resolver

Resolves the ordered list of files that form the chain for a
given target logical name.

# Public

## Package

`package chainresolver`

## Interface

```go
type ChainItem struct {
    LogicalName string
    FilePath    string
    Qualifier   *string
}

type ExternalItem struct {
    Path string
}

type Chain struct {
    Ancestors    []ChainItem
    Target       ChainItem
    Dependencies []ChainItem
    External     []ExternalItem
    Input        string
}

func ResolveChain(targetLogicalName string) (*Chain, error)
```

`ResolveChain` returns the chain for a target logical name.
Returns an error if the chain cannot be built.

### Chain assembly order

The chain is assembled in this order:

1. **Ancestors** — from `ROOT` down to (but not including) the
   target node. Each ancestor contributes its `# Public` section.
2. **Dependencies** — nodes listed in the target's `depends_on`
   frontmatter field. Each dependency contributes its `# Public`
   section (or a specific `## <qualifier>` subsection if the
   logical name has a qualifier). `ARTIFACT/` dependencies are
   resolved via `logicalnames.ArtifactRefParts` to find the
   node path and artifact ID.
3. **External** — files listed in the target's `external`
   frontmatter field. Each entry contributes the file path.
4. **Target** — the target node itself, contributing its
   `# Public` section and `# Agent` section.
5. **Input** — the target's `input` frontmatter field, if
   present.

Each `ChainItem` has a single `FilePath` and an optional
`Qualifier`. When `Qualifier` is nil, the caller should use
the `# Public` section of the file. When `Qualifier` is
non-nil, the caller should use only the `## <qualifier>`
subsection within `# Public`.

### Error handling

- If `logicalnames.PathFromLogicalName` returns false for any
  ROOT/ logical name → return error:
  `"cannot resolve logical name: <name>"`.
- If `ParseFrontmatter` fails → return error wrapping the
  underlying error.
