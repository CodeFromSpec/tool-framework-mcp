package mcpcreatetoken

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/subagenttoken"
)

var (
	ErrNotASpecReference    = errors.New("logical name is not a SPEC/ reference")
	ErrQualifierNotAllowed  = errors.New("logical name contains a parenthetical qualifier")
)

func MCPCreateToken(logicalName string) (string, error) {
	if !strings.HasPrefix(logicalName, "SPEC/") {
		return "", ErrNotASpecReference
	}

	if strings.Contains(logicalName, "(") {
		return "", ErrQualifierNotAllowed
	}

	token, err := subagenttoken.SubagentTokenGenerate(logicalName)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return token, nil
}
