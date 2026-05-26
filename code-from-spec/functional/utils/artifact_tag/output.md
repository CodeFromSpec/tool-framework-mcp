<!-- code-from-spec: ROOT/functional/utils/artifact_tag@aOo6CinxiLbYIz0W2SSLcFVWfZQ -->

# artifact_tag

Utilities for locating and parsing the `code-from-spec:` artifact tag that is
embedded in every generated source file. The tag records which spec node
produced the file and at which version (hash), enabling the tool to detect
whether a file is out of date.

---

## Data structures

```
record ArtifactTag
  logical_name: string   -- the spec node name, e.g. "ROOT/golang/server"
  hash:         string   -- exactly 27 base64url characters identifying the spec version
```

---

## Functions

### ExtractArtifactTag

```
function ExtractArtifactTag(file_path) -> ArtifactTag
  errors:
    - "file unreadable"  : the file cannot be opened or read.
    - "no tag found"     : no line in the file contains the substring "code-from-spec: ".
    - "malformed tag"    : the tag line exists but the content after "code-from-spec: "
                           cannot be parsed (missing "@", empty logical name,
                           or hash length is not exactly 27 characters).
```

**Step-by-step logic**

1. Open the file at `file_path` using `OpenFileReader`.
   If the file cannot be opened, raise error `"file unreadable"`.

2. Loop: read the next line using `ReadLine`.
   If `ReadLine` raises `"end of file"`, stop the loop and raise error `"no tag found"`.

3. Check whether the current line contains the substring `"code-from-spec: "`.
   If it does not, go back to step 2.

4. A matching line has been found. Stop reading the file.
   Take everything that follows the first occurrence of `"code-from-spec: "` on that line,
   up to the end of the line, and trim any trailing whitespace.
   Call this value `raw`.

5. Find the last occurrence of `"@"` in `raw`.
   If `"@"` is not present, raise error `"malformed tag"`.

6. Split `raw` at that last `"@"`:
   - `logical_name` = everything before the `"@"`.
   - `hash`         = everything after the `"@"`.

7. Validate:
   - If `logical_name` is empty, raise error `"malformed tag"`.
   - If `hash` is not exactly 27 characters long, raise error `"malformed tag"`.

8. Return an `ArtifactTag` record with the extracted `logical_name` and `hash`.
```

---

## Contracts and invariants

- The file is read only until the first line that contains `"code-from-spec: "`.
  Lines after the match are never read.
- Comment syntax (`//`, `#`, `/* */`, `--`, `<!-- -->`, etc.) is not parsed.
  The function matches on the raw substring regardless of surrounding context.
- The hash field of a valid tag is always exactly 27 characters (base64url encoding).
- `logical_name` may contain `/`, `_`, `-`, and alphanumeric characters
  but must not be empty.
