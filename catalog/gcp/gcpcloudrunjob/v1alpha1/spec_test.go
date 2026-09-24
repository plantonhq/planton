package gcpcloudrunjobv1alpha1

import (
	"errors"
	"testing"

	"buf.build/go/protovalidate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestGcpCloudRunJobSpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GcpCloudRunJobSpec Validation Suite")
}

// violatedRules lists the rule ids a validation error names, so a case pins
// the rule it exists for rather than any failure at all.
func violatedRules(err error) []string {
	var validationErr *protovalidate.ValidationError
	if !errors.As(err, &validationErr) {
		return nil
	}
	ids := make([]string, 0, len(validationErr.Violations))
	for _, violation := range validationErr.Violations {
		ids = append(ids, violation.Proto.GetRuleId())
	}
	return ids
}

var _ = Describe("GcpCloudRunJobSpec validations", func() {

	strVal := func(v string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
		}
	}

	makeValidSpec := func() *GcpCloudRunJobSpec {
		return &GcpCloudRunJobSpec{
			ProjectId: strVal("my-gcp-project"),
			Region:    "us-central1",
			Template: &GcpCloudRunJobTemplate{
				Containers: []*GcpCloudRunJobContainer{
					{Image: "us-docker.pkg.dev/my-project/repo/worker:v1.0.0"},
				},
			},
		}
	}

	Context("Required fields", func() {
		It("accepts a minimal valid spec", func() {
			Expect(protovalidate.Validate(makeValidSpec())).To(BeNil())
		})

		It("accepts a spec without project_id (ambient project)", func() {
			spec := makeValidSpec()
			spec.ProjectId = nil
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects spec with missing region", func() {
			spec := makeValidSpec()
			spec.Region = ""
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects spec with missing template", func() {
			spec := makeValidSpec()
			spec.Template = nil
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects template with no containers", func() {
			spec := makeValidSpec()
			spec.Template.Containers = nil
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects a container with no image", func() {
			spec := makeValidSpec()
			spec.Template.Containers[0].Image = ""
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("Region validation", func() {
		It("accepts a standard region", func() {
			spec := makeValidSpec()
			spec.Region = "europe-west1"
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects a zone as region", func() {
			spec := makeValidSpec()
			spec.Region = "us-central1-a"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("Job name validation", func() {
		It("accepts a valid job name", func() {
			spec := makeValidSpec()
			spec.JobName = "my-batch-job"
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects an invalid job name", func() {
			spec := makeValidSpec()
			spec.JobName = "My_Job"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("Task count and parallelism", func() {
		It("accepts valid task_count and parallelism", func() {
			spec := makeValidSpec()
			spec.TaskCount = proto.Int32(10)
			spec.Parallelism = proto.Int32(5)
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects parallelism greater than task_count", func() {
			spec := makeValidSpec()
			spec.TaskCount = proto.Int32(3)
			spec.Parallelism = proto.Int32(5)
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects task_count below 1", func() {
			spec := makeValidSpec()
			spec.TaskCount = proto.Int32(0)
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("GPU zonal redundancy", func() {
		It("rejects gpu_zonal_redundancy_disabled without node_selector", func() {
			spec := makeValidSpec()
			spec.GpuZonalRedundancyDisabled = true
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts gpu_zonal_redundancy_disabled with node_selector", func() {
			spec := makeValidSpec()
			spec.GpuZonalRedundancyDisabled = true
			spec.Template.NodeSelector = &GcpCloudRunJobNodeSelector{Accelerator: "nvidia-l4"}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})
	})

	Context("Environment variables", func() {
		It("accepts a literal env var", func() {
			spec := makeValidSpec()
			spec.Template.Containers[0].Env = []*GcpCloudRunJobEnvVar{{Name: "BATCH_SIZE", Value: "100"}}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts a secret env var", func() {
			spec := makeValidSpec()
			spec.Template.Containers[0].Env = []*GcpCloudRunJobEnvVar{{
				Name:            "DB_PASSWORD",
				ValueFromSecret: &GcpCloudRunJobSecretEnvSource{Secret: "db-password", Version: "latest"},
			}}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts a secret value the component stores", func() {
			spec := makeValidSpec()
			spec.Template.Containers[0].Env = []*GcpCloudRunJobEnvVar{{Name: "API_TOKEN", SecretValue: "$secret/api-token"}}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects a secret value beside a literal value", func() {
			spec := makeValidSpec()
			spec.Template.Containers[0].Env = []*GcpCloudRunJobEnvVar{{Name: "API_TOKEN", Value: "plain", SecretValue: "$secret/api-token"}}
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
			Expect(violatedRules(err)).To(ContainElement("env.value_xor_secret"))
		})

		It("rejects env var with both value and secret", func() {
			spec := makeValidSpec()
			spec.Template.Containers[0].Env = []*GcpCloudRunJobEnvVar{{
				Name:            "X",
				Value:           "v",
				ValueFromSecret: &GcpCloudRunJobSecretEnvSource{Secret: "s"},
			}}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects env var with invalid name", func() {
			spec := makeValidSpec()
			spec.Template.Containers[0].Env = []*GcpCloudRunJobEnvVar{{Name: "1BAD", Value: "v"}}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("Container resources", func() {
		It("accepts valid cpu and memory", func() {
			spec := makeValidSpec()
			spec.Template.Containers[0].Resources = &GcpCloudRunJobContainerResources{Cpu: "2", Memory: "4Gi"}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects invalid memory format", func() {
			spec := makeValidSpec()
			spec.Template.Containers[0].Resources = &GcpCloudRunJobContainerResources{Memory: "512"}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("Volumes", func() {
		It("accepts a Cloud SQL volume", func() {
			spec := makeValidSpec()
			spec.Template.Volumes = []*GcpCloudRunJobVolume{{
				Name: "cloudsql",
				Source: &GcpCloudRunJobVolume_CloudSqlInstance{
					CloudSqlInstance: &GcpCloudRunJobVolumeCloudSql{
						Instances: []*foreignkeyv1.StringValueOrRef{strVal("p:r:i")},
					},
				},
			}}
			spec.Template.Containers[0].VolumeMounts = []*GcpCloudRunJobVolumeMount{{Name: "cloudsql", MountPath: "/cloudsql"}}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects Cloud SQL volume without instances", func() {
			spec := makeValidSpec()
			spec.Template.Volumes = []*GcpCloudRunJobVolume{{
				Name: "cloudsql",
				Source: &GcpCloudRunJobVolume_CloudSqlInstance{
					CloudSqlInstance: &GcpCloudRunJobVolumeCloudSql{},
				},
			}}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts a secret volume", func() {
			spec := makeValidSpec()
			spec.Template.Volumes = []*GcpCloudRunJobVolume{{
				Name: "certs",
				Source: &GcpCloudRunJobVolume_Secret{
					Secret: &GcpCloudRunJobVolumeSecret{
						Secret: "tls-cert",
						Items:  []*GcpCloudRunJobVolumeSecretItem{{Path: "tls.crt", Version: "latest"}},
					},
				},
			}}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts an empty_dir volume", func() {
			spec := makeValidSpec()
			spec.Template.Volumes = []*GcpCloudRunJobVolume{{
				Name: "scratch",
				Source: &GcpCloudRunJobVolume_EmptyDir{
					EmptyDir: &GcpCloudRunJobVolumeEmptyDir{Medium: "MEMORY", SizeLimit: "256Mi"},
				},
			}}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts a GCS volume", func() {
			spec := makeValidSpec()
			spec.Template.Volumes = []*GcpCloudRunJobVolume{{
				Name: "data",
				Source: &GcpCloudRunJobVolume_Gcs{
					Gcs: &GcpCloudRunJobVolumeGcs{Bucket: strVal("my-bucket"), ReadOnly: true},
				},
			}}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts an NFS volume", func() {
			spec := makeValidSpec()
			spec.Template.Volumes = []*GcpCloudRunJobVolume{{
				Name: "share",
				Source: &GcpCloudRunJobVolume_Nfs{
					Nfs: &GcpCloudRunJobVolumeNfs{Server: "10.0.0.5", Path: "/share1"},
				},
			}}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects volume without source", func() {
			spec := makeValidSpec()
			spec.Template.Volumes = []*GcpCloudRunJobVolume{{Name: "empty"}}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects mount path not absolute", func() {
			spec := makeValidSpec()
			spec.Template.Containers[0].VolumeMounts = []*GcpCloudRunJobVolumeMount{{Name: "v", MountPath: "data"}}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("VPC access", func() {
		It("accepts connector-only vpc access", func() {
			spec := makeValidSpec()
			spec.Template.VpcAccess = &GcpCloudRunJobVpcAccess{Connector: strVal("my-connector")}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts direct VPC network_interfaces", func() {
			spec := makeValidSpec()
			spec.Template.VpcAccess = &GcpCloudRunJobVpcAccess{
				NetworkInterfaces: []*GcpCloudRunJobNetworkInterface{{
					Subnetwork: strVal("my-subnet"),
				}},
				Egress: "PRIVATE_RANGES_ONLY",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects connector and network_interfaces together", func() {
			spec := makeValidSpec()
			spec.Template.VpcAccess = &GcpCloudRunJobVpcAccess{
				Connector: strVal("my-connector"),
				NetworkInterfaces: []*GcpCloudRunJobNetworkInterface{{
					Subnetwork: strVal("my-subnet"),
				}},
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("Binary authorization", func() {
		It("accepts use_default", func() {
			spec := makeValidSpec()
			spec.BinaryAuthorization = &GcpCloudRunJobBinaryAuthorization{UseDefault: true}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects use_default and policy together", func() {
			spec := makeValidSpec()
			spec.BinaryAuthorization = &GcpCloudRunJobBinaryAuthorization{
				UseDefault: true,
				Policy:     "projects/p/platforms/cloudRun/policies/default",
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("Timeout and retries", func() {
		It("accepts valid timeout_seconds", func() {
			spec := makeValidSpec()
			spec.Template.TimeoutSeconds = proto.Int32(3600)
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects timeout_seconds above 86400", func() {
			spec := makeValidSpec()
			spec.Template.TimeoutSeconds = proto.Int32(90000)
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts max_retries of zero", func() {
			spec := makeValidSpec()
			spec.Template.MaxRetries = proto.Int32(0)
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})
	})

	Context("Launch stage", func() {
		It("accepts BETA launch stage", func() {
			spec := makeValidSpec()
			spec.LaunchStage = "BETA"
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects invalid launch stage", func() {
			spec := makeValidSpec()
			spec.LaunchStage = "PREVIEW"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("Startup probe", func() {
		It("accepts a TCP startup probe with a declared port", func() {
			spec := makeValidSpec()
			spec.Template.Containers[0].Ports = &GcpCloudRunJobContainerPort{ContainerPort: proto.Int32(8080)}
			spec.Template.Containers[0].StartupProbe = &GcpCloudRunJobProbe{
				PeriodSeconds:    proto.Int32(2),
				FailureThreshold: proto.Int32(3),
				Handler:          &GcpCloudRunJobProbe_TcpSocket{TcpSocket: &GcpCloudRunJobTcpSocketAction{}},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts an HTTP startup probe with headers", func() {
			spec := makeValidSpec()
			spec.Template.Containers[0].StartupProbe = &GcpCloudRunJobProbe{
				Handler: &GcpCloudRunJobProbe_HttpGet{HttpGet: &GcpCloudRunJobHttpGetAction{
					Path:        "/healthz",
					Port:        proto.Int32(8080),
					HttpHeaders: []*GcpCloudRunJobHttpHeader{{Name: "X-Probe", Value: "1"}},
				}},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts a gRPC startup probe", func() {
			spec := makeValidSpec()
			spec.Template.Containers[0].StartupProbe = &GcpCloudRunJobProbe{
				Handler: &GcpCloudRunJobProbe_Grpc{Grpc: &GcpCloudRunJobGrpcAction{Service: "my.Worker"}},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects a probe without a handler", func() {
			spec := makeValidSpec()
			spec.Template.Containers[0].StartupProbe = &GcpCloudRunJobProbe{PeriodSeconds: proto.Int32(10)}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects timeout exceeding period", func() {
			spec := makeValidSpec()
			spec.Template.Containers[0].StartupProbe = &GcpCloudRunJobProbe{
				TimeoutSeconds: proto.Int32(20),
				PeriodSeconds:  proto.Int32(10),
				Handler:        &GcpCloudRunJobProbe_TcpSocket{TcpSocket: &GcpCloudRunJobTcpSocketAction{}},
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects a startup window exceeding 240 seconds", func() {
			spec := makeValidSpec()
			spec.Template.Containers[0].StartupProbe = &GcpCloudRunJobProbe{
				FailureThreshold: proto.Int32(5),
				PeriodSeconds:    proto.Int32(60),
				Handler:          &GcpCloudRunJobProbe_TcpSocket{TcpSocket: &GcpCloudRunJobTcpSocketAction{}},
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects a period above 240", func() {
			spec := makeValidSpec()
			spec.Template.Containers[0].StartupProbe = &GcpCloudRunJobProbe{
				PeriodSeconds:    proto.Int32(241),
				FailureThreshold: proto.Int32(1),
				Handler:          &GcpCloudRunJobProbe_TcpSocket{TcpSocket: &GcpCloudRunJobTcpSocketAction{}},
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects an unknown port protocol", func() {
			spec := makeValidSpec()
			spec.Template.Containers[0].Ports = &GcpCloudRunJobContainerPort{Name: "tcp"}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("Execution tokens", func() {
		It("accepts a start token alone", func() {
			spec := makeValidSpec()
			spec.StartExecutionToken = "start-once-created"
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts a run token alone", func() {
			spec := makeValidSpec()
			spec.RunExecutionToken = "migrate-v42"
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects both tokens together", func() {
			spec := makeValidSpec()
			spec.StartExecutionToken = "start"
			spec.RunExecutionToken = "run"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("Execution metadata and deletion policy", func() {
		It("accepts execution labels and annotations", func() {
			spec := makeValidSpec()
			spec.ExecutionLabels = map[string]string{"cost-center": "1234"}
			spec.ExecutionAnnotations = map[string]string{"external-tool/trace": "on"}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts each documented deletion_policy value", func() {
			for _, p := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
				spec := makeValidSpec()
				spec.DeletionPolicy = p
				Expect(protovalidate.Validate(spec)).To(BeNil(), "deletion_policy=%s", p)
			}
		})

		It("rejects an unknown deletion_policy", func() {
			spec := makeValidSpec()
			spec.DeletionPolicy = "KEEP"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("Mount refinements", func() {
		It("accepts a GCS volume with mount options and a sub_path mount", func() {
			spec := makeValidSpec()
			spec.Template.Volumes = []*GcpCloudRunJobVolume{{
				Name: "datalake",
				Source: &GcpCloudRunJobVolume_Gcs{Gcs: &GcpCloudRunJobVolumeGcs{
					Bucket:       strVal("my-bucket"),
					ReadOnly:     true,
					MountOptions: []string{"implicit-dirs", "only-dir=input"},
				}},
			}}
			spec.Template.Containers[0].VolumeMounts = []*GcpCloudRunJobVolumeMount{{
				Name:      "datalake",
				MountPath: "/data",
				SubPath:   "batch-42",
			}}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})
	})
})
