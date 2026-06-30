package kuberneteskeycloakv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	cloudresourcekind "github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestKubernetesKeycloak(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesKeycloak Suite")
}

var _ = ginkgo.Describe("KubernetesKeycloak Custom Validation Tests", func() {
	var input *KubernetesKeycloak

	ginkgo.BeforeEach(func() {
		input = &KubernetesKeycloak{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesKeycloak",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-keycloak",
			},
			Spec: &KubernetesKeycloakSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "test-namespace",
					},
				},
				Container: &KubernetesKeycloakContainer{
					Resources: &kubernetes.ContainerResources{
						Limits: &kubernetes.CpuMemory{
							Cpu:    "1000m",
							Memory: "1Gi",
						},
						Requests: &kubernetes.CpuMemory{
							Cpu:    "50m",
							Memory: "100Mi",
						},
					},
				},
				Ingress: &KubernetesKeycloakIngress{
					Enabled:  true,
					Hostname: "keycloak.example.com",
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("keycloak_kubernetes", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Ingress validation", func() {
		ginkgo.Context("When ingress is enabled without hostname", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Ingress.Hostname = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("When ingress is disabled", func() {
			ginkgo.It("should not require hostname", func() {
				input.Spec.Ingress.Enabled = false
				input.Spec.Ingress.Hostname = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
