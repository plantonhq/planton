package azuremongoclusterv1alpha1

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

func TestAzureMongoClusterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMongoClusterSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const testClusterId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.DocumentDB/mongoClusters/source-cluster"

const testIdentityId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/app-uai"

const testKeyId = "https://app-vault.vault.azure.net/keys/mongo-cmk"

// validResource returns a valid Default-mode cluster that individual
// cases mutate into the shape under test.
func validResource() *AzureMongoCluster {
	return &AzureMongoCluster{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureMongoCluster",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-mongo",
		},
		Spec: &AzureMongoClusterSpec{
			ResourceGroup:         literal("app-rg"),
			Name:                  "acme-orders-db",
			Region:                "eastus",
			AdministratorUsername: "mongoadmin",
			AdministratorPassword: literal("S3cure-Passw0rd"),
			Version:               proto.String("8.0"),
			ComputeTier:           proto.String("M30"),
			StorageSizeInGb:       proto.Int32(128),
			ShardCount:            proto.Int32(1),
			HighAvailabilityMode:  proto.String("ZoneRedundantPreferred"),
		},
	}
}

// geoReplicaResource returns a valid GeoReplica-mode cluster.
func geoReplicaResource() *AzureMongoCluster {
	input := validResource()
	input.Spec = &AzureMongoClusterSpec{
		ResourceGroup:  literal("app-rg"),
		Name:           "acme-orders-replica",
		Region:         "westus3",
		CreateMode:     proto.String("GeoReplica"),
		SourceServerId: literal(testClusterId),
		SourceLocation: "eastus",
	}
	return input
}

var _ = ginkgo.Describe("AzureMongoClusterSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_mongo_cluster", func() {

			ginkgo.It("should not return a validation error for the Default-mode canonical shape", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an explicit Default create_mode", func() {
				input := validResource()
				input.Spec.CreateMode = proto.String("Default")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a GeoReplica-mode cluster without sizing fields", func() {
				err := protovalidate.Validate(geoReplicaResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a PointInTimeRestore-mode cluster with the restore block", func() {
				input := validResource()
				input.Spec = &AzureMongoClusterSpec{
					ResourceGroup: literal("app-rg"),
					Name:          "acme-orders-clone",
					Region:        "eastus",
					CreateMode:    proto.String("PointInTimeRestore"),
					Restore: &AzureMongoClusterRestore{
						PointInTimeUtc: "2026-08-01T12:00:00Z",
						SourceId:       literal(testClusterId),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the Free tier with one shard and HA disabled", func() {
				input := validResource()
				input.Spec.ComputeTier = proto.String("Free")
				input.Spec.HighAvailabilityMode = proto.String("Disabled")
				input.Spec.ShardCount = proto.Int32(1)
				input.Spec.StorageSizeInGb = proto.Int32(32)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept authentication methods including Entra", func() {
				input := validResource()
				input.Spec.AuthenticationMethods = []string{"NativeAuth", "MicrosoftEntraID"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a customer-managed key with the identity attached", func() {
				input := validResource()
				input.Spec.UserAssignedIdentityIds = []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)}
				input.Spec.CustomerManagedKey = &AzureMongoClusterCustomerManagedKey{
					KeyVaultKeyId:          literal(testKeyId),
					UserAssignedIdentityId: literal(testIdentityId),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept firewall rules with unique names", func() {
				input := validResource()
				input.Spec.FirewallRules = []*AzureMongoClusterFirewallRule{
					{Name: "office", StartIpAddress: "203.0.113.0", EndIpAddress: "203.0.113.255"},
					{Name: "vpn-egress", StartIpAddress: "198.51.100.7", EndIpAddress: "198.51.100.7"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept preview features on a Default-mode source", func() {
				input := validResource()
				input.Spec.PreviewFeatures = []string{"GeoReplicas"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the Data API on a Default-mode cluster", func() {
				input := validResource()
				input.Spec.DataApiModeEnabled = proto.Bool(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_mongo_cluster", func() {

			ginkgo.It("should reject a missing resource group", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject names violating the service format", func() {
				input := validResource()
				input.Spec.Name = "Acme-Orders"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Name = "ab"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Name = "a" + strings.Repeat("b", 40)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Name = "-orders"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown create_mode", func() {
				input := validResource()
				input.Spec.CreateMode = proto.String("Replica")
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a Default-mode cluster missing each sizing requirement", func() {
				input := validResource()
				input.Spec.AdministratorUsername = ""
				input.Spec.AdministratorPassword = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.Version = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.ComputeTier = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.StorageSizeInGb = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.ShardCount = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.HighAvailabilityMode = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a username without a password and a password without a username", func() {
				input := validResource()
				input.Spec.AdministratorPassword = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = geoReplicaResource()
				input.Spec.AdministratorPassword = literal("S3cure-Passw0rd")
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a GeoReplica-mode cluster missing its source coordinates", func() {
				input := geoReplicaResource()
				input.Spec.SourceServerId = nil
				input.Spec.SourceLocation = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = geoReplicaResource()
				input.Spec.SourceLocation = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject source_location without source_server_id", func() {
				input := validResource()
				input.Spec.SourceLocation = "eastus"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a PointInTimeRestore-mode cluster without the restore block", func() {
				input := validResource()
				input.Spec = &AzureMongoClusterSpec{
					ResourceGroup: literal("app-rg"),
					Name:          "acme-orders-clone",
					Region:        "eastus",
					CreateMode:    proto.String("PointInTimeRestore"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a malformed restore timestamp", func() {
				input := validResource()
				input.Spec = &AzureMongoClusterSpec{
					ResourceGroup: literal("app-rg"),
					Name:          "acme-orders-clone",
					Region:        "eastus",
					CreateMode:    proto.String("PointInTimeRestore"),
					Restore: &AzureMongoClusterRestore{
						PointInTimeUtc: "2026-08-01 12:00:00",
						SourceId:       literal(testClusterId),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject the Free and M25 tiers with zone-redundant HA or extra shards", func() {
				input := validResource()
				input.Spec.ComputeTier = proto.String("Free")
				input.Spec.HighAvailabilityMode = proto.String("ZoneRedundantPreferred")
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.ComputeTier = proto.String("M25")
				input.Spec.HighAvailabilityMode = proto.String("Disabled")
				input.Spec.ShardCount = proto.Int32(2)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject the Free tier with MicrosoftEntraID authentication", func() {
				input := validResource()
				input.Spec.ComputeTier = proto.String("Free")
				input.Spec.HighAvailabilityMode = proto.String("Disabled")
				input.Spec.ShardCount = proto.Int32(1)
				input.Spec.StorageSizeInGb = proto.Int32(32)
				input.Spec.AuthenticationMethods = []string{"NativeAuth", "MicrosoftEntraID"}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				// The smallest paid tier lifts the restriction.
				input.Spec.ComputeTier = proto.String("M10")
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should reject the Data API on a non-Default-mode cluster", func() {
				input := geoReplicaResource()
				input.Spec.DataApiModeEnabled = proto.Bool(false)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject storage size outside 32-32768", func() {
				input := validResource()
				input.Spec.StorageSizeInGb = proto.Int32(16)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.StorageSizeInGb = proto.Int32(65536)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject storage_type without storage_size_in_gb", func() {
				input := geoReplicaResource()
				input.Spec.StorageType = proto.String("PremiumSSDv2")
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown authentication method and duplicates", func() {
				input := validResource()
				input.Spec.AuthenticationMethods = []string{"Kerberos"}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.AuthenticationMethods = []string{"NativeAuth", "NativeAuth"}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a customer-managed key without any attached identity", func() {
				input := validResource()
				input.Spec.CustomerManagedKey = &AzureMongoClusterCustomerManagedKey{
					KeyVaultKeyId:          literal(testKeyId),
					UserAssignedIdentityId: literal(testIdentityId),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject duplicate preview features", func() {
				input := validResource()
				input.Spec.PreviewFeatures = []string{"GeoReplicas", "GeoReplicas"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject duplicate firewall rule names and malformed addresses", func() {
				input := validResource()
				input.Spec.FirewallRules = []*AzureMongoClusterFirewallRule{
					{Name: "office", StartIpAddress: "203.0.113.0", EndIpAddress: "203.0.113.255"},
					{Name: "office", StartIpAddress: "198.51.100.7", EndIpAddress: "198.51.100.7"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.FirewallRules = []*AzureMongoClusterFirewallRule{
					{Name: "office", StartIpAddress: "203.0.113.0/24", EndIpAddress: "203.0.113.255"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an out-of-vocabulary version, tier, storage type, and HA mode", func() {
				input := validResource()
				input.Spec.Version = proto.String("4.2")
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.ComputeTier = proto.String("M300")
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.StorageType = proto.String("StandardSSD")
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.HighAvailabilityMode = proto.String("SameZone")
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
