package awsneptuneclusterv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAwsNeptuneClusterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsNeptuneClusterSpec Validation Suite")
}

func int32Ptr(i int32) *int32 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	return &s
}

var _ = ginkgo.Describe("AwsNeptuneClusterSpec validations", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with minimal valid fields using subnet_ids", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsNeptuneClusterSpec{
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					SkipFinalSnapshot: boolPtr(true),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with neptune_subnet_group_name instead of subnet_ids", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsNeptuneClusterSpec{
					NeptuneSubnetGroupName: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-subnet-group"},
					},
					SkipFinalSnapshot: boolPtr(true),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with final_snapshot_identifier when skip_final_snapshot is false", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsNeptuneClusterSpec{
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					SkipFinalSnapshot:       boolPtr(false),
					FinalSnapshotIdentifier: "my-neptune-final-snapshot",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with valid cloudwatch logs exports", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsNeptuneClusterSpec{
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					SkipFinalSnapshot:            boolPtr(true),
					EnabledCloudwatchLogsExports: []string{"audit", "slowquery"},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with valid CIDR blocks", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsNeptuneClusterSpec{
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					SkipFinalSnapshot: boolPtr(true),
					AllowedCidrBlocks: []string{"10.0.0.0/16", "192.168.1.0/24"},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with valid storage_type", func() {
			ginkgo.It("should not return a validation error for iopt1", func() {
				spec := &AwsNeptuneClusterSpec{
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					SkipFinalSnapshot: boolPtr(true),
					StorageType:       "iopt1",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with valid cluster parameters", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsNeptuneClusterSpec{
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					SkipFinalSnapshot: boolPtr(true),
					ClusterParameters: []*AwsNeptuneClusterParameter{
						{
							Name:        "neptune_enable_audit_log",
							Value:       "1",
							ApplyMethod: "pending-reboot",
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with serverless configuration", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsNeptuneClusterSpec{
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					SkipFinalSnapshot: boolPtr(true),
					InstanceClass:     stringPtr("db.serverless"),
					ServerlessV2Scaling: &AwsNeptuneClusterServerlessV2ScalingConfiguration{
						MinCapacity: 2.5,
						MaxCapacity: 128.0,
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("with only one subnet and no neptune_subnet_group_name", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsNeptuneClusterSpec{
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
					},
					SkipFinalSnapshot: boolPtr(true),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with skip_final_snapshot false and missing final_snapshot_identifier", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsNeptuneClusterSpec{
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					SkipFinalSnapshot: boolPtr(false),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid cloudwatch logs exports", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsNeptuneClusterSpec{
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					SkipFinalSnapshot:            boolPtr(true),
					EnabledCloudwatchLogsExports: []string{"audit", "invalid_type"},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid CIDR block format", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsNeptuneClusterSpec{
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					SkipFinalSnapshot: boolPtr(true),
					AllowedCidrBlocks: []string{"invalid-cidr"},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with duplicate CIDR blocks", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsNeptuneClusterSpec{
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					SkipFinalSnapshot: boolPtr(true),
					AllowedCidrBlocks: []string{"10.0.0.0/16", "10.0.0.0/16"},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid backup_retention_period", func() {
			ginkgo.It("should return a validation error when less than 1", func() {
				spec := &AwsNeptuneClusterSpec{
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					SkipFinalSnapshot:     boolPtr(true),
					BackupRetentionPeriod: int32Ptr(0),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when greater than 35", func() {
				spec := &AwsNeptuneClusterSpec{
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					SkipFinalSnapshot:     boolPtr(true),
					BackupRetentionPeriod: int32Ptr(36),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid port", func() {
			ginkgo.It("should return a validation error when port is 0", func() {
				spec := &AwsNeptuneClusterSpec{
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					SkipFinalSnapshot: boolPtr(true),
					Port:              int32Ptr(0),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when port exceeds 65535", func() {
				spec := &AwsNeptuneClusterSpec{
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					SkipFinalSnapshot: boolPtr(true),
					Port:              int32Ptr(70000),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid preferred_backup_window format", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsNeptuneClusterSpec{
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					SkipFinalSnapshot:     boolPtr(true),
					PreferredBackupWindow: "invalid-format",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid preferred_maintenance_window format", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsNeptuneClusterSpec{
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					SkipFinalSnapshot:          boolPtr(true),
					PreferredMaintenanceWindow: "invalid-format",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid storage_type", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsNeptuneClusterSpec{
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					SkipFinalSnapshot: boolPtr(true),
					StorageType:       "invalid-type",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid apply_method in cluster_parameters", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsNeptuneClusterSpec{
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					SkipFinalSnapshot: boolPtr(true),
					ClusterParameters: []*AwsNeptuneClusterParameter{
						{
							Name:        "neptune_enable_audit_log",
							Value:       "1",
							ApplyMethod: "invalid-method",
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with serverless_v2_scaling where max < min", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsNeptuneClusterSpec{
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					SkipFinalSnapshot: boolPtr(true),
					InstanceClass:     stringPtr("db.serverless"),
					ServerlessV2Scaling: &AwsNeptuneClusterServerlessV2ScalingConfiguration{
						MinCapacity: 64.0,
						MaxCapacity: 16.0,
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
