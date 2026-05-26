<!-- code-from-spec: ROOT/functional/utils/artifact_tag@aOo6CinxiLbYIz0W2SSLcFVWfZQ -->

# ArtifactTag

## Data Structures

```
record ArtifactTag
  logical_name: string
  hash: string
```

## Functions

### ExtractArtifactTag

```
function ExtractArtifactTag(file_path) -> ArtifactTag
```

Parameters:
- `file_path`: string — path to the file to inspect

Returns: an ArtifactTag record with `logical_name` and `hash` fields

Errors:
- `"file unreadable"`: the file cannot be opened or read
- `"no tag found"`: no line in the file contains the substring `code-from-spec: `
- `"malformed tag"`: the tag line was found but could not be parsed (no `@` present, logical name is empty, or hash is not exactly 27 characters)

#### Step-by-step logic

1. Call `OpenFileReader(file_path)`.
   If the file cannot be opened, raise error `"file unreadable"`.

2. Repeat:
   a. Call `ReadLine(reader)` to get the next line.
      If `ReadLine` raises `"end of file"`, stop reading and raise error `"no tag found"`.
      If `ReadLine` raises any other read error, raise error `"file unreadable"`.
   b. Check whether the line contains the substring `"code-from-spec: "`.
      If it does not, continue to the next iteration.
      If it does, proceed to step 3.

   (Reading stops as soon as the first matching line is found — the rest of the file is not read.)

3. Take everything after the first occurrence of `"code-from-spec: "` in the matched line.
   Trim any trailing whitespace from this extracted portion.
   Call this value `raw_tag`.

4. Find the last occurrence of `"@"` in `raw_tag`.
   If `"@"` is not present, raise error `"malformed tag"`.

5. Split `raw_tag` at that last `"@"`:
   - `logical_name` = everything before the last `"@"`
   - `hash` = everything after the last `"@"`

6. Validate:
   - If `logical_name` is empty, raise error `"malformed tag"`.
   - If `hash` is not exactly 27 characters long, raise error `"malformed tag"`.

7. Return an ArtifactTag record with:
   - `logical_name` set to the extracted logical name
   - `hash` set to the extracted hash
```
