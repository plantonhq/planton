package awsdocumentdbv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAwsDocumentDbSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsDocumentDbSpec Validation Suite")
}

// Helper functions for pointer types
func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

var _ = ginkgo.Describe("AwsDocumentDbSpec validations", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with minimal valid fields using subnets", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsDocumentDbSpec{
					Region: "us-west-2",
					Subnets: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					MasterPassword:             "MySecureP@ssword123",
					SkipFinalSnapshot:          boolPtr(true),
					PreferredBackupWindow:      "03:00-04:00",
					PreferredMaintenanceWindow: "sun:05:00-sun:06:00",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with db_subnet_group instead of subnets", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsDocumentDbSpec{
					Region: "us-west-2",
					DbSubnetGroup: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-subnet-group"},
					},
					MasterPassword:    "MySecureP@ssword123",
					SkipFinalSnapshot: boolPtr(true),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with final_snapshot_identifier when skip_final_snapshot is false", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsDocumentDbSpec{
					Region: "us-west-2",
					Subnets: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					MasterPassword:          "MySecureP@ssword123",
					SkipFinalSnapshot:       boolPtr(false),
					FinalSnapshotIdentifier: "my-final-snapshot",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with valid cloudwatch logs exports", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsDocumentDbSpec{
					Region: "us-west-2",
					Subnets: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					MasterPassword:               "MySecureP@ssword123",
					SkipFinalSnapshot:            boolPtr(true),
					EnabledCloudwatchLogsExports: []string{"audit", "profiler"},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with valid CIDR blocks", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsDocumentDbSpec{
					Region: "us-west-2",
					Subnets: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					MasterPassword:    "MySecureP@ssword123",
					SkipFinalSnapshot: boolPtr(true),
					AllowedCidrs:      []string{"10.0.0.0/16", "192.168.1.0/24"},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("with missing master_password", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsDocumentDbSpec{
					Region: "us-west-2",
					Subnets: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					SkipFinalSnapshot: boolPtr(true),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with only one subnet and no db_subnet_group", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsDocumentDbSpec{
					Region: "us-west-2",
					Subnets: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
					},
					MasterPassword:    "MySecureP@ssword123",
					SkipFinalSnapshot: boolPtr(true),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with skip_final_snapshot false and missing final_snapshot_identifier", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsDocumentDbSpec{
					Region: "us-west-2",
					Subnets: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					MasterPassword:    "MySecureP@ssword123",
					SkipFinalSnapshot: boolPtr(false),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid cloudwatch logs exports", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsDocumentDbSpec{
					Region: "us-west-2",
					Subnets: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					MasterPassword:               "MySecureP@ssword123",
					SkipFinalSnapshot:            boolPtr(true),
					EnabledCloudwatchLogsExports: []string{"audit", "invalid_log_type"},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid CIDR block format", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsDocumentDbSpec{
					Region: "us-west-2",
					Subnets: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					MasterPassword:    "MySecureP@ssword123",
					SkipFinalSnapshot: boolPtr(true),
					AllowedCidrs:      []string{"invalid-cidr"},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with duplicate CIDR blocks", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsDocumentDbSpec{
					Region: "us-west-2",
					Subnets: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					MasterPassword:    "MySecureP@ssword123",
					SkipFinalSnapshot: boolPtr(true),
					AllowedCidrs:      []string{"10.0.0.0/16", "10.0.0.0/16"},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid backup_retention_period", func() {
			ginkgo.It("should return a validation error when less than 1", func() {
				spec := &AwsDocumentDbSpec{
					Region: "us-west-2",
					Subnets: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					MasterPassword:        "MySecureP@ssword123",
					SkipFinalSnapshot:     boolPtr(true),
					BackupRetentionPeriod: int32Ptr(0),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when greater than 35", func() {
				spec := &AwsDocumentDbSpec{
					Region: "us-west-2",
					Subnets: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					MasterPassword:        "MySecureP@ssword123",
					SkipFinalSnapshot:     boolPtr(true),
					BackupRetentionPeriod: int32Ptr(36),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid port", func() {
			ginkgo.It("should return a validation error when port is 0", func() {
				spec := &AwsDocumentDbSpec{
					Region: "us-west-2",
					Subnets: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					MasterPassword:    "MySecureP@ssword123",
					SkipFinalSnapshot: boolPtr(true),
					Port:              int32Ptr(0),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when port exceeds 65535", func() {
				spec := &AwsDocumentDbSpec{
					Region: "us-west-2",
					Subnets: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					MasterPassword:    "MySecureP@ssword123",
					SkipFinalSnapshot: boolPtr(true),
					Port:              int32Ptr(70000),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid preferred_backup_window format", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsDocumentDbSpec{
					Region: "us-west-2",
					Subnets: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					MasterPassword:        "MySecureP@ssword123",
					SkipFinalSnapshot:     boolPtr(true),
					PreferredBackupWindow: "invalid-format",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid preferred_maintenance_window format", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsDocumentDbSpec{
					Region: "us-west-2",
					Subnets: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					MasterPassword:             "MySecureP@ssword123",
					SkipFinalSnapshot:          boolPtr(true),
					PreferredMaintenanceWindow: "invalid-format",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid apply_method in cluster_parameters", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsDocumentDbSpec{
					Region: "us-west-2",
					Subnets: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
					},
					MasterPassword:    "MySecureP@ssword123",
					SkipFinalSnapshot: boolPtr(true),
					ClusterParameters: []*AwsDocumentDbParameter{
						{
							Name:        "tls",
							Value:       "enabled",
							ApplyMethod: "invalid-method",
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
