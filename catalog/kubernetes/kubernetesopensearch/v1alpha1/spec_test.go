package kubernetesopensearchv1alpha1

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

func TestKubernetesOpenSearch(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesOpenSearch Suite")
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

// allRolesPool returns the minimal valid pool — TWO nodes carrying both
// cluster_manager and data roles (the manager floor: a single
// manager-eligible replica is rejected by the spec because it cannot
// survive the operator's bootstrap handoff; the dev shape every test
// mutates from).
func allRolesPool() *KubernetesOpenSearchNodePool {
	return &KubernetesOpenSearchNodePool{
		Name:     "masters",
		Replicas: 2,
		Roles:    []string{"cluster_manager", "data"},
	}
}

var _ = ginkgo.Describe("KubernetesOpenSearch Validation Tests", func() {
	var input *KubernetesOpenSearch

	ginkgo.BeforeEach(func() {
		input = &KubernetesOpenSearch{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesOpenSearch",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-opensearch",
			},
			Spec: &KubernetesOpenSearchSpec{
				Namespace: literal("search"),
				Version:   "2.19.4",
				NodePools: []*KubernetesOpenSearchNodePool{allRolesPool()},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "search", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("separate manager and data pools should be valid (production shape)", func() {
			input.Spec.NodePools = []*KubernetesOpenSearchNodePool{
				{Name: "managers", Replicas: 3, Roles: []string{"cluster_manager"}},
				{Name: "data", Replicas: 3, Roles: []string{"data", "ingest"}, DiskSize: "100Gi"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("PVC-backed persistence with a storage class should be valid", func() {
			input.Spec.NodePools[0].Persistence = &KubernetesOpenSearchPersistence{
				Source: &KubernetesOpenSearchPersistence_Pvc{
					Pvc: &KubernetesOpenSearchPvc{StorageClass: literal("gp3")},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("emptyDir persistence with a size limit should be valid (throwaway shape)", func() {
			input.Spec.NodePools[0].Persistence = &KubernetesOpenSearchPersistence{
				Source: &KubernetesOpenSearchPersistence_EmptyDir{
					EmptyDir: &KubernetesOpenSearchEmptyDir{SizeLimit: "10Gi"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a PDB bound by min_available alone should be valid", func() {
			input.Spec.NodePools[0].Pdb = &KubernetesOpenSearchPdb{Enable: true, MinAvailable: "2"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a PDB bound by max_unavailable alone should be valid", func() {
			input.Spec.NodePools[0].Pdb = &KubernetesOpenSearchPdb{Enable: true, MaxUnavailable: "25%"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("every Dashboards service type in the vocabulary should be valid", func() {
			for _, serviceType := range []string{"ClusterIP", "NodePort", "LoadBalancer"} {
				input.Spec.Dashboards = &KubernetesOpenSearchDashboards{
					Enabled: true,
					Service: &KubernetesOpenSearchDashboardsService{Type: stringPtr(serviceType)},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("an additional volume sourced from a Secret alone should be valid", func() {
			input.Spec.AdditionalVolumes = []*KubernetesOpenSearchAdditionalVolume{
				{Name: "certs", Path: "/mnt/certs", SecretName: "search-certs"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an additional volume sourced from a ConfigMap alone should be valid", func() {
			input.Spec.AdditionalVolumes = []*KubernetesOpenSearchAdditionalVolume{
				{Name: "extra-config", Path: "/mnt/config", ConfigMapName: "search-config", RestartPods: true},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full surface (security, dashboards, monitoring, keystore, snapshots, image) should be valid", func() {
			input.Spec.CreateNamespace = true
			input.Spec.HttpPort = int32Ptr(9201)
			input.Spec.NodePools[0].Resources = &kubernetes.ContainerResources{
				Requests: &kubernetes.CpuMemory{Cpu: "1", Memory: "2Gi"},
				Limits:   &kubernetes.CpuMemory{Cpu: "2", Memory: "4Gi"},
			}
			input.Spec.NodePools[0].Jvm = "-Xmx1G -Xms1G"
			input.Spec.NodePools[0].DiskSize = "30Gi"
			input.Spec.NodePools[0].AdditionalConfig = map[string]string{"indices.query.bool.max_clause_count": "2048"}
			input.Spec.NodePools[0].NodeSelector = map[string]string{"kubernetes.io/os": "linux"}
			input.Spec.NodePools[0].Tolerations = []*kubernetes.WorkloadToleration{
				{Key: "dedicated", Operator: "Equal", Value: "search", Effect: "NoSchedule"},
			}
			input.Spec.NodePools[0].Pdb = &KubernetesOpenSearchPdb{Enable: true, MinAvailable: "1"}
			input.Spec.AdditionalConfig = map[string]string{"action.auto_create_index": "false"}
			input.Spec.ServiceAnnotations = map[string]string{"service.beta.kubernetes.io/aws-load-balancer-internal": "true"}
			input.Spec.SetVmMaxMapCount = boolPtr(true)
			input.Spec.DrainDataNodes = true
			input.Spec.PluginsList = []string{"repository-s3"}
			input.Spec.Bootstrap = &KubernetesOpenSearchBootstrap{
				Jvm:              "-Xmx512M -Xms512M",
				AdditionalConfig: map[string]string{"cluster.routing.allocation.disk.threshold_enabled": "false"},
			}
			input.Spec.Security = &KubernetesOpenSearchSecurity{
				TransportTls: &KubernetesOpenSearchTlsTransport{
					Generate: boolPtr(false),
					Secret:   literal("transport-certs"),
					CaSecret: literal("transport-ca"),
					NodesDn:  []string{"CN=*.search.svc"},
					AdminDn:  []string{"CN=admin"},
				},
				HttpTls: &KubernetesOpenSearchTlsHttp{
					Generate: boolPtr(false),
					Secret:   literal("http-certs"),
				},
				Config: &KubernetesOpenSearchSecurityConfig{
					SecurityConfigSecret:   literal("security-config"),
					AdminSecret:            literal("admin-cert"),
					AdminCredentialsSecret: literal("admin-credentials"),
				},
			}
			input.Spec.Dashboards = &KubernetesOpenSearchDashboards{
				Enabled:  true,
				Replicas: int32Ptr(2),
				Version:  "2.19.4",
				Tls:      &KubernetesOpenSearchDashboardsTls{Enable: true, Generate: boolPtr(true)},
				BasePath: "/dashboards",
				AdditionalConfig: map[string]string{
					"opensearch_security.multitenancy.enabled": "true",
				},
				OpensearchCredentialsSecret: literal("dashboards-credentials"),
				Service: &KubernetesOpenSearchDashboardsService{
					Type:                     stringPtr("LoadBalancer"),
					LoadBalancerSourceRanges: []string{"10.0.0.0/8"},
				},
				PluginsList: []string{"some-dashboards-plugin"},
			}
			input.Spec.Monitoring = &KubernetesOpenSearchMonitoring{
				Enabled:              true,
				ScrapeInterval:       "30s",
				MonitoringUserSecret: literal("monitoring-user"),
				PluginUrl:            "https://mirror.example.com/prometheus-exporter-2.19.4.zip",
			}
			input.Spec.Keystore = []*KubernetesOpenSearchKeystoreValue{
				{
					Secret:      literal("s3-credentials"),
					KeyMappings: map[string]string{"accessKey": "s3.client.default.access_key"},
				},
			}
			input.Spec.SnapshotRepositories = []*KubernetesOpenSearchSnapshotRepo{
				{
					Name:     "nightly",
					Type:     "s3",
					Settings: map[string]string{"bucket": "backups", "region": "us-west-2"},
				},
			}
			input.Spec.AdditionalVolumes = []*KubernetesOpenSearchAdditionalVolume{
				{Name: "certs", Path: "/mnt/certs", SubPath: "tls", SecretName: "search-certs"},
			}
			input.Spec.Image = &kubernetes.ContainerImage{Repo: "mirror.example.com/opensearchproject/opensearch", Tag: "2.19.4"}
			input.Spec.ImagePullSecrets = []string{"mirror-pull"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When top-level fields are invalid", func() {
		ginkgo.It("missing namespace should fail (required)", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a single cluster_manager-eligible replica should fail (manager quorum floor)", func() {
			// One manager-eligible node cannot survive the operator's
			// bootstrap handoff (verified live) — the spec rejects the shape.
			input.Spec.NodePools = []*KubernetesOpenSearchNodePool{
				{Name: "all", Replicas: 1, Roles: []string{"cluster_manager", "data"}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("one manager-eligible replica per pool across TWO pools should be valid (floor is the total)", func() {
			input.Spec.NodePools = []*KubernetesOpenSearchNodePool{
				{Name: "m1", Replicas: 1, Roles: []string{"cluster_manager"}},
				{Name: "m2", Replicas: 1, Roles: []string{"cluster_manager", "data"}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("missing version should fail (required)", func() {
			input.Spec.Version = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an http port of zero should fail (gte 1)", func() {
			input.Spec.HttpPort = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an http port above 65535 should fail (lte 65535)", func() {
			input.Spec.HttpPort = int32Ptr(70000)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When node pools are invalid", func() {
		ginkgo.It("an empty node pool list should fail (min_items)", func() {
			input.Spec.NodePools = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a missing pool name should fail (required)", func() {
			input.Spec.NodePools[0].Name = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an uppercase pool name should fail (DNS-safe pattern)", func() {
			input.Spec.NodePools[0].Name = "Masters"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a pool name over 20 characters should fail (max_len)", func() {
			input.Spec.NodePools[0].Name = "a-very-long-pool-name-indeed"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero replicas should fail (required + gte 1)", func() {
			input.Spec.NodePools[0].Replicas = 0
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an empty roles list should fail (min_items)", func() {
			input.Spec.NodePools[0].Roles = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a PDB declaring both bounds should fail (spec.pdb.one_bound)", func() {
			input.Spec.NodePools[0].Pdb = &KubernetesOpenSearchPdb{
				Enable:         true,
				MinAvailable:   "2",
				MaxUnavailable: "1",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When optional blocks are invalid", func() {
		ginkgo.It("an unknown Dashboards service type should fail (spec.dashboards.service_type_enum)", func() {
			input.Spec.Dashboards = &KubernetesOpenSearchDashboards{
				Enabled: true,
				Service: &KubernetesOpenSearchDashboardsService{Type: stringPtr("ExternalName")},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero Dashboards replicas should fail (gte 1)", func() {
			input.Spec.Dashboards = &KubernetesOpenSearchDashboards{
				Enabled:  true,
				Replicas: int32Ptr(0),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a keystore entry without a secret should fail (required)", func() {
			input.Spec.Keystore = []*KubernetesOpenSearchKeystoreValue{{}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a snapshot repository without a name should fail (required)", func() {
			input.Spec.SnapshotRepositories = []*KubernetesOpenSearchSnapshotRepo{{Type: "s3"}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a snapshot repository without a type should fail (required)", func() {
			input.Spec.SnapshotRepositories = []*KubernetesOpenSearchSnapshotRepo{{Name: "nightly"}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an additional volume without a name should fail (required)", func() {
			input.Spec.AdditionalVolumes = []*KubernetesOpenSearchAdditionalVolume{
				{Path: "/mnt/certs", SecretName: "search-certs"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an additional volume without a path should fail (required)", func() {
			input.Spec.AdditionalVolumes = []*KubernetesOpenSearchAdditionalVolume{
				{Name: "certs", SecretName: "search-certs"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an additional volume with both sources should fail (spec.additional_volume.one_source)", func() {
			input.Spec.AdditionalVolumes = []*KubernetesOpenSearchAdditionalVolume{
				{Name: "certs", Path: "/mnt/certs", SecretName: "search-certs", ConfigMapName: "search-config"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an additional volume with neither source should fail (spec.additional_volume.one_source)", func() {
			input.Spec.AdditionalVolumes = []*KubernetesOpenSearchAdditionalVolume{
				{Name: "certs", Path: "/mnt/certs"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
