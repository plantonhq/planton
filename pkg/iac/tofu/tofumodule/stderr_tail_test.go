package tofumodule

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/plantonhq/planton/shared/iac/terraform"
	"google.golang.org/protobuf/types/known/structpb"
)

// fakeEngine puts a "tofu" on PATH that fails before emitting any JSON: its
// only explanation is on stderr, the way a refused flag or an unparsable
// module fails.
func fakeEngine(t *testing.T, stderrLine string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("test relies on a POSIX shell")
	}
	bin := t.TempDir()
	script := "#!/bin/sh\necho '" + stderrLine + "' >&2\nexit 1\n"
	if err := os.WriteFile(filepath.Join(bin, "tofu"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func moduleDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".terraform"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func drained() chan string {
	lines := make(chan string, 16)
	go func() {
		for range lines {
		}
	}()
	return lines
}

func TestRunOperation_JSONMode_AFailureCarriesTheEnginesOwnSentence(t *testing.T) {
	fakeEngine(t, "Error: Unsupported block type")
	manifest, _ := structpb.NewStruct(map[string]any{})
	lines := drained()
	defer close(lines)

	err := RunOperation(context.Background(), "tofu", moduleDir(t), terraform.TerraformOperationType_apply,
		true, false, manifest, nil, true, lines)

	if err == nil {
		t.Fatal("expected the failed apply to return an error")
	}
	if !strings.Contains(err.Error(), "\nstderr: Error: Unsupported block type") {
		t.Fatalf("the failure lost the engine's own sentence; got: %v", err)
	}
}

func TestInit_JSONMode_AFailureCarriesTheEnginesOwnSentence(t *testing.T) {
	fakeEngine(t, "Error: Failed to query available provider packages")
	manifest, _ := structpb.NewStruct(map[string]any{})
	lines := drained()
	defer close(lines)

	err := Init(context.Background(), "tofu", moduleDir(t), manifest, terraform.TerraformBackendType_local,
		nil, nil, false, true, lines)

	if err == nil {
		t.Fatal("expected the failed init to return an error")
	}
	if !strings.Contains(err.Error(), "\nstderr: Error: Failed to query available provider packages") {
		t.Fatalf("the failure lost the engine's own sentence; got: %v", err)
	}
}

func TestWithStderr_KeepsTheCauseAndBoundsTheTail(t *testing.T) {
	cause := errors.New("failed to execute tofu command: exit status 1")

	if withStderr(nil, "anything") != nil {
		t.Fatal("no failure means no error")
	}
	if got := withStderr(cause, "  \n "); got != cause {
		t.Fatalf("an empty stderr leaves the failure as it was; got %v", got)
	}

	long := strings.Repeat("noise line\n", 2000) + "Error: the real cause"
	got := withStderr(cause, long)
	if !errors.Is(got, cause) {
		t.Fatal("the failure still unwraps to its cause")
	}
	if !strings.HasSuffix(got.Error(), "Error: the real cause") {
		t.Fatalf("the tail keeps the engine's last words; got ...%q", got.Error()[len(got.Error())-40:])
	}
	tail := got.Error()[strings.Index(got.Error(), "\nstderr: ")+len("\nstderr: "):]
	if len(tail) > maxStderrTail {
		t.Fatalf("the tail is bounded to %d bytes; got %d", maxStderrTail, len(tail))
	}
	if strings.HasPrefix(tail, "oise") || !strings.HasPrefix(tail, "noise line") {
		t.Fatalf("the tail starts on a whole line; got %q", tail[:20])
	}
}
