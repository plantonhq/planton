package kubernetesflagdflagfilev1alpha1

import (
	"errors"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestKubernetesFlagdFlagFile(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesFlagdFlagFile Suite")
}

func strPtr(s string) *string { return &s }

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

func boolValue(b bool) *KubernetesFlagdValue {
	return &KubernetesFlagdValue{Value: &KubernetesFlagdValue_BoolValue{BoolValue: b}}
}

func onOff() *KubernetesFlagdFlag {
	return &KubernetesFlagdFlag{
		State:          "ENABLED",
		Variants:       map[string]*KubernetesFlagdValue{"on": boolValue(true), "off": boolValue(false)},
		DefaultVariant: "off",
		Targeting: mustStruct(map[string]any{"if": []any{
			map[string]any{"in": []any{map[string]any{"var": "org"}, []any{"planton"}}}, "on", "off",
		}}),
	}
}

var _ = ginkgo.Describe("KubernetesFlagdFlagFile Validation Tests", func() {
	var input *KubernetesFlagdFlagFile

	ginkgo.BeforeEach(func() {
		input = &KubernetesFlagdFlagFile{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesFlagdFlagFile",
			Metadata:   &shared.CatalogObjectMetadata{Name: "release-flags"},
			Spec: &KubernetesFlagdFlagFileSpec{
				Namespace: literal("feature-flags"),
				Flags:     map[string]*KubernetesFlagdFlag{"assistant": onOff()},
			},
		}
	})

	invalid := func(want string) {
		gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement(gomega.ContainSubstring(want)))
	}

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("an on/off flag with JSONLogic targeting should be valid", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("evaluators, flag-set metadata, a disabled flag and a code-default flag should be valid", func() {
			input.Spec.Key = strPtr("release.flagd.json")
			input.Spec.Evaluators = map[string]*structpb.Struct{
				"is_staff": mustStruct(map[string]any{"ends_with": []any{map[string]any{"var": "email"}, "@example.com"}}),
			}
			input.Spec.Metadata = mustStruct(map[string]any{"flagSetId": "web"})
			parked := onOff()
			parked.State = "DISABLED"
			input.Spec.Flags["parked"] = parked
			codeDefault := onOff()
			codeDefault.DefaultVariant = ""
			codeDefault.Metadata = mustStruct(map[string]any{"owner": "web"})
			input.Spec.Flags["code-default"] = codeDefault
			input.Spec.Flags["color"] = &KubernetesFlagdFlag{
				State: "ENABLED",
				Variants: map[string]*KubernetesFlagdValue{
					"red":  {Value: &KubernetesFlagdValue_StringValue{StringValue: "#f00"}},
					"blue": {Value: &KubernetesFlagdValue_StringValue{StringValue: "#00f"}},
				},
				DefaultVariant: "red",
				Targeting:      mustStruct(map[string]any{"fractional": []any{[]any{"red", 50}, []any{"blue", 50}}}),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("an empty evaluator should fail", func() {
			input.Spec.Evaluators = map[string]*structpb.Struct{"is_staff": mustStruct(map[string]any{})}
			invalid("spec.evaluators.non_empty")
		})

		ginkgo.It("flag-set metadata holding an object should fail", func() {
			input.Spec.Metadata = mustStruct(map[string]any{"owner": map[string]any{"team": "web"}})
			invalid("spec.metadata.scalars")
		})

		ginkgo.It("flag metadata holding a list should fail", func() {
			f := onOff()
			f.Metadata = mustStruct(map[string]any{"tags": []any{"a", "b"}})
			input.Spec.Flags["assistant"] = f
			invalid("flag.metadata.scalars")
		})

		ginkgo.It("an empty variant name should fail", func() {
			f := onOff()
			f.Variants[""] = boolValue(true)
			input.Spec.Flags["assistant"] = f
			invalid("flag.variants.names")
		})

		ginkgo.It("a missing namespace should fail", func() {
			input.Spec.Namespace = nil
			invalid("namespace")
		})

		ginkgo.It("a key without the .json extension should fail", func() {
			input.Spec.Key = strPtr("flags.yaml")
			invalid("spec.key")
		})

		ginkgo.It("a flag key with whitespace should fail", func() {
			input.Spec.Flags["new banner"] = onOff()
			invalid("spec.flags.names")
		})

		ginkgo.It("an unknown state should fail", func() {
			f := onOff()
			f.State = "ON"
			input.Spec.Flags["assistant"] = f
			invalid("flag.state")
		})

		ginkgo.It("a flag without variants should fail", func() {
			f := onOff()
			f.Variants = nil
			f.DefaultVariant = ""
			input.Spec.Flags["assistant"] = f
			invalid("variants")
		})

		ginkgo.It("variants of mixed types should fail", func() {
			f := onOff()
			f.Variants["off"] = &KubernetesFlagdValue{Value: &KubernetesFlagdValue_StringValue{StringValue: "off"}}
			input.Spec.Flags["assistant"] = f
			invalid("flag.variants.one_type")
		})

		ginkgo.It("a default variant the flag does not define should fail", func() {
			f := onOff()
			f.DefaultVariant = "disabled"
			input.Spec.Flags["assistant"] = f
			invalid("flag.default_variant.known")
		})
	})
})
