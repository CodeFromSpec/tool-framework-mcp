// code-from-spec: ROOT/golang/server/tests@2CJNwmlSVxbZBHOq3ovxq2X-K5g

// Package main contains integration tests for the framework-mcp binary.
//
// Tests compile the binary once via TestMain and then exercise it as a
// subprocess, checking exit codes, stdout/stderr output, and MCP protocol
// responses. No test framework beyond the standard "testing" package is used.
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// testBinPath holds the path to the compiled binary built in TestMain.
// It is set once and shared across all tests in this package.
var testBinPath string

// TestMain builds the binary into a temp directory before running tests,
// and cleans up afterward. This avoids rebuilding on every test case.
func TestMain(m *testing.M) {
	// Create a temporary directory for the binary.
	tmpDir, err := os.MkdirTemp("", "framework-mcp-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	// On Windows, the binary must carry a .exe suffix.
	binName := "framework-mcp"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	testBinPath = filepath.Join(tmpDir, binName)

	// Build the binary. The package path uses the module prefix.
	buildCmd := exec.Command(
		"go", "build",
		"-o", testBinPath,
		"github.com/CodeFromSpec/tool-framework-mcp/v2/cmd/framework-mcp",
	)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build binary: %v\n", err)
		os.Exit(1)
	}

	// Run tests and capture the exit code.
	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// Helper types and functions — all prefixed with "test" per project convention.
// ---------------------------------------------------------------------------

// testRunResult captures the outcome of running the binary as a subprocess.
type testRunResult struct {
	exitCode int
	stdout   string
	stderr   string
}

// testRun executes the compiled binary with the given arguments and returns
// stdout, stderr, and the exit code. It never fails the test on a non-zero
// exit code; callers check the exit code themselves.
func testRun(t *testing.T, args ...string) testRunResult {
	t.Helper()
	cmd := exec.Command(testBinPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := testRunResult{
		stdout: stdout.String(),
		stderr: stderr.String(),
	}

	if err != nil {
		// Extract the exit code from the error.
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("unexpected error running binary: %v", err)
		}
	}
	// exitCode remains 0 if err == nil (clean exit).

	return result
}

// testUsageSubstring is a short fragment of the usage message that all usage
// checks can look for without being brittle to minor whitespace differences.
// Using the tool list header line is unambiguous and stable.
const testUsageSubstring = "Starts an MCP server over stdin/stdout"

// testContains reports whether s contains substr.
func testContains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// ---------------------------------------------------------------------------
// Help flag tests — Happy Path
// ---------------------------------------------------------------------------

// TestHelpFlagLong verifies that "--help" prints the usage message to stdout
// and exits with code 0.
func TestHelpFlagLong(t *testing.T) {
	result := testRun(t, "--help")

	if result.exitCode != 0 {
		t.Errorf("expected exit 0, got %d", result.exitCode)
	}
	if !testContains(result.stdout, testUsageSubstring) {
		t.Errorf("expected usage in stdout; got:\n%s", result.stdout)
	}
}

// TestHelpWord verifies that "help" prints the usage message to stdout
// and exits with code 0.
func TestHelpWord(t *testing.T) {
	result := testRun(t, "help")

	if result.exitCode != 0 {
		t.Errorf("expected exit 0, got %d", result.exitCode)
	}
	if !testContains(result.stdout, testUsageSubstring) {
		t.Errorf("expected usage in stdout; got:\n%s", result.stdout)
	}
}

// TestHelpFlagShort verifies that "-h" prints the usage message to stdout
// and exits with code 0.
func TestHelpFlagShort(t *testing.T) {
	result := testRun(t, "-h")

	if result.exitCode != 0 {
		t.Errorf("expected exit 0, got %d", result.exitCode)
	}
	if !testContains(result.stdout, testUsageSubstring) {
		t.Errorf("expected usage in stdout; got:\n%s", result.stdout)
	}
}

// ---------------------------------------------------------------------------
// Unrecognized argument tests — Failure Cases
// ---------------------------------------------------------------------------

// TestUnrecognizedArgument verifies that a single unrecognized argument causes
// the binary to print the usage message to stderr and exit with code 1.
func TestUnrecognizedArgument(t *testing.T) {
	result := testRun(t, "something")

	if result.exitCode != 1 {
		t.Errorf("expected exit 1, got %d", result.exitCode)
	}
	if !testContains(result.stderr, testUsageSubstring) {
		t.Errorf("expected usage in stderr; got:\n%s", result.stderr)
	}
}

// TestMultipleArguments verifies that multiple arguments cause the binary to
// print the usage message to stderr and exit with code 1.
func TestMultipleArguments(t *testing.T) {
	result := testRun(t, "foo", "bar")

	if result.exitCode != 1 {
		t.Errorf("expected exit 1, got %d", result.exitCode)
	}
	if !testContains(result.stderr, testUsageSubstring) {
		t.Errorf("expected usage in stderr; got:\n%s", result.stderr)
	}
}

// ---------------------------------------------------------------------------
// MCP protocol helpers
// ---------------------------------------------------------------------------

// testJSONRPCRequest is a minimal JSON-RPC 2.0 request envelope.
type testJSONRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// testJSONRPCResponse is a minimal JSON-RPC 2.0 response envelope used to
// decode the server's reply. The Result field is kept as raw JSON for further
// inspection.
type testJSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// testToolsListResult mirrors the structure of a tools/list result.
type testToolsListResult struct {
	Tools []testToolEntry `json:"tools"`
}

// testToolEntry represents a single tool in the tools/list result.
// Meta is decoded into a generic map so we can inspect arbitrary keys.
type testToolEntry struct {
	Name string         `json:"name"`
	Meta map[string]any `json:"_meta"`
}

// testMCPSession starts the binary as a long-running subprocess, sends the
// MCP initialize handshake followed by the provided requests, and returns the
// JSON-RPC responses for those requests (not the initialize response).
//
// The caller is responsible for providing requests with distinct IDs starting
// from 2 (ID 1 is reserved for the initialize request).
func testMCPSession(t *testing.T, requests []testJSONRPCRequest) []testJSONRPCResponse {
	t.Helper()

	// Start the server subprocess.
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	cmd := exec.CommandContext(ctx, testBinPath)

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}

	// Capture stderr for diagnostics if the test fails.
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start binary: %v", err)
	}
	t.Cleanup(func() {
		// Close stdin so the server's Run loop can detect EOF and exit cleanly.
		stdinPipe.Close()
		// Wait for the process to exit to avoid resource leaks.
		_ = cmd.Wait()
		if t.Failed() {
			t.Logf("server stderr:\n%s", stderrBuf.String())
		}
	})

	// Helper: write a single JSON-RPC request followed by a newline.
	sendRequest := func(req testJSONRPCRequest) {
		t.Helper()
		data, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}
		if _, err := fmt.Fprintf(stdinPipe, "%s\n", data); err != nil {
			t.Fatalf("failed to write request: %v", err)
		}
	}

	// Helper: read the next JSON-RPC response line from stdout.
	reader := bufio.NewReader(stdoutPipe)
	readResponse := func() testJSONRPCResponse {
		t.Helper()
		// The MCP SDK writes one JSON object per line over stdio.
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			t.Fatalf("failed to read response line: %v", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			t.Fatal("received empty response line")
		}
		var resp testJSONRPCResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			t.Fatalf("failed to unmarshal response %q: %v", line, err)
		}
		return resp
	}

	// --- MCP initialize handshake ---
	// The MCP protocol requires an initialize/initialized exchange before any
	// other requests can be made.
	initRequest := testJSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2024-11-05",
			"clientInfo": map[string]any{
				"name":    "test-client",
				"version": "0.0.1",
			},
			"capabilities": map[string]any{},
		},
	}
	sendRequest(initRequest)
	// Consume the initialize response (we don't inspect it here).
	_ = readResponse()

	// Send the initialized notification (no ID — it is a notification).
	initializedNotification := map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	}
	notifData, err := json.Marshal(initializedNotification)
	if err != nil {
		t.Fatalf("failed to marshal initialized notification: %v", err)
	}
	if _, err := fmt.Fprintf(stdinPipe, "%s\n", notifData); err != nil {
		t.Fatalf("failed to write initialized notification: %v", err)
	}

	// --- Send caller-supplied requests and collect responses ---
	for _, req := range requests {
		sendRequest(req)
	}

	responses := make([]testJSONRPCResponse, len(requests))
	for i := range responses {
		responses[i] = readResponse()
	}

	return responses
}

// testToolsListRequest returns a tools/list JSON-RPC request with the given ID.
func testToolsListRequest(id int) testJSONRPCRequest {
	return testJSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  "tools/list",
	}
}

// testDecodeToolsListResult decodes a tools/list JSON-RPC result from the raw
// response. Fails the test if the result cannot be decoded.
func testDecodeToolsListResult(t *testing.T, resp testJSONRPCResponse) testToolsListResult {
	t.Helper()
	if resp.Error != nil {
		t.Fatalf("tools/list returned error: code=%d message=%s", resp.Error.Code, resp.Error.Message)
	}
	var result testToolsListResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("failed to decode tools/list result: %v\nraw: %s", err, resp.Result)
	}
	return result
}

// testFindTool searches a tools list for a tool by name. Returns the entry and
// true if found, or an empty entry and false otherwise.
func testFindTool(tools []testToolEntry, name string) (testToolEntry, bool) {
	for _, tool := range tools {
		if tool.Name == name {
			return tool, true
		}
	}
	return testToolEntry{}, false
}

// ---------------------------------------------------------------------------
// MCP protocol tests
// ---------------------------------------------------------------------------

// TestToolsListAllFourTools verifies that the tools/list response advertises
// all four expected tools: load_chain, write_file, validate_specs, hash_fragment.
func TestToolsListAllFourTools(t *testing.T) {
	responses := testMCPSession(t, []testJSONRPCRequest{
		testToolsListRequest(2),
	})

	result := testDecodeToolsListResult(t, responses[0])

	expectedTools := []string{"load_chain", "write_file", "validate_specs", "hash_fragment"}
	for _, name := range expectedTools {
		if _, found := testFindTool(result.Tools, name); !found {
			t.Errorf("tool %q not found in tools/list; got tools: %v", name, testToolNames(result.Tools))
		}
	}
}

// TestToolsListLoadChainMaxResultSize verifies that the load_chain tool entry
// carries _meta["anthropic/maxResultSizeChars"] == 500000 in the tools/list
// response, so the client can allocate appropriate buffers.
func TestToolsListLoadChainMaxResultSize(t *testing.T) {
	responses := testMCPSession(t, []testJSONRPCRequest{
		testToolsListRequest(2),
	})

	result := testDecodeToolsListResult(t, responses[0])

	entry, found := testFindTool(result.Tools, "load_chain")
	if !found {
		t.Fatalf("tool load_chain not found in tools/list")
	}

	// The meta value is decoded as a float64 by encoding/json (all JSON numbers
	// decode to float64 in a map[string]any). Cast accordingly.
	rawVal, ok := entry.Meta["anthropic/maxResultSizeChars"]
	if !ok {
		t.Fatalf("load_chain _meta missing key anthropic/maxResultSizeChars; meta: %v", entry.Meta)
	}

	// JSON numbers decode to float64 in a map[string]any context.
	floatVal, ok := rawVal.(float64)
	if !ok {
		t.Fatalf("anthropic/maxResultSizeChars is %T (%v), expected float64", rawVal, rawVal)
	}

	const wantSize = 500000
	if int(floatVal) != wantSize {
		t.Errorf("anthropic/maxResultSizeChars = %v, want %d", floatVal, wantSize)
	}
}

// ---------------------------------------------------------------------------
// Internal helpers (not test cases)
// ---------------------------------------------------------------------------

// testToolNames extracts the names from a slice of tool entries for use in
// diagnostic messages.
func testToolNames(tools []testToolEntry) []string {
	names := make([]string, len(tools))
	for i, tool := range tools {
		names[i] = tool.Name
	}
	return names
}
