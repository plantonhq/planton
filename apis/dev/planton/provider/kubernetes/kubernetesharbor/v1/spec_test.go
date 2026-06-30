package kubernetesharborv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestKubernetesHarborSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesHarborSpec Validation Suite")
}

var _ = ginkgo.Describe("KubernetesHarborSpec validations", func() {
	var spec *KubernetesHarborSpec

	ginkgo.BeforeEach(func() {
		spec = &KubernetesHarborSpec{
			TargetCluster: &kubernetes.KubernetesClusterSelector{
				ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
				ClusterName: "test-cluster",
			},
			Namespace: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "test-namespace",
				},
			},
			CoreContainer: &KubernetesHarborContainer{
				Replicas: 1,
				Resources: &kubernetes.ContainerResources{
					Limits: &kubernetes.CpuMemory{
						Cpu:    "1000m",
						Memory: "2Gi",
					},
					Requests: &kubernetes.CpuMemory{
						Cpu:    "200m",
						Memory: "512Mi",
					},
				},
			},
			PortalContainer: &KubernetesHarborContainer{
				Replicas: 1,
				Resources: &kubernetes.ContainerResources{
					Limits: &kubernetes.CpuMemory{
						Cpu:    "500m",
						Memory: "512Mi",
					},
					Requests: &kubernetes.CpuMemory{
						Cpu:    "100m",
						Memory: "256Mi",
					},
				},
			},
			RegistryContainer: &KubernetesHarborContainer{
				Replicas: 1,
				Resources: &kubernetes.ContainerResources{
					Limits: &kubernetes.CpuMemory{
						Cpu:    "1000m",
						Memory: "2Gi",
					},
					Requests: &kubernetes.CpuMemory{
						Cpu:    "200m",
						Memory: "512Mi",
					},
				},
			},
			JobserviceContainer: &KubernetesHarborContainer{
				Replicas: 1,
				Resources: &kubernetes.ContainerResources{
					Limits: &kubernetes.CpuMemory{
						Cpu:    "1000m",
						Memory: "1Gi",
					},
					Requests: &kubernetes.CpuMemory{
						Cpu:    "100m",
						Memory: "256Mi",
					},
				},
			},
			Database: &KubernetesHarborDatabaseConfig{
				IsExternal: false,
				ManagedDatabase: &KubernetesHarborManagedPostgresql{
					Container: &KubernetesHarborPostgresqlContainer{
						Replicas:           1,
						PersistenceEnabled: true,
						DiskSize:           "20Gi",
						Resources: &kubernetes.ContainerResources{
							Limits: &kubernetes.CpuMemory{
								Cpu:    "1000m",
								Memory: "2Gi",
							},
							Requests: &kubernetes.CpuMemory{
								Cpu:    "200m",
								Memory: "512Mi",
							},
						},
					},
				},
			},
			Cache: &KubernetesHarborCacheConfig{
				IsExternal: false,
				ManagedCache: &KubernetesHarborManagedRedis{
					Container: &KubernetesHarborRedisContainer{
						Replicas:           1,
						PersistenceEnabled: true,
						DiskSize:           "8Gi",
						Resources: &kubernetes.ContainerResources{
							Limits: &kubernetes.CpuMemory{
								Cpu:    "500m",
								Memory: "512Mi",
							},
							Requests: &kubernetes.CpuMemory{
								Cpu:    "100m",
								Memory: "256Mi",
							},
						},
					},
				},
			},
			Storage: &KubernetesHarborStorageConfig{
				Type: KubernetesHarborStorageType_s3,
				S3: &KubernetesHarborS3Storage{
					Bucket:    "test-bucket",
					Region:    "us-west-2",
					AccessKey: "test-key",
					SecretKey: "test-secret",
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("spec with managed database and cache", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with external database", func() {
			ginkgo.It("should not return a validation error when external_database is provided", func() {
				spec.Database = &KubernetesHarborDatabaseConfig{
					IsExternal: true,
					ExternalDatabase: &KubernetesHarborExternalPostgresql{
						Host:     "postgres.example.com",
						Username: "harbor",
						Password: "secret",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with external cache", func() {
			ginkgo.It("should not return a validation error when external_cache is provided", func() {
				spec.Cache = &KubernetesHarborCacheConfig{
					IsExternal: true,
					ExternalCache: &KubernetesHarborExternalRedis{
						Host:     "redis.example.com",
						Password: "secret",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with GCS storage", func() {
			ginkgo.It("should not return a validation error when GCS config is provided", func() {
				spec.Storage = &KubernetesHarborStorageConfig{
					Type: KubernetesHarborStorageType_gcs,
					Gcs: &KubernetesHarborGcsStorage{
						Bucket:  "test-bucket",
						KeyData: "base64-encoded-key",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with Azure storage", func() {
			ginkgo.It("should not return a validation error when Azure config is provided", func() {
				spec.Storage = &KubernetesHarborStorageConfig{
					Type: KubernetesHarborStorageType_azure,
					Azure: &KubernetesHarborAzureStorage{
						AccountName: "test-account",
						AccountKey:  "test-key",
						Container:   "test-container",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with OSS storage", func() {
			ginkgo.It("should not return a validation error when OSS config is provided", func() {
				spec.Storage = &KubernetesHarborStorageConfig{
					Type: KubernetesHarborStorageType_oss,
					Oss: &KubernetesHarborOssStorage{
						Bucket:          "test-bucket",
						Endpoint:        "oss-cn-hangzhou.aliyuncs.com",
						AccessKeyId:     "test-key",
						AccessKeySecret: "test-secret",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with filesystem storage", func() {
			ginkgo.It("should not return a validation error when filesystem config is provided", func() {
				spec.Storage = &KubernetesHarborStorageConfig{
					Type: KubernetesHarborStorageType_filesystem,
					Filesystem: &KubernetesHarborFilesystemStorage{
						DiskSize: "100Gi",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with ingress enabled", func() {
			ginkgo.It("should not return a validation error when hostname is provided", func() {
				spec.Ingress = &KubernetesHarborIngress{
					Core: &KubernetesHarborIngressEndpoint{
						Enabled:  true,
						Hostname: "harbor.example.com",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with Redis Sentinel", func() {
			ginkgo.It("should not return a validation error when sentinel_master_set is provided", func() {
				spec.Cache = &KubernetesHarborCacheConfig{
					IsExternal: true,
					ExternalCache: &KubernetesHarborExternalRedis{
						Host:              "redis.example.com",
						UseSentinel:       true,
						SentinelMasterSet: "mymaster",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with PostgreSQL without persistence", func() {
			ginkgo.It("should not return a validation error when disk_size is empty", func() {
				spec.Database.ManagedDatabase.Container.PersistenceEnabled = false
				spec.Database.ManagedDatabase.Container.DiskSize = ""
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with Redis without persistence", func() {
			ginkgo.It("should not return a validation error when disk_size is empty", func() {
				spec.Cache.ManagedCache.Container.PersistenceEnabled = false
				spec.Cache.ManagedCache.Container.DiskSize = ""
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("database configuration", func() {
			ginkgo.It("should return a validation error when is_external is true but external_database is missing", func() {
				spec.Database = &KubernetesHarborDatabaseConfig{
					IsExternal: true,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("cache configuration", func() {
			ginkgo.It("should return a validation error when is_external is true but external_cache is missing", func() {
				spec.Cache = &KubernetesHarborCacheConfig{
					IsExternal: true,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("storage configuration", func() {
			ginkgo.It("should return a validation error when type is s3 but s3 config is missing", func() {
				spec.Storage = &KubernetesHarborStorageConfig{
					Type: KubernetesHarborStorageType_s3,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when type is gcs but gcs config is missing", func() {
				spec.Storage = &KubernetesHarborStorageConfig{
					Type: KubernetesHarborStorageType_gcs,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when type is azure but azure config is missing", func() {
				spec.Storage = &KubernetesHarborStorageConfig{
					Type: KubernetesHarborStorageType_azure,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when type is oss but oss config is missing", func() {
				spec.Storage = &KubernetesHarborStorageConfig{
					Type: KubernetesHarborStorageType_oss,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when type is filesystem but filesystem config is missing", func() {
				spec.Storage = &KubernetesHarborStorageConfig{
					Type: KubernetesHarborStorageType_filesystem,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("PostgreSQL disk size validation", func() {
			ginkgo.It("should return a validation error when persistence is enabled but disk_size is empty", func() {
				spec.Database.ManagedDatabase.Container.PersistenceEnabled = true
				spec.Database.ManagedDatabase.Container.DiskSize = ""
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when disk_size format is invalid", func() {
				spec.Database.ManagedDatabase.Container.PersistenceEnabled = true
				spec.Database.ManagedDatabase.Container.DiskSize = "invalid-size"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("Redis disk size validation", func() {
			ginkgo.It("should return a validation error when persistence is enabled but disk_size is empty", func() {
				spec.Cache.ManagedCache.Container.PersistenceEnabled = true
				spec.Cache.ManagedCache.Container.DiskSize = ""
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when disk_size format is invalid", func() {
				spec.Cache.ManagedCache.Container.PersistenceEnabled = true
				spec.Cache.ManagedCache.Container.DiskSize = "not-a-size"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("filesystem storage disk size validation", func() {
			ginkgo.It("should return a validation error when disk_size is empty", func() {
				spec.Storage = &KubernetesHarborStorageConfig{
					Type: KubernetesHarborStorageType_filesystem,
					Filesystem: &KubernetesHarborFilesystemStorage{
						DiskSize: "",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when disk_size format is invalid", func() {
				spec.Storage = &KubernetesHarborStorageConfig{
					Type: KubernetesHarborStorageType_filesystem,
					Filesystem: &KubernetesHarborFilesystemStorage{
						DiskSize: "100gigabytes",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("ingress hostname validation", func() {
			ginkgo.It("should return a validation error when enabled is true but hostname is empty", func() {
				spec.Ingress = &KubernetesHarborIngress{
					Core: &KubernetesHarborIngressEndpoint{
						Enabled:  true,
						Hostname: "",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("Redis Sentinel validation", func() {
			ginkgo.It("should return a validation error when use_sentinel is true but sentinel_master_set is empty", func() {
				spec.Cache = &KubernetesHarborCacheConfig{
					IsExternal: true,
					ExternalCache: &KubernetesHarborExternalRedis{
						Host:              "redis.example.com",
						UseSentinel:       true,
						SentinelMasterSet: "",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("replica count validation", func() {
			ginkgo.It("should return a validation error when core container replicas is 0", func() {
				spec.CoreContainer.Replicas = 0
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when portal container replicas is 0", func() {
				spec.PortalContainer.Replicas = 0
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when registry container replicas is 0", func() {
				spec.RegistryContainer.Replicas = 0
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when jobservice container replicas is 0", func() {
				spec.JobserviceContainer.Replicas = 0
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when PostgreSQL replicas is 0", func() {
				spec.Database.ManagedDatabase.Container.Replicas = 0
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when Redis replicas is 0", func() {
				spec.Cache.ManagedCache.Container.Replicas = 0
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("port validation for external PostgreSQL", func() {
			ginkgo.It("should return a validation error when port is 0", func() {
				var invalidPort int32 = 0
				spec.Database = &KubernetesHarborDatabaseConfig{
					IsExternal: true,
					ExternalDatabase: &KubernetesHarborExternalPostgresql{
						Host:     "postgres.example.com",
						Port:     &invalidPort,
						Username: "harbor",
						Password: "secret",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when port is greater than 65535", func() {
				var invalidPort int32 = 70000
				spec.Database = &KubernetesHarborDatabaseConfig{
					IsExternal: true,
					ExternalDatabase: &KubernetesHarborExternalPostgresql{
						Host:     "postgres.example.com",
						Port:     &invalidPort,
						Username: "harbor",
						Password: "secret",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("port validation for external Redis", func() {
			ginkgo.It("should return a validation error when port is 0", func() {
				var invalidPort int32 = 0
				spec.Cache = &KubernetesHarborCacheConfig{
					IsExternal: true,
					ExternalCache: &KubernetesHarborExternalRedis{
						Host: "redis.example.com",
						Port: &invalidPort,
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when port is greater than 65535", func() {
				var invalidPort int32 = 80000
				spec.Cache = &KubernetesHarborCacheConfig{
					IsExternal: true,
					ExternalCache: &KubernetesHarborExternalRedis{
						Host: "redis.example.com",
						Port: &invalidPort,
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
