<!-- code-from-spec: ROOT/functional/logic/utils/text_normalization@hcicup4toMEgf_0WvZnVAAAt12E -->

# Text Normalization

## function NormalizeText(raw_string) -> string

Parameters:
- raw_string: string — the input text to normalize

Returns:
- string — the normalized text

Contracts:
- Pure function: no I/O, no side effects, no errors raised.
- Deterministic: identical input always produces identical output.

Steps:

  1. If raw_string is empty, return "".

  2. Trim all leading and trailing whitespace characters
     from raw_string.
     Whitespace characters are:
       - space (U+0020)
       - horizontal tab (U+0009)

  3. Scan the trimmed string from left to right.
     For each run of one or more consecutive whitespace characters
     (U+0020 or U+0009), replace the entire run with a single
     space (U+0020).

  4. Apply Unicode simple case folding to the resulting string.
     This converts each character to its case-folded equivalent
     (e.g., "A" -> "a", "Straße" -> "strasse").

  5. Return the final string.
