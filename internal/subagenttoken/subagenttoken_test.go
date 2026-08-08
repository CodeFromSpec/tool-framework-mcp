package subagenttoken_test

import (
	"encoding/base64"
	"errors"
	"regexp"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/subagenttoken"
)

func TestRoundTrip_SpecLogicalName(t *testing.T) {
	token, err := subagenttoken.SubagentTokenGenerate("SPEC/a/b")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	logicalName, err := subagenttoken.SubagentTokenValidate(token)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if logicalName != "SPEC/a/b" {
		t.Fatalf("got %q, want %q", logicalName, "SPEC/a/b")
	}
}

func TestRoundTrip_ArtifactLogicalName(t *testing.T) {
	token, err := subagenttoken.SubagentTokenGenerate("ARTIFACT/x/y")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	logicalName, err := subagenttoken.SubagentTokenValidate(token)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if logicalName != "ARTIFACT/x/y" {
		t.Fatalf("got %q, want %q", logicalName, "ARTIFACT/x/y")
	}
}

func TestRoundTrip_EmptyLogicalName(t *testing.T) {
	token, err := subagenttoken.SubagentTokenGenerate("")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	logicalName, err := subagenttoken.SubagentTokenValidate(token)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if logicalName != "" {
		t.Fatalf("got %q, want %q", logicalName, "")
	}
}

func TestTokenProperties_SameNameDifferAcrossCalls(t *testing.T) {
	token1, err := subagenttoken.SubagentTokenGenerate("SPEC/a")
	if err != nil {
		t.Fatalf("generate token1: %v", err)
	}
	token2, err := subagenttoken.SubagentTokenGenerate("SPEC/a")
	if err != nil {
		t.Fatalf("generate token2: %v", err)
	}
	if token1 == token2 {
		t.Fatal("expected tokens to differ across calls")
	}
}

func TestTokenProperties_DifferentNamesDifferentTokens(t *testing.T) {
	token1, err := subagenttoken.SubagentTokenGenerate("SPEC/a")
	if err != nil {
		t.Fatalf("generate token1: %v", err)
	}
	token2, err := subagenttoken.SubagentTokenGenerate("SPEC/b")
	if err != nil {
		t.Fatalf("generate token2: %v", err)
	}
	if token1 == token2 {
		t.Fatal("expected tokens for different logical names to differ")
	}
}

func TestTokenProperties_URLSafeBase64(t *testing.T) {
	token, err := subagenttoken.SubagentTokenGenerate("SPEC/a")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	re := regexp.MustCompile(`^[A-Za-z0-9\-_]+$`)
	if !re.MatchString(token) {
		t.Fatalf("token %q contains non-URL-safe characters", token)
	}
}

func TestErrorCases_MalformedBase64(t *testing.T) {
	_, err := subagenttoken.SubagentTokenValidate("not-valid-base64!!!")
	if !errors.Is(err, subagenttoken.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestErrorCases_TooShort(t *testing.T) {
	_, err := subagenttoken.SubagentTokenValidate("QQ")
	if !errors.Is(err, subagenttoken.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestErrorCases_TamperedCiphertext(t *testing.T) {
	token, err := subagenttoken.SubagentTokenGenerate("SPEC/a")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	raw[len(raw)-1] ^= 0xFF
	tamperedToken := base64.RawURLEncoding.EncodeToString(raw)
	_, err = subagenttoken.SubagentTokenValidate(tamperedToken)
	if !errors.Is(err, subagenttoken.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestErrorCases_EmptyString(t *testing.T) {
	_, err := subagenttoken.SubagentTokenValidate("")
	if !errors.Is(err, subagenttoken.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}
