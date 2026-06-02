// code-from-spec: ROOT/golang/tests/server@o6-TqRHzbmGI_pipuPKick-yYCA
package main_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "framework-mcp-test-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create temp dir:", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	name := "framework-mcp"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binaryPath = dir + "/" + name

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to build binary:", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestHelpFlag(t *testing.T) {
	cmd := exec.Command(binaryPath, "--help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0, got: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("expected usage message in stdout, got: %s", string(out))
	}
}

func TestHelpWord(t *testing.T) {
	cmd := exec.Command(binaryPath, "help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0, got: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("expected usage message in stdout, got: %s", string(out))
	}
}

func TestShortHelpFlag(t *testing.T) {
	cmd := exec.Command(binaryPath, "-h")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0, got: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("expected usage message in stdout, got: %s", string(out))
	}
}

func TestUnrecognizedArgument(t *testing.T) {
	cmd := exec.Command(binaryPath, "something")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected exit 1, got exit 0")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 1 {
		t.Fatalf("expected exit code 1, got: %v", err)
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Errorf("expected usage message in stderr, got: %s", stderr.String())
	}
}

func TestMultipleArguments(t *testing.T) {
	cmd := exec.Command(binaryPath, "foo", "bar")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected exit 1, got exit 0")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 1 {
		t.Fatalf("expected exit code 1, got: %v", err)
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Errorf("expected usage message in stderr, got: %s", stderr.String())
	}
}

type jsonrpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type jsonrpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   json.RawMessage `json:"error"`
}

func testSendRequest(t *testing.T, enc *json.Encoder, id int, method string, params any) {
	t.Helper()
	req := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
	if err := enc.Encode(req); err != nil {
		t.Fatalf("failed to send request %s: %v", method, err)
	}
}

func testReadResponse(t *testing.T, scanner *bufio.Scanner, id int) jsonrpcResponse {
	t.Helper()
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		var resp jsonrpcResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			t.Fatalf("failed to parse response line: %v\nline: %s", err, line)
		}
		if resp.ID == nil {
			continue
		}
		if *resp.ID == id {
			return resp
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner error: %v", err)
	}
	t.Fatalf("no response found for id %d", id)
	return jsonrpcResponse{}
}

func testStartMCPSession(t *testing.T) (*json.Encoder, *bufio.Scanner, func()) {
	t.Helper()

	cmd := exec.Command(binaryPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("failed to get stdin pipe: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("failed to get stdout pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start binary: %v", err)
	}

	enc := json.NewEncoder(stdin)
	scanner := bufio.NewScanner(stdout)

	testSendRequest(t, enc, 1, "initialize", map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "test", "version": "0.0.1"},
	})
	testReadResponse(t, scanner, 1)

	cleanup := func() {
		stdin.Close()
		cmd.Wait()
	}

	return enc, scanner, cleanup
}

func TestToolsListAdvertisesAllTools(t *testing.T) {
	enc, scanner, cleanup := testStartMCPSession(t)
	defer cleanup()

	testSendRequest(t, enc, 2, "tools/list", nil)
	resp := testReadResponse(t, scanner, 2)

	var result struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("failed to parse tools/list result: %v", err)
	}

	wantTools := []string{"load_chain", "write_file", "validate_specs", "version"}
	found := make(map[string]bool)
	for _, tool := range result.Tools {
		found[tool.Name] = true
	}
	for _, name := range wantTools {
		if !found[name] {
			t.Errorf("expected tool %q in tools/list response", name)
		}
	}
}

func TestToolsListAdvertisesMaxResultSizeCharsForLoadChain(t *testing.T) {
	enc, scanner, cleanup := testStartMCPSession(t)
	defer cleanup()

	testSendRequest(t, enc, 2, "tools/list", nil)
	resp := testReadResponse(t, scanner, 2)

	var result struct {
		Tools []struct {
			Name string          `json:"name"`
			Meta json.RawMessage `json:"_meta"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("failed to parse tools/list result: %v", err)
	}

	for _, tool := range result.Tools {
		if tool.Name != "load_chain" {
			continue
		}
		if tool.Meta == nil {
			t.Fatal("load_chain tool has no _meta field")
		}
		var meta map[string]any
		if err := json.Unmarshal(tool.Meta, &meta); err != nil {
			t.Fatalf("failed to parse _meta: %v", err)
		}
		val, ok := meta["anthropic/maxResultSizeChars"]
		if !ok {
			t.Fatal("load_chain _meta missing anthropic/maxResultSizeChars")
		}
		numVal, ok := val.(float64)
		if !ok {
			t.Fatalf("anthropic/maxResultSizeChars has unexpected type %T", val)
		}
		if int(numVal) != 500000 {
			t.Errorf("expected anthropic/maxResultSizeChars=500000, got %v", numVal)
		}
		return
	}
	t.Fatal("load_chain tool not found in tools/list response")
}
