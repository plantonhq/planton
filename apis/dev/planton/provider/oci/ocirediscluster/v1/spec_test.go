package ociredisclusterv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestOciRedisClusterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciRedisClusterSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidRedisCluster() *OciRedisCluster {
	return &OciRedisCluster{
		ApiVersion: "oci.planton.dev/v1",
		Kind:       "OciRedisCluster",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-redis",
		},
		Spec: &OciRedisClusterSpec{
			CompartmentId:   newStringValueOrRef("ocid1.compartment.oc1..example"),
			SubnetId:        newStringValueOrRef("ocid1.subnet.oc1.phx.example"),
			NodeCount:       3,
			NodeMemoryInGbs: 8,
			SoftwareVersion: "V7.0.5",
		},
	}
}

var _ = ginkgo.Describe("OciRedisClusterSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_redis_redis_cluster", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidRedisCluster()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with display_name", func() {
				input := minimalValidRedisCluster()
				input.Spec.DisplayName = "Production Cache"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with explicit nonsharded mode", func() {
				input := minimalValidRedisCluster()
				input.Spec.ClusterMode = OciRedisClusterSpec_nonsharded
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with sharded mode and shard_count", func() {
				input := minimalValidRedisCluster()
				input.Spec.ClusterMode = OciRedisClusterSpec_sharded
				input.Spec.ShardCount = 3
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with nsg_ids", func() {
				input := minimalValidRedisCluster()
				input.Spec.NsgIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.nsg.oc1.phx.1"),
					newStringValueOrRef("ocid1.nsg.oc1.phx.2"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with config_set_id", func() {
				input := minimalValidRedisCluster()
				input.Spec.ConfigSetId = newStringValueOrRef("ocid1.ocicacheconfigset.oc1.phx.example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with minimum node_count of 1", func() {
				input := minimalValidRedisCluster()
				input.Spec.NodeCount = 1
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with small node_memory_in_gbs", func() {
				input := minimalValidRedisCluster()
				input.Spec.NodeMemoryInGbs = 2
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with large node_memory_in_gbs", func() {
				input := minimalValidRedisCluster()
				input.Spec.NodeMemoryInGbs = 64
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via valueFrom ref", func() {
				input := minimalValidRedisCluster()
				input.Spec.CompartmentId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-compartment",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with subnet_id via valueFrom ref", func() {
				input := minimalValidRedisCluster()
				input.Spec.SubnetId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-private-subnet",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full configuration", func() {
				input := minimalValidRedisCluster()
				input.Spec.DisplayName = "Production Sharded Cache"
				input.Spec.ClusterMode = OciRedisClusterSpec_sharded
				input.Spec.ShardCount = 5
				input.Spec.NodeCount = 3
				input.Spec.NodeMemoryInGbs = 32
				input.Spec.SoftwareVersion = "V7.1.1"
				input.Spec.NsgIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.nsg.oc1.phx.1"),
				}
				input.Spec.ConfigSetId = newStringValueOrRef("ocid1.ocicacheconfigset.oc1.phx.example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with nonsharded mode and shard_count zero", func() {
				input := minimalValidRedisCluster()
				input.Spec.ClusterMode = OciRedisClusterSpec_nonsharded
				input.Spec.ShardCount = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_redis_redis_cluster", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidRedisCluster()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidRedisCluster()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidRedisCluster()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciRedisCluster{
					ApiVersion: "oci.planton.dev/v1",
					Kind:       "OciRedisCluster",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-redis"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidRedisCluster()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when subnet_id is missing", func() {
				input := minimalValidRedisCluster()
				input.Spec.SubnetId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when node_count is zero", func() {
				input := minimalValidRedisCluster()
				input.Spec.NodeCount = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when software_version is empty", func() {
				input := minimalValidRedisCluster()
				input.Spec.SoftwareVersion = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when sharded mode without shard_count", func() {
				input := minimalValidRedisCluster()
				input.Spec.ClusterMode = OciRedisClusterSpec_sharded
				input.Spec.ShardCount = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
