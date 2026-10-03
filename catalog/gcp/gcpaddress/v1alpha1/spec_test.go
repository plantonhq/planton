package gcpaddressv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpAddressSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func ptr(s string) *string {
	return &s
}

func intPtr(i int32) *int32 {
	return &i
}

var _ = ginkgo.Describe("GcpAddressSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimalExternal := func() *GcpAddress {
		return &GcpAddress{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpAddress",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-regional-address",
			},
			Spec: &GcpAddressSpec{
				AddressName: "nat-external-ip",
				Region:      "us-central1",
			},
		}
	}

	minimalInternalEndpoint := func() *GcpAddress {
		return &GcpAddress{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpAddress",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-internal-endpoint",
			},
			Spec: &GcpAddressSpec{
				AddressName: "gce-endpoint-ip",
				Region:      "us-central1",
				AddressType: ptr("INTERNAL"),
				Purpose:     "GCE_ENDPOINT",
				Subnetwork: litRef(
					"projects/my-proj/regions/us-central1/subnetworks/my-subnet",
				),
			},
		}
	}

	minimalVpcPeering := func() *GcpAddress {
		pl := int32(20)
		return &GcpAddress{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpAddress",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-peering-range",
			},
			Spec: &GcpAddressSpec{
				AddressName:  "vpc-peering-range",
				Region:       "us-central1",
				AddressType:  ptr("INTERNAL"),
				Purpose:      "VPC_PEERING",
				PrefixLength: &pl,
				Network: litRef(
					"projects/my-proj/global/networks/my-vpc",
				),
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal external address spec", func() {
		gomega.Expect(validator.Validate(minimalExternal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a project_id literal", func() {
		target := minimalExternal()
		target.Spec.ProjectId = litRef("my-gcp-project-123")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a fully populated external address spec", func() {
		target := minimalExternal()
		target.Spec.ProjectId = litRef("my-gcp-project-123")
		target.Spec.Address = "34.120.0.1"
		target.Spec.AddressType = ptr("EXTERNAL")
		target.Spec.Description = "Static IP for Cloud NAT"
		target.Spec.IpVersion = ptr("IPV4")
		target.Spec.NetworkTier = "PREMIUM"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an IPv6 external address with endpoint type", func() {
		target := minimalExternal()
		target.Spec.IpVersion = ptr("IPV6")
		target.Spec.Ipv6EndpointType = "VM"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a valid internal GCE endpoint address", func() {
		gomega.Expect(validator.Validate(minimalInternalEndpoint())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a valid internal shared LB VIP", func() {
		target := minimalExternal()
		target.Spec.AddressType = ptr("INTERNAL")
		target.Spec.Purpose = "SHARED_LOADBALANCER_VIP"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a valid internal VPC peering range", func() {
		gomega.Expect(validator.Validate(minimalVpcPeering())).To(gomega.Succeed())
	})

	ginkgo.It("should accept prefix_length at lower bound (8)", func() {
		target := minimalVpcPeering()
		pl := int32(8)
		target.Spec.PrefixLength = &pl
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept prefix_length at upper bound (29)", func() {
		target := minimalVpcPeering()
		pl := int32(29)
		target.Spec.PrefixLength = &pl
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept IPSEC_INTERCONNECT with network", func() {
		target := minimalVpcPeering()
		target.Spec.Purpose = "IPSEC_INTERCONNECT"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept DNS_RESOLVER with subnetwork", func() {
		target := minimalInternalEndpoint()
		target.Spec.Purpose = "DNS_RESOLVER"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept user labels", func() {
		target := minimalExternal()
		target.Spec.Labels = map[string]string{"team": "platform"}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a BYOIP ip_collection PDP reference", func() {
		target := minimalExternal()
		target.Spec.IpCollection = "projects/my-proj/regions/us-central1/publicDelegatedPrefixes/my-pdp"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept each deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
			target := minimalExternal()
			target.Spec.DeletionPolicy = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject an invalid deletion_policy", func() {
		target := minimalExternal()
		target.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject when address_name is missing", func() {
		target := minimalExternal()
		target.Spec.AddressName = ""
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when region is missing", func() {
		target := minimalExternal()
		target.Spec.Region = ""
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid address_name", func() {
		target := minimalExternal()
		target.Spec.AddressName = "Invalid_Name"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "address_name")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an invalid region", func() {
		target := minimalExternal()
		target.Spec.Region = "US_CENTRAL1"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid address_type", func() {
		target := minimalExternal()
		target.Spec.AddressType = ptr("PUBLIC")
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid ip_version", func() {
		target := minimalExternal()
		target.Spec.IpVersion = ptr("IPV5")
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid network_tier", func() {
		target := minimalExternal()
		target.Spec.NetworkTier = "BASIC"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject network_tier on INTERNAL address_type", func() {
		target := minimalInternalEndpoint()
		target.Spec.NetworkTier = "PREMIUM"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "network_tier")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an invalid purpose", func() {
		target := minimalInternalEndpoint()
		target.Spec.Purpose = "PRIVATE_SERVICE_CONNECT"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject purpose with EXTERNAL address_type", func() {
		target := minimalExternal()
		target.Spec.Purpose = "GCE_ENDPOINT"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject purpose with unset address_type (defaults to EXTERNAL)", func() {
		target := minimalExternal()
		target.Spec.Purpose = "VPC_PEERING"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject VPC_PEERING without network", func() {
		target := minimalVpcPeering()
		target.Spec.Network = nil
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "network")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject GCE_ENDPOINT without subnetwork", func() {
		target := minimalInternalEndpoint()
		target.Spec.Subnetwork = nil
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "subnetwork")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject SHARED_LOADBALANCER_VIP with EXTERNAL address_type", func() {
		target := minimalExternal()
		target.Spec.AddressType = ptr("EXTERNAL")
		target.Spec.Purpose = "SHARED_LOADBALANCER_VIP"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject prefix_length below 8", func() {
		target := minimalVpcPeering()
		pl := int32(4)
		target.Spec.PrefixLength = &pl
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject prefix_length above 29", func() {
		target := minimalVpcPeering()
		pl := int32(31)
		target.Spec.PrefixLength = &pl
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid ipv6_endpoint_type", func() {
		target := minimalExternal()
		target.Spec.Ipv6EndpointType = "LB"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong kind constant", func() {
		target := minimalExternal()
		target.Kind = "GcpGlobalAddress"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong api_version", func() {
		target := minimalExternal()
		target.ApiVersion = "gcp.planton.dev/v2"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
