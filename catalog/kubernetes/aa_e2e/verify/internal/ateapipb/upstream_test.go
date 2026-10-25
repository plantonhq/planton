package ateapipb

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

// upstreamRelease is the release the copies come from; it moves with the
// KubernetesAgentSubstrate module's pin.
const upstreamRelease = "v0.6.0"

// TestCopiesMatchUpstream holds the two generated files byte-equal to
// upstream's at the pinned tag, so the verifier speaks exactly the API the
// installed release serves.
func TestCopiesMatchUpstream(t *testing.T) {
	if os.Getenv("AGENT_SUBSTRATE_UPSTREAM_PIN") != "1" {
		t.Skip("set AGENT_SUBSTRATE_UPSTREAM_PIN=1 to compare the copies with upstream's files at " + upstreamRelease)
	}
	client := &http.Client{Timeout: time.Minute}
	for _, name := range []string{"ateapi.pb.go", "ateapi_grpc.pb.go"} {
		url := "https://raw.githubusercontent.com/agent-substrate/substrate/" + upstreamRelease + "/pkg/proto/ateapipb/" + name
		resp, err := client.Get(url)
		if err != nil {
			t.Fatalf("fetching %s: %v", url, err)
		}
		want, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("fetching %s: status %d, %v", url, resp.StatusCode, err)
		}
		got, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s differs from upstream's at %s; copy it again, unedited", name, upstreamRelease)
		}
	}
}
