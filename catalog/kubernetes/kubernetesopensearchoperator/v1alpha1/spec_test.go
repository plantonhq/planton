package kubernetesopensearchoperatorv1alpha1

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

func TestKubernetesOpenSearchOperator(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesOpenSearchOperator Suite")
}

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

var _ = ginkgo.Describe("KubernetesOpenSearchOperator Validation Tests", func() {
	var input *KubernetesOpenSearchOperator

	ginkgo.BeforeEach(func() {
		input = &KubernetesOpenSearchOperator{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesOpenSearchOperator",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-opensearch-operator",
			},
			Spec: &KubernetesOpenSearchOperatorSpec{
				Namespace: literal("opensearch-operator"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error (every optional block omitted)", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("every CRD lifecycle position (bring-your-own CRDs, delete on uninstall) should be valid", func() {
			input.Spec.Crds = &KubernetesOpenSearchOperatorCrds{Install: boolPtr(false), KeepOnUninstall: boolPtr(false)}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "opensearch-operator", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("every log level in the operator's vocabulary should be valid", func() {
			for _, level := range []string{"debug", "info", "warn", "error"} {
				input.Spec.LogLevel = stringPtr(level)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("a watch_namespace fence alone should be valid", func() {
			input.Spec.WatchNamespace = "search-team"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("use_role_bindings paired with watch_namespace should be valid (the namespace-scoped RBAC shape)", func() {
			input.Spec.WatchNamespace = "search-team"
			input.Spec.UseRoleBindings = true
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full surface (chart pin, dns, gates, resources, scheduling, image, helm values) should be valid", func() {
			input.Spec.CreateNamespace = true
			input.Spec.ChartVersion = stringPtr("2.8.0")
			input.Spec.WatchNamespace = "search-team"
			input.Spec.UseRoleBindings = true
			input.Spec.LogLevel = stringPtr("debug")
			input.Spec.DnsBase = stringPtr("cluster.local")
			input.Spec.ParallelRecoveryEnabled = boolPtr(false)
			input.Spec.PprofEndpointsEnabled = true
			input.Spec.KubeRbacProxyEnabled = boolPtr(false)
			input.Spec.Resources = &kubernetes.ContainerResources{
				Requests: &kubernetes.CpuMemory{Cpu: "100m", Memory: "350Mi"},
				Limits:   &kubernetes.CpuMemory{Cpu: "200m", Memory: "500Mi"},
			}
			input.Spec.NodeSelector = map[string]string{"kubernetes.io/os": "linux"}
			input.Spec.Tolerations = []*kubernetes.WorkloadToleration{
				{Key: "dedicated", Operator: "Equal", Value: "search", Effect: "NoSchedule"},
			}
			input.Spec.ImagePullSecrets = []string{"mirror-pull"}
			input.Spec.Image = &KubernetesOpenSearchOperatorImage{
				Repository: "mirror.example.com/opensearchproject/opensearch-operator",
				Tag:        "2.8.0",
			}
			input.Spec.HelmValues = "manager:\n  podLabels:\n    team: search\n"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("missing namespace should fail (required)", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown log level should fail (spec.log_level_enum)", func() {
			input.Spec.LogLevel = stringPtr("trace")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("use_role_bindings without watch_namespace should fail (spec.role_bindings_require_watch_namespace)", func() {
			input.Spec.UseRoleBindings = true
			input.Spec.WatchNamespace = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
