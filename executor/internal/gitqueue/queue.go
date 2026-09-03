package gitqueue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Ank-work/Agentic_Build/executor/internal/job"
)

// ErrNotImplemented is returned by git pull/commit stubs.
var ErrNotImplemented = errors.New("git-mediated queue git ops not implemented")

// This package is the §6.3 fallback only. Happy path is Cursor worker →
// job on AK → results return. Do not treat git as the science executor.

// GitOps is the pull/commit surface. Real git is not wired in this slice.
type GitOps interface {
	Pull(ctx context.Context) error
	Commit(ctx context.Context, paths []string, message string) error
}

// UnimplementedGit returns ErrNotImplemented for pull and commit.
type UnimplementedGit struct{}

func (UnimplementedGit) Pull(ctx context.Context) error {
	_ = ctx
	return ErrNotImplemented
}

func (UnimplementedGit) Commit(ctx context.Context, paths []string, message string) error {
	_ = ctx
	_ = paths
	_ = message
	return ErrNotImplemented
}

// Result is a stub reply written to jobs/outbox/<id>.json.
type Result struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Note   string `json:"note"`
}

const fallbackNote = "git-mediated queue is fallback only (§6.3); not the happy path"

// ProcessInbox reads jobs/inbox/*.json, Parse+Validate, and writes a stub
// reply to jobs/outbox/<id>.json. Does not run binaries. Capability tokens
// from the inbox document are never copied to the outbox.
func ProcessInbox(inboxDir, outboxDir string) error {
	entries, err := os.ReadDir(inboxDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(outboxDir, 0o755); err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		path := filepath.Join(inboxDir, e.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		res := Result{Note: fallbackNote}
		j, perr := job.Parse(raw)
		if perr != nil {
			res.ID = strings.TrimSuffix(e.Name(), ".json")
			res.Status = "rejected"
			res.Error = perr.Error()
		} else {
			res.ID = j.Metadata.ID
			res.Status = "accepted-stub"
		}
		out, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			return err
		}
		id := res.ID
		if id == "" {
			id = strings.TrimSuffix(e.Name(), ".json")
		}
		if err := os.WriteFile(filepath.Join(outboxDir, id+".json"), append(out, '\n'), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// OutboxDir returns the sibling outbox for an inbox path (.../jobs/inbox → .../jobs/outbox).
func OutboxDir(inboxDir string) string {
	return filepath.Join(filepath.Dir(inboxDir), "outbox")
}

// String reports the fallback-only contract (for logs that must not dump jobs).
func String() string {
	return fmt.Sprintf("gitqueue fallback watcher; %s", fallbackNote)
}
