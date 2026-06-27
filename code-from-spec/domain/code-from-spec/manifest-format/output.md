[//]: # (code-from-spec: SPEC/domain/code-from-spec/manifest@46H21Guh5w3hIRodQFTqKTnBR_A)

# Manifest format

## Header

The first line of the manifest identifies the framework and version:

```
code-from-spec: v5
```

## Entry

Each subsequent line is one artifact entry. Fields are in fixed order, separated by `;`:

```
ARTIFACT/payments/fees/calculation;path:internal/fees/calculation.go;checksum:Kx9mP2vB7wY2tHsJ8dFak4Xz9pQ;chain:Jz3qR7nL5cW1gT4yK8mDfAx0vBe
```

Fields:
1. `ARTIFACT/<name>` — logical name
2. `path:<path>` — output file path, relative to project root
3. `checksum:<hash>` — hash of file content at generation time
4. `chain:<hash>` — chain hash at generation time
