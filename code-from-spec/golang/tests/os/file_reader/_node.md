---
depends_on:
  - ARTIFACT/golang/interfaces/os/file_reader(interface)
  - ARTIFACT/golang/implementation/os/file_reader(filereader)
input: ARTIFACT/functional/tests/os/file_reader(file_reader_tests)
outputs:
  - id: filereader_test
    path: internal/filereader/filereader_test.go
---

# ROOT/golang/tests/os/file_reader
