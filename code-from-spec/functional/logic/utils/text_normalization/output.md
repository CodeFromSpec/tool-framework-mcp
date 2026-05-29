<!-- code-from-spec: ROOT/functional/logic/utils/text_normalization@Oso--NTz8yT0bgPb2Dokj1uJOuo -->

# text_normalization

## function NormalizeText(raw_string) -> string

Parameters:
- raw_string: string

Returns: string

### Logic

1. If raw_string is empty, return "".

2. Trim all leading and trailing whitespace characters
   (space U+0020 and horizontal tab U+0009) from raw_string.

3. Scan the trimmed string for runs of one or more whitespace
   characters (space U+0020 and horizontal tab U+0009).
   Replace each such run with a single space (U+0020).

4. Apply Unicode simple case folding to the result.
   For each character, replace it with its case-folded
   equivalent (e.g., "A" -> "a", "Straße" -> "strasse").

5. Return the resulting string.

### Contracts

- Pure function — no I/O, no errors raised under any input.
- Deterministic — same input always produces same output.

### Examples

| Input                      | Output                    |
|----------------------------|---------------------------|
| `"  Interface  "`          | `"interface"`             |
| `"PUBLIC"`                 | `"public"`                |
| `"Straße"`                 | `"strasse"`               |
| `"Testes   de   aceitação"`| `"testes de aceitação"`   |
| `""`                       | `""`                      |
