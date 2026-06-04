package gcpcloudrunv1

import (
	"testing"

	"buf.build/go/protovalidate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestGcpCloudRunSpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GcpCloudRunSpec Validation Suite")
}

var _ = Describe("GcpCloudRunSpec validations", func() {

	strVal := func(v string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
		}
	}

	makeValidSpec := func() *GcpCloudRunSpec {
		return &GcpCloudRunSpec{
			ProjectId: strVal("my-gcp-project"),
			Region:    "us-central1",
			Container: &GcpCloudRunContainer{
				Image: &GcpCloudRunContainerImage{
					Repo: "us-docker.pkg.dev/my-project/repo/app",
					Tag:  "v1.0.0",
				},
				Port:   8080,
				Cpu:    1,
				Memory: 512,
				Replicas: &GcpCloudRunContainerReplicas{
					Min: 0,
					Max: 10,
				},
			},
		}
	}

	Context("Required fields", func() {
		It("accepts a minimal valid spec", func() {
			spec := makeValidSpec()
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects spec with missing project_id", func() {
			spec := makeValidSpec()
			spec.ProjectId = nil
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("rejects spec with missing region", func() {
			spec := makeValidSpec()
			spec.Region = ""
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("rejects spec with missing container", func() {
			spec := makeValidSpec()
			spec.Container = nil
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Region validation", func() {
		It("accepts valid region format", func() {
			spec := makeValidSpec()
			spec.Region = "us-west1"
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts multi-word region", func() {
			spec := makeValidSpec()
			spec.Region = "europe-west2"
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects region without trailing number", func() {
			spec := makeValidSpec()
			spec.Region = "us-central"
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("rejects region with uppercase letters", func() {
			spec := makeValidSpec()
			spec.Region = "US-CENTRAL1"
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Service name validation", func() {
		It("accepts valid service name", func() {
			spec := makeValidSpec()
			spec.ServiceName = "my-api-service"
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts empty service name (defaults to metadata.name)", func() {
			spec := makeValidSpec()
			spec.ServiceName = ""
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects service name with uppercase", func() {
			spec := makeValidSpec()
			spec.ServiceName = "MyService"
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("rejects service name starting with hyphen", func() {
			spec := makeValidSpec()
			spec.ServiceName = "-my-service"
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Container configuration", func() {
		It("accepts cpu = 1", func() {
			spec := makeValidSpec()
			spec.Container.Cpu = 1
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts cpu = 2", func() {
			spec := makeValidSpec()
			spec.Container.Cpu = 2
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts cpu = 4", func() {
			spec := makeValidSpec()
			spec.Container.Cpu = 4
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects cpu = 3 (not in allowed set)", func() {
			spec := makeValidSpec()
			spec.Container.Cpu = 3
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("accepts memory at minimum (128 MiB)", func() {
			spec := makeValidSpec()
			spec.Container.Memory = 128
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts memory at maximum (32768 MiB)", func() {
			spec := makeValidSpec()
			spec.Container.Memory = 32768
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects memory below minimum (127 MiB)", func() {
			spec := makeValidSpec()
			spec.Container.Memory = 127
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("rejects memory above maximum (32769 MiB)", func() {
			spec := makeValidSpec()
			spec.Container.Memory = 32769
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("accepts valid port", func() {
			spec := makeValidSpec()
			spec.Container.Port = 3000
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects port above 65535", func() {
			spec := makeValidSpec()
			spec.Container.Port = 65536
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Max concurrency validation", func() {
		It("accepts concurrency within range", func() {
			spec := makeValidSpec()
			spec.MaxConcurrency = 100
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts concurrency at maximum (1000)", func() {
			spec := makeValidSpec()
			spec.MaxConcurrency = 1000
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects concurrency above maximum (1001)", func() {
			spec := makeValidSpec()
			spec.MaxConcurrency = 1001
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("accepts zero concurrency (uses default)", func() {
			spec := makeValidSpec()
			spec.MaxConcurrency = 0
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})
	})

	Context("Timeout validation", func() {
		It("accepts timeout within range", func() {
			spec := makeValidSpec()
			spec.TimeoutSeconds = 60
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts timeout at maximum (3600)", func() {
			spec := makeValidSpec()
			spec.TimeoutSeconds = 3600
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects timeout above maximum (3601)", func() {
			spec := makeValidSpec()
			spec.TimeoutSeconds = 3601
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("accepts zero timeout (uses default)", func() {
			spec := makeValidSpec()
			spec.TimeoutSeconds = 0
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})
	})

	Context("DNS configuration - CEL validation", func() {
		It("accepts enabled DNS with hostnames and managed_zone", func() {
			spec := makeValidSpec()
			spec.Dns = &GcpCloudRunDns{
				Enabled:     true,
				Hostnames:   []string{"api.example.com"},
				ManagedZone: strVal("example-zone"),
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects enabled DNS without hostnames", func() {
			spec := makeValidSpec()
			spec.Dns = &GcpCloudRunDns{
				Enabled:     true,
				Hostnames:   []string{},
				ManagedZone: strVal("example-zone"),
			}
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("rejects enabled DNS without managed_zone", func() {
			spec := makeValidSpec()
			spec.Dns = &GcpCloudRunDns{
				Enabled:     true,
				Hostnames:   []string{"api.example.com"},
				ManagedZone: strVal(""),
			}
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("accepts disabled DNS without hostnames", func() {
			spec := makeValidSpec()
			spec.Dns = &GcpCloudRunDns{
				Enabled: false,
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects duplicate hostnames", func() {
			spec := makeValidSpec()
			spec.Dns = &GcpCloudRunDns{
				Enabled:     true,
				Hostnames:   []string{"api.example.com", "api.example.com"},
				ManagedZone: strVal("example-zone"),
			}
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("rejects hostname with uppercase", func() {
			spec := makeValidSpec()
			spec.Dns = &GcpCloudRunDns{
				Enabled:     true,
				Hostnames:   []string{"API.example.com"},
				ManagedZone: strVal("example-zone"),
			}
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Cloud SQL connectivity - CEL validation", func() {
		It("accepts native connection with instances", func() {
			spec := makeValidSpec()
			spec.CloudSql = &GcpCloudRunCloudSqlConnection{
				Connection: &GcpCloudRunCloudSqlDirectConnection{
					Instances: []*foreignkeyv1.StringValueOrRef{
						strVal("my-project:us-central1:my-db"),
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts auth proxy with instances", func() {
			spec := makeValidSpec()
			spec.CloudSql = &GcpCloudRunCloudSqlConnection{
				AuthProxy: &GcpCloudRunCloudSqlAuthProxy{
					Instances: []*foreignkeyv1.StringValueOrRef{
						strVal("my-project:us-central1:my-db"),
					},
					Port: 5432,
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects both connection and auth_proxy set (mutual exclusion)", func() {
			spec := makeValidSpec()
			spec.CloudSql = &GcpCloudRunCloudSqlConnection{
				Connection: &GcpCloudRunCloudSqlDirectConnection{
					Instances: []*foreignkeyv1.StringValueOrRef{
						strVal("my-project:us-central1:my-db"),
					},
				},
				AuthProxy: &GcpCloudRunCloudSqlAuthProxy{
					Instances: []*foreignkeyv1.StringValueOrRef{
						strVal("my-project:us-central1:my-db"),
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("rejects cloud_sql with neither connection nor auth_proxy", func() {
			spec := makeValidSpec()
			spec.CloudSql = &GcpCloudRunCloudSqlConnection{}
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("accepts spec without cloud_sql (optional)", func() {
			spec := makeValidSpec()
			spec.CloudSql = nil
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts native connection with multiple instances", func() {
			spec := makeValidSpec()
			spec.CloudSql = &GcpCloudRunCloudSqlConnection{
				Connection: &GcpCloudRunCloudSqlDirectConnection{
					Instances: []*foreignkeyv1.StringValueOrRef{
						strVal("my-project:us-central1:db-1"),
						strVal("my-project:us-central1:db-2"),
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects native connection with empty instances", func() {
			spec := makeValidSpec()
			spec.CloudSql = &GcpCloudRunCloudSqlConnection{
				Connection: &GcpCloudRunCloudSqlDirectConnection{
					Instances: []*foreignkeyv1.StringValueOrRef{},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("accepts auth proxy with use_private_ip", func() {
			spec := makeValidSpec()
			spec.CloudSql = &GcpCloudRunCloudSqlConnection{
				AuthProxy: &GcpCloudRunCloudSqlAuthProxy{
					Instances: []*foreignkeyv1.StringValueOrRef{
						strVal("my-project:us-central1:my-db"),
					},
					Port:         5432,
					UsePrivateIp: true,
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts auth proxy with custom port", func() {
			spec := makeValidSpec()
			spec.CloudSql = &GcpCloudRunCloudSqlConnection{
				AuthProxy: &GcpCloudRunCloudSqlAuthProxy{
					Instances: []*foreignkeyv1.StringValueOrRef{
						strVal("my-project:us-central1:my-db"),
					},
					Port: 3306,
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts native connection with value_from reference", func() {
			spec := makeValidSpec()
			spec.CloudSql = &GcpCloudRunCloudSqlConnection{
				Connection: &GcpCloudRunCloudSqlDirectConnection{
					Instances: []*foreignkeyv1.StringValueOrRef{
						{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
								ValueFrom: &foreignkeyv1.ValueFromRef{
									Name:      "my-cloudsql",
									FieldPath: "status.outputs.connection_name",
								},
							},
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})
	})

	Context("VPC access configuration", func() {
		It("accepts VPC access with network and subnet", func() {
			spec := makeValidSpec()
			spec.VpcAccess = &GcpCloudRunVpcAccess{
				Network: strVal("my-vpc"),
				Subnet:  strVal("my-subnet"),
				Egress:  "PRIVATE_RANGES_ONLY",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("accepts VPC access with ALL_TRAFFIC egress", func() {
			spec := makeValidSpec()
			spec.VpcAccess = &GcpCloudRunVpcAccess{
				Network: strVal("my-vpc"),
				Subnet:  strVal("my-subnet"),
				Egress:  "ALL_TRAFFIC",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("rejects VPC access with invalid egress", func() {
			spec := makeValidSpec()
			spec.VpcAccess = &GcpCloudRunVpcAccess{
				Network: strVal("my-vpc"),
				Subnet:  strVal("my-subnet"),
				Egress:  "INVALID_EGRESS",
			}
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
		})

		It("accepts spec without VPC access (optional)", func() {
			spec := makeValidSpec()
			spec.VpcAccess = nil
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})
	})

	Context("Complete production spec", func() {
		It("accepts a full-featured Cloud Run spec with Cloud SQL native connection", func() {
			spec := &GcpCloudRunSpec{
				ProjectId:   strVal("prod-project-123"),
				Region:      "us-central1",
				ServiceName: "production-api",
				Container: &GcpCloudRunContainer{
					Image: &GcpCloudRunContainerImage{
						Repo: "us-docker.pkg.dev/prod/containers/api",
						Tag:  "v3.0.0",
					},
					Cpu:    2,
					Memory: 2048,
					Port:   8080,
					Replicas: &GcpCloudRunContainerReplicas{
						Min: 2,
						Max: 50,
					},
					Env: &GcpCloudRunContainerEnv{
						Variables: map[string]string{
							"LOG_LEVEL": "info",
						},
						Secrets: map[string]string{
							"API_KEY": "projects/prod/secrets/api-key:latest",
						},
					},
				},
				MaxConcurrency:       100,
				TimeoutSeconds:       60,
				Ingress:              GcpCloudRunIngress_INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER,
				AllowUnauthenticated: false,
				ExecutionEnvironment: GcpCloudRunExecutionEnvironment_EXECUTION_ENVIRONMENT_GEN2,
				DeleteProtection:     true,
				CloudSql: &GcpCloudRunCloudSqlConnection{
					Connection: &GcpCloudRunCloudSqlDirectConnection{
						Instances: []*foreignkeyv1.StringValueOrRef{
							strVal("prod-project:us-central1:prod-db"),
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})
	})
})
