package mcpcreatetoken_test

import (
	"errors"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/mcpcreatetoken"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/subagenttoken"
)

func TestMCPCreateToken_RoundTrip(t *testing.T) {
	token, err := mcpcreatetoken.MCPCreateToken("SPEC/a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	logicalName, err := subagenttoken.SubagentTokenValidate(token)
	if err != nil {
		t.Fatalf("unexpected error validating token: %v", err)
	}
	if logicalName != "SPEC/a/b" {
		t.Errorf("expected %q, got %q", "SPEC/a/b", logicalName)
	}
}

func TestMCPCreateToken_DifferentTokensForSameInput(t *testing.T) {
	token1, err := mcpcreatetoken.MCPCreateToken("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	token2, err := mcpcreatetoken.MCPCreateToken("SPEC/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token1 == token2 {
		t.Error("expected different tokens for the same logical name")
	}
}

func TestMCPCreateToken_ArtifactReference(t *testing.T) {
	_, err := mcpcreatetoken.MCPCreateToken("ARTIFACT/x")
	if !errors.Is(err, mcpcreatetoken.ErrNotASpecReference) {
		t.Errorf("expected ErrNotASpecReference, got %v", err)
	}
}

func TestMCPCreateToken_ExternalReference(t *testing.T) {
	_, err := mcpcreatetoken.MCPCreateToken("EXTERNAL/x")
	if !errors.Is(err, mcpcreatetoken.ErrNotASpecReference) {
		t.Errorf("expected ErrNotASpecReference, got %v", err)
	}
}

func TestMCPCreateToken_QualifierNotAllowed(t *testing.T) {
	_, err := mcpcreatetoken.MCPCreateToken("SPEC/a(interface)")
	if !errors.Is(err, mcpcreatetoken.ErrQualifierNotAllowed) {
		t.Errorf("expected ErrQualifierNotAllowed, got %v", err)
	}
}
