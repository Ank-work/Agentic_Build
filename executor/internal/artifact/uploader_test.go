package artifact

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestStubUploaderNoSecretsAndNotImplemented(t *testing.T) {
	u := StubUploader{}
	uri, err := u.Upload(context.Background(), "/tmp/out.tgz", "aeon-artifacts", "jobs/abc")
	if !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("want ErrNotImplemented, got %v", err)
	}
	if !strings.HasPrefix(uri, "r2://") {
		t.Fatalf("uri %q should be r2:// placeholder", uri)
	}
}
