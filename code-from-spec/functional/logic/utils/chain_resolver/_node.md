---
depends_on:
  - ROOT/functional/logic/utils/logical_names
  - ROOT/functional/logic/parsing/frontmatter
outputs:
  - id: chain_resolver
    path: code-from-spec/functional/logic/utils/chain_resolver/output.md
---

# ROOT/functional/logic/utils/chain_resolver

Resolves the ordered list of files that form the chain
for a given target logical name.

Review status: pending

# Public

## Interface

```
record ChainItem
  logical_name: string
  file_path: string
  qualifier: optional string

record ExternalItem
  path: string

record Chain
  ancestors: list of ChainItem
  target: ChainItem
  dependencies: list of ChainItem
  external: list of ExternalItem
  input: string

function ResolveChain(target_logical_name: string) -> Chain
  errors:
    - cannot resolve logical name: a logical name cannot
      be resolved to a file path.
    - unreadable frontmatter: the target's frontmatter
      cannot be parsed.
```

`ResolveChain` returns the chain for a target logical
name — the ordered list of files that a downstream tool
needs to assemble context for artifact generation.

### Chain assembly order

1. **Ancestors** — from root down to (but not including)
   the target node.
2. **Dependencies** — nodes listed in the target's
   `depends_on`. Each dependency contributes its file path
   and an optional qualifier targeting a specific
   subsection.
3. **External** — files listed in the target's `external`.
4. **Target** — the target node itself.
5. **Input** — the target's `input` field, if present.

### Qualifier semantics

When qualifier is absent, the caller uses the full
`# Public` section. When present, the caller uses only
the `## <qualifier>` subsection within `# Public`.

# Agent

## Behavior

### Step 1 — Ancestors and target

Starting from the target logical name, walk upward using
`GetParent` repeatedly, collecting each logical name.
Sort the collected list alphabetically by logical name.

For each name, resolve the file path using `ResolvePath`.
Create a ChainItem with qualifier absent.

The last item in the sorted list is the target; the
remaining items form the ancestors list.

### Step 2 — Dependencies

Read the target node's frontmatter using
`ParseFrontmatter`. For each entry in `depends_on`:

1. Resolve the file path using `ResolvePath`.
2. If the logical name has a qualifier (using
   `ExtractQualifier`), set qualifier to that value.
   Otherwise, leave qualifier absent.
3. Verify the file exists. If not, raise error
   "cannot resolve logical name: <name>".
4. Add to the dependencies list.

Sort dependencies alphabetically by file path, then by
qualifier (absent sorts before present).

### Step 3 — External

Extract the `external` entries from the target's
frontmatter. Each entry contributes its path.

### Step 4 — Input

Extract the `input` field from the target's frontmatter.
If present, record it in the chain.

### Step 5 — Normalize file paths

All file paths use forward slashes as separators,
regardless of the operating system.

### Step 6 — Deduplicate

Remove duplicate entries from ancestors and dependencies.
Two entries are duplicates when they have the same file
path and the same qualifier.

When an entry exists with a given file path and no
qualifier (meaning the full `# Public` section), any
other entry with the same file path and a qualifier is
redundant — the full section already includes every
subsection. Remove the redundant entry.

Keep the first occurrence when removing duplicates.

## Contracts

- The chain is fully resolved — all file paths are valid
  and point to existing files.
- File paths use forward slashes regardless of OS.
- No duplicate entries in the result.
