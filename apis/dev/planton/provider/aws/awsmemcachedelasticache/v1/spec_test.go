package awsmemcachedelasticachev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAwsMemcachedElasticacheSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsMemcachedElasticacheSpec Validation Suite")
}

// helper to create a StringValueOrRef with a literal value.
func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

var _ = ginkgo.Describe("AwsMemcachedElasticacheSpec validations", func() {
	var spec *AwsMemcachedElasticacheSpec

	ginkgo.BeforeEach(func() {
		// Minimal valid spec: single-node Memcached cluster.
		spec = &AwsMemcachedElasticacheSpec{
			Region:        "us-west-2",
			EngineVersion: "1.6.22",
			NodeType:      "cache.t3.micro",
			NumCacheNodes: 1,
		}
	})

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal single-node Memcached cluster", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a multi-node cluster", func() {
		spec.NumCacheNodes = 3
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts cross-az mode with multiple nodes", func() {
		spec.NumCacheNodes = 3
		spec.AzMode = "cross-az"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts single-az mode explicitly", func() {
		spec.AzMode = "single-az"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts transit encryption enabled", func() {
		spec.TransitEncryptionEnabled = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts custom port", func() {
		port := int32(11212)
		spec.Port = &port
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

	ginkgo.It("accepts custom parameters with family", func() {
		spec.ParameterGroupFamily = "memcached1.6"
		spec.Parameters = []*AwsMemcachedElasticacheParameter{
			{Name: "chunk_size", Value: "96"},
			{Name: "binding_protocol", Value: "auto"},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts maintenance window", func() {
		spec.MaintenanceWindow = "sun:05:00-sun:06:00"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts notification topic ARN", func() {
		spec.NotificationTopicArn = strRef("arn:aws:sns:us-east-1:123456789012:my-topic")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts preferred availability zones matching node count", func() {
		spec.NumCacheNodes = 3
		spec.PreferredAvailabilityZones = []string{"us-east-1a", "us-east-1b", "us-east-1c"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a production-ready configuration", func() {
		port := int32(11211)
		spec.Region = "us-west-2"
		spec.EngineVersion = "1.6.22"
		spec.NodeType = "cache.r7g.large"
		spec.NumCacheNodes = 3
		spec.AzMode = "cross-az"
		spec.Port = &port
		spec.TransitEncryptionEnabled = true
		spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
			strRef("subnet-aaa"), strRef("subnet-bbb"), strRef("subnet-ccc"),
		}
		spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
			strRef("sg-123"),
		}
		spec.MaintenanceWindow = "sun:05:00-sun:06:00"
		spec.AutoMinorVersionUpgrade = true
		spec.NotificationTopicArn = strRef("arn:aws:sns:us-east-1:123456789012:my-topic")
		spec.ParameterGroupFamily = "memcached1.6"
		spec.Parameters = []*AwsMemcachedElasticacheParameter{
			{Name: "chunk_size", Value: "96"},
		}
		spec.PreferredAvailabilityZones = []string{"us-east-1a", "us-east-1b", "us-east-1c"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts maximum node count", func() {
		spec.NumCacheNodes = 40
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts apply_immediately flag", func() {
		spec.ApplyImmediately = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Required field validations
	// -------------------------------------------------------------------------

	ginkgo.It("fails when engine_version is missing", func() {
		spec.EngineVersion = ""
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when node_type is missing", func() {
		spec.NodeType = ""
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: az_mode valid values
	// -------------------------------------------------------------------------

	ginkgo.It("fails when az_mode has an invalid value", func() {
		spec.AzMode = "multi-az"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: cross-az requires multi-node
	// -------------------------------------------------------------------------

	ginkgo.It("fails when cross-az is set with single node", func() {
		spec.NumCacheNodes = 1
		spec.AzMode = "cross-az"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: preferred_availability_zones length mismatch
	// -------------------------------------------------------------------------

	ginkgo.It("fails when preferred_availability_zones length does not match num_cache_nodes", func() {
		spec.NumCacheNodes = 3
		spec.PreferredAvailabilityZones = []string{"us-east-1a", "us-east-1b"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when preferred_availability_zones provided for single node but length is 2", func() {
		spec.NumCacheNodes = 1
		spec.PreferredAvailabilityZones = []string{"us-east-1a", "us-east-1b"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Field-level: num_cache_nodes range
	// -------------------------------------------------------------------------

	ginkgo.It("fails when num_cache_nodes is 0", func() {
		spec.NumCacheNodes = 0
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when num_cache_nodes exceeds 40", func() {
		spec.NumCacheNodes = 41
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Field-level: port range
	// -------------------------------------------------------------------------

	ginkgo.It("fails when port is 0", func() {
		port := int32(0)
		spec.Port = &port
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when port exceeds 65535", func() {
		port := int32(65536)
		spec.Port = &port
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Field-level: maintenance_window pattern
	// -------------------------------------------------------------------------

	ginkgo.It("fails when maintenance_window has an invalid format", func() {
		spec.MaintenanceWindow = "sunday:05:00-sunday:06:00"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when maintenance_window uses uppercase", func() {
		spec.MaintenanceWindow = "Sun:05:00-Sun:06:00"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: parameters require parameter_group_family
	// -------------------------------------------------------------------------

	ginkgo.It("fails when parameters are set without parameter_group_family", func() {
		spec.Parameters = []*AwsMemcachedElasticacheParameter{
			{Name: "chunk_size", Value: "96"},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Nested message: parameter validations
	// -------------------------------------------------------------------------

	ginkgo.It("fails when parameter name is missing", func() {
		spec.ParameterGroupFamily = "memcached1.6"
		spec.Parameters = []*AwsMemcachedElasticacheParameter{
			{Name: "", Value: "96"},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when parameter value is missing", func() {
		spec.ParameterGroupFamily = "memcached1.6"
		spec.Parameters = []*AwsMemcachedElasticacheParameter{
			{Name: "chunk_size", Value: ""},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
