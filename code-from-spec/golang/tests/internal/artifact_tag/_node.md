---
depends_on:
  - ARTIFACT/golang/implementation/internal/artifact_tag/code(artifacttag)
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
  - ARTIFACT/golang/interfaces/os/file_reader(interface)
input: ARTIFACT/functional/tests/parsing/artifact_tag(artifact_tag_tests)
outputs:
  - id: artifacttag_test
    path: internal/artifacttag/artifacttag_test.go
---

# ROOT/golang/tests/internal/artifact_tag
