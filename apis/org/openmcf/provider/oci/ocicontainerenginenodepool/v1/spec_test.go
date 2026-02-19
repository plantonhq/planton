package ocicontainerenginenodepoolv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOciContainerEngineNodePoolSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciContainerEngineNodePoolSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidNodePool() *OciContainerEngineNodePool {
	return &OciContainerEngineNodePool{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciContainerEngineNodePool",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-node-pool",
		},
		Spec: &OciContainerEngineNodePoolSpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
			ClusterId:     newStringValueOrRef("ocid1.cluster.oc1.iad.example"),
			NodeShape:     "VM.Standard.E4.Flex",
			NodeConfigDetails: &OciContainerEngineNodePoolSpec_NodeConfigDetails{
				PlacementConfigs: []*OciContainerEngineNodePoolSpec_PlacementConfig{
					{
						AvailabilityDomain: "Uocm:PHX-AD-1",
						SubnetId:           newStringValueOrRef("ocid1.subnet.oc1.iad.example"),
					},
				},
				Size: 3,
			},
		},
	}
}

var _ = ginkgo.Describe("OciContainerEngineNodePoolSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_container_engine_node_pool", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidNodePool()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with display name set", func() {
				input := minimalValidNodePool()
				input.Spec.Name = "Production Worker Nodes"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with explicit kubernetes version", func() {
				input := minimalValidNodePool()
				input.Spec.KubernetesVersion = "v1.28.2"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with flex shape config", func() {
				input := minimalValidNodePool()
				input.Spec.NodeShapeConfig = &OciContainerEngineNodePoolSpec_NodeShapeConfig{
					Ocpus:       2.0,
					MemoryInGbs: 32.0,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with node source details", func() {
				input := minimalValidNodePool()
				input.Spec.NodeSourceDetails = &OciContainerEngineNodePoolSpec_NodeSourceDetails{
					ImageId:             "ocid1.image.oc1.iad.example",
					BootVolumeSizeInGbs: 100,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with SSH public key", func() {
				input := minimalValidNodePool()
				input.Spec.SshPublicKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQ..."
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with initial node labels", func() {
				input := minimalValidNodePool()
				input.Spec.InitialNodeLabels = []*OciContainerEngineNodePoolSpec_NodeLabel{
					{Key: "workload-type", Value: "gpu"},
					{Key: "team", Value: "ml-platform"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with node metadata", func() {
				input := minimalValidNodePool()
				input.Spec.NodeMetadata = map[string]string{
					"user_data": "IyEvYmluL2Jhc2g=",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple placement configs across ADs", func() {
				input := minimalValidNodePool()
				input.Spec.NodeConfigDetails.PlacementConfigs = []*OciContainerEngineNodePoolSpec_PlacementConfig{
					{
						AvailabilityDomain: "Uocm:PHX-AD-1",
						SubnetId:           newStringValueOrRef("ocid1.subnet.oc1.iad.ad1"),
					},
					{
						AvailabilityDomain: "Uocm:PHX-AD-2",
						SubnetId:           newStringValueOrRef("ocid1.subnet.oc1.iad.ad2"),
					},
					{
						AvailabilityDomain: "Uocm:PHX-AD-3",
						SubnetId:           newStringValueOrRef("ocid1.subnet.oc1.iad.ad3"),
					},
				}
				input.Spec.NodeConfigDetails.Size = 6
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with placement fault domains", func() {
				input := minimalValidNodePool()
				input.Spec.NodeConfigDetails.PlacementConfigs[0].FaultDomains = []string{
					"FAULT-DOMAIN-1", "FAULT-DOMAIN-2",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with preemptible node config", func() {
				input := minimalValidNodePool()
				preserve := true
				input.Spec.NodeConfigDetails.PlacementConfigs[0].PreemptibleNodeConfig = &OciContainerEngineNodePoolSpec_PreemptibleNodeConfig{
					IsPreserveBootVolume: &preserve,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with capacity reservation", func() {
				input := minimalValidNodePool()
				input.Spec.NodeConfigDetails.PlacementConfigs[0].CapacityReservationId = newStringValueOrRef("ocid1.capacityreservation.oc1.iad.example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with node NSGs", func() {
				input := minimalValidNodePool()
				input.Spec.NodeConfigDetails.NsgIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.networksecuritygroup.oc1.iad.worker1"),
					newStringValueOrRef("ocid1.networksecuritygroup.oc1.iad.worker2"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with KMS key and PV encryption", func() {
				input := minimalValidNodePool()
				input.Spec.NodeConfigDetails.KmsKeyId = newStringValueOrRef("ocid1.key.oc1.iad.bootvol")
				input.Spec.NodeConfigDetails.IsPvEncryptionInTransitEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with VCN-native pod network options", func() {
				input := minimalValidNodePool()
				input.Spec.NodeConfigDetails.PodNetworkOptionDetails = &OciContainerEngineNodePoolSpec_PodNetworkOptionDetails{
					CniType:        OciContainerEngineNodePoolSpec_oci_vcn_ip_native,
					MaxPodsPerNode: 31,
					PodNsgIds: []*foreignkeyv1.StringValueOrRef{
						newStringValueOrRef("ocid1.networksecuritygroup.oc1.iad.pods"),
					},
					PodSubnetIds: []*foreignkeyv1.StringValueOrRef{
						newStringValueOrRef("ocid1.subnet.oc1.iad.pods"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with flannel overlay pod network options", func() {
				input := minimalValidNodePool()
				input.Spec.NodeConfigDetails.PodNetworkOptionDetails = &OciContainerEngineNodePoolSpec_PodNetworkOptionDetails{
					CniType: OciContainerEngineNodePoolSpec_flannel_overlay,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with node eviction settings", func() {
				input := minimalValidNodePool()
				forceAction := true
				forceDelete := false
				input.Spec.NodeEvictionSettings = &OciContainerEngineNodePoolSpec_NodeEvictionSettings{
					EvictionGraceDuration:           "PT30M",
					IsForceActionAfterGraceDuration: &forceAction,
					IsForceDeleteAfterGraceDuration: &forceDelete,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with node pool cycling details", func() {
				input := minimalValidNodePool()
				input.Spec.NodePoolCyclingDetails = &OciContainerEngineNodePoolSpec_NodePoolCyclingDetails{
					IsNodeCyclingEnabled: true,
					MaximumSurge:         "25%",
					MaximumUnavailable:   "0",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via valueFrom ref", func() {
				input := minimalValidNodePool()
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

			ginkgo.It("should not return a validation error with cluster_id via valueFrom ref", func() {
				input := minimalValidNodePool()
				input.Spec.ClusterId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-oke-cluster",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all options populated", func() {
				input := minimalValidNodePool()
				input.Spec.Name = "full-node-pool"
				input.Spec.KubernetesVersion = "v1.28.2"
				input.Spec.NodeShapeConfig = &OciContainerEngineNodePoolSpec_NodeShapeConfig{
					Ocpus:       4.0,
					MemoryInGbs: 64.0,
				}
				input.Spec.NodeSourceDetails = &OciContainerEngineNodePoolSpec_NodeSourceDetails{
					ImageId:             "ocid1.image.oc1.iad.example",
					BootVolumeSizeInGbs: 100,
				}
				input.Spec.SshPublicKey = "ssh-rsa AAAAB3..."
				input.Spec.InitialNodeLabels = []*OciContainerEngineNodePoolSpec_NodeLabel{
					{Key: "env", Value: "production"},
				}
				input.Spec.NodeMetadata = map[string]string{"startup": "script"}
				preserve := false
				input.Spec.NodeConfigDetails = &OciContainerEngineNodePoolSpec_NodeConfigDetails{
					PlacementConfigs: []*OciContainerEngineNodePoolSpec_PlacementConfig{
						{
							AvailabilityDomain: "Uocm:PHX-AD-1",
							SubnetId:           newStringValueOrRef("ocid1.subnet.oc1.iad.ad1"),
							FaultDomains:       []string{"FAULT-DOMAIN-1"},
							PreemptibleNodeConfig: &OciContainerEngineNodePoolSpec_PreemptibleNodeConfig{
								IsPreserveBootVolume: &preserve,
							},
						},
					},
					Size: 5,
					NsgIds: []*foreignkeyv1.StringValueOrRef{
						newStringValueOrRef("ocid1.nsg.oc1.iad.worker"),
					},
					KmsKeyId:                       newStringValueOrRef("ocid1.key.oc1.iad.boot"),
					IsPvEncryptionInTransitEnabled:  true,
					PodNetworkOptionDetails: &OciContainerEngineNodePoolSpec_PodNetworkOptionDetails{
						CniType:        OciContainerEngineNodePoolSpec_oci_vcn_ip_native,
						MaxPodsPerNode: 31,
						PodSubnetIds:   []*foreignkeyv1.StringValueOrRef{newStringValueOrRef("ocid1.subnet.oc1.iad.pods")},
					},
				}
				forceAction := true
				input.Spec.NodeEvictionSettings = &OciContainerEngineNodePoolSpec_NodeEvictionSettings{
					EvictionGraceDuration:           "PT60M",
					IsForceActionAfterGraceDuration: &forceAction,
				}
				input.Spec.NodePoolCyclingDetails = &OciContainerEngineNodePoolSpec_NodePoolCyclingDetails{
					IsNodeCyclingEnabled: true,
					MaximumSurge:         "1",
					MaximumUnavailable:   "0",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_container_engine_node_pool", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidNodePool()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidNodePool()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidNodePool()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciContainerEngineNodePool{
					ApiVersion: "oci.openmcf.org/v1",
					Kind:       "OciContainerEngineNodePool",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-pool"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidNodePool()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when cluster_id is missing", func() {
				input := minimalValidNodePool()
				input.Spec.ClusterId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when node_shape is empty", func() {
				input := minimalValidNodePool()
				input.Spec.NodeShape = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when node_config_details is missing", func() {
				input := minimalValidNodePool()
				input.Spec.NodeConfigDetails = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when placement_configs is empty", func() {
				input := minimalValidNodePool()
				input.Spec.NodeConfigDetails.PlacementConfigs = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when size is zero", func() {
				input := minimalValidNodePool()
				input.Spec.NodeConfigDetails.Size = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when size is negative", func() {
				input := minimalValidNodePool()
				input.Spec.NodeConfigDetails.Size = -1
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when placement availability_domain is empty", func() {
				input := minimalValidNodePool()
				input.Spec.NodeConfigDetails.PlacementConfigs[0].AvailabilityDomain = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when placement subnet_id is missing", func() {
				input := minimalValidNodePool()
				input.Spec.NodeConfigDetails.PlacementConfigs[0].SubnetId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when node_source_details image_id is empty", func() {
				input := minimalValidNodePool()
				input.Spec.NodeSourceDetails = &OciContainerEngineNodePoolSpec_NodeSourceDetails{
					ImageId: "",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when pod_network cni_type is unspecified", func() {
				input := minimalValidNodePool()
				input.Spec.NodeConfigDetails.PodNetworkOptionDetails = &OciContainerEngineNodePoolSpec_PodNetworkOptionDetails{
					CniType: OciContainerEngineNodePoolSpec_cni_unspecified,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
