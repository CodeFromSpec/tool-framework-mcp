# Context for AI Agent

## What this project is

An MCP server (`framework-mcp`) for Code from Spec v3 projects.
It provides 4 tools: `load_chain`, `write_file`, `validate_specs`,
`hash_fragment`.

The project follows the Code from Spec methodology — specs are
the source of truth, code is generated from them. See
`CODE_FROM_SPEC.md` for the framework rules.

## Spec tree structure

```
code-from-spec/
├── _node.md                          ← ROOT
├── functional/                       ← functional layer (language-agnostic)
│   ├── _node.md                      ← output format, constraints, domain info
│   ├── dependencies/
│   │   └── owasp-path-traversal/     ← security reference
│   ├── utils/                        ← internal components
│   │   ├── file_reader/              ← sequential line reader
│   │   ├── frontmatter/              ← YAML frontmatter parser
│   │   ├── name_normalization/       ← heading text normalization
│   │   ├── logical_names/            ← logical name ↔ file path mapping
│   │   ├── path_validation/          ← path traversal prevention
│   │   ├── node_parsing/             ← spec node body parser
│   │   ├── node_discovery/           ← find all _node.md in tree
│   │   ├── node_ranking/             ← iterative ranking + cycle detection
│   │   ├── artifact_tag/             ← extract code-from-spec: name@hash
│   │   ├── format_validation/        ← linter for spec structural rules
│   │   └── chain_hash/               ← compute chain hash from raw files
│   └── mcp_tools/                    ← MCP tool specifications
│       ├── load_chain/               ← load spec chain + hash
│       ├── write_file/               ← write generated file
│       ├── validate_specs/           ← validate tree + staleness
│       └── hash_fragment/            ← hash a line range
├── golang/                           ← Go implementation layer
│   ├── _node.md                      ← Go module, language, conventions
│   ├── dependencies/                 ← Go libraries
│   │   ├── goccy-go-yaml/
│   │   ├── golang-x-text/
│   │   ├── google-uuid/
│   │   ├── mcp-go-sdk/
│   │   └── yuin-goldmark/
│   ├── go_module/
│   ├── server/                       ← main entry point
│   │   ├── code/
│   │   └── tests/
│   └── internal/                     ← Go packages
│       ├── chain_hash/               ← NEW, reads raw files for hashing
│       ├── chain_resolver/
│       ├── file_reader/
│       ├── frontmatter/
│       ├── logical_names/
│       ├── normalizename/
│       ├── parsenode/
│       ├── pathvalidation/
│       ├── node_discovery/
│       ├── node_ranking/
│       ├── artifact_tag/
│       ├── format_validation/
│       └── tools/
│           ├── load_chain/
│           ├── write_file/
│           ├── validate_specs/
│           └── hash_fragment/
```

Each leaf node in `functional/` has `outputs` pointing to an
`output.md` (pseudocode) in the same directory. Each leaf `code/`
node in `golang/` has `input: ARTIFACT/functional/...` to consume
the pseudocode, and `outputs` pointing to a `.go` file.

## Layers

- **functional/** — language-agnostic specs. Leaf nodes generate
  pseudocode (`output.md`). Interface in `# Public`, behavior in
  `# Agent`.
- **golang/** — Go implementation. Leaf `code/` nodes consume
  functional pseudocode via `input:` and generate `.go` files.
  Leaf `tests/` nodes generate `_test.go` files.

## What works

- The MCP server compiles and runs. Binary at `tools/framework-mcp.exe`.
- `load_chain` — loads chain, computes chain hash (via `chainhash`
  package), returns `chain_hash: <hash>\n\n<context stream>`.
  Input artifact separated by `\n--- input ---\n` marker.
- `write_file` — writes files, validates against `outputs`.
- `validate_specs` — discovers nodes, validates format, detects
  cycles (via node ranking), checks staleness.
- `hash_fragment` — computes SHA-1 hash of a file line range.
- Subagent generation works: `.claude/agents/code-from-spec-code-generation.md`
  dispatches subagents that call `load_chain` then `write_file`.
- Chain hash is now computed from raw file bytes (not reconstructed
  data) in `internal/chainhash/chainhash.go`, shared by both
  `load_chain` and `validate_specs`.
- Artifact tag extraction handles comment syntax (e.g., `<!-- -->`)
  by extracting only base64url characters after `@`.

## Current bugs

### Node ranking fails on ARTIFACT/ input references

**File**: `internal/noderanking/noderanking.go`

The `DetectCycles` function builds an `allEntries` map containing
node logical names and artifact file paths. But `input` fields
contain ARTIFACT/ logical names (e.g.,
`ARTIFACT/functional/utils/frontmatter(frontmatter)`), which are
NOT in `allEntries`. This causes `ErrUnresolvableRef`, ranking
fails silently, all ranks stay 0, and staleness entries come out
in alphabetical order instead of rank order.

**Fix needed**: The ranking must understand ARTIFACT/ references.
An ARTIFACT/ reference like `ARTIFACT/functional/utils/frontmatter(frontmatter)`
should resolve to the artifact entry for the output with id
`frontmatter` in node `ROOT/functional/utils/frontmatter`. The
`allEntries` map should include ARTIFACT/ logical names mapped
to their corresponding artifact entries.

### Path validation on Windows

**File**: `internal/pathvalidation/pathvalidation.go`

Fixed for case-insensitive comparison on Windows, but may still
have edge cases. The `resolveExistingPrefix` function walks up
to find the deepest existing ancestor when the full path doesn't
exist yet.

## What is stale / needs regeneration

Run `validate_specs` to get the current list. As of now:

- **14 functional output.md** — some were regenerated with correct
  hashes, some still have old hashes, 1 (chain_hash) is missing.
- **16 golang .go files** — all stale (artifact tags have `@PENDING`).
- **16 golang _test.go files** — all missing.
- **1 golang main.go** — stale.
- **1 golang main_test.go** — missing.

## Key design decisions

- **Chain hash from raw bytes**: The chain hash MUST be computed
  from raw file content read from disk, with only CRLF→LF
  normalization. Never from parsed/reconstructed data. This is
  implemented in `internal/chainhash/chainhash.go`.
- **No heredoc delimiters**: `load_chain` returns a continuous
  context stream with no file boundaries or metadata markers.
- **3 parts in load_chain result**: `chain_hash: <hash>\n\n`
  prefix, then context stream, then optional `\n--- input ---\n`
  with input content.
- **Artifacts alongside specs**: `output.md` files live next to
  their `_node.md` in `code-from-spec/functional/...`.
- **Iterative ranking** for cycle detection and processing order
  (not DFS). Based on the algorithm from `tool-staleness-check`.
- **Pseudocode format**: Defined in `ROOT/functional # Public`.
  Uses plain types, no language syntax, step-by-step logic.

## Companion documents

- `CODE_FROM_SPEC.md` — framework rules (v3)
- `CHAIN_HASH.md` — chain hash algorithm
- `FILE_FORMAT.md` — file format details
- `ARTIFACT_GENERATION.md` — how to generate artifacts
- `.claude/agents/code-from-spec-code-generation.md` — subagent definition

## Global rules (from CLAUDE.md)

- Never run git commands. User manages git manually.
- Prefer native file tools over Bash.
- Do not save memories automatically.
