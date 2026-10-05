package kubernetesgofeatureflagflagfilev1alpha1

import (
	"errors"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestKubernetesGoFeatureFlagFlagFile(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesGoFeatureFlagFlagFile Suite")
}

func strPtr(s string) *string       { return &s }
func boolPtr(b bool) *bool          { return &b }
func float64Ptr(f float64) *float64 { return &f }

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value}}
}

func violations(err error) []string {
	gomega.Expect(err).NotTo(gomega.BeNil())
	var verr *protovalidate.ValidationError
	gomega.Expect(errors.As(err, &verr)).To(gomega.BeTrue())
	out := []string{}
	for _, v := range verr.Violations {
		out = append(out, v.Proto.GetRuleId(), protovalidate.FieldPathString(v.Proto.GetField()))
	}
	return out
}

func mustStruct(m map[string]any) *structpb.Struct {
	s, err := structpb.NewStruct(m)
	gomega.Expect(err).To(gomega.BeNil())
	return s
}

func boolValue(b bool) *KubernetesGoFeatureFlagValue {
	return &KubernetesGoFeatureFlagValue{Value: &KubernetesGoFeatureFlagValue_BoolValue{BoolValue: b}}
}

func stringValue(v string) *KubernetesGoFeatureFlagValue {
	return &KubernetesGoFeatureFlagValue{Value: &KubernetesGoFeatureFlagValue_StringValue{StringValue: v}}
}

// onOff is the common release flag: on for the listed organizations, off for
// everyone else.
func onOff() *KubernetesGoFeatureFlagFlag {
	return &KubernetesGoFeatureFlagFlag{
		Variations: map[string]*KubernetesGoFeatureFlagValue{"enabled": boolValue(true), "disabled": boolValue(false)},
		Targeting: []*KubernetesGoFeatureFlagRule{
			{Name: "first-organizations", Query: `org in ["planton"]`, Variation: "enabled"},
		},
		DefaultRule: &KubernetesGoFeatureFlagRule{Variation: "disabled"},
	}
}

var _ = ginkgo.Describe("KubernetesGoFeatureFlagFlagFile Validation Tests", func() {
	var input *KubernetesGoFeatureFlagFlagFile

	ginkgo.BeforeEach(func() {
		input = &KubernetesGoFeatureFlagFlagFile{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesGoFeatureFlagFlagFile",
			Metadata:   &shared.CatalogObjectMetadata{Name: "release-flags"},
			Spec: &KubernetesGoFeatureFlagFlagFileSpec{
				Namespace: literal("feature-flags"),
				Flags:     map[string]*KubernetesGoFeatureFlagFlag{"assistant": onOff()},
			},
		}
	})

	invalid := func(want string) {
		gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement(gomega.ContainSubstring(want)))
	}

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("an on/off release flag should be valid", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an empty flag file should be valid", func() {
			input.Spec.Flags = nil
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference and a custom key should be valid", func() {
			input.Spec.Namespace = &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{
				Kind: catalogkind.CatalogKind_KubernetesNamespace, Name: "feature-flags", FieldPath: "spec.name",
			}}}
			input.Spec.Key = strPtr("release.goff.yaml")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a fully populated flag (every rule shape, rollout and step) should be valid", func() {
			input.Spec.Flags["checkout-theme"] = &KubernetesGoFeatureFlagFlag{
				Variations: map[string]*KubernetesGoFeatureFlagValue{"light": stringValue("light"), "dark": stringValue("dark"), "auto": stringValue("auto")},
				Targeting: []*KubernetesGoFeatureFlagRule{
					{Name: "staff", Query: `email ew "@example.com"`, Variation: "dark"},
					{Name: "beta", Query: `beta eq true`, Percentage: map[string]float64{"dark": 25, "light": 75}},
					{Name: "ramp", Query: `country eq "FR"`, ProgressiveRollout: &KubernetesGoFeatureFlagProgressiveRollout{
						Initial: &KubernetesGoFeatureFlagProgressiveRolloutStep{Variation: "light", Percentage: float64Ptr(0), Date: "2026-11-01T09:00:00Z"},
						End:     &KubernetesGoFeatureFlagProgressiveRolloutStep{Variation: "dark", Percentage: float64Ptr(100), Date: "2026-11-15T09:00:00+05:30"},
					}},
					{Name: "parked", Disable: true},
				},
				DefaultRule:     &KubernetesGoFeatureFlagRule{Percentage: map[string]float64{"light": 90, "auto": 10}},
				BucketingKey:    "companyId",
				TrackEvents:     boolPtr(false),
				Version:         "3",
				Metadata:        mustStruct(map[string]any{"description": "checkout theme", "owner": "web"}),
				Experimentation: &KubernetesGoFeatureFlagExperimentation{Start: "2026-11-01T00:00:00Z", End: "2026-12-01T00:00:00.5Z"},
				ScheduledRollout: []*KubernetesGoFeatureFlagScheduledStep{
					{Date: "2026-11-08T09:00:00Z", Targeting: []*KubernetesGoFeatureFlagRule{{Name: "beta", Percentage: map[string]float64{"dark": 50, "light": 50}}}},
					{Date: "2026-12-01T09:00:00Z", DefaultRule: &KubernetesGoFeatureFlagRule{Variation: "dark"}, Disable: boolPtr(false), TrackEvents: boolPtr(true), Version: "4",
						Variations:      map[string]*KubernetesGoFeatureFlagValue{"contrast": stringValue("contrast")},
						Experimentation: &KubernetesGoFeatureFlagExperimentation{Start: "2026-12-01T00:00:00Z", End: "2027-01-01T00:00:00Z"}},
					{Date: "2026-12-15T09:00:00Z", Targeting: []*KubernetesGoFeatureFlagRule{{Name: "beta", Percentage: map[string]float64{"light": -1, "dark": 100}}}},
				},
			}
			input.Spec.Flags["limits"] = &KubernetesGoFeatureFlagFlag{
				Variations: map[string]*KubernetesGoFeatureFlagValue{
					"small": {Value: &KubernetesGoFeatureFlagValue_ObjectValue{ObjectValue: mustStruct(map[string]any{"rps": 10})}},
					"large": {Value: &KubernetesGoFeatureFlagValue_ObjectValue{ObjectValue: mustStruct(map[string]any{"rps": 100})}},
				},
				DefaultRule: &KubernetesGoFeatureFlagRule{Variation: "small"},
			}
			lv, _ := structpb.NewList([]any{"a", "b"})
			input.Spec.Flags["regions"] = &KubernetesGoFeatureFlagFlag{
				Variations:  map[string]*KubernetesGoFeatureFlagValue{"default": {Value: &KubernetesGoFeatureFlagValue_ListValue{ListValue: lv}}},
				DefaultRule: &KubernetesGoFeatureFlagRule{Variation: "default"},
			}
			input.Spec.Flags["ratio"] = &KubernetesGoFeatureFlagFlag{
				Variations:  map[string]*KubernetesGoFeatureFlagValue{"low": {Value: &KubernetesGoFeatureFlagValue_NumberValue{NumberValue: 0.1}}, "high": {Value: &KubernetesGoFeatureFlagValue_NumberValue{NumberValue: 0.9}}},
				DefaultRule: &KubernetesGoFeatureFlagRule{Variation: "low"},
				Disable:     true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a missing namespace should fail", func() {
			input.Spec.Namespace = nil
			invalid("namespace")
		})

		ginkgo.It("a key with a slash should fail", func() {
			input.Spec.Key = strPtr("flags/goff.yaml")
			invalid("spec.key")
		})

		ginkgo.It("a flag name with whitespace should fail", func() {
			input.Spec.Flags["new banner"] = onOff()
			invalid("spec.flags.names")
		})

		ginkgo.It("a flag without variations should fail", func() {
			f := onOff()
			f.Variations = nil
			f.Targeting = nil
			f.DefaultRule = &KubernetesGoFeatureFlagRule{Percentage: map[string]float64{}}
			input.Spec.Flags["assistant"] = f
			invalid("variations")
		})

		ginkgo.It("variations of mixed types should fail", func() {
			f := onOff()
			f.Variations["disabled"] = stringValue("off")
			input.Spec.Flags["assistant"] = f
			invalid("flag.variations.one_type")
		})

		ginkgo.It("a variation with no value should fail", func() {
			f := onOff()
			f.Variations["disabled"] = &KubernetesGoFeatureFlagValue{}
			input.Spec.Flags["assistant"] = f
			invalid("value")
		})

		ginkgo.It("a flag without a default rule should fail", func() {
			f := onOff()
			f.DefaultRule = nil
			input.Spec.Flags["assistant"] = f
			invalid("default_rule")
		})

		ginkgo.It("a default rule with a query should fail", func() {
			f := onOff()
			f.DefaultRule.Query = `org eq "x"`
			input.Spec.Flags["assistant"] = f
			invalid("flag.default_rule.no_query")
		})

		ginkgo.It("a default rule that resolves to nothing should fail", func() {
			f := onOff()
			f.DefaultRule = &KubernetesGoFeatureFlagRule{}
			input.Spec.Flags["assistant"] = f
			invalid("flag.default_rule.resolves")
		})

		ginkgo.It("an enabled targeting rule that resolves to nothing should fail", func() {
			f := onOff()
			f.Targeting = append(f.Targeting, &KubernetesGoFeatureFlagRule{Name: "empty", Query: `org eq "x"`})
			input.Spec.Flags["assistant"] = f
			invalid("flag.targeting.resolves")
		})

		ginkgo.It("a rule naming an unknown variation should fail", func() {
			f := onOff()
			f.Targeting[0].Variation = "on"
			input.Spec.Flags["assistant"] = f
			invalid("flag.rules.known_variations")
		})

		ginkgo.It("a percentage naming an unknown variation should fail", func() {
			f := onOff()
			f.DefaultRule = &KubernetesGoFeatureFlagRule{Percentage: map[string]float64{"enabled": 10, "off": 90}}
			input.Spec.Flags["assistant"] = f
			invalid("flag.rules.known_variations")
		})

		ginkgo.It("duplicated rule names should fail", func() {
			f := onOff()
			f.Targeting = append(f.Targeting, &KubernetesGoFeatureFlagRule{Name: "first-organizations", Query: `org eq "y"`, Variation: "enabled"})
			input.Spec.Flags["assistant"] = f
			invalid("flag.targeting.unique_names")
		})

		ginkgo.It("a rule with two outcomes should fail", func() {
			f := onOff()
			f.Targeting[0].Percentage = map[string]float64{"enabled": 50, "disabled": 50}
			input.Spec.Flags["assistant"] = f
			invalid("flag.rules.one_outcome")
		})

		ginkgo.It("a negative percentage share should fail", func() {
			f := onOff()
			f.DefaultRule = &KubernetesGoFeatureFlagRule{Percentage: map[string]float64{"enabled": -5, "disabled": 10}}
			input.Spec.Flags["assistant"] = f
			invalid("flag.rules.percentage_shares")
		})

		ginkgo.It("an all-zero percentage split should fail", func() {
			f := onOff()
			f.DefaultRule = &KubernetesGoFeatureFlagRule{Percentage: map[string]float64{"enabled": 0, "disabled": 0}}
			input.Spec.Flags["assistant"] = f
			invalid("flag.rules.percentage_shares")
		})

		ginkgo.It("a progressive rollout without an end should fail", func() {
			f := onOff()
			f.DefaultRule = &KubernetesGoFeatureFlagRule{ProgressiveRollout: &KubernetesGoFeatureFlagProgressiveRollout{
				Initial: &KubernetesGoFeatureFlagProgressiveRolloutStep{Variation: "disabled", Date: "2026-11-01T09:00:00Z"},
			}}
			input.Spec.Flags["assistant"] = f
			invalid("end")
		})

		ginkgo.It("a progressive rollout step with a non-RFC 3339 date should fail", func() {
			f := onOff()
			f.DefaultRule = &KubernetesGoFeatureFlagRule{ProgressiveRollout: &KubernetesGoFeatureFlagProgressiveRollout{
				Initial: &KubernetesGoFeatureFlagProgressiveRolloutStep{Variation: "disabled", Date: "2026-11-01"},
				End:     &KubernetesGoFeatureFlagProgressiveRolloutStep{Variation: "enabled", Date: "2026-11-15T09:00:00Z"},
			}}
			input.Spec.Flags["assistant"] = f
			invalid("progressive_rollout.date")
		})

		ginkgo.It("a progressive rollout share above 100 should fail", func() {
			f := onOff()
			f.DefaultRule = &KubernetesGoFeatureFlagRule{ProgressiveRollout: &KubernetesGoFeatureFlagProgressiveRollout{
				Initial: &KubernetesGoFeatureFlagProgressiveRolloutStep{Variation: "disabled", Date: "2026-11-01T09:00:00Z"},
				End:     &KubernetesGoFeatureFlagProgressiveRolloutStep{Variation: "enabled", Percentage: float64Ptr(101), Date: "2026-11-15T09:00:00Z"},
			}}
			input.Spec.Flags["assistant"] = f
			invalid("percentage")
		})

		ginkgo.It("an experimentation window without an end should fail", func() {
			f := onOff()
			f.Experimentation = &KubernetesGoFeatureFlagExperimentation{Start: "2026-11-01T00:00:00Z"}
			input.Spec.Flags["assistant"] = f
			invalid("end")
		})

		ginkgo.It("an experimentation start that is not RFC 3339 should fail", func() {
			f := onOff()
			f.Experimentation = &KubernetesGoFeatureFlagExperimentation{Start: "tomorrow", End: "2026-12-01T00:00:00Z"}
			input.Spec.Flags["assistant"] = f
			invalid("experimentation.start")
		})

		ginkgo.It("a date that does not exist should fail", func() {
			f := onOff()
			f.Experimentation = &KubernetesGoFeatureFlagExperimentation{Start: "2026-02-30T25:61:00Z", End: "2026-12-01T00:00:00Z"}
			input.Spec.Flags["assistant"] = f
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("February 29 in a non-leap year should fail", func() {
			f := onOff()
			f.Experimentation = &KubernetesGoFeatureFlagExperimentation{Start: "2027-02-29T00:00:00Z", End: "2027-12-01T00:00:00Z"}
			input.Spec.Flags["assistant"] = f
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an enabled targeting rule without a query should fail", func() {
			f := onOff()
			f.Targeting[0].Query = ""
			input.Spec.Flags["assistant"] = f
			invalid("flag.targeting.query_required")
		})

		ginkgo.It("a progressive rollout naming an unknown variation should fail", func() {
			f := onOff()
			f.DefaultRule = &KubernetesGoFeatureFlagRule{ProgressiveRollout: &KubernetesGoFeatureFlagProgressiveRollout{
				Initial: &KubernetesGoFeatureFlagProgressiveRolloutStep{Variation: "disabled", Date: "2026-11-01T09:00:00Z"},
				End:     &KubernetesGoFeatureFlagProgressiveRolloutStep{Variation: "on", Percentage: float64Ptr(100), Date: "2026-11-15T09:00:00Z"},
			}}
			input.Spec.Flags["assistant"] = f
			invalid("flag.rules.known_variations")
		})

		ginkgo.It("a progressive rollout that ramps backward should fail", func() {
			f := onOff()
			f.DefaultRule = &KubernetesGoFeatureFlagRule{ProgressiveRollout: &KubernetesGoFeatureFlagProgressiveRollout{
				Initial: &KubernetesGoFeatureFlagProgressiveRolloutStep{Variation: "disabled", Percentage: float64Ptr(50), Date: "2026-11-01T09:00:00Z"},
				End:     &KubernetesGoFeatureFlagProgressiveRolloutStep{Variation: "enabled", Percentage: float64Ptr(30), Date: "2026-11-15T09:00:00Z"},
			}}
			input.Spec.Flags["assistant"] = f
			invalid("progressive_rollout.ramps_forward")
		})

		ginkgo.It("a progressive rollout with one variation at both ends should fail", func() {
			f := onOff()
			f.DefaultRule = &KubernetesGoFeatureFlagRule{ProgressiveRollout: &KubernetesGoFeatureFlagProgressiveRollout{
				Initial: &KubernetesGoFeatureFlagProgressiveRolloutStep{Variation: "enabled", Date: "2026-11-01T09:00:00Z"},
				End:     &KubernetesGoFeatureFlagProgressiveRolloutStep{Variation: "enabled", Percentage: float64Ptr(100), Date: "2026-11-15T09:00:00Z"},
			}}
			input.Spec.Flags["assistant"] = f
			invalid("progressive_rollout.distinct_variations")
		})

		ginkgo.It("an empty end share ramps to 100 and should pass", func() {
			f := onOff()
			f.DefaultRule = &KubernetesGoFeatureFlagRule{ProgressiveRollout: &KubernetesGoFeatureFlagProgressiveRollout{
				Initial: &KubernetesGoFeatureFlagProgressiveRolloutStep{Variation: "disabled", Percentage: float64Ptr(50), Date: "2026-11-01T09:00:00Z"},
				End:     &KubernetesGoFeatureFlagProgressiveRolloutStep{Variation: "enabled", Date: "2026-11-15T09:00:00Z"},
			}}
			input.Spec.Flags["assistant"] = f
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an experimentation end that is not a date should fail", func() {
			f := onOff()
			f.Experimentation = &KubernetesGoFeatureFlagExperimentation{Start: "2026-11-01T00:00:00Z", End: "next week"}
			input.Spec.Flags["assistant"] = f
			invalid("experimentation.end")
		})

		ginkgo.It("a scheduled step date that does not exist should fail", func() {
			f := onOff()
			f.ScheduledRollout = []*KubernetesGoFeatureFlagScheduledStep{{Date: "2026-04-31T09:00:00Z", Disable: boolPtr(true)}}
			input.Spec.Flags["assistant"] = f
			invalid("scheduled_step.date")
		})

		ginkgo.It("a progressive rollout ending before it starts should fail", func() {
			f := onOff()
			f.DefaultRule = &KubernetesGoFeatureFlagRule{ProgressiveRollout: &KubernetesGoFeatureFlagProgressiveRollout{
				Initial: &KubernetesGoFeatureFlagProgressiveRolloutStep{Variation: "disabled", Date: "2026-11-15T09:00:00Z"},
				End:     &KubernetesGoFeatureFlagProgressiveRolloutStep{Variation: "enabled", Percentage: float64Ptr(100), Date: "2026-11-01T09:00:00Z"},
			}}
			input.Spec.Flags["assistant"] = f
			invalid("progressive_rollout.end_after_initial")
		})

		ginkgo.It("a disabled default rule should fail", func() {
			f := onOff()
			f.DefaultRule.Disable = true
			input.Spec.Flags["assistant"] = f
			invalid("flag.default_rule.not_disabled")
		})

		ginkgo.It("a disabled targeting rule naming an unknown variation should pass", func() {
			f := onOff()
			f.Targeting = append(f.Targeting, &KubernetesGoFeatureFlagRule{Name: "parked", Variation: "gone", Disable: true})
			input.Spec.Flags["assistant"] = f
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a scheduled step without a date should fail", func() {
			f := onOff()
			f.ScheduledRollout = []*KubernetesGoFeatureFlagScheduledStep{{Disable: boolPtr(true)}}
			input.Spec.Flags["assistant"] = f
			invalid("date")
		})
	})
})
