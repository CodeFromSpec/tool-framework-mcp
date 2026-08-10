# Manifest

The manifest tracks the state of every generated
artifact and verdict.

This document assumes familiarity with
[CODE_FROM_SPEC.md](../CODE_FROM_SPEC.md).

---

## Location

The manifest is a single file at
`code-from-spec/.manifest`. It is committed to version
control.

If the manifest is lost, it can be regenerated, but the
record of which chain hash produced each artifact and
verdict is lost — and so is each verdict's `result`. All
are treated as stale and must be regenerated.

---

## Format

One line per artifact or verdict, terminated by a
newline (LF).
Entries are ordered alphabetically by logical name.

```
code-from-spec: v6
ARTIFACT/payments/fees/calculation;path:internal/fees/calculation.go;checksum:Kx9mP2vB7wY2tHsJ8dFak4Xz9pQ;chain:Jz3qR7nL5cW1gT4yK8mDfAx0vBe
VERDICT/review/fees;path:code-from-spec/review/fees/verdict.md;checksum:Ux1mP2vB7wY2tHsJ8dFak4Xz9pQ;chain:Wz3qR7nL5cW1gT4yK8mDfAx0vBe;result:pass
```

The first line identifies the framework and version.
Subsequent lines are artifact and verdict entries.

Each line has four fields in fixed order, separated by
`;` — verdict entries carry a fifth:

- **`ARTIFACT/<name>`** or **`VERDICT/<name>`** — the
  logical name of the artifact or verdict.
- **`path:<path>`** — the output file path, relative to
  the project root.
- **`checksum:<hash>`** — hash of the file content at
  the time of generation.
- **`chain:<hash>`** — the chain hash at the time of
  generation.
- **`result:<value>`** — verdict entries only: `pass`,
  `fail`, or `accepted` (set by `accept`).

All hashes use the same algorithm and encoding defined
in CHAIN_HASH.md (SHA-1, base64url, 27 characters).

---

## Artifact and verdict status

The `validate_specs` tool reports the status of each
artifact and verdict. Five states exist, applying
equally to both. A verdict's `result` is orthogonal to
its status — a verdict can be up-to-date and `fail`.

### Up-to-date

The chain hash in the manifest matches the current chain
hash of the node, and the checksum in the manifest matches
the hash of the file on disk.

### Stale

The chain hash in the manifest does not match the current
chain hash of the node. The specification has changed since
the artifact or verdict was last generated. It must be
regenerated.

### Modified

The checksum in the manifest does not match the hash of
the file on disk. The file was modified outside of the
framework. An entry can be both stale and modified.

### Missing

The artifact or verdict file does not exist on disk.

### Orphan

The manifest contains an entry whose logical name does not
correspond to any existing node in the spec tree. The node
was deleted or renamed, but the file and manifest entry
remain.

---

## Concurrency

The manifest may be read and written concurrently during
artifact generation. Any process that reads the manifest
must acquire a shared lock; any process that writes it
must acquire an exclusive lock. A read must not see a
partially-written manifest, and two writes must not
interleave.
