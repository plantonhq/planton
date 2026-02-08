package openstackcomputekeypairv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestOpenStackComputeKeypairSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackComputeKeypairSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("OpenStackComputeKeypairSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_compute_keypair", func() {

			ginkgo.It("should not return a validation error for minimal valid keypair with public key", func() {
				input := &OpenStackComputeKeypair{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackComputeKeypair",
					Metadata: &shared.CloudResourceMetadata{
						Name: "my-ssh-key",
					},
					Spec: &OpenStackComputeKeypairSpec{
						PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDxyz user@host",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for keypair without public key (generated)", func() {
				input := &OpenStackComputeKeypair{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackComputeKeypair",
					Metadata: &shared.CloudResourceMetadata{
						Name: "generated-keypair",
					},
					Spec: &OpenStackComputeKeypairSpec{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for keypair with region override", func() {
				input := &OpenStackComputeKeypair{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackComputeKeypair",
					Metadata: &shared.CloudResourceMetadata{
						Name: "regional-keypair",
					},
					Spec: &OpenStackComputeKeypairSpec{
						PublicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA user@host",
						Region:    "RegionOne",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for keypair with labels", func() {
				input := &OpenStackComputeKeypair{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackComputeKeypair",
					Metadata: &shared.CloudResourceMetadata{
						Name: "labeled-keypair",
						Labels: map[string]string{
							"team": "platform",
							"env":  "production",
						},
					},
					Spec: &OpenStackComputeKeypairSpec{
						PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDxyz user@host",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for keypair with org and env metadata", func() {
				input := &OpenStackComputeKeypair{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackComputeKeypair",
					Metadata: &shared.CloudResourceMetadata{
						Name: "org-keypair",
						Org:  "my-org",
						Env:  "staging",
					},
					Spec: &OpenStackComputeKeypairSpec{
						PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDxyz user@host",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("openstack_compute_keypair", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := &OpenStackComputeKeypair{
					ApiVersion: "wrong.openmcf.org/v1",
					Kind:       "OpenStackComputeKeypair",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-keypair",
					},
					Spec: &OpenStackComputeKeypairSpec{
						PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDxyz user@host",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := &OpenStackComputeKeypair{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "WrongKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-keypair",
					},
					Spec: &OpenStackComputeKeypairSpec{
						PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDxyz user@host",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := &OpenStackComputeKeypair{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackComputeKeypair",
					Spec: &OpenStackComputeKeypairSpec{
						PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDxyz user@host",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackComputeKeypair{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackComputeKeypair",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-keypair",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
