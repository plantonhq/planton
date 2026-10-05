package kubernetesperconamysqloperatorv1alpha1

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

func TestKubernetesPerconaMysqlOperator(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesPerconaMysqlOperator Suite")
}

func int32Ptr(i int32) *int32    { return &i }
func boolPtr(b bool) *bool       { return &b }
func stringPtr(s string) *string { return &s }

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

var _ = ginkgo.Describe("KubernetesPerconaMysqlOperator Validation Tests", func() {
	var input *KubernetesPerconaMysqlOperator

	ginkgo.BeforeEach(func() {
		input = &KubernetesPerconaMysqlOperator{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesPerconaMysqlOperator",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-pxc-operator",
			},
			Spec: &KubernetesPerconaMysqlOperatorSpec{
				Namespace: literal("mysql-operator"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error (every optional block omitted)", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "mysql-operator", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("one replica should be valid (gte 1 boundary)", func() {
			input.Spec.Replicas = int32Ptr(1)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("cluster-wide watch without namespaces should be valid (cluster_wide_xor_namespaces)", func() {
			input.Spec.Watch = &KubernetesPerconaMysqlOperatorWatch{ClusterWide: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a fenced namespace list without cluster_wide should be valid (cluster_wide_xor_namespaces)", func() {
			input.Spec.Watch = &KubernetesPerconaMysqlOperatorWatch{
				Namespaces: []string{"databases", "staging"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an empty watch block should be valid (own-namespace posture)", func() {
			input.Spec.Watch = &KubernetesPerconaMysqlOperatorWatch{}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("each allowed log level should be valid (level_enum)", func() {
			for _, level := range []string{"DEBUG", "INFO", "ERROR"} {
				input.Spec.Log = &KubernetesPerconaMysqlOperatorLog{
					Structured: true,
					Level:      stringPtr(level),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("leader election with tuned lease timings should be valid", func() {
			input.Spec.Replicas = int32Ptr(2)
			input.Spec.LeaderElection = &KubernetesPerconaMysqlOperatorLeaderElection{
				Enabled:       boolPtr(true),
				LeaseDuration: stringPtr("60s"),
				RenewDeadline: stringPtr("40s"),
				RetryPeriod:   stringPtr("10s"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full-surface spec with every block populated should be valid", func() {
			input.Spec = &KubernetesPerconaMysqlOperatorSpec{
				Namespace:               literal("mysql-operator"),
				CreateNamespace:         true,
				ChartVersion:            stringPtr("1.20.0"),
				Replicas:                int32Ptr(2),
				Watch:                   &KubernetesPerconaMysqlOperatorWatch{ClusterWide: true},
				MaxConcurrentReconciles: int32Ptr(4),
				S3WorkersLimit:          int32Ptr(10),
				Log: &KubernetesPerconaMysqlOperatorLog{
					Structured: true,
					Level:      stringPtr("INFO"),
				},
				DisableTelemetry: true,
				LeaderElection: &KubernetesPerconaMysqlOperatorLeaderElection{
					Enabled:       boolPtr(true),
					LeaseDuration: stringPtr("60s"),
					RenewDeadline: stringPtr("40s"),
					RetryPeriod:   stringPtr("10s"),
				},
				XtrabackupSidecar: true,
				Resources: &kubernetes.ContainerResources{
					Requests: &kubernetes.CpuMemory{Cpu: "100m", Memory: "20Mi"},
					Limits:   &kubernetes.CpuMemory{Cpu: "200m", Memory: "500Mi"},
				},
				NodeSelector: map[string]string{"workload": "operators"},
				Tolerations: []*kubernetes.WorkloadToleration{
					{Key: "dedicated", Operator: "Equal", Value: "operators", Effect: "NoSchedule"},
				},
				ImagePullSecrets: []string{"registry-pull"},
				Image: &KubernetesPerconaMysqlOperatorImage{
					Repository: "my-mirror.example.com/percona-xtradb-cluster-operator",
					Tag:        "1.20.0",
				},
				HelmValues: "podAnnotations:\n  team: data\n",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("missing namespace should fail (required)", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("zero replicas should fail (gte 1)", func() {
			input.Spec.Replicas = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("zero max_concurrent_reconciles should fail (gte 1)", func() {
			input.Spec.MaxConcurrentReconciles = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("zero s3_workers_limit should fail (gte 1)", func() {
			input.Spec.S3WorkersLimit = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("cluster_wide combined with namespaces should fail (cluster_wide_xor_namespaces)", func() {
			input.Spec.Watch = &KubernetesPerconaMysqlOperatorWatch{
				ClusterWide: true,
				Namespaces:  []string{"databases"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown log level should fail (level_enum)", func() {
			input.Spec.Log = &KubernetesPerconaMysqlOperatorLog{Level: stringPtr("WARN")}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
