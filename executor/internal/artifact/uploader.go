package artifact

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// ErrNotImplemented is returned by StubUploader. This slice does not
// talk to object storage.
var ErrNotImplemented = errors.New("artifact upload not implemented")

// Uploader sends science artifacts PC → Cloudflare R2 direct (§6.1, §7).
// Do not hairpin through Cursor's artifact bucket.
//
// Credentials belong in Doppler / the AK environment — never in this
// repository. Interface only; no SDK and no keys in this tree.
type Uploader interface {
	Upload(ctx context.Context, localPath, bucket, prefix string) (uri string, err error)
}

// StubUploader is a placeholder. It never contacts the network.
type StubUploader struct{}

// Upload returns an r2:// URI placeholder and ErrNotImplemented.
func (StubUploader) Upload(ctx context.Context, localPath, bucket, prefix string) (string, error) {
	_ = ctx
	base := filepath.Base(localPath)
	prefix = strings.Trim(prefix, "/")
	bucket = strings.TrimSpace(bucket)
	var uri string
	switch {
	case bucket == "":
		uri = "r2://"
	case prefix == "":
		uri = fmt.Sprintf("r2://%s/%s", bucket, base)
	default:
		uri = fmt.Sprintf("r2://%s/%s/%s", bucket, prefix, base)
	}
	return uri, ErrNotImplemented
}
