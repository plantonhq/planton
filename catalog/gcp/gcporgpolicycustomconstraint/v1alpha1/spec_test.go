package gcporgpolicycustomconstraintv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpOrgPolicyCustomConstraintSpec Suite")
}

var _ = ginkgo.Describe("GcpOrgPolicyCustomConstraintSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpOrgPolicyCustomConstraint {
		return &GcpOrgPolicyCustomConstraint{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpOrgPolicyCustomConstraint",
			Metadata: &shared.CloudResourceMetadata{
				Name: "deny-gke-auto-upgrade-off",
			},
			Spec: &GcpOrgPolicyCustomConstraintSpec{
				OrganizationId: "123456789012",
				ResourceTypes:  []string{"container.googleapis.com/NodePool"},
				MethodTypes:    []string{"CREATE", "UPDATE"},
				Condition:      "resource.management.autoUpgrade == false",
				ActionType:     "DENY",
			},
		}
	}

	expectError := func(target *GcpOrgPolicyCustomConstraint, substring string) {
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(substring))
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal DENY constraint named from metadata", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept an explicit constraint name, display name, description, every method type, and ALLOW", func() {
		target := minimal()
		target.Spec.ConstraintName = "requireShieldedVm"
		target.Spec.DisplayName = "Require Shielded VM"
		target.Spec.Description = "Compute instances must enable Secure Boot; set shieldedInstanceConfig.enableSecureBoot to true"
		target.Spec.ResourceTypes = []string{"compute.googleapis.com/Instance"}
		target.Spec.MethodTypes = []string{"CREATE", "UPDATE", "DELETE", "REMOVE_GRANT", "GOVERN_TAGS"}
		target.Spec.Condition = "resource.shieldedInstanceConfig.enableSecureBoot == true"
		target.Spec.ActionType = "ALLOW"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept every deletion_policy value and the empty default", func() {
		for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = policy
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed(), policy)
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing or non-numeric organization ID", func() {
		target := minimal()
		target.Spec.OrganizationId = ""
		expectError(target, "organization_id")

		target.Spec.OrganizationId = "organizations/123456789012"
		expectError(target, "organization_id")
	})

	ginkgo.It("should reject a constraint name carrying the custom. prefix or illegal characters", func() {
		for _, name := range []string{"custom.disableGkeAutoUpgrade", "disable-gke-auto-upgrade", "1disable", "a123456789012345678901234567890123456789012345678901234567890123"} {
			target := minimal()
			target.Spec.ConstraintName = name
			expectError(target, "WITHOUT the custom. prefix")
		}
	})

	ginkgo.It("should reject an empty or malformed resource type list", func() {
		target := minimal()
		target.Spec.ResourceTypes = nil
		expectError(target, "resource_types")

		target.Spec.ResourceTypes = []string{"NodePool"}
		expectError(target, "resource_types")
	})

	ginkgo.It("should reject an empty or unknown method type", func() {
		target := minimal()
		target.Spec.MethodTypes = nil
		expectError(target, "method_types")

		target.Spec.MethodTypes = []string{"CREATE", "PATCH"}
		expectError(target, "method_types entries must each be one of")
	})

	ginkgo.It("should reject a missing condition and an unknown action type", func() {
		target := minimal()
		target.Spec.Condition = ""
		expectError(target, "condition")

		target = minimal()
		target.Spec.ActionType = "BLOCK"
		expectError(target, "action_type must be one of")
	})

	ginkgo.It("should reject an unknown deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "RETAIN"
		expectError(target, "deletion_policy must be one of")
	})
})
