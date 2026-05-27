<!-- code-from-spec: ROOT/functional/utils/name_normalization@pExIebEWgCYHxTq_qT26O6-47V0 -->

# NormalizeName

## Function signatures

```
function NormalizeName(raw_string) -> string
```

**Parameters**

- `raw_string` — string — the raw input to normalize; may be empty, have leading/trailing
  whitespace, mixed case, or Unicode characters.

**Returns**

- `string` — the normalized form of `raw_string`.

**Contracts**

- Pure function — no I/O, no external state, no errors raised under any input.
- Deterministic — identical inputs always produce identical outputs.

---

## Step-by-step logic

function NormalizeName(raw_string) -> string

  1. If raw_string is empty, return "".

  2. Trim leading and trailing whitespace from raw_string.
     Whitespace for this purpose means only:
       - space character (U+0020)
       - horizontal tab character (U+0009)
     All other characters — including other Unicode whitespace — are not trimmed.
     Call the result trimmed.

  3. Collapse internal runs of whitespace in trimmed.
     A "run" is one or more consecutive whitespace characters
     (space U+0020 or horizontal tab U+0009).
     Replace each such run with a single space (U+0020).
     Call the result collapsed.

  4. Apply Unicode simple case folding to collapsed.
     Unicode simple case folding maps each code point to its
     case-folded equivalent using the Unicode Simple_Case_Folding
     mapping (as defined in CaseFolding.txt, "S" and "C" entries).
     This mapping is applied code point by code point.
     Note: simple case folding may change the byte length of the
     string (e.g., "ß" U+00DF folds to "ss" — two code points).
     Non-cased characters (digits, punctuation, accented vowels
     that have no case pair, etc.) are left unchanged.
     Call the result folded.

  5. Return folded.

---

## Examples

| Input                      | Step 2 (trim)          | Step 3 (collapse)    | Step 4 (case fold)   | Output               |
|----------------------------|------------------------|----------------------|----------------------|----------------------|
| `"  Interface  "`          | `"Interface"`          | `"Interface"`        | `"interface"`        | `"interface"`        |
| `"PUBLIC"`                 | `"PUBLIC"`             | `"PUBLIC"`           | `"public"`           | `"public"`           |
| `"Straße"`                 | `"Straße"`             | `"Straße"`           | `"strasse"`          | `"strasse"`          |
| `"Testes   de   aceitação"`| `"Testes   de   aceitação"` | `"Testes de aceitação"` | `"testes de aceitação"` | `"testes de aceitação"` |
| `""`                       | `""`                   | —                    | —                    | `""`                 |

---

## Error conditions

None. This is a pure function. It never raises an error for any input.
