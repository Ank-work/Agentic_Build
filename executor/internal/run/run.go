// Package run executes validated Aeon jobs as os/exec children (§6.1, §12.1–§12.5).
//
// Jobs are argv arrays only: exec.CommandContext(ctx, argv[0], argv[1:]...).
// Never a shell string, never sh -c, never OpenClaw exec/spawn/shell/fs_write.
//
// network.mode none (default, §6.1, §12.3) grants no destinations. Child env is
// exactly spec.env (explicit keys only; no parent inheritance, no shell
// expansion). Proxy and related network variables are not set for mode none.
// True Linux network-namespace isolation is best-effort, optional, and must not
// be required for tests (no root / CAP_SYS_ADMIN assumption). Portable
// enforcement is: no destinations granted and no proxy/network env on the child.
// A real per-process network jail is an AK/Windows concern, not a test
// requirement of this package.
//
// Capability tokens are never logged, printed, or included in Result or error
// strings (§12.5). stage: reader is refused execution and does not Start a
// process (§6.2).
package run

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/Ank-work/Agentic_Build/executor/internal/allowlist"
	"github.com/Ank-work/Agentic_Build/executor/internal/capability"
	"github.com/Ank-work/Agentic_Build/executor/internal/job"
)

// ErrReaderStage is returned when a reader job would otherwise be executed.
// The reader must not Start a process (§6.2).
var ErrReaderStage = errors.New("reader stage must not execute")

var proxyEnvKeys = map[string]struct{}{
	"HTTP_PROXY": {}, "HTTPS_PROXY": {}, "ALL_PROXY": {}, "NO_PROXY": {},
	"http_proxy": {}, "https_proxy": {}, "all_proxy": {}, "no_proxy": {},
	"FTP_PROXY": {}, "ftp_proxy": {},
}

// Result is the outcome of a run. It omits capabilityToken and must never
// serialize or fmt the raw token (§12.5).
type Result struct {
	ID          string `json:"id"`
	Stage       string `json:"stage"`
	NetworkMode string `json:"networkMode"`
	ExitCode    int    `json:"exitCode"`
	TimedOut    bool   `json:"timedOut"`
	Stdout      string `json:"stdout"`
	Stderr      string `json:"stderr"`
}

// String reports identity and exit fields only. It never includes a capability token.
func (r Result) String() string {
	return fmt.Sprintf("id=%s stage=%s network.mode=%s exit=%d timedOut=%v",
		r.ID, r.Stage, r.NetworkMode, r.ExitCode, r.TimedOut)
}

// MarshalJSON encodes Result without a capabilityToken field.
func (r Result) MarshalJSON() ([]byte, error) {
	type out Result
	return json.Marshal(out(r))
}

// Run allowlist-checks and capability-validates j, then either refuses a
// reader job or runs argv as an os/exec child with timeout. The raw
// capability token is never placed in errors or Result.
func Run(ctx context.Context, j *job.Job, allowed allowlist.Set) (*Result, error) {
	if j == nil {
		return nil, fmt.Errorf("job is nil")
	}
	if err := allowlist.Check([]string(j.Spec.Argv), allowed); err != nil {
		return nil, err
	}
	if err := capability.Validate(j.Spec.Stage, j.Spec.CapabilityToken, j.Spec.Argv[0], j.DestinationHosts()); err != nil {
		return nil, err
	}
	if j.Spec.Stage == job.StageReader {
		return nil, ErrReaderStage
	}

	timeout := time.Duration(j.Spec.TimeoutSeconds) * time.Second
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	argv := j.Spec.Argv
	cmd := exec.CommandContext(runCtx, argv[0], argv[1:]...)
	if j.Spec.Cwd != "" {
		cmd.Dir = j.Spec.Cwd
	}
	cmd.Env = childEnv(j.Spec.Env, j.NetworkMode())

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	res := &Result{
		ID:          j.Metadata.ID,
		Stage:       j.Spec.Stage,
		NetworkMode: j.NetworkMode(),
		Stdout:      stdout.String(),
		Stderr:      stderr.String(),
	}
	if runCtx.Err() == context.DeadlineExceeded {
		res.TimedOut = true
	}
	if cmd.ProcessState != nil {
		res.ExitCode = cmd.ProcessState.ExitCode()
	} else if err != nil {
		res.ExitCode = -1
	}
	if res.TimedOut && res.ExitCode == 0 {
		res.ExitCode = -1
	}
	_ = err
	return res, nil
}

func childEnv(specEnv map[string]string, networkMode string) []string {
	// Non-nil empty slice: do not inherit the parent environment.
	out := make([]string, 0, len(specEnv))
	for k, v := range specEnv {
		if networkMode == job.NetworkNone {
			if _, deny := proxyEnvKeys[k]; deny {
				continue
			}
		}
		if strings.ContainsAny(k, "=\x00") {
			continue
		}
		out = append(out, k+"="+v)
	}
	return out
}
