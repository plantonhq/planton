package kuberneteskeycloakoperatorv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesKeycloakOperator(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesKeycloakOperator Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func valueFrom(kind catalogkind.CatalogKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Kind:      kind,
				Name:      name,
				FieldPath: fieldPath,
			},
		},
	}
}

var _ = ginkgo.Describe("KubernetesKeycloakOperator Validation Tests", func() {
	var input *KubernetesKeycloakOperator

	ginkgo.BeforeEach(func() {
		input = &KubernetesKeycloakOperator{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesKeycloakOperator",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "keycloak-operator",
			},
			Spec: &KubernetesKeycloakOperatorSpec{
				Namespace: literal("keycloak"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("a minimal spec (namespace only) should be valid", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "keycloak", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a maximal spec (every field populated) should be valid", func() {
			input.Spec.CreateNamespace = true
			input.Spec.ClusterWide = true
			input.Spec.OperatorImage = "mirror.example.com/keycloak/keycloak-operator:26.0.7"
			input.Spec.DefaultKeycloakImage = "mirror.example.com/keycloak/keycloak:26.0.7"
			input.Spec.Resources = &kubernetes.ContainerResources{
				Requests: &kubernetes.CpuMemory{Cpu: "300m", Memory: "450Mi"},
				Limits:   &kubernetes.CpuMemory{Cpu: "700m", Memory: "450Mi"},
			}
			input.Spec.Scheduling = &KubernetesKeycloakOperatorScheduling{
				NodeSelector: map[string]string{"workload": "platform"},
				Tolerations: []*kubernetes.WorkloadToleration{
					{Key: "platform", Operator: "Equal", Value: "true", Effect: "NoSchedule"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a missing namespace should be invalid", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a missing spec should be invalid", func() {
			input.Spec = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
