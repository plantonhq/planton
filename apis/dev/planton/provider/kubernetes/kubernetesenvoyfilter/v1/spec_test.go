package kubernetesenvoyfilterv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestKubernetesEnvoyFilter(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesEnvoyFilter Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func valueFrom(kind cloudresourcekind.CloudResourceKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Kind:      kind,
				Name:      name,
				FieldPath: fieldPath,
			},
		},
	}
}

// ptr returns a pointer to v, for setting proto3 `optional` scalar fields.
func ptr[T any](v T) *T { return &v }

// mustStruct builds a google.protobuf.Struct from a Go map, failing the test on error.
func mustStruct(m map[string]interface{}) *structpb.Struct {
	s, err := structpb.NewStruct(m)
	gomega.Expect(err).To(gomega.BeNil())
	return s
}

// clusterMergePatch returns a minimal valid config patch: MERGE a small JSON value into a
// matched cluster.
func clusterMergePatch() *KubernetesEnvoyFilterConfigPatch {
	return &KubernetesEnvoyFilterConfigPatch{
		ApplyTo: ptr("CLUSTER"),
		Match: &KubernetesEnvoyFilterEnvoyConfigObjectMatch{
			Context: ptr("SIDECAR_OUTBOUND"),
			Cluster: &KubernetesEnvoyFilterClusterMatch{
				Service: ptr("reviews.default.svc.cluster.local"),
			},
		},
		Patch: &KubernetesEnvoyFilterPatch{
			Operation: ptr("MERGE"),
			Value: mustStruct(map[string]interface{}{
				"connect_timeout": "5s",
			}),
		},
	}
}

// httpFilterInsertPatch returns a config patch that inserts an HTTP filter before a named
// sub-filter in the HTTP connection manager — exercising the deepest match nesting.
func httpFilterInsertPatch() *KubernetesEnvoyFilterConfigPatch {
	return &KubernetesEnvoyFilterConfigPatch{
		ApplyTo: ptr("HTTP_FILTER"),
		Match: &KubernetesEnvoyFilterEnvoyConfigObjectMatch{
			Context: ptr("SIDECAR_INBOUND"),
			Listener: &KubernetesEnvoyFilterListenerMatch{
				PortNumber: ptr(uint32(8080)),
				FilterChain: &KubernetesEnvoyFilterFilterChainMatch{
					Filter: &KubernetesEnvoyFilterFilterMatch{
						Name: ptr("envoy.filters.network.http_connection_manager"),
						SubFilter: &KubernetesEnvoyFilterSubFilterMatch{
							Name: ptr("envoy.filters.http.router"),
						},
					},
				},
			},
		},
		Patch: &KubernetesEnvoyFilterPatch{
			Operation: ptr("INSERT_BEFORE"),
			Value: mustStruct(map[string]interface{}{
				"name": "envoy.filters.http.cors",
			}),
		},
	}
}

func targetRef(kind, name string) *kubernetes.KubernetesIstioApiPolicyTargetReference {
	return &kubernetes.KubernetesIstioApiPolicyTargetReference{
		Group: "gateway.networking.k8s.io",
		Kind:  kind,
		Name:  name,
	}
}

var _ = ginkgo.Describe("KubernetesEnvoyFilter Validation Tests", func() {
	var input *KubernetesEnvoyFilter

	ginkgo.BeforeEach(func() {
		input = &KubernetesEnvoyFilter{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesEnvoyFilter",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-envoy-filter",
			},
			Spec: &KubernetesEnvoyFilterSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace:     literal("default"),
				ConfigPatches: []*KubernetesEnvoyFilterConfigPatch{clusterMergePatch()},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with a minimal cluster MERGE patch and no attachment", func() {
			ginkgo.It("should not return a validation error", func() {
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a workload_selector only", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.WorkloadSelector = &kubernetes.KubernetesIstioApiNetworkingWorkloadSelector{
					Labels: map[string]string{"app": "reviews"},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with target_refs only", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.TargetRefs = []*kubernetes.KubernetesIstioApiPolicyTargetReference{
					targetRef("Gateway", "my-gateway"),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with neither selector nor target_refs (match-all)", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.WorkloadSelector = nil
				input.Spec.TargetRefs = nil
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with priority set", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Priority = ptr(int32(-10))
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a deeply nested HTTP_FILTER listener match", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.ConfigPatches = []*KubernetesEnvoyFilterConfigPatch{httpFilterInsertPatch()}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a route_configuration / vhost / route match", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.ConfigPatches = []*KubernetesEnvoyFilterConfigPatch{
					{
						ApplyTo: ptr("HTTP_ROUTE"),
						Match: &KubernetesEnvoyFilterEnvoyConfigObjectMatch{
							Context: ptr("GATEWAY"),
							RouteConfiguration: &KubernetesEnvoyFilterRouteConfigurationMatch{
								PortNumber: ptr(uint32(443)),
								Vhost: &KubernetesEnvoyFilterVirtualHostMatch{
									Name: ptr("example.com:443"),
									Route: &KubernetesEnvoyFilterRouteMatch{
										Action: ptr("ROUTE"),
									},
								},
							},
						},
						Patch: &KubernetesEnvoyFilterPatch{
							Operation: ptr("MERGE"),
							Value:     mustStruct(map[string]interface{}{"name": "example"}),
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a REMOVE operation and no value", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.ConfigPatches = []*KubernetesEnvoyFilterConfigPatch{
					{
						ApplyTo: ptr("HTTP_FILTER"),
						Match: &KubernetesEnvoyFilterEnvoyConfigObjectMatch{
							Listener: &KubernetesEnvoyFilterListenerMatch{
								FilterChain: &KubernetesEnvoyFilterFilterChainMatch{
									Filter: &KubernetesEnvoyFilterFilterMatch{
										Name: ptr("envoy.filters.network.http_connection_manager"),
										SubFilter: &KubernetesEnvoyFilterSubFilterMatch{
											Name: ptr("envoy.filters.http.fault"),
										},
									},
								},
							},
						},
						Patch: &KubernetesEnvoyFilterPatch{Operation: ptr("REMOVE")},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with apply_to BOOTSTRAP (deprecated but accepted)", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.ConfigPatches[0].ApplyTo = ptr("BOOTSTRAP")
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with no config_patches (a valid no-op)", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.ConfigPatches = nil
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with the namespace resolved via a valueFrom reference", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Namespace = valueFrom(cloudresourcekind.CloudResourceKind_KubernetesNamespace, "mesh-ns", "spec.name")
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When the envelope is invalid", func() {
		ginkgo.Context("with the wrong api_version", func() {
			ginkgo.It("should return a validation error", func() {
				input.ApiVersion = "v1"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with the wrong kind", func() {
			ginkgo.It("should return a validation error", func() {
				input.Kind = "KubernetesServiceEntry"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("without metadata", func() {
			ginkgo.It("should return a validation error", func() {
				input.Metadata = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("without a spec", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("without a namespace", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Namespace = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When the attachment model is violated", func() {
		ginkgo.Context("with both workload_selector and target_refs set", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.WorkloadSelector = &kubernetes.KubernetesIstioApiNetworkingWorkloadSelector{
					Labels: map[string]string{"app": "reviews"},
				}
				input.Spec.TargetRefs = []*kubernetes.KubernetesIstioApiPolicyTargetReference{
					targetRef("Gateway", "my-gateway"),
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a wildcard in a workload_selector label value", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.WorkloadSelector = &kubernetes.KubernetesIstioApiNetworkingWorkloadSelector{
					Labels: map[string]string{"app": "reviews-*"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When target_refs are invalid", func() {
		ginkgo.Context("with more than 16 target_refs", func() {
			ginkgo.It("should return a validation error", func() {
				refs := make([]*kubernetes.KubernetesIstioApiPolicyTargetReference, 0, 17)
				for i := 0; i < 17; i++ {
					refs = append(refs, targetRef("Gateway", "g"))
				}
				input.Spec.TargetRefs = refs
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a target_ref missing its kind", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TargetRefs = []*kubernetes.KubernetesIstioApiPolicyTargetReference{
					{Group: "gateway.networking.k8s.io", Name: "my-gateway"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a target_ref missing its name", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TargetRefs = []*kubernetes.KubernetesIstioApiPolicyTargetReference{
					{Group: "gateway.networking.k8s.io", Kind: "Gateway"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a cross-namespace target_ref", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TargetRefs = []*kubernetes.KubernetesIstioApiPolicyTargetReference{
					{Group: "gateway.networking.k8s.io", Kind: "Gateway", Name: "my-gateway", Namespace: "other-ns"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When the match object_types exclusivity is violated", func() {
		ginkgo.Context("with both listener and cluster set", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.ConfigPatches[0].Match = &KubernetesEnvoyFilterEnvoyConfigObjectMatch{
					Listener: &KubernetesEnvoyFilterListenerMatch{Name: ptr("0.0.0.0_8080")},
					Cluster:  &KubernetesEnvoyFilterClusterMatch{Name: ptr("outbound|80||x")},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with both listener and route_configuration set", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.ConfigPatches[0].Match = &KubernetesEnvoyFilterEnvoyConfigObjectMatch{
					Listener:           &KubernetesEnvoyFilterListenerMatch{Name: ptr("0.0.0.0_8080")},
					RouteConfiguration: &KubernetesEnvoyFilterRouteConfigurationMatch{Name: ptr("http_proxy")},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When closed-set enums carry an unknown value", func() {
		ginkgo.Context("with an unknown apply_to", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.ConfigPatches[0].ApplyTo = ptr("SIDECAR")
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with an unknown match context", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.ConfigPatches[0].Match.Context = ptr("INGRESS")
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with an unknown patch operation", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.ConfigPatches[0].Patch.Operation = ptr("UPSERT")
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with an unknown filter_class", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.ConfigPatches[0].Patch.FilterClass = ptr("CACHE")
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with an unknown route action", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.ConfigPatches = []*KubernetesEnvoyFilterConfigPatch{
					{
						ApplyTo: ptr("HTTP_ROUTE"),
						Match: &KubernetesEnvoyFilterEnvoyConfigObjectMatch{
							RouteConfiguration: &KubernetesEnvoyFilterRouteConfigurationMatch{
								Vhost: &KubernetesEnvoyFilterVirtualHostMatch{
									Route: &KubernetesEnvoyFilterRouteMatch{Action: ptr("FORWARD")},
								},
							},
						},
						Patch: &KubernetesEnvoyFilterPatch{Operation: ptr("MERGE")},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When a port number is out of range", func() {
		ginkgo.Context("with a cluster port_number above 65535", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.ConfigPatches[0].Match.Cluster.PortNumber = ptr(uint32(70000))
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a listener filter-chain destination_port above 65535", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.ConfigPatches = []*KubernetesEnvoyFilterConfigPatch{
					{
						ApplyTo: ptr("FILTER_CHAIN"),
						Match: &KubernetesEnvoyFilterEnvoyConfigObjectMatch{
							Listener: &KubernetesEnvoyFilterListenerMatch{
								FilterChain: &KubernetesEnvoyFilterFilterChainMatch{
									DestinationPort: ptr(uint32(70000)),
								},
							},
						},
						Patch: &KubernetesEnvoyFilterPatch{Operation: ptr("MERGE")},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
