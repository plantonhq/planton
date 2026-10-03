package azureprivateendpointv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzurePrivateEndpointSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzurePrivateEndpointSpec Validation Tests")
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

// validResource returns a minimal valid AzurePrivateEndpoint that individual
// cases then mutate into the shape under test.
func validResource() *AzurePrivateEndpoint {
	return &AzurePrivateEndpoint{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzurePrivateEndpoint",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "pg-pe",
		},
		Spec: &AzurePrivateEndpointSpec{
			Region:        "eastus",
			ResourceGroup: literal("my-rg"),
			Name:          "pg-private-endpoint",
			SubnetId:      literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/pe-subnet"),
			PrivateServiceConnection: &AzurePrivateEndpointServiceConnection{
				PrivateConnectionResourceId: literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/test-pg"),
				SubresourceNames:            []string{"postgresqlServer"},
			},
		},
	}
}

var _ = ginkgo.Describe("AzurePrivateEndpointSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_private_endpoint", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept references for subnet, resource group, and connection target", func() {
				input := validResource()
				input.Spec.ResourceGroup = ref("network-rg")
				input.Spec.SubnetId = ref("pe-subnet")
				input.Spec.PrivateServiceConnection.PrivateConnectionResourceId = ref("prod-postgres")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a connection by alias instead of resource id", func() {
				input := validResource()
				input.Spec.PrivateServiceConnection.PrivateConnectionResourceId = nil
				input.Spec.PrivateServiceConnection.ConnectionAlias = "partner-pls.d20286c8-4ea5-11eb-9584.centralus.azure.privatelinkservice"
				input.Spec.PrivateServiceConnection.SubresourceNames = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a manual connection with a request message", func() {
				input := validResource()
				manual := true
				input.Spec.PrivateServiceConnection.IsManualConnection = &manual
				input.Spec.PrivateServiceConnection.RequestMessage = "please approve cross-tenant access"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept private DNS zone references", func() {
				input := validResource()
				input.Spec.PrivateDnsZoneIds = []*foreignkeyv1.StringValueOrRef{ref("postgres-privatelink-zone")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept static ip configurations", func() {
				input := validResource()
				input.Spec.IpConfigurations = []*AzurePrivateEndpointIpConfiguration{
					{Name: "pg-ip", PrivateIpAddress: "10.0.1.10", SubresourceName: "postgresqlServer"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept application security group references", func() {
				input := validResource()
				input.Spec.ApplicationSecurityGroupIds = []*foreignkeyv1.StringValueOrRef{ref("data-tier")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_private_endpoint", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when subnet is missing", func() {
				input := validResource()
				input.Spec.SubnetId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the connection block is missing", func() {
				input := validResource()
				input.Spec.PrivateServiceConnection = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when both resource id and alias are set", func() {
				input := validResource()
				input.Spec.PrivateServiceConnection.ConnectionAlias = "some-alias"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when neither resource id nor alias is set", func() {
				input := validResource()
				input.Spec.PrivateServiceConnection.PrivateConnectionResourceId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the alias lacks the privatelinkservice suffix", func() {
				input := validResource()
				input.Spec.PrivateServiceConnection.PrivateConnectionResourceId = nil
				input.Spec.PrivateServiceConnection.ConnectionAlias = "partner-pls.centralus.example.com"
				input.Spec.PrivateServiceConnection.SubresourceNames = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a manual connection has no request message", func() {
				input := validResource()
				manual := true
				input.Spec.PrivateServiceConnection.IsManualConnection = &manual
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a request message is set on an auto-approved connection", func() {
				input := validResource()
				input.Spec.PrivateServiceConnection.RequestMessage = "should not be here"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when an ip configuration has no private ip", func() {
				input := validResource()
				input.Spec.IpConfigurations = []*AzurePrivateEndpointIpConfiguration{
					{Name: "pg-ip"},
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

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := validResource()
				input.ApiVersion = "wrong.version/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := validResource()
				input.Kind = "WrongKind"
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
