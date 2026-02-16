package awscloudfrontv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAwsCloudFrontSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsCloudFrontSpec Validation Suite")
}

var _ = ginkgo.Describe("AwsCloudFrontSpec validations", func() {
	var spec *AwsCloudFrontSpec

	ginkgo.BeforeEach(func() {
		spec = &AwsCloudFrontSpec{
			Enabled: true,
			Aliases: []string{"cdn.example.com"},
			CertificateArn: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
				},
			},
			PriceClass:     AwsCloudFrontSpec_PRICE_CLASS_100,
			Origins: []*AwsCloudFrontSpec_Origin{
				{
					DomainName: "my-bucket.s3.amazonaws.com",
					OriginPath: "/assets",
					IsDefault:  true,
				},
				{
					DomainName: "api.example.com",
					OriginPath: "",
					IsDefault:  false,
				},
			},
			DefaultRootObject: "index.html",
		}
	})

	ginkgo.It("accepts a valid CloudFront spec", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a valid spec without aliases and certificate", func() {
		spec.Aliases = []string{}
		spec.CertificateArn = nil
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a valid spec with single origin", func() {
		spec.Origins = []*AwsCloudFrontSpec_Origin{
			{
				DomainName: "my-bucket.s3.amazonaws.com",
				OriginPath: "",
				IsDefault:  true,
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// Field validation tests
	ginkgo.It("fails when aliases contain invalid domain names", func() {
		spec.Aliases = []string{"invalid-domain"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when aliases contain duplicate values", func() {
		spec.Aliases = []string{"cdn.example.com", "cdn.example.com"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a valid spec with certificate_arn as StringValueOrRef", func() {
		spec.CertificateArn = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: "arn:aws:acm:us-east-1:123456789012:certificate/abcdef01-2345-6789-abcd-ef0123456789",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when price_class is undefined", func() {
		spec.PriceClass = AwsCloudFrontSpec_PriceClass(99)
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when origins is empty", func() {
		spec.Origins = []*AwsCloudFrontSpec_Origin{}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when origin domain_name is empty", func() {
		spec.Origins[0].DomainName = ""
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when origin domain_name has invalid format", func() {
		spec.Origins[0].DomainName = "invalid-domain"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when origin_path doesn't start with /", func() {
		spec.Origins[0].OriginPath = "assets"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when default_root_object has invalid characters", func() {
		spec.DefaultRootObject = "index.html!"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// CEL expression tests
	ginkgo.It("fails when aliases are provided but certificate_arn is nil (aliases_require_cert)", func() {
		spec.Aliases = []string{"cdn.example.com"}
		spec.CertificateArn = nil
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("passes when aliases are empty and certificate_arn is nil (aliases_require_cert)", func() {
		spec.Aliases = []string{}
		spec.CertificateArn = nil
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("passes when aliases are provided and certificate_arn is set (aliases_require_cert)", func() {
		spec.Aliases = []string{"cdn.example.com"}
		spec.CertificateArn = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when no origin is marked as default (exactly_one_default_origin)", func() {
		spec.Origins[0].IsDefault = false
		spec.Origins[1].IsDefault = false
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when multiple origins are marked as default (exactly_one_default_origin)", func() {
		spec.Origins[0].IsDefault = true
		spec.Origins[1].IsDefault = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("passes when exactly one origin is marked as default (exactly_one_default_origin)", func() {
		spec.Origins[0].IsDefault = true
		spec.Origins[1].IsDefault = false
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("passes when exactly one origin is marked as default in different position (exactly_one_default_origin)", func() {
		spec.Origins[0].IsDefault = false
		spec.Origins[1].IsDefault = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// Edge cases
	ginkgo.It("accepts valid domain names with hyphens", func() {
		spec.Aliases = []string{"my-cdn.example.com"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts valid domain names with numbers", func() {
		spec.Aliases = []string{"cdn1.example.com"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts valid origin paths with underscores and hyphens", func() {
		spec.Origins[0].OriginPath = "/assets/images"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts empty origin_path", func() {
		spec.Origins[0].OriginPath = ""
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts valid default_root_object with underscores", func() {
		spec.DefaultRootObject = "index_main.html"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts empty default_root_object", func() {
		spec.DefaultRootObject = ""
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})
})
