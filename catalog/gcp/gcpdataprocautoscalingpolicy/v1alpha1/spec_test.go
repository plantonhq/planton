package gcpdataprocautoscalingpolicyv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpDataprocAutoscalingPolicySpec Suite")
}

var _ = ginkgo.Describe("GcpDataprocAutoscalingPolicySpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper for StringValueOrRef with a literal value.
	svr := func(v string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
		}
	}

	// Helper for explicit-presence doubles (the scale factors), where a
	// set 0.0 is a legitimate value distinct from unset.
	f64 := func(v float64) *float64 { return &v }

	// Helper to build a minimal valid GcpDataprocAutoscalingPolicy.
	minimal := func() *GcpDataprocAutoscalingPolicy {
		return &GcpDataprocAutoscalingPolicy{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpDataprocAutoscalingPolicy",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-autoscaling-policy",
			},
			Spec: &GcpDataprocAutoscalingPolicySpec{
				PolicyId: "etl-autoscaling",
				Location: "us-central1",
				WorkerConfig: &GcpDataprocAutoscalingPolicyWorkerConfig{
					MaxInstances: 10,
				},
				BasicAlgorithm: &GcpDataprocAutoscalingPolicyBasicAlgorithm{
					YarnConfig: &GcpDataprocAutoscalingPolicyYarnConfig{
						GracefulDecommissionTimeout: "3600s",
						ScaleUpFactor:               f64(0.5),
						ScaleDownFactor:             f64(1.0),
					},
				},
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec", func() {
		err := validator.Validate(minimal())
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a project_id literal", func() {
		msg := minimal()
		msg.Spec.ProjectId = svr("my-gcp-project")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a multi-digit region", func() {
		msg := minimal()
		msg.Spec.Location = "europe-west12"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept policy_id at minimum length (3 chars)", func() {
		msg := minimal()
		msg.Spec.PolicyId = "abc"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept policy_id at maximum length (50 chars)", func() {
		msg := minimal()
		id := "a"
		for i := 0; i < 48; i++ {
			id += "b"
		}
		id += "z"
		msg.Spec.PolicyId = id
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept policy_id with underscores and hyphens", func() {
		msg := minimal()
		msg.Spec.PolicyId = "etl_batch-policy2"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept worker_config min_instances of 0 (API default)", func() {
		msg := minimal()
		msg.Spec.WorkerConfig.MinInstances = 0
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept worker_config min_instances of 2 or more", func() {
		msg := minimal()
		msg.Spec.WorkerConfig.MinInstances = 2
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept worker_config with a weight", func() {
		msg := minimal()
		msg.Spec.WorkerConfig.Weight = 1
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a secondary_worker_config with bounds and weight", func() {
		msg := minimal()
		msg.Spec.SecondaryWorkerConfig = &GcpDataprocAutoscalingPolicySecondaryWorkerConfig{
			MaxInstances: 20,
			MinInstances: 5,
			Weight:       3,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a scale-to-zero secondary_worker_config", func() {
		msg := minimal()
		msg.Spec.SecondaryWorkerConfig = &GcpDataprocAutoscalingPolicySecondaryWorkerConfig{
			MaxInstances: 50,
			MinInstances: 0,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a cooldown_period in seconds", func() {
		msg := minimal()
		msg.Spec.BasicAlgorithm.CooldownPeriod = "120s"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept scale factors at the 1.0 upper bound", func() {
		msg := minimal()
		msg.Spec.BasicAlgorithm.YarnConfig.ScaleUpFactor = f64(1.0)
		msg.Spec.BasicAlgorithm.YarnConfig.ScaleDownFactor = f64(1.0)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept small fractional scale factors", func() {
		msg := minimal()
		msg.Spec.BasicAlgorithm.YarnConfig.ScaleUpFactor = f64(0.05)
		msg.Spec.BasicAlgorithm.YarnConfig.ScaleDownFactor = f64(0.05)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an explicit 0.0 scale_down_factor (scale-down disabled)", func() {
		msg := minimal()
		msg.Spec.BasicAlgorithm.YarnConfig.ScaleDownFactor = f64(0.0)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an explicit 0.0 scale_up_factor", func() {
		msg := minimal()
		msg.Spec.BasicAlgorithm.YarnConfig.ScaleUpFactor = f64(0.0)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept min_worker_fractions within 0.0-1.0", func() {
		msg := minimal()
		msg.Spec.BasicAlgorithm.YarnConfig.ScaleUpMinWorkerFraction = 0.1
		msg.Spec.BasicAlgorithm.YarnConfig.ScaleDownMinWorkerFraction = 0.25
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a fully-featured policy", func() {
		msg := minimal()
		msg.Spec.ProjectId = svr("my-gcp-project")
		msg.Spec.WorkerConfig = &GcpDataprocAutoscalingPolicyWorkerConfig{
			MaxInstances: 10,
			MinInstances: 2,
			Weight:       1,
		}
		msg.Spec.SecondaryWorkerConfig = &GcpDataprocAutoscalingPolicySecondaryWorkerConfig{
			MaxInstances: 20,
			MinInstances: 0,
			Weight:       3,
		}
		msg.Spec.BasicAlgorithm = &GcpDataprocAutoscalingPolicyBasicAlgorithm{
			CooldownPeriod: "120s",
			YarnConfig: &GcpDataprocAutoscalingPolicyYarnConfig{
				GracefulDecommissionTimeout: "3600s",
				ScaleUpFactor:               f64(0.5),
				ScaleDownFactor:             f64(1.0),
				ScaleUpMinWorkerFraction:    0.05,
				ScaleDownMinWorkerFraction:  0.05,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing policy_id", func() {
		msg := minimal()
		msg.Spec.PolicyId = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("policy_id"))
	})

	ginkgo.It("should reject a policy_id shorter than 3 characters", func() {
		msg := minimal()
		msg.Spec.PolicyId = "ab"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a policy_id longer than 50 characters", func() {
		msg := minimal()
		id := ""
		for i := 0; i < 51; i++ {
			id += "a"
		}
		msg.Spec.PolicyId = id
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a policy_id starting with a hyphen", func() {
		msg := minimal()
		msg.Spec.PolicyId = "-bad-policy"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a policy_id ending with an underscore", func() {
		msg := minimal()
		msg.Spec.PolicyId = "bad_policy_"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing location", func() {
		msg := minimal()
		msg.Spec.Location = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("location"))
	})

	ginkgo.It("should reject a zone where a region is expected", func() {
		msg := minimal()
		msg.Spec.Location = "us-central1-a"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing worker_config", func() {
		msg := minimal()
		msg.Spec.WorkerConfig = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject worker_config without max_instances", func() {
		msg := minimal()
		msg.Spec.WorkerConfig.MaxInstances = 0
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject worker_config min_instances of 1", func() {
		msg := minimal()
		msg.Spec.WorkerConfig.MinInstances = 1
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("min_instances"))
	})

	ginkgo.It("should reject worker_config with max below min", func() {
		msg := minimal()
		msg.Spec.WorkerConfig.MaxInstances = 3
		msg.Spec.WorkerConfig.MinInstances = 5
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("max_instances"))
	})

	ginkgo.It("should reject worker_config with a negative weight", func() {
		msg := minimal()
		msg.Spec.WorkerConfig.Weight = -1
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject secondary_worker_config with max below min", func() {
		msg := minimal()
		msg.Spec.SecondaryWorkerConfig = &GcpDataprocAutoscalingPolicySecondaryWorkerConfig{
			MaxInstances: 3,
			MinInstances: 5,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject secondary_worker_config with negative bounds", func() {
		msg := minimal()
		msg.Spec.SecondaryWorkerConfig = &GcpDataprocAutoscalingPolicySecondaryWorkerConfig{
			MaxInstances: -1,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing basic_algorithm", func() {
		msg := minimal()
		msg.Spec.BasicAlgorithm = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing yarn_config", func() {
		msg := minimal()
		msg.Spec.BasicAlgorithm.YarnConfig = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing graceful_decommission_timeout", func() {
		msg := minimal()
		msg.Spec.BasicAlgorithm.YarnConfig.GracefulDecommissionTimeout = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid graceful_decommission_timeout format", func() {
		msg := minimal()
		msg.Spec.BasicAlgorithm.YarnConfig.GracefulDecommissionTimeout = "1h"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid cooldown_period format", func() {
		msg := minimal()
		msg.Spec.BasicAlgorithm.CooldownPeriod = "2m"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("cooldown_period"))
	})

	ginkgo.It("should reject a cooldown_period below 2 minutes", func() {
		msg := minimal()
		msg.Spec.BasicAlgorithm.CooldownPeriod = "10s"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("must be between 2 minutes"))
	})

	ginkgo.It("should reject a cooldown_period above 1 day", func() {
		msg := minimal()
		msg.Spec.BasicAlgorithm.CooldownPeriod = "90000s"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("must be between 2 minutes"))
	})

	ginkgo.It("should accept cooldown_period at both bounds", func() {
		msg := minimal()
		msg.Spec.BasicAlgorithm.CooldownPeriod = "86400s"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.HaveOccurred())
		msg.Spec.BasicAlgorithm.CooldownPeriod = "120s"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a graceful_decommission_timeout above 1 day", func() {
		msg := minimal()
		msg.Spec.BasicAlgorithm.YarnConfig.GracefulDecommissionTimeout = "90000s"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("must be at most 1 day"))
	})

	ginkgo.It("should accept graceful_decommission_timeout at both bounds", func() {
		msg := minimal()
		msg.Spec.BasicAlgorithm.YarnConfig.GracefulDecommissionTimeout = "0s"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.HaveOccurred())
		msg.Spec.BasicAlgorithm.YarnConfig.GracefulDecommissionTimeout = "86400s"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a scale_up_factor above 1.0", func() {
		msg := minimal()
		msg.Spec.BasicAlgorithm.YarnConfig.ScaleUpFactor = f64(1.5)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a scale_down_factor above 1.0", func() {
		msg := minimal()
		msg.Spec.BasicAlgorithm.YarnConfig.ScaleDownFactor = f64(2.0)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an unset scale_up_factor", func() {
		msg := minimal()
		msg.Spec.BasicAlgorithm.YarnConfig.ScaleUpFactor = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an unset scale_down_factor", func() {
		msg := minimal()
		msg.Spec.BasicAlgorithm.YarnConfig.ScaleDownFactor = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject secondary_worker_config with a floor but no ceiling", func() {
		msg := minimal()
		msg.Spec.SecondaryWorkerConfig = &GcpDataprocAutoscalingPolicySecondaryWorkerConfig{
			MaxInstances: 0,
			MinInstances: 5,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("max_instances"))
	})

	ginkgo.It("should reject a scale_up_min_worker_fraction above 1.0", func() {
		msg := minimal()
		msg.Spec.BasicAlgorithm.YarnConfig.ScaleUpMinWorkerFraction = 1.1
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a scale_down_min_worker_fraction above 1.0", func() {
		msg := minimal()
		msg.Spec.BasicAlgorithm.YarnConfig.ScaleDownMinWorkerFraction = 1.1
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject wrong api_version", func() {
		msg := minimal()
		msg.ApiVersion = "wrong/v1"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject wrong kind", func() {
		msg := minimal()
		msg.Kind = "WrongKind"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing metadata", func() {
		msg := minimal()
		msg.Metadata = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing spec", func() {
		msg := minimal()
		msg.Spec = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept every valid deletion_policy", func() {
		for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			msg := minimal()
			msg.Spec.DeletionPolicy = policy
			gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should reject an unknown deletion_policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "RETAIN"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
