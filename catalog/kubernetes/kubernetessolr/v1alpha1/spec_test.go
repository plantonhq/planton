package kubernetessolrv1alpha1

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

func TestKubernetesSolr(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesSolr Suite")
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

// secretKey returns a minimal valid Secret+key reference.
func secretKey(name, key string) *KubernetesSolrSecretKeyRef {
	return &KubernetesSolrSecretKeyRef{Name: name, Key: key}
}

// minimalTls returns the smallest valid TLS block (both required
// keystore references present).
func minimalTls() *KubernetesSolrTls {
	return &KubernetesSolrTls{
		Pkcs12Secret:           secretKey("solr-keystore", "keystore.p12"),
		KeystorePasswordSecret: secretKey("solr-keystore-password", "password"),
	}
}

var _ = ginkgo.Describe("KubernetesSolr Validation Tests", func() {
	var input *KubernetesSolr

	ginkgo.BeforeEach(func() {
		input = &KubernetesSolr{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesSolr",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-solr",
			},
			Spec: &KubernetesSolrSpec{
				Namespace: literal("search"),
				Version:   "9.10.0",
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error (namespace + version only)", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "search", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("every log level in the vocabulary should be valid", func() {
			for _, level := range []string{"ERROR", "WARN", "INFO", "DEBUG", "TRACE"} {
				input.Spec.LogLevel = stringPtr(level)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("a provided ZooKeeper ensemble should be valid", func() {
			input.Spec.Zookeeper = &KubernetesSolrZookeeper{
				Source: &KubernetesSolrZookeeper_Provided{
					Provided: &KubernetesSolrProvidedZookeeper{
						Replicas: int32Ptr(3),
						Persistence: &KubernetesSolrProvidedZookeeperPersistence{
							Size:         "5Gi",
							StorageClass: literal("gp3"),
						},
						Chroot: stringPtr("/solr"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an external ZooKeeper ensemble should be valid", func() {
			input.Spec.Zookeeper = &KubernetesSolrZookeeper{
				Source: &KubernetesSolrZookeeper_External{
					External: &KubernetesSolrExternalZookeeper{
						ConnectionString: "zk-0.zk:2181,zk-1.zk:2181",
						Chroot:           stringPtr("/solr"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("persistent storage with each reclaim policy should be valid", func() {
			for _, policy := range []string{"Retain", "Delete"} {
				input.Spec.Storage = &KubernetesSolrStorage{
					Source: &KubernetesSolrStorage_Persistent{
						Persistent: &KubernetesSolrPersistentStorage{
							Size:          "20Gi",
							StorageClass:  literal("gp3"),
							ReclaimPolicy: stringPtr(policy),
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("ephemeral storage with a size limit should be valid (dev shape)", func() {
			input.Spec.Storage = &KubernetesSolrStorage{
				Source: &KubernetesSolrStorage_Ephemeral{
					Ephemeral: &KubernetesSolrEphemeralStorage{SizeLimit: "10Gi"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("basic authentication should be valid", func() {
			input.Spec.Security = &KubernetesSolrSecurity{
				AuthenticationType:    stringPtr("basic"),
				BasicAuthSecret:       literal("solr-basic-auth"),
				ProbesRequireAuth:     true,
				BootstrapSecurityJson: secretKey("solr-security-json", "security.json"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("every client_auth demand in the vocabulary should be valid on a TLS listener", func() {
			for _, clientAuth := range []string{"None", "Want", "Need"} {
				input.Spec.Tls = minimalTls()
				input.Spec.Tls.ClientAuth = stringPtr(clientAuth)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("an S3 backup repository should be valid", func() {
			input.Spec.BackupRepositories = []*KubernetesSolrBackupRepository{{
				Name: "nightly",
				Backend: &KubernetesSolrBackupRepository_S3{
					S3: &KubernetesSolrS3Repository{
						Region:       "us-west-2",
						Bucket:       "solr-backups",
						BaseLocation: "/prod",
						Credentials: &KubernetesSolrS3Credentials{
							AccessKeyIdSecret:     secretKey("s3-creds", "access-key-id"),
							SecretAccessKeySecret: secretKey("s3-creds", "secret-access-key"),
						},
					},
				},
			}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a GCS backup repository should be valid", func() {
			input.Spec.BackupRepositories = []*KubernetesSolrBackupRepository{{
				Name: "nightly",
				Backend: &KubernetesSolrBackupRepository_Gcs{
					Gcs: &KubernetesSolrGcsRepository{
						Bucket:              "solr-backups",
						GcsCredentialSecret: secretKey("gcs-creds", "key.json"),
					},
				},
			}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a volume backup repository should be valid", func() {
			input.Spec.BackupRepositories = []*KubernetesSolrBackupRepository{{
				Name: "shared-nfs",
				Backend: &KubernetesSolrBackupRepository_Volume{
					Volume: &KubernetesSolrVolumeRepository{
						PvcClaimName: "solr-backups-rwx",
						Directory:    "prod",
					},
				},
			}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("every update-strategy method in the vocabulary should be valid", func() {
			for _, method := range []string{"Managed", "StatefulSet", "Manual"} {
				input.Spec.UpdateStrategy = &KubernetesSolrUpdateStrategy{Method: stringPtr(method)}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("every external exposure method in the vocabulary should be valid", func() {
			for _, method := range []string{"Ingress", "ExternalDNS"} {
				input.Spec.External = &KubernetesSolrExternal{
					Method:     method,
					DomainName: "search.example.com",
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("full surface (jvm, tls, backups, strategy, scaling, external, scheduling) should be valid", func() {
			input.Spec.CreateNamespace = true
			input.Spec.Replicas = int32Ptr(3)
			input.Spec.ImageRepository = "mirror.example.com/solr"
			input.Spec.Zookeeper = &KubernetesSolrZookeeper{
				Source: &KubernetesSolrZookeeper_Provided{
					Provided: &KubernetesSolrProvidedZookeeper{Replicas: int32Ptr(3)},
				},
			}
			input.Spec.Storage = &KubernetesSolrStorage{
				Source: &KubernetesSolrStorage_Persistent{
					Persistent: &KubernetesSolrPersistentStorage{Size: "20Gi"},
				},
			}
			input.Spec.JavaMem = "-Xms1g -Xmx1g"
			input.Spec.SolrOpts = "-Dsolr.autoSoftCommit.maxTime=10000"
			input.Spec.LogLevel = stringPtr("INFO")
			input.Spec.GcTune = "-XX:+UseG1GC -XX:MaxGCPauseMillis=200"
			input.Spec.Resources = &kubernetes.ContainerResources{
				Requests: &kubernetes.CpuMemory{Cpu: "1", Memory: "2Gi"},
				Limits:   &kubernetes.CpuMemory{Cpu: "2", Memory: "2Gi"},
			}
			input.Spec.Security = &KubernetesSolrSecurity{AuthenticationType: stringPtr("basic")}
			input.Spec.Tls = minimalTls()
			input.Spec.Tls.TruststoreSecret = secretKey("solr-truststore", "truststore.p12")
			input.Spec.Tls.TruststorePasswordSecret = secretKey("solr-truststore-password", "password")
			input.Spec.Tls.ClientAuth = stringPtr("Want")
			input.Spec.Tls.VerifyClientHostname = true
			input.Spec.BackupRepositories = []*KubernetesSolrBackupRepository{{
				Name: "nightly",
				Backend: &KubernetesSolrBackupRepository_S3{
					S3: &KubernetesSolrS3Repository{Region: "us-west-2", Bucket: "solr-backups"},
				},
			}}
			input.Spec.SolrModules = []string{"analytics", "ltr"}
			input.Spec.AdditionalLibs = []string{"/opt/solr/custom-libs"}
			input.Spec.UpdateStrategy = &KubernetesSolrUpdateStrategy{
				Method:                      stringPtr("Managed"),
				MaxPodsUnavailable:          "25%",
				MaxShardReplicasUnavailable: "1",
				RestartSchedule:             "@every 168h",
			}
			input.Spec.Availability = &KubernetesSolrAvailability{PdbEnabled: boolPtr(true)}
			input.Spec.Scaling = &KubernetesSolrScaling{
				VacatePodsOnScaleDown: boolPtr(true),
				PopulatePodsOnScaleUp: boolPtr(true),
			}
			input.Spec.External = &KubernetesSolrExternal{
				Method:             "Ingress",
				DomainName:         "search.example.com",
				UseExternalAddress: true,
				HideNodes:          true,
			}
			input.Spec.PodPort = int32Ptr(8983)
			input.Spec.NodeSelector = map[string]string{"kubernetes.io/os": "linux"}
			input.Spec.Tolerations = []*kubernetes.WorkloadToleration{
				{Key: "dedicated", Operator: "Equal", Value: "search", Effect: "NoSchedule"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When top-level fields are invalid", func() {
		ginkgo.It("missing namespace should fail (required)", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("missing version should fail (required)", func() {
			input.Spec.Version = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero replicas should fail (gte 1)", func() {
			input.Spec.Replicas = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown log level should fail (spec.log_level_enum)", func() {
			input.Spec.LogLevel = stringPtr("VERBOSE")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a pod port of zero should fail (gte 1)", func() {
			input.Spec.PodPort = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a pod port above 65535 should fail (lte 65535)", func() {
			input.Spec.PodPort = int32Ptr(70000)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When ZooKeeper or storage blocks are invalid", func() {
		ginkgo.It("a zero-replica provided ensemble should fail (gte 1)", func() {
			input.Spec.Zookeeper = &KubernetesSolrZookeeper{
				Source: &KubernetesSolrZookeeper_Provided{
					Provided: &KubernetesSolrProvidedZookeeper{Replicas: int32Ptr(0)},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an external ensemble without a connection string should fail (required)", func() {
			input.Spec.Zookeeper = &KubernetesSolrZookeeper{
				Source: &KubernetesSolrZookeeper_External{
					External: &KubernetesSolrExternalZookeeper{},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("persistent storage without a size should fail (required)", func() {
			input.Spec.Storage = &KubernetesSolrStorage{
				Source: &KubernetesSolrStorage_Persistent{
					Persistent: &KubernetesSolrPersistentStorage{},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown reclaim policy should fail (spec.storage.reclaim_policy_enum)", func() {
			input.Spec.Storage = &KubernetesSolrStorage{
				Source: &KubernetesSolrStorage_Persistent{
					Persistent: &KubernetesSolrPersistentStorage{
						Size:          "20Gi",
						ReclaimPolicy: stringPtr("Recycle"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When security or TLS blocks are invalid", func() {
		ginkgo.It("an unknown authentication type should fail (spec.security.authentication_type_enum)", func() {
			input.Spec.Security = &KubernetesSolrSecurity{AuthenticationType: stringPtr("oauth")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a TLS block without the pkcs12 keystore should fail (required)", func() {
			input.Spec.Tls = minimalTls()
			input.Spec.Tls.Pkcs12Secret = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a TLS block without the keystore password should fail (required)", func() {
			input.Spec.Tls = minimalTls()
			input.Spec.Tls.KeystorePasswordSecret = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a secret reference missing its key should fail (required)", func() {
			input.Spec.Tls = minimalTls()
			input.Spec.Tls.Pkcs12Secret = &KubernetesSolrSecretKeyRef{Name: "solr-keystore"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown client_auth demand should fail (spec.tls.client_auth_enum)", func() {
			input.Spec.Tls = minimalTls()
			input.Spec.Tls.ClientAuth = stringPtr("Require")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When backup repositories are invalid", func() {
		ginkgo.It("a repository without any backend should fail (spec.backup_repository.backend_required)", func() {
			input.Spec.BackupRepositories = []*KubernetesSolrBackupRepository{{Name: "nightly"}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a repository name starting with a dash should fail (name pattern)", func() {
			input.Spec.BackupRepositories = []*KubernetesSolrBackupRepository{{
				Name: "-nightly",
				Backend: &KubernetesSolrBackupRepository_Volume{
					Volume: &KubernetesSolrVolumeRepository{PvcClaimName: "solr-backups-rwx"},
				},
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a repository name over 100 characters should fail (max_len)", func() {
			longName := make([]byte, 101)
			for i := range longName {
				longName[i] = 'a'
			}
			input.Spec.BackupRepositories = []*KubernetesSolrBackupRepository{{
				Name: string(longName),
				Backend: &KubernetesSolrBackupRepository_Volume{
					Volume: &KubernetesSolrVolumeRepository{PvcClaimName: "solr-backups-rwx"},
				},
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an S3 repository without a region should fail (required)", func() {
			input.Spec.BackupRepositories = []*KubernetesSolrBackupRepository{{
				Name: "nightly",
				Backend: &KubernetesSolrBackupRepository_S3{
					S3: &KubernetesSolrS3Repository{Bucket: "solr-backups"},
				},
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an S3 repository without a bucket should fail (required)", func() {
			input.Spec.BackupRepositories = []*KubernetesSolrBackupRepository{{
				Name: "nightly",
				Backend: &KubernetesSolrBackupRepository_S3{
					S3: &KubernetesSolrS3Repository{Region: "us-west-2"},
				},
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a GCS repository without a bucket should fail (required)", func() {
			input.Spec.BackupRepositories = []*KubernetesSolrBackupRepository{{
				Name: "nightly",
				Backend: &KubernetesSolrBackupRepository_Gcs{
					Gcs: &KubernetesSolrGcsRepository{},
				},
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a volume repository without a PVC claim name should fail (required)", func() {
			input.Spec.BackupRepositories = []*KubernetesSolrBackupRepository{{
				Name: "nightly",
				Backend: &KubernetesSolrBackupRepository_Volume{
					Volume: &KubernetesSolrVolumeRepository{},
				},
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When update strategy or external exposure is invalid", func() {
		ginkgo.It("an unknown update method should fail (spec.update_strategy.method_enum)", func() {
			input.Spec.UpdateStrategy = &KubernetesSolrUpdateStrategy{Method: stringPtr("Rolling")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an external block without a method should fail (required)", func() {
			input.Spec.External = &KubernetesSolrExternal{DomainName: "search.example.com"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown external method should fail (spec.external.method_enum)", func() {
			input.Spec.External = &KubernetesSolrExternal{
				Method:     "NodePort",
				DomainName: "search.example.com",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an external block without a domain name should fail (required)", func() {
			input.Spec.External = &KubernetesSolrExternal{Method: "Ingress"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
