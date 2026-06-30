# Cache storage

## Directory layout

```
code-from-spec/.cache/
├── .content/    ← position content, keyed by content hash
└── .chains/     ← chain structure, keyed by chain hash
```

## File naming convention

Files in both stores are named with a dot prefix, a 27-character base64url hash, and no extension: `.<hash>`.

## Content file format

Each file in `.cache/.content/` stores the processed text of one chain position:

- For unqualified `SPEC/` references: the concatenation of all `##` subsections of `# Public`.
- For qualified `SPEC/x(y)` references: the single `## y` subsection.
- For `ARTIFACT/` and `EXTERNAL/` references: the full file content.

Identical content produces the same hash and is stored once regardless of how many chains reference it.

## Chain file format

Each file in `.cache/.chains/` contains the ordered list of positions that produced the chain hash. Each line has a label and a content hash separated by a colon and space:

```
SPEC/payments: d4e5f6g7h8i9j0k1l2m3n4o5p6q
SPEC/payments/fees: g7h8i9j0k1l2m3n4o5p6q7r8s
SPEC/architecture/backend/config(interface): a3b4c5d6e7f8g9h0i1j2k3l4m5n
AGENT[SPEC/payments/fees/calculation]: p6q7r8s9t0u1v2w3x4y5z6a
INPUT[ARTIFACT/functional/calc]: s9t0u1v2w3x4y5z6a7b8c9d
```

Labels for ancestors, dependencies, and the target node's `# Public` use the logical name directly. The target node's `# Agent` section is wrapped as `AGENT[...]` and the input as `INPUT[...]`. The content hash points to the corresponding file in `.cache/.content/`.

## Write-once semantics and atomic writes

Cache files are write-once — once created, they are never modified. Writes must be atomic: write to a temporary file, then rename. A cache file either exists completely or does not exist.

## Concurrency

Concurrent writes of the same hash produce identical content; one rename wins and the result is correct either way. Reads always see a complete file or no file. Concurrent deletes of the same unreferenced file are idempotent.
