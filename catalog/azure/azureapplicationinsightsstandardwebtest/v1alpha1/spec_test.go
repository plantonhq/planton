package azureapplicationinsightsstandardwebtestv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureApplicationInsightsStandardWebTestSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureApplicationInsightsStandardWebTestSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func ref(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

func validResource() *AzureApplicationInsightsStandardWebTest {
	return &AzureApplicationInsightsStandardWebTest{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureApplicationInsightsStandardWebTest",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-webtest",
		},
		Spec: &AzureApplicationInsightsStandardWebTestSpec{
			ResourceGroup:         literal("test-rg"),
			Name:                  "test-webtest",
			Region:                "eastus",
			ApplicationInsightsId: ref("app-insights"),
			Request: &AzureApplicationInsightsStandardWebTestRequest{
				Url: "https://example.com/health",
			},
			GeoLocations: []string{"us-east-1-azr"},
		},
	}
}

var _ = ginkgo.Describe("AzureApplicationInsightsStandardWebTestSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_application_insights_standard_web_test", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept frequency, timeout, and retry settings", func() {
				input := validResource()
				freq := int32(600)
				timeout := int32(60)
				retry := true
				input.Spec.Frequency = &freq
				input.Spec.Timeout = &timeout
				input.Spec.RetryEnabled = &retry
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a POST request with headers and body", func() {
				input := validResource()
				input.Spec.Request = &AzureApplicationInsightsStandardWebTestRequest{
					Url:      "https://example.com/api",
					HttpVerb: "POST",
					Body:     "{\"ping\":true}",
					Headers: []*AzureApplicationInsightsStandardWebTestHeader{
						{Name: "Content-Type", Value: "application/json"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept validation rules with content and ssl checks", func() {
				input := validResource()
				code := int32(204)
				life := int32(30)
				sslCheck := true
				input.Spec.ValidationRules = &AzureApplicationInsightsStandardWebTestValidationRules{
					ExpectedStatusCode:       &code,
					SslCheckEnabled:          &sslCheck,
					SslCertRemainingLifetime: &life,
					Content: &AzureApplicationInsightsStandardWebTestContentValidation{
						ContentMatch: "healthy",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept multiple geo locations", func() {
				input := validResource()
				input.Spec.GeoLocations = []string{"us-east-1-azr", "us-west-2-azr", "emea-nl-ams-azr"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_application_insights_standard_web_test", func() {

			ginkgo.It("should return a validation error when the request is missing", func() {
				input := validResource()
				input.Spec.Request = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the request url is missing", func() {
				input := validResource()
				input.Spec.Request = &AzureApplicationInsightsStandardWebTestRequest{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when geo_locations is empty", func() {
				input := validResource()
				input.Spec.GeoLocations = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when application_insights_id is missing", func() {
				input := validResource()
				input.Spec.ApplicationInsightsId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid frequency", func() {
				input := validResource()
				freq := int32(120)
				input.Spec.Frequency = &freq
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid http verb", func() {
				input := validResource()
				input.Spec.Request.HttpVerb = "FETCH"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an out-of-range ssl lifetime", func() {
				input := validResource()
				life := int32(400)
				sslCheck := true
				input.Spec.ValidationRules = &AzureApplicationInsightsStandardWebTestValidationRules{
					SslCheckEnabled:          &sslCheck,
					SslCertRemainingLifetime: &life,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the ssl lifetime is set without the ssl check", func() {
				input := validResource()
				life := int32(30)
				input.Spec.ValidationRules = &AzureApplicationInsightsStandardWebTestValidationRules{
					SslCertRemainingLifetime: &life,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the ssl check is enabled on an http url", func() {
				input := validResource()
				input.Spec.Request.Url = "http://example.com/health"
				sslCheck := true
				input.Spec.ValidationRules = &AzureApplicationInsightsStandardWebTestValidationRules{
					SslCheckEnabled: &sslCheck,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should accept an http url when no ssl checks are requested", func() {
				input := validResource()
				input.Spec.Request.Url = "http://internal.example.com/health"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when content_match is empty", func() {
				input := validResource()
				input.Spec.ValidationRules = &AzureApplicationInsightsStandardWebTestValidationRules{
					Content: &AzureApplicationInsightsStandardWebTestContentValidation{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a header value is missing", func() {
				input := validResource()
				input.Spec.Request.Headers = []*AzureApplicationInsightsStandardWebTestHeader{
					{Name: "X-Custom"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := validResource()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
