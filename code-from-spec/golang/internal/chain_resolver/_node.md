# ROOT/golang/internal/chain_resolver

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

type Chain struct {
    Ancestors    []ChainItem
    Target       ChainItem
    Dependencies []ChainItem
    Code         []string
}

func ResolveChain(targetLogicalName string) (*Chain, error)
```

`ResolveChain` returns the chain separated into ancestors, target,
and dependencies. Returns an error if the chain cannot be built.

Each `ChainItem` has a single `FilePath` and an optional
`Qualifier`. When `Qualifier` is nil, the caller should use
the `# Public` section of the file. When `Qualifier` is
non-nil, the caller should use only the `## <qualifier>`
subsection within `# Public`.

### Error handling

- If `logicalnames.PathFromLogicalName` returns false for any logical name →
  return error: `"cannot resolve logical name: <name>"`.
- If `ParseFrontmatter` fails → return error wrapping the
  underlying error.
