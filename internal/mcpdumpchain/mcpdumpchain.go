package mcpdumpchain

import (
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/mcploadchain"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/oslayer"
)

func MCPDumpChain(logicalName string) (string, error) {
	chainContent, err := mcploadchain.MCPLoadChain(logicalName)
	if err != nil {
		return "", err
	}

	handle, err := oslayer.OpenFile(oslayer.CfsPath("dump_chain.xml"), "overwrite", 30000)
	if err != nil {
		return "", err
	}

	err = handle.Write(chainContent)
	if err != nil {
		handle.Close()
		return "", err
	}

	handle.Close()

	return "wrote dump_chain.xml", nil
}
