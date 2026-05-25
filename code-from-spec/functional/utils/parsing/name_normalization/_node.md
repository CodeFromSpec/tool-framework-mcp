---
outputs:
  - id: name_normalization
    path: artifacts/functional/utils/parsing/name_normalization/output.md
---

# ROOT/functional/utils/parsing/name_normalization

Normalizes heading text and qualifier text for comparison.

# Public

## Interface

```
function NormalizeName(raw_string) -> string
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
