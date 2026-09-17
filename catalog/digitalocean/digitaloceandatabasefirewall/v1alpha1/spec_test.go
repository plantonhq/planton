package digitaloceandatabasefirewallv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	fk "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestDigitalOceanDatabaseFirewallSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanDatabaseFirewallSpec Validation Suite")
}

var _ = ginkgo.Describe("DigitalOceanDatabaseFirewallSpec validations", func() {

	newRef := func(value string) *fk.StringValueOrRef {
		return &fk.StringValueOrRef{
			LiteralOrRef: &fk.StringValueOrRef_Value{Value: value},
		}
	}

	makeValidSpec := func() *DigitalOceanDatabaseFirewallSpec {
		return &DigitalOceanDatabaseFirewallSpec{
			Cluster: newRef("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"),
			IpRules: []string{"203.0.113.10"},
		}
	}

	ginkgo.Context("Required fields", func() {
		ginkgo.It("accepts a minimal valid spec (one IP rule)", func() {
			err := protovalidate.Validate(makeValidSpec())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects spec with missing cluster", func() {
			spec := makeValidSpec()
			spec.Cluster = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("At least one rule", func() {
		ginkgo.It("rejects a spec with no rules in any list", func() {
			spec := makeValidSpec()
			spec.IpRules = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts a droplet reference as the only rule", func() {
			spec := makeValidSpec()
			spec.IpRules = nil
			spec.DropletIds = []*fk.StringValueOrRef{newRef("123456789")}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a kubernetes cluster reference as the only rule", func() {
			spec := makeValidSpec()
			spec.IpRules = nil
			spec.KubernetesClusterIds = []*fk.StringValueOrRef{newRef("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts an app reference as the only rule", func() {
			spec := makeValidSpec()
			spec.IpRules = nil
			spec.AppIds = []*fk.StringValueOrRef{newRef("cccccccc-dddd-eeee-ffff-000000000000")}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a tag as the only rule", func() {
			spec := makeValidSpec()
			spec.IpRules = nil
			spec.Tags = []string{"backend"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("ip_rules", func() {
		ginkgo.It("accepts a CIDR block", func() {
			spec := makeValidSpec()
			spec.IpRules = []string{"10.10.0.0/16"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a /32 CIDR block", func() {
			spec := makeValidSpec()
			spec.IpRules = []string{"203.0.113.10/32"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		// DigitalOcean's database firewall rejects every IPv6 shape at apply
		// (`422 invalid ip format`), so the spec refuses them at validation.
		ginkgo.It("rejects an IPv6 address", func() {
			spec := makeValidSpec()
			spec.IpRules = []string{"2001:db8::1"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an IPv6 CIDR block", func() {
			spec := makeValidSpec()
			spec.IpRules = []string{"2001:db8::/32"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a value that is neither IP nor CIDR", func() {
			spec := makeValidSpec()
			spec.IpRules = []string{"office-network"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("tags", func() {
		ginkgo.It("rejects a tag with forbidden characters", func() {
			spec := makeValidSpec()
			spec.Tags = []string{"has spaces"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
