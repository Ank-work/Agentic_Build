package job

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Ank-work/Agentic_Build/executor/internal/allowlist"
	"github.com/Ank-work/Agentic_Build/executor/internal/capability"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	p := filepath.Join("..", "..", "..", "schemas", "fixtures", name)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return b
}

func TestParseValidMinimal(t *testing.T) {
	j, err := Parse(fixture(t, "valid-minimal.json"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if j.NetworkMode() != NetworkNone {
		t.Fatalf("network.mode=%q want none", j.NetworkMode())
	}
	if len(j.Spec.Argv) == 0 {
		t.Fatal("argv empty")
	}
	if err := allowlist.Check([]string(j.Spec.Argv), allowlist.Default()); err != nil {
		t.Fatalf("allowlist: %v", err)
	}
	if err := capability.Validate(j.Spec.Stage, j.Spec.CapabilityToken, j.Spec.Argv[0], j.DestinationHosts()); err != nil {
		t.Fatalf("capability: %v", err)
	}
}

func TestShellStringReject(t *testing.T) {
	_, err := Parse(fixture(t, "invalid-shell-string.json"))
	if err == nil {
		t.Fatal("expected reject of shell-string argv")
	}
	if !strings.Contains(err.Error(), "array") && !strings.Contains(err.Error(), "shell") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCommandShellFieldsReject(t *testing.T) {
	_, err := Parse(fixture(t, "invalid-shell-command-field.json"))
	if err == nil {
		t.Fatal("expected reject of unknown command/shell fields")
	}
}

func TestMissingNetworkDefaultsToNone(t *testing.T) {
	j, err := Parse(fixture(t, "missing-network.json"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if j.NetworkMode() != NetworkNone {
		t.Fatalf("network.mode=%q want none", j.NetworkMode())
	}
}

func TestReaderWithTokenReject(t *testing.T) {
	raw := []byte(`{
		"apiVersion":"aeon.dev/v1","kind":"Job",
		"metadata":{"id":"r1","createdAt":"2026-09-03T00:00:00Z"},
		"spec":{"stage":"reader","argv":["gromacs"],"timeoutSeconds":60,
			"network":{"mode":"none"},"capabilityToken":"should-not-be-here"}
	}`)
	_, err := Parse(raw)
	if err == nil {
		t.Fatal("expected reader+token reject")
	}
	if strings.Contains(err.Error(), "should-not-be-here") {
		t.Fatal("error must not log the token")
	}
}

func TestExecutorNoTokenReject(t *testing.T) {
	raw := []byte(`{
		"apiVersion":"aeon.dev/v1","kind":"Job",
		"metadata":{"id":"e1","createdAt":"2026-09-03T00:00:00Z"},
		"spec":{"stage":"executor","argv":["gromacs"],"timeoutSeconds":60,
			"network":{"mode":"none"}}
	}`)
	_, err := Parse(raw)
	if err == nil {
		t.Fatal("expected executor without token to reject")
	}
}

func TestAllowlistModeWithoutDestinationsReject(t *testing.T) {
	raw := []byte(`{
		"apiVersion":"aeon.dev/v1","kind":"Job",
		"metadata":{"id":"e2","createdAt":"2026-09-03T00:00:00Z"},
		"spec":{"stage":"executor","argv":["gromacs"],"timeoutSeconds":60,
			"network":{"mode":"allowlist"},"capabilityToken":"t"}
	}`)
	_, err := Parse(raw)
	if err == nil {
		t.Fatal("expected allowlist mode without destinations to reject")
	}
}

func TestUnknownKeyReject(t *testing.T) {
	raw := []byte(`{
		"apiVersion":"aeon.dev/v1","kind":"Job",
		"metadata":{"id":"e3","createdAt":"2026-09-03T00:00:00Z"},
		"spec":{"stage":"executor","argv":["gromacs"],"timeoutSeconds":60,
			"network":{"mode":"none"},"capabilityToken":"t","extra":true}
	}`)
	_, err := Parse(raw)
	if err == nil {
		t.Fatal("expected unknown key reject")
	}
}

func TestExecutorTokenAllowlistedArgv0Accept(t *testing.T) {
	j, err := Parse(fixture(t, "valid-minimal.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := allowlist.Check([]string(j.Spec.Argv), allowlist.Default()); err != nil {
		t.Fatal(err)
	}
	if err := capability.Validate(j.Spec.Stage, j.Spec.CapabilityToken, j.Spec.Argv[0], nil); err != nil {
		t.Fatal(err)
	}
}
