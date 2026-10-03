package gcpgkefleetscopev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/pkg/refannotations"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpGkeFleetScopeSpec Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value}}
}

func reference(kind catalogkind.CatalogKind, name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Kind: kind, Name: name}}}
}

var _ = ginkgo.Describe("GcpGkeFleetScopeSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpGkeFleetScope {
		return &GcpGkeFleetScope{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpGkeFleetScope",
			Metadata:   &shared.CatalogObjectMetadata{Name: "team-orders"},
			Spec:       &GcpGkeFleetScopeSpec{},
		}
	}

	full := func() *GcpGkeFleetScope {
		msg := minimal()
		msg.Spec.ProjectId = reference(catalogkind.CatalogKind_GcpGkeFleet, "platform-fleet")
		msg.Spec.ScopeId = "orders"
		msg.Spec.Labels = map[string]string{"team": "orders"}
		msg.Spec.NamespaceLabels = map[string]string{"cost-center": "orders"}
		msg.Spec.Namespaces = []*GcpGkeFleetScopeNamespace{
			{ScopeNamespaceId: "orders-api", NamespaceLabels: map[string]string{"tier": "api"}},
			{ScopeNamespaceId: "orders-workers"},
		}
		msg.Spec.RbacRoleBindings = []*GcpGkeFleetScopeRbacRoleBinding{
			{ScopeRbacRoleBindingId: "orders-devs", Group: literal("orders-devs@example.com"), Role: &GcpGkeFleetScopeRole{PredefinedRole: "EDIT"}},
			{ScopeRbacRoleBindingId: "orders-deployer", User: reference(catalogkind.CatalogKind_GcpServiceAccount, "orders-deployer"), Role: &GcpGkeFleetScopeRole{CustomRole: "deployer"}},
		}
		msg.Spec.MembershipBindings = []*GcpGkeFleetScopeMembershipBinding{
			{MembershipBindingId: "orders-uc1", Membership: reference(catalogkind.CatalogKind_GcpGkeCluster, "orders-uc1")},
			{MembershipBindingId: "orders-ew4", Membership: literal("projects/platform-host/locations/europe-west4/memberships/orders-ew4")},
		}
		return msg
	}

	ginkgo.It("should accept an empty scope and a fully declared one", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
		gomega.Expect(validator.Validate(full())).To(gomega.Succeed())
	})

	ginkgo.It("should reject a namespace Google reserves, or one that is not a DNS label", func() {
		for _, id := range []string{"kube-system", "default", "config-management-system", "gke-connect", "Orders", "orders_api"} {
			msg := minimal()
			msg.Spec.Namespaces = []*GcpGkeFleetScopeNamespace{{ScopeNamespaceId: id}}
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), id)
		}
	})

	ginkgo.It("should reject duplicate child IDs", func() {
		msg := full()
		msg.Spec.Namespaces[1].ScopeNamespaceId = "orders-api"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.RbacRoleBindings[1].ScopeRbacRoleBindingId = "orders-devs"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.MembershipBindings[1].MembershipBindingId = "orders-uc1"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require exactly one principal per role binding", func() {
		msg := full()
		msg.Spec.RbacRoleBindings[0].User = literal("alice@example.com")
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.RbacRoleBindings[0].Group = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require exactly one role form and a known predefined role", func() {
		msg := full()
		msg.Spec.RbacRoleBindings[0].Role = &GcpGkeFleetScopeRole{PredefinedRole: "VIEW", CustomRole: "viewer"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.RbacRoleBindings[0].Role = &GcpGkeFleetScopeRole{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.RbacRoleBindings[0].Role = &GcpGkeFleetScopeRole{PredefinedRole: "OWNER"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.RbacRoleBindings[0].Role = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a membership, as a full name when literal", func() {
		msg := full()
		msg.Spec.MembershipBindings[1].Membership = literal("orders-ew4")
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.MembershipBindings[1].Membership = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed scope ID and an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.ScopeId = "Orders"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})

// A cluster binding composes from either way a cluster joins a fleet: an
// explicit membership's name, or the fleet_membership a GcpGkeCluster with
// fleet_project exports.
func TestMembershipBindingAcceptsBothRegistrationPaths(t *testing.T) {
	field := refannotations.Of((&GcpGkeFleetScopeMembershipBinding{}).ProtoReflect().Descriptor().Fields().ByName("membership"))
	if path, ok := field.DefaultPath(catalogkind.CatalogKind_GcpGkeCluster); !ok || path != "status.outputs.fleet_membership" {
		t.Fatalf("membership composes from a GcpGkeCluster at %q (ok=%t), want status.outputs.fleet_membership", path, ok)
	}
	if path, ok := field.DefaultPath(catalogkind.CatalogKind_GcpGkeFleetMembership); !ok || path != "status.outputs.name" {
		t.Fatalf("membership composes from a GcpGkeFleetMembership at %q (ok=%t), want status.outputs.name", path, ok)
	}
	if kind := field.EffectiveKind(catalogkind.CatalogKind_unspecified); kind != catalogkind.CatalogKind_GcpGkeFleetMembership {
		t.Fatalf("a kindless valueFrom on membership reads as %s, want GcpGkeFleetMembership", kind)
	}
}
