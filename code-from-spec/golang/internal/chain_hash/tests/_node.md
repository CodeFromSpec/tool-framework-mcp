---
outputs:
  - id: chainhash_test
    path: internal/chainhash/chainhash_test.go
---

# ROOT/golang/internal/chain_hash/tests

Tests for the chainhash package.

# Agent

Use `t.TempDir()` to create isolated spec trees.
Verify that the hash is deterministic, 27 characters,
and changes when any file in the chain changes.
