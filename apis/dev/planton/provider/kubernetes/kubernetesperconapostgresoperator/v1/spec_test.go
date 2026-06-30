package kubernetesperconapostgresoperatorv1

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

func TestKubernetesPerconaPostgresOperator(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesPerconaPostgresOperator Suite")
}

var _ = ginkgo.Describe("KubernetesPerconaPostgresOperator Validation Tests", func() {
	var input *KubernetesPerconaPostgresOperator

	ginkgo.BeforeEach(func() {
		input = &KubernetesPerconaPostgresOperator{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesPerconaPostgresOperator",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-percona-postgres-operator",
			},
			Spec: &KubernetesPerconaPostgresOperatorSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "percona-postgres-operator",
					},
				},
				Container: &KubernetesPerconaPostgresOperatorSpecContainer{
					Resources: &kubernetes.ContainerResources{
						Limits: &kubernetes.CpuMemory{
							Cpu:    "1000m",
							Memory: "1Gi",
						},
						Requests: &kubernetes.CpuMemory{
							Cpu:    "100m",
							Memory: "256Mi",
						},
					},
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("percona_postgres_operator", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When namespace is provided", func() {
		ginkgo.Context("with valid lowercase namespace", func() {
			ginkgo.It("should pass validation", func() {
				input.Spec.Namespace = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-namespace-123",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When container resources are required", func() {
		ginkgo.Context("with missing container spec", func() {
			ginkgo.It("should fail validation", func() {
				input.Spec.Container = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
