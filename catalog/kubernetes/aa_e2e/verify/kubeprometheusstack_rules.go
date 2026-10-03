package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// proveRuleTuning checks the manifest's per-alert tuning of the curated
// rules against what Prometheus actually loaded (/api/v1/rules), because the
// chart skips an unknown alert name without a word: a switch that names no
// rule must fail here, not pass the apply. Every disabled alert is absent,
// every overridden alert is loaded with its declared hold and severity, and
// the Watchdog still loads, so the rest of the curated set survived the
// switch. The rule files reach Prometheus through the operator's config
// reloader, so the check waits for the set to settle.
func (v *KubePrometheusStackVerifier) proveRuleTuning(ctx context.Context, base string) error {
	if len(v.DisabledAlerts) == 0 && len(v.AlertOverrides) == 0 {
		return nil
	}
	if err := awaitHTTPCondition(ctx, base+"/api/v1/rules?type=alert", 5*time.Minute, func(body string) error {
		loaded, err := loadedAlertRules(body)
		if err != nil {
			return err
		}
		if _, ok := loaded["Watchdog"]; !ok {
			return errors.New("the curated rules have not loaded yet (no Watchdog)")
		}
		for _, alert := range v.DisabledAlerts {
			if _, ok := loaded[alert]; ok {
				return errors.Errorf("%s is disabled but still loaded", alert)
			}
		}
		for _, alert := range overriddenAlerts(v.AlertOverrides) {
			want := v.AlertOverrides[alert]
			rules, ok := loaded[alert]
			if !ok {
				return errors.Errorf("%s is overridden but no rule by that name is loaded (a name the chart does not have?)", alert)
			}
			for _, r := range rules {
				if want.For != "" {
					seconds, err := prometheusDurationSeconds(want.For)
					if err != nil {
						return err
					}
					if r.Duration != seconds {
						return errors.Errorf("%s holds %vs, want %s (%vs)", alert, r.Duration, want.For, seconds)
					}
				}
				if want.Severity != "" && r.Labels["severity"] != want.Severity {
					return errors.Errorf("%s carries severity %q, want %q", alert, r.Labels["severity"], want.Severity)
				}
			}
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "RULE-TUNING: the loaded rules never matched the declared tuning")
	}
	fmt.Printf("  [verify] RULE-TUNING: %d alert(s) absent from the loaded rules, %d carrying their declared hold and severity, the Watchdog still loaded\n",
		len(v.DisabledAlerts), len(v.AlertOverrides))
	return nil
}

type loadedAlertRule struct {
	Duration float64           `json:"duration"`
	Labels   map[string]string `json:"labels"`
}

// loadedAlertRules indexes a /api/v1/rules?type=alert payload by alert name
// (one name can appear in several groups).
func loadedAlertRules(body string) (map[string][]loadedAlertRule, error) {
	var payload struct {
		Data struct {
			Groups []struct {
				Rules []struct {
					Name string `json:"name"`
					loadedAlertRule
				} `json:"rules"`
			} `json:"groups"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return nil, errors.Wrap(err, "parsing the rules response")
	}
	loaded := map[string][]loadedAlertRule{}
	for _, g := range payload.Data.Groups {
		for _, r := range g.Rules {
			loaded[r.Name] = append(loaded[r.Name], r.loadedAlertRule)
		}
	}
	return loaded, nil
}

var prometheusDurationPart = regexp.MustCompile(`([0-9]+)(ms|y|w|d|h|m|s)`)

// prometheusDurationSeconds converts a Prometheus duration ("45m", "1h30m",
// "2d") to seconds, the unit the rules API reports a hold in.
func prometheusDurationSeconds(d string) (float64, error) {
	unit := map[string]float64{"ms": 0.001, "s": 1, "m": 60, "h": 3600, "d": 86400, "w": 7 * 86400, "y": 365 * 86400}
	parts := prometheusDurationPart.FindAllStringSubmatch(d, -1)
	if d != "0" && len(parts) == 0 {
		return 0, errors.Errorf("%q is not a Prometheus duration", d)
	}
	var total float64
	for _, p := range parts {
		n, _ := strconv.ParseFloat(p[1], 64)
		total += n * unit[p[2]]
	}
	return total, nil
}

func overriddenAlerts(m map[string]kpsAlertOverride) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
