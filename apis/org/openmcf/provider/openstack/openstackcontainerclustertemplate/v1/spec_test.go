package openstackcontainerclustertemplatev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOpenStackContainerClusterTemplateSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackContainerClusterTemplateSpec Validation Tests")
}

func int32Ptr(i int32) *int32 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidTemplate() *OpenStackContainerClusterTemplate {
	return &OpenStackContainerClusterTemplate{
		ApiVersion: "openstack.openmcf.org/v1",
		Kind:       "OpenStackContainerClusterTemplate",
		Metadata: &shared.CloudResourceMetadata{
			Name: "k8s-template",
		},
		Spec: &OpenStackContainerClusterTemplateSpec{
			Coe:   "kubernetes",
			Image: newStringValueOrRef("fedora-coreos-39"),
		},
	}
}

var _ = ginkgo.Describe("OpenStackContainerClusterTemplateSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_container_cluster_template", func() {

			ginkgo.It("should not return a validation error for minimal valid template", func() {
				input := minimalValidTemplate()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for template with keypair", func() {
				input := minimalValidTemplate()
				input.Spec.Keypair = newStringValueOrRef("my-keypair")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for template with external_network", func() {
				input := minimalValidTemplate()
				input.Spec.ExternalNetwork = newStringValueOrRef("public-net-uuid")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for template with fixed_network and fixed_subnet", func() {
				input := minimalValidTemplate()
				input.Spec.FixedNetwork = newStringValueOrRef("tenant-net-uuid")
				input.Spec.FixedSubnet = newStringValueOrRef("tenant-subnet-uuid")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for template with network and volume drivers", func() {
				input := minimalValidTemplate()
				input.Spec.NetworkDriver = "flannel"
				input.Spec.VolumeDriver = "cinder"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for template with dns_nameserver", func() {
				input := minimalValidTemplate()
				input.Spec.DnsNameserver = "8.8.8.8"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for template with docker_volume_size", func() {
				input := minimalValidTemplate()
				input.Spec.DockerVolumeSize = int32Ptr(50)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for template with flavors", func() {
				input := minimalValidTemplate()
				input.Spec.Flavor = "m1.medium"
				input.Spec.MasterFlavor = "m1.large"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for template with floating_ip_enabled", func() {
				input := minimalValidTemplate()
				input.Spec.FloatingIpEnabled = boolPtr(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for template with master_lb_enabled", func() {
				input := minimalValidTemplate()
				input.Spec.MasterLbEnabled = boolPtr(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for template with tls_disabled", func() {
				input := minimalValidTemplate()
				input.Spec.TlsDisabled = boolPtr(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for template with labels", func() {
				input := minimalValidTemplate()
				input.Spec.Labels = map[string]string{
					"kube_tag":          "v1.28.4",
					"container_runtime": "containerd",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for template with region", func() {
				input := minimalValidTemplate()
				input.Spec.Region = "RegionOne"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fully specified template", func() {
				input := minimalValidTemplate()
				input.Spec.Keypair = newStringValueOrRef("my-keypair")
				input.Spec.ExternalNetwork = newStringValueOrRef("public-net")
				input.Spec.FixedNetwork = newStringValueOrRef("tenant-net")
				input.Spec.FixedSubnet = newStringValueOrRef("tenant-subnet")
				input.Spec.NetworkDriver = "calico"
				input.Spec.VolumeDriver = "cinder"
				input.Spec.DnsNameserver = "8.8.8.8"
				input.Spec.DockerVolumeSize = int32Ptr(100)
				input.Spec.Flavor = "m1.xlarge"
				input.Spec.MasterFlavor = "m1.large"
				input.Spec.FloatingIpEnabled = boolPtr(true)
				input.Spec.MasterLbEnabled = boolPtr(true)
				input.Spec.TlsDisabled = boolPtr(false)
				input.Spec.Labels = map[string]string{"env": "production"}
				input.Spec.Region = "RegionOne"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for image with value_from ref", func() {
				input := minimalValidTemplate()
				input.Spec.Image = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-image",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("openstack_container_cluster_template", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidTemplate()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidTemplate()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidTemplate()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackContainerClusterTemplate{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackContainerClusterTemplate",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when coe is empty", func() {
				input := minimalValidTemplate()
				input.Spec.Coe = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when image is missing", func() {
				input := minimalValidTemplate()
				input.Spec.Image = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
