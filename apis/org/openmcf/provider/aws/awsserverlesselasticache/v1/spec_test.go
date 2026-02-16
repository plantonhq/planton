package awsserverlesselasticachev1

import (
	"testing"

	"buf.build/go/protovalidate"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestAwsServerlessElasticacheSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsServerlessElasticacheSpec Validation Suite")
}

// helper to create a StringValueOrRef with a literal value.
func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

var _ = ginkgo.Describe("AwsServerlessElasticacheSpec validations", func() {
	var spec *AwsServerlessElasticacheSpec

	ginkgo.BeforeEach(func() {
		// Minimal valid spec: Redis serverless cache with only required field.
		spec = &AwsServerlessElasticacheSpec{
			Engine: "redis",
		}
	})

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal Redis serverless cache", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a minimal Valkey serverless cache", func() {
		spec.Engine = "valkey"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a minimal Memcached serverless cache", func() {
		spec.Engine = "memcached"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Redis with major engine version", func() {
		spec.MajorEngineVersion = "7"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts data storage limits", func() {
		spec.DataStorageMinGb = 1
		spec.DataStorageMaxGb = 100
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts only data_storage_max_gb without min", func() {
		spec.DataStorageMaxGb = 500
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts only data_storage_min_gb without max", func() {
		spec.DataStorageMinGb = 10
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts ECPU limits", func() {
		spec.EcpuMin = 1000
		spec.EcpuMax = 15000000
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts only ecpu_max without min", func() {
		spec.EcpuMax = 5000
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts VPC networking configuration", func() {
		spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
			strRef("subnet-aaa"), strRef("subnet-bbb"),
		}
		spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
			strRef("sg-123"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts KMS encryption key", func() {
		spec.KmsKeyId = strRef("arn:aws:kms:us-east-1:123456789012:key/12345678-abcd-efgh-1234-abcdefgh")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Redis daily snapshot time", func() {
		spec.DailySnapshotTime = "05:00"
		spec.SnapshotRetentionLimit = 7
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Redis user group", func() {
		spec.UserGroupId = "my-user-group"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a production-ready Redis configuration", func() {
		spec.Engine = "redis"
		spec.MajorEngineVersion = "7"
		spec.Description = "Production session cache"
		spec.DataStorageMinGb = 5
		spec.DataStorageMaxGb = 100
		spec.EcpuMin = 1000
		spec.EcpuMax = 100000
		spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
			strRef("subnet-aaa"), strRef("subnet-bbb"), strRef("subnet-ccc"),
		}
		spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
			strRef("sg-123"),
		}
		spec.KmsKeyId = strRef("arn:aws:kms:us-east-1:123456789012:key/my-key")
		spec.DailySnapshotTime = "03:00"
		spec.SnapshotRetentionLimit = 14
		spec.UserGroupId = "app-users"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts maximum data storage", func() {
		spec.DataStorageMaxGb = 5000
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts maximum ECPU", func() {
		spec.EcpuMax = 15000000
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts description", func() {
		spec.Description = "My serverless cache for API responses"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Required field validations
	// -------------------------------------------------------------------------

	ginkgo.It("fails when engine is missing", func() {
		spec.Engine = ""
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: engine valid values
	// -------------------------------------------------------------------------

	ginkgo.It("fails when engine is invalid", func() {
		spec.Engine = "dynamodb"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when engine is empty string after required check", func() {
		spec.Engine = "Redis"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: data_storage min/max ordering
	// -------------------------------------------------------------------------

	ginkgo.It("fails when data_storage_min_gb exceeds data_storage_max_gb", func() {
		spec.DataStorageMinGb = 100
		spec.DataStorageMaxGb = 10
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts data_storage_min_gb equal to data_storage_max_gb", func() {
		spec.DataStorageMinGb = 50
		spec.DataStorageMaxGb = 50
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: ecpu min/max ordering
	// -------------------------------------------------------------------------

	ginkgo.It("fails when ecpu_min exceeds ecpu_max", func() {
		spec.EcpuMin = 10000
		spec.EcpuMax = 1000
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts ecpu_min equal to ecpu_max", func() {
		spec.EcpuMin = 5000
		spec.EcpuMax = 5000
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: engine-specific field guards (Memcached)
	// -------------------------------------------------------------------------

	ginkgo.It("fails when daily_snapshot_time is set for memcached", func() {
		spec.Engine = "memcached"
		spec.DailySnapshotTime = "05:00"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when snapshot_retention_limit is set for memcached", func() {
		spec.Engine = "memcached"
		spec.SnapshotRetentionLimit = 7
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when user_group_id is set for memcached", func() {
		spec.Engine = "memcached"
		spec.UserGroupId = "my-group"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: engine-specific guards allow Valkey
	// -------------------------------------------------------------------------

	ginkgo.It("accepts daily_snapshot_time for valkey", func() {
		spec.Engine = "valkey"
		spec.DailySnapshotTime = "09:00"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts user_group_id for valkey", func() {
		spec.Engine = "valkey"
		spec.UserGroupId = "valkey-users"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Field-level: data_storage range
	// -------------------------------------------------------------------------

	ginkgo.It("fails when data_storage_max_gb exceeds 5000", func() {
		spec.DataStorageMaxGb = 5001
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when data_storage_min_gb exceeds 5000", func() {
		spec.DataStorageMinGb = 5001
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Field-level: ECPU range
	// -------------------------------------------------------------------------

	ginkgo.It("fails when ecpu_max exceeds 15000000", func() {
		spec.EcpuMax = 15000001
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when ecpu_min exceeds 15000000", func() {
		spec.EcpuMin = 15000001
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: ECPU/storage floor validation
	// -------------------------------------------------------------------------

	ginkgo.It("fails when ecpu_min is below 1000 but non-zero", func() {
		spec.EcpuMin = 500
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when ecpu_max is below 1000 but non-zero", func() {
		spec.EcpuMax = 999
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Field-level: snapshot_retention_limit range
	// -------------------------------------------------------------------------

	ginkgo.It("fails when snapshot_retention_limit exceeds 35", func() {
		spec.SnapshotRetentionLimit = 36
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
