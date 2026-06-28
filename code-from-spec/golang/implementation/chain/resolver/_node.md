---
depends_on:
  - SPEC/golang/implementation/os/file/impl
  - SPEC/golang/implementation/os/path_utils
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

Implement the chain resolver as a Go package.

## Logic

### Step 1 — Resolve ancestors and target

If target_logical_name is "SPEC":
  Resolve the file path by calling
  LogicalNameToPath(target_logical_name).
  If it fails, propagate the error.
  Create a ChainItem with
    unqualified_logical_name = target_logical_name,
    file_path = resolved path, qualifier = absent.
  Set ancestors to an empty list.
  Set target to that ChainItem.
  Skip to step 2.

Otherwise:
  Initialize a name list containing
  target_logical_name.
  Set current_name = target_logical_name.
  Loop:
    Call LogicalNameGetParent(current_name).
    If it fails, propagate the error.
    Add the returned parent name to the name list.
    Set current_name = parent name.
    If current_name is "SPEC", stop the loop.

  Sort the name list alphabetically. This produces
  root-first order.

  For each name in the sorted name list:
    Call LogicalNameToPath(name).
    If it fails, propagate the error.
    Create a ChainItem with
      unqualified_logical_name = name,
      file_path = resolved path, qualifier = absent.

  The last item in the sorted list becomes the target.
  All preceding items form the ancestors list.

### Step 2 — Resolve dependencies

Call FrontmatterParse(target.file_path).
If it fails, raise error "UnreadableFrontmatter".

Initialize an empty dependency list.

For each entry in frontmatter.depends_on:

  If LogicalNameIsSpec(entry) is true:
    Call LogicalNameGetQualifier(entry) to get the
    qualifier (absent if none).
    Call LogicalNameStripQualifier(entry) to get the
    bare logical name.
    Call LogicalNameToPath(bare logical name).
    If it fails, propagate the error.
    Create ChainItem with
      unqualified_logical_name = bare logical name,
      file_path = resolved path,
      qualifier = extracted qualifier (absent if none).
    Add to dependency list.

  Else if LogicalNameIsArtifact(entry) is true:
    Call LogicalNameGetArtifactGenerator(entry) to get
    the generating node's logical name.
    If it fails, propagate the error.
    Call LogicalNameToPath(generating node's logical
    name). If it fails, propagate the error.
    Call FrontmatterParse(generating node's file path).
    If it fails, raise error "UnreadableFrontmatter".
    If generating node's frontmatter.output is empty,
    raise error "UnresolvableArtifact".
    Create ChainItem with
      unqualified_logical_name = entry (as-is),
      file_path = frontmatter.output as PathCfs,
      qualifier = absent.
    Add to dependency list.

  Else if LogicalNameIsExternal(entry) is true:
    Call LogicalNameExternalToPath(entry).
    If it fails, propagate the error.
    Create ChainItem with
      unqualified_logical_name = entry (as-is),
      file_path = resolved path,
      qualifier = absent.
    Add to dependency list.

  Else:
    raise error "UnresolvableArtifact".

Sort the dependency list alphabetically by
unqualified_logical_name, then by qualifier (absent
sorts before present), in a single pass.

### Step 3 — Deduplicate dependencies

Initialize an empty deduplicated dependency list.

For each entry in the sorted dependency list:

  If LogicalNameIsSpec(entry.unqualified_logical_name):
    Check if an entry with the same
    unqualified_logical_name and the same qualifier
    already exists in the deduplicated list.
    If yes, skip (duplicate).
    Also check if an entry with the same
    unqualified_logical_name and no qualifier already
    exists. If yes, skip (full section covers every
    subsection).
    Otherwise, add to deduplicated list.

  Else if LogicalNameIsArtifact(entry.unqualified_logical_name):
    Check if an entry with the same
    unqualified_logical_name already exists.
    If yes, skip. Otherwise, add.

  Else if LogicalNameIsExternal(entry.unqualified_logical_name):
    Check if an entry with the same
    unqualified_logical_name already exists.
    If yes, skip. Otherwise, add.

Replace the dependency list with the deduplicated list.

### Step 4 — Resolve input

If frontmatter.input is empty:
  Set the Chain's input field to absent.
Else:
  Set input_entry = frontmatter.input.

  If LogicalNameIsArtifact(input_entry) is true:
    Call LogicalNameGetArtifactGenerator(input_entry).
    If it fails, propagate the error.
    Call LogicalNameToPath(generating node's logical
    name). If it fails, propagate the error.
    Call FrontmatterParse(generating node's file path).
    If it fails, raise error "UnreadableFrontmatter".
    If frontmatter.output is empty, raise error
    "UnresolvableArtifact".
    Create ChainItem with
      unqualified_logical_name = input_entry (as-is),
      file_path = frontmatter.output as PathCfs,
      qualifier = absent.
    Set the Chain's input field to that ChainItem.

  Else if LogicalNameIsExternal(input_entry) is true:
    Call LogicalNameExternalToPath(input_entry).
    If it fails, propagate the error.
    Create ChainItem with
      unqualified_logical_name = input_entry (as-is),
      file_path = resolved path,
      qualifier = absent.
    Set the Chain's input field to that ChainItem.

Return Chain with ancestors, dependencies, target,
input.

## Go-specific guidance

- Use the `logicalnames` package for
  `LogicalNameGetParent`, `LogicalNameToPath`,
  `LogicalNameGetQualifier`, `LogicalNameStripQualifier`,
  `LogicalNameGetArtifactGenerator`,
  `LogicalNameIsArtifact`.
- Use the `frontmatter` package for `FrontmatterParse`
  and the `Frontmatter`, `FrontmatterExternal`,
  `FrontmatterOutput` records.
- Use the `pathutils` package for `PathCfs`.
- The package name should be `chainresolver`.
- `ChainItem` and `Chain` are exported structs in this
  package.
- `FrontmatterExternal` in the `Chain.External` field
  uses the type from the `frontmatter` package directly.
