---
output: internal/subagenttoken/subagenttoken.go
---

# SPEC/golang/implementation/subagent_token

Generates and validates opaque tokens that stand in for
a logical name when handed to a generation subagent. A
subagent cannot construct a valid token on its own, so
passing tokens instead of raw logical names to
`load_chain` and `write_file` confines the subagent to
the node it was dispatched for.

# Public

## Package

`package subagenttoken`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/subagenttoken"`

## Interface

```go
func SubagentTokenGenerate(logicalName string) (string, error)
func SubagentTokenValidate(token string) (string, error)
```

### SubagentTokenGenerate

Receives a logical name and returns an opaque token
string that encodes it.

### SubagentTokenValidate

Receives a token string and returns the logical name it
encodes. Fails if the token was not produced by
`SubagentTokenGenerate` with the fixed key below (wrong
size, malformed base64, or failed AEAD authentication).

### Errors

- `ErrInvalidToken`: the token is malformed or fails
  authentication.

# Agent

Implement the subagent token component as a Go package.

## Logic

### Fixed key

Use this hardcoded 32-byte (AES-256) key, declared as a
package-level `var` initialized from a hex literal
decoded with `encoding/hex`:

```
4b1e9f2a7c3d8e05f61a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e6f7
```

This key is fixed in source on purpose — it is not a
security boundary against someone reading the code, only
a way to make tokens unforgeable by a subagent that can
only call the two functions below.

### SubagentTokenGenerate(logicalName: string) -> (string, error)

1. Create an AES cipher block from the fixed key using
   `aes.NewCipher`.
2. Wrap it in a GCM `cipher.AEAD` with
   `cipher.NewGCM`.
3. Generate `aead.NonceSize()` random bytes with
   `crypto/rand.Read` as the nonce. If it fails,
   propagate the error.
4. Compute `sealed = aead.Seal(nil, nonce, []byte(logicalName), nil)`.
5. Concatenate `nonce || sealed`.
6. Encode the concatenation with
   `base64.RawURLEncoding` and return it.

### SubagentTokenValidate(token: string) -> (string, error)

1. Decode `token` with `base64.RawURLEncoding`. If it
   fails, return `ErrInvalidToken`.
2. Create the AES-GCM AEAD exactly as in generation.
3. If the decoded data is shorter than
   `aead.NonceSize()`, return `ErrInvalidToken`.
4. Split into `nonce = data[:aead.NonceSize()]` and
   `sealed = data[aead.NonceSize():]`.
5. Compute `plaintext, err = aead.Open(nil, nonce, sealed, nil)`.
   If `err` is not nil, return `ErrInvalidToken`.
6. Return `string(plaintext)`.

## Go-specific guidance

- Use `crypto/aes`, `crypto/cipher`, `crypto/rand`,
  `encoding/base64`, and `encoding/hex` from the standard
  library only.
- The package name should be `subagenttoken`.
- Decode the fixed key once, at package initialization
  (`var` with `hex.DecodeString` inside an `init()` or a
  package-level `var` assignment using
  `must`-style panic-on-error, consistent with how other
  packages in this codebase initialize package-level
  constants derived from literals).
