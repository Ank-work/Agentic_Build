package gitqueue

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessInboxFallbackStub(t *testing.T) {
	dir := t.TempDir()
	inbox := filepath.Join(dir, "inbox")
	outbox := filepath.Join(dir, "outbox")
	if err := os.Mkdir(inbox, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{
		"apiVersion":"aeon.dev/v1","kind":"Job",
		"metadata":{"id":"job-q1","createdAt":"2026-09-03T00:00:00Z"},
		"spec":{"stage":"executor","argv":["gromacs"],"timeoutSeconds":60,
			"network":{"mode":"none"},"capabilityToken":"inbox-secret-token"}
	}`
	if err := os.WriteFile(filepath.Join(inbox, "job-q1.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ProcessInbox(inbox, outbox); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(outbox, "job-q1.json"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	if strings.Contains(s, "inbox-secret-token") {
		t.Fatal("outbox must not copy capabilityToken")
	}
	if !strings.Contains(s, "accepted-stub") {
		t.Fatalf("unexpected outbox: %s", s)
	}
	if !strings.Contains(s, "fallback") {
		t.Fatalf("missing fallback note: %s", s)
	}
}

func TestUnimplementedGit(t *testing.T) {
	var g UnimplementedGit
	if err := g.Pull(nil); err != ErrNotImplemented {
		t.Fatalf("Pull: %v", err)
	}
	if err := g.Commit(nil, nil, "m"); err != ErrNotImplemented {
		t.Fatalf("Commit: %v", err)
	}
}
