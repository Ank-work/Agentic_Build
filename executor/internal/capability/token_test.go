package capability

import (
	"strings"
	"testing"
)

func TestReaderUnusableToken(t *testing.T) {
	if err := Validate("reader", "", "gromacs", nil); err != nil {
		t.Fatalf("empty reader token should pass: %v", err)
	}
	err := Validate("reader", "tok-abc", "gromacs", nil)
	if err == nil {
		t.Fatal("reader with token must reject")
	}
	if strings.Contains(err.Error(), "tok-abc") {
		t.Fatal("error must not log the token")
	}
}

func TestExecutorRequiresToken(t *testing.T) {
	if err := Validate("executor", "", "gromacs", nil); err == nil {
		t.Fatal("executor without token must reject")
	}
}

func TestExecutorOpaqueTokenCoversArgv0(t *testing.T) {
	if err := Validate("executor", "placeholder-capability-token", "gromacs", nil); err != nil {
		t.Fatalf("opaque token with no destinations: %v", err)
	}
}

func TestExecutorStructuredCoversDestinations(t *testing.T) {
	tok := `{"binaries":["gromacs"],"hosts":["example.com"]}`
	if err := Validate("executor", tok, "gromacs", []string{"example.com"}); err != nil {
		t.Fatalf("structured token: %v", err)
	}
	if err := Validate("executor", tok, "gromacs", []string{"evil.example"}); err == nil {
		t.Fatal("expected destination miss to reject")
	}
	if err := Validate("executor", tok, "not-gromacs", nil); err == nil {
		t.Fatal("expected argv0 miss to reject")
	}
}

func TestOpaqueTokenDoesNotCoverHosts(t *testing.T) {
	if err := Validate("executor", "opaque", "gromacs", []string{"example.com"}); err == nil {
		t.Fatal("opaque token must not cover destinations")
	}
}
