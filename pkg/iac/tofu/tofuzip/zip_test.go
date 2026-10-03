package tofuzip

import (
	"strings"
	"testing"

	"github.com/plantonhq/planton/internal/cli/version"
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

	want := downloads.BuildTerraformDownloadURL(kindDir, release)
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

// Dev builds have no released artifacts, so the zip fast path must stay off
// and resolution must fall through to staging.
func TestCanUseZipModeIsOffForDevBuilds(t *testing.T) {
	saved := version.Version
	defer func() { version.Version = saved }()

	version.Version = ""
	if CanUseZipMode() {
		t.Error("CanUseZipMode() must be false for an unset version")
	}
	version.Version = version.DefaultVersion
	if CanUseZipMode() {
		t.Error("CanUseZipMode() must be false for the dev default version")
	}
	version.Version = "v0.3.50"
	if !CanUseZipMode() {
		t.Error("CanUseZipMode() must be true for a released version")
	}
}
