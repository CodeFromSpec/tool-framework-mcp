---
outputs:
  - id: text_normalization
    path: code-from-spec/functional/logic/utils/text_normalization/output.md
---

# ROOT/functional/logic/utils/text_normalization

Normalizes text for comparison.

# Public

## Interface

```
function NormalizeText(raw_string) -> string
```

### Examples

| Input | Output |
|---|---|
| `"  Interface  "` | `"interface"` |
| `"PUBLIC"` | `"public"` |
| `"Straße"` | `"strasse"` |
| `"Testes   de   aceitação"` | `"testes de aceitação"` |
| `""` | `""` |

# Agent

## Behavior

Given a raw string:

1. Trim leading and trailing whitespace.
2. Collapse each run of whitespace characters to a single
   space (U+0020).
3. Apply Unicode simple case folding.

Whitespace characters are space (U+0020) and horizontal tab
(U+0009).

## Contracts

- Pure function — no I/O, no errors.
- Deterministic — same input always produces same output.
