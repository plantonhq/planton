package awsmwaaenvironmentv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAwsMwaaEnvironmentSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsMwaaEnvironmentSpec Validation Tests")
}

func validMinimalSpec() *AwsMwaaEnvironment {
	return &AwsMwaaEnvironment{
		ApiVersion: "aws.openmcf.org/v1",
		Kind:       "AwsMwaaEnvironment",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-mwaa-env",
		},
		Spec: &AwsMwaaEnvironmentSpec{
			Region: "us-west-2",
			SourceBucketArn: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:s3:::my-airflow-bucket"},
			},
			DagS3Path: "dags/",
			ExecutionRoleArn: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::111122223333:role/mwaa-execution-role"},
			},
			SubnetIds: []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-aaa"}},
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-bbb"}},
			},
			AssociateSecurityGroupIds: []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-mwaa-123"}},
			},
		},
	}
}

var _ = ginkgo.Describe("AwsMwaaEnvironmentSpec Validation Tests", func() {

	// ===== HAPPY PATH TESTS =====

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal valid environment", func() {
			input := validMinimalSpec()
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an environment with airflow_version specified", func() {
			input := validMinimalSpec()
			input.Spec.AirflowVersion = "2.10.1"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an environment with airflow_configuration_options", func() {
			input := validMinimalSpec()
			input.Spec.AirflowConfigurationOptions = map[string]string{
				"core.default_timezone":      "utc",
				"webserver.dag_default_view": "grid",
				"celery.worker_autoscale":    "10,2",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an environment with plugins_s3_path and requirements_s3_path", func() {
			input := validMinimalSpec()
			input.Spec.PluginsS3Path = "plugins/plugins.zip"
			input.Spec.RequirementsS3Path = "requirements/requirements.txt"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an environment with startup_script_s3_path", func() {
			input := validMinimalSpec()
			input.Spec.StartupScriptS3Path = "scripts/startup.sh"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an environment with S3 object version pinning", func() {
			input := validMinimalSpec()
			input.Spec.PluginsS3Path = "plugins/plugins.zip"
			input.Spec.PluginsS3ObjectVersion = "abc123def456"
			input.Spec.RequirementsS3Path = "requirements/requirements.txt"
			input.Spec.RequirementsS3ObjectVersion = "ghi789jkl012"
			input.Spec.StartupScriptS3Path = "scripts/startup.sh"
			input.Spec.StartupScriptS3ObjectVersion = "mno345pqr678"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an environment with KMS encryption", func() {
			input := validMinimalSpec()
			input.Spec.KmsKeyArn = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:kms:us-east-1:111122223333:key/mrk-abc"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an environment with environment_class mw1.medium", func() {
			input := validMinimalSpec()
			input.Spec.EnvironmentClass = "mw1.medium"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an environment with min/max workers specified", func() {
			input := validMinimalSpec()
			input.Spec.MinWorkers = 2
			input.Spec.MaxWorkers = 10
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an environment with min/max webservers specified", func() {
			input := validMinimalSpec()
			input.Spec.MinWebservers = 2
			input.Spec.MaxWebservers = 4
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an environment with schedulers specified", func() {
			input := validMinimalSpec()
			input.Spec.Schedulers = 3
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an environment with webserver_access_mode PUBLIC_ONLY", func() {
			input := validMinimalSpec()
			input.Spec.WebserverAccessMode = proto.String("PUBLIC_ONLY")
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an environment with endpoint_management CUSTOMER", func() {
			input := validMinimalSpec()
			input.Spec.EndpointManagement = "CUSTOMER"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an environment with logging_configuration (all 5 modules enabled at INFO)", func() {
			input := validMinimalSpec()
			input.Spec.LoggingConfiguration = &AwsMwaaEnvironmentLoggingConfiguration{
				DagProcessingLogs: &AwsMwaaEnvironmentLoggingModuleConfig{
					Enabled:  true,
					LogLevel: "INFO",
				},
				SchedulerLogs: &AwsMwaaEnvironmentLoggingModuleConfig{
					Enabled:  true,
					LogLevel: "INFO",
				},
				TaskLogs: &AwsMwaaEnvironmentLoggingModuleConfig{
					Enabled:  true,
					LogLevel: "INFO",
				},
				WebserverLogs: &AwsMwaaEnvironmentLoggingModuleConfig{
					Enabled:  true,
					LogLevel: "INFO",
				},
				WorkerLogs: &AwsMwaaEnvironmentLoggingModuleConfig{
					Enabled:  true,
					LogLevel: "INFO",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an environment with logging_configuration (single module at DEBUG)", func() {
			input := validMinimalSpec()
			input.Spec.LoggingConfiguration = &AwsMwaaEnvironmentLoggingConfiguration{
				TaskLogs: &AwsMwaaEnvironmentLoggingModuleConfig{
					Enabled:  true,
					LogLevel: "DEBUG",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an environment with weekly_maintenance_window_start", func() {
			input := validMinimalSpec()
			input.Spec.WeeklyMaintenanceWindowStart = "TUE:03:30"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an environment with worker_replacement_strategy GRACEFUL", func() {
			input := validMinimalSpec()
			input.Spec.WorkerReplacementStrategy = "GRACEFUL"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an environment with managed security group (vpc_id + security_group_ids)", func() {
			input := validMinimalSpec()
			input.Spec.VpcId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "vpc-abc123"},
			}
			input.Spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-source-123"}},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an environment with managed security group + allowed_cidr_blocks", func() {
			input := validMinimalSpec()
			input.Spec.VpcId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "vpc-abc123"},
			}
			input.Spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-source-123"}},
			}
			input.Spec.AllowedCidrBlocks = []string{"10.0.0.0/16", "172.16.0.0/12"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an environment with valueFrom references", func() {
			input := validMinimalSpec()
			input.Spec.SourceBucketArn = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{
						Kind: cloudresourcekind.CloudResourceKind_AwsS3Bucket,
						Name: "my-s3-bucket",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a production-ready environment with full configuration", func() {
			input := validMinimalSpec()
			input.Spec.AirflowVersion = "2.10.1"
			input.Spec.AirflowConfigurationOptions = map[string]string{
				"core.default_timezone":      "utc",
				"webserver.dag_default_view": "grid",
			}
			input.Spec.PluginsS3Path = "plugins/plugins.zip"
			input.Spec.PluginsS3ObjectVersion = "v1abc"
			input.Spec.RequirementsS3Path = "requirements/requirements.txt"
			input.Spec.RequirementsS3ObjectVersion = "v2def"
			input.Spec.StartupScriptS3Path = "scripts/startup.sh"
			input.Spec.StartupScriptS3ObjectVersion = "v3ghi"
			input.Spec.KmsKeyArn = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:kms:us-east-1:111122223333:key/mrk-prod"},
			}
			input.Spec.EnvironmentClass = "mw1.large"
			input.Spec.MinWorkers = 2
			input.Spec.MaxWorkers = 25
			input.Spec.MinWebservers = 2
			input.Spec.MaxWebservers = 5
			input.Spec.Schedulers = 3
			input.Spec.WebserverAccessMode = proto.String("PRIVATE_ONLY")
			input.Spec.EndpointManagement = "SERVICE"
			input.Spec.LoggingConfiguration = &AwsMwaaEnvironmentLoggingConfiguration{
				DagProcessingLogs: &AwsMwaaEnvironmentLoggingModuleConfig{
					Enabled:  true,
					LogLevel: "INFO",
				},
				SchedulerLogs: &AwsMwaaEnvironmentLoggingModuleConfig{
					Enabled:  true,
					LogLevel: "INFO",
				},
				TaskLogs: &AwsMwaaEnvironmentLoggingModuleConfig{
					Enabled:  true,
					LogLevel: "WARNING",
				},
				WebserverLogs: &AwsMwaaEnvironmentLoggingModuleConfig{
					Enabled:  true,
					LogLevel: "ERROR",
				},
				WorkerLogs: &AwsMwaaEnvironmentLoggingModuleConfig{
					Enabled:  true,
					LogLevel: "INFO",
				},
			}
			input.Spec.WeeklyMaintenanceWindowStart = "SUN:00:00"
			input.Spec.WorkerReplacementStrategy = "GRACEFUL"
			input.Spec.VpcId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "vpc-prod-123"},
			}
			input.Spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-prod-source"}},
			}
			input.Spec.AllowedCidrBlocks = []string{"10.0.0.0/8"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// ===== FAILURE TESTS =====

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.Context("required fields", func() {
			ginkgo.It("should fail when source_bucket_arn is missing", func() {
				input := validMinimalSpec()
				input.Spec.SourceBucketArn = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should fail when dag_s3_path is empty", func() {
				input := validMinimalSpec()
				input.Spec.DagS3Path = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should fail when execution_role_arn is missing", func() {
				input := validMinimalSpec()
				input.Spec.ExecutionRoleArn = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should fail when fewer than 2 subnet_ids are provided", func() {
				input := validMinimalSpec()
				input.Spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
					{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-aaa"}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("CEL validations", func() {
			ginkgo.It("should fail when dag_s3_path starts with /", func() {
				input := validMinimalSpec()
				input.Spec.DagS3Path = "/dags/"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("dag_s3_path must be a relative path"))
			})

			ginkgo.It("should fail when max_workers < min_workers", func() {
				input := validMinimalSpec()
				input.Spec.MinWorkers = 10
				input.Spec.MaxWorkers = 5
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("max_workers must be >= min_workers"))
			})

			ginkgo.It("should fail when max_webservers < min_webservers", func() {
				input := validMinimalSpec()
				input.Spec.MinWebservers = 4
				input.Spec.MaxWebservers = 2
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("max_webservers must be >= min_webservers"))
			})

			ginkgo.It("should fail when no security coverage is provided", func() {
				input := validMinimalSpec()
				input.Spec.VpcId = nil
				input.Spec.AssociateSecurityGroupIds = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("at least one of vpc_id"))
			})
		})

		ginkgo.Context("invalid enum values", func() {
			ginkgo.It("should fail for invalid environment_class", func() {
				input := validMinimalSpec()
				input.Spec.EnvironmentClass = "mw1.jumbo"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should fail for invalid webserver_access_mode", func() {
				input := validMinimalSpec()
				input.Spec.WebserverAccessMode = proto.String("HYBRID")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should fail for invalid endpoint_management", func() {
				input := validMinimalSpec()
				input.Spec.EndpointManagement = "SELF_MANAGED"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should fail for invalid worker_replacement_strategy", func() {
				input := validMinimalSpec()
				input.Spec.WorkerReplacementStrategy = "ROLLING"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("scheduler constraints", func() {
			ginkgo.It("should fail when schedulers < 2", func() {
				input := validMinimalSpec()
				input.Spec.Schedulers = 1
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("logging validations", func() {
			ginkgo.It("should fail for invalid log_level", func() {
				input := validMinimalSpec()
				input.Spec.LoggingConfiguration = &AwsMwaaEnvironmentLoggingConfiguration{
					TaskLogs: &AwsMwaaEnvironmentLoggingModuleConfig{
						Enabled:  true,
						LogLevel: "TRACE",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("CIDR validation", func() {
			ginkgo.It("should fail for invalid CIDR block format", func() {
				input := validMinimalSpec()
				input.Spec.VpcId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "vpc-abc123"},
				}
				input.Spec.AllowedCidrBlocks = []string{"not-a-cidr"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("API envelope validation", func() {
			ginkgo.It("should fail for wrong apiVersion", func() {
				input := validMinimalSpec()
				input.ApiVersion = "gcp.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should fail for wrong kind", func() {
				input := validMinimalSpec()
				input.Kind = "AwsRedshiftCluster"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should fail when metadata is missing", func() {
				input := validMinimalSpec()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should fail when spec is missing", func() {
				input := &AwsMwaaEnvironment{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsMwaaEnvironment",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-mwaa-env",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should accept a valid complete envelope", func() {
				input := validMinimalSpec()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
