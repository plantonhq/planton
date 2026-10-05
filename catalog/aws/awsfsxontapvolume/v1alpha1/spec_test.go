package awsfsxontapvolumev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	shared "github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsFsxOntapVolumeSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsFsxOntapVolumeSpec Validation Suite")
}

func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func int32Ptr(i int32) *int32 {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}

var _ = ginkgo.Describe("AwsFsxOntapVolumeSpec validations", func() {
	var spec *AwsFsxOntapVolumeSpec

	ginkgo.BeforeEach(func() {
		spec = &AwsFsxOntapVolumeSpec{
			Region:                  "us-west-2",
			StorageVirtualMachineId: strRef("svm-0123456789abcdef0"),
			Name:                    "vol_default",
			SizeInMegabytes:         int32Ptr(1024),
		}
	})

	// -------------------------------------------------------------------------
	// Happy path — valid configurations
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal valid spec", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a full production configuration", func() {
		spec.JunctionPath = "/data/prod"
		spec.OntapVolumeType = stringPtr("RW")
		spec.VolumeStyle = stringPtr("FLEXVOL")
		spec.SecurityStyle = "UNIX"
		spec.SnapshotPolicy = "default"
		spec.StorageEfficiencyEnabled = boolPtr(true)
		spec.CopyTagsToBackups = boolPtr(true)
		spec.FinalBackupTags = map[string]string{"retention": "final"}
		spec.TieringPolicy = &AwsFsxOntapVolumeTieringPolicy{
			Name:          "AUTO",
			CoolingPeriod: 31,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts the byte-precise size arm", func() {
		spec.SizeInMegabytes = nil
		spec.SizeInBytes = int64Ptr(4398046511104) // 4 TiB
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a size beyond 2 PiB via size_in_bytes", func() {
		spec.SizeInMegabytes = nil
		spec.SizeInBytes = int64Ptr(3000000000000000) // ~2.7 PiB
		spec.VolumeStyle = stringPtr("FLEXGROUP")
		spec.AggregateConfiguration = &AwsFsxOntapVolumeAggregateConfiguration{
			Aggregates: []string{"aggr1", "aggr2"},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts explicit storage_efficiency_enabled false", func() {
		spec.StorageEfficiencyEnabled = boolPtr(false)
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts DP volume type", func() {
		spec.OntapVolumeType = stringPtr("DP")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts FLEXGROUP volume style", func() {
		spec.VolumeStyle = stringPtr("FLEXGROUP")
		spec.AggregateConfiguration = &AwsFsxOntapVolumeAggregateConfiguration{
			Aggregates:               []string{"aggr1", "aggr2"},
			ConstituentsPerAggregate: 8,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts NTFS security style", func() {
		spec.SecurityStyle = "NTFS"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts MIXED security style", func() {
		spec.SecurityStyle = "MIXED"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts junction path with nested directories", func() {
		spec.JunctionPath = "/shares/finance/reports"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts name at 203 characters", func() {
		longName := ""
		for i := 0; i < 203; i++ {
			longName += "a"
		}
		spec.Name = longName
		gomega.Expect(len(spec.Name)).To(gomega.Equal(203))
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts name with underscores and digits", func() {
		spec.Name = "vol_prod_01_data"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts skip_final_backup true", func() {
		spec.SkipFinalBackup = boolPtr(true)
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Tiering policy — valid configurations
	// -------------------------------------------------------------------------

	ginkgo.It("accepts NONE tiering policy", func() {
		spec.TieringPolicy = &AwsFsxOntapVolumeTieringPolicy{Name: "NONE"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts SNAPSHOT_ONLY with cooling period", func() {
		spec.TieringPolicy = &AwsFsxOntapVolumeTieringPolicy{
			Name:          "SNAPSHOT_ONLY",
			CoolingPeriod: 2,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts AUTO with cooling period 183", func() {
		spec.TieringPolicy = &AwsFsxOntapVolumeTieringPolicy{
			Name:          "AUTO",
			CoolingPeriod: 183,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts ALL tiering policy without cooling period", func() {
		spec.TieringPolicy = &AwsFsxOntapVolumeTieringPolicy{Name: "ALL"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// SnapLock — valid configurations
	// -------------------------------------------------------------------------

	ginkgo.It("accepts SnapLock ENTERPRISE with defaults", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType: "ENTERPRISE",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts SnapLock COMPLIANCE with full retention periods", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType: "COMPLIANCE",
			RetentionPeriod: &AwsFsxOntapVolumeRetentionPeriod{
				DefaultRetention: &AwsFsxOntapVolumeRetentionDuration{
					Type:  "YEARS",
					Value: 5,
				},
				MinimumRetention: &AwsFsxOntapVolumeRetentionDuration{
					Type:  "YEARS",
					Value: 1,
				},
				MaximumRetention: &AwsFsxOntapVolumeRetentionDuration{
					Type:  "YEARS",
					Value: 10,
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts SnapLock ENTERPRISE with privileged delete enabled", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType:     "ENTERPRISE",
			PrivilegedDelete: stringPtr("ENABLED"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts SnapLock with autocommit period", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType: "ENTERPRISE",
			AutocommitPeriod: &AwsFsxOntapVolumeAutocommitPeriod{
				Type:  "HOURS",
				Value: 24,
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts SnapLock with INFINITE retention", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType: "COMPLIANCE",
			RetentionPeriod: &AwsFsxOntapVolumeRetentionPeriod{
				DefaultRetention: &AwsFsxOntapVolumeRetentionDuration{
					Type: "INFINITE",
				},
				MinimumRetention: &AwsFsxOntapVolumeRetentionDuration{
					Type:  "DAYS",
					Value: 1,
				},
				MaximumRetention: &AwsFsxOntapVolumeRetentionDuration{
					Type: "INFINITE",
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts SnapLock with volume append mode enabled", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType:            "ENTERPRISE",
			VolumeAppendModeEnabled: boolPtr(true),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts SnapLock audit log volume", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType:   "ENTERPRISE",
			AuditLogVolume: boolPtr(true),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Field-level failures — required fields
	// -------------------------------------------------------------------------

	ginkgo.It("rejects missing storage_virtual_machine_id", func() {
		spec.StorageVirtualMachineId = nil
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects empty name", func() {
		spec.Name = ""
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects name exceeding 203 characters", func() {
		longName := ""
		for i := 0; i < 204; i++ {
			longName += "a"
		}
		spec.Name = longName
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects size_in_megabytes below minimum (20)", func() {
		spec.SizeInMegabytes = int32Ptr(19)
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL validation — size_exactly_one_arm
	// -------------------------------------------------------------------------

	ginkgo.It("rejects a spec with no size arm", func() {
		spec.SizeInMegabytes = nil
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a spec with both size arms", func() {
		spec.SizeInBytes = int64Ptr(1073741824)
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects size_in_bytes below the 20 MiB minimum", func() {
		spec.SizeInMegabytes = nil
		spec.SizeInBytes = int64Ptr(1048576)
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects size_in_bytes above the ~20 PiB ceiling", func() {
		spec.SizeInMegabytes = nil
		spec.SizeInBytes = int64Ptr(22517998000000001)
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL validation failures — name format
	// -------------------------------------------------------------------------

	ginkgo.It("rejects name with hyphens", func() {
		spec.Name = "vol-prod-01"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects name with spaces", func() {
		spec.Name = "vol prod 01"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects name with special characters", func() {
		spec.Name = "vol@prod#01"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL validation failures — enum values
	// -------------------------------------------------------------------------

	ginkgo.It("rejects invalid ontap_volume_type", func() {
		spec.OntapVolumeType = stringPtr("INVALID")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects invalid volume_style", func() {
		spec.VolumeStyle = stringPtr("INVALID")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects invalid security_style", func() {
		spec.SecurityStyle = "INVALID"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL validation failures — junction_path format
	// -------------------------------------------------------------------------

	ginkgo.It("rejects junction_path not starting with /", func() {
		spec.JunctionPath = "vol1"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL validation failures — tiering policy
	// -------------------------------------------------------------------------

	ginkgo.It("rejects invalid tiering policy name", func() {
		spec.TieringPolicy = &AwsFsxOntapVolumeTieringPolicy{Name: "INVALID"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects cooling_period with NONE tiering policy", func() {
		spec.TieringPolicy = &AwsFsxOntapVolumeTieringPolicy{
			Name:          "NONE",
			CoolingPeriod: 30,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects cooling_period with ALL tiering policy", func() {
		spec.TieringPolicy = &AwsFsxOntapVolumeTieringPolicy{
			Name:          "ALL",
			CoolingPeriod: 30,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects cooling_period below minimum (2)", func() {
		spec.TieringPolicy = &AwsFsxOntapVolumeTieringPolicy{
			Name:          "AUTO",
			CoolingPeriod: 1,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects cooling_period above maximum (183)", func() {
		spec.TieringPolicy = &AwsFsxOntapVolumeTieringPolicy{
			Name:          "AUTO",
			CoolingPeriod: 184,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a tiering policy without a name (name is required when the block is present)", func() {
		spec.TieringPolicy = &AwsFsxOntapVolumeTieringPolicy{}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL validation failures — SnapLock
	// -------------------------------------------------------------------------

	ginkgo.It("rejects SnapLock with missing snaplock_type", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType: "",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects invalid snaplock_type", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType: "INVALID",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects invalid privileged_delete value", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType:     "ENTERPRISE",
			PrivilegedDelete: stringPtr("INVALID"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects invalid autocommit period type", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType: "ENTERPRISE",
			AutocommitPeriod: &AwsFsxOntapVolumeAutocommitPeriod{
				Type:  "INVALID",
				Value: 10,
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects invalid retention duration type", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType: "COMPLIANCE",
			RetentionPeriod: &AwsFsxOntapVolumeRetentionPeriod{
				DefaultRetention: &AwsFsxOntapVolumeRetentionDuration{
					Type:  "INVALID",
					Value: 5,
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an autocommit period without a type (the AWS API requires it)", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType:     "ENTERPRISE",
			AutocommitPeriod: &AwsFsxOntapVolumeAutocommitPeriod{Value: 30},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an autocommit value below its unit's AWS floor (MINUTES < 5)", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType:     "ENTERPRISE",
			AutocommitPeriod: &AwsFsxOntapVolumeAutocommitPeriod{Type: "MINUTES", Value: 4},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an autocommit value above its unit's AWS ceiling (DAYS > 3650)", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType:     "ENTERPRISE",
			AutocommitPeriod: &AwsFsxOntapVolumeAutocommitPeriod{Type: "DAYS", Value: 3651},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an autocommit value above its unit's AWS ceiling (YEARS > 10)", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType:     "ENTERPRISE",
			AutocommitPeriod: &AwsFsxOntapVolumeAutocommitPeriod{Type: "YEARS", Value: 11},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a value alongside autocommit type NONE (dead configuration)", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType:     "ENTERPRISE",
			AutocommitPeriod: &AwsFsxOntapVolumeAutocommitPeriod{Type: "NONE", Value: 1},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts autocommit MINUTES at the AWS floor (5)", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType:     "ENTERPRISE",
			AutocommitPeriod: &AwsFsxOntapVolumeAutocommitPeriod{Type: "MINUTES", Value: 5},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("rejects a retention duration without a type (the AWS API requires it)", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType: "COMPLIANCE",
			RetentionPeriod: &AwsFsxOntapVolumeRetentionPeriod{
				DefaultRetention: &AwsFsxOntapVolumeRetentionDuration{Value: 5},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a retention value above its unit's AWS ceiling (HOURS > 24)", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType: "COMPLIANCE",
			RetentionPeriod: &AwsFsxOntapVolumeRetentionPeriod{
				MinimumRetention: &AwsFsxOntapVolumeRetentionDuration{Type: "HOURS", Value: 25},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a retention value above its unit's AWS ceiling (YEARS > 100)", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType: "COMPLIANCE",
			RetentionPeriod: &AwsFsxOntapVolumeRetentionPeriod{
				MaximumRetention: &AwsFsxOntapVolumeRetentionDuration{Type: "YEARS", Value: 101},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a value alongside retention type INFINITE (AWS takes no value)", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType: "COMPLIANCE",
			RetentionPeriod: &AwsFsxOntapVolumeRetentionPeriod{
				MaximumRetention: &AwsFsxOntapVolumeRetentionDuration{Type: "INFINITE", Value: 1},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a zero-value retention duration for a unit type (0 days is meaningful)", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType: "COMPLIANCE",
			RetentionPeriod: &AwsFsxOntapVolumeRetentionPeriod{
				MinimumRetention: &AwsFsxOntapVolumeRetentionDuration{Type: "DAYS", Value: 0},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL validation failures — aggregate configuration
	// -------------------------------------------------------------------------

	ginkgo.It("rejects too many aggregates (>12)", func() {
		spec.VolumeStyle = stringPtr("FLEXGROUP")
		spec.AggregateConfiguration = &AwsFsxOntapVolumeAggregateConfiguration{
			Aggregates: []string{
				"aggr1", "aggr2", "aggr3", "aggr4", "aggr5", "aggr6",
				"aggr7", "aggr8", "aggr9", "aggr10", "aggr11", "aggr12", "aggr13",
			},
			ConstituentsPerAggregate: 8,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an aggregate name outside the aggrN pattern", func() {
		spec.VolumeStyle = stringPtr("FLEXGROUP")
		spec.AggregateConfiguration = &AwsFsxOntapVolumeAggregateConfiguration{
			Aggregates: []string{"aggregate1"},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects constituents_per_aggregate above 200", func() {
		spec.VolumeStyle = stringPtr("FLEXGROUP")
		spec.AggregateConfiguration = &AwsFsxOntapVolumeAggregateConfiguration{
			Aggregates:               []string{"aggr1"},
			ConstituentsPerAggregate: 201,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects any aggregate_configuration on a FLEXVOL volume (even an empty one)", func() {
		spec.VolumeStyle = stringPtr("FLEXVOL")
		spec.AggregateConfiguration = &AwsFsxOntapVolumeAggregateConfiguration{}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts constituents_per_aggregate without aggregates on FLEXGROUP (AWS picks the aggregates)", func() {
		spec.VolumeStyle = stringPtr("FLEXGROUP")
		spec.AggregateConfiguration = &AwsFsxOntapVolumeAggregateConfiguration{
			ConstituentsPerAggregate: 8,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Field bounds — junction path, snapshot policy, retention values
	// -------------------------------------------------------------------------

	ginkgo.It("rejects junction_path exceeding 255 characters", func() {
		longPath := "/"
		for i := 0; i < 255; i++ {
			longPath += "a"
		}
		spec.JunctionPath = longPath
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects retention duration value above 65535", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType: "COMPLIANCE",
			RetentionPeriod: &AwsFsxOntapVolumeRetentionPeriod{
				DefaultRetention: &AwsFsxOntapVolumeRetentionDuration{
					Type:  "DAYS",
					Value: 65536,
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects autocommit value above 65535", func() {
		spec.SnaplockConfiguration = &AwsFsxOntapVolumeSnaplockConfiguration{
			SnaplockType: "ENTERPRISE",
			AutocommitPeriod: &AwsFsxOntapVolumeAutocommitPeriod{
				Type:  "DAYS",
				Value: 65536,
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// API envelope validations
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a valid API envelope", func() {
		vol := &AwsFsxOntapVolume{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsFsxOntapVolume",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "my-ontap-volume",
				Id:   "awsfxov-abc123",
				Org:  "my-org",
				Env:  "dev",
			},
			Spec: spec,
		}
		err := protovalidate.Validate(vol)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("rejects invalid api_version in envelope", func() {
		vol := &AwsFsxOntapVolume{
			ApiVersion: "invalid/v1",
			Kind:       "AwsFsxOntapVolume",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "my-ontap-volume",
				Id:   "awsfxov-abc123",
				Org:  "my-org",
				Env:  "dev",
			},
			Spec: spec,
		}
		err := protovalidate.Validate(vol)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects invalid kind in envelope", func() {
		vol := &AwsFsxOntapVolume{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "InvalidKind",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "my-ontap-volume",
				Id:   "awsfxov-abc123",
				Org:  "my-org",
				Env:  "dev",
			},
			Spec: spec,
		}
		err := protovalidate.Validate(vol)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
