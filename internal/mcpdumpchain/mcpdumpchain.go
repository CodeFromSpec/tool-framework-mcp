package mcpdumpchain

import (
	"strings"

	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/mcploadchain"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/subagenttoken"
)

func MCPDumpChain(logicalName string) (string, error) {
	token, err := subagenttoken.SubagentTokenGenerate(logicalName)
	if err != nil {
		return "", err
	}

	chainContent, err := mcploadchain.MCPLoadChain(token)
	if err != nil {
		return "", err
	}

	dumpPath := "code-from-spec/.dump/" + strings.ReplaceAll(logicalName, "/", "_") + ".xml"

	handle, err := oslayer.OpenFile(oslayer.CfsPath(dumpPath), "overwrite", 30000)
	if err != nil {
		return "", err
	}

	err = handle.Write(chainContent)
	if err != nil {
		handle.Close()
		return "", err
	}

	handle.Close()

	return "wrote " + dumpPath, nil
}
