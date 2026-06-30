package kuberneteszalandopostgresoperatorv1

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

func TestKubernetesZalandoPostgresOperator(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesZalandoPostgresOperator Suite")
}

var _ = ginkgo.Describe("KubernetesZalandoPostgresOperator Custom Validation Tests", func() {
	var input *KubernetesZalandoPostgresOperator

	ginkgo.BeforeEach(func() {
		input = &KubernetesZalandoPostgresOperator{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesZalandoPostgresOperator",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-kubernetes-zalando-postgres-operator",
			},
			Spec: &KubernetesZalandoPostgresOperatorSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "test-namespace",
					},
				},
				Container: &KubernetesZalandoPostgresOperatorSpecContainer{},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("kubernetes_zalando_postgres_operator", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When backup_config is provided with a bucket and credentials", func() {
		ginkgo.Context("kubernetes_zalando_postgres_operator with backup", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.BackupConfig = &KubernetesZalandoPostgresOperatorBackupConfig{
					Bucket: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-bucket"},
					},
					ObjectPrefix: "prod",
					Schedule:     "0 2 * * *",
					Credentials: &KubernetesZalandoPostgresOperatorR2Credentials{
						CloudflareAccountId: "test-account-id",
						AccessKeyId:         "test-access-key",
						SecretAccessKey:     "test-secret-key",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When backup_config is missing its bucket", func() {
		ginkgo.Context("kubernetes_zalando_postgres_operator with incomplete backup", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.BackupConfig = &KubernetesZalandoPostgresOperatorBackupConfig{
					Schedule: "0 2 * * *",
					Credentials: &KubernetesZalandoPostgresOperatorR2Credentials{
						CloudflareAccountId: "test-account-id",
						AccessKeyId:         "test-access-key",
						SecretAccessKey:     "test-secret-key",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
