package azuresearchservicev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureSearchServiceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureSearchServiceSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testStorageId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/searchdata"
	testUaiId     = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/search-uai"
)

// validResource returns a minimal valid search service that
// individual cases mutate into the shape under test.
func validResource() *AzureSearchService {
	return &AzureSearchService{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureSearchService",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-search-service",
		},
		Spec: &AzureSearchServiceSpec{
			Region:        "eastus",
			ResourceGroup: literal("search-rg"),
			Name:          "acme-search",
			Sku:           "standard",
		},
	}
}

var _ = ginkgo.Describe("AzureSearchServiceSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_search_service", func() {

			ginkgo.It("should not return a validation error for a minimal standard service", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a free service with counts unset", func() {
				input := validResource()
				input.Spec.Sku = "free"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a basic service at its caps (3 replicas, 3 partitions)", func() {
				input := validResource()
				input.Spec.Sku = "basic"
				input.Spec.ReplicaCount = proto.Int32(3)
				input.Spec.PartitionCount = proto.Int32(3)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept high-density hosting on standard3 within its partition cap", func() {
				input := validResource()
				input.Spec.Sku = "standard3"
				input.Spec.HostingMode = AzureSearchServiceHostingMode_HIGH_DENSITY
				input.Spec.PartitionCount = proto.Int32(3)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept 12 partitions on a standard service in default hosting", func() {
				input := validResource()
				input.Spec.PartitionCount = proto.Int32(12)
				input.Spec.ReplicaCount = proto.Int32(12)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an authentication failure mode while local auth stays enabled", func() {
				input := validResource()
				input.Spec.AuthenticationFailureMode = "http403"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the RBAC-only posture (local auth off, no failure mode)", func() {
				input := validResource()
				localAuth := false
				input.Spec.LocalAuthenticationEnabled = &localAuth
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept semantic search on a paid sku", func() {
				input := validResource()
				input.Spec.SemanticSearchSku = "standard"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept allowed_ips as addresses and CIDR ranges", func() {
				input := validResource()
				input.Spec.AllowedIps = []string{"203.0.113.7", "203.0.113.0/24"}
				input.Spec.NetworkRuleBypassOption = "AzureServices"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a system-assigned identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureSearchServiceIdentity{
					Type: AzureSearchServiceIdentityType_SYSTEM_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a shared private link to a storage blob target", func() {
				input := validResource()
				input.Spec.SharedPrivateLinkServices = []*AzureSearchServiceSharedPrivateLink{
					{
						Name:             "to-blob",
						SubresourceName:  "blob",
						TargetResourceId: literal(testStorageId),
						RequestMessage:   "please approve",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_search_service", func() {

			ginkgo.It("should reject an unknown sku", func() {
				input := validResource()
				input.Spec.Sku = "premium"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a second partition on the free sku", func() {
				input := validResource()
				input.Spec.Sku = "free"
				input.Spec.PartitionCount = proto.Int32(2)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a second replica on the free sku", func() {
				input := validResource()
				input.Spec.Sku = "free"
				input.Spec.ReplicaCount = proto.Int32(2)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject more than 3 partitions on the basic sku", func() {
				input := validResource()
				input.Spec.Sku = "basic"
				input.Spec.PartitionCount = proto.Int32(4)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject more than 3 replicas on the basic sku", func() {
				input := validResource()
				input.Spec.Sku = "basic"
				input.Spec.ReplicaCount = proto.Int32(4)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a partition count outside the legal set", func() {
				input := validResource()
				input.Spec.PartitionCount = proto.Int32(5)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject more than 12 replicas", func() {
				input := validResource()
				input.Spec.ReplicaCount = proto.Int32(13)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject high-density hosting off the standard3 sku", func() {
				input := validResource()
				input.Spec.HostingMode = AzureSearchServiceHostingMode_HIGH_DENSITY
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject more than 3 partitions in high-density hosting", func() {
				input := validResource()
				input.Spec.Sku = "standard3"
				input.Spec.HostingMode = AzureSearchServiceHostingMode_HIGH_DENSITY
				input.Spec.PartitionCount = proto.Int32(6)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an authentication failure mode in the RBAC-only posture", func() {
				input := validResource()
				localAuth := false
				input.Spec.LocalAuthenticationEnabled = &localAuth
				input.Spec.AuthenticationFailureMode = "http403"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown authentication failure mode", func() {
				input := validResource()
				input.Spec.AuthenticationFailureMode = "http418"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject semantic search on the free sku", func() {
				input := validResource()
				input.Spec.Sku = "free"
				input.Spec.SemanticSearchSku = "free"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an allowed_ips entry that is not an IPv4 or CIDR", func() {
				input := validResource()
				input.Spec.AllowedIps = []string{"not-an-ip"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown network rule bypass option", func() {
				input := validResource()
				input.Spec.NetworkRuleBypassOption = "Everything"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject duplicate shared private link names", func() {
				input := validResource()
				input.Spec.SharedPrivateLinkServices = []*AzureSearchServiceSharedPrivateLink{
					{Name: "to-blob", SubresourceName: "blob", TargetResourceId: literal(testStorageId)},
					{Name: "to-blob", SubresourceName: "table", TargetResourceId: literal(testStorageId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a shared private link subresource name shorter than 3 characters", func() {
				input := validResource()
				input.Spec.SharedPrivateLinkServices = []*AzureSearchServiceSharedPrivateLink{
					{Name: "to-blob", SubresourceName: "ab", TargetResourceId: literal(testStorageId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a user-assigned identity without identity_ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureSearchServiceIdentity{
					Type: AzureSearchServiceIdentityType_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a service without a name", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
