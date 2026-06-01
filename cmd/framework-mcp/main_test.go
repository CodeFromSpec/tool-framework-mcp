// code-from-spec: ROOT/golang/tests/server@bS6NYCvg4tdCJsfwufWcaNmcJbo
package main_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "framework-mcp-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	binName := "framework-mcp"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binaryPath = filepath.Join(tmpDir, binName)

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: failed to build binary: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// TestHelpFlagPrintsUsageToStdout verifies that --help exits 0 and prints usage to stdout.
func TestHelpFlagPrintsUsageToStdout(t *testing.T) {
	cmd := exec.Command(binaryPath, "--help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0, got error: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("expected stdout to contain usage message, got: %s", string(out))
	}
}

// TestHelpWordPrintsUsageToStdout verifies that the "help" argument exits 0 and prints usage to stdout.
func TestHelpWordPrintsUsageToStdout(t *testing.T) {
	cmd := exec.Command(binaryPath, "help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0, got error: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("expected stdout to contain usage message, got: %s", string(out))
	}
}

// TestShortHelpFlagPrintsUsageToStdout verifies that -h exits 0 and prints usage to stdout.
func TestShortHelpFlagPrintsUsageToStdout(t *testing.T) {
	cmd := exec.Command(binaryPath, "-h")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0, got error: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("expected stdout to contain usage message, got: %s", string(out))
	}
}

// TestUnrecognizedArgumentPrintsUsageToStderr verifies that an unrecognized argument exits 1
// and prints usage to stderr.
func TestUnrecognizedArgumentPrintsUsageToStderr(t *testing.T) {
	cmd := exec.Command(binaryPath, "something")
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected exit 1, got exit 0")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected *exec.ExitError, got: %T %v", err, err)
	}
	if exitErr.ExitCode() != 1 {
		t.Errorf("expected exit code 1, got: %d", exitErr.ExitCode())
	}
	if !strings.Contains(stderrBuf.String(), "Usage:") {
		t.Errorf("expected stderr to contain usage message, got: %s", stderrBuf.String())
	}
}

// TestMultipleArgumentsPrintsUsageToStderr verifies that multiple arguments exit 1
// and print usage to stderr.
func TestMultipleArgumentsPrintsUsageToStderr(t *testing.T) {
	cmd := exec.Command(binaryPath, "foo", "bar")
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected exit 1, got exit 0")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected *exec.ExitError, got: %T %v", err, err)
	}
	if exitErr.ExitCode() != 1 {
		t.Errorf("expected exit code 1, got: %d", exitErr.ExitCode())
	}
	if !strings.Contains(stderrBuf.String(), "Usage:") {
		t.Errorf("expected stderr to contain usage message, got: %s", stderrBuf.String())
	}
}

// testSendMCPRequests starts the binary as a subprocess, writes JSON-RPC messages to stdin,
// and returns all JSON-RPC response lines from stdout.
func testSendMCPRequests(t *testing.T, requests []map[string]any) []map[string]any {
	t.Helper()

	cmd := exec.Command(binaryPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("testSendMCPRequests: StdinPipe: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("testSendMCPRequests: StdoutPipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("testSendMCPRequests: Start: %v", err)
	}
	t.Cleanup(func() {
		_ = stdin.Close()
		_ = cmd.Wait()
	})

	for _, req := range requests {
		data, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("testSendMCPRequests: Marshal: %v", err)
		}
		if _, err := fmt.Fprintf(stdin, "%s\n", data); err != nil {
			t.Fatalf("testSendMCPRequests: write stdin: %v", err)
		}
	}

	var responses []map[string]any
	scanner := bufio.NewScanner(stdout)
	for len(responses) < len(requests) && scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		var resp map[string]any
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			t.Fatalf("testSendMCPRequests: Unmarshal response: %v\nline: %s", err, line)
		}
		responses = append(responses, resp)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("testSendMCPRequests: scanner: %v", err)
	}

	return responses
}

// testMCPRequests returns the standard initialize + tools/list request sequence.
func testMCPRequests() []map[string]any {
	return []map[string]any{
		{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "initialize",
			"params": map[string]any{
				"protocolVersion": "2024-11-05",
				"capabilities":    map[string]any{},
				"clientInfo": map[string]any{
					"name":    "test-client",
					"version": "0.0.1",
				},
			},
		},
		{
			"jsonrpc": "2.0",
			"id":      2,
			"method":  "tools/list",
			"params":  map[string]any{},
		},
	}
}

// testParseToolsList extracts the tools array from a tools/list JSON-RPC response.
func testParseToolsList(t *testing.T, responses []map[string]any) []map[string]any {
	t.Helper()

	// Find the response with id == 2 (tools/list).
	var toolsResp map[string]any
	for _, r := range responses {
		if id, ok := r["id"]; ok {
			switch v := id.(type) {
			case float64:
				if v == 2 {
					toolsResp = r
				}
			}
		}
	}
	if toolsResp == nil {
		t.Fatalf("testParseToolsList: tools/list response not found in %v", responses)
	}

	result, ok := toolsResp["result"].(map[string]any)
	if !ok {
		t.Fatalf("testParseToolsList: result field missing or wrong type: %v", toolsResp)
	}
	toolsRaw, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("testParseToolsList: tools field missing or wrong type: %v", result)
	}

	tools := make([]map[string]any, 0, len(toolsRaw))
	for _, raw := range toolsRaw {
		tool, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("testParseToolsList: tool entry wrong type: %T", raw)
		}
		tools = append(tools, tool)
	}
	return tools
}

// testFindTool finds a tool by name from a list; returns nil if not found.
func testFindTool(tools []map[string]any, name string) map[string]any {
	for _, tool := range tools {
		if tool["name"] == name {
			return tool
		}
	}
	return nil
}

// TestToolsListAdvertisesMaxResultSizeCharsForLoadChain verifies that load_chain
// has _meta["anthropic/maxResultSizeChars"] == 500000.
func TestToolsListAdvertisesMaxResultSizeCharsForLoadChain(t *testing.T) {
	responses := testSendMCPRequests(t, testMCPRequests())
	tools := testParseToolsList(t, responses)

	loadChainTool := testFindTool(tools, "load_chain")
	if loadChainTool == nil {
		t.Fatal("load_chain tool not found in tools/list response")
	}

	meta, ok := loadChainTool["_meta"].(map[string]any)
	if !ok {
		t.Fatalf("load_chain tool missing _meta field or wrong type: %v", loadChainTool)
	}

	maxSize, ok := meta["anthropic/maxResultSizeChars"]
	if !ok {
		t.Fatal("load_chain _meta missing anthropic/maxResultSizeChars")
	}

	maxSizeFloat, ok := maxSize.(float64)
	if !ok {
		t.Fatalf("anthropic/maxResultSizeChars wrong type %T, value: %v", maxSize, maxSize)
	}
	if maxSizeFloat != 500000 {
		t.Errorf("expected anthropic/maxResultSizeChars == 500000, got %v", maxSizeFloat)
	}
}

// TestToolsListAdvertisesAllFourTools verifies that all four expected tools are advertised.
func TestToolsListAdvertisesAllFourTools(t *testing.T) {
	responses := testSendMCPRequests(t, testMCPRequests())
	tools := testParseToolsList(t, responses)

	expectedTools := []string{"load_chain", "write_file", "validate_specs", "hash_fragment"}
	for _, name := range expectedTools {
		if testFindTool(tools, name) == nil {
			t.Errorf("expected tool %q not found in tools/list response", name)
		}
	}
}
