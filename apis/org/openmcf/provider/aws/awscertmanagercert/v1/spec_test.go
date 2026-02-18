package awscertmanagercertv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAwsCertManagerCert(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsCertManagerCert Suite")
}

var _ = ginkgo.Describe("AwsCertManagerCert", func() {

	var input *AwsCertManagerCert

	ginkgo.BeforeEach(func() {
		input = &AwsCertManagerCert{
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCertManagerCert",
			Metadata: &shared.CloudResourceMetadata{
				Name: "a-test-name",
			},
			Spec: &AwsCertManagerCertSpec{
				Region:            "us-west-2",
				PrimaryDomainName: "example.com",
				AlternateDomainNames: []string{
					"www.example.com",
					"test.example.com",
				},
				ValidationMethod: proto.String("DNS"),
				Route53HostedZoneId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-zone-id"},
				},
			},
		}
	})

	ginkgo.Context("when valid input is passed", func() {
		ginkgo.It("should not return a validation error", func() {
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Domain Pattern Validations", func() {

		ginkgo.Context("PrimaryDomainName", func() {
			ginkgo.It("should accept a valid apex domain", func() {
				input.Spec.PrimaryDomainName = "example.com"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a valid wildcard domain", func() {
				input.Spec.PrimaryDomainName = "*.example.com"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should reject a domain missing a TLD", func() {
				input.Spec.PrimaryDomainName = "example"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject multiple wildcard asterisks", func() {
				input.Spec.PrimaryDomainName = "**.example.com"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a domain with invalid characters", func() {
				input.Spec.PrimaryDomainName = "exa@mple.com"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("AlternateDomainNames", func() {
			ginkgo.It("should accept multiple valid domains", func() {
				input.Spec.AlternateDomainNames = []string{"www.example.com", "*.foo.org"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should reject if any domain is invalid", func() {
				input.Spec.AlternateDomainNames = []string{"www.example.com", "invalid@@domain"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
