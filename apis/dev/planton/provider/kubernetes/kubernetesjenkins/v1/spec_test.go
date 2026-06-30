package kubernetesjenkinsv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestKubernetesJenkins(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesJenkins Suite")
}

var _ = ginkgo.Describe("KubernetesJenkins Custom Validation Tests", func() {
	var input *KubernetesJenkins

	ginkgo.BeforeEach(func() {
		input = &KubernetesJenkins{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesJenkins",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-jenkins",
			},
			Spec: &KubernetesJenkinsSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-k8s-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "jenkins",
					},
				},
				ContainerResources: &kubernetes.ContainerResources{
					Limits: &kubernetes.CpuMemory{
						Cpu:    "1000m",
						Memory: "1Gi",
					},
					Requests: &kubernetes.CpuMemory{
						Cpu:    "50m",
						Memory: "100Mi",
					},
				},
				Ingress: &KubernetesJenkinsIngress{
					Enabled:  true,
					Hostname: "jenkins.example.com",
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("jenkins_kubernetes", func() {
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
