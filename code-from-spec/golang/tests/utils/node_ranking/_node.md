---
depends_on:
  - ARTIFACT/golang/interfaces/utils/node_ranking(interface)
  - ARTIFACT/golang/interfaces/parsing/frontmatter(interface)
  - ARTIFACT/golang/interfaces/utils/logical_names(interface)
input: ARTIFACT/functional/tests/utils/node_ranking(node_ranking_tests)
outputs:
  - id: noderanking_test
    path: internal/noderanking/noderanking_test.go
---

# ROOT/golang/tests/utils/node_ranking
