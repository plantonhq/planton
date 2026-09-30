package awsneptuneclusterv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsNeptuneClusterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsNeptuneClusterSpec Custom Validation Tests")
}

// twoSubnets returns the minimum valid subnet list (two literal subnet IDs).
func twoSubnets() []*foreignkeyv1.StringValueOrRef {
	return []*foreignkeyv1.StringValueOrRef{
		{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
		{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
	}
}

// validCluster returns a minimal valid Neptune cluster that individual
// tests mutate into specific scenarios.
func validCluster() *AwsNeptuneCluster {
	return &AwsNeptuneCluster{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsNeptuneCluster",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-neptune",
		},
		Spec: &AwsNeptuneClusterSpec{
			Region:            "us-west-2",
			SubnetIds:         twoSubnets(),
			SkipFinalSnapshot: true,
			Instances: []*AwsNeptuneClusterInstance{
				{Name: "writer", InstanceClass: "db.r6g.large"},
			},
		},
	}
}

var _ = ginkgo.Describe("AwsNeptuneClusterSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("accepts a minimal provisioned cluster", func() {
			err := protovalidate.Validate(validCluster())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a cluster with writer and reader instances", func() {
			input := validCluster()
			input.Spec.Instances = []*AwsNeptuneClusterInstance{
				{Name: "writer", InstanceClass: "db.r6g.large", PromotionTier: 0},
				{Name: "reader-1", InstanceClass: "db.r6g.large", PromotionTier: 1},
			}
			input.Spec.EnabledCloudwatchLogsExports = []string{"audit", "slowquery"}
			input.Spec.BackupRetentionPeriod = 7
			input.Spec.IamDatabaseAuthenticationEnabled = true
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a Neptune Serverless cluster", func() {
			input := validCluster()
			input.Spec.ServerlessV2Scaling = &AwsNeptuneClusterServerlessV2Scaling{
				MinCapacity: 1,
				MaxCapacity: 32,
			}
			input.Spec.Instances = []*AwsNeptuneClusterInstance{
				{Name: "writer", InstanceClass: "db.serverless"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts an existing subnet group instead of subnet ids", func() {
			input := validCluster()
			input.Spec.SubnetIds = nil
			input.Spec.NeptuneSubnetGroupName = "existing-group"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a headless snapshot restore without instances", func() {
			input := validCluster()
			input.Spec.Instances = nil
			input.Spec.SnapshotIdentifier = "my-snapshot"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a headless replica without instances", func() {
			input := validCluster()
			input.Spec.Instances = nil
			input.Spec.ReplicationSourceIdentifier = "arn:aws:rds:us-west-2:123456789012:cluster:source"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts inline cluster parameters with a pinned engine version", func() {
			input := validCluster()
			input.Spec.EngineVersion = "1.4.5.1"
			input.Spec.Parameters = []*AwsNeptuneClusterParameter{
				{Name: "neptune_enable_audit_log", Value: "1", ApplyMethod: "pending-reboot"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a major version upgrade with the instance parameter group", func() {
			input := validCluster()
			input.Spec.AllowMajorVersionUpgrade = true
			input.Spec.NeptuneInstanceParameterGroupName = "upgrade-instance-params"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts iopt1 storage", func() {
			input := validCluster()
			input.Spec.StorageType = "iopt1"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("networking validations", func() {

		ginkgo.It("rejects a spec with neither subnets nor a subnet group", func() {
			input := validCluster()
			input.Spec.SubnetIds = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("provide at least two subnet_ids (distinct AZs) or an"))
		})

		ginkgo.It("rejects a single subnet without a subnet group", func() {
			input := validCluster()
			input.Spec.SubnetIds = input.Spec.SubnetIds[:1]
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("provide at least two subnet_ids (distinct AZs) or an"))
		})

		ginkgo.It("rejects more than three availability zones", func() {
			input := validCluster()
			input.Spec.AvailabilityZones = []string{"us-west-2a", "us-west-2b", "us-west-2c", "us-west-2d"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an out-of-range port", func() {
			input := validCluster()
			input.Spec.Port = 70000
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid storage type", func() {
			input := validCluster()
			input.Spec.StorageType = "gp3"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("cluster shape validations", func() {

		ginkgo.It("rejects an empty instances list on a regular cluster", func() {
			input := validCluster()
			input.Spec.Instances = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("instances is required"))
		})

		ginkgo.It("rejects a provisioned instance class under serverless scaling", func() {
			input := validCluster()
			input.Spec.ServerlessV2Scaling = &AwsNeptuneClusterServerlessV2Scaling{
				MinCapacity: 1,
				MaxCapacity: 32,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("every instance must use instance_class 'db.serverless'"))
		})

		ginkgo.It("rejects a db.serverless instance without scaling bounds", func() {
			input := validCluster()
			input.Spec.Instances = []*AwsNeptuneClusterInstance{
				{Name: "writer", InstanceClass: "db.serverless"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("serverless_v2_scaling is required when any instance"))
		})

		ginkgo.It("rejects an instance name that is not a valid identifier fragment", func() {
			input := validCluster()
			input.Spec.Instances[0].Name = "Writer_1"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an instance class without the db. prefix", func() {
			input := validCluster()
			input.Spec.Instances[0].InstanceClass = "r6g.large"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a promotion tier above 15", func() {
			input := validCluster()
			input.Spec.Instances[0].PromotionTier = 16
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("serverless scaling validations", func() {

		ginkgo.It("rejects max capacity below min capacity", func() {
			input := validCluster()
			input.Spec.Instances = []*AwsNeptuneClusterInstance{
				{Name: "writer", InstanceClass: "db.serverless"},
			}
			input.Spec.ServerlessV2Scaling = &AwsNeptuneClusterServerlessV2Scaling{
				MinCapacity: 64,
				MaxCapacity: 32,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("max_capacity must be greater than or equal to min_capacity"))
		})

		ginkgo.It("rejects a min capacity below the Neptune floor", func() {
			input := validCluster()
			input.Spec.Instances = []*AwsNeptuneClusterInstance{
				{Name: "writer", InstanceClass: "db.serverless"},
			}
			input.Spec.ServerlessV2Scaling = &AwsNeptuneClusterServerlessV2Scaling{
				MinCapacity: 0.5,
				MaxCapacity: 32,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a max capacity above the Neptune ceiling", func() {
			input := validCluster()
			input.Spec.Instances = []*AwsNeptuneClusterInstance{
				{Name: "writer", InstanceClass: "db.serverless"},
			}
			input.Spec.ServerlessV2Scaling = &AwsNeptuneClusterServerlessV2Scaling{
				MinCapacity: 1,
				MaxCapacity: 256,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("deletion safety validations", func() {

		ginkgo.It("rejects a missing final snapshot name while not skipping the snapshot", func() {
			input := validCluster()
			input.Spec.SkipFinalSnapshot = false
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("final_snapshot_identifier is required when"))
		})
	})

	ginkgo.Describe("observability validations", func() {

		ginkgo.It("rejects an unsupported log export type", func() {
			input := validCluster()
			input.Spec.EnabledCloudwatchLogsExports = []string{"general"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("must contain only 'audit' or 'slowquery'"))
		})
	})

	ginkgo.Describe("parameter group validations", func() {

		ginkgo.It("rejects inline parameters alongside an existing group name", func() {
			input := validCluster()
			input.Spec.EngineVersion = "1.4.5.1"
			input.Spec.NeptuneClusterParameterGroupName = "existing-group"
			input.Spec.Parameters = []*AwsNeptuneClusterParameter{
				{Name: "neptune_enable_audit_log", Value: "1"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("parameters and neptune_cluster_parameter_group_name are mutually exclusive"))
		})

		ginkgo.It("rejects inline parameters without a pinned engine version", func() {
			input := validCluster()
			input.Spec.Parameters = []*AwsNeptuneClusterParameter{
				{Name: "neptune_enable_audit_log", Value: "1"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("inline parameters require a pinned engine_version"))
		})

		ginkgo.It("rejects an invalid apply method", func() {
			input := validCluster()
			input.Spec.EngineVersion = "1.4.5.1"
			input.Spec.Parameters = []*AwsNeptuneClusterParameter{
				{Name: "neptune_enable_audit_log", Value: "1", ApplyMethod: "eventually"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("apply_method must be 'immediate' or 'pending-reboot'"))
		})
	})

	ginkgo.Describe("upgrade validations", func() {

		ginkgo.It("rejects a major version upgrade without the instance parameter group", func() {
			input := validCluster()
			input.Spec.AllowMajorVersionUpgrade = true
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("allow_major_version_upgrade requires neptune_instance_parameter_group_name"))
		})
	})

	ginkgo.Describe("custom endpoint validations", func() {

		ginkgo.It("accepts a reader custom endpoint with static members", func() {
			input := validCluster()
			input.Spec.CustomEndpoints = []*AwsNeptuneClusterCustomEndpoint{{
				Name:          "analytics",
				EndpointType:  "READER",
				StaticMembers: []string{input.Spec.Instances[0].Name},
			}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("rejects an endpoint that pins and excludes members at once", func() {
			input := validCluster()
			input.Spec.CustomEndpoints = []*AwsNeptuneClusterCustomEndpoint{{
				Name:            "analytics",
				EndpointType:    "READER",
				StaticMembers:   []string{input.Spec.Instances[0].Name},
				ExcludedMembers: []string{input.Spec.Instances[0].Name},
			}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("mutually exclusive"))
		})

		ginkgo.It("rejects duplicate endpoint names", func() {
			input := validCluster()
			input.Spec.CustomEndpoints = []*AwsNeptuneClusterCustomEndpoint{
				{Name: "analytics", EndpointType: "READER"},
				{Name: "analytics", EndpointType: "ANY"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("must be unique"))
		})

		ginkgo.It("rejects a member that names no instance entry", func() {
			input := validCluster()
			input.Spec.CustomEndpoints = []*AwsNeptuneClusterCustomEndpoint{{
				Name:          "analytics",
				EndpointType:  "READER",
				StaticMembers: []string{"ghost"},
			}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("must name entries of spec.instances"))
		})
	})

	ginkgo.Describe("instance parameter validations", func() {

		ginkgo.It("accepts inline instance parameters with a pinned engine version", func() {
			input := validCluster()
			input.Spec.EngineVersion = "1.4.5.1"
			input.Spec.InstanceParameters = []*AwsNeptuneClusterParameter{
				{Name: "neptune_query_timeout", Value: "120000"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("rejects inline instance parameters alongside an existing instance group name", func() {
			input := validCluster()
			input.Spec.EngineVersion = "1.4.5.1"
			input.Spec.NeptuneInstanceParameterGroupName = "existing-group"
			input.Spec.InstanceParameters = []*AwsNeptuneClusterParameter{
				{Name: "neptune_query_timeout", Value: "120000"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("instance_parameters and neptune_instance_parameter_group_name are mutually exclusive"))
		})

		ginkgo.It("rejects inline instance parameters without a pinned engine version", func() {
			input := validCluster()
			input.Spec.InstanceParameters = []*AwsNeptuneClusterParameter{
				{Name: "neptune_query_timeout", Value: "120000"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("inline instance_parameters require a pinned engine_version"))
		})
	})

	ginkgo.Describe("identifier format validations", func() {

		ginkgo.It("rejects an instance name with a trailing hyphen -- the provider's identifier rule", func() {
			input := validCluster()
			input.Spec.Instances[0].Name = "reader-"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a final snapshot identifier with a double hyphen", func() {
			input := validCluster()
			input.Spec.SkipFinalSnapshot = false
			input.Spec.FinalSnapshotIdentifier = "final--snap"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a global cluster identifier that starts with a digit", func() {
			input := validCluster()
			input.Spec.GlobalClusterIdentifier = "1global"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
