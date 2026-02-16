package awsfsxontapvolumev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	shared "github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
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

func int32Val(i int32) int32 {
	return i
}

var _ = ginkgo.Describe("AwsFsxOntapVolumeSpec validations", func() {
	var spec *AwsFsxOntapVolumeSpec

	ginkgo.BeforeEach(func() {
		spec = &AwsFsxOntapVolumeSpec{
			StorageVirtualMachineId: strRef("svm-0123456789abcdef0"),
			Name:                    "vol_default",
			SizeInMegabytes:         1024,
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
		spec.StorageEfficiencyEnabled = true
		spec.CopyTagsToBackups = boolPtr(true)
		spec.TieringPolicy = &AwsFsxOntapVolumeTieringPolicy{
			Name:          "AUTO",
			CoolingPeriod: 31,
		}
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
		spec.SizeInMegabytes = 19
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects size_in_megabytes of zero", func() {
		spec.SizeInMegabytes = 0
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

	// -------------------------------------------------------------------------
	// API envelope validations
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a valid API envelope", func() {
		vol := &AwsFsxOntapVolume{
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsFsxOntapVolume",
			Metadata: &shared.CloudResourceMetadata{
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
			Metadata: &shared.CloudResourceMetadata{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "InvalidKind",
			Metadata: &shared.CloudResourceMetadata{
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
