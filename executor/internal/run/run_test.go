package run

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Ank-work/Agentic_Build/executor/internal/allowlist"
	"github.com/Ank-work/Agentic_Build/executor/internal/job"
)

const secretToken = "tok-chunk-a-secret"

func tempAllowlist(bins ...string) allowlist.Set {
	s := make(allowlist.Set)
	for _, b := range bins {
		s[b] = struct{}{}
	}
	return s
}

func parseJob(t *testing.T, raw string) *job.Job {
	t.Helper()
	j, err := job.Parse([]byte(raw))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return j
}

func TestRunTable(t *testing.T) {
	allowed := tempAllowlist("/bin/true", "/bin/echo", "/bin/sleep")
	ctx := context.Background()

	t.Run("executor echo hello succeeds", func(t *testing.T) {
		j := parseJob(t, `{
			"apiVersion":"aeon.dev/v1","kind":"Job",
			"metadata":{"id":"echo-ok","createdAt":"2026-09-03T00:00:00Z"},
			"spec":{"stage":"executor","argv":["/bin/echo","hello"],
				"timeoutSeconds":5,"network":{"mode":"none"},
				"capabilityToken":"`+secretToken+`"}
		}`)
		res, err := Run(ctx, j, allowed)
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if res.ExitCode != 0 {
			t.Fatalf("exit=%d stdout=%q stderr=%q", res.ExitCode, res.Stdout, res.Stderr)
		}
		if !strings.Contains(res.Stdout, "hello") {
			t.Fatalf("stdout=%q want hello", res.Stdout)
		}
		if res.Stage != job.StageExecutor {
			t.Fatalf("stage=%q", res.Stage)
		}
		assertNoToken(t, res, err)
	})

	t.Run("executor true succeeds", func(t *testing.T) {
		j := parseJob(t, `{
			"apiVersion":"aeon.dev/v1","kind":"Job",
			"metadata":{"id":"true-ok","createdAt":"2026-09-03T00:00:00Z"},
			"spec":{"stage":"executor","argv":["/bin/true"],
				"timeoutSeconds":5,"network":{"mode":"none"},
				"capabilityToken":"`+secretToken+`"}
		}`)
		res, err := Run(ctx, j, allowed)
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if res.ExitCode != 0 {
			t.Fatalf("exit=%d", res.ExitCode)
		}
		assertNoToken(t, res, err)
	})

	t.Run("reader refused with no process", func(t *testing.T) {
		j := parseJob(t, `{
			"apiVersion":"aeon.dev/v1","kind":"Job",
			"metadata":{"id":"reader-no","createdAt":"2026-09-03T00:00:00Z"},
			"spec":{"stage":"reader","argv":["/bin/echo","READER-SHOULD-NOT-RUN"],
				"timeoutSeconds":5,"network":{"mode":"none"}}
		}`)
		if j.Spec.CapabilityToken != "" {
			t.Fatal("reader fixture must have empty token")
		}
		res, err := Run(ctx, j, allowed)
		if !errors.Is(err, ErrReaderStage) {
			t.Fatalf("err=%v want ErrReaderStage", err)
		}
		if res != nil {
			t.Fatalf("result=%v; reader must not start a process", res)
		}
		if err != nil && strings.Contains(err.Error(), "READER-SHOULD-NOT-RUN") {
			t.Fatal("reader error must not include child output")
		}
		assertNoToken(t, res, err)
	})

	t.Run("timeout sleep", func(t *testing.T) {
		j := parseJob(t, `{
			"apiVersion":"aeon.dev/v1","kind":"Job",
			"metadata":{"id":"sleep-to","createdAt":"2026-09-03T00:00:00Z"},
			"spec":{"stage":"executor","argv":["/bin/sleep","30"],
				"timeoutSeconds":1,"network":{"mode":"none"},
				"capabilityToken":"`+secretToken+`"}
		}`)
		start := time.Now()
		res, err := Run(ctx, j, allowed)
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if !res.TimedOut {
			t.Fatal("expected TimedOut")
		}
		if res.ExitCode == 0 {
			t.Fatal("timeout must be non-zero exit")
		}
		if elapsed > 4*time.Second {
			t.Fatalf("timeout too slow: %s", elapsed)
		}
		assertNoToken(t, res, err)
	})

	t.Run("network mode none recorded no destinations", func(t *testing.T) {
		j := parseJob(t, `{
			"apiVersion":"aeon.dev/v1","kind":"Job",
			"metadata":{"id":"net-none","createdAt":"2026-09-03T00:00:00Z"},
			"spec":{"stage":"executor","argv":["/bin/true"],
				"timeoutSeconds":5,"network":{"mode":"none"},
				"capabilityToken":"`+secretToken+`",
				"env":{"HTTP_PROXY":"http://127.0.0.1:9","FOO":"bar"}}
		}`)
		if j.NetworkMode() != job.NetworkNone {
			t.Fatalf("job mode=%q", j.NetworkMode())
		}
		if len(j.DestinationHosts()) != 0 {
			t.Fatalf("destinations granted=%v", j.DestinationHosts())
		}
		res, err := Run(ctx, j, allowed)
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if res.NetworkMode != job.NetworkNone {
			t.Fatalf("result mode=%q want none", res.NetworkMode)
		}
		if len(j.DestinationHosts()) != 0 {
			t.Fatalf("mode none must grant no destinations: %v", j.DestinationHosts())
		}
		env := childEnv(j.Spec.Env, j.NetworkMode())
		for _, e := range env {
			if strings.HasPrefix(e, "HTTP_PROXY=") {
				t.Fatal("mode none must not set proxy env")
			}
		}
		assertNoToken(t, res, err)
	})

	t.Run("allowlist miss fails before exec", func(t *testing.T) {
		j := parseJob(t, `{
			"apiVersion":"aeon.dev/v1","kind":"Job",
			"metadata":{"id":"allow-miss","createdAt":"2026-09-03T00:00:00Z"},
			"spec":{"stage":"executor","argv":["/bin/echo","ALLOWLIST-SHOULD-NOT-RUN"],
				"timeoutSeconds":5,"network":{"mode":"none"},
				"capabilityToken":"`+secretToken+`"}
		}`)
		res, err := Run(ctx, j, tempAllowlist("gromacs"))
		if err == nil {
			t.Fatal("expected allowlist miss")
		}
		if res != nil {
			t.Fatalf("must not exec on allowlist miss: %v", res)
		}
		if !strings.Contains(err.Error(), "allowlist") {
			t.Fatalf("err=%v", err)
		}
		assertNoToken(t, res, err)
	})

	t.Run("missing executor token fails before exec", func(t *testing.T) {
		j := &job.Job{
			Metadata: job.Metadata{ID: "no-token"},
			Spec: job.Spec{
				Stage:          job.StageExecutor,
				Argv:           job.Argv{"/bin/echo", "TOKEN-MISS-SHOULD-NOT-RUN"},
				TimeoutSeconds: 5,
				Network:        &job.Network{Mode: job.NetworkNone},
			},
		}
		res, err := Run(ctx, j, allowed)
		if err == nil {
			t.Fatal("expected missing token reject")
		}
		if res != nil {
			t.Fatalf("must not exec without token: %v", res)
		}
		assertNoToken(t, res, err)
	})
}

func assertNoToken(t *testing.T, res *Result, err error) {
	t.Helper()
	blob := fmt.Sprintf("err=%v res=%v", err, res)
	if res != nil {
		blob += res.String()
		b, mErr := json.Marshal(res)
		if mErr != nil {
			t.Fatalf("marshal: %v", mErr)
		}
		blob += string(b)
	}
	if strings.Contains(blob, secretToken) {
		t.Fatal("capability token must not appear in error or result fmt")
	}
}
