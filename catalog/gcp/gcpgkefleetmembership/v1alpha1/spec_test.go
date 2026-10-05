package gcpgkefleetmembershipv1alpha1

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
	ginkgo.RunSpecs(t, "GcpGkeFleetMembershipSpec Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value}}
}

var _ = ginkgo.Describe("GcpGkeFleetMembershipSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpGkeFleetMembership {
		return &GcpGkeFleetMembership{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpGkeFleetMembership",
			Metadata:   &shared.CatalogObjectMetadata{Name: "orders-cluster"},
			Spec: &GcpGkeFleetMembershipSpec{
				GkeCluster: literal("projects/orders-prod/locations/us-central1/clusters/orders"),
			},
		}
	}

	ginkgo.It("should accept a cluster registration with defaults", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set, including fleet Workload Identity", func() {
		msg := minimal()
		msg.Spec.ProjectId = literal("platform-host")
		msg.Spec.MembershipId = "orders-prod-uc1"
		msg.Spec.Location = "us-central1"
		msg.Spec.Issuer = "https://container.googleapis.com/v1/projects/orders-prod/locations/us-central1/clusters/orders"
		msg.Spec.Labels = map[string]string{"team": "orders"}
		msg.Spec.DeletionPolicy = "ABANDON"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a membership with only an issuer (a cluster outside Google Cloud)", func() {
		msg := minimal()
		msg.Spec.GkeCluster = nil
		msg.Spec.Issuer = "https://oidc.example.com/clusters/edge-1"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed membership ID or location", func() {
		for _, id := range []string{"Orders", "-orders", "orders-", "orders_cluster"} {
			msg := minimal()
			msg.Spec.MembershipId = id
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), id)
		}
		msg := minimal()
		msg.Spec.Location = "US-Central1"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an issuer that is not an https URL", func() {
		msg := minimal()
		msg.Spec.Issuer = "http://container.googleapis.com/v1/projects/p/locations/l/clusters/c"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
