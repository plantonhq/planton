package gcpsharedvpcserviceprojectv1alpha1

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
	ginkgo.RunSpecs(t, "GcpSharedVpcServiceProjectSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func nameRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: v}},
	}
}

var _ = ginkgo.Describe("GcpSharedVpcServiceProjectSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpSharedVpcServiceProject {
		return &GcpSharedVpcServiceProject{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpSharedVpcServiceProject",
			Metadata: &shared.CloudResourceMetadata{
				Name: "payments-attach",
			},
			Spec: &GcpSharedVpcServiceProjectSpec{
				HostProjectId:    litRef("acme-network-host"),
				ServiceProjectId: litRef("acme-payments-prod"),
			},
		}
	}

	expectError := func(target *GcpSharedVpcServiceProject, substring string) {
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(substring))
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept both projects by literal ID", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept the host by reference to a GcpSharedVpcHost and the service project by reference to a GcpProject", func() {
		target := minimal()
		target.Spec.HostProjectId = nameRef("network-host")
		target.Spec.ServiceProjectId = nameRef("payments-prod")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept the empty deletion_policy (detach) and ABANDON", func() {
		for _, policy := range []string{"", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = policy
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed(), policy)
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing host or service project", func() {
		target := minimal()
		target.Spec.HostProjectId = nil
		expectError(target, "host_project_id")

		target = minimal()
		target.Spec.ServiceProjectId = nil
		expectError(target, "service_project_id")
	})

	ginkgo.It("should reject the DELETE and PREVENT policies this resource does not accept", func() {
		for _, policy := range []string{"DELETE", "PREVENT", "RETAIN"} {
			target := minimal()
			target.Spec.DeletionPolicy = policy
			expectError(target, "deletion_policy must be empty (detach on destroy) or ABANDON")
		}
	})
})
