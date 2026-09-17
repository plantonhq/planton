package kubernetesvalkeyv1alpha1

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

func TestKubernetesValkey(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesValkey Suite")
}

func int32Ptr(i int32) *int32    { return &i }
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

// validReplication returns a minimal valid replication block (persistence
// declared inside, as replication mode requires) for tests that mutate one
// replication rule at a time.
func validReplication() *KubernetesValkeyReplication {
	return &KubernetesValkeyReplication{
		Persistence: &KubernetesValkeyPersistence{Size: "5Gi"},
	}
}

// validAuth returns a minimal valid auth block — the 'default' user must
// always be declared for auth to be meaningful.
func validAuth() *KubernetesValkeyAuth {
	return &KubernetesValkeyAuth{
		Users: []*KubernetesValkeyAclUser{
			{Name: "default", Password: "admin-pass"},
		},
	}
}

var _ = ginkgo.Describe("KubernetesValkey Validation Tests", func() {
	var input *KubernetesValkey

	ginkgo.BeforeEach(func() {
		input = &KubernetesValkey{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesValkey",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-valkey",
			},
			Spec: &KubernetesValkeySpec{
				Namespace: literal("caches"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error (every optional block omitted)", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(cloudresourcekind.CloudResourceKind_KubernetesNamespace, "caches", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("standalone with top-level persistence should be valid", func() {
			input.Spec.Persistence = &KubernetesValkeyPersistence{
				Size:            "5Gi",
				StorageClass:    literal("fast-ssd"),
				KeepOnUninstall: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("fractional persistence size '1.5Ti' should be valid (size_quantity)", func() {
			input.Spec.Persistence = &KubernetesValkeyPersistence{Size: "1.5Ti"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("replication with persistence inside the block should be valid", func() {
			input.Spec.Replication = validReplication()
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("one replica should be valid (gte 1 boundary)", func() {
			replication := validReplication()
			replication.Replicas = int32Ptr(1)
			input.Spec.Replication = replication
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("write-safety thresholds should be valid", func() {
			replication := validReplication()
			replication.MinReplicasToWrite = 1
			replication.MinReplicasMaxLag = int32Ptr(10)
			input.Spec.Replication = replication
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("read service of each allowed type should be valid (type_enum)", func() {
			for _, serviceType := range []string{"ClusterIP", "NodePort", "LoadBalancer"} {
				replication := validReplication()
				replication.ReadService = &KubernetesValkeyReadService{
					Enabled: boolPtr(true),
					Type:    stringPtr(serviceType),
				}
				input.Spec.Replication = replication
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("AOF with RDB save points should be valid", func() {
			input.Spec.Config = &KubernetesValkeyConfig{
				AppendOnly:    true,
				RdbSavePoints: []string{"900 1", "300 10"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("snapshots disabled without save points should be valid (save_points_xor_disabled)", func() {
			input.Spec.Config = &KubernetesValkeyConfig{
				SnapshotsDisabled: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("cache posture with max_memory and an eviction policy should be valid", func() {
			input.Spec.Config = &KubernetesValkeyConfig{
				MaxMemory:       "256mb",
				MaxMemoryPolicy: "allkeys-lru",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("max_memory '1gb' should be valid (max_memory_format)", func() {
			input.Spec.Config = &KubernetesValkeyConfig{MaxMemory: "1gb"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("auth with the default user should be valid (default_required)", func() {
			input.Spec.Auth = validAuth()
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an ACL user without a password should be valid (the module generates one)", func() {
			// A user needs only a name: the password is the module's to mint
			// into the <name>-auth Secret when the spec leaves it empty.
			input.Spec.Auth = &KubernetesValkeyAuth{
				Users: []*KubernetesValkeyAclUser{
					{Name: "default"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("declared and generated passwords may mix within one auth block", func() {
			input.Spec.Auth = &KubernetesValkeyAuth{
				Users: []*KubernetesValkeyAclUser{
					{Name: "default"},
					{Name: "reader", Password: "$secret/valkey-reader", Permissions: stringPtr("~* -@all +@read +ping +info")},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("auth with distinct users and custom permissions should be valid (unique_names)", func() {
			auth := validAuth()
			auth.Users = append(auth.Users, &KubernetesValkeyAclUser{
				Name:        "reader",
				Password:    "reader-pass",
				Permissions: stringPtr("~* -@all +@read +ping +info"),
			})
			input.Spec.Auth = auth
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("TLS enabled with a certificate secret should be valid (enabled_needs_secret)", func() {
			input.Spec.Tls = &KubernetesValkeyTls{
				Enabled:           true,
				CertificateSecret: valueFrom(cloudresourcekind.CloudResourceKind_KubernetesCertificate, "valkey-cert", "status.outputs.secret_name"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("mutual TLS with TLS enabled should be valid (mtls_needs_enabled)", func() {
			input.Spec.Tls = &KubernetesValkeyTls{
				Enabled:                  true,
				CertificateSecret:        literal("valkey-cert"),
				RequireClientCertificate: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("write service port boundaries should be valid (gte 1, lte 65535)", func() {
			for _, port := range []int32{1, 6379, 65535} {
				input.Spec.Service = &KubernetesValkeyService{
					Type: stringPtr("LoadBalancer"),
					Port: int32Ptr(port),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("metrics with a ServiceMonitor should be valid (monitor_needs_exporter)", func() {
			input.Spec.Metrics = &KubernetesValkeyMetrics{
				Enabled:               true,
				ServiceMonitorEnabled: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("PDB with only max_unavailable should be valid (single_bound)", func() {
			input.Spec.PodDisruptionBudget = &KubernetesValkeyPodDisruptionBudget{
				Enabled:        true,
				MaxUnavailable: 1,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("PDB with only min_available should be valid (single_bound)", func() {
			input.Spec.PodDisruptionBudget = &KubernetesValkeyPodDisruptionBudget{
				Enabled:      true,
				MinAvailable: 2,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("each allowed log level should be valid (log_level_enum)", func() {
			for _, level := range []string{"debug", "verbose", "notice", "warning"} {
				input.Spec.LogLevel = stringPtr(level)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("full-surface spec with every block populated should be valid", func() {
			input.Spec = &KubernetesValkeySpec{
				Namespace:       literal("caches"),
				CreateNamespace: true,
				ChartVersion:    stringPtr("0.11.0"),
				Image: &KubernetesValkeyImage{
					Registry:   "my-mirror.example.com",
					Repository: "valkey/valkey",
					Tag:        "9.1.1",
				},
				Replication: &KubernetesValkeyReplication{
					Replicas: int32Ptr(2),
					Persistence: &KubernetesValkeyPersistence{
						Size:         "10Gi",
						StorageClass: literal("fast-ssd"),
					},
					ReplicationUser:    stringPtr("replicator"),
					DisklessSync:       true,
					MinReplicasToWrite: 1,
					MinReplicasMaxLag:  int32Ptr(10),
					ReadService: &KubernetesValkeyReadService{
						Enabled:     boolPtr(true),
						Type:        stringPtr("ClusterIP"),
						Annotations: map[string]string{"prometheus.io/scrape": "true"},
					},
				},
				Config: &KubernetesValkeyConfig{
					AppendOnly:      true,
					RdbSavePoints:   []string{"900 1", "300 10"},
					MaxMemory:       "1gb",
					MaxMemoryPolicy: "noeviction",
					ExtraDirectives: "tcp-keepalive 300",
				},
				Auth: &KubernetesValkeyAuth{
					Users: []*KubernetesValkeyAclUser{
						{Name: "default", Password: "admin-pass"},
						{Name: "replicator", Password: "repl-pass", Permissions: stringPtr("+psync +replconf +ping")},
						{Name: "reader", Password: "reader-pass", Permissions: stringPtr("~* -@all +@read +ping +info")},
					},
				},
				Tls: &KubernetesValkeyTls{
					Enabled:                  true,
					CertificateSecret:        literal("valkey-cert"),
					RequireClientCertificate: true,
				},
				Service: &KubernetesValkeyService{
					Type:        stringPtr("LoadBalancer"),
					Port:        int32Ptr(6379),
					Annotations: map[string]string{"service.beta.kubernetes.io/aws-load-balancer-type": "nlb"},
				},
				Resources: &kubernetes.ContainerResources{
					Requests: &kubernetes.CpuMemory{Cpu: "250m", Memory: "1Gi"},
					Limits:   &kubernetes.CpuMemory{Cpu: "1", Memory: "2Gi"},
				},
				Metrics: &KubernetesValkeyMetrics{
					Enabled:               true,
					ServiceMonitorEnabled: true,
				},
				Scheduling: &KubernetesValkeyScheduling{
					NodeSelector: map[string]string{"workload": "caches"},
					Tolerations: []*kubernetes.WorkloadToleration{
						{Key: "dedicated", Operator: "Equal", Value: "caches", Effect: "NoSchedule"},
					},
					PriorityClassName: "cache-critical",
				},
				PodDisruptionBudget: &KubernetesValkeyPodDisruptionBudget{
					Enabled:        true,
					MaxUnavailable: 1,
				},
				LogLevel:         stringPtr("notice"),
				ImagePullSecrets: []string{"registry-pull"},
				HelmValues:       "networkPolicy:\n  enabled: true\n",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("missing namespace should fail (required)", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("replication combined with top-level persistence should fail (replication_persistence_placement)", func() {
			input.Spec.Replication = validReplication()
			input.Spec.Persistence = &KubernetesValkeyPersistence{Size: "5Gi"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("replication without persistence should fail (required)", func() {
			input.Spec.Replication = &KubernetesValkeyReplication{}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("zero replicas should fail (gte 1)", func() {
			replication := validReplication()
			replication.Replicas = int32Ptr(0)
			input.Spec.Replication = replication
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("negative min_replicas_to_write should fail (gte 0)", func() {
			replication := validReplication()
			replication.MinReplicasToWrite = -1
			input.Spec.Replication = replication
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("zero min_replicas_max_lag should fail (gte 1)", func() {
			replication := validReplication()
			replication.MinReplicasMaxLag = int32Ptr(0)
			input.Spec.Replication = replication
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown read service type should fail (read_service.type_enum)", func() {
			replication := validReplication()
			replication.ReadService = &KubernetesValkeyReadService{Type: stringPtr("ExternalName")}
			input.Spec.Replication = replication
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("empty persistence size should fail (required)", func() {
			input.Spec.Persistence = &KubernetesValkeyPersistence{}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("persistence size '5GB' should fail (size_quantity — GB is not a Kubernetes suffix)", func() {
			input.Spec.Persistence = &KubernetesValkeyPersistence{Size: "5GB"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("malformed replication persistence size should fail (size_quantity)", func() {
			input.Spec.Replication = &KubernetesValkeyReplication{
				Persistence: &KubernetesValkeyPersistence{Size: "abc"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("RDB save point without a change count should fail (rdb_save_point_format)", func() {
			input.Spec.Config = &KubernetesValkeyConfig{
				RdbSavePoints: []string{"900"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("save points combined with snapshots_disabled should fail (save_points_xor_disabled)", func() {
			input.Spec.Config = &KubernetesValkeyConfig{
				RdbSavePoints:     []string{"900 1"},
				SnapshotsDisabled: true,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("uppercase max_memory should fail (max_memory_format — Valkey sizes are lowercase)", func() {
			input.Spec.Config = &KubernetesValkeyConfig{MaxMemory: "256MB"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown max_memory_policy should fail (max_memory_policy_enum)", func() {
			input.Spec.Config = &KubernetesValkeyConfig{MaxMemoryPolicy: "lru"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("auth without users should fail (min_items 1)", func() {
			input.Spec.Auth = &KubernetesValkeyAuth{}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("duplicate ACL user names should fail (unique_names)", func() {
			input.Spec.Auth = &KubernetesValkeyAuth{
				Users: []*KubernetesValkeyAclUser{
					{Name: "default", Password: "pass-one"},
					{Name: "default", Password: "pass-two"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("auth without the default user should fail (default_required)", func() {
			input.Spec.Auth = &KubernetesValkeyAuth{
				Users: []*KubernetesValkeyAclUser{
					{Name: "app", Password: "app-pass"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("ACL user without a name should fail (required)", func() {
			input.Spec.Auth = &KubernetesValkeyAuth{
				Users: []*KubernetesValkeyAclUser{
					{Name: "default", Password: "admin-pass"},
					{Password: "orphan-pass"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})


		ginkgo.It("TLS enabled without a certificate secret should fail (enabled_needs_secret)", func() {
			input.Spec.Tls = &KubernetesValkeyTls{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("client certificates required without TLS enabled should fail (mtls_needs_enabled)", func() {
			input.Spec.Tls = &KubernetesValkeyTls{
				CertificateSecret:        literal("valkey-cert"),
				RequireClientCertificate: true,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown write service type should fail (service.type_enum)", func() {
			input.Spec.Service = &KubernetesValkeyService{Type: stringPtr("ExternalName")}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("zero service port should fail (gte 1)", func() {
			input.Spec.Service = &KubernetesValkeyService{Port: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("service port above 65535 should fail (lte 65535)", func() {
			input.Spec.Service = &KubernetesValkeyService{Port: int32Ptr(70000)}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("ServiceMonitor without the exporter should fail (monitor_needs_exporter)", func() {
			input.Spec.Metrics = &KubernetesValkeyMetrics{ServiceMonitorEnabled: true}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("PDB with both bounds should fail (single_bound)", func() {
			input.Spec.PodDisruptionBudget = &KubernetesValkeyPodDisruptionBudget{
				Enabled:        true,
				MaxUnavailable: 1,
				MinAvailable:   2,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("negative PDB max_unavailable should fail (gte 0)", func() {
			input.Spec.PodDisruptionBudget = &KubernetesValkeyPodDisruptionBudget{MaxUnavailable: -1}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown log_level should fail (log_level_enum)", func() {
			input.Spec.LogLevel = stringPtr("trace")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
