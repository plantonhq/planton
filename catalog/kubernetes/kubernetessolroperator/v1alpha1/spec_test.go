package kubernetessolroperatorv1alpha1

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

func TestKubernetesSolrOperator(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesSolrOperator Suite")
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

var _ = ginkgo.Describe("KubernetesSolrOperator Validation Tests", func() {
	var input *KubernetesSolrOperator

	ginkgo.BeforeEach(func() {
		input = &KubernetesSolrOperator{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesSolrOperator",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-solr-operator",
			},
			Spec: &KubernetesSolrOperatorSpec{
				Namespace: literal("solr-operator"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error (every optional block omitted)", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("every CRD lifecycle position (bring-your-own CRDs, delete on uninstall) should be valid", func() {
			input.Spec.Crds = &KubernetesSolrOperatorCrds{Install: boolPtr(false), KeepOnUninstall: boolPtr(false)}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "solr-operator", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a namespace fence via watch_namespaces should be valid", func() {
			input.Spec.WatchNamespaces = []string{"team-a", "team-b"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("skipping the bundled zookeeper-operator while using an existing one should be valid", func() {
			input.Spec.ZookeeperOperator = &KubernetesSolrOperatorZookeeperOperator{
				Install:     boolPtr(false),
				UseExisting: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full surface (replicas, mtls, resources, scheduling, image, helm values) should be valid", func() {
			input.Spec.CreateNamespace = true
			input.Spec.ChartVersion = stringPtr("0.9.1")
			input.Spec.Replicas = int32Ptr(2)
			input.Spec.WatchNamespaces = []string{"search"}
			input.Spec.ZookeeperOperator = &KubernetesSolrOperatorZookeeperOperator{Install: boolPtr(true)}
			input.Spec.LeaderElectionEnabled = boolPtr(true)
			input.Spec.MetricsEnabled = boolPtr(false)
			input.Spec.Mtls = &KubernetesSolrOperatorMtls{
				ClientCertSecret:   literal("solr-operator-client-cert"),
				CaCertSecret:       literal("solr-ca"),
				CaCertSecretKey:    stringPtr("ca-cert.pem"),
				InsecureSkipVerify: boolPtr(true),
				WatchForUpdates:    boolPtr(true),
			}
			input.Spec.Resources = &kubernetes.ContainerResources{
				Requests: &kubernetes.CpuMemory{Cpu: "100m", Memory: "128Mi"},
				Limits:   &kubernetes.CpuMemory{Cpu: "500m", Memory: "256Mi"},
			}
			input.Spec.NodeSelector = map[string]string{"kubernetes.io/os": "linux"}
			input.Spec.Tolerations = []*kubernetes.WorkloadToleration{
				{Key: "dedicated", Operator: "Equal", Value: "search", Effect: "NoSchedule"},
			}
			input.Spec.ImagePullSecret = "mirror-pull"
			input.Spec.Image = &KubernetesSolrOperatorImage{
				Repository: "mirror.example.com/apache/solr-operator",
				Tag:        "v0.9.1",
			}
			input.Spec.HelmValues = "zookeeper-operator:\n  watchNamespace: search\n"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("missing namespace should fail (required)", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero replicas should fail (replicas gte 1)", func() {
			input.Spec.Replicas = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("mtls block without client_cert_secret should fail (required)", func() {
			// An mtls block without a client certificate would render
			// nothing and silently leave the operator without an identity.
			input.Spec.Mtls = &KubernetesSolrOperatorMtls{
				InsecureSkipVerify: boolPtr(true),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
