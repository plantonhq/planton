package openstackkeypairv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestOpenStackKeypairSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackKeypairSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("OpenStackKeypairSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_keypair", func() {

			ginkgo.It("should not return a validation error for minimal valid keypair with public key", func() {
				input := &OpenStackKeypair{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackKeypair",
					Metadata: &shared.CloudResourceMetadata{
						Name: "my-ssh-key",
					},
					Spec: &OpenStackKeypairSpec{
						PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDxyz user@host",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for keypair without public key (generated)", func() {
				input := &OpenStackKeypair{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackKeypair",
					Metadata: &shared.CloudResourceMetadata{
						Name: "generated-keypair",
					},
					Spec: &OpenStackKeypairSpec{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for keypair with region override", func() {
				input := &OpenStackKeypair{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackKeypair",
					Metadata: &shared.CloudResourceMetadata{
						Name: "regional-keypair",
					},
					Spec: &OpenStackKeypairSpec{
						PublicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA user@host",
						Region:    "RegionOne",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for keypair with labels", func() {
				input := &OpenStackKeypair{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackKeypair",
					Metadata: &shared.CloudResourceMetadata{
						Name: "labeled-keypair",
						Labels: map[string]string{
							"team": "platform",
							"env":  "production",
						},
					},
					Spec: &OpenStackKeypairSpec{
						PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDxyz user@host",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for keypair with org and env metadata", func() {
				input := &OpenStackKeypair{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackKeypair",
					Metadata: &shared.CloudResourceMetadata{
						Name: "org-keypair",
						Org:  "my-org",
						Env:  "staging",
					},
					Spec: &OpenStackKeypairSpec{
						PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDxyz user@host",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("openstack_keypair", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := &OpenStackKeypair{
					ApiVersion: "wrong.openmcf.org/v1",
					Kind:       "OpenStackKeypair",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-keypair",
					},
					Spec: &OpenStackKeypairSpec{
						PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDxyz user@host",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := &OpenStackKeypair{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "WrongKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-keypair",
					},
					Spec: &OpenStackKeypairSpec{
						PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDxyz user@host",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := &OpenStackKeypair{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackKeypair",
					Spec: &OpenStackKeypairSpec{
						PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDxyz user@host",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackKeypair{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackKeypair",
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
