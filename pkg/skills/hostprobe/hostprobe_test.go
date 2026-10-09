package hostprobe

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestCodingAgentJourney runs the developer's journey through every coding
// agent installed on the machine, on the skill content under test and --
// when a baseline is given -- on the content that predates the claim, so
// the same checks are seen to FAIL on the old content and PASS on the new.
//
// Gated by PLANTON_SKILL_HOST_PROBE=1: each leg spends real model tokens and
// needs the host's CLI signed in. Inputs, all optional:
//
//	PLANTON_SKILL_HOST_PROBE_CONTENT   the planton skill directory under proof
//	                                   (default: ../../../skills/planton)
//	PLANTON_SKILL_HOST_PROBE_CATALOG   a multi-cloud-catalog skill directory,
//	                                   ideally a packaged copy carrying
//	                                   kinds/ (default: the working
//	                                   tree's, which has no pack)
//	PLANTON_SKILL_HOST_PROBE_BASELINE  a planton skill directory expected to
//	                                   FAIL the checks (a released copy); the
//	                                   RED leg is skipped when unset
//	PLANTON_SKILL_HOST_PROBE_HOSTS     comma-free single host name to run
//	                                   only that host ("cursor", "claude")
//	PLANTON_SKILL_HOST_PROBE_KEEP=1    keep fixtures and transcripts on disk
//
// This authoring probe inherits the host's environment and home. Run it only
// in an isolated environment without live platform credentials: a temporary
// repository alone does not remove an existing CLI login. Activity probes
// use RunActivityHost's restricted tools and closed fixture transport instead.
func TestCodingAgentJourney(t *testing.T) {
	if os.Getenv("PLANTON_SKILL_HOST_PROBE") != "1" {
		t.Skip("set PLANTON_SKILL_HOST_PROBE=1 to run the live coding-agent probe (spends tokens; needs a signed-in agent CLI)")
	}
	content := Content{
		Planton: envOr("PLANTON_SKILL_HOST_PROBE_CONTENT", filepath.Join("..", "..", "..", "skills", "planton")),
		Catalog: envOr("PLANTON_SKILL_HOST_PROBE_CATALOG", filepath.Join("..", "..", "..", "skills", "multi-cloud-catalog")),
	}
	baseline := os.Getenv("PLANTON_SKILL_HOST_PROBE_BASELINE")
	only := os.Getenv("PLANTON_SKILL_HOST_PROBE_HOSTS")
	validate := plantonValidate(t)

	for _, host := range Hosts {
		host := host
		if only != "" && only != host.Name {
			continue
		}
		t.Run(host.Name, func(t *testing.T) {
			if _, err := exec.LookPath(host.Binary); err != nil {
				t.Skipf("%s is not on PATH; install and sign in to run this leg", host.Binary)
			}
			if baseline != "" {
				t.Run("red-baseline", func(t *testing.T) {
					vs := runLeg(t, host, Content{Planton: baseline, Catalog: content.Catalog}, validate)
					if len(vs) == 0 {
						t.Errorf("the baseline content passed every check -- the claim is not discriminative on this host; record it per the validation doctrine")
					}
					for _, v := range vs {
						t.Logf("baseline violated (expected): %s", v)
					}
				})
			}
			t.Run("green", func(t *testing.T) {
				vs := runLeg(t, host, content, validate)
				for _, v := range vs {
					t.Errorf("violated: %s", v)
				}
			})
		})
	}
}

func runLeg(t *testing.T, host Host, content Content, validate func(string) error) []Violation {
	t.Helper()
	fx, err := NewFixture(host, content)
	if err != nil {
		t.Fatalf("building fixture: %v", err)
	}
	keep := os.Getenv("PLANTON_SKILL_HOST_PROBE_KEEP") == "1"
	defer func() {
		if !keep && !t.Failed() {
			fx.Remove()
		} else {
			t.Logf("fixture kept at %s", fx.Repo)
		}
	}()
	t.Logf("%s: running in %s", host.Name, fx.Repo)
	transcript, runErr := fx.Run(context.Background(), host, Prompt, 20*time.Minute)
	path := filepath.Join(fx.Parent, host.Name+"-transcript.jsonl")
	_ = os.WriteFile(path, transcript, 0o644)
	t.Logf("transcript: %s (%d bytes)", path, len(transcript))
	if runErr != nil {
		t.Logf("%s exited with: %v (the tree is judged regardless)", host.Binary, runErr)
	}
	return fx.Judge(transcript, validate)
}

// plantonValidate returns the CLI-backed manifest oracle, or nil (with a
// logged reason) when the CLI is absent so the leg still judges the tree.
func plantonValidate(t *testing.T) func(string) error {
	if _, err := exec.LookPath("planton"); err != nil {
		t.Log("planton CLI not on PATH; manifests are checked for shape only")
		return nil
	}
	return func(manifest string) error {
		out, err := exec.Command("planton", "validate", manifest).CombinedOutput()
		if err != nil {
			return fmt.Errorf("%w: %s", err, out)
		}
		return nil
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
