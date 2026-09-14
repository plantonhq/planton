package gcpsubnetworkv1alpha1

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

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpSubnetworkSpec Suite")
}

func strPtr(s string) *string {
	return &s
}

var _ = ginkgo.Describe("GcpSubnetworkSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpSubnetwork.
	minimal := func() *GcpSubnetwork {
		return &GcpSubnetwork{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpSubnetwork",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-subnetwork",
			},
			Spec: &GcpSubnetworkSpec{
				VpcSelfLink: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "projects/test-project/global/networks/test-vpc",
					},
				},
				SubnetworkName: "test-subnet",
				Region:         "us-central1",
				IpCidrRange:    "10.10.0.0/20",
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept flow logs declared and switched off", func() {
		target := minimal()
		target.Spec.LogConfig = &GcpSubnetworkLogConfig{Enabled: proto.Bool(false), AggregationInterval: strPtr("INTERVAL_1_MIN")}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a GKE-shaped subnet with secondary ranges and flow logs", func() {
		target := minimal()
		target.Spec.Description = "gke workload subnet"
		target.Spec.SecondaryIpRanges = []*GcpSubnetworkSecondaryRange{
			{RangeName: "pods", IpCidrRange: "10.16.0.0/14"},
			{RangeName: "services", IpCidrRange: "10.20.0.0/20"},
		}
		target.Spec.PrivateIpGoogleAccess = true
		target.Spec.LogConfig = &GcpSubnetworkLogConfig{
			AggregationInterval: strPtr("INTERVAL_1_MIN"),
			Metadata:            strPtr("INCLUDE_ALL_METADATA"),
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a proxy-only subnet with an ACTIVE role", func() {
		target := minimal()
		target.Spec.Purpose = strPtr("REGIONAL_MANAGED_PROXY")
		target.Spec.Role = "ACTIVE"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a Private Service Connect subnet", func() {
		target := minimal()
		target.Spec.Purpose = strPtr("PRIVATE_SERVICE_CONNECT")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a dual-stack subnet with external IPv6", func() {
		target := minimal()
		target.Spec.StackType = strPtr("IPV4_IPV6")
		target.Spec.Ipv6AccessType = "EXTERNAL"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an IPv6-only subnet without an IPv4 range", func() {
		target := minimal()
		target.Spec.IpCidrRange = ""
		target.Spec.StackType = strPtr("IPV6_ONLY")
		target.Spec.Ipv6AccessType = "EXTERNAL"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a custom-metadata flow-log config with fields", func() {
		target := minimal()
		target.Spec.LogConfig = &GcpSubnetworkLogConfig{
			Metadata:       strPtr("CUSTOM_METADATA"),
			MetadataFields: []string{"src_instance", "dest_vpc"},
			FilterExpr:     "connection.dest_port == 443",
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a spec without vpc_self_link", func() {
		target := minimal()
		target.Spec.VpcSelfLink = nil
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing subnetwork_name", func() {
		target := minimal()
		target.Spec.SubnetworkName = ""
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid subnetwork_name", func() {
		target := minimal()
		target.Spec.SubnetworkName = "INVALID-NAME"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing ip_cidr_range on an IPv4 subnet", func() {
		target := minimal()
		target.Spec.IpCidrRange = ""
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "IPV6_ONLY")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an ip_cidr_range on an IPV6_ONLY subnet", func() {
		target := minimal()
		target.Spec.StackType = strPtr("IPV6_ONLY")
		target.Spec.Ipv6AccessType = "EXTERNAL"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid purpose", func() {
		target := minimal()
		target.Spec.Purpose = strPtr("PUBLIC")
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a role on a PRIVATE subnet", func() {
		target := minimal()
		target.Spec.Role = "ACTIVE"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "REGIONAL_MANAGED_PROXY")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an invalid role value", func() {
		target := minimal()
		target.Spec.Purpose = strPtr("REGIONAL_MANAGED_PROXY")
		target.Spec.Role = "STANDBY"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a dual-stack subnet without ipv6_access_type", func() {
		target := minimal()
		target.Spec.StackType = strPtr("IPV4_IPV6")
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "ipv6_access_type")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject ipv6_access_type on an IPV4_ONLY subnet", func() {
		target := minimal()
		target.Spec.Ipv6AccessType = "EXTERNAL"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject external_ipv6_prefix without EXTERNAL access", func() {
		target := minimal()
		target.Spec.StackType = strPtr("IPV4_IPV6")
		target.Spec.Ipv6AccessType = "INTERNAL"
		target.Spec.ExternalIpv6Prefix = "2600:1901:0:1234::/64"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid private_ipv6_google_access", func() {
		target := minimal()
		target.Spec.PrivateIpv6GoogleAccess = "ENABLE_EVERYTHING"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a secondary range with an invalid name", func() {
		target := minimal()
		target.Spec.SecondaryIpRanges = []*GcpSubnetworkSecondaryRange{
			{RangeName: "Pods_Range", IpCidrRange: "10.16.0.0/14"},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid flow-log aggregation interval", func() {
		target := minimal()
		target.Spec.LogConfig = &GcpSubnetworkLogConfig{
			AggregationInterval: strPtr("INTERVAL_2_MIN"),
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a flow_sampling above 1", func() {
		target := minimal()
		sampling := 1.5
		target.Spec.LogConfig = &GcpSubnetworkLogConfig{FlowSampling: &sampling}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject metadata_fields without CUSTOM_METADATA", func() {
		target := minimal()
		target.Spec.LogConfig = &GcpSubnetworkLogConfig{
			MetadataFields: []string{"src_instance"},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "CUSTOM_METADATA")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a wrong kind constant", func() {
		target := minimal()
		target.Kind = "GcpSubnet"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept reserved_internal_range in place of ip_cidr_range", func() {
		target := minimal()
		target.Spec.IpCidrRange = ""
		target.Spec.ReservedInternalRange = "networkconnectivity.googleapis.com/projects/p/locations/global/internalRanges/app-range"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should reject both ip_cidr_range and reserved_internal_range", func() {
		target := minimal()
		target.Spec.ReservedInternalRange = "networkconnectivity.googleapis.com/projects/p/locations/global/internalRanges/app-range"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "at most one")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a secondary range with both CIDR sources", func() {
		target := minimal()
		target.Spec.SecondaryIpRanges = []*GcpSubnetworkSecondaryRange{{
			RangeName:             "pods",
			IpCidrRange:           "10.16.0.0/14",
			ReservedInternalRange: "networkconnectivity.googleapis.com/projects/p/locations/global/internalRanges/pods-range",
		}}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "exactly one CIDR source")).To(gomega.BeTrue())
	})

	ginkgo.It("should accept a secondary range backed by a reserved internal range", func() {
		target := minimal()
		target.Spec.SecondaryIpRanges = []*GcpSubnetworkSecondaryRange{{
			RangeName:             "pods",
			ReservedInternalRange: "networkconnectivity.googleapis.com/projects/p/locations/global/internalRanges/pods-range",
		}}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a secondary range with no CIDR source", func() {
		target := minimal()
		target.Spec.SecondaryIpRanges = []*GcpSubnetworkSecondaryRange{{
			RangeName: "pods",
		}}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject internal_ipv6_prefix without INTERNAL access", func() {
		target := minimal()
		target.Spec.InternalIpv6Prefix = "fd20:1:2:3::/64"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "INTERNAL")).To(gomega.BeTrue())
	})

	ginkgo.It("should accept internal_ipv6_prefix on an INTERNAL dual-stack subnet", func() {
		target := minimal()
		st := "IPV4_IPV6"
		target.Spec.StackType = &st
		target.Spec.Ipv6AccessType = "INTERNAL"
		target.Spec.InternalIpv6Prefix = "fd20:1:2:3::/64"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should reject ip_collection on an IPv4-only subnet", func() {
		target := minimal()
		target.Spec.IpCollection = "projects/p/regions/us-central1/publicDelegatedPrefixes/sub-pdp"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "BYOIP")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an unknown resolve_subnet_mask", func() {
		target := minimal()
		target.Spec.ResolveSubnetMask = "ARP_EVERYTHING"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "resolve_subnet_mask")).To(gomega.BeTrue())
	})

	ginkgo.It("should accept every deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON", ""} {
			target := minimal()
			target.Spec.DeletionPolicy = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should reject an unknown deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "FORCE"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
