package kubernetespostgresv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestKubernetesPostgres(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesPostgres Suite")
}

var _ = ginkgo.Describe("KubernetesPostgres Custom Validation Tests", func() {
	var input *KubernetesPostgres

	ginkgo.BeforeEach(func() {
		input = &KubernetesPostgres{
			ApiVersion: "kubernetes.openmcf.org/v1",
			Kind:       "KubernetesPostgres",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-pg",
			},
			Spec: &KubernetesPostgresSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "test-namespace",
					},
				},
				Container: &KubernetesPostgresContainer{
					Replicas: 1,
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
					DiskSize: "2Gi", // valid disk size
				},
				Ingress: &KubernetesPostgresIngress{
					Enabled:  true,
					Hostname: "postgres.example.com",
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("postgres_kubernetes", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a backup config", func() {
			ginkgo.It("should accept a literal bucket with credentials", func() {
				input.Spec.BackupConfig = &KubernetesPostgresBackupConfig{
					Enabled: true,
					Bucket: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "pg-backups"},
					},
					ObjectPrefix: "prod",
					Schedule:     "0 2 * * *",
					RetainCount:  14,
					Credentials: &KubernetesPostgresR2Credentials{
						CloudflareAccountId: "acct-123",
						AccessKeyId:         "ak-123",
						SecretAccessKey:     "sk-123",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("with restore enabled but no source", func() {
			ginkgo.It("should require bucket, object_prefix, and credentials", func() {
				input.Spec.BackupConfig = &KubernetesPostgresBackupConfig{
					Restore: &KubernetesPostgresRestoreConfig{
						Enabled: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
