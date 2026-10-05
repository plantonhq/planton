package pulumibinary

import (
	"strings"
	"testing"

	"github.com/plantonhq/planton/pkg/downloads"
)

// Pins the wiring, not the shape (pkg/downloads owns the shape): this package
// must compose exactly the key the release lanes upload, for a kind the
// registry knows.
func TestBuildDownloadURLMatchesDownloadsGrammar(t *testing.T) {
	const kindDir = "AwsS3Bucket"
	const release = "v0.3.50"

	got, err := BuildDownloadURL(kindDir, release)
	if err != nil {
		t.Fatalf("BuildDownloadURL() error: %v", err)
	}

	want := downloads.BuildPulumiDownloadURL(kindDir, release, GetPlatformSuffix())
	if got != want {
		t.Errorf("BuildDownloadURL() = %q, want %q", got, want)
	}
}

func TestBuildDownloadURLUnknownKindFailsPlainly(t *testing.T) {
	_, err := BuildDownloadURL("NoSuchKind", "v0.3.50")
	if err == nil {
		t.Fatal("BuildDownloadURL() with an unknown kind must fail instead of composing a URL that 404s")
	}
	if !strings.Contains(err.Error(), "NoSuchKind") {
		t.Errorf("error %q does not name the unresolvable kind", err.Error())
	}
}
