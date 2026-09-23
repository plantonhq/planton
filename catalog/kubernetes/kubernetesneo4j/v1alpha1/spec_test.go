package kubernetesneo4jv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesNeo4J(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesNeo4j Suite")
}

func boolPtr(b bool) *bool       { return &b }
func stringPtr(s string) *string { return &s }

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func valueFrom(kind cloudresourcekind.CloudResourceKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
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

var _ = ginkgo.Describe("KubernetesNeo4j Validation Tests", func() {
	var input *KubernetesNeo4J

	ginkgo.BeforeEach(func() {
		input = &KubernetesNeo4J{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesNeo4j",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-neo4j",
			},
			Spec: &KubernetesNeo4JSpec{
				Namespace: literal("graph"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error (every optional block omitted)", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(cloudresourcekind.CloudResourceKind_KubernetesNamespace, "graph", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("the community edition declared explicitly should be valid", func() {
			input.Spec.Edition = stringPtr("community")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("the enterprise edition with the license accepted should be valid", func() {
			input.Spec.Edition = stringPtr("enterprise")
			input.Spec.AcceptLicenseAgreement = true
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a cluster member (cluster_name on enterprise with license) should be valid", func() {
			input.Spec.Edition = stringPtr("enterprise")
			input.Spec.AcceptLicenseAgreement = true
			input.Spec.ClusterName = "graph-cluster"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("no auth declared should be valid (the module generates the admin password)", func() {
			input.Spec.Auth = nil
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a declared admin password should be valid", func() {
			input.Spec.Auth = &KubernetesNeo4JAuth{
				Source: &KubernetesNeo4JAuth_Password{Password: "super-secret"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an existing auth Secret should be valid", func() {
			input.Spec.Auth = &KubernetesNeo4JAuth{
				Source: &KubernetesNeo4JAuth_ExistingSecret{ExistingSecret: "neo4j-auth"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("every service type in the vocabulary should be valid", func() {
			for _, serviceType := range []string{"ClusterIP", "NodePort", "LoadBalancer"} {
				input.Spec.Service = &KubernetesNeo4JService{Type: stringPtr(serviceType)}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("full surface (auth, memory, config, service, ssl, scheduling, image) should be valid", func() {
			input.Spec.CreateNamespace = true
			input.Spec.ChartVersion = stringPtr("2026.6.0")
			input.Spec.Edition = stringPtr("enterprise")
			input.Spec.AcceptLicenseAgreement = true
			input.Spec.ClusterName = "graph-cluster"
			input.Spec.Auth = &KubernetesNeo4JAuth{
				Source: &KubernetesNeo4JAuth_Password{Password: "super-secret"},
			}
			input.Spec.Resources = &kubernetes.ContainerResources{
				Requests: &kubernetes.CpuMemory{Cpu: "500m", Memory: "2Gi"},
				Limits:   &kubernetes.CpuMemory{Cpu: "2", Memory: "4Gi"},
			}
			input.Spec.DataVolume = &KubernetesNeo4JDataVolume{
				Size:         stringPtr("50Gi"),
				StorageClass: literal("gp3"),
			}
			input.Spec.Memory = &KubernetesNeo4JMemory{
				HeapInitial: "1G",
				HeapMax:     "1G",
				PageCache:   "1G",
			}
			input.Spec.Config = map[string]string{"server.default_listen_address": "0.0.0.0"}
			input.Spec.ApocConfig = map[string]string{"apoc.export.file.enabled": "true"}
			input.Spec.AdditionalJvmArguments = []string{"-XX:+UseG1GC"}
			input.Spec.UseDefaultJvmArguments = boolPtr(true)
			input.Spec.Service = &KubernetesNeo4JService{
				Type:        stringPtr("LoadBalancer"),
				Annotations: map[string]string{"service.beta.kubernetes.io/aws-load-balancer-internal": "true"},
			}
			input.Spec.Ssl = &KubernetesNeo4JSsl{
				Bolt:  &KubernetesNeo4JSslScope{Secret: literal("neo4j-bolt-tls")},
				Https: &KubernetesNeo4JSslScope{Secret: literal("neo4j-https-tls")},
			}
			input.Spec.Scheduling = &KubernetesNeo4JScheduling{
				NodeSelector: map[string]string{"kubernetes.io/os": "linux"},
				Tolerations: []*kubernetes.WorkloadToleration{
					{Key: "dedicated", Operator: "Equal", Value: "graph", Effect: "NoSchedule"},
				},
				PodAntiAffinity:   boolPtr(true),
				PriorityClassName: "high-priority",
			}
			input.Spec.ServiceMonitorEnabled = true
			input.Spec.Image = &KubernetesNeo4JImage{
				Registry:   "mirror.example.com",
				Repository: "neo4j",
				Tag:        "2026.6.0-enterprise",
			}
			input.Spec.HelmValues = "podSpec:\n  labels:\n    team: graph\n"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("missing namespace should fail (required)", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown edition should fail (spec.edition_enum)", func() {
			input.Spec.Edition = stringPtr("professional")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("the enterprise edition without accepting the license should fail (spec.enterprise_requires_license_acceptance)", func() {
			input.Spec.Edition = stringPtr("enterprise")
			input.Spec.AcceptLicenseAgreement = false
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("cluster_name on the community edition should fail (spec.cluster_requires_enterprise)", func() {
			input.Spec.Edition = stringPtr("community")
			input.Spec.ClusterName = "graph-cluster"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("cluster_name with the edition left to its default should fail (spec.cluster_requires_enterprise)", func() {
			input.Spec.ClusterName = "graph-cluster"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown service type should fail (spec.service.type_enum)", func() {
			input.Spec.Service = &KubernetesNeo4JService{Type: stringPtr("ExternalName")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
