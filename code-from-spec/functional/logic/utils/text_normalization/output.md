<!-- code-from-spec: SPEC/functional/logic/utils/text_normalization@zZPhFivo301KTHUiVwAW85I0VGM -->

function NormalizeText(raw_string) -> string

  1. If raw_string is empty, return "".

  2. Trim leading and trailing whitespace characters from raw_string,
     where whitespace is defined as space (U+0020) and horizontal tab (U+0009) only.

  3. Collapse each consecutive run of whitespace characters
     (space U+0020 and horizontal tab U+0009) to a single space (U+0020).

  4. Apply Unicode simple case folding to the resulting string.
     This converts uppercase characters to their lowercase equivalents,
     including Unicode mappings (e.g., "Straße" -> "strasse").

  5. Return the normalized string.

Contracts:
  - Pure function — no I/O, no side effects, no errors raised.
  - Deterministic — same input always produces the same output.
  - Whitespace is defined strictly as U+0020 (space) and U+0009 (horizontal tab).
    Do not use standard library "isspace" functions that may match U+00A0 or other
    non-standard whitespace characters.
