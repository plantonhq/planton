package module

import (
	"testing"

	kubernetesflagdflagfilev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesflagdflagfile/v1alpha1"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestFlagdFileDeclaresTheSchemaAndOmitsAnEmptyDefaultVariant(t *testing.T) {
	targeting, _ := structpb.NewStruct(map[string]interface{}{"if": []interface{}{map[string]interface{}{"in": []interface{}{map[string]interface{}{"var": "org"}, []interface{}{"planton"}}}, "on", "off"}})
	spec := &kubernetesflagdflagfilev1alpha1.KubernetesFlagdFlagFileSpec{
		Flags: map[string]*kubernetesflagdflagfilev1alpha1.KubernetesFlagdFlag{
			"assistant": {
				State: "ENABLED",
				Variants: map[string]*kubernetesflagdflagfilev1alpha1.KubernetesFlagdValue{
					"on":  {Value: &kubernetesflagdflagfilev1alpha1.KubernetesFlagdValue_BoolValue{BoolValue: true}},
					"off": {Value: &kubernetesflagdflagfilev1alpha1.KubernetesFlagdValue_BoolValue{BoolValue: false}},
				},
				Targeting: targeting,
			},
		},
	}
	doc, err := renderFlagFile(spec)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"$schema":"https://flagd.dev/schema/v0/flags.json","flags":{"assistant":{"state":"ENABLED","targeting":{"if":[{"in":[{"var":"org"},["planton"]]},"on","off"]},"variants":{"off":false,"on":true}}}}`
	if doc != want {
		t.Fatalf("got  %s\nwant %s", doc, want)
	}
}
