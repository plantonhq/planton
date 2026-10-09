package hostprobe

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

func TestActivityJourneyCannotPassWhenHostApprovalBlocksEveryTool(t *testing.T) {
	// A CLI refusal scene may legitimately need no command. It must still
	// load the skill; "I could not call any tool" is not behavioral proof.
	scenario := ActivityScenario{Exchanges: []ActivityExchange{{Tool: "run_command", Optional: true}}}
	if len(judgeActivityJourney(scenario, fstest.MapFS{}, nil)) == 0 {
		t.Fatal("blocked host mistaken for a successful skill journey")
	}
}

func TestActivityHostEnvironmentDoesNotCarryPlatformAccess(t *testing.T) {
	t.Setenv("PLANTON_API_ENDPOINT", "live.example:443")
	t.Setenv("PLANTON_ACCESS_TOKEN", "not-a-real-token")
	t.Setenv("AWS_ACCESS_KEY_ID", "not-a-real-key")
	t.Setenv("KUBECONFIG", "/not/a/real/config")
	t.Setenv("CODEX_THREAD_ID", "outer-thread")
	for _, entry := range activityHostEnv() {
		for _, prefix := range []string{"PLANTON_", "AWS_", "KUBECONFIG=", "CODEX_THREAD_ID="} {
			if strings.HasPrefix(entry, prefix) {
				t.Fatalf("inherited %s", strings.SplitN(entry, "=", 2)[0])
			}
		}
	}
}

// Opt-in because this spends model tokens. Every recorded scenario can run
// through either verified restricted host, without any live Planton access.
// Evidence directories are retained and printed even when a host fails.
func TestActivityAgentJourneys(t *testing.T) {
	if os.Getenv("PLANTON_SKILL_ACTIVITY_PROBE") != "1" {
		t.Skip("set PLANTON_SKILL_ACTIVITY_PROBE=1 for external-agent fixture journeys")
	}
	binary := os.Getenv("PLANTON_SKILL_ACTIVITY_FIXTURE")
	if binary == "" {
		t.Fatal("build ./pkg/skills/hostprobe/cmd/activityfixture and set PLANTON_SKILL_ACTIVITY_FIXTURE to its absolute path")
	}
	skill, err := filepath.Abs(envOr("PLANTON_SKILL_HOST_PROBE_CONTENT", "../../../skills/planton"))
	if err != nil {
		t.Fatal(err)
	}
	paths, err := filepath.Glob("testdata/activity/*.json")
	if err != nil || len(paths) == 0 {
		t.Fatal("activity scenarios not found", err)
	}
	for _, host := range []string{"claude", "codex"} {
		if only := os.Getenv("PLANTON_SKILL_HOST_PROBE_HOSTS"); only != "" && only != host {
			continue
		}
		t.Run(host, func(t *testing.T) {
			if _, err := exec.LookPath(host); err != nil {
				t.Skipf("%s unavailable: %v", host, err)
			}
			for _, path := range paths {
				name := strings.TrimSuffix(filepath.Base(path), ".json")
				if only := os.Getenv("PLANTON_SKILL_ACTIVITY_SCENARIO"); only != "" && only != name {
					continue
				}
				t.Run(name, func(t *testing.T) {
					full, _ := filepath.Abs(path)
					result, err := RunActivityHost(context.Background(), host, binary, skill, full, 4*time.Minute)
					t.Logf("evidence: %s", result.Directory)
					if err != nil {
						t.Fatalf("host did not complete: %v", err)
					}
					for _, v := range result.Violations {
						t.Error(v)
					}
				})
			}
		})
	}
}
