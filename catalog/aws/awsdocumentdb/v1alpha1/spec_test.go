package awsdocumentdbv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsDocumentDbSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsDocumentDbSpec Custom Validation Tests")
}

// twoSubnets returns the minimum valid subnet list (two literal subnet IDs).
func twoSubnets() []*foreignkeyv1.StringValueOrRef {
	return []*foreignkeyv1.StringValueOrRef{
		{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
		{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
	}
}

// validCluster returns a minimal valid DocumentDB cluster that individual
// tests mutate into specific scenarios.
func validCluster() *AwsDocumentDb {
	return &AwsDocumentDb{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsDocumentDb",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-docdb",
		},
		Spec: &AwsDocumentDbSpec{
			Region:                   "us-west-2",
			SubnetIds:                twoSubnets(),
			MasterUsername:           "docadmin",
			ManageMasterUserPassword: true,
			SkipFinalSnapshot:        true,
			Instances: []*AwsDocumentDbInstance{
				{Name: "writer", InstanceClass: "db.t4g.medium"},
			},
		},
	}
}

var _ = ginkgo.Describe("AwsDocumentDbSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("accepts a minimal provisioned cluster", func() {
			err := protovalidate.Validate(validCluster())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a cluster with writer and reader instances", func() {
			input := validCluster()
			input.Spec.Instances = []*AwsDocumentDbInstance{
				{Name: "writer", InstanceClass: "db.r6g.large", PromotionTier: 0},
				{Name: "reader-1", InstanceClass: "db.r6g.large", PromotionTier: 1},
			}
			input.Spec.EnabledCloudwatchLogsExports = []string{"audit", "profiler"}
			input.Spec.BackupRetentionPeriod = 7
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a DocumentDB Serverless cluster", func() {
			input := validCluster()
			input.Spec.ServerlessV2Scaling = &AwsDocumentDbServerlessV2Scaling{
				MinCapacity: 0.5,
				MaxCapacity: 4,
			}
			input.Spec.Instances = []*AwsDocumentDbInstance{
				{Name: "writer", InstanceClass: "db.serverless"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a direct master password without managed password", func() {
			input := validCluster()
			input.Spec.ManageMasterUserPassword = false
			input.Spec.MasterPassword = "super-secret-password"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts an existing subnet group instead of subnet ids", func() {
			input := validCluster()
			input.Spec.SubnetIds = nil
			input.Spec.DbSubnetGroupName = "existing-group"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a headless snapshot restore without username or instances", func() {
			input := validCluster()
			input.Spec.MasterUsername = ""
			input.Spec.Instances = nil
			input.Spec.SnapshotIdentifier = "my-snapshot"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a point-in-time restore with latest restorable time", func() {
			input := validCluster()
			input.Spec.RestoreToPointInTime = &AwsDocumentDbRestoreToPointInTime{
				SourceClusterIdentifier: "source-cluster",
				UseLatestRestorableTime: true,
				RestoreType:             "copy-on-write",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a global cluster secondary without credentials", func() {
			input := validCluster()
			input.Spec.MasterUsername = ""
			input.Spec.ManageMasterUserPassword = false
			input.Spec.Instances = nil
			input.Spec.GlobalClusterIdentifier = "global-docdb"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts inline cluster parameters with a pinned engine version", func() {
			input := validCluster()
			input.Spec.EngineVersion = "5.0.0"
			input.Spec.Parameters = []*AwsDocumentDbParameter{
				{Name: "audit_logs", Value: "enabled", ApplyMethod: "immediate"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a custom port in the DocumentDB range", func() {
			input := validCluster()
			input.Spec.Port = 27018
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

		ginkgo.It("rejects an invalid network_type", func() {
			input := validCluster()
			input.Spec.NetworkType = "IPV6"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("network_type must be 'IPV4' or 'DUAL' when set"))
		})

		ginkgo.It("rejects a port below the DocumentDB floor", func() {
			input := validCluster()
			input.Spec.Port = 1024
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("port must be between 1150 and 65535"))
		})
	})

	ginkgo.Describe("password strategy validations", func() {

		ginkgo.It("rejects a master password alongside the managed password", func() {
			input := validCluster()
			input.Spec.MasterPassword = "also-a-password"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("master_password cannot be set when"))
		})

		ginkgo.It("rejects a new cluster without a master username", func() {
			input := validCluster()
			input.Spec.MasterUsername = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("master_username is required"))
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
			input.Spec.ServerlessV2Scaling = &AwsDocumentDbServerlessV2Scaling{
				MinCapacity: 0.5,
				MaxCapacity: 4,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("every instance must use instance_class 'db.serverless'"))
		})

		ginkgo.It("rejects a db.serverless instance without scaling bounds", func() {
			input := validCluster()
			input.Spec.Instances = []*AwsDocumentDbInstance{
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
			input.Spec.Instances = []*AwsDocumentDbInstance{
				{Name: "writer", InstanceClass: "db.serverless"},
			}
			input.Spec.ServerlessV2Scaling = &AwsDocumentDbServerlessV2Scaling{
				MinCapacity: 8,
				MaxCapacity: 4,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("max_capacity must be greater than or equal to min_capacity"))
		})

		ginkgo.It("rejects capacities that are not half-step multiples", func() {
			input := validCluster()
			input.Spec.Instances = []*AwsDocumentDbInstance{
				{Name: "writer", InstanceClass: "db.serverless"},
			}
			input.Spec.ServerlessV2Scaling = &AwsDocumentDbServerlessV2Scaling{
				MinCapacity: 0.7,
				MaxCapacity: 4,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("must be multiples of 0.5"))
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

		ginkgo.It("rejects a final snapshot identifier with consecutive hyphens", func() {
			input := validCluster()
			input.Spec.SkipFinalSnapshot = false
			input.Spec.FinalSnapshotIdentifier = "my--final"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts a well-formed final snapshot identifier", func() {
			input := validCluster()
			input.Spec.SkipFinalSnapshot = false
			input.Spec.FinalSnapshotIdentifier = "docdb-final-2026"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("instance maintenance validations", func() {

		ginkgo.It("accepts an explicit certificate-rotation restart opt-out with immediate apply", func() {
			input := validCluster()
			noRestart := false
			input.Spec.Instances[0].CertificateRotationRestart = &noRestart
			input.Spec.Instances[0].ApplyImmediately = true
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("restore validations", func() {

		ginkgo.It("rejects both restore sources together", func() {
			input := validCluster()
			input.Spec.SnapshotIdentifier = "my-snapshot"
			input.Spec.RestoreToPointInTime = &AwsDocumentDbRestoreToPointInTime{
				SourceClusterIdentifier: "source-cluster",
				UseLatestRestorableTime: true,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("mutually exclusive create-time restore sources"))
		})

		ginkgo.It("rejects a point-in-time restore with both time strategies", func() {
			input := validCluster()
			input.Spec.RestoreToPointInTime = &AwsDocumentDbRestoreToPointInTime{
				SourceClusterIdentifier: "source-cluster",
				RestoreToTime:           "2026-07-01T09:45:00Z",
				UseLatestRestorableTime: true,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("exactly one of restore_to_time or use_latest_restorable_time"))
		})

		ginkgo.It("rejects an invalid restore type", func() {
			input := validCluster()
			input.Spec.RestoreToPointInTime = &AwsDocumentDbRestoreToPointInTime{
				SourceClusterIdentifier: "source-cluster",
				UseLatestRestorableTime: true,
				RestoreType:             "shallow-copy",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("restore_type must be 'full-copy' or 'copy-on-write'"))
		})
	})

	ginkgo.Describe("observability validations", func() {

		ginkgo.It("rejects an unsupported log export type", func() {
			input := validCluster()
			input.Spec.EnabledCloudwatchLogsExports = []string{"general"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("must contain only 'audit' or 'profiler'"))
		})
	})

	ginkgo.Describe("parameter group validations", func() {

		ginkgo.It("rejects inline parameters alongside an existing group name", func() {
			input := validCluster()
			input.Spec.EngineVersion = "5.0.0"
			input.Spec.DbClusterParameterGroupName = "existing-group"
			input.Spec.Parameters = []*AwsDocumentDbParameter{
				{Name: "audit_logs", Value: "enabled"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("parameters and db_cluster_parameter_group_name are mutually exclusive"))
		})

		ginkgo.It("rejects inline parameters without a pinned engine version", func() {
			input := validCluster()
			input.Spec.Parameters = []*AwsDocumentDbParameter{
				{Name: "audit_logs", Value: "enabled"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("inline parameters require a pinned engine_version"))
		})

		ginkgo.It("rejects an invalid apply method", func() {
			input := validCluster()
			input.Spec.EngineVersion = "5.0.0"
			input.Spec.Parameters = []*AwsDocumentDbParameter{
				{Name: "audit_logs", Value: "enabled", ApplyMethod: "eventually"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("apply_method must be 'immediate' or 'pending-reboot'"))
		})
	})
})
