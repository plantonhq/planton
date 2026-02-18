package awsredshiftclusterv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAwsRedshiftClusterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsRedshiftClusterSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("AwsRedshiftClusterSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_redshift_cluster", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &AwsRedshiftCluster{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsRedshiftCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-redshift-cluster",
					},
					Spec: &AwsRedshiftClusterSpec{
						Region:   "us-west-2",
						NodeType: "dc2.large",
						SubnetIds: []*foreignkeyv1.StringValueOrRef{
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"},
							},
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"},
							},
						},
						SkipFinalSnapshot: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.Context("password_mutual_exclusion", func() {
			ginkgo.It("should return a validation error when manage_master_password is true and master_password is set", func() {
				input := &AwsRedshiftCluster{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsRedshiftCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-redshift-cluster",
					},
					Spec: &AwsRedshiftClusterSpec{
						Region:   "us-west-2",
						NodeType: "dc2.large",
						SubnetIds: []*foreignkeyv1.StringValueOrRef{
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"},
							},
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"},
							},
						},
						ManageMasterPassword: true,
						MasterPassword:       "SomeP@ss1",
						SkipFinalSnapshot:    true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("master_password cannot be set when manage_master_password is true"))
			})
		})

		ginkgo.Context("final_snapshot_required", func() {
			ginkgo.It("should return a validation error when skip_final_snapshot is false and final_snapshot_identifier is empty", func() {
				input := &AwsRedshiftCluster{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsRedshiftCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-redshift-cluster",
					},
					Spec: &AwsRedshiftClusterSpec{
						Region:   "us-west-2",
						NodeType: "dc2.large",
						SubnetIds: []*foreignkeyv1.StringValueOrRef{
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"},
							},
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"},
							},
						},
						SkipFinalSnapshot:       false,
						FinalSnapshotIdentifier: "",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("final_snapshot_identifier is required when skip_final_snapshot is false"))
			})
		})

		ginkgo.Context("subnets_or_group", func() {
			ginkgo.It("should return a validation error when neither subnet_ids nor cluster_subnet_group_name is provided", func() {
				input := &AwsRedshiftCluster{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsRedshiftCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-redshift-cluster",
					},
					Spec: &AwsRedshiftClusterSpec{
						Region:                  "us-west-2",
						NodeType:                "dc2.large",
						SkipFinalSnapshot:       true,
						FinalSnapshotIdentifier: "",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("provide either subnet_ids"))
			})
		})

		ginkgo.Context("logging s3_bucket_required", func() {
			ginkgo.It("should return a validation error when log_destination_type is s3 and s3_bucket_name is empty", func() {
				input := &AwsRedshiftCluster{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsRedshiftCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-redshift-cluster",
					},
					Spec: &AwsRedshiftClusterSpec{
						Region:   "us-west-2",
						NodeType: "dc2.large",
						SubnetIds: []*foreignkeyv1.StringValueOrRef{
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"},
							},
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"},
							},
						},
						SkipFinalSnapshot: true,
						Logging: &AwsRedshiftClusterLogging{
							LogDestinationType: "s3",
							S3BucketName:       "",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("s3_bucket_name is required when log_destination_type is"))
			})
		})

		ginkgo.Context("logging cloudwatch_exports_required", func() {
			ginkgo.It("should return a validation error when log_destination_type is cloudwatch and log_exports is empty", func() {
				input := &AwsRedshiftCluster{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsRedshiftCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-redshift-cluster",
					},
					Spec: &AwsRedshiftClusterSpec{
						Region:   "us-west-2",
						NodeType: "dc2.large",
						SubnetIds: []*foreignkeyv1.StringValueOrRef{
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"},
							},
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"},
							},
						},
						SkipFinalSnapshot: true,
						Logging: &AwsRedshiftClusterLogging{
							LogDestinationType: "cloudwatch",
							LogExports:         []string{},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("log_exports must have at least one entry when log_destination_type is"))
			})
		})
	})
})
