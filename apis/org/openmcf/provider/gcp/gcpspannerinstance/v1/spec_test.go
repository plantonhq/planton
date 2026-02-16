package gcpspannerinstancev1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpSpannerInstanceSpec Suite")
}

var _ = ginkgo.Describe("GcpSpannerInstanceSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpSpannerInstance with num_nodes.
	minimal := func() *GcpSpannerInstance {
		return &GcpSpannerInstance{
			ApiVersion: "gcp.openmcf.org/v1",
			Kind:       "GcpSpannerInstance",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-spanner",
			},
			Spec: &GcpSpannerInstanceSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-gcp-project",
					},
				},
				InstanceName: "my-spanner-db",
				Config:       "regional-us-central1",
				DisplayName:  "My Spanner Instance",
				NumNodes:     1,
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec with num_nodes", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a minimal valid spec with processing_units", func() {
		msg := minimal()
		msg.Spec.NumNodes = 0
		msg.Spec.ProcessingUnits = 1000
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a valid spec with autoscaling using nodes", func() {
		msg := minimal()
		msg.Spec.NumNodes = 0
		msg.Spec.AutoscalingConfig = &GcpSpannerInstanceAutoscalingConfig{
			AutoscalingLimits: &GcpSpannerInstanceAutoscalingLimits{
				MinNodes: 1,
				MaxNodes: 3,
			},
			AutoscalingTargets: &GcpSpannerInstanceAutoscalingTargets{
				HighPriorityCpuUtilizationPercent: 65,
				StorageUtilizationPercent:         80,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a valid spec with autoscaling using processing_units", func() {
		msg := minimal()
		msg.Spec.NumNodes = 0
		msg.Spec.AutoscalingConfig = &GcpSpannerInstanceAutoscalingConfig{
			AutoscalingLimits: &GcpSpannerInstanceAutoscalingLimits{
				MinProcessingUnits: 1000,
				MaxProcessingUnits: 5000,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept autoscaling with only limits (no targets)", func() {
		msg := minimal()
		msg.Spec.NumNodes = 0
		msg.Spec.AutoscalingConfig = &GcpSpannerInstanceAutoscalingConfig{
			AutoscalingLimits: &GcpSpannerInstanceAutoscalingLimits{
				MinNodes: 1,
				MaxNodes: 5,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept FREE_INSTANCE with no capacity", func() {
		msg := minimal()
		msg.Spec.NumNodes = 0
		msg.Spec.InstanceType = "FREE_INSTANCE"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept PROVISIONED instance type explicitly", func() {
		msg := minimal()
		msg.Spec.InstanceType = "PROVISIONED"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept edition STANDARD", func() {
		msg := minimal()
		msg.Spec.Edition = "STANDARD"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept edition ENTERPRISE", func() {
		msg := minimal()
		msg.Spec.Edition = "ENTERPRISE"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept edition ENTERPRISE_PLUS", func() {
		msg := minimal()
		msg.Spec.Edition = "ENTERPRISE_PLUS"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept instance_name at minimum boundary (6 chars)", func() {
		msg := minimal()
		msg.Spec.InstanceName = "abcde1"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept instance_name at maximum boundary (30 chars)", func() {
		msg := minimal()
		msg.Spec.InstanceName = "a" + strings.Repeat("b", 28) + "c"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept instance_name with hyphens and numbers", func() {
		msg := minimal()
		msg.Spec.InstanceName = "spanner-prod-01"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept display_name at minimum boundary (4 chars)", func() {
		msg := minimal()
		msg.Spec.DisplayName = "Test"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept display_name at maximum boundary (30 chars)", func() {
		msg := minimal()
		msg.Spec.DisplayName = strings.Repeat("A", 30)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept default_backup_schedule_type NONE", func() {
		msg := minimal()
		msg.Spec.DefaultBackupScheduleType = "NONE"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept default_backup_schedule_type AUTOMATIC", func() {
		msg := minimal()
		msg.Spec.DefaultBackupScheduleType = "AUTOMATIC"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept force_destroy true", func() {
		msg := minimal()
		msg.Spec.ForceDestroy = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept multi-region config", func() {
		msg := minimal()
		msg.Spec.Config = "nam-eur-asia1"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept full-load scenario with all optional fields", func() {
		msg := minimal()
		msg.Spec.NumNodes = 0
		msg.Spec.InstanceType = "PROVISIONED"
		msg.Spec.Edition = "ENTERPRISE"
		msg.Spec.DefaultBackupScheduleType = "AUTOMATIC"
		msg.Spec.ForceDestroy = true
		msg.Spec.AutoscalingConfig = &GcpSpannerInstanceAutoscalingConfig{
			AutoscalingLimits: &GcpSpannerInstanceAutoscalingLimits{
				MinNodes: 1,
				MaxNodes: 10,
			},
			AutoscalingTargets: &GcpSpannerInstanceAutoscalingTargets{
				HighPriorityCpuUtilizationPercent: 65,
				StorageUtilizationPercent:         80,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept empty instance_type (defaults to PROVISIONED)", func() {
		msg := minimal()
		msg.Spec.InstanceType = ""
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept empty edition", func() {
		msg := minimal()
		msg.Spec.Edition = ""
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept autoscaling targets at boundary (0 and 100)", func() {
		msg := minimal()
		msg.Spec.NumNodes = 0
		msg.Spec.AutoscalingConfig = &GcpSpannerInstanceAutoscalingConfig{
			AutoscalingLimits: &GcpSpannerInstanceAutoscalingLimits{
				MinNodes: 1,
				MaxNodes: 3,
			},
			AutoscalingTargets: &GcpSpannerInstanceAutoscalingTargets{
				HighPriorityCpuUtilizationPercent: 0,
				StorageUtilizationPercent:         100,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject when project_id is missing", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when instance_name is empty", func() {
		msg := minimal()
		msg.Spec.InstanceName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject instance_name shorter than 6 characters", func() {
		msg := minimal()
		msg.Spec.InstanceName = "abcde"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject instance_name longer than 30 characters", func() {
		msg := minimal()
		msg.Spec.InstanceName = "a" + strings.Repeat("b", 30)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject instance_name starting with a digit", func() {
		msg := minimal()
		msg.Spec.InstanceName = "1spanner-db"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject instance_name starting with a hyphen", func() {
		msg := minimal()
		msg.Spec.InstanceName = "-spanner-db"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject instance_name with uppercase letters", func() {
		msg := minimal()
		msg.Spec.InstanceName = "MySpanner"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject instance_name ending with a hyphen", func() {
		msg := minimal()
		msg.Spec.InstanceName = "spanner-"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject instance_name with underscores", func() {
		msg := minimal()
		msg.Spec.InstanceName = "spanner_db"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when config is empty", func() {
		msg := minimal()
		msg.Spec.Config = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when display_name is empty", func() {
		msg := minimal()
		msg.Spec.DisplayName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject display_name shorter than 4 characters", func() {
		msg := minimal()
		msg.Spec.DisplayName = "abc"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject display_name longer than 30 characters", func() {
		msg := minimal()
		msg.Spec.DisplayName = strings.Repeat("A", 31)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when metadata is missing", func() {
		msg := minimal()
		msg.Metadata = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when spec is missing", func() {
		msg := minimal()
		msg.Spec = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid instance_type", func() {
		msg := minimal()
		msg.Spec.InstanceType = "MANAGED"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid edition", func() {
		msg := minimal()
		msg.Spec.Edition = "PREMIUM"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid default_backup_schedule_type", func() {
		msg := minimal()
		msg.Spec.DefaultBackupScheduleType = "DAILY"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject num_nodes and processing_units both set", func() {
		msg := minimal()
		msg.Spec.NumNodes = 1
		msg.Spec.ProcessingUnits = 1000
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject num_nodes and autoscaling_config both set", func() {
		msg := minimal()
		msg.Spec.NumNodes = 1
		msg.Spec.AutoscalingConfig = &GcpSpannerInstanceAutoscalingConfig{
			AutoscalingLimits: &GcpSpannerInstanceAutoscalingLimits{
				MinNodes: 1,
				MaxNodes: 3,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject processing_units and autoscaling_config both set", func() {
		msg := minimal()
		msg.Spec.NumNodes = 0
		msg.Spec.ProcessingUnits = 1000
		msg.Spec.AutoscalingConfig = &GcpSpannerInstanceAutoscalingConfig{
			AutoscalingLimits: &GcpSpannerInstanceAutoscalingLimits{
				MinProcessingUnits: 1000,
				MaxProcessingUnits: 5000,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject FREE_INSTANCE with num_nodes", func() {
		msg := minimal()
		msg.Spec.InstanceType = "FREE_INSTANCE"
		msg.Spec.NumNodes = 1
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject FREE_INSTANCE with processing_units", func() {
		msg := minimal()
		msg.Spec.InstanceType = "FREE_INSTANCE"
		msg.Spec.NumNodes = 0
		msg.Spec.ProcessingUnits = 1000
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject FREE_INSTANCE with autoscaling_config", func() {
		msg := minimal()
		msg.Spec.InstanceType = "FREE_INSTANCE"
		msg.Spec.NumNodes = 0
		msg.Spec.AutoscalingConfig = &GcpSpannerInstanceAutoscalingConfig{
			AutoscalingLimits: &GcpSpannerInstanceAutoscalingLimits{
				MinNodes: 1,
				MaxNodes: 3,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject FREE_INSTANCE with edition set", func() {
		msg := minimal()
		msg.Spec.InstanceType = "FREE_INSTANCE"
		msg.Spec.NumNodes = 0
		msg.Spec.Edition = "ENTERPRISE"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject FREE_INSTANCE with AUTOMATIC backup schedule", func() {
		msg := minimal()
		msg.Spec.InstanceType = "FREE_INSTANCE"
		msg.Spec.NumNodes = 0
		msg.Spec.DefaultBackupScheduleType = "AUTOMATIC"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject autoscaling with mixed units (min_nodes + max_processing_units)", func() {
		msg := minimal()
		msg.Spec.NumNodes = 0
		msg.Spec.AutoscalingConfig = &GcpSpannerInstanceAutoscalingConfig{
			AutoscalingLimits: &GcpSpannerInstanceAutoscalingLimits{
				MinNodes:           1,
				MaxProcessingUnits: 5000,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject autoscaling with max_nodes < min_nodes", func() {
		msg := minimal()
		msg.Spec.NumNodes = 0
		msg.Spec.AutoscalingConfig = &GcpSpannerInstanceAutoscalingConfig{
			AutoscalingLimits: &GcpSpannerInstanceAutoscalingLimits{
				MinNodes: 5,
				MaxNodes: 2,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject autoscaling with max_processing_units < min_processing_units", func() {
		msg := minimal()
		msg.Spec.NumNodes = 0
		msg.Spec.AutoscalingConfig = &GcpSpannerInstanceAutoscalingConfig{
			AutoscalingLimits: &GcpSpannerInstanceAutoscalingLimits{
				MinProcessingUnits: 5000,
				MaxProcessingUnits: 1000,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject autoscaling targets CPU utilization > 100", func() {
		msg := minimal()
		msg.Spec.NumNodes = 0
		msg.Spec.AutoscalingConfig = &GcpSpannerInstanceAutoscalingConfig{
			AutoscalingLimits: &GcpSpannerInstanceAutoscalingLimits{
				MinNodes: 1,
				MaxNodes: 3,
			},
			AutoscalingTargets: &GcpSpannerInstanceAutoscalingTargets{
				HighPriorityCpuUtilizationPercent: 101,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject autoscaling targets storage utilization > 100", func() {
		msg := minimal()
		msg.Spec.NumNodes = 0
		msg.Spec.AutoscalingConfig = &GcpSpannerInstanceAutoscalingConfig{
			AutoscalingLimits: &GcpSpannerInstanceAutoscalingLimits{
				MinNodes: 1,
				MaxNodes: 3,
			},
			AutoscalingTargets: &GcpSpannerInstanceAutoscalingTargets{
				StorageUtilizationPercent: 150,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject autoscaling without limits", func() {
		msg := minimal()
		msg.Spec.NumNodes = 0
		msg.Spec.AutoscalingConfig = &GcpSpannerInstanceAutoscalingConfig{
			AutoscalingTargets: &GcpSpannerInstanceAutoscalingTargets{
				HighPriorityCpuUtilizationPercent: 65,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
