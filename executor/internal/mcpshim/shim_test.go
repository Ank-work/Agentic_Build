package mcpshim

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestListToolsOnlySurface(t *testing.T) {
	tools := ListTools()
	if len(tools) != 3 {
		t.Fatalf("got %d tools", len(tools))
	}
	names := map[string]bool{}
	for _, tool := range tools {
		names[tool.Name] = true
	}
	for _, want := range []string{ToolApprovals, ToolTelegram, ToolMemory} {
		if !names[want] {
			t.Errorf("missing %s", want)
		}
	}
	for _, deny := range []string{"exec", "spawn", "shell", "fs_write", "science-run"} {
		if names[deny] {
			t.Errorf("must not advertise %s", deny)
		}
	}
}

func TestDeniedTools(t *testing.T) {
	for _, name := range []string{"exec", "spawn", "shell", "fs_write", "science-run"} {
		res := CallTool(name)
		if !res.Denied || res.Message != "deny" {
			t.Errorf("%s: %+v", name, res)
		}
	}
}

func TestHandleToolsCallDeny(t *testing.T) {
	line := []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"exec"}}`)
	out, err := HandleLine(line)
	if err != nil {
		t.Fatal(err)
	}
	var resp map[string]any
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(resp["result"])
	if !strings.Contains(string(raw), "deny") {
		t.Fatalf("expected deny in %s", raw)
	}
}
