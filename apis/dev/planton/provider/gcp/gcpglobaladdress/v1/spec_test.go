package gcpglobaladdressv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpGlobalAddressSpec Suite")
}

func ptr(s string) *string {
	return &s
}

func intPtr(i int32) *int32 {
	return &i
}

var _ = ginkgo.Describe("GcpGlobalAddressSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpGlobalAddress for external IP (the default use case).
	minimalExternal := func() *GcpGlobalAddress {
		return &GcpGlobalAddress{
			ApiVersion: "gcp.planton.dev/v1",
			Kind:       "GcpGlobalAddress",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-global-address",
			},
			Spec: &GcpGlobalAddressSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-gcp-project"},
				},
				AddressName: "lb-external-ip",
			},
		}
	}

	// Helper to build a minimal valid internal VPC peering address.
	minimalVpcPeering := func() *GcpGlobalAddress {
		prefixLen := int32(20)
		return &GcpGlobalAddress{
			ApiVersion: "gcp.planton.dev/v1",
			Kind:       "GcpGlobalAddress",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-peering-range",
			},
			Spec: &GcpGlobalAddressSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-gcp-project"},
				},
				AddressName:  "vpc-peering-range",
				AddressType:  ptr("INTERNAL"),
				Purpose:      "VPC_PEERING",
				PrefixLength: &prefixLen,
				Network: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/my-gcp-project/global/networks/my-vpc"},
				},
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal external address spec", func() {
		msg := minimalExternal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a fully populated external address spec", func() {
		msg := minimalExternal()
		msg.Spec.Address = "34.120.0.1"
		msg.Spec.AddressType = ptr("EXTERNAL")
		msg.Spec.Description = "Static IP for HTTPS load balancer"
		msg.Spec.IpVersion = ptr("IPV4")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an IPv6 external address", func() {
		msg := minimalExternal()
		msg.Spec.IpVersion = ptr("IPV6")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a valid internal VPC peering range", func() {
		msg := minimalVpcPeering()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a private service connect address", func() {
		msg := minimalExternal()
		msg.Spec.AddressType = ptr("INTERNAL")
		msg.Spec.Purpose = "PRIVATE_SERVICE_CONNECT"
		msg.Spec.Network = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/my-gcp-project/global/networks/my-vpc"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept prefix_length at lower bound (8)", func() {
		msg := minimalVpcPeering()
		pl := int32(8)
		msg.Spec.PrefixLength = &pl
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept prefix_length at upper bound (29)", func() {
		msg := minimalVpcPeering()
		pl := int32(29)
		msg.Spec.PrefixLength = &pl
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject when project_id is missing", func() {
		msg := minimalExternal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when address_name is missing", func() {
		msg := minimalExternal()
		msg.Spec.AddressName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid address_name (starts with number)", func() {
		msg := minimalExternal()
		msg.Spec.AddressName = "123-invalid"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an address_name with uppercase", func() {
		msg := minimalExternal()
		msg.Spec.AddressName = "Invalid-Name"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid address_type", func() {
		msg := minimalExternal()
		msg.Spec.AddressType = ptr("PUBLIC")
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid ip_version", func() {
		msg := minimalExternal()
		msg.Spec.IpVersion = ptr("IPV5")
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid purpose", func() {
		msg := minimalExternal()
		msg.Spec.AddressType = ptr("INTERNAL")
		msg.Spec.Purpose = "INVALID_PURPOSE"
		msg.Spec.Network = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/my-gcp-project/global/networks/my-vpc"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject purpose with EXTERNAL address_type", func() {
		msg := minimalExternal()
		msg.Spec.AddressType = ptr("EXTERNAL")
		msg.Spec.Purpose = "VPC_PEERING"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject purpose with unset address_type (defaults to EXTERNAL)", func() {
		msg := minimalExternal()
		msg.Spec.Purpose = "VPC_PEERING"
		// address_type not set — proto3 default is empty string, not "INTERNAL"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject VPC_PEERING without prefix_length", func() {
		msg := minimalVpcPeering()
		msg.Spec.PrefixLength = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject prefix_length below 8", func() {
		msg := minimalVpcPeering()
		pl := int32(4)
		msg.Spec.PrefixLength = &pl
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject prefix_length above 29", func() {
		msg := minimalVpcPeering()
		pl := int32(31)
		msg.Spec.PrefixLength = &pl
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject INTERNAL address_type without network", func() {
		msg := minimalExternal()
		msg.Spec.AddressType = ptr("INTERNAL")
		msg.Spec.Purpose = "PRIVATE_SERVICE_CONNECT"
		// network not set — CEL rule should catch this
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
