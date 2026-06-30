# Cache

The cache stores spec chain content from previous
generations so that the tooling can show the subagent
what changed.

The cache is best-effort infrastructure
— without it, the framework works, but the subagent
cannot see what the spec looked like in the previous
generation.

This document assumes familiarity with
CODE_FROM_SPEC.md and MANIFEST.md.

---

## Location

The cache lives at `code-from-spec/.cache/` and is
gitignored. It contains two subdirectories:

```
code-from-spec/.cache/
├── .content/    ← position content, keyed by content hash
└── .chains/     ← chain structure, keyed by chain hash
```

---

## Content store

`.cache/.content/` stores the processed content of each
chain position — the exact content that participates in
the chain hash.

Each file is named `.<content-hash>` (dot-prefixed,
27-character base64url hash, no extension). The file
content is the processed text of the position: for
unqualified `SPEC/` references, the concatenation of
all `##` subsections of `# Public`; for qualified
`SPEC/x(y)` references, the single `## y` subsection;
for `ARTIFACT/` and `EXTERNAL/` references, the full
file content.

Deduplication is automatic — identical content produces
the same hash and is stored once regardless of how many
chains reference it.

---

## Chain store

`.cache/.chains/` stores the structure of each chain at
the time it was computed. Each file is named
`.<chain-hash>` (dot-prefixed, 27-character base64url
hash, no extension).

The file contains the ordered list of positions that
produced the chain hash. Each line has a label and a
content hash:

```
SPEC/payments: d4e5f6g7h8i9j0k1l2m3n4o5p6q
SPEC/payments/fees: g7h8i9j0k1l2m3n4o5p6q7r8s
SPEC/architecture/backend/config(interface): a3b4c5d6e7f8g9h0i1j2k3l4m5n
SPEC/integrations/database: j0k1l2m3n4o5p6q7r8s9t0u
SPEC/payments/fees/calculation: m3n4o5p6q7r8s9t0u1v2w3x
AGENT[SPEC/payments/fees/calculation]: p6q7r8s9t0u1v2w3x4y5z6a
INPUT[ARTIFACT/functional/calc]: s9t0u1v2w3x4y5z6a7b8c9d
```

The label identifies the position. Labels for
ancestors, dependencies, and the target node's
`# Public` use the logical name directly. The target
node's `# Agent` section, if present, is wrapped as
`AGENT[...]` and the input, if present, as
`INPUT[...]`. The content hash points to the
corresponding file in `.cache/.content/`.

---

## Cache population

The cache is populated as a side effect of normal
operations. Whenever the tooling processes a spec
chain, it writes the content of each position to
`.cache/.content/` and the chain structure to
`.cache/.chains/`.

Over the course of a session, the cache self-completes.

The tooling may implement a `reconstruct_cache`
operation that reads the manifest and populates the
cache from the current state of all files. It is
idempotent — only fills gaps, skipping content and
chain files that already exist.

---

## Pruning

The tooling may implement a `prune_cache` operation
that removes unreferenced files from the cache:

- Content files in `.cache/.content/` whose hash is
  not referenced by any chain file in `.cache/.chains/`.
- Chain files in `.cache/.chains/` whose hash is not
  referenced by any manifest entry.

---

## Concurrency

Cache files are write-once — once created, they are
never modified. This eliminates the need for file
locking on the cache.

- **Writes** must be atomic (write to a temporary file,
  then rename). A cache file either exists completely
  or does not exist.
- **Concurrent writes** of the same hash produce
  identical content. One rename wins; the result is
  correct either way.
- **Reads** always see a complete file or no file.
- **Pruning** deletes only unreferenced files.
  Concurrent deletes of the same file are idempotent.

---

## Cache usage

When assembling the spec chain for a stale artifact,
the tooling checks whether cache data is available
for the previous generation:

1. Read the old chain hash from the manifest entry.
2. Look up `.cache/.chains/<old-chain-hash>`. If the
   file does not exist, cache is not available.
3. For each position in the old chain structure,
   verify that `.cache/.content/<content-hash>` exists.
   If any content file is missing, cache is not
   available.

When cache is not available, the spec chain is
assembled without `<previous_*>` sections or
`disposition` attributes.

When cache is available, the tooling compares old and
current chains to compute a `disposition` for each
position:

1. Read the old chain structure from
   `.cache/.chains/<old-chain-hash>`.
2. Compute the current chain positions (labels and
   content hashes).
3. For constraints entries, compare by label:
   - **`unchanged`** — same label, same content hash.
   - **`changed`** — same label, different content
     hash. Old content is read from
     `.cache/.content/`.
   - **`removed`** — label exists in old but not in
     current. Old content is read from
     `.cache/.content/`.
   - **`added`** — label exists in current but not
     in old.
4. For instructions and input, compare by content
   hash only (the name of the reference does not
   matter):
   - **`unchanged`** — same content hash.
   - **`changed`** — different content hash. Old
     content is read from `.cache/.content/`.
   - **`removed`** — existed before, absent now.
     Old content is read from `.cache/.content/`.
   - **`added`** — absent before, exists now.

The disposition is delivered in the spec chain XML:
current sections (`<constraints>`, `<instructions>`,
`<input>`) carry disposition on every entry. Previous
sections (`<previous_constraints>`,
`<previous_instructions>`, `<previous_input>`) contain
only `changed` and `removed` entries with their old
content. See CHAIN_ASSEMBLY.md for the full format.

---

## Resources

| Document | Description |
|---|---|
| [CODE_FROM_SPEC.md](https://github.com/CodeFromSpec/framework/blob/main/CODE_FROM_SPEC.md) | Full methodology specification |
| [CHAIN_HASH.md](https://github.com/CodeFromSpec/framework/blob/main/rules/CHAIN_HASH.md) | Chain hash algorithm for staleness detection |
| [CHAIN_ASSEMBLY.md](https://github.com/CodeFromSpec/framework/blob/main/rules/CHAIN_ASSEMBLY.md) | Chain format, assembly order, and delivery |
| [MANIFEST.md](https://github.com/CodeFromSpec/framework/blob/main/rules/MANIFEST.md) | Manifest format and artifact status |
