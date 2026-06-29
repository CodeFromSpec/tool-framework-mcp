---
depends_on:
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing/frontmatter
  - SPEC/golang/implementation/utils/logical_names
output: internal/chainresolver/chainresolver.go
---

# SPEC/golang/implementation/chain/resolver

Resolves the ordered list of positions that form the
chain for a given target logical name.

# Public

## Package

`package chainresolver`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/chainresolver"`

## Interface

```go
type ChainItem struct {
	UnqualifiedLogicalName string
	FilePath               oslayer.CfsPath
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

Implement the chain resolver as a Go package.

## Logic

### Step 1 — Resolve ancestors and target

Call LogicalNameParse(target_logical_name).
If it fails, propagate the error. Let `target_ln`
be the result.

Create the target ChainItem with
  unqualified_logical_name = target_ln.Name,
  file_path = CfsPath(target_ln.Path),
  qualifier = absent.

If target_ln.Parent is nil:
  Set ancestors to an empty list.
  Skip to step 2.

Otherwise:
  Initialize a name list containing
  target_logical_name.
  Set current_ln = target_ln.
  Loop:
    If current_ln.Parent is nil, stop the loop.
    Add *current_ln.Parent to the name list.
    Call LogicalNameParse(*current_ln.Parent).
    If it fails, propagate the error.
    Set current_ln = the result.

  Sort the name list alphabetically. This produces
  root-first order.

  For each name in the sorted name list:
    Call LogicalNameParse(name).
    If it fails, propagate the error. Let `ln` be
    the result.
    Create a ChainItem with
      unqualified_logical_name = ln.Name,
      file_path = CfsPath(ln.Path),
      qualifier = absent.

  The last item in the sorted list becomes the target.
  All preceding items form the ancestors list.

### Step 2 — Resolve dependencies

Call FrontmatterParse(CfsPath(target_ln.Path)).
If it fails, raise ErrUnreadableFrontmatter.

Initialize an empty dependency list.

For each entry in frontmatter.depends_on:

  Call LogicalNameParse(entry).
  If it fails, raise ErrUnresolvableArtifact
  (wrapping the original error). Let `ln` be
  the result.

  If ln.Type is NodeTypeSpec:
    Create ChainItem with
      unqualified_logical_name = ln.Name,
      file_path = CfsPath(ln.Path),
      qualifier = if ln.Qualifier is not nil then
        *ln.Qualifier, else absent.
    Add to dependency list.

  Else if ln.Type is NodeTypeArtifact:
    Create ChainItem with
      unqualified_logical_name = ln.Name,
      file_path = CfsPath(ln.Path),
      qualifier = absent.
    Add to dependency list.

  Else if ln.Type is NodeTypeExternal:
    Create ChainItem with
      unqualified_logical_name = ln.Name,
      file_path = CfsPath(ln.Path),
      qualifier = absent.
    Add to dependency list.

  Else:
    raise ErrUnresolvableArtifact.

Sort the dependency list alphabetically by
unqualified_logical_name, then by qualifier (absent
sorts before present), in a single pass.

### Step 3 — Deduplicate dependencies

Initialize an empty deduplicated dependency list.

For each item in the sorted dependency list:

  If item has a SPEC-prefixed unqualified_logical_name:
    Check if an entry with the same
    unqualified_logical_name and the same qualifier
    already exists in the deduplicated list.
    If yes, skip (duplicate).
    Also check if an entry with the same
    unqualified_logical_name and no qualifier already
    exists. If yes, skip (full section covers every
    subsection).
    Otherwise, add to deduplicated list.

  Else if item has an ARTIFACT-prefixed
  unqualified_logical_name:
    Check if an entry with the same
    unqualified_logical_name already exists.
    If yes, skip. Otherwise, add.

  Else if item has an EXTERNAL-prefixed
  unqualified_logical_name:
    Check if an entry with the same
    unqualified_logical_name already exists.
    If yes, skip. Otherwise, add.

Replace the dependency list with the deduplicated list.

### Step 4 — Resolve input

If frontmatter.input is empty:
  Set the Chain's input field to absent.
Else:
  Call LogicalNameParse(frontmatter.input).
  If it fails, propagate the error. Let `input_ln`
  be the result.

  Create ChainItem with
    unqualified_logical_name = input_ln.Name,
    file_path = CfsPath(input_ln.Path),
    qualifier = if input_ln.Qualifier is not nil then
      *input_ln.Qualifier, else absent.
  Set the Chain's input field to that ChainItem.

Return Chain with ancestors, dependencies, target,
input.

## Go-specific guidance

- Use the `logicalnames` package for
  `LogicalNameParse`, `LogicalName`, `NodeTypeSpec`,
  `NodeTypeArtifact`, `NodeTypeExternal`.
- Use the `frontmatter` package for `FrontmatterParse`.
- Use the `oslayer` package for `CfsPath`.
- The package name should be `chainresolver`.
- `ChainItem` and `Chain` are exported structs in this
  package.
- Convert `ln.Path` to `CfsPath(ln.Path)` when
  assigning to ChainItem.FilePath.
