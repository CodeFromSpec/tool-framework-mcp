---
depends_on:
  - ROOT/functional/logic/utils/logical_names
  - ROOT/functional/logic/parsing/frontmatter
  - ROOT/functional/logic/os/path_utils(interface)
output: code-from-spec/functional/logic/chain/resolver/output.md
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
  unqualified_logical_name: string
  file_path: pathutils.PathCfs
  qualifier: optional string

record Chain
  ancestors: list of ChainItem
  dependencies: list of ChainItem
  target: ChainItem
  input: optional ChainItem
```

`ChainItem` fields:

- `unqualified_logical_name` — the logical name without
  the parenthetical qualifier. For a `depends_on` entry
  like `SPEC/a/b(interface)`, this field holds
  `SPEC/a/b`. For `ARTIFACT/` and `EXTERNAL/`
  references (which never have qualifiers), this is the
  full logical name as-is.
- `file_path` — the resolved file path as a `PathCfs`.
  For `SPEC/` references, this is the `_node.md` path.
  For `ARTIFACT/` references, this is the artifact
  output path. For `EXTERNAL/` references, this is the
  project-root-relative path.
- `qualifier` — the parenthetical qualifier, if present
  in the original `depends_on` entry. Absent for
  references without a qualifier and for all `ARTIFACT/`
  and `EXTERNAL/` references.

```

function ChainResolve(target_logical_name: string) -> Chain
  errors:
    - UnreadableFrontmatter: a node's frontmatter
      cannot be parsed.
    - UnresolvableArtifact: an ARTIFACT/ reference
      cannot be resolved.
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
2. **Dependencies** — all entries from the target's
   `depends_on` (`SPEC/`, `ARTIFACT/`, `EXTERNAL/`),
   sorted alphabetically by logical name in a single
   pass. Each entry has its resolved file path and an
   optional qualifier.
3. **Target** — the target node itself.
4. **Input** — the target's `input`, if present.

# Agent

## Behavior

### Step 1 — Resolve ancestors and target

If the target logical name is `"SPEC"`, create a single
`ChainItem` for it (resolve file path using
`LogicalNameToPath`, qualifier absent). Ancestors list
is empty. Skip to Step 2.

Otherwise, add the target logical name to the list.
Then walk upward using `LogicalNameGetParent` repeatedly,
adding each parent to the list until reaching `SPEC`
(inclusive). If `LogicalNameGetParent` fails, propagate
the error.

Sort the collected list alphabetically by logical name.
This produces root-first order (e.g. `SPEC`, `SPEC/a`,
`SPEC/a/b`).

For each name, resolve the file path using
`LogicalNameToPath`. If it fails, propagate the error.
Create a `ChainItem` with qualifier absent.

The last item in the sorted list is the target; the
remaining items form the ancestors list.

### Step 2 — Resolve dependencies

Read the target node's frontmatter using
`FrontmatterParse` (pass the target's file path). If
parsing fails, raise "unreadable frontmatter".

For each entry in `frontmatter.depends_on`, classify
it using `LogicalNameIsSpec`, `LogicalNameIsArtifact`,
and `LogicalNameIsExternal`.

**`SPEC/` references** (detected by `LogicalNameIsSpec`):
1. Extract the qualifier using `LogicalNameGetQualifier`
   (absent if none). Strip the qualifier using
   `LogicalNameStripQualifier` to get the bare logical
   name.
2. Resolve the bare logical name to a file path using
   `LogicalNameToPath`. If it fails, propagate the error.
3. Create a `ChainItem` with the bare logical name as
   `unqualified_logical_name`, the resolved file path,
   and the qualifier (if any).

**`ARTIFACT/` references:**
1. Derive the generating node's logical name using
   `LogicalNameGetArtifactGenerator`. If it fails,
   propagate the error.
2. Resolve the generating node's logical name to a file
   path using `LogicalNameToPath`.
3. Read the generating node's frontmatter using
   `FrontmatterParse`. If parsing fails, raise
   "unreadable frontmatter".
4. If `frontmatter.output` is empty, raise
   "unresolvable artifact" — the generating node
   declares no output.
5. The output path is `frontmatter.output`. Do not
   verify existence — the artifact may not have been
   generated yet.
6. Create a `ChainItem` with the `ARTIFACT/` logical
   name as `unqualified_logical_name`, the output path
   as `PathCfs`, and qualifier absent.

**`EXTERNAL/` references:**
1. Convert to a file path using
   `LogicalNameExternalToPath`. If it fails, propagate
   the error.
2. Create a `ChainItem` with the `EXTERNAL/` logical
   name as `unqualified_logical_name`, the path as
   `PathCfs`, and qualifier absent.

If none of the three prefixes match, raise
"unresolvable artifact".

Sort all dependencies alphabetically by logical name,
then by qualifier (absent sorts before present). This
is a single sort pass across all reference types.

### Step 3 — Deduplicate dependencies

Remove duplicate entries from the dependencies list.

Determine the entry type using `LogicalNameIsArtifact`,
`LogicalNameIsSpec`, and `LogicalNameIsExternal`.

For `SPEC/` entries:
- Two entries are duplicates when they have the same
  logical name (after normalization) and the same
  qualifier.
- When an entry exists with a given logical name and no
  qualifier (meaning the full `# Public` section), any
  other entry with the same logical name and a qualifier
  is redundant — the full section already includes
  every subsection. Remove the redundant entry.

For `ARTIFACT/` entries: two entries are duplicates when
they have the same logical name.

For `EXTERNAL/` entries: two entries are duplicates when
they have the same logical name.

Keep the first occurrence when removing duplicates.

### Step 4 — Resolve input

If the target's frontmatter has a non-empty `input`
field, classify it using `LogicalNameIsArtifact` and
`LogicalNameIsExternal`.

**`ARTIFACT/` input:**
1. Derive the generating node's logical name using
   `LogicalNameGetArtifactGenerator`. If it fails,
   propagate the error.
2. Resolve the generating node's logical name to a file
   path using `LogicalNameToPath`.
3. Read the generating node's frontmatter using
   `FrontmatterParse`. If parsing fails, raise
   "unreadable frontmatter".
4. If `frontmatter.output` is empty, raise
   "unresolvable artifact" — the generating node
   declares no output.
5. The output path is `frontmatter.output`. Do not
   verify existence.
6. Create a `ChainItem` with the `ARTIFACT/` logical
   name as `unqualified_logical_name`, the output path
   as `PathCfs`, and qualifier absent.

**`EXTERNAL/` input:**
1. Convert to a file path using
   `LogicalNameExternalToPath`. If it fails, propagate
   the error.
2. Create a `ChainItem` with the `EXTERNAL/` logical
   name as `unqualified_logical_name`, the path as
   `PathCfs`, and qualifier absent.

If `input` is empty, the `input` field in the returned
`Chain` is absent.

## Contracts

- The chain is fully resolved — all file paths are
  derived from logical names and frontmatter. Existence
  is not verified; the caller handles missing files.
- File paths are `PathCfs` values (forward slashes).
- No duplicate entries in the dependencies list.
- Ancestors are in root-first order.
- Dependencies are sorted alphabetically by logical
  name, then by qualifier.
