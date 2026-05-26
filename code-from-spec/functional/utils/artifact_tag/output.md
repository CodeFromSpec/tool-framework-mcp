<!-- code-from-spec: ROOT/functional/utils/artifact_tag@rkmykwMB1Rop5Zm_of96dj5Z2Zc -->

# artifact_tag

Utilities for locating and parsing the `code-from-spec` tag embedded in
generated files.

---

## Data structures

```
record ArtifactTag
  logical_name: string   -- the spec node name, e.g. "ROOT/golang/server"
  hash:         string   -- exactly 27 characters (base64url chain hash)
```

---

## Functions

### ExtractArtifactTag

```
function ExtractArtifactTag(file_path) -> ArtifactTag
```

**Parameters**

- `file_path` — path to the file to inspect.

**Returns** an `ArtifactTag` record on success.

**Errors**

- `"file unreadable"` — the file cannot be opened or read.
- `"no tag found"` — the file contains no `code-from-spec: ` substring.
- `"malformed tag"` — the tag exists but cannot be parsed
  (no `@` separator, empty logical name, or hash not exactly 27 characters).

**Steps**

1. Open the file at `file_path` using `OpenFileReader`.
   If the file cannot be opened, raise error `"file unreadable"`.

2. Read lines one by one using `ReadLine`.
   For each line:

   a. Check whether the line contains the substring `"code-from-spec: "`.
      If it does not, continue to the next line.

   b. When a matching line is found, stop reading further lines.
      Extract the raw value by taking everything that follows
      `"code-from-spec: "` up to the end of the line,
      then trimming any trailing whitespace.

3. If all lines have been read and no matching line was found,
   raise error `"no tag found"`.

4. Parse the raw value extracted in step 2b:

   a. Find the last occurrence of `"@"` in the raw value.
      If `"@"` is not present, raise error `"malformed tag"`.

   b. Set `logical_name` to everything before the last `"@"`.
      Set `hash` to everything after the last `"@"`.

   c. If `logical_name` is empty, raise error `"malformed tag"`.

   d. If `hash` is not exactly 27 characters long,
      raise error `"malformed tag"`.

5. Return an `ArtifactTag` record with the extracted
   `logical_name` and `hash`.
```

---

## Contracts and invariants

- The file is read only until the first matching line.
  Lines after the first match are never read.
- The `@` used to split name from hash is always the **last** `@`
  in the raw value, allowing logical names that themselves contain `@`.
- The returned `hash` is always exactly 27 characters.
- Comment syntax (`//`, `#`, `/* */`, `--`, `<!-- -->`) is ignored;
  the scan is purely substring-based.
