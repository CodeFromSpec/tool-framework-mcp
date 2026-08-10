---
input: EXTERNAL/code-from-spec/external/code-from-spec/MANIFEST.md
output: code-from-spec/external/code-from-spec/manifest-format/output.md
---

# SPEC/external/code-from-spec/manifest-format

Extracts manifest file format information from the
Code from Spec v6 MANIFEST.md specification document.

# Agent

Extract exactly the following from the input:

1. The header line format (with example).
2. The entry line format — fields, separators, ordering
   (with example). Include the verdict entry format with
   the `result` field.

Nothing else.

Place all content under a single `# Manifest format`
heading.

The output will be used as minimal context to teach an
AI agent how to parse and write the manifest file.
Keep it concise.
