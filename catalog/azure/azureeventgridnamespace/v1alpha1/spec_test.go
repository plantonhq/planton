package azureeventgridnamespacev1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureEventgridNamespaceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureEventgridNamespaceSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const testIdentityId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/app-uai"

const testTopicId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.EventGrid/topics/app-topic"

// validResource returns a valid namespace that individual cases
// mutate into the shape under test.
func validResource() *AzureEventgridNamespace {
	return &AzureEventgridNamespace{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureEventgridNamespace",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-egns",
		},
		Spec: &AzureEventgridNamespaceSpec{
			ResourceGroup: literal("app-rg"),
			Name:          "acme-events",
			Region:        "eastus",
		},
	}
}

var _ = ginkgo.Describe("AzureEventgridNamespaceSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_eventgrid_namespace", func() {

			ginkgo.It("should not return a validation error for the minimal shape", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept explicit capacity within 1-40", func() {
				input := validResource()
				input.Spec.Capacity = proto.Int32(40)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept inbound IP rules up to the cap", func() {
				input := validResource()
				input.Spec.InboundIpRules = []string{"203.0.113.0/24", "198.51.100.7"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a 3-character and a 50-character name", func() {
				input := validResource()
				input.Spec.Name = "abc"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.Name = strings.Repeat("a", 50)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a system-assigned identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureEventgridNamespaceIdentity{
					Type: AzureEventgridNamespaceIdentityType_SYSTEM_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the combined identity mode carrying an identity id", func() {
				input := validResource()
				input.Spec.Identity = &AzureEventgridNamespaceIdentity{
					Type:        AzureEventgridNamespaceIdentityType_SYSTEM_AND_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a full MQTT topic-spaces block", func() {
				input := validResource()
				input.Spec.TopicSpacesConfiguration = &AzureEventgridNamespaceTopicSpacesConfiguration{
					AlternativeAuthenticationNameSources: []string{
						"ClientCertificateSubject",
						"ClientCertificateDns",
					},
					MaximumClientSessionsPerAuthenticationName: proto.Int32(5),
					MaximumSessionExpiryInHours:                proto.Int32(8),
					RouteTopicId:                               literal(testTopicId),
					DynamicRoutingEnrichments: []*AzureEventgridNamespaceRoutingEnrichment{
						{Key: "clientname", Value: "${client.authenticationName}"},
					},
					StaticRoutingEnrichments: []*AzureEventgridNamespaceRoutingEnrichment{
						{Key: "source", Value: "mqtt-broker"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an empty MQTT block (presence is the enable switch)", func() {
				input := validResource()
				input.Spec.TopicSpacesConfiguration = &AzureEventgridNamespaceTopicSpacesConfiguration{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_eventgrid_namespace", func() {

			ginkgo.It("should reject a missing resource group", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a 2-character and a 51-character name", func() {
				input := validResource()
				input.Spec.Name = "ab"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Name = strings.Repeat("a", 51)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing region", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject capacity outside 1-40", func() {
				input := validResource()
				input.Spec.Capacity = proto.Int32(0)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Capacity = proto.Int32(41)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty inbound IP rule entry", func() {
				input := validResource()
				input.Spec.InboundIpRules = []string{""}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an identity block without a flavor", func() {
				input := validResource()
				input.Spec.Identity = &AzureEventgridNamespaceIdentity{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a user-assigned identity without identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureEventgridNamespaceIdentity{
					Type: AzureEventgridNamespaceIdentityType_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a system-assigned identity carrying identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureEventgridNamespaceIdentity{
					Type:        AzureEventgridNamespaceIdentityType_SYSTEM_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown alternative authentication name source", func() {
				input := validResource()
				input.Spec.TopicSpacesConfiguration = &AzureEventgridNamespaceTopicSpacesConfiguration{
					AlternativeAuthenticationNameSources: []string{"ClientCertificateSerialNumber"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject MQTT session dials outside their ranges", func() {
				input := validResource()
				input.Spec.TopicSpacesConfiguration = &AzureEventgridNamespaceTopicSpacesConfiguration{
					MaximumClientSessionsPerAuthenticationName: proto.Int32(101),
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.TopicSpacesConfiguration = &AzureEventgridNamespaceTopicSpacesConfiguration{
					MaximumSessionExpiryInHours: proto.Int32(9),
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a routing enrichment with an oversize key or empty value", func() {
				input := validResource()
				input.Spec.TopicSpacesConfiguration = &AzureEventgridNamespaceTopicSpacesConfiguration{
					StaticRoutingEnrichments: []*AzureEventgridNamespaceRoutingEnrichment{
						{Key: strings.Repeat("k", 21), Value: "v"},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.TopicSpacesConfiguration = &AzureEventgridNamespaceTopicSpacesConfiguration{
					StaticRoutingEnrichments: []*AzureEventgridNamespaceRoutingEnrichment{
						{Key: "k", Value: ""},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
