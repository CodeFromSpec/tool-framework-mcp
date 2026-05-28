# ROOT/golang/implementation/internal/textnormalization

Normalizes text for comparison.

# Public

## Package

`package textnormalization`

## Dependencies

- `golang.org/x/text/cases` — Unicode simple case folding.

## Interface

```go
func NormalizeText(raw string) string
```

Applies the framework normalization rules to a raw heading
or qualifier text:

1. Trim leading and trailing whitespace.
2. Collapse each sequence of one or more whitespace characters
   to a single space (`U+0020`).
3. Apply Unicode simple case folding using `cases.Fold()` from
   `golang.org/x/text/cases`.

Whitespace characters are space (`U+0020`) and horizontal tab
(`U+0009`).
