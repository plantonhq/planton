package kubernetesagentsubstrateworkerpoolv1alpha1

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
// the WorkerPool: the envelope (a namespace, label grammar), a size that is
// always written (zero included), a worker image by reference or literal,
// exactly one sandbox class from the closed set, its optional SandboxConfig by
// reference or literal (a name of at most 253 characters), every reference
// non-empty when given, and the worker pod template's bounds (64 labels and
// annotations with Kubernetes' key and value grammar, none in the ate.dev
// domains Substrate reserves, 16 tolerations).

func TestKubernetesAgentSubstrateWorkerPool(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesAgentSubstrateWorkerPool Suite")
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

var _ = ginkgo.Describe("KubernetesAgentSubstrateWorkerPool Validation Tests", func() {
	var input *KubernetesAgentSubstrateWorkerPool

	ginkgo.BeforeEach(func() {
		input = &KubernetesAgentSubstrateWorkerPool{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesAgentSubstrateWorkerPool",
			Metadata:   &shared.CatalogObjectMetadata{Name: "acme-workers"},
			Spec: &KubernetesAgentSubstrateWorkerPoolSpec{
				Namespace:   literal("acme-agents"),
				Labels:      map[string]string{"tenant": "acme"},
				Replicas:    ptr(int32(3)),
				WorkerImage: valueFrom(catalogkind.CatalogKind_KubernetesAgentSubstrate, "substrate", "status.outputs.gvisor_worker_image"),
				Template: &KubernetesAgentSubstrateWorkerPoolPodTemplate{
					Labels:       map[string]string{"team": "platform"},
					NodeSelector: map[string]string{"pool": "agents"},
					Tolerations: []*kubernetes.KubernetesCoreV1ApiToleration{
						{Key: "tenant", Operator: ptr("Equal"), Value: "acme", Effect: ptr("NoSchedule")},
					},
					Resources: &kubernetes.KubernetesCoreV1ApiResourceRequirements{
						Limits:   map[string]string{"cpu": "1", "memory": "1Gi"},
						Requests: map[string]string{"cpu": "250m", "memory": "512Mi"},
					},
					ServiceAccountName: valueFrom(catalogkind.CatalogKind_KubernetesServiceAccount, "workers", "status.outputs.service_account_name"),
				},
				SandboxClasses: []*KubernetesAgentSubstrateWorkerPoolSandboxClass{{
					Name: "gvisor",
					ConfigRef: &KubernetesAgentSubstrateWorkerPoolSandboxConfigRef{
						Name: valueFrom(catalogkind.CatalogKind_KubernetesAgentSubstrateSandboxConfig, "gvisor-pinned", "status.outputs.sandbox_config_name"),
					},
				}},
			},
		}
	})

	ginkgo.It("accepts a pool a person would write", func() {
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("accepts a pool of zero workers", func() {
		input.Spec.Replicas = ptr(int32(0))
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("accepts a pool on the class default, with a literal image and no template", func() {
		input.Spec.SandboxClasses[0].ConfigRef = nil
		input.Spec.WorkerImage = literal("ghcr.io/plantonhq/planton/agent-substrate/ateworker-gvisor:v0.6.0@sha256:e3f59217a1e63c280ea1cc4149dfe0686353f9ef583fe2ff1c129d4ad42f8cba")
		input.Spec.Template = nil
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("accepts a microVM pool", func() {
		input.Spec.SandboxClasses[0].Name = "microvm"
		input.Spec.WorkerImage = valueFrom(catalogkind.CatalogKind_KubernetesAgentSubstrate, "substrate", "status.outputs.microvm_worker_image")
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("requires a namespace", func() {
		input.Spec.Namespace = nil
		expectViolation(protovalidate.Validate(input), "required")
	})

	ginkgo.It("requires the size, so every apply writes it", func() {
		input.Spec.Replicas = nil
		expectViolation(protovalidate.Validate(input), "required")
	})

	ginkgo.It("refuses a negative size", func() {
		input.Spec.Replicas = ptr(int32(-1))
		expectViolation(protovalidate.Validate(input), "int32.gte")
	})

	ginkgo.It("requires a worker image", func() {
		input.Spec.WorkerImage = nil
		expectViolation(protovalidate.Validate(input), "required")
	})

	ginkgo.It("requires exactly one sandbox class", func() {
		input.Spec.SandboxClasses = nil
		expectViolation(protovalidate.Validate(input), "repeated.min_items")
		input.Spec.SandboxClasses = []*KubernetesAgentSubstrateWorkerPoolSandboxClass{{Name: "gvisor"}, {Name: "microvm"}}
		expectViolation(protovalidate.Validate(input), "repeated.max_items")
	})

	ginkgo.It("refuses a class outside gvisor and microvm", func() {
		input.Spec.SandboxClasses[0].Name = "kata"
		expectViolation(protovalidate.Validate(input), "string.in")
	})

	ginkgo.It("requires a config reference's name when one is given", func() {
		input.Spec.SandboxClasses[0].ConfigRef = &KubernetesAgentSubstrateWorkerPoolSandboxConfigRef{}
		expectViolation(protovalidate.Validate(input), "required")
	})

	ginkgo.It("refuses worker pod labels in the ate.dev domains", func() {
		input.Spec.Template.Labels = map[string]string{"ate.dev/worker-pool": "x"}
		expectViolation(protovalidate.Validate(input), "template.labels.reserved")
		input.Spec.Template.Labels = map[string]string{"pods.ate.dev/role": "x"}
		expectViolation(protovalidate.Validate(input), "template.labels.reserved")
	})

	ginkgo.It("refuses worker pod annotations in the ate.dev domains", func() {
		input.Spec.Template.Annotations = map[string]string{"sandboxconfig.ate.dev/x": "y"}
		expectViolation(protovalidate.Validate(input), "template.annotations.reserved")
	})

	ginkgo.It("refuses more than 64 worker pod labels", func() {
		labels := map[string]string{}
		for i := 0; i < 65; i++ {
			labels["k"+string(rune('a'+i%26))+string(rune('a'+i/26))] = "v"
		}
		input.Spec.Template.Labels = labels
		expectViolation(protovalidate.Validate(input), "map.max_pairs")
	})

	ginkgo.It("refuses more than 16 tolerations", func() {
		var tolerations []*kubernetes.KubernetesCoreV1ApiToleration
		for i := 0; i < 17; i++ {
			tolerations = append(tolerations, &kubernetes.KubernetesCoreV1ApiToleration{Key: "k", Operator: ptr("Exists")})
		}
		input.Spec.Template.Tolerations = tolerations
		expectViolation(protovalidate.Validate(input), "repeated.max_items")
	})

	ginkgo.It("refuses more than 64 worker pod annotations", func() {
		annotations := map[string]string{}
		for i := 0; i < 65; i++ {
			annotations["k"+string(rune('a'+i%26))+string(rune('a'+i/26))] = "v"
		}
		input.Spec.Template.Annotations = annotations
		expectViolation(protovalidate.Validate(input), "map.max_pairs")
	})

	ginkgo.It("refuses worker pod label keys and values outside Kubernetes' grammar", func() {
		input.Spec.Template.Labels = map[string]string{"Team.Example/role": "x"}
		expectViolation(protovalidate.Validate(input), "string.pattern")
		input.Spec.Template.Labels = map[string]string{"team": "-platform"}
		expectViolation(protovalidate.Validate(input), "string.pattern")
	})

	ginkgo.It("refuses a worker pod annotation key outside Kubernetes' grammar", func() {
		input.Spec.Template.Annotations = map[string]string{"not a key": "y"}
		expectViolation(protovalidate.Validate(input), "string.pattern")
	})

	ginkgo.It("refuses an invalid pool label key", func() {
		input.Spec.Labels = map[string]string{"tenant/": "acme"}
		expectViolation(protovalidate.Validate(input), "string.pattern")
	})

	ginkgo.It("refuses an empty class name", func() {
		input.Spec.SandboxClasses[0].Name = ""
		expectViolation(protovalidate.Validate(input), "required")
	})

	ginkgo.It("refuses a SandboxConfig name longer than 253 characters", func() {
		long := make([]byte, 254)
		for i := range long {
			long[i] = 'a'
		}
		input.Spec.SandboxClasses[0].ConfigRef.Name = literal(string(long))
		expectViolation(protovalidate.Validate(input), "config_ref.name.max_len")
		input.Spec.SandboxClasses[0].ConfigRef.Name = literal(string(long[:253]))
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("accepts a PriorityClass by reference or by literal name", func() {
		input.Spec.Template.PriorityClassName = valueFrom(catalogkind.CatalogKind_KubernetesPriorityClass, "agents", "status.outputs.priority_class_name")
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		input.Spec.Template.PriorityClassName = literal("system-node-critical")
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("refuses an empty literal wherever a reference or a literal goes", func() {
		input.Spec.WorkerImage = literal("")
		expectViolation(protovalidate.Validate(input), "string_value_or_ref.non_empty")
		input.Spec.WorkerImage = literal("ghcr.io/plantonhq/planton/agent-substrate/ateworker-gvisor:v0.6.0")
		input.Spec.SandboxClasses[0].ConfigRef.Name = literal("")
		expectViolation(protovalidate.Validate(input), "string_value_or_ref.non_empty")
		input.Spec.SandboxClasses[0].ConfigRef = nil
		input.Spec.Template.ServiceAccountName = literal("")
		expectViolation(protovalidate.Validate(input), "string_value_or_ref.non_empty")
		input.Spec.Template.ServiceAccountName = nil
		input.Spec.Template.PriorityClassName = literal("")
		expectViolation(protovalidate.Validate(input), "string_value_or_ref.non_empty")
	})

	ginkgo.It("refuses an invalid pool label value", func() {
		input.Spec.Labels = map[string]string{"tenant": "not valid"}
		expectViolation(protovalidate.Validate(input), "string.pattern")
	})
})
