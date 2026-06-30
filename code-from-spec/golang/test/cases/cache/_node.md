---
depends_on:
  - SPEC/golang/test/utils/chdir
  - SPEC/golang/implementation/cache
  - SPEC/golang/implementation/chain/hash
  - SPEC/golang/implementation/oslayer(interface)
output: internal/cache/cache_test.go
---

# SPEC/golang/test/cases/cache

# Agent

## Test setup guidance

Cache operations use `oslayer.OpenFile`, `RenameFile`,
`DeleteFile`, and `ListAllFiles` internally, which
require a valid project root. Use `testutils.Chdir(t)`
for isolation.

Cache directories (`code-from-spec/.cache/.content/`
and `code-from-spec/.cache/.chains/`) are created
automatically by `oslayer.OpenFile` in "overwrite" mode.

## Test cases

### WriteContent

#### Writes content file

Setup: none (empty cache).

Actions:
1. Call `cache.WriteContent("abcdefghijklmnopqrstuvwxyza",
   "hello world\n")`.

Expected:
- No error.
- File exists at
  `code-from-spec/.cache/.content/.abcdefghijklmnopqrstuvwxyza`
  with content `"hello world\n"`.

#### Write-once — skips existing

Setup:
1. Call `cache.WriteContent("abcdefghijklmnopqrstuvwxyza",
   "first")`.

Actions:
1. Call `cache.WriteContent("abcdefghijklmnopqrstuvwxyza",
   "second")`.

Expected:
- No error.
- File content is still `"first"` (not overwritten).

### WriteChain

#### Writes chain file

Setup: none.

Actions:
1. Call `cache.WriteChain("zyxwvutsrqponmlkjihgfedcbaz",
   []chainhash.ContentHash{
     {Label: "SPEC/root", Hash: "aaaaaaaaaaaaaaaaaaaaaaaaaa1"},
     {Label: "SPEC/root/a", Hash: "bbbbbbbbbbbbbbbbbbbbbbbbbbb"},
     {Label: "AGENT[SPEC/root/a]", Hash: "ccccccccccccccccccccccccccc"},
   })`.

Expected:
- No error.
- File exists at
  `code-from-spec/.cache/.chains/.zyxwvutsrqponmlkjihgfedcbaz`.
- File content is three lines:
  `SPEC/root: aaaaaaaaaaaaaaaaaaaaaaaaaa1\n`
  `SPEC/root/a: bbbbbbbbbbbbbbbbbbbbbbbbbbb\n`
  `AGENT[SPEC/root/a]: ccccccccccccccccccccccccccc\n`

#### Write-once — skips existing

Setup:
1. Write a chain file with one position.

Actions:
1. Call `cache.WriteChain` with same hash but different
   positions.

Expected:
- No error. Original content preserved.

### ReadContent

#### Reads existing content

Setup:
1. Call `cache.WriteContent(hash, "test content\n")`.

Actions:
1. Call `cache.ReadContent(hash)`.

Expected:
- Returns `"test content\n"`, no error.

#### Returns ErrNotFound for missing

Actions:
1. Call `cache.ReadContent("nonexistenthashvalue12345ab")`.

Expected:
- Returns `cache.ErrNotFound`.

### ReadChain

#### Reads existing chain

Setup:
1. Write a chain file with known positions.

Actions:
1. Call `cache.ReadChain(hash)`.

Expected:
- Returns the positions with correct labels and hashes.

#### Returns ErrNotFound for missing

Actions:
1. Call `cache.ReadChain("nonexistenthashvalue12345ab")`.

Expected:
- Returns `cache.ErrNotFound`.

#### Returns ErrChainFileCorrupted for malformed line

Setup:
1. Manually write a file to
   `code-from-spec/.cache/.chains/.<hash>` containing
   a line without `: ` separator (e.g. `"bad line"`).

Actions:
1. Call `cache.ReadChain(hash)`.

Expected:
- Returns `cache.ErrChainFileCorrupted`.

### ListContentHashes

#### Lists hashes

Setup:
1. Write two content files with different hashes.

Actions:
1. Call `cache.ListContentHashes()`.

Expected:
- Returns both hashes (without dot prefix).

#### Empty directory returns nil

Actions:
1. Call `cache.ListContentHashes()` with no cache
   directory.

Expected:
- Returns nil, no error.

### ListChainHashes

#### Lists hashes

Setup:
1. Write two chain files with different hashes.

Actions:
1. Call `cache.ListChainHashes()`.

Expected:
- Returns both hashes (without dot prefix).

### DeleteContent

#### Deletes content file

Setup:
1. Write a content file.

Actions:
1. Call `cache.DeleteContent(hash)`.
2. Call `cache.ReadContent(hash)`.

Expected:
- Delete succeeds.
- ReadContent returns `cache.ErrNotFound`.

### DeleteChain

#### Deletes chain file

Setup:
1. Write a chain file.

Actions:
1. Call `cache.DeleteChain(hash)`.
2. Call `cache.ReadChain(hash)`.

Expected:
- Delete succeeds.
- ReadChain returns `cache.ErrNotFound`.

### Temporary files filtered

#### List does not include temp files

Setup:
1. Write a content file (creates the directory).
2. Manually create a file named `._tmp_somehash` in
   `.cache/.content/`.

Actions:
1. Call `cache.ListContentHashes()`.

Expected:
- Only the real hash is returned, not `_tmp_somehash`.

## Go-specific guidance

- The package name is `cache_test` (external test
  package).
- Use `testutils.Chdir(t)` to create a temp dir and
  set the working directory.
- Use `errors.Is` for error sentinel checks.
- To verify file content on disk, use
  `oslayer.OpenFile` in "read" mode and read lines.
- For the malformed chain file test, write the file
  directly with `oslayer.OpenFile` in "overwrite" mode.
