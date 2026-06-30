package kubernetestemporalv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

// Helper function to create a string password value
func stringPassword(value string) *kubernetes.KubernetesSensitiveValue {
	return &kubernetes.KubernetesSensitiveValue{
		SensitiveValue: &kubernetes.KubernetesSensitiveValue_Value{
			Value: value,
		},
	}
}

// Helper function to create a secret ref password value
func secretRefPassword(name, key string) *kubernetes.KubernetesSensitiveValue {
	return &kubernetes.KubernetesSensitiveValue{
		SensitiveValue: &kubernetes.KubernetesSensitiveValue_SecretRef{
			SecretRef: &kubernetes.KubernetesSecretKeyRef{
				Name: name,
				Key:  key,
			},
		},
	}
}

func TestKubernetesTemporal(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesTemporal Suite")
}

var _ = ginkgo.Describe("KubernetesTemporal Custom Validation Tests", func() {
	var input *KubernetesTemporal

	ginkgo.BeforeEach(func() {
		input = &KubernetesTemporal{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesTemporal",
			Metadata: &shared.CloudResourceMetadata{
				Name: "temporal-demo",
			},
			Spec: &KubernetesTemporalSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "test-namespace",
					},
				},
				DisableWebUi: false,
				Ingress: &KubernetesTemporalIngress{
					Frontend: &KubernetesTemporalFrontendIngressEndpoint{
						Enabled:      true,
						GrpcHostname: "temporal-frontend.example.com",
					},
					WebUi: &KubernetesTemporalWebUiIngressEndpoint{
						Enabled:  true,
						Hostname: "temporal-ui.example.com",
					},
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with cassandra external database and external elasticsearch", func() {
			ginkgo.BeforeEach(func() {
				input.Spec.Database = &KubernetesTemporalDatabaseConfig{
					Backend: KubernetesTemporalDatabaseBackend_cassandra,
					ExternalDatabase: &KubernetesTemporalExternalDatabase{
						Host:     "cassandra.example.com",
						Port:     9042,
						Username: "temporal_user",
						Password: stringPassword("secret"),
					},
					DatabaseName:           proto.String("temporal"),
					VisibilityName:         proto.String("temporal_visibility"),
					DisableAutoSchemaSetup: true,
				}
				input.Spec.ExternalElasticsearch = &KubernetesTemporalExternalElasticsearch{
					Host:     "es.example.com",
					Port:     9200,
					User:     "es_user",
					Password: stringPassword("es_password"),
				}
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with postgresql external database", func() {
			ginkgo.BeforeEach(func() {
				input.Spec.Database = &KubernetesTemporalDatabaseConfig{
					Backend: KubernetesTemporalDatabaseBackend_postgresql,
					ExternalDatabase: &KubernetesTemporalExternalDatabase{
						Host:     "postgres.example.com",
						Port:     5432,
						Username: "pg_user",
						Password: stringPassword("pg_password"),
					},
					DatabaseName:           proto.String("temporal_pg"),
					VisibilityName:         proto.String("temporal_visibility_pg"),
					DisableAutoSchemaSetup: false,
				}
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with mysql external database", func() {
			ginkgo.BeforeEach(func() {
				input.Spec.Database = &KubernetesTemporalDatabaseConfig{
					Backend: KubernetesTemporalDatabaseBackend_mysql,
					ExternalDatabase: &KubernetesTemporalExternalDatabase{
						Host:     "mysql.example.com",
						Port:     3306,
						Username: "mysql_user",
						Password: stringPassword("mysql_password"),
					},
					DatabaseName:   proto.String("temporal_mysql"),
					VisibilityName: proto.String("temporal_visibility_mysql"),
				}
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Search Attribute Validation Tests", func() {
		ginkgo.BeforeEach(func() {
			// Set up a valid base configuration with database
			input.Spec.Database = &KubernetesTemporalDatabaseConfig{
				Backend: KubernetesTemporalDatabaseBackend_postgresql,
				ExternalDatabase: &KubernetesTemporalExternalDatabase{
					Host:     "postgres.example.com",
					Port:     5432,
					Username: "pg_user",
					Password: stringPassword("pg_password"),
				},
				DatabaseName:   proto.String("temporal_pg"),
				VisibilityName: proto.String("temporal_visibility_pg"),
			}
		})
	})

	ginkgo.Describe("Password Secret Reference Tests", func() {
		ginkgo.BeforeEach(func() {
			// Set up a valid base configuration
			input.Spec.Database = &KubernetesTemporalDatabaseConfig{
				Backend:        KubernetesTemporalDatabaseBackend_postgresql,
				DatabaseName:   proto.String("temporal"),
				VisibilityName: proto.String("temporal_visibility"),
			}
		})

		ginkgo.Context("with database password using secret reference", func() {
			ginkgo.BeforeEach(func() {
				input.Spec.Database.ExternalDatabase = &KubernetesTemporalExternalDatabase{
					Host:     "postgres.example.com",
					Port:     5432,
					Username: "pg_user",
					Password: secretRefPassword("db-credentials", "password"),
				}
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with elasticsearch password using secret reference", func() {
			ginkgo.BeforeEach(func() {
				input.Spec.Database.ExternalDatabase = &KubernetesTemporalExternalDatabase{
					Host:     "postgres.example.com",
					Port:     5432,
					Username: "pg_user",
					Password: stringPassword("pg_password"),
				}
				input.Spec.ExternalElasticsearch = &KubernetesTemporalExternalElasticsearch{
					Host:     "es.example.com",
					Port:     9200,
					User:     "es_user",
					Password: secretRefPassword("es-credentials", "password"),
				}
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Dynamic Configuration Tests", func() {
		ginkgo.BeforeEach(func() {
			// Set up a valid base configuration with database
			input.Spec.Database = &KubernetesTemporalDatabaseConfig{
				Backend: KubernetesTemporalDatabaseBackend_postgresql,
				ExternalDatabase: &KubernetesTemporalExternalDatabase{
					Host:     "postgres.example.com",
					Port:     5432,
					Username: "pg_user",
					Password: stringPassword("pg_password"),
				},
				DatabaseName:   proto.String("temporal"),
				VisibilityName: proto.String("temporal_visibility"),
			}
		})

		ginkgo.Context("with valid dynamic config values", func() {
			ginkgo.BeforeEach(func() {
				input.Spec.DynamicConfig = &KubernetesTemporalDynamicConfig{
					HistorySizeLimitError:  proto.Int64(104857600), // 100 MB
					HistoryCountLimitError: proto.Int64(102400),    // 100K events
					HistorySizeLimitWarn:   proto.Int64(52428800),  // 50 MB
					HistoryCountLimitWarn:  proto.Int64(51200),     // 50K events
				}
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with history_size_limit_error below minimum", func() {
			ginkgo.BeforeEach(func() {
				input.Spec.DynamicConfig = &KubernetesTemporalDynamicConfig{
					HistorySizeLimitError: proto.Int64(100), // Below 1 MB minimum
				}
			})

			ginkgo.It("should return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with history_count_limit_error below minimum", func() {
			ginkgo.BeforeEach(func() {
				input.Spec.DynamicConfig = &KubernetesTemporalDynamicConfig{
					HistoryCountLimitError: proto.Int64(100), // Below 1000 minimum
				}
			})

			ginkgo.It("should return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("History Shards Configuration Tests", func() {
		ginkgo.BeforeEach(func() {
			// Set up a valid base configuration with database
			input.Spec.Database = &KubernetesTemporalDatabaseConfig{
				Backend: KubernetesTemporalDatabaseBackend_postgresql,
				ExternalDatabase: &KubernetesTemporalExternalDatabase{
					Host:     "postgres.example.com",
					Port:     5432,
					Username: "pg_user",
					Password: stringPassword("pg_password"),
				},
				DatabaseName:   proto.String("temporal"),
				VisibilityName: proto.String("temporal_visibility"),
			}
		})

		ginkgo.Context("with valid num_history_shards", func() {
			ginkgo.BeforeEach(func() {
				input.Spec.NumHistoryShards = proto.Int32(512)
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with num_history_shards at minimum (1)", func() {
			ginkgo.BeforeEach(func() {
				input.Spec.NumHistoryShards = proto.Int32(1)
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with num_history_shards at maximum (16384)", func() {
			ginkgo.BeforeEach(func() {
				input.Spec.NumHistoryShards = proto.Int32(16384)
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with num_history_shards below minimum", func() {
			ginkgo.BeforeEach(func() {
				input.Spec.NumHistoryShards = proto.Int32(0)
			})

			ginkgo.It("should return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with num_history_shards above maximum", func() {
			ginkgo.BeforeEach(func() {
				input.Spec.NumHistoryShards = proto.Int32(20000)
			})

			ginkgo.It("should return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Services Configuration Tests", func() {
		ginkgo.BeforeEach(func() {
			// Set up a valid base configuration with database
			input.Spec.Database = &KubernetesTemporalDatabaseConfig{
				Backend: KubernetesTemporalDatabaseBackend_postgresql,
				ExternalDatabase: &KubernetesTemporalExternalDatabase{
					Host:     "postgres.example.com",
					Port:     5432,
					Username: "pg_user",
					Password: stringPassword("pg_password"),
				},
				DatabaseName:   proto.String("temporal"),
				VisibilityName: proto.String("temporal_visibility"),
			}
		})

		ginkgo.Context("with valid services configuration", func() {
			ginkgo.BeforeEach(func() {
				input.Spec.Services = &KubernetesTemporalServices{
					Frontend: &KubernetesTemporalServiceConfig{
						Replicas: proto.Int32(2),
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
					History: &KubernetesTemporalServiceConfig{
						Replicas: proto.Int32(3),
						Resources: &kubernetes.ContainerResources{
							Limits: &kubernetes.CpuMemory{
								Cpu:    "2000m",
								Memory: "4Gi",
							},
							Requests: &kubernetes.CpuMemory{
								Cpu:    "500m",
								Memory: "1Gi",
							},
						},
					},
					Matching: &KubernetesTemporalServiceConfig{
						Replicas: proto.Int32(2),
					},
					Worker: &KubernetesTemporalServiceConfig{
						Replicas: proto.Int32(1),
					},
				}
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with service replicas below minimum", func() {
			ginkgo.BeforeEach(func() {
				input.Spec.Services = &KubernetesTemporalServices{
					Frontend: &KubernetesTemporalServiceConfig{
						Replicas: proto.Int32(0), // Below 1 minimum
					},
				}
			})

			ginkgo.It("should return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with service replicas above maximum", func() {
			ginkgo.BeforeEach(func() {
				input.Spec.Services = &KubernetesTemporalServices{
					History: &KubernetesTemporalServiceConfig{
						Replicas: proto.Int32(150), // Above 100 maximum
					},
				}
			})

			ginkgo.It("should return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
