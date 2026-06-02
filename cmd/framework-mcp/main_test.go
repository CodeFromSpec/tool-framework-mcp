// code-from-spec: ROOT/golang/tests/server@GaEeScH0L46QRU10Zh81i6ixAEQ
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

var testBinaryPath string

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "framework-mcp-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	binName := "framework-mcp"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	testBinaryPath = filepath.Join(tmpDir, binName)

	cmd := exec.Command("go", "build", "-o", testBinaryPath, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build binary: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestHelp_Flag(t *testing.T) {
	cmd := exec.Command(testBinaryPath, "--help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("--help exited with error: %v", err)
	}
	if !strings.Contains(string(out), "Usage") {
		t.Errorf("stdout does not contain usage message, got: %s", out)
	}
}

func TestHelp_Word(t *testing.T) {
	cmd := exec.Command(testBinaryPath, "help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("help exited with error: %v", err)
	}
	if !strings.Contains(string(out), "Usage") {
		t.Errorf("stdout does not contain usage message, got: %s", out)
	}
}

func TestHelp_ShortFlag(t *testing.T) {
	cmd := exec.Command(testBinaryPath, "-h")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("-h exited with error: %v", err)
	}
	if !strings.Contains(string(out), "Usage") {
		t.Errorf("stdout does not contain usage message, got: %s", out)
	}
}

func TestUnrecognizedArgument(t *testing.T) {
	cmd := exec.Command(testBinaryPath, "something")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected exit 1 for unrecognized argument")
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
		}
	}
	if !strings.Contains(stderr.String(), "Usage") {
		t.Errorf("stderr does not contain usage message, got: %s", stderr.String())
	}
}

func TestMultipleArguments(t *testing.T) {
	cmd := exec.Command(testBinaryPath, "foo", "bar")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected exit 1 for multiple arguments")
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
		}
	}
	if !strings.Contains(stderr.String(), "Usage") {
		t.Errorf("stderr does not contain usage message, got: %s", stderr.String())
	}
}

type testJSONRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type testJSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   json.RawMessage `json:"error"`
}

func testReadResponse(t *testing.T, scanner *bufio.Scanner) testJSONRPCResponse {
	t.Helper()
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		var resp testJSONRPCResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			t.Fatalf("failed to parse response line %q: %v", line, err)
		}
		if resp.ID == nil {
			continue
		}
		return resp
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner error: %v", err)
	}
	t.Fatal("no response received from server")
	return testJSONRPCResponse{}
}

func testStartServer(t *testing.T) (stdin *os.File, scanner *bufio.Scanner, cmd *exec.Cmd, cleanup func()) {
	t.Helper()

	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating stdin pipe: %v", err)
	}
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating stdout pipe: %v", err)
	}

	cmd = exec.Command(testBinaryPath)
	cmd.Stdin = stdinR
	cmd.Stdout = stdoutW

	if err := cmd.Start(); err != nil {
		t.Fatalf("starting server: %v", err)
	}

	stdinR.Close()
	stdoutW.Close()

	scanner = bufio.NewScanner(stdoutR)

	cleanup = func() {
		stdinW.Close()
		stdoutR.Close()
		cmd.Wait()
	}

	return stdinW, scanner, cmd, cleanup
}

func testSendRequest(t *testing.T, w *os.File, req testJSONRPCRequest) {
	t.Helper()
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	data = append(data, '\n')
	if _, err := w.Write(data); err != nil {
		t.Fatalf("write request: %v", err)
	}
}

func TestMCP_ToolsList_AllTools(t *testing.T) {
	stdinW, scanner, _, cleanup := testStartServer(t)
	defer cleanup()

	initReq := testJSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2024-11-05",
			"clientInfo":      map[string]any{"name": "test", "version": "0"},
			"capabilities":    map[string]any{},
		},
	}
	testSendRequest(t, stdinW, initReq)
	testReadResponse(t, scanner)

	listReq := testJSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}
	testSendRequest(t, stdinW, listReq)
	resp := testReadResponse(t, scanner)

	var result struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal tools/list result: %v", err)
	}

	wantTools := []string{"load_chain", "write_file", "validate_specs", "chain_hash", "version"}
	for _, want := range wantTools {
		found := false
		for _, tool := range result.Tools {
			if tool.Name == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("tool %q not found in tools/list response", want)
		}
	}
}

func TestMCP_ToolsList_LoadChainMaxResultSize(t *testing.T) {
	stdinW, scanner, _, cleanup := testStartServer(t)
	defer cleanup()

	initReq := testJSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2024-11-05",
			"clientInfo":      map[string]any{"name": "test", "version": "0"},
			"capabilities":    map[string]any{},
		},
	}
	testSendRequest(t, stdinW, initReq)
	testReadResponse(t, scanner)

	listReq := testJSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}
	testSendRequest(t, stdinW, listReq)
	resp := testReadResponse(t, scanner)

	var result struct {
		Tools []struct {
			Name string          `json:"name"`
			Meta json.RawMessage `json:"_meta"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal tools/list result: %v", err)
	}

	for _, tool := range result.Tools {
		if tool.Name != "load_chain" {
			continue
		}
		var meta map[string]any
		if err := json.Unmarshal(tool.Meta, &meta); err != nil {
			t.Fatalf("unmarshal _meta for load_chain: %v", err)
		}
		val, ok := meta["anthropic/maxResultSizeChars"]
		if !ok {
			t.Fatal("load_chain _meta missing anthropic/maxResultSizeChars")
		}
		numVal, ok := val.(float64)
		if !ok {
			t.Fatalf("anthropic/maxResultSizeChars is not a number: %T", val)
		}
		if int(numVal) != 500000 {
			t.Errorf("anthropic/maxResultSizeChars = %v, want 500000", numVal)
		}
		return
	}
	t.Fatal("load_chain tool not found in tools/list response")
}
