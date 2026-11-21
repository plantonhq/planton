package gcpbigqueryreservationgroupv1alpha1

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
	ginkgo.RunSpecs(t, "GcpBigQueryReservationGroupSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpBigQueryReservationGroupSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpBigQueryReservationGroup {
		return &GcpBigQueryReservationGroup{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpBigQueryReservationGroup",
			Metadata:   &shared.CatalogObjectMetadata{Name: "tier-1"},
			Spec:       &GcpBigQueryReservationGroupSpec{},
		}
	}

	ginkgo.It("should accept a group with nothing set", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("bq-admin")
		msg.Spec.Location = "EU"
		msg.Spec.ReservationGroupName = "tier-1"
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a location in the wrong case", func() {
		for _, loc := range []string{"us", "eu", "Europe-West1"} {
			msg := minimal()
			msg.Spec.Location = loc
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), loc)
		}
		msg := minimal()
		msg.Spec.Location = "europe-west1"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed name and an unknown deletion policy", func() {
		for _, name := range []string{"tier_1", "Tier-1", "1tier", "tier-",
			"a234567890123456789012345678901234567890123456789012345678901234x"} {
			msg := minimal()
			msg.Spec.ReservationGroupName = name
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), name)
		}
		msg := minimal()
		msg = minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
