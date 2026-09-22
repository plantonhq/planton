package gcpcloudidentitygroupv1alpha1

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
	ginkgo.RunSpecs(t, "GcpCloudIdentityGroupSpec Suite")
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

var _ = ginkgo.Describe("GcpCloudIdentityGroupSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpCloudIdentityGroup {
		return &GcpCloudIdentityGroup{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpCloudIdentityGroup",
			Metadata:   &shared.CloudResourceMetadata{Name: "platform-admins"},
			Spec: &GcpCloudIdentityGroupSpec{
				GroupEmail: "platform-admins@example.com",
				CustomerId: "customers/C01abc2de",
			},
		}
	}

	withMembers := func() *GcpCloudIdentityGroup {
		msg := minimal()
		msg.Spec.Memberships = []*GcpCloudIdentityGroupMembership{
			{Member: litRef("alice@example.com"), Roles: []*GcpCloudIdentityGroupMembershipRole{{Name: "MEMBER"}, {Name: "OWNER"}}},
			{Member: nameRef("deployer-sa")},
			{Member: litRef("contractor@example.com"), Roles: []*GcpCloudIdentityGroupMembershipRole{{Name: "MEMBER", ExpireTime: "2027-01-01T00:00:00Z"}}, CreateIgnoreAlreadyExists: true},
		}
		return msg
	}

	ginkgo.It("should accept the minimal group and a group with members", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
		gomega.Expect(validator.Validate(withMembers())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every lever together", func() {
		msg := withMembers()
		msg.Spec.DisplayName = "Platform Admins"
		msg.Spec.Description = "Owners of the platform projects"
		msg.Spec.Security = true
		msg.Spec.InitialGroupConfig = proto.String("WITH_INITIAL_OWNER")
		msg.Spec.GroupNamespace = "identitysources/abc"
		msg.Spec.Memberships[0].MemberNamespace = "identitysources/abc"
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require group_email and customer_id in their forms", func() {
		msg := minimal()
		msg.Spec.GroupEmail = ""
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = minimal()
		msg.Spec.GroupEmail = "platform-admins"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = minimal()
		msg.Spec.CustomerId = "C01abc2de"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should require MEMBER among listed roles and unique roles", func() {
		msg := withMembers()
		msg.Spec.Memberships[0].Roles = []*GcpCloudIdentityGroupMembershipRole{{Name: "OWNER"}}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = withMembers()
		msg.Spec.Memberships[0].Roles = []*GcpCloudIdentityGroupMembershipRole{{Name: "MEMBER"}, {Name: "MEMBER"}}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid role, an expiry on a non-MEMBER role, and a malformed expiry", func() {
		msg := withMembers()
		msg.Spec.Memberships[0].Roles[1].Name = "ADMIN"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = withMembers()
		msg.Spec.Memberships[0].Roles[1].ExpireTime = "2027-01-01T00:00:00Z"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = withMembers()
		msg.Spec.Memberships[2].Roles[0].ExpireTime = "next year"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a membership without a member and duplicate members", func() {
		msg := withMembers()
		msg.Spec.Memberships[1].Member = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = withMembers()
		msg.Spec.Memberships = append(msg.Spec.Memberships, &GcpCloudIdentityGroupMembership{Member: litRef("alice@example.com")})
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid initial_group_config or deletion_policy", func() {
		msg := minimal()
		msg.Spec.InitialGroupConfig = proto.String("WITH_OWNER")
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})
})
