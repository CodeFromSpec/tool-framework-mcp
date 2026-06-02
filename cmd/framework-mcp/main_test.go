// code-from-spec: ROOT/golang/tests/server@cTNrII6dHP7w7HDOkA2rV0theeY
package main_test

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
	binaryPath = filepath.Join(dir, name)

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
		t.Fatalf("expected exit 0, got error: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("expected usage message in stdout, got: %s", out)
	}
}

func TestHelpWord(t *testing.T) {
	cmd := exec.Command(binaryPath, "help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0, got error: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("expected usage message in stdout, got: %s", out)
	}
}

func TestShortHelpFlag(t *testing.T) {
	cmd := exec.Command(binaryPath, "-h")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0, got error: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("expected usage message in stdout, got: %s", out)
	}
}

func TestUnrecognizedArgument(t *testing.T) {
	cmd := exec.Command(binaryPath, "something")
	var stderr bytes.Buffer
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
	var stderr bytes.Buffer
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

func testSendMCPRequests(t *testing.T, requests []map[string]any) []map[string]any {
	t.Helper()
	cmd := exec.Command(binaryPath)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("failed to get stdin pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start binary: %v", err)
	}

	for _, req := range requests {
		data, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}
		fmt.Fprintf(stdin, "%s\n", data)
	}
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			t.Fatalf("binary exited with error: %v", err)
		}
	}

	var responses []map[string]any
	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		var resp map[string]any
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			continue
		}
		responses = append(responses, resp)
	}
	return responses
}

func TestToolsListMaxResultSizeChars(t *testing.T) {
	requests := []map[string]any{
		{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "initialize",
			"params": map[string]any{
				"protocolVersion": "2024-11-05",
				"clientInfo": map[string]any{
					"name":    "test-client",
					"version": "1.0",
				},
				"capabilities": map[string]any{},
			},
		},
		{
			"jsonrpc": "2.0",
			"id":      2,
			"method":  "tools/list",
			"params":  map[string]any{},
		},
	}

	responses := testSendMCPRequests(t, requests)

	var toolsListResp map[string]any
	for _, resp := range responses {
		id, _ := resp["id"].(float64)
		if id == 2 {
			toolsListResp = resp
			break
		}
	}
	if toolsListResp == nil {
		t.Fatal("did not receive tools/list response")
	}

	result, ok := toolsListResp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result object, got: %v", toolsListResp["result"])
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("expected tools array, got: %v", result["tools"])
	}

	for _, toolAny := range tools {
		tool, ok := toolAny.(map[string]any)
		if !ok {
			continue
		}
		name, _ := tool["name"].(string)
		if name != "load_chain" {
			continue
		}
		meta, ok := tool["_meta"].(map[string]any)
		if !ok {
			t.Fatalf("load_chain tool missing _meta field")
		}
		val, ok := meta["anthropic/maxResultSizeChars"]
		if !ok {
			t.Fatal("load_chain tool missing anthropic/maxResultSizeChars in _meta")
		}
		numVal, ok := val.(float64)
		if !ok || numVal != 500000 {
			t.Errorf("expected anthropic/maxResultSizeChars to be 500000, got: %v", val)
		}
		return
	}
	t.Fatal("load_chain tool not found in tools/list response")
}

func TestToolsListAdvertisesAllTools(t *testing.T) {
	requests := []map[string]any{
		{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "initialize",
			"params": map[string]any{
				"protocolVersion": "2024-11-05",
				"clientInfo": map[string]any{
					"name":    "test-client",
					"version": "1.0",
				},
				"capabilities": map[string]any{},
			},
		},
		{
			"jsonrpc": "2.0",
			"id":      2,
			"method":  "tools/list",
			"params":  map[string]any{},
		},
	}

	responses := testSendMCPRequests(t, requests)

	var toolsListResp map[string]any
	for _, resp := range responses {
		id, _ := resp["id"].(float64)
		if id == 2 {
			toolsListResp = resp
			break
		}
	}
	if toolsListResp == nil {
		t.Fatal("did not receive tools/list response")
	}

	result, ok := toolsListResp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result object, got: %v", toolsListResp["result"])
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("expected tools array, got: %v", result["tools"])
	}

	expected := []string{"load_chain", "write_file", "validate_specs", "chain_hash", "version"}
	found := make(map[string]bool)
	for _, toolAny := range tools {
		tool, ok := toolAny.(map[string]any)
		if !ok {
			continue
		}
		name, _ := tool["name"].(string)
		found[name] = true
	}

	for _, name := range expected {
		if !found[name] {
			t.Errorf("expected tool %q to be present in tools/list", name)
		}
	}
}
