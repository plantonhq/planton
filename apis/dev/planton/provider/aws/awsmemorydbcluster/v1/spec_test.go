package awsmemorydbclusterv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAwsMemorydbClusterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsMemorydbClusterSpec Validation Suite")
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

func svr(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

var _ = ginkgo.Describe("AwsMemorydbClusterSpec validations", func() {

	// -----------------------------------------------------------------
	// Valid inputs
	// -----------------------------------------------------------------
	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.Context("with minimal required fields", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsMemorydbClusterSpec{
					Region:   "us-west-2",
					Engine:   "redis",
					NodeType: "db.t4g.small",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with full production configuration", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsMemorydbClusterSpec{
					Region:              "us-west-2",
					Engine:              "redis",
					EngineVersion:       "7.1",
					Description:         "Production session store",
					NodeType:            "db.r7g.large",
					Port:                int32Ptr(6379),
					NumShards:           int32Ptr(2),
					NumReplicasPerShard: int32Ptr(2),
					AclName:             stringPtr("my-prod-acl"),
					SubnetIds: []*foreignkeyv1.StringValueOrRef{
						svr("subnet-11111111"),
						svr("subnet-22222222"),
					},
					SecurityGroupIds: []*foreignkeyv1.StringValueOrRef{
						svr("sg-12345678"),
					},
					TlsEnabled:             boolPtr(true),
					KmsKeyId:               svr("arn:aws:kms:us-east-1:123456789012:key/abc-123"),
					MaintenanceWindow:      "sun:05:00-sun:06:00",
					SnapshotRetentionLimit: 7,
					SnapshotWindow:         "03:00-04:00",
					FinalSnapshotName:      "final-snap",
					ParameterGroupFamily:   "memorydb_redis7",
					Parameters: []*AwsMemorydbClusterParameter{
						{Name: "activedefrag", Value: "yes"},
					},
					SnsTopicArn:             svr("arn:aws:sns:us-east-1:123456789012:alerts"),
					AutoMinorVersionUpgrade: boolPtr(true),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with valkey engine", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsMemorydbClusterSpec{
					Region:   "us-west-2",
					Engine:   "valkey",
					NodeType: "db.r7g.large",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with tls_enabled=false and open-access ACL", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsMemorydbClusterSpec{
					Region:     "us-west-2",
					Engine:     "redis",
					NodeType:   "db.t4g.small",
					TlsEnabled: boolPtr(false),
					AclName:    stringPtr("open-access"),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with snapshot restore from ARNs", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsMemorydbClusterSpec{
					Region:       "us-west-2",
					Engine:       "redis",
					NodeType:     "db.r7g.large",
					SnapshotArns: []string{"arn:aws:s3:::my-bucket/snapshot.rdb"},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with data tiering enabled", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsMemorydbClusterSpec{
					Region:      "us-west-2",
					Engine:      "redis",
					NodeType:    "db.r6gd.xlarge",
					DataTiering: true,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	// -----------------------------------------------------------------
	// Invalid inputs
	// -----------------------------------------------------------------
	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.Context("with missing engine", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsMemorydbClusterSpec{
					Region:   "us-west-2",
					NodeType: "db.t4g.small",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with missing node_type", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsMemorydbClusterSpec{
					Region: "us-west-2",
					Engine: "redis",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid engine value", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsMemorydbClusterSpec{
					Region:   "us-west-2",
					Engine:   "memcached",
					NodeType: "db.t4g.small",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with tls_enabled=false and non-open-access ACL", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsMemorydbClusterSpec{
					Region:     "us-west-2",
					Engine:     "redis",
					NodeType:   "db.t4g.small",
					TlsEnabled: boolPtr(false),
					AclName:    stringPtr("my-custom-acl"),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with parameters but no parameter_group_family", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsMemorydbClusterSpec{
					Region:   "us-west-2",
					Engine:   "redis",
					NodeType: "db.t4g.small",
					Parameters: []*AwsMemorydbClusterParameter{
						{Name: "activedefrag", Value: "yes"},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with both snapshot_arns and snapshot_name", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsMemorydbClusterSpec{
					Region:       "us-west-2",
					Engine:       "redis",
					NodeType:     "db.t4g.small",
					SnapshotArns: []string{"arn:aws:s3:::bucket/snap.rdb"},
					SnapshotName: "my-snapshot",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid port (out of range)", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsMemorydbClusterSpec{
					Region:   "us-west-2",
					Engine:   "redis",
					NodeType: "db.t4g.small",
					Port:     int32Ptr(0),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with num_replicas_per_shard out of range", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsMemorydbClusterSpec{
					Region:              "us-west-2",
					Engine:              "redis",
					NodeType:            "db.t4g.small",
					NumReplicasPerShard: int32Ptr(6),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with snapshot_retention_limit out of range", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsMemorydbClusterSpec{
					Region:                 "us-west-2",
					Engine:                 "redis",
					NodeType:               "db.t4g.small",
					SnapshotRetentionLimit: 36,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid maintenance_window format", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsMemorydbClusterSpec{
					Region:            "us-west-2",
					Engine:            "redis",
					NodeType:          "db.t4g.small",
					MaintenanceWindow: "invalid-format",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid snapshot_window format", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsMemorydbClusterSpec{
					Region:         "us-west-2",
					Engine:         "redis",
					NodeType:       "db.t4g.small",
					SnapshotWindow: "bad-window",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with parameter missing name", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsMemorydbClusterSpec{
					Region:               "us-west-2",
					Engine:               "redis",
					NodeType:             "db.t4g.small",
					ParameterGroupFamily: "memorydb_redis7",
					Parameters: []*AwsMemorydbClusterParameter{
						{Value: "yes"},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
