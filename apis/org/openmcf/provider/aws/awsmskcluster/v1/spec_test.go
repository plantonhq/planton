package awsmskclusterv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAwsMskClusterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsMskClusterSpec Validation Tests")
}

// helper to build a valid minimal spec used as a baseline for mutation tests
func validMinimalSpec() *AwsMskCluster {
	return &AwsMskCluster{
		ApiVersion: "aws.openmcf.org/v1",
		Kind:       "AwsMskCluster",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-msk-cluster",
		},
		Spec: &AwsMskClusterSpec{
			KafkaVersion:        "3.6.0",
			NumberOfBrokerNodes: 3,
			InstanceType:        "kafka.m5.large",
			SubnetIds: []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-aaa"}},
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-bbb"}},
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-ccc"}},
			},
		},
	}
}

var _ = ginkgo.Describe("AwsMskClusterSpec Validation Tests", func() {

	// ===== HAPPY PATH TESTS =====

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal valid cluster", func() {
			input := validMinimalSpec()
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a cluster with SASL/IAM authentication", func() {
			input := validMinimalSpec()
			input.Spec.Authentication = &AwsMskClusterAuthentication{
				SaslIamEnabled: true,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a cluster with multiple auth methods", func() {
			input := validMinimalSpec()
			input.Spec.Authentication = &AwsMskClusterAuthentication{
				SaslIamEnabled:   true,
				SaslScramEnabled: true,
				TlsEnabled:       true,
				TlsCertificateAuthorityArns: []*foreignkeyv1.StringValueOrRef{
					{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:acm-pca:us-east-1:111122223333:certificate-authority/abc"}},
				},
				Unauthenticated: false,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a cluster with KMS encryption", func() {
			input := validMinimalSpec()
			input.Spec.KmsKeyArn = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:kms:us-east-1:111122223333:key/mrk-abc"},
			}
			input.Spec.ClientBrokerEncryption = proto.String("TLS")
			input.Spec.InClusterEncryption = proto.Bool(true)
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a cluster with EBS storage config", func() {
			input := validMinimalSpec()
			input.Spec.EbsVolumeSizeGib = proto.Int32(500)
			input.Spec.ProvisionedThroughputEnabled = true
			input.Spec.ProvisionedThroughputMbs = 500
			input.Spec.StorageMode = "TIERED"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a cluster with server_properties", func() {
			input := validMinimalSpec()
			input.Spec.ServerProperties = map[string]string{
				"auto.create.topics.enable":  "false",
				"default.replication.factor": "3",
				"min.insync.replicas":        "2",
				"num.partitions":             "6",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a cluster with external configuration reference", func() {
			input := validMinimalSpec()
			input.Spec.ConfigurationArn = "arn:aws:kafka:us-east-1:111122223333:configuration/my-config/abc-123"
			input.Spec.ConfigurationRevision = 1
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a cluster with all three logging destinations", func() {
			input := validMinimalSpec()
			input.Spec.Logging = &AwsMskClusterLogging{
				CloudwatchLogs: &AwsMskClusterCloudwatchLogging{
					Enabled: true,
					LogGroup: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "/aws/msk/my-cluster"},
					},
				},
				Firehose: &AwsMskClusterFirehoseLogging{
					Enabled: true,
					DeliveryStream: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-firehose-stream"},
					},
				},
				S3: &AwsMskClusterS3Logging{
					Enabled: true,
					Bucket: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-msk-logs-bucket"},
					},
					Prefix: "msk-logs/",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a cluster with monitoring config", func() {
			input := validMinimalSpec()
			input.Spec.EnhancedMonitoring = "PER_TOPIC_PER_BROKER"
			input.Spec.JmxExporterEnabled = true
			input.Spec.NodeExporterEnabled = true
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a cluster with public access enabled", func() {
			input := validMinimalSpec()
			input.Spec.PublicAccessType = "SERVICE_PROVIDED_EIPS"
			input.Spec.Authentication = &AwsMskClusterAuthentication{
				SaslIamEnabled: true,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a cluster with networking and managed security group", func() {
			input := validMinimalSpec()
			input.Spec.VpcId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "vpc-abc123"},
			}
			input.Spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-source-123"}},
			}
			input.Spec.AllowedCidrBlocks = []string{"10.0.0.0/16"}
			input.Spec.AssociateSecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-extra-456"}},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a production-ready cluster with full configuration", func() {
			input := validMinimalSpec()
			input.Spec.KafkaVersion = "3.6.0"
			input.Spec.NumberOfBrokerNodes = 6
			input.Spec.InstanceType = "kafka.m7g.xlarge"
			input.Spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-az1"}},
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-az2"}},
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-az3"}},
			}
			input.Spec.EbsVolumeSizeGib = proto.Int32(1000)
			input.Spec.StorageMode = "TIERED"
			input.Spec.KmsKeyArn = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:kms:us-east-1:111122223333:key/mrk-prod"},
			}
			input.Spec.ClientBrokerEncryption = proto.String("TLS")
			input.Spec.InClusterEncryption = proto.Bool(true)
			input.Spec.Authentication = &AwsMskClusterAuthentication{
				SaslIamEnabled: true,
			}
			input.Spec.ServerProperties = map[string]string{
				"auto.create.topics.enable":  "false",
				"default.replication.factor": "3",
				"min.insync.replicas":        "2",
			}
			input.Spec.Logging = &AwsMskClusterLogging{
				CloudwatchLogs: &AwsMskClusterCloudwatchLogging{
					Enabled: true,
					LogGroup: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "/aws/msk/prod-cluster"},
					},
				},
			}
			input.Spec.EnhancedMonitoring = "PER_TOPIC_PER_BROKER"
			input.Spec.JmxExporterEnabled = true
			input.Spec.NodeExporterEnabled = true
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// ===== FAILURE TESTS =====

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.Context("required fields", func() {
			ginkgo.It("should fail when kafka_version is empty", func() {
				input := validMinimalSpec()
				input.Spec.KafkaVersion = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should fail when instance_type is empty", func() {
				input := validMinimalSpec()
				input.Spec.InstanceType = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should fail when number_of_broker_nodes is zero", func() {
				input := validMinimalSpec()
				input.Spec.NumberOfBrokerNodes = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should fail when subnet_ids is empty", func() {
				input := validMinimalSpec()
				input.Spec.SubnetIds = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("storage constraints", func() {
			ginkgo.It("should fail when ebs_volume_size_gib exceeds 16384", func() {
				input := validMinimalSpec()
				input.Spec.EbsVolumeSizeGib = proto.Int32(20000)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should fail when provisioned_throughput_mbs is below 250", func() {
				input := validMinimalSpec()
				input.Spec.ProvisionedThroughputMbs = 100
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should fail when provisioned_throughput_mbs is above 2375", func() {
				input := validMinimalSpec()
				input.Spec.ProvisionedThroughputMbs = 3000
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("provisioned_throughput_requires_mbs", func() {
			ginkgo.It("should fail when provisioned_throughput_enabled is true but mbs is zero", func() {
				input := validMinimalSpec()
				input.Spec.ProvisionedThroughputEnabled = true
				input.Spec.ProvisionedThroughputMbs = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("provisioned_throughput_mbs must be set"))
			})
		})

		ginkgo.Context("invalid enum values", func() {
			ginkgo.It("should fail for invalid storage_mode", func() {
				input := validMinimalSpec()
				input.Spec.StorageMode = "INVALID"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should fail for invalid client_broker_encryption", func() {
				input := validMinimalSpec()
				input.Spec.ClientBrokerEncryption = proto.String("AES256")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should fail for invalid enhanced_monitoring", func() {
				input := validMinimalSpec()
				input.Spec.EnhancedMonitoring = "ULTRA_DETAILED"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should fail for invalid public_access_type", func() {
				input := validMinimalSpec()
				input.Spec.PublicAccessType = "PUBLIC"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("configuration_mutual_exclusion", func() {
			ginkgo.It("should fail when both configuration_arn and server_properties are set", func() {
				input := validMinimalSpec()
				input.Spec.ConfigurationArn = "arn:aws:kafka:us-east-1:111122223333:configuration/my-config/abc-123"
				input.Spec.ConfigurationRevision = 1
				input.Spec.ServerProperties = map[string]string{
					"auto.create.topics.enable": "false",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("configuration_arn and server_properties are mutually exclusive"))
			})
		})

		ginkgo.Context("configuration_revision_required", func() {
			ginkgo.It("should fail when configuration_arn is set without configuration_revision", func() {
				input := validMinimalSpec()
				input.Spec.ConfigurationArn = "arn:aws:kafka:us-east-1:111122223333:configuration/my-config/abc-123"
				input.Spec.ConfigurationRevision = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("configuration_revision is required"))
			})
		})

		ginkgo.Context("logging validations", func() {
			ginkgo.It("should fail when CloudWatch logging enabled without log_group", func() {
				input := validMinimalSpec()
				input.Spec.Logging = &AwsMskClusterLogging{
					CloudwatchLogs: &AwsMskClusterCloudwatchLogging{
						Enabled: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("log_group is required"))
			})

			ginkgo.It("should fail when Firehose logging enabled without delivery_stream", func() {
				input := validMinimalSpec()
				input.Spec.Logging = &AwsMskClusterLogging{
					Firehose: &AwsMskClusterFirehoseLogging{
						Enabled: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("delivery_stream is required"))
			})

			ginkgo.It("should fail when S3 logging enabled without bucket", func() {
				input := validMinimalSpec()
				input.Spec.Logging = &AwsMskClusterLogging{
					S3: &AwsMskClusterS3Logging{
						Enabled: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("bucket is required"))
			})
		})

		ginkgo.Context("CIDR validation", func() {
			ginkgo.It("should fail for invalid CIDR in allowed_cidr_blocks", func() {
				input := validMinimalSpec()
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
				input := &AwsMskCluster{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsMskCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-msk-cluster",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
