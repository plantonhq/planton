package openstackcontainerclusterv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOpenStackContainerClusterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackContainerClusterSpec Validation Tests")
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

func minimalValidCluster() *OpenStackContainerCluster {
	return &OpenStackContainerCluster{
		ApiVersion: "openstack.openmcf.org/v1",
		Kind:       "OpenStackContainerCluster",
		Metadata: &shared.CloudResourceMetadata{
			Name: "k8s-cluster",
		},
		Spec: &OpenStackContainerClusterSpec{
			ClusterTemplate: newStringValueOrRef("template-uuid-1234"),
		},
	}
}

var _ = ginkgo.Describe("OpenStackContainerClusterSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_container_cluster", func() {

			ginkgo.It("should not return a validation error for minimal valid cluster", func() {
				input := minimalValidCluster()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for cluster with master_count", func() {
				input := minimalValidCluster()
				input.Spec.MasterCount = int32Ptr(3)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for cluster with node_count", func() {
				input := minimalValidCluster()
				input.Spec.NodeCount = int32Ptr(5)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for cluster with keypair", func() {
				input := minimalValidCluster()
				input.Spec.Keypair = newStringValueOrRef("my-keypair")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for cluster with flavor overrides", func() {
				input := minimalValidCluster()
				input.Spec.Flavor = "m1.xlarge"
				input.Spec.MasterFlavor = "m1.large"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for cluster with docker_volume_size", func() {
				input := minimalValidCluster()
				input.Spec.DockerVolumeSize = int32Ptr(100)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for cluster with labels", func() {
				input := minimalValidCluster()
				input.Spec.Labels = map[string]string{
					"kube_tag":          "v1.28.4",
					"container_runtime": "containerd",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for cluster with create_timeout", func() {
				input := minimalValidCluster()
				input.Spec.CreateTimeout = int32Ptr(60)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for cluster with floating_ip_enabled", func() {
				input := minimalValidCluster()
				input.Spec.FloatingIpEnabled = boolPtr(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for cluster with region", func() {
				input := minimalValidCluster()
				input.Spec.Region = "RegionOne"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fully specified cluster", func() {
				input := minimalValidCluster()
				input.Spec.MasterCount = int32Ptr(3)
				input.Spec.NodeCount = int32Ptr(5)
				input.Spec.Keypair = newStringValueOrRef("my-keypair")
				input.Spec.Flavor = "m1.xlarge"
				input.Spec.MasterFlavor = "m1.large"
				input.Spec.DockerVolumeSize = int32Ptr(100)
				input.Spec.Labels = map[string]string{"env": "production"}
				input.Spec.CreateTimeout = int32Ptr(90)
				input.Spec.FloatingIpEnabled = boolPtr(true)
				input.Spec.Region = "RegionOne"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for cluster_template with value_from ref", func() {
				input := minimalValidCluster()
				input.Spec.ClusterTemplate = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-template",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("openstack_container_cluster", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidCluster()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidCluster()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidCluster()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackContainerCluster{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackContainerCluster",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when cluster_template is missing", func() {
				input := minimalValidCluster()
				input.Spec.ClusterTemplate = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
