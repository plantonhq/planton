package ocicontainerengineclusterv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOciContainerEngineClusterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciContainerEngineClusterSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidCluster() *OciContainerEngineCluster {
	return &OciContainerEngineCluster{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciContainerEngineCluster",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-oke-cluster",
		},
		Spec: &OciContainerEngineClusterSpec{
			CompartmentId:    newStringValueOrRef("ocid1.compartment.oc1..example"),
			VcnId:            newStringValueOrRef("ocid1.vcn.oc1.iad.example"),
			KubernetesVersion: "v1.28.2",
		},
	}
}

var _ = ginkgo.Describe("OciContainerEngineClusterSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_container_engine_cluster", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidCluster()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with display name set", func() {
				input := minimalValidCluster()
				input.Spec.Name = "My Production Cluster"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for basic cluster with public endpoint", func() {
				input := minimalValidCluster()
				input.Spec.Type = OciContainerEngineClusterSpec_basic_cluster
				isPublic := true
				input.Spec.EndpointConfig = &OciContainerEngineClusterSpec_EndpointConfig{
					SubnetId:          newStringValueOrRef("ocid1.subnet.oc1.iad.example"),
					IsPublicIpEnabled: &isPublic,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for enhanced cluster with private endpoint and NSGs", func() {
				input := minimalValidCluster()
				input.Spec.Type = OciContainerEngineClusterSpec_enhanced_cluster
				isPublic := false
				input.Spec.EndpointConfig = &OciContainerEngineClusterSpec_EndpointConfig{
					SubnetId:          newStringValueOrRef("ocid1.subnet.oc1.iad.example"),
					IsPublicIpEnabled: &isPublic,
					NsgIds: []*foreignkeyv1.StringValueOrRef{
						newStringValueOrRef("ocid1.networksecuritygroup.oc1.iad.example1"),
						newStringValueOrRef("ocid1.networksecuritygroup.oc1.iad.example2"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with VCN-native CNI and kubernetes network config", func() {
				input := minimalValidCluster()
				input.Spec.CniType = OciContainerEngineClusterSpec_oci_vcn_ip_native
				input.Spec.Options = &OciContainerEngineClusterSpec_ClusterOptions{
					KubernetesNetworkConfig: &OciContainerEngineClusterSpec_KubernetesNetworkConfig{
						PodsCidr:     "10.244.0.0/16",
						ServicesCidr: "10.96.0.0/16",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with flannel overlay CNI", func() {
				input := minimalValidCluster()
				input.Spec.CniType = OciContainerEngineClusterSpec_flannel_overlay
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with KMS encryption key", func() {
				input := minimalValidCluster()
				input.Spec.KmsKeyId = newStringValueOrRef("ocid1.key.oc1.iad.example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with service LB subnet IDs", func() {
				input := minimalValidCluster()
				input.Spec.Options = &OciContainerEngineClusterSpec_ClusterOptions{
					ServiceLbSubnetIds: []*foreignkeyv1.StringValueOrRef{
						newStringValueOrRef("ocid1.subnet.oc1.iad.lb1"),
						newStringValueOrRef("ocid1.subnet.oc1.iad.lb2"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with service LB config", func() {
				input := minimalValidCluster()
				input.Spec.Options = &OciContainerEngineClusterSpec_ClusterOptions{
					ServiceLbConfig: &OciContainerEngineClusterSpec_ServiceLbConfig{
						BackendNsgIds: []*foreignkeyv1.StringValueOrRef{
							newStringValueOrRef("ocid1.networksecuritygroup.oc1.iad.backend"),
						},
						FreeformTags: map[string]string{"Department": "Engineering"},
						DefinedTags:  map[string]string{"Operations.CostCenter": "42"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with persistent volume config", func() {
				input := minimalValidCluster()
				input.Spec.Options = &OciContainerEngineClusterSpec_ClusterOptions{
					PersistentVolumeConfig: &OciContainerEngineClusterSpec_PersistentVolumeConfig{
						FreeformTags: map[string]string{"team": "platform"},
						DefinedTags:  map[string]string{"Operations.CostCenter": "42"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with OIDC inline config", func() {
				input := minimalValidCluster()
				input.Spec.Options = &OciContainerEngineClusterSpec_ClusterOptions{
					OpenIdConnectTokenAuthenticationConfig: &OciContainerEngineClusterSpec_OpenIdConnectTokenAuthenticationConfig{
						IsOpenIdConnectAuthEnabled: true,
						IssuerUrl:                  "https://accounts.google.com",
						ClientId:                   "my-client-id",
						UsernameClaim:              "email",
						GroupsClaim:                "groups",
						SigningAlgorithms:          []string{"RS256"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with OIDC configuration file", func() {
				input := minimalValidCluster()
				input.Spec.Options = &OciContainerEngineClusterSpec_ClusterOptions{
					OpenIdConnectTokenAuthenticationConfig: &OciContainerEngineClusterSpec_OpenIdConnectTokenAuthenticationConfig{
						IsOpenIdConnectAuthEnabled: true,
						ConfigurationFile:          "eyJhcGlWZXJzaW9uIjogInYxIn0=",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with OIDC discovery enabled", func() {
				input := minimalValidCluster()
				input.Spec.Options = &OciContainerEngineClusterSpec_ClusterOptions{
					IsOpenIdConnectDiscoveryEnabled: true,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with image policy config", func() {
				input := minimalValidCluster()
				input.Spec.ImagePolicyConfig = &OciContainerEngineClusterSpec_ImagePolicyConfig{
					IsPolicyEnabled: true,
					KeyDetails: []*OciContainerEngineClusterSpec_ImagePolicyKeyDetail{
						{KmsKeyId: newStringValueOrRef("ocid1.key.oc1.iad.example")},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all options populated", func() {
				input := minimalValidCluster()
				input.Spec.Name = "full-cluster"
				input.Spec.Type = OciContainerEngineClusterSpec_enhanced_cluster
				input.Spec.CniType = OciContainerEngineClusterSpec_oci_vcn_ip_native
				isPublic := false
				input.Spec.EndpointConfig = &OciContainerEngineClusterSpec_EndpointConfig{
					SubnetId:          newStringValueOrRef("ocid1.subnet.oc1.iad.api"),
					IsPublicIpEnabled: &isPublic,
					NsgIds:            []*foreignkeyv1.StringValueOrRef{newStringValueOrRef("ocid1.nsg.oc1.iad.api")},
				}
				input.Spec.KmsKeyId = newStringValueOrRef("ocid1.key.oc1.iad.secrets")
				input.Spec.ImagePolicyConfig = &OciContainerEngineClusterSpec_ImagePolicyConfig{
					IsPolicyEnabled: true,
					KeyDetails:      []*OciContainerEngineClusterSpec_ImagePolicyKeyDetail{{KmsKeyId: newStringValueOrRef("ocid1.key.oc1.iad.images")}},
				}
				input.Spec.Options = &OciContainerEngineClusterSpec_ClusterOptions{
					KubernetesNetworkConfig: &OciContainerEngineClusterSpec_KubernetesNetworkConfig{
						PodsCidr:     "10.244.0.0/16",
						ServicesCidr: "10.96.0.0/16",
					},
					ServiceLbSubnetIds: []*foreignkeyv1.StringValueOrRef{
						newStringValueOrRef("ocid1.subnet.oc1.iad.lb"),
					},
					IpFamilies: []OciContainerEngineClusterSpec_IpFamily{
						OciContainerEngineClusterSpec_ipv4,
					},
					ServiceLbConfig: &OciContainerEngineClusterSpec_ServiceLbConfig{
						BackendNsgIds: []*foreignkeyv1.StringValueOrRef{newStringValueOrRef("ocid1.nsg.oc1.iad.lb")},
						FreeformTags:  map[string]string{"team": "platform"},
					},
					PersistentVolumeConfig: &OciContainerEngineClusterSpec_PersistentVolumeConfig{
						FreeformTags: map[string]string{"team": "platform"},
					},
					OpenIdConnectTokenAuthenticationConfig: &OciContainerEngineClusterSpec_OpenIdConnectTokenAuthenticationConfig{
						IsOpenIdConnectAuthEnabled: true,
						IssuerUrl:                  "https://idp.example.com",
						ClientId:                   "my-client",
					},
					IsOpenIdConnectDiscoveryEnabled: true,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with dual-stack IPv4+IPv6", func() {
				input := minimalValidCluster()
				input.Spec.Options = &OciContainerEngineClusterSpec_ClusterOptions{
					IpFamilies: []OciContainerEngineClusterSpec_IpFamily{
						OciContainerEngineClusterSpec_ipv4,
						OciContainerEngineClusterSpec_ipv6,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via value_from ref", func() {
				input := minimalValidCluster()
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

			ginkgo.It("should not return a validation error with vcn_id via value_from ref", func() {
				input := minimalValidCluster()
				input.Spec.VcnId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-vcn",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_container_engine_cluster", func() {

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
				input := &OciContainerEngineCluster{
					ApiVersion: "oci.openmcf.org/v1",
					Kind:       "OciContainerEngineCluster",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-cluster"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidCluster()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when vcn_id is missing", func() {
				input := minimalValidCluster()
				input.Spec.VcnId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kubernetes_version is empty", func() {
				input := minimalValidCluster()
				input.Spec.KubernetesVersion = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when endpoint_config is present but subnet_id is missing", func() {
				input := minimalValidCluster()
				input.Spec.EndpointConfig = &OciContainerEngineClusterSpec_EndpointConfig{
					SubnetId: nil,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
