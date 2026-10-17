package kubernetesagentsandboxtemplatev1alpha1

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
// the SandboxTemplate: the envelope (a namespace, label grammar), the pod
// template (required, with at least one container), the volume claim
// templates (named, uniquely), the closed sets of the policy management and
// the two claim policies, and the NetworkPolicy combination rules (an IP block
// stands alone; an end port needs a numeric port). The pod's own rules are the
// core/v1 library's, pinned in catalog/kubernetes/core_v1_api_test.go.

func TestKubernetesAgentSandboxTemplate(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesAgentSandboxTemplate Suite")
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

var _ = ginkgo.Describe("KubernetesAgentSandboxTemplate Validation Tests", func() {
	var input *KubernetesAgentSandboxTemplate

	routerPeer := func() *kubernetes.KubernetesNetworkingV1ApiNetworkPolicyPeer {
		return &kubernetes.KubernetesNetworkingV1ApiNetworkPolicyPeer{
			NamespaceSelector: &kubernetes.KubernetesCoreV1ApiLabelSelector{MatchLabels: map[string]string{"kubernetes.io/metadata.name": "agent-sandbox-system"}},
			PodSelector:       &kubernetes.KubernetesCoreV1ApiLabelSelector{MatchLabels: map[string]string{"app": "sandbox-router"}},
		}
	}

	ginkgo.BeforeEach(func() {
		input = &KubernetesAgentSandboxTemplate{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesAgentSandboxTemplate",
			Metadata:   &shared.CatalogObjectMetadata{Name: "coding-agent"},
			Spec: &KubernetesAgentSandboxTemplateSpec{
				Namespace: literal("agents"),
				PodTemplate: &kubernetes.KubernetesAgentSandboxApiPodTemplate{
					Spec: &kubernetes.KubernetesCoreV1ApiPodSpec{
						RuntimeClassName: ptr("gvisor"),
						Containers: []*kubernetes.KubernetesCoreV1ApiContainer{{
							Name:  "agent",
							Image: "us-docker.pkg.dev/example/agents/coding-agent:2.3.0",
						}},
					},
				},
				VolumeClaimTemplates: []*kubernetes.KubernetesAgentSandboxApiPersistentVolumeClaimTemplate{workspace("workspace")},
				NetworkPolicy: &KubernetesAgentSandboxTemplateNetworkPolicy{
					Ingress: []*kubernetes.KubernetesNetworkingV1ApiNetworkPolicyIngressRule{{
						From:  []*kubernetes.KubernetesNetworkingV1ApiNetworkPolicyPeer{routerPeer()},
						Ports: []*kubernetes.KubernetesNetworkingV1ApiNetworkPolicyPort{{Protocol: ptr("TCP"), Port: ptr("8888")}},
					}},
					Egress: []*kubernetes.KubernetesNetworkingV1ApiNetworkPolicyEgressRule{{
						To: []*kubernetes.KubernetesNetworkingV1ApiNetworkPolicyPeer{{
							IpBlock: &kubernetes.KubernetesNetworkingV1ApiIPBlock{Cidr: "0.0.0.0/0", Except: []string{"10.0.0.0/8"}},
						}},
						Ports: []*kubernetes.KubernetesNetworkingV1ApiNetworkPolicyPort{{Port: ptr("8000"), EndPort: ptr(int32(8100))}},
					}},
				},
				NetworkPolicyManagement:    ptr("Managed"),
				EnvVarsInjectionPolicy:     ptr("Allowed"),
				VolumeClaimTemplatesPolicy: ptr("Overrides"),
			},
		}
	})

	ginkgo.It("accepts a template a person would write", func() {
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("accepts an empty network policy, which denies everything both ways", func() {
		input.Spec.NetworkPolicy = &KubernetesAgentSandboxTemplateNetworkPolicy{}
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

	ginkgo.It("requires at least one container", func() {
		input.Spec.PodTemplate.Spec.Containers = nil
		expectViolation(protovalidate.Validate(input), "repeated.min_items")
	})

	ginkgo.It("refuses an invalid label key on the object", func() {
		input.Spec.Labels = map[string]string{"-bad": "x"}
		expectViolation(protovalidate.Validate(input), "string.pattern")
	})

	ginkgo.It("refuses two volume claim templates with one name", func() {
		input.Spec.VolumeClaimTemplates = append(input.Spec.VolumeClaimTemplates, workspace("workspace"))
		expectViolation(protovalidate.Validate(input), "volume_claim_templates.unique_names")
	})

	ginkgo.It("refuses an unknown policy management", func() {
		input.Spec.NetworkPolicyManagement = ptr("Cilium")
		expectViolation(protovalidate.Validate(input), "string.in")
	})

	ginkgo.It("refuses an unknown env injection policy", func() {
		input.Spec.EnvVarsInjectionPolicy = ptr("Merge")
		expectViolation(protovalidate.Validate(input), "string.in")
	})

	ginkgo.It("refuses an unknown volume claim templates policy", func() {
		input.Spec.VolumeClaimTemplatesPolicy = ptr("Replace")
		expectViolation(protovalidate.Validate(input), "string.in")
	})

	ginkgo.It("refuses an IP block combined with a selector in one peer", func() {
		input.Spec.NetworkPolicy.Egress[0].To[0].PodSelector = &kubernetes.KubernetesCoreV1ApiLabelSelector{}
		expectViolation(protovalidate.Validate(input), "network_policy_peer.ip_block_alone")
	})

	ginkgo.It("requires an IP block's CIDR", func() {
		input.Spec.NetworkPolicy.Egress[0].To[0].IpBlock.Cidr = ""
		expectViolation(protovalidate.Validate(input), "string.min_len")
	})

	ginkgo.It("refuses an end port with a named port", func() {
		input.Spec.NetworkPolicy.Egress[0].Ports[0].Port = ptr("http")
		expectViolation(protovalidate.Validate(input), "network_policy_port.end_port_range")
	})

	ginkgo.It("refuses an end port below its port", func() {
		input.Spec.NetworkPolicy.Egress[0].Ports[0].EndPort = ptr(int32(7000))
		expectViolation(protovalidate.Validate(input), "network_policy_port.end_port_range")
	})

	ginkgo.It("refuses a policy port out of range", func() {
		input.Spec.NetworkPolicy.Ingress[0].Ports[0].Port = ptr("70000")
		expectViolation(protovalidate.Validate(input), "network_policy_port.port")
	})

	ginkgo.It("refuses an unknown protocol", func() {
		input.Spec.NetworkPolicy.Ingress[0].Ports[0].Protocol = ptr("ICMP")
		expectViolation(protovalidate.Validate(input), "string.in")
	})
})
