package gcpgkefleetv1alpha1

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
	ginkgo.RunSpecs(t, "GcpGkeFleetSpec Suite")
}

var _ = ginkgo.Describe("GcpGkeFleetSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpGkeFleet {
		return &GcpGkeFleet{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpGkeFleet",
			Metadata:   &shared.CatalogObjectMetadata{Name: "platform-fleet"},
			Spec:       &GcpGkeFleetSpec{},
		}
	}

	ginkgo.It("should accept an empty spec (the provider's project, Google's display name)", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set", func() {
		msg := minimal()
		msg.Spec.ProjectId = &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "platform-host"}}
		msg.Spec.DisplayName = "Platform 'prod' fleet!"
		msg.Spec.DefaultClusterConfig = &GcpGkeFleetDefaultClusterConfig{
			BinaryAuthorizationConfig: &GcpGkeFleetBinaryAuthorizationConfig{
				EvaluationMode: "POLICY_BINDINGS",
				PolicyBindings: []string{"projects/123456789/platforms/gke/policies/baseline"},
			},
			SecurityPostureConfig: &GcpGkeFleetSecurityPostureConfig{Mode: "BASIC", VulnerabilityMode: "VULNERABILITY_BASIC"},
		}
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a display name outside 4-30 allowed characters", func() {
		for _, name := range []string{"abc", "a fleet name that is far too long!", "fleet/prod", "fleet_prod"} {
			msg := minimal()
			msg.Spec.DisplayName = name
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), name)
		}
	})

	ginkgo.It("should require POLICY_BINDINGS mode for policy bindings", func() {
		msg := minimal()
		msg.Spec.DefaultClusterConfig = &GcpGkeFleetDefaultClusterConfig{
			BinaryAuthorizationConfig: &GcpGkeFleetBinaryAuthorizationConfig{
				EvaluationMode: "DISABLED",
				PolicyBindings: []string{"projects/123/platforms/gke/policies/p"},
			},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a policy binding that is not a platform policy name, and duplicates", func() {
		for _, bindings := range [][]string{
			{"projects/my-project/platforms/gke/policies/p"},
			{"baseline"},
			{"projects/1/platforms/gke/policies/p", "projects/1/platforms/gke/policies/p"},
		} {
			msg := minimal()
			msg.Spec.DefaultClusterConfig = &GcpGkeFleetDefaultClusterConfig{
				BinaryAuthorizationConfig: &GcpGkeFleetBinaryAuthorizationConfig{EvaluationMode: "POLICY_BINDINGS", PolicyBindings: bindings},
			}
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), bindings)
		}
	})

	ginkgo.It("should reject unknown modes and deletion policies", func() {
		msg := minimal()
		msg.Spec.DefaultClusterConfig = &GcpGkeFleetDefaultClusterConfig{BinaryAuthorizationConfig: &GcpGkeFleetBinaryAuthorizationConfig{EvaluationMode: "ENFORCED"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DefaultClusterConfig = &GcpGkeFleetDefaultClusterConfig{SecurityPostureConfig: &GcpGkeFleetSecurityPostureConfig{Mode: "STANDARD"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DefaultClusterConfig = &GcpGkeFleetDefaultClusterConfig{SecurityPostureConfig: &GcpGkeFleetSecurityPostureConfig{VulnerabilityMode: "BASIC"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DeletionPolicy = "RETAIN"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
