package module

import (
	"reflect"
	"testing"

	kuberneteskubeprometheusstackv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskubeprometheusstack/v1alpha1"
	"github.com/plantonhq/planton/shared"
)

// These tests pin how the curated rules' tuning reaches the chart: whole
// groups under defaultRules.rules, single alerts under defaultRules.disabled,
// and per-alert holds and severities under the top-level customRules map,
// with only the keys the manifest sets. The OpenTofu module's
// default_rules_values and custom_rules_values render the same shapes.

func defaultRulesLocals(rules *kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackDefaultRules) *Locals {
	return initializeLocals(nil, &kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackIacInput{
		Target: &kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStack{
			Metadata: &shared.CatalogObjectMetadata{Name: "management-metrics"},
			Spec: &kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackSpec{
				Namespace:    literal("observability"),
				DefaultRules: rules,
			},
		},
	})
}

func TestCuratedRuleTuningRendersTheChartsOwnKeys(t *testing.T) {
	values, err := buildHelmValues(defaultRulesLocals(&kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackDefaultRules{
		DisabledGroups: []string{"etcd"},
		DisabledAlerts: []string{"KubeCPUOvercommit", "KubeMemoryOvercommit"},
		AlertOverrides: []*kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackAlertOverride{
			{Alert: "KubePodNotReady", ForDuration: "30m"},
			{Alert: "CPUThrottlingHigh", Severity: "info"},
			{Alert: "KubeJobFailed", ForDuration: "1h", Severity: "warning"},
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	wantDefaultRules := map[string]interface{}{
		"rules":    map[string]interface{}{"etcd": false},
		"disabled": map[string]interface{}{"KubeCPUOvercommit": true, "KubeMemoryOvercommit": true},
	}
	if !reflect.DeepEqual(values["defaultRules"], wantDefaultRules) {
		t.Errorf("defaultRules = %v, want %v", values["defaultRules"], wantDefaultRules)
	}
	wantCustomRules := map[string]interface{}{
		"KubePodNotReady":   map[string]interface{}{"for": "30m"},
		"CPUThrottlingHigh": map[string]interface{}{"severity": "info"},
		"KubeJobFailed":     map[string]interface{}{"for": "1h", "severity": "warning"},
	}
	if !reflect.DeepEqual(values["customRules"], wantCustomRules) {
		t.Errorf("customRules = %v, want %v", values["customRules"], wantCustomRules)
	}
}

func TestUntunedCuratedRulesLeaveTheChartDefaults(t *testing.T) {
	for name, rules := range map[string]*kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackDefaultRules{
		"no block":    nil,
		"empty block": {},
	} {
		values, err := buildHelmValues(defaultRulesLocals(rules))
		if err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{"defaultRules", "customRules"} {
			if v, ok := values[key]; ok {
				t.Errorf("%s: %s rendered %v; an untuned stack must leave the chart's own", name, key, v)
			}
		}
	}
}
