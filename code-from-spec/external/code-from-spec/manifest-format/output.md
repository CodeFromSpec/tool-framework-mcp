# Manifest format

## Header

The first line of the manifest identifies the framework and version:

```
code-from-spec: v6
```

## Entry

Each subsequent line is one artifact or verdict entry. Fields are in fixed order, separated by `;`:

```
ARTIFACT/payments/fees/calculation;path:internal/fees/calculation.go;checksum:Kx9mP2vB7wY2tHsJ8dFak4Xz9pQ;chain:Jz3qR7nL5cW1gT4yK8mDfAx0vBe
VERDICT/review/fees;path:code-from-spec/review/fees/verdict.md;checksum:Ux1mP2vB7wY2tHsJ8dFak4Xz9pQ;chain:Wz3qR7nL5cW1gT4yK8mDfAx0vBe;result:pass
```

Fields:
1. `ARTIFACT/<name>` or `VERDICT/<name>` — logical name
2. `path:<path>` — output file path, relative to project root
3. `checksum:<hash>` — hash of file content at generation time
4. `chain:<hash>` — chain hash at generation time
5. `result:<value>` — verdict entries only: `pass`, `fail`, or `accepted`
