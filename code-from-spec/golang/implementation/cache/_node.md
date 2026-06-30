---
depends_on:
  - ARTIFACT/domain/code-from-spec/cache-details
  - SPEC/golang/implementation/chain/hash
  - SPEC/golang/implementation/oslayer(interface)
output: internal/cache/cache.go
---

# SPEC/golang/implementation/cache

Manages the spec chain cache — a local, best-effort
store that records the content of each chain position
and the structure of each chain at the time they were
computed. This allows the tooling to show subagents
what changed between generations.

# Public

## Package

`package cache`

## Interface

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/cache"`

```go
func WriteContent(contentHash string, content string) error
func WriteChain(chainHash string, positions []chainhash.ContentHash) error
func ReadContent(contentHash string) (string, error)
func ReadChain(chainHash string) ([]chainhash.ContentHash, error)
func ListContentHashes() ([]string, error)
func ListChainHashes() ([]string, error)
func DeleteContent(contentHash string) error
func DeleteChain(chainHash string) error
```

Uses `chainhash.ContentHash` for chain positions — the
same type returned by `ChainHashCompute`.

### WriteContent

Writes the processed content of a chain position to
the content store. Write-once: if a file with the
given hash already exists, returns nil without writing.
The write is atomic — content is written to a temporary
file and renamed.

### WriteChain

Writes the structure of a chain to the chain store.
Write-once: if a file with the given hash already
exists, returns nil without writing. The write is
atomic.

### ReadContent

Returns the content stored under the given hash.
Returns `ErrNotFound` if the file does not exist.

### ReadChain

Returns the chain structure stored under the given
hash. Returns `ErrNotFound` if the file does not
exist. Returns `ErrChainFileCorrupted` if a line
does not match the expected `label: content-hash`
format — this signals that the cache file was
altered and cannot be trusted.

### ListContentHashes

Returns all content hashes present in the content
store. Each hash is the 27-character base64url string
(without the dot prefix used in the filename).

### ListChainHashes

Returns all chain hashes present in the chain store.
Each hash is the 27-character base64url string
(without the dot prefix used in the filename).

### DeleteContent

Deletes a content file from the content store.

### DeleteChain

Deletes a chain file from the chain store.

## Storage layout

The cache lives at `code-from-spec/.cache/` with two
subdirectories:

```
code-from-spec/.cache/
├── .content/    ← content files, keyed by content hash
└── .chains/     ← chain files, keyed by chain hash
```

Content files are named `.<content-hash>` (dot-prefixed,
27-character base64url hash, no extension). The file
content is the processed text of the position.

Chain files are named `.<chain-hash>` (same naming). Each
line has a label and a content hash separated by `: `:

```
SPEC/payments: d4e5f6g7h8i9j0k1l2m3n4o5p6q
SPEC/payments/fees: g7h8i9j0k1l2m3n4o5p6q7r8s
AGENT[SPEC/payments/fees/calculation]: p6q7r8s9t0u1v2w3x4y5z6a
INPUT[ARTIFACT/functional/calc]: s9t0u1v2w3x4y5z6a7b8c9d
```

## Constraints

- Cache files are write-once — once created, they are
  never modified.
- Writes are atomic (write to temporary file, then
  rename). A cache file either exists completely or
  does not exist.
- Concurrent writes of the same hash produce identical
  content. One rename wins; the result is correct
  either way.
- Reads always see a complete file or no file.
- The cache is best-effort infrastructure. Errors
  during write are not fatal — the framework works
  without cache, it just cannot show diffs.

# Agent

Implement the cache component as a Go package.

## Constants

```go
const (
    contentDir = "code-from-spec/.cache/.content"
    chainDir   = "code-from-spec/.cache/.chains"
)
```

## Logic

### WriteContent

1. Build the target path: `contentDir + "/." + contentHash`.
2. Try `oslayer.OpenFile` on the target CfsPath with
   mode "read" and timeout 0.
   If it succeeds (file exists), call `.Close()` and
   return nil.
   If it returns `oslayer.ErrFileUnreadable`, continue
   (file does not exist).
   If it returns `oslayer.ErrLockTimeout`, continue
   (file is being written by another process — we
   proceed with our own write attempt since the rename
   will resolve the race).
   If it returns any other error, return the error.
3. Generate a temporary path:
   `contentDir + "/._tmp_" + contentHash`.
4. Open the temporary path with `oslayer.OpenFile` in
   mode "overwrite" and timeout 30000.
5. Call `.Write(content)`.
6. Call `.Close()`.
7. Call `oslayer.RenameFile(tempPath, targetPath)`.
8. Return nil.

### WriteChain

1. Build the target path: `chainDir + "/." + chainHash`.
2. Check if file exists (same logic as WriteContent
   step 2).
3. Generate a temporary path:
   `chainDir + "/._tmp_" + chainHash`.
4. Open the temporary path with `oslayer.OpenFile` in
   mode "overwrite" and timeout 30000.
5. For each position in positions:
   Write `position.Label + ": " + position.Hash + "\n"`.
6. Call `.Close()`.
7. Call `oslayer.RenameFile(tempPath, targetPath)`.
8. Return nil.

### ReadContent

1. Build the path: `contentDir + "/." + contentHash`.
2. Try `oslayer.OpenFile` on the CfsPath with mode "read"
   and timeout 30000.
   If it returns `oslayer.ErrFileUnreadable`, return
   `ErrNotFound`.
   If it returns any other error, return the error.
3. Read all lines with `.ReadLine()` until
   `oslayer.ErrEndOfFile`.
4. Call `.Close()`.
5. Join lines with "\n" and append a trailing "\n".
6. Return the content.

### ReadChain

1. Build the path: `chainDir + "/." + chainHash`.
2. Try `oslayer.OpenFile` on the CfsPath with mode "read"
   and timeout 30000.
   If it returns `oslayer.ErrFileUnreadable`, return
   `ErrNotFound`.
   If it returns any other error, return the error.
3. Read all lines with `.ReadLine()` until
   `oslayer.ErrEndOfFile`.
4. Call `.Close()`.
5. For each line:
   Split on ": " (first occurrence) into label and
   hash. If the line does not contain ": ",
   return `ErrChainFileCorrupted`.
   Build a `chainhash.ContentHash` with these values.
6. Return the positions.

### ListContentHashes

1. Call `oslayer.ListAllFiles` on CfsPath(contentDir).
   If it returns `oslayer.ErrDirectoryNotFound`, return
   `(nil, nil)`.
2. For each returned CfsPath, extract the filename
   (last path component). Strip the leading dot to get
   the hash.
3. Return the collected hashes.

### ListChainHashes

1. Call `oslayer.ListAllFiles` on CfsPath(chainDir).
   If it returns `oslayer.ErrDirectoryNotFound`, return
   `(nil, nil)`.
2. For each returned CfsPath, extract the filename
   (last path component). Strip the leading dot to get
   the hash.
3. Return the collected hashes.

### DeleteContent

1. Build the path: `contentDir + "/." + contentHash`.
2. Call `oslayer.DeleteFile` on the CfsPath.

### DeleteChain

1. Build the path: `chainDir + "/." + chainHash`.
2. Call `oslayer.DeleteFile` on the CfsPath.

## Go-specific guidance

- The package name is `cache`.
- Use the `chainhash` package for `ContentHash`.
- Use the `oslayer` package for `OpenFile`, `CfsPath`,
  `RenameFile`, `DeleteFile`, `ListAllFiles`,
  and file methods (`ReadLine`, `Write`, `Close`).
- Use `strings.SplitN(line, ": ", 2)` for parsing
  chain file lines.
- Use `path` (not `path/filepath`) for extracting
  filename components from CfsPath values — CfsPath
  uses forward slashes regardless of OS.
- Skip files whose name does not start with "." or that
  start with "._tmp_" when listing hashes — these are
  temporary files from interrupted writes.
- Wrap errors with `fmt.Errorf` using `%w`.
- Define sentinel errors: `ErrNotFound` (file does not
  exist in cache), `ErrChainFileCorrupted` (malformed
  chain file line — cache was altered externally).
  Callers treat both as cache unavailable.
