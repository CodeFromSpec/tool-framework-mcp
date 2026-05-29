---
depends_on:
  - ARTIFACT/golang/interfaces/chain/chain_resolver(interface)
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
  - ARTIFACT/golang/interfaces/os/file_reader(interface)
  - ARTIFACT/golang/interfaces/parsing/frontmatter(interface)
  - ARTIFACT/golang/interfaces/utils/logical_names(interface)
input: ARTIFACT/functional/tests/chain/chain_resolver(chain_resolver_tests)
outputs:
  - id: chainresolver_test
    path: internal/chainresolver/chainresolver_test.go
---

# ROOT/golang/tests/chain/chain_resolver
