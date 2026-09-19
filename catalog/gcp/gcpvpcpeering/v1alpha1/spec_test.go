package gcpvpcpeeringv1alpha1

import (
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
	ginkgo.RunSpecs(t, "GcpVpcPeeringSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func nameRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: v}},
	}
}

var _ = ginkgo.Describe("GcpVpcPeeringSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// The CREATE form: this side creates the peering to a named peer.
	createForm := func() *GcpVpcPeering {
		return &GcpVpcPeering{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpVpcPeering",
			Metadata: &shared.CloudResourceMetadata{
				Name: "hub-to-spoke",
			},
			Spec: &GcpVpcPeeringSpec{
				Network:     litRef("projects/acme-net/global/networks/hub"),
				PeerNetwork: litRef("projects/acme-app/global/networks/spoke"),
			},
		}
	}

	// The ROUTES-CONFIG form: the peering exists; only route exchange is managed.
	routesConfigForm := func() *GcpVpcPeering {
		target := createForm()
		target.Metadata.Name = "sql-routes"
		target.Spec.PeerNetwork = nil
		target.Spec.PeeringName = "servicenetworking-googleapis-com"
		target.Spec.ExportCustomRoutes = true
		return target
	}

	expectError := func(target *GcpVpcPeering, substring string) {
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(substring))
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept the create form with both networks as literals, named from metadata", func() {
		gomega.Expect(validator.Validate(createForm())).To(gomega.Succeed())
	})

	ginkgo.It("should accept both networks by reference with every create-form knob set", func() {
		target := createForm()
		target.Spec.Network = nameRef("hub")
		target.Spec.PeerNetwork = nameRef("spoke")
		target.Spec.PeeringName = "hub-to-spoke-a"
		target.Spec.ExportCustomRoutes = true
		target.Spec.ImportCustomRoutes = true
		target.Spec.ExportSubnetRoutesWithPublicIp = proto.Bool(false)
		target.Spec.ImportSubnetRoutesWithPublicIp = proto.Bool(true)
		target.Spec.StackType = "IPV4_IPV6"
		target.Spec.UpdateStrategy = "CONSENSUS"
		target.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept the routes-config form on a Google-managed peering", func() {
		gomega.Expect(validator.Validate(routesConfigForm())).To(gomega.Succeed())
	})

	ginkgo.It("should accept the routes-config form with the public-IP subnet flags set", func() {
		target := routesConfigForm()
		target.Spec.ExportSubnetRoutesWithPublicIp = proto.Bool(true)
		target.Spec.ImportSubnetRoutesWithPublicIp = proto.Bool(false)
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept every stack_type, update_strategy, and deletion_policy value on the create form", func() {
		for _, stackType := range []string{"", "IPV4_ONLY", "IPV4_IPV6"} {
			target := createForm()
			target.Spec.StackType = stackType
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed(), stackType)
		}
		for _, strategy := range []string{"", "INDEPENDENT", "CONSENSUS"} {
			target := createForm()
			target.Spec.UpdateStrategy = strategy
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed(), strategy)
		}
		for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			target := createForm()
			target.Spec.DeletionPolicy = policy
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed(), policy)
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a peering without its own network", func() {
		target := createForm()
		target.Spec.Network = nil
		expectError(target, "network")
	})

	ginkgo.It("should reject a malformed peering_name", func() {
		for _, name := range []string{"Hub", "-hub", "hub-", "hub_spoke", string(make([]byte, 64))} {
			target := createForm()
			target.Spec.PeeringName = name
			expectError(target, "peering_name must be 1-63 characters")
		}
	})

	ginkgo.It("should reject create-form-only knobs on the routes-config form", func() {
		target := routesConfigForm()
		target.Spec.StackType = "IPV4_ONLY"
		expectError(target, "stack_type applies only when this resource creates the peering")

		target = routesConfigForm()
		target.Spec.UpdateStrategy = "CONSENSUS"
		expectError(target, "update_strategy applies only when this resource creates the peering")

		target = routesConfigForm()
		target.Spec.DeletionPolicy = "ABANDON"
		expectError(target, "deletion_policy applies only when this resource creates the peering")
	})

	ginkgo.It("should reject unknown enum values", func() {
		target := createForm()
		target.Spec.StackType = "IPV6_ONLY"
		expectError(target, "stack_type must be IPV4_ONLY or IPV4_IPV6")

		target = createForm()
		target.Spec.UpdateStrategy = "EAGER"
		expectError(target, "update_strategy must be INDEPENDENT or CONSENSUS")

		target = createForm()
		target.Spec.DeletionPolicy = "RETAIN"
		expectError(target, "deletion_policy must be one of")
	})
})
