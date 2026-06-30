---
input: EXTERNAL/code-from-spec/domain/code-from-spec/CACHE.md
output: code-from-spec/domain/code-from-spec/cache-details/output.md
---

# SPEC/domain/code-from-spec/cache-details

Extracts cache storage details from the Code from Spec
v5 CACHE.md specification document.

# Agent

Extract exactly the following from the input:

1. The directory layout (content store and chain store
   paths).
2. File naming convention (dot prefix, hash length,
   encoding, no extension).
3. Content file format — what is stored.
4. Chain file format — line format with label and
   content hash (with example).
5. Write-once semantics and atomic write requirement.
6. Concurrency rules.

Nothing else.

Place all content under a single `# Cache storage`
heading.

The output will be used as minimal context to teach an
AI agent how to read and write cache files. Keep it
concise.
