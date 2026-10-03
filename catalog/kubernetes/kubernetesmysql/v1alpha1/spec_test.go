package kubernetesmysqlv1alpha1

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

func TestKubernetesMysql(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesMysql Suite")
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

// s3Storage returns a minimal valid S3 backup storage — the base every
// backup test mutates from.
func s3Storage(name string) *KubernetesMysqlBackupStorage {
	return &KubernetesMysqlBackupStorage{
		Name: name,
		Backend: &KubernetesMysqlBackupStorage_S3{
			S3: &KubernetesMysqlS3Storage{
				Bucket: "mysql-backups",
				Region: "us-east-1",
				AccessKeys: &KubernetesMysqlS3AccessKeys{
					AccessKeyId:     "AKIAEXAMPLE",
					SecretAccessKey: "secret",
				},
			},
		},
	}
}

// validBackup returns a minimal valid backup block (one S3 storage, no
// schedules) for tests that mutate one backup rule at a time.
func validBackup() *KubernetesMysqlBackup {
	return &KubernetesMysqlBackup{
		Storages: []*KubernetesMysqlBackupStorage{s3Storage("primary")},
	}
}

var _ = ginkgo.Describe("KubernetesMysql Validation Tests", func() {
	var input *KubernetesMysql

	ginkgo.BeforeEach(func() {
		input = &KubernetesMysql{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesMysql",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-mysql",
			},
			Spec: &KubernetesMysqlSpec{
				Namespace: literal("databases"),
				Storage:   &KubernetesMysqlStorage{Size: "10Gi"},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error (every optional block omitted)", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "databases", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("three instances should be valid (Galera quorum shape)", func() {
			input.Spec.Instances = int32Ptr(3)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("five instances should be valid", func() {
			input.Spec.Instances = int32Ptr(5)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("one instance with unsafe.cluster_size should be valid (instances_quorum_or_unsafe)", func() {
			input.Spec.Instances = int32Ptr(1)
			input.Spec.Unsafe = &KubernetesMysqlUnsafe{ClusterSize: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("storage size '500Mi' should be valid (size_quantity)", func() {
			input.Spec.Storage.Size = "500Mi"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("fractional storage size '1.5Ti' should be valid (size_quantity)", func() {
			input.Spec.Storage.Size = "1.5Ti"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("storage class as a reference should be valid", func() {
			input.Spec.Storage.StorageClass = valueFrom(catalogkind.CatalogKind_KubernetesStorageClass, "fast-ssd", "status.outputs.storage_class_name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("HAProxy proxy with exposed services should be valid", func() {
			input.Spec.Proxy = &KubernetesMysqlProxy{
				Flavor: &KubernetesMysqlProxy_Haproxy{
					Haproxy: &KubernetesMysqlHaproxy{
						Replicas: int32Ptr(3),
						ExposePrimary: &KubernetesMysqlProxyService{
							Type: stringPtr("LoadBalancer"),
						},
						ExposeReplicas: &KubernetesMysqlHaproxyReplicasService{
							Enabled:     boolPtr(true),
							OnlyReaders: true,
							Type:        stringPtr("ClusterIP"),
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("ProxySQL proxy with its own storage should be valid", func() {
			input.Spec.Proxy = &KubernetesMysqlProxy{
				Flavor: &KubernetesMysqlProxy_Proxysql{
					Proxysql: &KubernetesMysqlProxysql{
						Replicas: int32Ptr(3),
						Storage:  &KubernetesMysqlStorage{Size: "2Gi"},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("one HAProxy replica with unsafe.proxy_size should be valid (proxy_replicas_or_unsafe)", func() {
			input.Spec.Proxy = &KubernetesMysqlProxy{
				Flavor: &KubernetesMysqlProxy_Haproxy{
					Haproxy: &KubernetesMysqlHaproxy{Replicas: int32Ptr(1)},
				},
			}
			input.Spec.Unsafe = &KubernetesMysqlUnsafe{ProxySize: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("one ProxySQL replica with unsafe.proxy_size should be valid (proxy_replicas_or_unsafe)", func() {
			input.Spec.Proxy = &KubernetesMysqlProxy{
				Flavor: &KubernetesMysqlProxy_Proxysql{
					Proxysql: &KubernetesMysqlProxysql{
						Replicas: int32Ptr(1),
						Storage:  &KubernetesMysqlStorage{Size: "2Gi"},
					},
				},
			}
			input.Spec.Unsafe = &KubernetesMysqlUnsafe{ProxySize: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("disabled TLS with unsafe.tls should be valid (tls_disabled_or_unsafe)", func() {
			input.Spec.Tls = &KubernetesMysqlTls{Enabled: boolPtr(false)}
			input.Spec.Unsafe = &KubernetesMysqlUnsafe{Tls: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("TLS with a cert-manager issuer should be valid", func() {
			input.Spec.Tls = &KubernetesMysqlTls{
				Enabled:    boolPtr(true),
				Issuer:     literal("org-ca-issuer"),
				IssuerKind: stringPtr("ClusterIssuer"),
				Sans:       []string{"mysql.example.com"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespaced Issuer kind should be valid (issuer_kind_enum)", func() {
			input.Spec.Tls = &KubernetesMysqlTls{
				Issuer:     literal("db-issuer"),
				IssuerKind: stringPtr("Issuer"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("users with distinct names should be valid (unique_names)", func() {
			input.Spec.Users = []*KubernetesMysqlUser{
				{Name: "app-writer", Dbs: []string{"orders"}, Grants: []string{"SELECT", "INSERT", "UPDATE", "DELETE"}},
				{Name: "app-reader", Dbs: []string{"orders"}, Hosts: []string{"%"}, Grants: []string{"SELECT"}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("backup to S3 with access keys should be valid", func() {
			input.Spec.Backup = validBackup()
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("backup to an S3-compatible endpoint with access keys should be valid (MinIO posture)", func() {
			backup := validBackup()
			s3 := backup.Storages[0].GetS3()
			s3.EndpointUrl = "http://minio.minio-system.svc:9000"
			s3.ForcePathStyle = true
			backup.Storages[0].VerifyTls = boolPtr(false)
			input.Spec.Backup = backup
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("backup to Azure Blob should be valid", func() {
			input.Spec.Backup = &KubernetesMysqlBackup{
				Storages: []*KubernetesMysqlBackupStorage{
					{
						Name: "azure-store",
						Backend: &KubernetesMysqlBackupStorage_Azure{
							Azure: &KubernetesMysqlAzureStorage{
								Container:      "mysql-backups",
								StorageAccount: "myaccount",
								AccessKey:      "base64key==",
							},
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("backup to a PVC should be valid", func() {
			input.Spec.Backup = &KubernetesMysqlBackup{
				Storages: []*KubernetesMysqlBackupStorage{
					{
						Name: "local-pvc",
						Backend: &KubernetesMysqlBackupStorage_Pvc{
							Pvc: &KubernetesMysqlPvcStorage{
								Volume: &KubernetesMysqlStorage{Size: "20Gi"},
							},
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("backup schedules with distinct names and five-field crons should be valid", func() {
			backup := validBackup()
			backup.Schedules = []*KubernetesMysqlBackupSchedule{
				{Name: "daily", Schedule: "0 2 * * *", StorageName: "primary", Keep: int32Ptr(7)},
				{Name: "weekly", Schedule: "30 4 * * 0", StorageName: "primary", DeleteFromStorage: boolPtr(false)},
			}
			input.Spec.Backup = backup
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("PITR with a dedicated storage should be valid (enabled_needs_storage)", func() {
			backup := validBackup()
			backup.Storages = append(backup.Storages, s3Storage("pitr-binlogs"))
			backup.Pitr = &KubernetesMysqlPitr{
				Enabled:            true,
				StorageName:        "pitr-binlogs",
				TimeBetweenUploads: int32Ptr(60),
			}
			input.Spec.Backup = backup
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("PDB with only max_unavailable should be valid (single_bound)", func() {
			input.Spec.PodDisruptionBudget = &KubernetesMysqlPodDisruptionBudget{MaxUnavailable: 1}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("PDB with only min_available should be valid (single_bound)", func() {
			input.Spec.PodDisruptionBudget = &KubernetesMysqlPodDisruptionBudget{MinAvailable: 2}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("each allowed update strategy should be valid (update_strategy_enum)", func() {
			for _, strategy := range []string{"SmartUpdate", "RollingUpdate", "OnDelete"} {
				input.Spec.UpdateStrategy = stringPtr(strategy)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("full-surface spec with every block populated should be valid", func() {
			input.Spec = &KubernetesMysqlSpec{
				Namespace:       literal("databases"),
				CreateNamespace: true,
				ImageName:       "percona/percona-xtradb-cluster:8.4.8-8.1",
				Instances:       int32Ptr(3),
				Storage: &KubernetesMysqlStorage{
					Size:         "100Gi",
					StorageClass: literal("fast-ssd"),
				},
				Resources: &kubernetes.ContainerResources{
					Requests: &kubernetes.CpuMemory{Cpu: "1", Memory: "4Gi"},
					Limits:   &kubernetes.CpuMemory{Cpu: "2", Memory: "8Gi"},
				},
				MysqlConfig:  "[mysqld]\nmax_connections=500\n",
				AutoRecovery: boolPtr(true),
				Proxy: &KubernetesMysqlProxy{
					Flavor: &KubernetesMysqlProxy_Haproxy{
						Haproxy: &KubernetesMysqlHaproxy{
							Replicas: int32Ptr(3),
							Resources: &kubernetes.ContainerResources{
								Requests: &kubernetes.CpuMemory{Cpu: "250m", Memory: "256Mi"},
								Limits:   &kubernetes.CpuMemory{Cpu: "500m", Memory: "512Mi"},
							},
							ExposePrimary: &KubernetesMysqlProxyService{
								Type:        stringPtr("LoadBalancer"),
								Annotations: map[string]string{"service.beta.kubernetes.io/aws-load-balancer-type": "nlb"},
							},
							ExposeReplicas: &KubernetesMysqlHaproxyReplicasService{
								Enabled:     boolPtr(true),
								OnlyReaders: true,
								Type:        stringPtr("ClusterIP"),
							},
						},
					},
				},
				Tls: &KubernetesMysqlTls{
					Enabled:    boolPtr(true),
					Issuer:     literal("org-ca-issuer"),
					IssuerKind: stringPtr("ClusterIssuer"),
					Sans:       []string{"mysql.example.com"},
				},
				Users: []*KubernetesMysqlUser{
					{
						Name:            "app-writer",
						Dbs:             []string{"orders"},
						Hosts:           []string{"%"},
						Grants:          []string{"SELECT", "INSERT", "UPDATE", "DELETE"},
						WithGrantOption: false,
						Password:        "writer-pass",
					},
					{Name: "app-reader", Dbs: []string{"orders"}, Grants: []string{"SELECT"}},
				},
				Backup: &KubernetesMysqlBackup{
					Storages: []*KubernetesMysqlBackupStorage{
						s3Storage("primary"),
						s3Storage("pitr-binlogs"),
					},
					Schedules: []*KubernetesMysqlBackupSchedule{
						{Name: "daily", Schedule: "0 2 * * *", StorageName: "primary", Keep: int32Ptr(7), DeleteFromStorage: boolPtr(true)},
					},
					Pitr: &KubernetesMysqlPitr{
						Enabled:            true,
						StorageName:        "pitr-binlogs",
						TimeBetweenUploads: int32Ptr(60),
					},
				},
				Scheduling: &KubernetesMysqlScheduling{
					AntiAffinityTopologyKey: "topology.kubernetes.io/zone",
					NodeSelector:            map[string]string{"workload": "databases"},
					Tolerations: []*kubernetes.WorkloadToleration{
						{Key: "dedicated", Operator: "Equal", Value: "databases", Effect: "NoSchedule"},
					},
					PriorityClassName: "database-critical",
				},
				PodDisruptionBudget: &KubernetesMysqlPodDisruptionBudget{MaxUnavailable: 1},
				LogCollector: &KubernetesMysqlLogCollector{
					Enabled: boolPtr(true),
					Resources: &kubernetes.ContainerResources{
						Requests: &kubernetes.CpuMemory{Cpu: "100m", Memory: "128Mi"},
						Limits:   &kubernetes.CpuMemory{Cpu: "200m", Memory: "256Mi"},
					},
				},
				UpdateStrategy: stringPtr("SmartUpdate"),
				Unsafe: &KubernetesMysqlUnsafe{
					BackupIfUnhealthy: true,
				},
				Pause:            false,
				ImagePullSecrets: []string{"registry-pull"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("missing namespace should fail (required)", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("zero instances should fail (gte 1)", func() {
			input.Spec.Instances = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("two instances without unsafe.cluster_size should fail (instances_quorum_or_unsafe)", func() {
			input.Spec.Instances = int32Ptr(2)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("one instance with unsafe declared but cluster_size false should fail (instances_quorum_or_unsafe)", func() {
			input.Spec.Instances = int32Ptr(1)
			input.Spec.Unsafe = &KubernetesMysqlUnsafe{Tls: true}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("missing storage should fail (required)", func() {
			input.Spec.Storage = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("empty storage size should fail (required)", func() {
			input.Spec.Storage.Size = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("storage size '10GB' should fail (size_quantity — GB is not a Kubernetes suffix)", func() {
			input.Spec.Storage.Size = "10GB"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("empty proxy block without a flavor should fail (flavor oneof required)", func() {
			input.Spec.Proxy = &KubernetesMysqlProxy{}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("one HAProxy replica without unsafe.proxy_size should fail (proxy_replicas_or_unsafe)", func() {
			input.Spec.Proxy = &KubernetesMysqlProxy{
				Flavor: &KubernetesMysqlProxy_Haproxy{
					Haproxy: &KubernetesMysqlHaproxy{Replicas: int32Ptr(1)},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("one ProxySQL replica without unsafe.proxy_size should fail (proxy_replicas_or_unsafe)", func() {
			input.Spec.Proxy = &KubernetesMysqlProxy{
				Flavor: &KubernetesMysqlProxy_Proxysql{
					Proxysql: &KubernetesMysqlProxysql{
						Replicas: int32Ptr(1),
						Storage:  &KubernetesMysqlStorage{Size: "2Gi"},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("one HAProxy replica with unsafe declared but proxy_size false should fail (proxy_replicas_or_unsafe)", func() {
			input.Spec.Proxy = &KubernetesMysqlProxy{
				Flavor: &KubernetesMysqlProxy_Haproxy{
					Haproxy: &KubernetesMysqlHaproxy{Replicas: int32Ptr(1)},
				},
			}
			input.Spec.Unsafe = &KubernetesMysqlUnsafe{ClusterSize: true}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("disabled TLS without unsafe.tls should fail (tls_disabled_or_unsafe)", func() {
			input.Spec.Tls = &KubernetesMysqlTls{Enabled: boolPtr(false)}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("zero HAProxy replicas should fail (gte 1)", func() {
			input.Spec.Proxy = &KubernetesMysqlProxy{
				Flavor: &KubernetesMysqlProxy_Haproxy{
					Haproxy: &KubernetesMysqlHaproxy{Replicas: int32Ptr(0)},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown HAProxy replicas service type should fail (expose_replicas.type_enum)", func() {
			input.Spec.Proxy = &KubernetesMysqlProxy{
				Flavor: &KubernetesMysqlProxy_Haproxy{
					Haproxy: &KubernetesMysqlHaproxy{
						ExposeReplicas: &KubernetesMysqlHaproxyReplicasService{Type: stringPtr("ExternalName")},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown proxy service type should fail (proxy.service.type_enum)", func() {
			input.Spec.Proxy = &KubernetesMysqlProxy{
				Flavor: &KubernetesMysqlProxy_Haproxy{
					Haproxy: &KubernetesMysqlHaproxy{
						ExposePrimary: &KubernetesMysqlProxyService{Type: stringPtr("ExternalName")},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("ProxySQL without storage should fail (required)", func() {
			input.Spec.Proxy = &KubernetesMysqlProxy{
				Flavor: &KubernetesMysqlProxy_Proxysql{
					Proxysql: &KubernetesMysqlProxysql{},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("zero ProxySQL replicas should fail (gte 1)", func() {
			input.Spec.Proxy = &KubernetesMysqlProxy{
				Flavor: &KubernetesMysqlProxy_Proxysql{
					Proxysql: &KubernetesMysqlProxysql{
						Replicas: int32Ptr(0),
						Storage:  &KubernetesMysqlStorage{Size: "2Gi"},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown issuer_kind should fail (issuer_kind_enum)", func() {
			input.Spec.Tls = &KubernetesMysqlTls{IssuerKind: stringPtr("Cluster")}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("duplicate user names should fail (unique_names)", func() {
			input.Spec.Users = []*KubernetesMysqlUser{
				{Name: "app"},
				{Name: "app"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("user without a name should fail (required)", func() {
			input.Spec.Users = []*KubernetesMysqlUser{
				{Dbs: []string{"orders"}},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown update_strategy should fail (update_strategy_enum)", func() {
			input.Spec.UpdateStrategy = stringPtr("Recreate")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("backup without storages should fail (min_items 1)", func() {
			input.Spec.Backup = &KubernetesMysqlBackup{}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("duplicate storage names should fail (storages.unique_names)", func() {
			input.Spec.Backup = &KubernetesMysqlBackup{
				Storages: []*KubernetesMysqlBackupStorage{
					s3Storage("primary"),
					s3Storage("primary"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("uppercase storage name should fail (storages.name_format — DNS label)", func() {
			backup := validBackup()
			backup.Storages[0].Name = "Primary"
			input.Spec.Backup = backup
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("storage without a backend arm should fail (oneof required)", func() {
			input.Spec.Backup = &KubernetesMysqlBackup{
				Storages: []*KubernetesMysqlBackupStorage{
					{Name: "primary"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("s3 without a bucket should fail (required)", func() {
			backup := validBackup()
			backup.Storages[0].GetS3().Bucket = ""
			input.Spec.Backup = backup
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("non-http endpoint_url should fail (endpoint_format)", func() {
			backup := validBackup()
			backup.Storages[0].GetS3().EndpointUrl = "minio.minio-system.svc:9000"
			input.Spec.Backup = backup
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("s3 without access keys should fail (required)", func() {
			backup := validBackup()
			backup.Storages[0].GetS3().AccessKeys = nil
			input.Spec.Backup = backup
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("access keys without access_key_id should fail (required)", func() {
			backup := validBackup()
			backup.Storages[0].GetS3().AccessKeys.AccessKeyId = ""
			input.Spec.Backup = backup
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("access keys without secret_access_key should fail (required)", func() {
			backup := validBackup()
			backup.Storages[0].GetS3().AccessKeys.SecretAccessKey = ""
			input.Spec.Backup = backup
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("azure without a container should fail (required)", func() {
			input.Spec.Backup = &KubernetesMysqlBackup{
				Storages: []*KubernetesMysqlBackupStorage{
					{
						Name: "azure-store",
						Backend: &KubernetesMysqlBackupStorage_Azure{
							Azure: &KubernetesMysqlAzureStorage{
								StorageAccount: "myaccount",
								AccessKey:      "base64key==",
							},
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("azure without a storage_account should fail (required)", func() {
			input.Spec.Backup = &KubernetesMysqlBackup{
				Storages: []*KubernetesMysqlBackupStorage{
					{
						Name: "azure-store",
						Backend: &KubernetesMysqlBackupStorage_Azure{
							Azure: &KubernetesMysqlAzureStorage{
								Container: "mysql-backups",
								AccessKey: "base64key==",
							},
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("azure without an access_key should fail (required)", func() {
			input.Spec.Backup = &KubernetesMysqlBackup{
				Storages: []*KubernetesMysqlBackupStorage{
					{
						Name: "azure-store",
						Backend: &KubernetesMysqlBackupStorage_Azure{
							Azure: &KubernetesMysqlAzureStorage{
								Container:      "mysql-backups",
								StorageAccount: "myaccount",
							},
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("pvc without a volume should fail (required)", func() {
			input.Spec.Backup = &KubernetesMysqlBackup{
				Storages: []*KubernetesMysqlBackupStorage{
					{
						Name:    "local-pvc",
						Backend: &KubernetesMysqlBackupStorage_Pvc{Pvc: &KubernetesMysqlPvcStorage{}},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("duplicate schedule names should fail (schedules.unique_names)", func() {
			backup := validBackup()
			backup.Schedules = []*KubernetesMysqlBackupSchedule{
				{Name: "daily", Schedule: "0 2 * * *", StorageName: "primary"},
				{Name: "daily", Schedule: "0 4 * * *", StorageName: "primary"},
			}
			input.Spec.Backup = backup
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("uppercase schedule name should fail (schedules.name_format — DNS label)", func() {
			backup := validBackup()
			backup.Schedules = []*KubernetesMysqlBackupSchedule{
				{Name: "Daily", Schedule: "0 2 * * *", StorageName: "primary"},
			}
			input.Spec.Backup = backup
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("four-field cron should fail (cron_five_fields)", func() {
			backup := validBackup()
			backup.Schedules = []*KubernetesMysqlBackupSchedule{
				{Name: "daily", Schedule: "0 2 * *", StorageName: "primary"},
			}
			input.Spec.Backup = backup
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("six-field cron should fail (cron_five_fields — seconds are not accepted)", func() {
			backup := validBackup()
			backup.Schedules = []*KubernetesMysqlBackupSchedule{
				{Name: "daily", Schedule: "0 0 2 * * *", StorageName: "primary"},
			}
			input.Spec.Backup = backup
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("schedule without storage_name should fail (required)", func() {
			backup := validBackup()
			backup.Schedules = []*KubernetesMysqlBackupSchedule{
				{Name: "daily", Schedule: "0 2 * * *"},
			}
			input.Spec.Backup = backup
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("zero schedule keep should fail (gte 1)", func() {
			backup := validBackup()
			backup.Schedules = []*KubernetesMysqlBackupSchedule{
				{Name: "daily", Schedule: "0 2 * * *", StorageName: "primary", Keep: int32Ptr(0)},
			}
			input.Spec.Backup = backup
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("PITR enabled without storage_name should fail (enabled_needs_storage)", func() {
			backup := validBackup()
			backup.Pitr = &KubernetesMysqlPitr{Enabled: true}
			input.Spec.Backup = backup
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("zero PITR time_between_uploads should fail (gte 1)", func() {
			backup := validBackup()
			backup.Pitr = &KubernetesMysqlPitr{
				Enabled:            true,
				StorageName:        "primary",
				TimeBetweenUploads: int32Ptr(0),
			}
			input.Spec.Backup = backup
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("PDB with both bounds should fail (single_bound)", func() {
			input.Spec.PodDisruptionBudget = &KubernetesMysqlPodDisruptionBudget{
				MaxUnavailable: 1,
				MinAvailable:   2,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("negative PDB max_unavailable should fail (gte 0)", func() {
			input.Spec.PodDisruptionBudget = &KubernetesMysqlPodDisruptionBudget{MaxUnavailable: -1}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
