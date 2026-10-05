package module

import (
	"encoding/json"

	"github.com/pkg/errors"
	kubernetesflagdflagfilev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesflagdflagfile/v1alpha1"
)

// SchemaURL is the flagd flag-definition schema the rendered file declares.
const SchemaURL = "https://flagd.dev/schema/v0/flags.json"

// renderFlagFile renders the flag definitions as flagd JSON. Both engines
// produce it byte for byte (Go json.Marshal and OpenTofu jsonencode sort
// keys and escape alike). Terraform twin: local.flag_file in locals.tf.
func renderFlagFile(spec *kubernetesflagdflagfilev1alpha1.KubernetesFlagdFlagFileSpec) (string, error) {
	flags := map[string]interface{}{}
	for name, f := range spec.GetFlags() {
		variants := map[string]interface{}{}
		for vn, v := range f.GetVariants() {
			variants[vn] = value(v)
		}
		flag := map[string]interface{}{
			"state":    f.GetState(),
			"variants": variants,
		}
		if f.GetDefaultVariant() != "" {
			flag["defaultVariant"] = f.GetDefaultVariant()
		}
		if len(f.GetTargeting().GetFields()) > 0 {
			flag["targeting"] = f.GetTargeting().AsMap()
		}
		if len(f.GetMetadata().GetFields()) > 0 {
			flag["metadata"] = f.GetMetadata().AsMap()
		}
		flags[name] = flag
	}
	doc := map[string]interface{}{
		"$schema": SchemaURL,
		"flags":   flags,
	}
	if len(spec.GetEvaluators()) > 0 {
		evaluators := map[string]interface{}{}
		for name, e := range spec.GetEvaluators() {
			evaluators[name] = e.AsMap()
		}
		doc["$evaluators"] = evaluators
	}
	if len(spec.GetMetadata().GetFields()) > 0 {
		doc["metadata"] = spec.GetMetadata().AsMap()
	}
	body, err := json.Marshal(doc)
	if err != nil {
		return "", errors.Wrap(err, "failed to encode the flag definitions")
	}
	return string(body), nil
}

func value(v *kubernetesflagdflagfilev1alpha1.KubernetesFlagdValue) interface{} {
	switch x := v.GetValue().(type) {
	case *kubernetesflagdflagfilev1alpha1.KubernetesFlagdValue_BoolValue:
		return x.BoolValue
	case *kubernetesflagdflagfilev1alpha1.KubernetesFlagdValue_StringValue:
		return x.StringValue
	case *kubernetesflagdflagfilev1alpha1.KubernetesFlagdValue_NumberValue:
		return x.NumberValue
	case *kubernetesflagdflagfilev1alpha1.KubernetesFlagdValue_ObjectValue:
		return x.ObjectValue.AsMap()
	}
	return nil
}
