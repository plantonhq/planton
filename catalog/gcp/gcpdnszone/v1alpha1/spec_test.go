package gcpdnszonev1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestGcpDnsZoneSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpDnsZoneSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func fromRef(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

func ptr(s string) *string {
	return &s
}

func ptrBool(b bool) *bool {
	return &b
}

var _ = ginkgo.Describe("GcpDnsZoneSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimalPublic := func() *GcpDnsZone {
		return &GcpDnsZone{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpDnsZone",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "example.com",
			},
			Spec: &GcpDnsZoneSpec{
				ProjectId: litRef("my-gcp-project"),
			},
		}
	}

	expectValid := func(z *GcpDnsZone) {
		gomega.Expect(validator.Validate(z)).To(gomega.Succeed())
	}

	expectInvalid := func(z *GcpDnsZone, substr string) {
		err := validator.Validate(z)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), substr)).To(
			gomega.BeTrue(), "expected error to contain %q, got: %s", substr, err)
	}

	ginkgo.It("accepts a minimal public zone with literal project_id", func() {
		expectValid(minimalPublic())
	})

	ginkgo.It("accepts a project_id reference", func() {
		z := minimalPublic()
		z.Spec.ProjectId = fromRef("my-project")
		expectValid(z)
	})

	ginkgo.It("accepts a missing project_id (falls back to the provider's default project)", func() {
		z := minimalPublic()
		z.Spec.ProjectId = nil
		expectValid(z)
	})

	ginkgo.It("accepts an explicit dns_name FQDN", func() {
		z := minimalPublic()
		z.Spec.DnsName = "apps.example.com."
		expectValid(z)
	})

	ginkgo.It("accepts an empty dns_name (modules derive from metadata.name)", func() {
		z := minimalPublic()
		z.Spec.DnsName = ""
		expectValid(z)
	})

	ginkgo.It("rejects dns_name without trailing dot", func() {
		z := minimalPublic()
		z.Spec.DnsName = "example.com"
		expectInvalid(z, "dns_name")
	})

	ginkgo.It("rejects an invalid visibility value", func() {
		z := minimalPublic()
		z.Spec.Visibility = ptr("hybrid")
		expectInvalid(z, "visibility")
	})

	ginkgo.It("accepts explicit public visibility", func() {
		z := minimalPublic()
		z.Spec.Visibility = ptr("public")
		expectValid(z)
	})

	ginkgo.It("accepts a private zone with VPC network visibility", func() {
		z := minimalPublic()
		z.Spec.Visibility = ptr("private")
		z.Spec.PrivateVisibilityConfig = &GcpDnsZonePrivateVisibilityConfig{
			Networks: []*GcpDnsZonePrivateVisibilityNetwork{
				{NetworkUrl: litRef("projects/p/global/networks/default")},
			},
		}
		expectValid(z)
	})

	ginkgo.It("accepts a private zone visible to a GKE cluster", func() {
		z := minimalPublic()
		z.Spec.Visibility = ptr("private")
		z.Spec.PrivateVisibilityConfig = &GcpDnsZonePrivateVisibilityConfig{
			GkeClusters: []*GcpDnsZonePrivateVisibilityGkeCluster{
				{GkeClusterName: litRef("projects/p/locations/us-central1/clusters/prod")},
			},
		}
		expectValid(z)
	})

	ginkgo.It("rejects private_visibility_config without targets", func() {
		z := minimalPublic()
		z.Spec.Visibility = ptr("private")
		z.Spec.PrivateVisibilityConfig = &GcpDnsZonePrivateVisibilityConfig{}
		expectInvalid(z, "private_visibility_config requires")
	})

	ginkgo.It("rejects private_visibility_config on a public zone", func() {
		z := minimalPublic()
		z.Spec.Visibility = ptr("public")
		z.Spec.PrivateVisibilityConfig = &GcpDnsZonePrivateVisibilityConfig{
			Networks: []*GcpDnsZonePrivateVisibilityNetwork{
				{NetworkUrl: litRef("projects/p/global/networks/default")},
			},
		}
		expectInvalid(z, "private_visibility_config requires visibility private")
	})

	ginkgo.It("accepts DNSSEC enabled with defaults", func() {
		z := minimalPublic()
		z.Spec.DnssecConfig = &GcpDnsZoneDnssecConfig{
			State: ptr("on"),
		}
		expectValid(z)
	})

	ginkgo.It("rejects an invalid DNSSEC state", func() {
		z := minimalPublic()
		z.Spec.DnssecConfig = &GcpDnsZoneDnssecConfig{
			State: ptr("maybe"),
		}
		expectInvalid(z, "state")
	})

	ginkgo.It("rejects an invalid DNSSEC non_existence value", func() {
		z := minimalPublic()
		z.Spec.DnssecConfig = &GcpDnsZoneDnssecConfig{
			NonExistence: "nsec5",
		}
		expectInvalid(z, "non_existence")
	})

	ginkgo.It("rejects an invalid DNSSEC key algorithm", func() {
		z := minimalPublic()
		z.Spec.DnssecConfig = &GcpDnsZoneDnssecConfig{
			DefaultKeySpecs: []*GcpDnsZoneDnssecKeySpec{
				{Algorithm: "md5", KeyType: "zoneSigning"},
			},
		}
		expectInvalid(z, "algorithm")
	})

	ginkgo.It("accepts a private forwarding zone", func() {
		z := minimalPublic()
		z.Spec.Visibility = ptr("private")
		z.Spec.ForwardingConfig = &GcpDnsZoneForwardingConfig{
			TargetNameServers: []*GcpDnsZoneForwardingTargetNameServer{
				{Ipv4Address: "10.0.0.2"},
			},
		}
		expectValid(z)
	})

	ginkgo.It("rejects forwarding_config on a public zone", func() {
		z := minimalPublic()
		z.Spec.Visibility = ptr("public")
		z.Spec.ForwardingConfig = &GcpDnsZoneForwardingConfig{
			TargetNameServers: []*GcpDnsZoneForwardingTargetNameServer{
				{Ipv4Address: "10.0.0.2"},
			},
		}
		expectInvalid(z, "forwarding_config requires visibility private")
	})

	ginkgo.It("rejects a forwarding target without an address", func() {
		z := minimalPublic()
		z.Spec.Visibility = ptr("private")
		z.Spec.ForwardingConfig = &GcpDnsZoneForwardingConfig{
			TargetNameServers: []*GcpDnsZoneForwardingTargetNameServer{
				{},
			},
		}
		expectInvalid(z, "target_name_server")
	})

	ginkgo.It("rejects an invalid forwarding_path", func() {
		z := minimalPublic()
		z.Spec.Visibility = ptr("private")
		z.Spec.ForwardingConfig = &GcpDnsZoneForwardingConfig{
			TargetNameServers: []*GcpDnsZoneForwardingTargetNameServer{
				{Ipv4Address: "10.0.0.2", ForwardingPath: "internet"},
			},
		}
		expectInvalid(z, "forwarding_path")
	})

	ginkgo.It("accepts a private peering zone", func() {
		z := minimalPublic()
		z.Spec.Visibility = ptr("private")
		z.Spec.PeeringConfig = &GcpDnsZonePeeringConfig{
			TargetNetwork: litRef("projects/p/global/networks/host"),
		}
		expectValid(z)
	})

	ginkgo.It("rejects peering_config on a public zone", func() {
		z := minimalPublic()
		z.Spec.Visibility = ptr("public")
		z.Spec.PeeringConfig = &GcpDnsZonePeeringConfig{
			TargetNetwork: litRef("projects/p/global/networks/host"),
		}
		expectInvalid(z, "peering_config requires visibility private")
	})

	ginkgo.It("rejects forwarding and peering together", func() {
		z := minimalPublic()
		z.Spec.Visibility = ptr("private")
		z.Spec.ForwardingConfig = &GcpDnsZoneForwardingConfig{
			TargetNameServers: []*GcpDnsZoneForwardingTargetNameServer{
				{Ipv4Address: "10.0.0.2"},
			},
		}
		z.Spec.PeeringConfig = &GcpDnsZonePeeringConfig{
			TargetNetwork: litRef("projects/p/global/networks/host"),
		}
		expectInvalid(z, "mutually exclusive")
	})

	ginkgo.It("accepts cloud logging configuration", func() {
		z := minimalPublic()
		z.Spec.CloudLoggingConfig = &GcpDnsZoneCloudLoggingConfig{
			EnableLogging: true,
		}
		expectValid(z)
	})

	ginkgo.It("accepts force_destroy and custom labels", func() {
		z := minimalPublic()
		z.Spec.ForceDestroy = ptrBool(true)
		z.Spec.Labels = map[string]string{"team": "platform"}
		expectValid(z)
	})

	ginkgo.It("accepts an IPv6 forwarding target", func() {
		z := minimalPublic()
		z.Spec.Visibility = ptr("private")
		z.Spec.ForwardingConfig = &GcpDnsZoneForwardingConfig{
			TargetNameServers: []*GcpDnsZoneForwardingTargetNameServer{
				{Ipv6Address: "2001:db8::53"},
			},
		}
		expectValid(z)
	})

	ginkgo.It("rejects a forwarding target with both IPv4 and IPv6", func() {
		z := minimalPublic()
		z.Spec.Visibility = ptr("private")
		z.Spec.ForwardingConfig = &GcpDnsZoneForwardingConfig{
			TargetNameServers: []*GcpDnsZoneForwardingTargetNameServer{
				{Ipv4Address: "10.1.2.53", Ipv6Address: "2001:db8::53"},
			},
		}
		expectInvalid(z, "never both")
	})

	ginkgo.It("accepts each deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
			z := minimalPublic()
			z.Spec.DeletionPolicy = v
			expectValid(z)
		}
	})

	ginkgo.It("rejects an invalid deletion_policy", func() {
		z := minimalPublic()
		z.Spec.DeletionPolicy = "KEEP"
		expectInvalid(z, "deletion_policy")
	})
})
