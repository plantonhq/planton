package gcpsccmuteconfigv1alpha1

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
	ginkgo.RunSpecs(t, "GcpSccMuteConfigSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpSccMuteConfigSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpSccMuteConfig {
		return &GcpSccMuteConfig{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpSccMuteConfig",
			Metadata:   &shared.CatalogObjectMetadata{Name: "sandbox-public-buckets"},
			Spec: &GcpSccMuteConfigSpec{
				MuteConfigId: "sandbox-public-buckets",
				Filter:       `category = "PUBLIC_BUCKET_ACL"`,
				Type:         "DYNAMIC",
			},
		}
	}

	ginkgo.It("should accept a dynamic rule on the provider's project", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a static folder rule with every field set", func() {
		msg := minimal()
		msg.Spec.Scope = &GcpSccMuteConfigScope{FolderId: litRef("456")}
		msg.Spec.Type = "STATIC"
		msg.Spec.Description = "Sandbox buckets are public by design"
		msg.Spec.Location = "global"
		msg.Spec.DeletionPolicy = "ABANDON"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require a filter and a known type", func() {
		msg := minimal()
		msg.Spec.Filter = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Type = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Type = "MUTE_CONFIG_TYPE_UNSPECIFIED"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a mute config id outside Google's rule", func() {
		for _, id := range []string{"", "Upper", "1starts-with-digit", "ends-with-", "under_score"} {
			msg := minimal()
			msg.Spec.MuteConfigId = id
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), id)
		}
	})

	ginkgo.It("should reject two scope arms", func() {
		msg := minimal()
		msg.Spec.Scope = &GcpSccMuteConfigScope{FolderId: litRef("456"), OrganizationId: "123"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
