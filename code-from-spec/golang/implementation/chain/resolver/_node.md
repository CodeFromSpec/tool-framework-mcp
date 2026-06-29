---
depends_on:
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing(interface)
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
type Chain struct {
	Ancestors    []parsing.CfsReference
	Dependencies []parsing.CfsReference
	Target       parsing.CfsReference
	Input        *parsing.CfsReference // nil if absent
}

func ChainResolve(targetLogicalName string) (Chain, error)
```

### Chain assembly order

1. **Ancestors** — from root down to (but not including)
   the target node.
2. **Dependencies** — all entries from the target's
   `depends_on`, sorted alphabetically by logical name.
3. **Target** — the target node itself.
4. **Input** — the target's `input`, if present.

### Errors

- `ErrUnreadableFrontmatter`: the node cannot be parsed.
- `ErrUnresolvableArtifact`: an ARTIFACT/ reference
  cannot be resolved.
- Propagated errors from `parsing` package.

# Agent

Implement the chain resolver as a Go package.

## Logic

### Step 1 — Resolve ancestors and target

Call parsing.CfsReferenceFromName(target_logical_name).
If it fails, propagate the error. Let `target_ref`
be the result.

If target_ref.ParentName is nil:
  Set ancestors to an empty list.
  Skip to step 2.

Otherwise:
  Initialize a ref list containing *target_ref.
  Set current_ref = target_ref.
  Loop:
    If current_ref.ParentName is nil, stop the loop.
    Call parsing.CfsReferenceFromName(
    *current_ref.ParentName).
    If it fails, propagate the error.
    Set current_ref = the result.
    Add *current_ref to the ref list.

  Sort the ref list alphabetically by LogicalName.
  This produces root-first order.

  The last item in the sorted list is the target.
  All preceding items form the ancestors list.

### Step 2 — Resolve dependencies

Call parsing.ParseNode(target_logical_name).
If it fails, raise ErrUnreadableFrontmatter.
Let `node` be the result. Let `fm` =
node.Frontmatter.

Initialize an empty dependency list.

For each entry in fm.DependsOn:

  Call parsing.CfsReferenceFromName(entry).
  If it fails, raise ErrUnresolvableArtifact
  (wrapping the original error). Let `ref` be
  the result.

  Add *ref to dependency list.

Sort the dependency list alphabetically by
LogicalName, then by Qualifier (nil sorts before
non-nil), in a single pass.

### Step 3 — Deduplicate dependencies

Initialize an empty deduplicated dependency list.

For each item in the sorted dependency list:

  If item.LogicalName starts with "SPEC/":
    Check if an entry with the same LogicalName and
    the same Qualifier already exists in the
    deduplicated list. If yes, skip (duplicate).
    Also check if an entry with the same LogicalName
    and nil Qualifier already exists. If yes, skip
    (full section covers every subsection).
    Otherwise, add to deduplicated list.

  Else if item.LogicalName starts with "ARTIFACT/":
    Check if an entry with the same LogicalName
    already exists. If yes, skip. Otherwise, add.

  Else if item.LogicalName starts with "EXTERNAL/":
    Check if an entry with the same LogicalName
    already exists. If yes, skip. Otherwise, add.

Replace the dependency list with the deduplicated
list.

### Step 4 — Resolve input

If fm.Input is nil:
  Set the Chain's Input field to nil.
Else:
  Call parsing.CfsReferenceFromName(*fm.Input).
  If it fails, propagate the error. Let `input_ref`
  be the result.
  Set the Chain's Input field to input_ref.

Return Chain with Ancestors, Dependencies, Target,
Input.

## Go-specific guidance

- Use the `parsing` package for
  `CfsReferenceFromName`, `CfsReference`,
  `ParseNode`.
- The package name should be `chainresolver`.
- `Chain` is the only exported struct in this package.
  It holds `parsing.CfsReference` values directly —
  no intermediate types.
