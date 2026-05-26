<!-- code-from-spec: ROOT/functional/utils/name_normalization@TZGR4KKTP-5o2xZnTcDQnk07yTs -->

# NormalizeName

## Function Signature

```
function NormalizeName(raw_string) -> string
```

**Parameters**

- `raw_string` — string — the raw input to normalize

**Returns**

- string — the normalized form of the input

**Contracts**

- Pure function: no I/O, no side effects, raises no errors
- Deterministic: same input always produces same output

---

## Step-by-Step Logic

1. If `raw_string` is empty, return `""`.

2. Trim leading and trailing whitespace from `raw_string`.
   Whitespace characters are:
   - Space (U+0020)
   - Horizontal tab (U+0009)

3. Scan the trimmed string from left to right.
   For each run of one or more consecutive whitespace characters
   (U+0020 or U+0009), replace the entire run with a single
   space (U+0020).
   Characters that are not whitespace are left in place unchanged.

4. Apply Unicode simple case folding to the result of step 3.
   Simple case folding maps each character to its folded
   equivalent as defined by the Unicode standard.
   Example: "Straße" folds to "strasse" because "ß" folds to "ss".

5. Return the resulting string.

---

## Examples

| Input                       | Output                  |
|-----------------------------|-------------------------|
| `"  Interface  "`           | `"interface"`           |
| `"PUBLIC"`                  | `"public"`              |
| `"Straße"`                  | `"strasse"`             |
| `"Testes   de   aceitação"` | `"testes de aceitação"` |
| `""`                        | `""`                    |

---

## Error Conditions

None. This function is pure and raises no errors under any input.
