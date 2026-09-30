package azurefrontdoororiginv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureFrontDoorOriginSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureFrontDoorOriginSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// ref builds a StringValueOrRef carrying a value_from reference.
func ref(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

const originGroupId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Cdn/profiles/planton-fd/originGroups/api-backends"
const appServiceId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Web/sites/my-app"
const plsId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/privateLinkServices/my-pls"

// minimal valid spec: a public origin with certificate checking on (the
// documented default).
func minimalSpec() *AzureFrontDoorOrigin {
	return &AzureFrontDoorOrigin{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureFrontDoorOrigin",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-front-door-origin",
		},
		Spec: &AzureFrontDoorOriginSpec{
			OriginGroupId: literal(originGroupId),
			OriginName:    "test-origin",
			HostName:      literal("myapp.azurewebsites.net"),
		},
	}
}

func validPrivateLink() *AzureFrontDoorOriginPrivateLink {
	return &AzureFrontDoorOriginPrivateLink{
		Location:            "eastus",
		PrivateLinkTargetId: literal(appServiceId),
		TargetType:          AzureFrontDoorOriginPrivateLinkTargetType_SITES,
	}
}

var _ = ginkgo.Describe("AzureFrontDoorOriginSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal origin", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept a host name resolved by reference", func() {
			input := minimalSpec()
			input.Spec.HostName = ref("my-web-app")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept all optional dials at their boundaries", func() {
			input := minimalSpec()
			input.Spec.OriginHostHeader = literal("myapp.azurewebsites.net")
			input.Spec.HttpPort = proto.Int32(1)
			input.Spec.HttpsPort = proto.Int32(65535)
			input.Spec.Priority = proto.Int32(5)
			input.Spec.Weight = proto.Int32(1000)
			input.Spec.Enabled = proto.Bool(false)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a private link with a target type", func() {
			input := minimalSpec()
			input.Spec.PrivateLink = validPrivateLink()
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a private link to a Private Link Service without a target type", func() {
			input := minimalSpec()
			input.Spec.PrivateLink = &AzureFrontDoorOriginPrivateLink{
				Location:            "eastus",
				PrivateLinkTargetId: literal(plsId),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept every private-link target type", func() {
			for _, targetType := range []AzureFrontDoorOriginPrivateLinkTargetType{
				AzureFrontDoorOriginPrivateLinkTargetType_SITES,
				AzureFrontDoorOriginPrivateLinkTargetType_BLOB,
				AzureFrontDoorOriginPrivateLinkTargetType_BLOB_SECONDARY,
				AzureFrontDoorOriginPrivateLinkTargetType_WEB,
				AzureFrontDoorOriginPrivateLinkTargetType_WEB_SECONDARY,
				AzureFrontDoorOriginPrivateLinkTargetType_MANAGED_ENVIRONMENTS,
				AzureFrontDoorOriginPrivateLinkTargetType_GATEWAY,
			} {
				input := minimalSpec()
				privateLink := validPrivateLink()
				privateLink.TargetType = targetType
				input.Spec.PrivateLink = privateLink
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "target type %v must be accepted", targetType)
			}
		})

		ginkgo.It("should accept a private link with an explicit certificate check", func() {
			input := minimalSpec()
			input.Spec.CertificateNameCheckEnabled = proto.Bool(true)
			input.Spec.PrivateLink = validPrivateLink()
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a 140-character private-link request message", func() {
			input := minimalSpec()
			privateLink := validPrivateLink()
			privateLink.RequestMessage = proto.String(strings.Repeat("m", 140))
			input.Spec.PrivateLink = privateLink
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept origin name boundaries (2 and 90 characters)", func() {
			input := minimalSpec()
			input.Spec.OriginName = "ab"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			input.Spec.OriginName = "a" + strings.Repeat("b", 88) + "c"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing origin group reference", func() {
			input := minimalSpec()
			input.Spec.OriginGroupId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing origin name", func() {
			input := minimalSpec()
			input.Spec.OriginName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an origin name over 90 characters", func() {
			input := minimalSpec()
			input.Spec.OriginName = "a" + strings.Repeat("b", 89) + "c"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an origin name with a leading hyphen", func() {
			input := minimalSpec()
			input.Spec.OriginName = "-origin"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing host name", func() {
			input := minimalSpec()
			input.Spec.HostName = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject http port 0 and 65536", func() {
			for _, port := range []int32{0, 65536} {
				input := minimalSpec()
				input.Spec.HttpPort = proto.Int32(port)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "port %d must be rejected", port)
			}
		})

		ginkgo.It("should reject https port 0 and 65536", func() {
			for _, port := range []int32{0, 65536} {
				input := minimalSpec()
				input.Spec.HttpsPort = proto.Int32(port)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "port %d must be rejected", port)
			}
		})

		ginkgo.It("should reject priority 0 and 6", func() {
			for _, priority := range []int32{0, 6} {
				input := minimalSpec()
				input.Spec.Priority = proto.Int32(priority)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "priority %d must be rejected", priority)
			}
		})

		ginkgo.It("should reject weight 0 and 1001", func() {
			for _, weight := range []int32{0, 1001} {
				input := minimalSpec()
				input.Spec.Weight = proto.Int32(weight)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "weight %d must be rejected", weight)
			}
		})

		ginkgo.It("should reject a private link with certificate checking disabled", func() {
			input := minimalSpec()
			input.Spec.CertificateNameCheckEnabled = proto.Bool(false)
			input.Spec.PrivateLink = validPrivateLink()
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a private link without a location", func() {
			input := minimalSpec()
			privateLink := validPrivateLink()
			privateLink.Location = ""
			input.Spec.PrivateLink = privateLink
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept a Private Link Service reference without a target type", func() {
			input := minimalSpec()
			input.Spec.PrivateLink = &AzureFrontDoorOriginPrivateLink{
				Location: "eastus",
				PrivateLinkTargetId: &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{Kind: cloudresourcekind.CloudResourceKind_AzurePrivateLinkService, Name: "internal-api-pls"},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should reject a web app reference without a target type", func() {
			input := minimalSpec()
			input.Spec.PrivateLink = &AzureFrontDoorOriginPrivateLink{
				Location: "eastus",
				PrivateLinkTargetId: &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{Kind: cloudresourcekind.CloudResourceKind_AzureLinuxWebApp, Name: "api"},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a private-link target that is not an ARM id", func() {
			input := minimalSpec()
			privateLink := validPrivateLink()
			privateLink.PrivateLinkTargetId = literal("my-app")
			input.Spec.PrivateLink = privateLink
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a non-PLS private-link target without a target type", func() {
			input := minimalSpec()
			input.Spec.PrivateLink = &AzureFrontDoorOriginPrivateLink{
				Location:            "eastus",
				PrivateLinkTargetId: literal(appServiceId),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a private-link request message over 140 characters", func() {
			input := minimalSpec()
			privateLink := validPrivateLink()
			privateLink.RequestMessage = proto.String(strings.Repeat("m", 141))
			input.Spec.PrivateLink = privateLink
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an undefined private-link target type", func() {
			input := minimalSpec()
			privateLink := validPrivateLink()
			privateLink.TargetType = AzureFrontDoorOriginPrivateLinkTargetType(99)
			input.Spec.PrivateLink = privateLink
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a wrong kind", func() {
			input := minimalSpec()
			input.Kind = "AzureFrontDoorOrigins"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject missing metadata", func() {
			input := minimalSpec()
			input.Metadata = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
