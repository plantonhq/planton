package module

import (
	"encoding/json"
	"testing"

	kubernetesgofeatureflagflagfilev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgofeatureflagflagfile/v1alpha1"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestFlagFileRendersGoFeatureFlagFormat(t *testing.T) {
	pct := 25.0
	meta, _ := structpb.NewStruct(map[string]interface{}{"owner": "web"})
	spec := &kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagFlagFileSpec{
		Flags: map[string]*kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagFlag{
			"assistant": {
				Variations: map[string]*kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagValue{
					"enabled":  {Value: &kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagValue_BoolValue{BoolValue: true}},
					"disabled": {Value: &kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagValue_BoolValue{BoolValue: false}},
				},
				Targeting: []*kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagRule{
					{Name: "first", Query: `org in ["planton"]`, Variation: "enabled"},
					{Name: "ramp", ProgressiveRollout: &kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagProgressiveRollout{
						Initial: &kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagProgressiveRolloutStep{Variation: "disabled", Date: "2026-11-01T00:00:00Z"},
						End:     &kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagProgressiveRolloutStep{Variation: "enabled", Percentage: &pct, Date: "2026-11-08T00:00:00Z"},
					}},
				},
				DefaultRule: &kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagRule{Variation: "disabled"},
				Metadata:    meta,
			},
		},
	}
	doc, err := renderFlagFile(spec)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(doc), &got); err != nil {
		t.Fatal(err)
	}
	flag := got["assistant"]
	if flag["variations"].(map[string]interface{})["enabled"] != true {
		t.Errorf("a boolean variation renders as a JSON boolean: %v", flag["variations"])
	}
	if flag["defaultRule"].(map[string]interface{})["variation"] != "disabled" {
		t.Errorf("defaultRule wrong: %v", flag["defaultRule"])
	}
	targeting := flag["targeting"].([]interface{})
	if targeting[0].(map[string]interface{})["query"] != `org in ["planton"]` {
		t.Errorf("query wrong: %v", targeting[0])
	}
	initial := targeting[1].(map[string]interface{})["progressiveRollout"].(map[string]interface{})["initial"].(map[string]interface{})
	if _, ok := initial["percentage"]; ok {
		t.Errorf("an unset rollout percentage is omitted so the relay's own default applies")
	}
	if flag["metadata"].(map[string]interface{})["owner"] != "web" {
		t.Errorf("metadata wrong: %v", flag["metadata"])
	}
}

func TestEmptyFlagFileRendersAnEmptyObject(t *testing.T) {
	doc, err := renderFlagFile(&kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagFlagFileSpec{})
	if err != nil || doc != "{}" {
		t.Fatalf("got %q, %v", doc, err)
	}
}
