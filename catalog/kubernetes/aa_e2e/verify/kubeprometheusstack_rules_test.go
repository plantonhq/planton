package verify

import (
	"os"
	"path/filepath"
	"testing"
)

// The tuning proof reads the scenario's own manifest and the rules API, so
// both readers are pinned offline: a verifier that misread either would
// pass a live run it should fail.
func TestKpsRuleTuningReadsTheScenario(t *testing.T) {
	path := filepath.Join("..", "..", "kuberneteskubeprometheusstack", "e2e", "scenarios", "default-rules-tuned.yaml")
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	disabled, overrides := kpsRuleTuning(manifestSpecMap(path))
	if len(disabled) != 2 || disabled[0] != "KubeCPUOvercommit" || disabled[1] != "KubeMemoryOvercommit" {
		t.Errorf("disabled = %v", disabled)
	}
	want := map[string]kpsAlertOverride{
		"KubePodNotReady":   {For: "45m"},
		"CPUThrottlingHigh": {Severity: "warning"},
		"KubeJobFailed":     {For: "1h30m", Severity: "info"},
	}
	if len(overrides) != len(want) {
		t.Fatalf("overrides = %v", overrides)
	}
	for alert, o := range want {
		if overrides[alert] != o {
			t.Errorf("%s = %+v, want %+v", alert, overrides[alert], o)
		}
	}
}

func TestPrometheusDurationSeconds(t *testing.T) {
	for in, want := range map[string]float64{"45m": 2700, "1h30m": 5400, "2d": 172800, "0": 0, "500ms": 0.5} {
		got, err := prometheusDurationSeconds(in)
		if err != nil || got != want {
			t.Errorf("%q = %v (%v), want %v", in, got, err, want)
		}
	}
	if _, err := prometheusDurationSeconds("45 minutes"); err == nil {
		t.Error("a duration with no Prometheus unit must be refused")
	}
}

func TestLoadedAlertRulesIndexesEveryGroup(t *testing.T) {
	body := `{"status":"success","data":{"groups":[
	  {"name":"general.rules","rules":[{"name":"Watchdog","duration":0,"labels":{"severity":"none"}}]},
	  {"name":"kubernetes-apps","rules":[{"name":"KubePodNotReady","duration":2700,"labels":{"severity":"warning"}}]},
	  {"name":"estate","rules":[{"name":"KubePodNotReady","duration":900,"labels":{"severity":"warning"}}]}]}}`
	loaded, err := loadedAlertRules(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded["KubePodNotReady"]) != 2 || loaded["KubePodNotReady"][0].Duration != 2700 || loaded["Watchdog"][0].Labels["severity"] != "none" {
		t.Errorf("loaded = %+v", loaded)
	}
}
