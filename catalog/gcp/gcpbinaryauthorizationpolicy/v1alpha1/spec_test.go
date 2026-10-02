package gcpbinaryauthorizationpolicyv1alpha1

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
	ginkgo.RunSpecs(t, "GcpBinaryAuthorizationPolicySpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpBinaryAuthorizationPolicySpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpBinaryAuthorizationPolicy {
		return &GcpBinaryAuthorizationPolicy{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpBinaryAuthorizationPolicy",
			Metadata:   &shared.CloudResourceMetadata{Name: "project-policy"},
			Spec: &GcpBinaryAuthorizationPolicySpec{
				DefaultAdmissionRule: &GcpBinaryAuthorizationPolicyAdmissionRule{
					EvaluationMode:  "ALWAYS_ALLOW",
					EnforcementMode: "ENFORCED_BLOCK_AND_AUDIT_LOG",
				},
			},
		}
	}

	attested := func() *GcpBinaryAuthorizationPolicy {
		msg := minimal()
		msg.Spec.DefaultAdmissionRule = &GcpBinaryAuthorizationPolicyAdmissionRule{
			EvaluationMode:        "REQUIRE_ATTESTATION",
			EnforcementMode:       "DRYRUN_AUDIT_LOG_ONLY",
			RequireAttestationsBy: []*foreignkeyv1.StringValueOrRef{litRef("projects/sec/attestors/built-by-ci")},
		}
		return msg
	}

	ginkgo.It("should accept an allow-all default rule", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set with per-cluster rules", func() {
		msg := attested()
		msg.Spec.ProjectId = litRef("prod")
		msg.Spec.Description = "Signed images only"
		msg.Spec.GlobalPolicyEvaluationMode = "ENABLE"
		msg.Spec.AdmissionWhitelistPatterns = []string{"us-docker.pkg.dev/prod/base/*"}
		msg.Spec.ClusterAdmissionRules = []*GcpBinaryAuthorizationPolicyClusterAdmissionRule{
			{Cluster: "us-central1.prod", EvaluationMode: "REQUIRE_ATTESTATION", EnforcementMode: "ENFORCED_BLOCK_AND_AUDIT_LOG",
				RequireAttestationsBy: []*foreignkeyv1.StringValueOrRef{litRef("built-by-ci"), litRef("qa-approved")}},
			{Cluster: "us-central1-a.sandbox", EvaluationMode: "ALWAYS_ALLOW", EnforcementMode: "DRYRUN_AUDIT_LOG_ONLY"},
		}
		msg.Spec.DeletionPolicy = "ABANDON"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require a default rule with both modes", func() {
		msg := minimal()
		msg.Spec.DefaultAdmissionRule = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DefaultAdmissionRule.EnforcementMode = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should pair attestors with REQUIRE_ATTESTATION only", func() {
		msg := attested()
		msg.Spec.DefaultAdmissionRule.RequireAttestationsBy = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DefaultAdmissionRule.RequireAttestationsBy = []*foreignkeyv1.StringValueOrRef{litRef("built-by-ci")}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.ClusterAdmissionRules = []*GcpBinaryAuthorizationPolicyClusterAdmissionRule{
			{Cluster: "us-central1.prod", EvaluationMode: "REQUIRE_ATTESTATION", EnforcementMode: "ENFORCED_BLOCK_AND_AUDIT_LOG"},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject two rules for one cluster and a malformed cluster key", func() {
		msg := minimal()
		rule := func(cluster string) *GcpBinaryAuthorizationPolicyClusterAdmissionRule {
			return &GcpBinaryAuthorizationPolicyClusterAdmissionRule{Cluster: cluster, EvaluationMode: "ALWAYS_DENY", EnforcementMode: "ENFORCED_BLOCK_AND_AUDIT_LOG"}
		}
		msg.Spec.ClusterAdmissionRules = []*GcpBinaryAuthorizationPolicyClusterAdmissionRule{rule("us-central1.prod"), rule("us-central1.prod")}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		for _, cluster := range []string{"prod", "projects/p/locations/us-central1/clusters/prod", "us-central1/prod"} {
			msg = minimal()
			msg.Spec.ClusterAdmissionRules = []*GcpBinaryAuthorizationPolicyClusterAdmissionRule{rule(cluster)}
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), cluster)
		}
	})

	ginkgo.It("should reject unknown modes", func() {
		msg := minimal()
		msg.Spec.DefaultAdmissionRule.EvaluationMode = "ALLOW"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.GlobalPolicyEvaluationMode = "ON"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
