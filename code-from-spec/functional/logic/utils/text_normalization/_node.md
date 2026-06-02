---
output: code-from-spec/functional/logic/utils/text_normalization/output.md
---

# ROOT/functional/logic/utils/text_normalization

Normalizes text for comparison.

# Public

## Interface

```
function NormalizeText(raw_string: string) -> string
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
(U+0009). Standard library functions that test for
whitespace (e.g. `isSpace`, `IsSpace`) may use a broader
definition that includes characters like non-breaking
space (U+00A0). Avoid them if they do not match this
definition.

## Contracts

- Pure function — no I/O, no errors.
- Deterministic — same input always produces same output.
