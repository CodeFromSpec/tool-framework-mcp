# ROOT/golang/implementation/internal/artifact_tag

Extracts the artifact tag from generated files for
staleness detection.

# Public

## Package

`package artifacttag`

## Interface

```go
type ArtifactTag struct {
    LogicalName string
    Hash        string
}

var (
    ErrFileUnreadable = errors.New("file unreadable")
    ErrNoTagFound     = errors.New("no tag found")
    ErrMalformedTag   = errors.New("malformed tag")
)

func ExtractArtifactTag(filePath string) (*ArtifactTag, error)
```

`ExtractArtifactTag` opens a file and scans line by line
for the `code-from-spec: <logical-name>@<hash>` pattern.
Returns the first match found.

### Error handling

All errors wrap a sentinel so callers can use `errors.Is()`:

| Sentinel | Returned when |
|---|---|
| `ErrFileUnreadable` | The file cannot be opened or read. |
| `ErrNoTagFound` | The file has no `code-from-spec:` substring. |
| `ErrMalformedTag` | The tag exists but cannot be parsed (no `@`, empty name, wrong hash length). |
