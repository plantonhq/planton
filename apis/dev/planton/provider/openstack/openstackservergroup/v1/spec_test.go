package openstackservergroupv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestOpenStackServerGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackServerGroupSpec Validation Tests")
}

// minimalValidServerGroup returns a minimal valid OpenStackServerGroup
// with the required policy field set.
func minimalValidServerGroup() *OpenStackServerGroup {
	return &OpenStackServerGroup{
		ApiVersion: "openstack.planton.dev/v1",
		Kind:       "OpenStackServerGroup",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-server-group",
		},
		Spec: &OpenStackServerGroupSpec{
			Policy: "anti-affinity",
		},
	}
}

var _ = ginkgo.Describe("OpenStackServerGroupSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_server_group", func() {

			ginkgo.It("should not return a validation error for minimal valid server group with anti-affinity", func() {
				input := minimalValidServerGroup()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for server group with affinity policy", func() {
				input := minimalValidServerGroup()
				input.Spec.Policy = "affinity"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for server group with soft-affinity policy", func() {
				input := minimalValidServerGroup()
				input.Spec.Policy = "soft-affinity"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for server group with soft-anti-affinity policy", func() {
				input := minimalValidServerGroup()
				input.Spec.Policy = "soft-anti-affinity"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for server group with region override", func() {
				input := minimalValidServerGroup()
				input.Spec.Region = "RegionTwo"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fully-specified server group", func() {
				input := &OpenStackServerGroup{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackServerGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-server-group",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &OpenStackServerGroupSpec{
						Policy: "anti-affinity",
						Region: "RegionOne",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("openstack_server_group", func() {

			ginkgo.It("should return a validation error when policy is empty", func() {
				input := minimalValidServerGroup()
				input.Spec.Policy = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when policy is an invalid value", func() {
				input := minimalValidServerGroup()
				input.Spec.Policy = "spread"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when policy has wrong case", func() {
				input := minimalValidServerGroup()
				input.Spec.Policy = "Anti-Affinity"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidServerGroup()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidServerGroup()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidServerGroup()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackServerGroup{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackServerGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-server-group",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
