package cloudflarezerotrustaccessservicetokenv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestCloudflareZeroTrustAccessServiceTokenSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareZeroTrustAccessServiceTokenSpec Custom Validation Tests")
}

const testAccountId = "023e105f4ecef8ad9ca31a8372d0c353"

func int32Ptr(v int32) *int32 { return &v }

func zoneRef() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "023e105f4ecef8ad9ca31a8372d0c353"}}
}

func validToken(spec *CloudflareZeroTrustAccessServiceTokenSpec) *CloudflareZeroTrustAccessServiceToken {
	return &CloudflareZeroTrustAccessServiceToken{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareZeroTrustAccessServiceToken",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-service-token",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("CloudflareZeroTrustAccessServiceTokenSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal account-scoped token", func() {
			input := validToken(&CloudflareZeroTrustAccessServiceTokenSpec{
				AccountId: testAccountId,
				Name:      "ci-deployer",
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a zone-scoped token", func() {
			input := validToken(&CloudflareZeroTrustAccessServiceTokenSpec{
				ZoneId: zoneRef(),
				Name:   "zone-ci-deployer",
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a Go-style duration and the forever value", func() {
			input := validToken(&CloudflareZeroTrustAccessServiceTokenSpec{
				AccountId: testAccountId,
				Name:      "ci-deployer",
				Duration:  "2h45m",
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())

			input.Spec.Duration = "forever"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept the rotation pair set together", func() {
			input := validToken(&CloudflareZeroTrustAccessServiceTokenSpec{
				AccountId:                     testAccountId,
				Name:                          "ci-deployer",
				ClientSecretVersion:           int32Ptr(2),
				PreviousClientSecretExpiresAt: "2026-09-01T00:00:00Z",
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a token with no scope", func() {
			input := validToken(&CloudflareZeroTrustAccessServiceTokenSpec{
				Name: "ci-deployer",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a token with both scopes", func() {
			input := validToken(&CloudflareZeroTrustAccessServiceTokenSpec{
				AccountId: testAccountId,
				ZoneId:    zoneRef(),
				Name:      "ci-deployer",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed duration", func() {
			input := validToken(&CloudflareZeroTrustAccessServiceTokenSpec{
				AccountId: testAccountId,
				Name:      "ci-deployer",
				Duration:  "1 year",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a rotation version without the expiry", func() {
			input := validToken(&CloudflareZeroTrustAccessServiceTokenSpec{
				AccountId:           testAccountId,
				Name:                "ci-deployer",
				ClientSecretVersion: int32Ptr(2),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a rotation expiry without the version", func() {
			input := validToken(&CloudflareZeroTrustAccessServiceTokenSpec{
				AccountId:                     testAccountId,
				Name:                          "ci-deployer",
				PreviousClientSecretExpiresAt: "2026-09-01T00:00:00Z",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a non-RFC3339 rotation expiry", func() {
			input := validToken(&CloudflareZeroTrustAccessServiceTokenSpec{
				AccountId:                     testAccountId,
				Name:                          "ci-deployer",
				ClientSecretVersion:           int32Ptr(2),
				PreviousClientSecretExpiresAt: "next tuesday",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
