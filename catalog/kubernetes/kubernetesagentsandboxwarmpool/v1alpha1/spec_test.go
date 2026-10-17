package kubernetesagentsandboxwarmpoolv1alpha1

import (
	"errors"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

// The validation suite pins what the spec refuses before the API server sees
// the SandboxWarmPool: the envelope (a namespace, label grammar), the template
// reference (required, literal or by reference), a non-negative size (unset
// left to the cluster, so an autoscaler can own it), and the closed set of
// update strategies.

func TestKubernetesAgentSandboxWarmPool(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesAgentSandboxWarmPool Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value}}
}

func valueFrom(kind catalogkind.CatalogKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Kind: kind, Name: name, FieldPath: fieldPath},
		},
	}
}

func ptr[T any](v T) *T { return &v }

func expectViolation(err error, ruleID string) {
	gomega.Expect(err).NotTo(gomega.BeNil())
	var validationErr *protovalidate.ValidationError
	gomega.Expect(errors.As(err, &validationErr)).To(gomega.BeTrue(), "not a validation error: %v", err)
	var ids []string
	for _, violation := range validationErr.Violations {
		ids = append(ids, violation.Proto.GetRuleId())
	}
	gomega.Expect(ids).To(gomega.ContainElement(ruleID), "violations: %v", err)
}

var _ = ginkgo.Describe("KubernetesAgentSandboxWarmPool Validation Tests", func() {
	var input *KubernetesAgentSandboxWarmPool

	ginkgo.BeforeEach(func() {
		input = &KubernetesAgentSandboxWarmPool{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesAgentSandboxWarmPool",
			Metadata:   &shared.CatalogObjectMetadata{Name: "coding-agent-pool"},
			Spec: &KubernetesAgentSandboxWarmPoolSpec{
				Namespace: literal("agents"),
				Replicas:  ptr(int32(3)),
				SandboxTemplateRef: &KubernetesAgentSandboxWarmPoolTemplateRef{
					Name: valueFrom(catalogkind.CatalogKind_KubernetesAgentSandboxTemplate, "coding-agent", "status.outputs.sandbox_template_name"),
				},
				UpdateStrategy: &KubernetesAgentSandboxWarmPoolUpdateStrategy{Type: ptr("Recreate")},
			},
		}
	})

	ginkgo.It("accepts a pool a person would write", func() {
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("accepts a pool whose size an autoscaler owns (replicas unset)", func() {
		input.Spec.Replicas = nil
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("accepts an empty pool", func() {
		input.Spec.Replicas = ptr(int32(0))
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("accepts the template by literal name", func() {
		input.Spec.SandboxTemplateRef.Name = literal("coding-agent")
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("requires a namespace", func() {
		input.Spec.Namespace = nil
		expectViolation(protovalidate.Validate(input), "required")
	})

	ginkgo.It("requires a template reference", func() {
		input.Spec.SandboxTemplateRef = nil
		expectViolation(protovalidate.Validate(input), "required")
	})

	ginkgo.It("requires the template's name", func() {
		input.Spec.SandboxTemplateRef.Name = nil
		expectViolation(protovalidate.Validate(input), "required")
	})

	ginkgo.It("refuses a negative size", func() {
		input.Spec.Replicas = ptr(int32(-1))
		expectViolation(protovalidate.Validate(input), "int32.gte")
	})

	ginkgo.It("refuses an unknown update strategy", func() {
		input.Spec.UpdateStrategy.Type = ptr("RollingUpdate")
		expectViolation(protovalidate.Validate(input), "string.in")
	})

	ginkgo.It("refuses an invalid label key on the object", func() {
		input.Spec.Labels = map[string]string{"-bad": "x"}
		expectViolation(protovalidate.Validate(input), "string.pattern")
	})
})
