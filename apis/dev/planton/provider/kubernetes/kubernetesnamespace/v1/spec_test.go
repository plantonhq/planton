package kubernetesnamespacev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestKubernetesNamespaceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesNamespaceSpec Validation Suite")
}

var _ = ginkgo.Describe("KubernetesNamespaceSpec validations", func() {

	ginkgo.Context("When valid specs are provided", func() {

		ginkgo.It("accepts a minimal valid spec", func() {
			spec := &KubernetesNamespaceSpec{
				Name: "dev-team",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a spec with built-in profile SMALL", func() {
			spec := &KubernetesNamespaceSpec{
				Name: "dev-namespace",
				ResourceProfile: &KubernetesNamespaceResourceProfile{
					ProfileConfig: &KubernetesNamespaceResourceProfile_Preset{
						Preset: KubernetesNamespaceResourceProfile_small,
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a spec with built-in profile LARGE", func() {
			spec := &KubernetesNamespaceSpec{
				Name: "prod-namespace",
				ResourceProfile: &KubernetesNamespaceResourceProfile{
					ProfileConfig: &KubernetesNamespaceResourceProfile_Preset{
						Preset: KubernetesNamespaceResourceProfile_large,
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a spec with custom resource quotas", func() {
			spec := &KubernetesNamespaceSpec{
				Name: "custom-namespace",
				ResourceProfile: &KubernetesNamespaceResourceProfile{
					ProfileConfig: &KubernetesNamespaceResourceProfile_Custom{
						Custom: &KubernetesNamespaceCustomQuotas{
							Cpu: &KubernetesNamespaceCpuQuota{
								Requests: "4",
								Limits:   "8",
							},
							Memory: &KubernetesNamespaceMemoryQuota{
								Requests: "8Gi",
								Limits:   "16Gi",
							},
							ObjectCounts: &KubernetesNamespaceObjectCountQuotas{
								Pods:                   50,
								Services:               20,
								Configmaps:             100,
								Secrets:                100,
								PersistentVolumeClaims: 10,
								LoadBalancers:          5,
							},
							DefaultLimits: &KubernetesNamespaceDefaultLimits{
								DefaultCpuRequest:    "100m",
								DefaultCpuLimit:      "500m",
								DefaultMemoryRequest: "128Mi",
								DefaultMemoryLimit:   "512Mi",
							},
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a spec with network isolation enabled", func() {
			spec := &KubernetesNamespaceSpec{
				Name: "secure-namespace",
				NetworkConfig: &KubernetesNamespaceNetworkConfig{
					IsolateIngress:           true,
					RestrictEgress:           true,
					AllowedIngressNamespaces: []string{"istio-system", "kube-system"},
					AllowedEgressCidrs:       []string{"10.0.0.0/8", "192.168.1.0/24"},
					AllowedEgressDomains:     []string{"api.stripe.com", "*.github.com"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a spec with Istio service mesh enabled", func() {
			spec := &KubernetesNamespaceSpec{
				Name: "mesh-namespace",
				ServiceMeshConfig: &KubernetesNamespaceServiceMeshConfig{
					Enabled:     true,
					MeshType:    KubernetesNamespaceServiceMeshConfig_istio,
					RevisionTag: "prod-stable",
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a spec with Linkerd service mesh enabled", func() {
			spec := &KubernetesNamespaceSpec{
				Name: "linkerd-namespace",
				ServiceMeshConfig: &KubernetesNamespaceServiceMeshConfig{
					Enabled:  true,
					MeshType: KubernetesNamespaceServiceMeshConfig_linkerd,
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a spec with pod security standard BASELINE", func() {
			spec := &KubernetesNamespaceSpec{
				Name:                "baseline-namespace",
				PodSecurityStandard: KubernetesNamespaceSpec_baseline,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a spec with pod security standard RESTRICTED", func() {
			spec := &KubernetesNamespaceSpec{
				Name:                "restricted-namespace",
				PodSecurityStandard: KubernetesNamespaceSpec_restricted,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a spec with labels and annotations", func() {
			spec := &KubernetesNamespaceSpec{
				Name: "annotated-namespace",
				Labels: map[string]string{
					"team":        "platform",
					"environment": "production",
					"cost-center": "engineering",
				},
				Annotations: map[string]string{
					"linkerd.io/inject": "enabled",
					"janitor/ttl":       "24h",
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("When invalid specs are provided", func() {

		ginkgo.It("rejects empty namespace name", func() {
			spec := &KubernetesNamespaceSpec{
				Name: "",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects namespace name with uppercase letters", func() {
			spec := &KubernetesNamespaceSpec{
				Name: "DevTeam",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects namespace name with underscores", func() {
			spec := &KubernetesNamespaceSpec{
				Name: "dev_team",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects namespace name starting with hyphen", func() {
			spec := &KubernetesNamespaceSpec{
				Name: "-devteam",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects namespace name ending with hyphen", func() {
			spec := &KubernetesNamespaceSpec{
				Name: "devteam-",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects namespace name longer than 63 characters", func() {
			spec := &KubernetesNamespaceSpec{
				Name: "this-is-a-very-long-namespace-name-that-exceeds-the-maximum-length-allowed",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects service mesh enabled without mesh type", func() {
			spec := &KubernetesNamespaceSpec{
				Name: "mesh-namespace",
				ServiceMeshConfig: &KubernetesNamespaceServiceMeshConfig{
					Enabled:  true,
					MeshType: KubernetesNamespaceServiceMeshConfig_service_mesh_type_unspecified,
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects empty CPU requests in custom quota", func() {
			spec := &KubernetesNamespaceSpec{
				Name: "custom-namespace",
				ResourceProfile: &KubernetesNamespaceResourceProfile{
					ProfileConfig: &KubernetesNamespaceResourceProfile_Custom{
						Custom: &KubernetesNamespaceCustomQuotas{
							Cpu: &KubernetesNamespaceCpuQuota{
								Requests: "",
								Limits:   "8",
							},
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects empty CPU limits in custom quota", func() {
			spec := &KubernetesNamespaceSpec{
				Name: "custom-namespace",
				ResourceProfile: &KubernetesNamespaceResourceProfile{
					ProfileConfig: &KubernetesNamespaceResourceProfile_Custom{
						Custom: &KubernetesNamespaceCustomQuotas{
							Cpu: &KubernetesNamespaceCpuQuota{
								Requests: "4",
								Limits:   "",
							},
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects empty memory requests in custom quota", func() {
			spec := &KubernetesNamespaceSpec{
				Name: "custom-namespace",
				ResourceProfile: &KubernetesNamespaceResourceProfile{
					ProfileConfig: &KubernetesNamespaceResourceProfile_Custom{
						Custom: &KubernetesNamespaceCustomQuotas{
							Memory: &KubernetesNamespaceMemoryQuota{
								Requests: "",
								Limits:   "16Gi",
							},
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects empty memory limits in custom quota", func() {
			spec := &KubernetesNamespaceSpec{
				Name: "custom-namespace",
				ResourceProfile: &KubernetesNamespaceResourceProfile{
					ProfileConfig: &KubernetesNamespaceResourceProfile_Custom{
						Custom: &KubernetesNamespaceCustomQuotas{
							Memory: &KubernetesNamespaceMemoryQuota{
								Requests: "8Gi",
								Limits:   "",
							},
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects zero or negative pod count in object quotas", func() {
			spec := &KubernetesNamespaceSpec{
				Name: "custom-namespace",
				ResourceProfile: &KubernetesNamespaceResourceProfile{
					ProfileConfig: &KubernetesNamespaceResourceProfile_Custom{
						Custom: &KubernetesNamespaceCustomQuotas{
							ObjectCounts: &KubernetesNamespaceObjectCountQuotas{
								Pods: 0,
							},
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects empty default CPU request in default limits", func() {
			spec := &KubernetesNamespaceSpec{
				Name: "custom-namespace",
				ResourceProfile: &KubernetesNamespaceResourceProfile{
					ProfileConfig: &KubernetesNamespaceResourceProfile_Custom{
						Custom: &KubernetesNamespaceCustomQuotas{
							DefaultLimits: &KubernetesNamespaceDefaultLimits{
								DefaultCpuRequest:    "",
								DefaultCpuLimit:      "500m",
								DefaultMemoryRequest: "128Mi",
								DefaultMemoryLimit:   "512Mi",
							},
						},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects revision tag longer than 63 characters", func() {
			spec := &KubernetesNamespaceSpec{
				Name: "mesh-namespace",
				ServiceMeshConfig: &KubernetesNamespaceServiceMeshConfig{
					Enabled:     true,
					MeshType:    KubernetesNamespaceServiceMeshConfig_istio,
					RevisionTag: "this-is-a-very-long-revision-tag-that-exceeds-the-maximum-length-allowed-for-revision-tags",
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
