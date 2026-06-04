<!-- code-from-spec: ROOT/functional/logic/utils/text_normalization@h7kvG4cIYhwxdpRkjUNoNwOEtV0 -->

function NormalizeText(raw_string: string) -> string

  1. Trim all leading and trailing whitespace characters from raw_string.
     Whitespace characters are space (U+0020) and horizontal tab (U+0009).

  2. Collapse each consecutive run of whitespace characters (U+0020 and U+0009)
     within the trimmed string to a single space (U+0020).

  3. Apply Unicode simple case folding to the result.
     This converts uppercase characters to their lowercase equivalents,
     including characters like "ß" -> "ss" where applicable.

  4. Return the resulting string.
