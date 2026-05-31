// code-from-spec: ROOT/golang/tests/server@mKB3EKrkCTFNN5c9LeEDdV6z_YY
package main_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

var testBinaryPath string

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "framework-mcp-test-*")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}
	defer os.RemoveAll(tmpDir)

	binaryName := "framework-mcp"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	testBinaryPath = filepath.Join(tmpDir, binaryName)

	buildCmd := exec.Command("go", "build", "-o", testBinaryPath, ".")
	buildCmd.Stdout = os.Stderr
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		panic("failed to build binary: " + err.Error())
	}

	os.Exit(m.Run())
}

// testRunBinary runs the compiled binary with the given arguments and returns
// stdout, stderr, and the exit code.
func testRunBinary(args ...string) (stdout, stderr string, exitCode int) {
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
			exitCode = -1
		}
	}
	return stdout, stderr, exitCode
}

// testMCPSession holds the state needed to communicate with a running MCP process.
type testMCPSession struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	decoder *json.Decoder
}

// testStartMCPProcess starts the binary as a subprocess for MCP communication.
func testStartMCPProcess(t *testing.T) *testMCPSession {
	t.Helper()
	cmd := exec.Command(testBinaryPath)

	in, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("testStartMCPProcess: stdin pipe: %v", err)
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("testStartMCPProcess: stdout pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("testStartMCPProcess: start: %v", err)
	}
	t.Cleanup(func() {
		in.Close()
		cmd.Wait() //nolint:errcheck
	})

	return &testMCPSession{
		cmd:     cmd,
		stdin:   in,
		decoder: json.NewDecoder(out),
	}
}

// testSendJSON sends a JSON-encoded value followed by a newline to the session's stdin.
func testSendJSON(t *testing.T, s *testMCPSession, v any) {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("testSendJSON: marshal: %v", err)
	}
	data = append(data, '\n')
	if _, err := s.stdin.Write(data); err != nil {
		t.Fatalf("testSendJSON: write: %v", err)
	}
}

// testDecodeResponse reads one JSON-RPC response from the session with a 10-second timeout.
func testDecodeResponse(t *testing.T, s *testMCPSession) map[string]any {
	t.Helper()
	type decodeResult struct {
		val map[string]any
		err error
	}
	ch := make(chan decodeResult, 1)
	go func() {
		var v map[string]any
		err := s.decoder.Decode(&v)
		ch <- decodeResult{v, err}
	}()
	select {
	case r := <-ch:
		if r.err != nil {
			t.Fatalf("testDecodeResponse: %v", r.err)
		}
		return r.val
	case <-time.After(10 * time.Second):
		t.Fatalf("testDecodeResponse: timed out waiting for response")
		return nil
	}
}

// testMCPInitialize performs the MCP handshake: sends initialize, reads response,
// then sends the initialized notification.
func testMCPInitialize(t *testing.T, s *testMCPSession) {
	t.Helper()

	testSendJSON(t, s, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2024-11-05",
			"clientInfo": map[string]any{
				"name":    "test-client",
				"version": "0.0.1",
			},
			"capabilities": map[string]any{},
		},
	})

	// Read and discard the initialize response.
	testDecodeResponse(t, s)

	// Send initialized notification (no response expected).
	testSendJSON(t, s, map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	})
}

// testMCPToolsList sends a tools/list request and returns the parsed JSON-RPC response.
func testMCPToolsList(t *testing.T, s *testMCPSession) map[string]any {
	t.Helper()
	testSendJSON(t, s, map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]any{},
	})
	return testDecodeResponse(t, s)
}

// testExtractTools extracts the tools slice from a tools/list JSON-RPC response.
func testExtractTools(t *testing.T, resp map[string]any) []map[string]any {
	t.Helper()
	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("testExtractTools: response missing 'result' field: %v", resp)
	}
	rawTools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("testExtractTools: result missing 'tools' field: %v", result)
	}
	tools := make([]map[string]any, 0, len(rawTools))
	for i, raw := range rawTools {
		tool, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("testExtractTools: tool[%d] is not an object: %v", i, raw)
		}
		tools = append(tools, tool)
	}
	return tools
}

// --- Help flag tests ---

func TestHelpFlag_PrintsUsageToStdout(t *testing.T) {
	stdout, _, exitCode := testRunBinary("--help")
	if exitCode != 0 {
		t.Errorf("expected exit 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("expected stdout to contain usage message, got: %q", stdout)
	}
}

func TestHelpWord_PrintsUsageToStdout(t *testing.T) {
	stdout, _, exitCode := testRunBinary("help")
	if exitCode != 0 {
		t.Errorf("expected exit 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("expected stdout to contain usage message, got: %q", stdout)
	}
}

func TestShortHelpFlag_PrintsUsageToStdout(t *testing.T) {
	stdout, _, exitCode := testRunBinary("-h")
	if exitCode != 0 {
		t.Errorf("expected exit 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("expected stdout to contain usage message, got: %q", stdout)
	}
}

// --- Failure cases ---

func TestUnrecognizedArgument_PrintsUsageToStderr(t *testing.T) {
	_, stderr, exitCode := testRunBinary("something")
	if exitCode != 1 {
		t.Errorf("expected exit 1, got %d", exitCode)
	}
	if !strings.Contains(stderr, "Usage:") {
		t.Errorf("expected stderr to contain usage message, got: %q", stderr)
	}
}

func TestMultipleArguments_PrintsUsageToStderr(t *testing.T) {
	_, stderr, exitCode := testRunBinary("foo", "bar")
	if exitCode != 1 {
		t.Errorf("expected exit 1, got %d", exitCode)
	}
	if !strings.Contains(stderr, "Usage:") {
		t.Errorf("expected stderr to contain usage message, got: %q", stderr)
	}
}

// --- MCP protocol tests ---

func TestMCPToolsList_LoadChainHasMaxResultSizeChars(t *testing.T) {
	s := testStartMCPProcess(t)
	testMCPInitialize(t, s)
	resp := testMCPToolsList(t, s)

	tools := testExtractTools(t, resp)
	for _, tool := range tools {
		name, _ := tool["name"].(string)
		if name != "load_chain" {
			continue
		}
		meta, ok := tool["_meta"].(map[string]any)
		if !ok {
			t.Fatalf("load_chain tool has no _meta field")
		}
		val, ok := meta["anthropic/maxResultSizeChars"]
		if !ok {
			t.Fatalf("load_chain _meta missing anthropic/maxResultSizeChars")
		}
		// JSON numbers decode as float64.
		numVal, ok := val.(float64)
		if !ok {
			t.Fatalf("anthropic/maxResultSizeChars is not a number, got %T: %v", val, val)
		}
		if int(numVal) != 500000 {
			t.Errorf("expected anthropic/maxResultSizeChars = 500000, got %v", numVal)
		}
		return
	}
	t.Errorf("load_chain tool not found in tools/list response")
}

func TestMCPToolsList_AdvertisesAllFourTools(t *testing.T) {
	s := testStartMCPProcess(t)
	testMCPInitialize(t, s)
	resp := testMCPToolsList(t, s)

	tools := testExtractTools(t, resp)

	wantTools := []string{"load_chain", "write_file", "validate_specs", "hash_fragment"}
	found := make(map[string]bool, len(tools))
	for _, tool := range tools {
		name, _ := tool["name"].(string)
		found[name] = true
	}
	for _, want := range wantTools {
		if !found[want] {
			t.Errorf("expected tool %q to be advertised, but it was not found", want)
		}
	}
}
