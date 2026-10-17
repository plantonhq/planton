package kubernetesagentsandboxv1alpha1

import (
	"errors"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

// The validation suite pins what the spec refuses before the API server sees
// the Sandbox: the envelope (a namespace, label grammar), the pod template
// (required, with at least one container), the volume claim templates (named,
// uniquely), the expiry's RFC 3339 form, and the closed sets of the shutdown
// policy and the operating mode. The pod's own rules are the core/v1
// library's, pinned in catalog/kubernetes/core_v1_api_test.go.

func TestKubernetesAgentSandbox(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesAgentSandbox Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value}}
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

func workspace(name string) *kubernetes.KubernetesAgentSandboxApiPersistentVolumeClaimTemplate {
	return &kubernetes.KubernetesAgentSandboxApiPersistentVolumeClaimTemplate{
		Metadata: &kubernetes.KubernetesAgentSandboxApiEmbeddedObjectMetadata{Name: name},
		Spec: &kubernetes.KubernetesCoreV1ApiPersistentVolumeClaimSpec{
			AccessModes: []string{"ReadWriteOnce"},
			Resources:   &kubernetes.KubernetesCoreV1ApiVolumeResourceRequirements{Requests: map[string]string{"storage": "1Gi"}},
		},
	}
}

var _ = ginkgo.Describe("KubernetesAgentSandbox Validation Tests", func() {
	var input *KubernetesAgentSandbox

	ginkgo.BeforeEach(func() {
		input = &KubernetesAgentSandbox{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesAgentSandbox",
			Metadata:   &shared.CatalogObjectMetadata{Name: "agent-workbench"},
			Spec: &KubernetesAgentSandboxSpec{
				Namespace: literal("agents"),
				PodTemplate: &kubernetes.KubernetesAgentSandboxApiPodTemplate{
					Spec: &kubernetes.KubernetesCoreV1ApiPodSpec{
						RuntimeClassName: ptr("gvisor"),
						Containers: []*kubernetes.KubernetesCoreV1ApiContainer{{
							Name:  "agent",
							Image: "us-docker.pkg.dev/example/agents/workbench:1.4.2",
							VolumeMounts: []*kubernetes.KubernetesCoreV1ApiVolumeMount{{
								Name: "workspace", MountPath: "/workspace",
							}},
						}},
					},
				},
				VolumeClaimTemplates: []*kubernetes.KubernetesAgentSandboxApiPersistentVolumeClaimTemplate{workspace("workspace")},
				Service:              ptr(true),
				ShutdownTime:         ptr("2099-01-01T00:00:00Z"),
				ShutdownPolicy:       ptr("Retain"),
				OperatingMode:        ptr("Suspended"),
			},
		}
	})

	ginkgo.It("accepts a sandbox a person would write", func() {
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("accepts the smallest sandbox: a namespace and one container", func() {
		input.Spec = &KubernetesAgentSandboxSpec{
			Namespace: literal("agents"),
			PodTemplate: &kubernetes.KubernetesAgentSandboxApiPodTemplate{Spec: &kubernetes.KubernetesCoreV1ApiPodSpec{
				Containers: []*kubernetes.KubernetesCoreV1ApiContainer{{Name: "agent", Image: "busybox:1.37"}},
			}},
		}
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("requires a namespace", func() {
		input.Spec.Namespace = nil
		expectViolation(protovalidate.Validate(input), "required")
	})

	ginkgo.It("requires a pod template", func() {
		input.Spec.PodTemplate = nil
		expectViolation(protovalidate.Validate(input), "required")
	})

	ginkgo.It("requires a pod specification inside the template", func() {
		input.Spec.PodTemplate.Spec = nil
		expectViolation(protovalidate.Validate(input), "required")
	})

	ginkgo.It("requires at least one container", func() {
		input.Spec.PodTemplate.Spec.Containers = nil
		expectViolation(protovalidate.Validate(input), "repeated.min_items")
	})

	ginkgo.It("refuses an invalid label key on the object", func() {
		input.Spec.Labels = map[string]string{"-bad": "x"}
		expectViolation(protovalidate.Validate(input), "string.pattern")
	})

	ginkgo.It("refuses an invalid label on the pod", func() {
		input.Spec.PodTemplate.Metadata = &kubernetes.KubernetesAgentSandboxApiPodMetadata{Labels: map[string]string{"app": "not a value"}}
		expectViolation(protovalidate.Validate(input), "string.pattern")
	})

	ginkgo.It("requires each volume claim template to be named", func() {
		input.Spec.VolumeClaimTemplates[0].Metadata.Name = ""
		expectViolation(protovalidate.Validate(input), "string.min_len")
	})

	ginkgo.It("requires a volume claim template's name to be a DNS label", func() {
		input.Spec.VolumeClaimTemplates[0].Metadata.Name = "Workspace_1"
		expectViolation(protovalidate.Validate(input), "string.pattern")
	})

	ginkgo.It("requires a volume claim template's specification", func() {
		input.Spec.VolumeClaimTemplates[0].Spec = nil
		expectViolation(protovalidate.Validate(input), "required")
	})

	ginkgo.It("refuses two volume claim templates with one name", func() {
		input.Spec.VolumeClaimTemplates = append(input.Spec.VolumeClaimTemplates, workspace("workspace"))
		expectViolation(protovalidate.Validate(input), "volume_claim_templates.unique_names")
	})

	ginkgo.It("refuses an expiry that is not an RFC 3339 time", func() {
		input.Spec.ShutdownTime = ptr("tomorrow at noon")
		expectViolation(protovalidate.Validate(input), "string.pattern")
	})

	ginkgo.It("accepts an expiry with fractional seconds and an offset", func() {
		input.Spec.ShutdownTime = ptr("2099-01-01T05:30:00.5+05:30")
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("refuses an unknown shutdown policy", func() {
		input.Spec.ShutdownPolicy = ptr("Archive")
		expectViolation(protovalidate.Validate(input), "string.in")
	})

	ginkgo.It("refuses an unknown operating mode", func() {
		input.Spec.OperatingMode = ptr("Paused")
		expectViolation(protovalidate.Validate(input), "string.in")
	})
})
