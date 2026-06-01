---
depends_on:
  - ROOT/functional/logic/utils/logical_names
  - ROOT/functional/logic/parsing/frontmatter
  - ROOT/functional/logic/os/path_utils(interface)
outputs:
  - id: chain_resolver
    path: code-from-spec/functional/logic/chain/resolver/output.md
---

# ROOT/functional/logic/chain/resolver

Resolves the ordered list of positions that form the
chain for a given target logical name.

# Public

## Namespace

    namespace: chainresolver

## Interface

```
record ChainItem
  logical_name: string
  file_path: pathutils.PathCfs
  qualifier: optional string

record Chain
  ancestors: list of ChainItem
  dependencies: list of ChainItem
  external: list of frontmatter.FrontmatterExternal
  target: ChainItem
  input: optional ChainItem

function ChainResolve(target_logical_name: string) -> Chain
  errors:
    - UnreadableFrontmatter: a node's frontmatter
      cannot be parsed.
    - UnresolvableArtifact: an ARTIFACT/ reference's
      output id does not match any declared output.
    - (LogicalNames.*): propagated from
      LogicalNameToPath, LogicalNameGetParent.
    - (Frontmatter.*): propagated from
      FrontmatterParse.
```

`ChainResolve` returns the chain for a target logical
name — the ordered list of positions that a downstream
tool needs to assemble context for artifact generation
or to compute the chain hash.

### Chain assembly order

1. **Ancestors** — from root down to (but not including)
   the target node.
2. **Dependencies** — entries from the target's
   `depends_on`, sorted alphabetically by file path
   then by qualifier, each with its resolved file path
   and an optional qualifier.
3. **External** — files from the target's `external`,
   sorted alphabetically by path, including fragment
   declarations when present.
4. **Target** — the target node itself.
5. **Input** — the target's `input` artifact, if present.

# Agent

## Behavior

### Step 1 — Resolve ancestors and target

If the target logical name is `"ROOT"`, create a single
`ChainItem` for it (resolve file path using
`LogicalNameToPath`, qualifier absent). Ancestors list
is empty. Skip to Step 2.

Otherwise, add the target logical name to the list.
Then walk upward using `LogicalNameGetParent` repeatedly,
adding each parent to the list until reaching `ROOT`
(inclusive). If `LogicalNameGetParent` fails, propagate
the error.

Sort the collected list alphabetically by logical name.
This produces root-first order (e.g. `ROOT`, `ROOT/a`,
`ROOT/a/b`).

For each name, resolve the file path using
`LogicalNameToPath`. If it fails, propagate the error.
Create a `ChainItem` with qualifier absent.

The last item in the sorted list is the target; the
remaining items form the ancestors list.

### Step 2 — Resolve dependencies

Read the target node's frontmatter using
`FrontmatterParse` (pass the target's file path). If
parsing fails, raise "unreadable frontmatter".

For each entry in `frontmatter.depends_on`, determine
whether it starts with `ROOT/` or `ARTIFACT/`. If
neither, raise "unresolvable artifact".

**`ROOT/` references:**
1. Extract the qualifier using `LogicalNameGetQualifier`
   (absent if none). Strip the qualifier using
   `LogicalNameStripQualifier` to get the bare logical
   name.
2. Resolve the bare logical name to a file path using
   `LogicalNameToPath`. If it fails, propagate the error.
3. Create a `ChainItem` with the bare logical name, the
   resolved file path, and the qualifier (if any).

**`ARTIFACT/` references:**
1. Extract the qualifier (artifact id) using
   `LogicalNameGetQualifier`. If absent, raise
   "unresolvable artifact" — `ARTIFACT/` references
   must include an id.
2. Derive the generating node's logical name using
   `LogicalNameGetArtifactGenerator`. If it fails,
   propagate the error.
3. Resolve the generating node's logical name to a file
   path using `LogicalNameToPath`.
4. Read the generating node's frontmatter using
   `FrontmatterParse`. If parsing fails, raise
   "unreadable frontmatter".
5. Find the output entry whose `id` matches the
   qualifier. If no match, raise "unresolvable artifact".
   The output's `path` is the artifact file path. Do not
   verify existence — the artifact may not have been
   generated yet.
6. Create a `ChainItem` with the original `ARTIFACT/`
   logical name, the artifact file path as `PathCfs`,
   and the qualifier.

Sort dependencies alphabetically by file path value,
then by qualifier (absent sorts before present).

### Step 3 — Deduplicate dependencies

Remove duplicate entries from the dependencies list.

Determine whether each entry is `ROOT/` or `ARTIFACT/`
using `LogicalNameIsArtifact`.

For `ROOT/` entries:
- Two entries are duplicates when they have the same
  file path and the same qualifier.
- When an entry exists with a given file path and no
  qualifier (meaning the full `# Public` section), any
  other entry with the same file path and a qualifier
  is redundant — the full section already includes
  every subsection. Remove the redundant entry.

For `ARTIFACT/` entries: two entries are duplicates only
when they have the exact same logical name (including
qualifier). The qualifier is the artifact id and is
always present — there is no "without qualifier"
subsumption for artifacts.

Keep the first occurrence when removing duplicates.

### Step 4 — Collect external

Copy the `external` list from the target's frontmatter
into the chain. The list preserves the `FrontmatterExternal`
records as-is, including any `fragments` declarations.
Sort entries alphabetically by `path`. Fragments within
each entry retain their declaration order.

### Step 5 — Resolve input

If the target's frontmatter has a non-empty `input`
field (it is an `ARTIFACT/` reference):
1. Extract the qualifier (artifact id) using
   `LogicalNameGetQualifier`. If absent, raise
   "unresolvable artifact".
2. Derive the generating node's logical name using
   `LogicalNameGetArtifactGenerator`. If it fails,
   propagate the error.
3. Resolve the generating node's logical name to a file
   path using `LogicalNameToPath`.
4. Read the generating node's frontmatter using
   `FrontmatterParse`. If parsing fails, raise
   "unreadable frontmatter".
5. Find the output entry whose `id` matches the
   qualifier. If no match, raise "unresolvable artifact".
   The output's `path` is the artifact file path. Do not
   verify existence.
6. Create a `ChainItem` with the original `ARTIFACT/`
   logical name, the artifact file path as `PathCfs`,
   and the qualifier.

If `input` is empty, the `input` field in the returned
`Chain` is absent.

## Contracts

- The chain is fully resolved — all file paths are
  derived from logical names and frontmatter. Existence
  is not verified; the caller handles missing files.
- File paths are `PathCfs` values (forward slashes).
- No duplicate entries in the dependencies list.
- Ancestors are in root-first order.
- Dependencies are sorted by file path then qualifier.
- External entries are sorted by path.
