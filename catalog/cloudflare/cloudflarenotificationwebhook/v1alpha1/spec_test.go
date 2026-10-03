package cloudflarenotificationwebhookv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestCloudflareNotificationWebhookSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareNotificationWebhookSpec Custom Validation Tests")
}

const testAccountID = "0da42c8d2132a9ddaf714f9e7c920711"

func secretRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func validWebhook(spec *CloudflareNotificationWebhookSpec) *CloudflareNotificationWebhook {
	return &CloudflareNotificationWebhook{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareNotificationWebhook",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-notification-webhook",
		},
		Spec: spec,
	}
}

func baseSpec() *CloudflareNotificationWebhookSpec {
	return &CloudflareNotificationWebhookSpec{
		AccountId: testAccountID,
		Name:      "ops-slack",
		Url:       "https://hooks.slack.com/services/T00/B00/XXXX",
	}
}

var _ = ginkgo.Describe("CloudflareNotificationWebhookSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a webhook without a secret", func() {
			gomega.Expect(protovalidate.Validate(validWebhook(baseSpec()))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a webhook with a secret", func() {
			spec := baseSpec()
			spec.Secret = secretRef("shared-secret-value")
			gomega.Expect(protovalidate.Validate(validWebhook(spec))).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing url", func() {
			spec := baseSpec()
			spec.Url = ""
			gomega.Expect(protovalidate.Validate(validWebhook(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing name", func() {
			spec := baseSpec()
			spec.Name = ""
			gomega.Expect(protovalidate.Validate(validWebhook(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed account id", func() {
			spec := baseSpec()
			spec.AccountId = "abc"
			gomega.Expect(protovalidate.Validate(validWebhook(spec))).NotTo(gomega.BeNil())
		})
	})
})
