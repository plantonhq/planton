package module

import (
	"encoding/json"

	"github.com/pkg/errors"
	kubernetesgofeatureflagflagfilev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgofeatureflagflagfile/v1alpha1"
)

// renderFlagFile renders the flags as one JSON object keyed by flag name.
// JSON is YAML, so the relay's default `yaml` file format reads it; both
// engines produce it byte for byte (Go json.Marshal and OpenTofu jsonencode
// sort keys and escape alike). Keys follow GO Feature Flag's flag format
// (variations, targeting, defaultRule, percentage, progressiveRollout,
// scheduledRollout, ...). Terraform twin: local.flag_file in locals.tf.
func renderFlagFile(spec *kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagFlagFileSpec) (string, error) {
	flags := map[string]interface{}{}
	for name, f := range spec.GetFlags() {
		flag := map[string]interface{}{
			"variations":  variations(f.GetVariations()),
			"defaultRule": rule(f.GetDefaultRule()),
		}
		if len(f.GetTargeting()) > 0 {
			flag["targeting"] = rules(f.GetTargeting())
		}
		setString(flag, "bucketingKey", f.GetBucketingKey())
		if f.TrackEvents != nil {
			flag["trackEvents"] = f.GetTrackEvents()
		}
		if f.GetDisable() {
			flag["disable"] = true
		}
		setString(flag, "version", f.GetVersion())
		if len(f.GetMetadata().GetFields()) > 0 {
			flag["metadata"] = f.GetMetadata().AsMap()
		}
		if e := f.GetExperimentation(); e != nil {
			flag["experimentation"] = map[string]interface{}{"start": e.GetStart(), "end": e.GetEnd()}
		}
		if len(f.GetScheduledRollout()) > 0 {
			steps := []interface{}{}
			for _, s := range f.GetScheduledRollout() {
				steps = append(steps, scheduledStep(s))
			}
			flag["scheduledRollout"] = steps
		}
		flags[name] = flag
	}
	body, err := json.Marshal(flags)
	if err != nil {
		return "", errors.Wrap(err, "failed to encode the flag file")
	}
	return string(body), nil
}

func variations(vs map[string]*kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagValue) map[string]interface{} {
	out := map[string]interface{}{}
	for name, v := range vs {
		out[name] = value(v)
	}
	return out
}

func value(v *kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagValue) interface{} {
	switch x := v.GetValue().(type) {
	case *kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagValue_BoolValue:
		return x.BoolValue
	case *kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagValue_StringValue:
		return x.StringValue
	case *kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagValue_NumberValue:
		return x.NumberValue
	case *kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagValue_ObjectValue:
		return x.ObjectValue.AsMap()
	case *kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagValue_ListValue:
		return x.ListValue.AsSlice()
	}
	return nil
}

func rules(rs []*kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagRule) []interface{} {
	out := []interface{}{}
	for _, r := range rs {
		out = append(out, rule(r))
	}
	return out
}

func rule(r *kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagRule) map[string]interface{} {
	out := map[string]interface{}{}
	if r == nil {
		return out
	}
	setString(out, "name", r.GetName())
	setString(out, "query", r.GetQuery())
	setString(out, "variation", r.GetVariation())
	if len(r.GetPercentage()) > 0 {
		pct := map[string]interface{}{}
		for k, v := range r.GetPercentage() {
			pct[k] = v
		}
		out["percentage"] = pct
	}
	if pr := r.GetProgressiveRollout(); pr != nil {
		out["progressiveRollout"] = map[string]interface{}{
			"initial": rolloutStep(pr.GetInitial()),
			"end":     rolloutStep(pr.GetEnd()),
		}
	}
	if r.GetDisable() {
		out["disable"] = true
	}
	return out
}

func rolloutStep(s *kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagProgressiveRolloutStep) map[string]interface{} {
	out := map[string]interface{}{"variation": s.GetVariation(), "date": s.GetDate()}
	if s.Percentage != nil {
		out["percentage"] = s.GetPercentage()
	}
	return out
}

func scheduledStep(s *kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagScheduledStep) map[string]interface{} {
	out := map[string]interface{}{"date": s.GetDate()}
	if len(s.GetVariations()) > 0 {
		out["variations"] = variations(s.GetVariations())
	}
	if len(s.GetTargeting()) > 0 {
		out["targeting"] = rules(s.GetTargeting())
	}
	if s.GetDefaultRule() != nil {
		out["defaultRule"] = rule(s.GetDefaultRule())
	}
	if s.TrackEvents != nil {
		out["trackEvents"] = s.GetTrackEvents()
	}
	if s.Disable != nil {
		out["disable"] = s.GetDisable()
	}
	setString(out, "version", s.GetVersion())
	if e := s.GetExperimentation(); e != nil {
		out["experimentation"] = map[string]interface{}{"start": e.GetStart(), "end": e.GetEnd()}
	}
	return out
}

func setString(m map[string]interface{}, key, value string) {
	if value != "" {
		m[key] = value
	}
}
