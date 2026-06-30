package hetznercloudserverv1

import (
	"testing"

	"buf.build/go/protovalidate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestHetznerCloudServerSpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HetznerCloudServerSpec Validation Suite")
}

var _ = Describe("HetznerCloudServerSpec validations", func() {

	Context("with valid specs", func() {
		It("should accept a minimal spec with only required fields", func() {
			spec := &HetznerCloudServerSpec{
				ServerType: "cx22",
				Image:      "ubuntu-24.04",
				Location:   "fsn1",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a spec with SSH keys", func() {
			spec := &HetznerCloudServerSpec{
				ServerType: "cx22",
				Image:      "ubuntu-24.04",
				Location:   "fsn1",
				SshKeys: []*foreignkeyv1.StringValueOrRef{
					{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-key"}},
					{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "12345"}},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a spec with user_data", func() {
			spec := &HetznerCloudServerSpec{
				ServerType: "cx22",
				Image:      "debian-12",
				Location:   "nbg1",
				UserData:   "#!/bin/bash\napt-get update",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a spec with placement group reference", func() {
			spec := &HetznerCloudServerSpec{
				ServerType: "cx22",
				Image:      "ubuntu-24.04",
				Location:   "fsn1",
				PlacementGroupId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "67890"},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a spec with firewall references", func() {
			spec := &HetznerCloudServerSpec{
				ServerType: "cx22",
				Image:      "ubuntu-24.04",
				Location:   "fsn1",
				FirewallIds: []*foreignkeyv1.StringValueOrRef{
					{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "111"}},
					{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "222"}},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a spec with network attachments", func() {
			spec := &HetznerCloudServerSpec{
				ServerType: "cx22",
				Image:      "ubuntu-24.04",
				Location:   "fsn1",
				Networks: []*HetznerCloudServerSpec_NetworkAttachment{
					{
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "100"},
						},
						Ip: "10.0.1.5",
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a spec with network attachment and alias IPs", func() {
			spec := &HetznerCloudServerSpec{
				ServerType: "cx22",
				Image:      "ubuntu-24.04",
				Location:   "fsn1",
				Networks: []*HetznerCloudServerSpec_NetworkAttachment{
					{
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "100"},
						},
						Ip:       "10.0.1.5",
						AliasIps: []string{"10.0.1.6", "10.0.1.7"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a spec with backups and protections enabled", func() {
			spec := &HetznerCloudServerSpec{
				ServerType:             "cpx11",
				Image:                  "rocky-9",
				Location:               "hel1",
				Backups:                true,
				KeepDisk:               true,
				DeleteProtection:       true,
				RebuildProtection:      true,
				ShutdownBeforeDeletion: true,
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a spec with dns_ptr", func() {
			spec := &HetznerCloudServerSpec{
				ServerType: "cx22",
				Image:      "ubuntu-24.04",
				Location:   "fsn1",
				DnsPtr:     "web.example.com",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a spec with public_net block", func() {
			ipv4Enabled := true
			ipv6Enabled := false
			spec := &HetznerCloudServerSpec{
				ServerType: "cx22",
				Image:      "ubuntu-24.04",
				Location:   "fsn1",
				PublicNet: &HetznerCloudServerSpec_PublicNet{
					Ipv4Enabled: &ipv4Enabled,
					Ipv6Enabled: &ipv6Enabled,
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a spec with primary IP references in public_net", func() {
			spec := &HetznerCloudServerSpec{
				ServerType: "cx22",
				Image:      "ubuntu-24.04",
				Location:   "fsn1",
				PublicNet: &HetznerCloudServerSpec_PublicNet{
					Ipv4: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "500"},
					},
					Ipv6: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "501"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a fully-populated spec", func() {
			ipv4Enabled := true
			ipv6Enabled := true
			spec := &HetznerCloudServerSpec{
				ServerType: "cax11",
				Image:      "ubuntu-24.04",
				Location:   "fsn1",
				SshKeys: []*foreignkeyv1.StringValueOrRef{
					{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-key"}},
				},
				UserData: "#cloud-config\npackage_update: true",
				PlacementGroupId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "200"},
				},
				FirewallIds: []*foreignkeyv1.StringValueOrRef{
					{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "300"}},
				},
				PublicNet: &HetznerCloudServerSpec_PublicNet{
					Ipv4Enabled: &ipv4Enabled,
					Ipv6Enabled: &ipv6Enabled,
				},
				Networks: []*HetznerCloudServerSpec_NetworkAttachment{
					{
						NetworkId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "400"},
						},
						Ip: "10.0.1.5",
					},
				},
				Backups:                true,
				KeepDisk:               true,
				DeleteProtection:       true,
				RebuildProtection:      true,
				ShutdownBeforeDeletion: true,
				DnsPtr:                 "web.example.com",
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})
	})

	Context("with invalid specs", func() {
		It("should reject an empty server_type", func() {
			spec := &HetznerCloudServerSpec{
				ServerType: "",
				Image:      "ubuntu-24.04",
				Location:   "fsn1",
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject a missing server_type (zero value)", func() {
			spec := &HetznerCloudServerSpec{
				Image:    "ubuntu-24.04",
				Location: "fsn1",
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject an empty image", func() {
			spec := &HetznerCloudServerSpec{
				ServerType: "cx22",
				Image:      "",
				Location:   "fsn1",
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject a missing image (zero value)", func() {
			spec := &HetznerCloudServerSpec{
				ServerType: "cx22",
				Location:   "fsn1",
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject an empty location", func() {
			spec := &HetznerCloudServerSpec{
				ServerType: "cx22",
				Image:      "ubuntu-24.04",
				Location:   "",
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject a missing location (zero value)", func() {
			spec := &HetznerCloudServerSpec{
				ServerType: "cx22",
				Image:      "ubuntu-24.04",
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject a network attachment without network_id", func() {
			spec := &HetznerCloudServerSpec{
				ServerType: "cx22",
				Image:      "ubuntu-24.04",
				Location:   "fsn1",
				Networks: []*HetznerCloudServerSpec_NetworkAttachment{
					{
						Ip: "10.0.1.5",
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})
	})
})
