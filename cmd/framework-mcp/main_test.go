// code-from-spec: ROOT/golang/server/tests@CGBDxLOUr5PZxyfvEmRzW9fMJYI

// Package main tests verify the compiled framework-mcp binary behavior:
// exit codes, stdout/stderr output, and MCP protocol responses.
//
// The binary is compiled once in TestMain into a temporary directory.
// All tests invoke it as a subprocess using os/exec.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// testBinaryPath holds the path to the compiled binary under test.
// It is set once in TestMain before any test runs.
var testBinaryPath string

// TestMain builds the binary once into a temp directory, then runs all tests.
// On Windows the binary must have the .exe extension.
func TestMain(m *testing.M) {
	// Determine binary name based on platform.
	binaryName := "framework-mcp-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	// Create a temporary directory for the binary.
	tmpDir, err := os.MkdirTemp("", "framework-mcp-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	testBinaryPath = filepath.Join(tmpDir, binaryName)

	// Build the binary from the cmd/framework-mcp package.
	// The working directory is the project root (as per spec: "always executed
	// from the project root directory").
	buildCmd := exec.Command("go", "build", "-o", testBinaryPath, "./cmd/framework-mcp")
	buildCmd.Stdout = os.Stderr // send build output to stderr so test log shows it
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build binary: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// Happy Path Tests
// ---------------------------------------------------------------------------

// TestHelpFlagPrintsUsageToStdout verifies that running with --help exits 0
// and prints the usage message to stdout.
func TestHelpFlagPrintsUsageToStdout(t *testing.T) {
	stdout, _, exitCode := testRunBinary(t, "--help")

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
	testAssertContainsUsage(t, "stdout", stdout)
}

// TestHelpWordPrintsUsageToStdout verifies that running with the word "help"
// exits 0 and prints the usage message to stdout.
func TestHelpWordPrintsUsageToStdout(t *testing.T) {
	stdout, _, exitCode := testRunBinary(t, "help")

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
	testAssertContainsUsage(t, "stdout", stdout)
}

// TestShortHelpFlagPrintsUsageToStdout verifies that running with -h exits 0
// and prints the usage message to stdout.
func TestShortHelpFlagPrintsUsageToStdout(t *testing.T) {
	stdout, _, exitCode := testRunBinary(t, "-h")

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
	testAssertContainsUsage(t, "stdout", stdout)
}

// ---------------------------------------------------------------------------
// Failure Case Tests
// ---------------------------------------------------------------------------

// TestUnrecognizedArgumentPrintsUsageToStderr verifies that running with an
// unrecognized argument exits 1 and prints the usage message to stderr.
func TestUnrecognizedArgumentPrintsUsageToStderr(t *testing.T) {
	_, stderr, exitCode := testRunBinary(t, "something")

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
	testAssertContainsUsage(t, "stderr", stderr)
}

// TestMultipleArgumentsPrintsUsageToStderr verifies that running with multiple
// arguments exits 1 and prints the usage message to stderr.
func TestMultipleArgumentsPrintsUsageToStderr(t *testing.T) {
	_, stderr, exitCode := testRunBinary(t, "foo", "bar")

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
	testAssertContainsUsage(t, "stderr", stderr)
}

// ---------------------------------------------------------------------------
// MCP Protocol Tests
// ---------------------------------------------------------------------------

// testJSONRPCRequest is a minimal JSON-RPC 2.0 request structure used when
// communicating with the binary over stdin.
type testJSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// testJSONRPCResponse is a generic JSON-RPC 2.0 response structure.
type testJSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   json.RawMessage `json:"error,omitempty"`
}

// testToolsListResult matches the structure returned by the tools/list method.
type testToolsListResult struct {
	Tools []testToolEntry `json:"tools"`
}

// testToolEntry represents a single tool entry in tools/list.
type testToolEntry struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Meta        map[string]interface{} `json:"_meta,omitempty"`
}

// TestToolsListAdvertisesMaxResultSizeCharsForLoadChain starts the binary as a
// subprocess, sends initialize + tools/list over stdin, and verifies that the
// load_chain tool advertises _meta["anthropic/maxResultSizeChars"] = 500000.
func TestToolsListAdvertisesMaxResultSizeCharsForLoadChain(t *testing.T) {
	toolsResult := testGetToolsList(t)

	// Find the load_chain tool.
	var loadChainTool *testToolEntry
	for i := range toolsResult.Tools {
		if toolsResult.Tools[i].Name == "load_chain" {
			loadChainTool = &toolsResult.Tools[i]
			break
		}
	}

	if loadChainTool == nil {
		t.Fatal("load_chain tool not found in tools/list response")
	}

	// Verify the _meta field contains the expected key.
	if loadChainTool.Meta == nil {
		t.Fatal("load_chain tool has no _meta field")
	}

	rawVal, ok := loadChainTool.Meta["anthropic/maxResultSizeChars"]
	if !ok {
		t.Fatal("load_chain _meta missing key 'anthropic/maxResultSizeChars'")
	}

	// JSON numbers unmarshal as float64 when using interface{}.
	numVal, ok := rawVal.(float64)
	if !ok {
		t.Fatalf("expected 'anthropic/maxResultSizeChars' to be a number, got %T", rawVal)
	}

	const want = 500000
	if int(numVal) != want {
		t.Errorf("expected anthropic/maxResultSizeChars = %d, got %d", want, int(numVal))
	}
}

// TestToolsListAdvertisesAllFourTools starts the binary as a subprocess, sends
// initialize + tools/list over stdin, and verifies that all four required tools
// are advertised.
func TestToolsListAdvertisesAllFourTools(t *testing.T) {
	toolsResult := testGetToolsList(t)

	// Build a set of advertised tool names.
	advertised := make(map[string]bool, len(toolsResult.Tools))
	for _, tool := range toolsResult.Tools {
		advertised[tool.Name] = true
	}

	required := []string{"load_chain", "write_file", "validate_specs", "hash_fragment"}
	for _, name := range required {
		if !advertised[name] {
			t.Errorf("expected tool %q to be advertised, but it was not found", name)
		}
	}
}

// ---------------------------------------------------------------------------
// Test Helpers
// ---------------------------------------------------------------------------

// testRunBinary runs the compiled binary with the given arguments and returns
// (stdout, stderr, exitCode). It does not fail the test on non-zero exit codes
// because callers assert exit codes themselves.
func testRunBinary(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	cmd := exec.Command(testBinaryPath, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			// Unexpected execution error (e.g., binary not found).
			t.Fatalf("unexpected error running binary: %v", err)
		}
	} else {
		exitCode = 0
	}

	return stdout, stderr, exitCode
}

// testAssertContainsUsage checks that the given output string contains key
// phrases from the usage message defined in the spec.
func testAssertContainsUsage(t *testing.T, stream, output string) {
	t.Helper()

	// These phrases are taken directly from the usage message in the spec.
	// Checking a representative subset is sufficient to confirm the usage
	// message was printed — we do not require exact byte-for-byte equality.
	phrases := []string{
		"Usage: framework-mcp",
		"load_chain",
		"write_file",
		"validate_specs",
		"hash_fragment",
	}

	for _, phrase := range phrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("%s does not contain expected phrase %q\nfull %s:\n%s",
				stream, phrase, stream, output)
		}
	}
}

// testGetToolsList is a shared helper that starts the binary as a subprocess,
// sends MCP initialize + tools/list requests, and returns the parsed
// testToolsListResult. The test is failed immediately if any step goes wrong.
func testGetToolsList(t *testing.T) testToolsListResult {
	t.Helper()

	// Start the binary. It reads JSON-RPC from stdin and writes responses to
	// stdout. We close stdin after writing to signal EOF.
	cmd := exec.Command(testBinaryPath)
	var outBuf, errBuf bytes.Buffer
	cmd.Stderr = &errBuf

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}
	cmd.Stdout = &outBuf

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start binary: %v", err)
	}

	// Build the two JSON-RPC requests.
	// 1. initialize — required by the MCP protocol before any tool calls.
	initReq := testJSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "0.0.1",
			},
		},
	}

	// 2. tools/list — retrieve the registered tools.
	listReq := testJSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	// Write both requests as newline-delimited JSON to stdin.
	initBytes, err := json.Marshal(initReq)
	if err != nil {
		t.Fatalf("failed to marshal initialize request: %v", err)
	}
	listBytes, err := json.Marshal(listReq)
	if err != nil {
		t.Fatalf("failed to marshal tools/list request: %v", err)
	}

	if _, err := fmt.Fprintf(stdin, "%s\n%s\n", initBytes, listBytes); err != nil {
		t.Fatalf("failed to write requests to stdin: %v", err)
	}
	// Close stdin so the server knows there is no more input and can shut down.
	if err := stdin.Close(); err != nil {
		t.Fatalf("failed to close stdin: %v", err)
	}

	// Wait for the process to exit.
	if err := cmd.Wait(); err != nil {
		// A non-zero exit after stdin is closed is tolerable if we got output.
		// Log for diagnostic purposes but do not fail outright.
		t.Logf("binary exited with error (may be normal on stdin close): %v", err)
		if errBuf.Len() > 0 {
			t.Logf("binary stderr: %s", errBuf.String())
		}
	}

	// Parse every JSON-RPC response line from stdout.
	// We are looking for the response with id=2 (the tools/list response).
	var toolsListResponse *testJSONRPCResponse

	scanner := bufio.NewScanner(&outBuf)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var resp testJSONRPCResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			// Skip non-JSON lines (e.g., SDK diagnostic output).
			t.Logf("skipping non-JSON line: %s", line)
			continue
		}

		if resp.ID == 2 {
			toolsListResponse = &resp
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("error reading stdout: %v", err)
	}

	if toolsListResponse == nil {
		t.Fatalf("did not receive a tools/list response (id=2) from the binary.\nstdout:\n%s\nstderr:\n%s",
			outBuf.String(), errBuf.String())
	}

	if toolsListResponse.Error != nil {
		t.Fatalf("tools/list returned an error: %s", toolsListResponse.Error)
	}

	// Parse the result field into our typed struct.
	var result testToolsListResult
	if err := json.Unmarshal(toolsListResponse.Result, &result); err != nil {
		t.Fatalf("failed to unmarshal tools/list result: %v\nraw result: %s",
			err, toolsListResponse.Result)
	}

	return result
}
