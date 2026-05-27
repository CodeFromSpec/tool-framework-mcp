---
depends_on:
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
  - ARTIFACT/golang/implementation/os/path_utils(pathutils)
input: ARTIFACT/functional/tests/os/path_utils(path_utils_tests)
outputs:
  - id: pathutils_test
    path: internal/pathutils/pathutils_test.go
---

# ROOT/golang/tests/os/path_utils
