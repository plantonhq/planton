package kubernetesagentsandboxclaimv1alpha1

import (
	"errors"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

// The validation suite pins what the spec refuses before the API server sees
// the SandboxClaim: the envelope (a namespace, label grammar), the pool
// reference, the lifecycle (an RFC 3339 expiry, a non-negative time to live,
// the closed set of shutdown policies), the environment variables (named as
// Kubernetes allows, valued, unique per container), unique volume claim
// templates, and the pod labels' grammar.

func TestKubernetesAgentSandboxClaim(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesAgentSandboxClaim Suite")
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

var _ = ginkgo.Describe("KubernetesAgentSandboxClaim Validation Tests", func() {
	var input *KubernetesAgentSandboxClaim

	ginkgo.BeforeEach(func() {
		input = &KubernetesAgentSandboxClaim{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesAgentSandboxClaim",
			Metadata:   &shared.CatalogObjectMetadata{Name: "alice-workbench"},
			Spec: &KubernetesAgentSandboxClaimSpec{
				Namespace: literal("agents"),
				WarmPoolRef: &KubernetesAgentSandboxClaimWarmPoolRef{
					Name: valueFrom(catalogkind.CatalogKind_KubernetesAgentSandboxWarmPool, "coding-agent-pool", "status.outputs.sandbox_warm_pool_name"),
				},
				Lifecycle: &KubernetesAgentSandboxClaimLifecycle{
					ShutdownTime:            ptr("2099-01-01T00:00:00Z"),
					TtlSecondsAfterFinished: ptr(int32(3600)),
					ShutdownPolicy:          ptr("Retain"),
				},
				AdditionalPodMetadata: &kubernetes.KubernetesAgentSandboxApiPodMetadata{
					Labels: map[string]string{"sandbox.users.io/user": "alice"},
				},
				Env: []*KubernetesAgentSandboxClaimEnvVar{
					{Name: "TASK_ID", Value: "task-4217", ContainerName: "agent"},
					{Name: "MODEL", Value: "frontier-large"},
				},
				VolumeClaimTemplates: []*kubernetes.KubernetesAgentSandboxApiPersistentVolumeClaimTemplate{{
					Metadata: &kubernetes.KubernetesAgentSandboxApiEmbeddedObjectMetadata{Name: "workspace"},
					Spec: &kubernetes.KubernetesCoreV1ApiPersistentVolumeClaimSpec{
						AccessModes: []string{"ReadWriteOnce"},
						Resources:   &kubernetes.KubernetesCoreV1ApiVolumeResourceRequirements{Requests: map[string]string{"storage": "50Gi"}},
					},
				}},
			},
		}
	})

	ginkgo.It("accepts a claim a person would write", func() {
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("accepts the smallest claim: a namespace and a pool", func() {
		input.Spec = &KubernetesAgentSandboxClaimSpec{
			Namespace:   literal("agents"),
			WarmPoolRef: &KubernetesAgentSandboxClaimWarmPoolRef{Name: literal("coding-agent-pool")},
		}
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("accepts an environment variable name with dots and dashes, as Kubernetes does", func() {
		input.Spec.Env[0].Name = "app.config-path"
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("requires a namespace", func() {
		input.Spec.Namespace = nil
		expectViolation(protovalidate.Validate(input), "required")
	})

	ginkgo.It("requires a warm pool reference", func() {
		input.Spec.WarmPoolRef = nil
		expectViolation(protovalidate.Validate(input), "required")
	})

	ginkgo.It("requires the pool's name", func() {
		input.Spec.WarmPoolRef.Name = nil
		expectViolation(protovalidate.Validate(input), "required")
	})

	ginkgo.It("refuses an expiry that is not an RFC 3339 time", func() {
		input.Spec.Lifecycle.ShutdownTime = ptr("next friday")
		expectViolation(protovalidate.Validate(input), "string.pattern")
	})

	ginkgo.It("refuses a negative time to live", func() {
		input.Spec.Lifecycle.TtlSecondsAfterFinished = ptr(int32(-1))
		expectViolation(protovalidate.Validate(input), "int32.gte")
	})

	ginkgo.It("refuses an unknown shutdown policy", func() {
		input.Spec.Lifecycle.ShutdownPolicy = ptr("Archive")
		expectViolation(protovalidate.Validate(input), "string.in")
	})

	ginkgo.It("requires an environment variable's name", func() {
		input.Spec.Env[0].Name = ""
		expectViolation(protovalidate.Validate(input), "string.min_len")
	})

	ginkgo.It("refuses an environment variable name containing =", func() {
		input.Spec.Env[0].Name = "A=B"
		expectViolation(protovalidate.Validate(input), "string.pattern")
	})

	ginkgo.It("requires an environment variable's value", func() {
		input.Spec.Env[0].Value = ""
		expectViolation(protovalidate.Validate(input), "string.min_len")
	})

	ginkgo.It("refuses one variable twice for one container", func() {
		input.Spec.Env[1] = &KubernetesAgentSandboxClaimEnvVar{Name: "TASK_ID", Value: "again", ContainerName: "agent"}
		expectViolation(protovalidate.Validate(input), "env.unique_names")
	})

	ginkgo.It("refuses two volume claim templates with one name", func() {
		input.Spec.VolumeClaimTemplates = append(input.Spec.VolumeClaimTemplates, input.Spec.VolumeClaimTemplates[0])
		expectViolation(protovalidate.Validate(input), "volume_claim_templates.unique_names")
	})

	ginkgo.It("refuses an invalid pod label", func() {
		input.Spec.AdditionalPodMetadata.Labels = map[string]string{"sandbox.users.io/user": "not valid"}
		expectViolation(protovalidate.Validate(input), "string.pattern")
	})
})
