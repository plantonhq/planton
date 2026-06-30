package kubernetestelemetryv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestKubernetesTelemetry(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesTelemetry Suite")
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

// targetRef returns a same-namespace PolicyTargetReference (namespace left empty,
// as upstream requires in the 1.26 line).
func targetRef(kind, name string) *kubernetes.KubernetesIstioApiPolicyTargetReference {
	return &kubernetes.KubernetesIstioApiPolicyTargetReference{
		Kind: kind,
		Name: name,
	}
}

var _ = ginkgo.Describe("KubernetesTelemetry Validation Tests", func() {
	var input *KubernetesTelemetry

	ginkgo.BeforeEach(func() {
		input = &KubernetesTelemetry{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesTelemetry",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-telemetry",
			},
			Spec: &KubernetesTelemetrySpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: literal("istio-system"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with only a namespace (an empty mesh-wide telemetry)", func() {
			ginkgo.It("should not return a validation error", func() {
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a selector targeting specific workloads", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Selector = &kubernetes.KubernetesIstioApiWorkloadSelector{
					MatchLabels: map[string]string{"app": "ratings"},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with target_refs instead of a selector", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.TargetRefs = []*kubernetes.KubernetesIstioApiPolicyTargetReference{
					targetRef("Gateway", "edge-gateway"),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a full tracing rule (match, providers, sampling, custom tags)", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Tracing = []*KubernetesTelemetryTracing{
					{
						Match:                    &KubernetesTelemetryTracingSelector{Mode: ptr("CLIENT")},
						Providers:                []*KubernetesTelemetryProviderRef{{Name: "zipkin"}},
						RandomSamplingPercentage: ptr(10.0),
						DisableSpanReporting:     ptr(false),
						EnableIstioTags:          ptr(true),
						// Advanced hidden-but-functional upstream knob; carried for fidelity.
						UseRequestIdForTraceSampling: ptr(false),
						CustomTags: map[string]*KubernetesTelemetryCustomTag{
							"lit": {Literal: &KubernetesTelemetryCustomTagLiteral{Value: "foo"}},
							"env": {Environment: &KubernetesTelemetryCustomTagEnvironment{Name: "POD_NAME", DefaultValue: "unknown"}},
							"hdr": {Header: &KubernetesTelemetryCustomTagRequestHeader{Name: "x-req-id"}},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with random_sampling_percentage at the bounds", func() {
			ginkgo.It("should accept 0 and 100", func() {
				for _, p := range []float64{0, 100} {
					input.Spec.Tracing = []*KubernetesTelemetryTracing{{RandomSamplingPercentage: ptr(p)}}
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "sampling %v should be valid", p)
				}
			})
		})

		ginkgo.Context("with a full metrics rule (providers, override, tag overrides, interval)", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Metrics = []*KubernetesTelemetryMetrics{
					{
						Providers:         []*KubernetesTelemetryProviderRef{{Name: "prometheus"}},
						ReportingInterval: ptr("5s"),
						Overrides: []*KubernetesTelemetryMetricsOverride{
							{
								Match:    &KubernetesTelemetryMetricSelector{Metric: ptr("REQUEST_COUNT"), Mode: ptr("SERVER")},
								Disabled: ptr(false),
								TagOverrides: map[string]*KubernetesTelemetryTagOverride{
									"request_method": {Operation: ptr("UPSERT"), Value: ptr("request.method")},
									"response_code":  {Operation: ptr("REMOVE")},
								},
							},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a custom_metric override", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Metrics = []*KubernetesTelemetryMetrics{
					{Overrides: []*KubernetesTelemetryMetricsOverride{
						{Match: &KubernetesTelemetryMetricSelector{CustomMetric: ptr("my_custom_metric")}},
					}},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a tag override that has no operation set (unset = no value constraint)", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Metrics = []*KubernetesTelemetryMetrics{
					{Overrides: []*KubernetesTelemetryMetricsOverride{
						{TagOverrides: map[string]*KubernetesTelemetryTagOverride{
							"foo": {}, // neither operation nor value -> valid (upstream treats unset operation as neither UPSERT nor REMOVE)
						}},
					}},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a full access logging rule (match, providers, filter)", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.AccessLogging = []*KubernetesTelemetryAccessLogging{
					{
						Match:     &KubernetesTelemetryAccessLoggingSelector{Mode: ptr("SERVER")},
						Providers: []*KubernetesTelemetryProviderRef{{Name: "envoy"}},
						Disabled:  ptr(false),
						Filter:    &KubernetesTelemetryAccessLoggingFilter{Expression: "response.code >= 400"},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with the namespace resolved via a valueFrom reference", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Namespace = valueFrom(cloudresourcekind.CloudResourceKind_KubernetesNamespace, "finance-ns", "spec.name")
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
				input.Kind = "KubernetesAuthorizationPolicy"
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

	ginkgo.Describe("When selector and target_refs are both set", func() {
		ginkgo.It("should return a validation error (at most one attachment)", func() {
			input.Spec.Selector = &kubernetes.KubernetesIstioApiWorkloadSelector{
				MatchLabels: map[string]string{"app": "ratings"},
			}
			input.Spec.TargetRefs = []*kubernetes.KubernetesIstioApiPolicyTargetReference{
				targetRef("Gateway", "edge-gateway"),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When too many target_refs are provided", func() {
		ginkgo.It("should reject more than 16 entries", func() {
			refs := make([]*kubernetes.KubernetesIstioApiPolicyTargetReference, 0, 17)
			for i := 0; i < 17; i++ {
				refs = append(refs, targetRef("Service", "svc"))
			}
			input.Spec.TargetRefs = refs
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When a target_ref is invalid", func() {
		ginkgo.Context("with a cross-namespace reference", func() {
			ginkgo.It("should return a validation error", func() {
				ref := targetRef("Gateway", "edge-gateway")
				ref.Namespace = "other-ns"
				input.Spec.TargetRefs = []*kubernetes.KubernetesIstioApiPolicyTargetReference{ref}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("without a kind", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TargetRefs = []*kubernetes.KubernetesIstioApiPolicyTargetReference{
					{Name: "edge-gateway"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When the selector match labels are invalid", func() {
		ginkgo.Context("with a wildcard in a label key", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Selector = &kubernetes.KubernetesIstioApiWorkloadSelector{
					MatchLabels: map[string]string{"app*": "ratings"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When tracing fields are invalid", func() {
		ginkgo.Context("with random_sampling_percentage below 0", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Tracing = []*KubernetesTelemetryTracing{{RandomSamplingPercentage: ptr(-1.0)}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with random_sampling_percentage above 100", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Tracing = []*KubernetesTelemetryTracing{{RandomSamplingPercentage: ptr(100.01)}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with an invalid tracing match mode", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Tracing = []*KubernetesTelemetryTracing{{Match: &KubernetesTelemetryTracingSelector{Mode: ptr("INBOUND")}}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a provider that has an empty name", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Tracing = []*KubernetesTelemetryTracing{{Providers: []*KubernetesTelemetryProviderRef{{Name: ""}}}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When a custom tag is invalid", func() {
		ginkgo.Context("with more than one source set", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Tracing = []*KubernetesTelemetryTracing{{
					CustomTags: map[string]*KubernetesTelemetryCustomTag{
						"bad": {
							Literal:     &KubernetesTelemetryCustomTagLiteral{Value: "foo"},
							Environment: &KubernetesTelemetryCustomTagEnvironment{Name: "POD_NAME"},
						},
					},
				}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a literal that has an empty value", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Tracing = []*KubernetesTelemetryTracing{{
					CustomTags: map[string]*KubernetesTelemetryCustomTag{
						"lit": {Literal: &KubernetesTelemetryCustomTagLiteral{Value: ""}},
					},
				}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with an environment source missing its name", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Tracing = []*KubernetesTelemetryTracing{{
					CustomTags: map[string]*KubernetesTelemetryCustomTag{
						"env": {Environment: &KubernetesTelemetryCustomTagEnvironment{DefaultValue: "x"}},
					},
				}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When a metric selector is invalid", func() {
		ginkgo.Context("with both metric and custom_metric set", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Metrics = []*KubernetesTelemetryMetrics{
					{Overrides: []*KubernetesTelemetryMetricsOverride{
						{Match: &KubernetesTelemetryMetricSelector{Metric: ptr("REQUEST_COUNT"), CustomMetric: ptr("my_metric")}},
					}},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a metric outside the closed set", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Metrics = []*KubernetesTelemetryMetrics{
					{Overrides: []*KubernetesTelemetryMetricsOverride{
						{Match: &KubernetesTelemetryMetricSelector{Metric: ptr("REQUEST_TOTAL")}},
					}},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with an empty custom_metric", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Metrics = []*KubernetesTelemetryMetrics{
					{Overrides: []*KubernetesTelemetryMetricsOverride{
						{Match: &KubernetesTelemetryMetricSelector{CustomMetric: ptr("")}},
					}},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with an invalid mode", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Metrics = []*KubernetesTelemetryMetrics{
					{Overrides: []*KubernetesTelemetryMetricsOverride{
						{Match: &KubernetesTelemetryMetricSelector{Mode: ptr("BOTH")}},
					}},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When a tag override is invalid", func() {
		ginkgo.Context("with operation UPSERT but no value", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Metrics = []*KubernetesTelemetryMetrics{
					{Overrides: []*KubernetesTelemetryMetricsOverride{
						{TagOverrides: map[string]*KubernetesTelemetryTagOverride{
							"foo": {Operation: ptr("UPSERT")},
						}},
					}},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with operation UPSERT but an empty value", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Metrics = []*KubernetesTelemetryMetrics{
					{Overrides: []*KubernetesTelemetryMetricsOverride{
						{TagOverrides: map[string]*KubernetesTelemetryTagOverride{
							"foo": {Operation: ptr("UPSERT"), Value: ptr("")},
						}},
					}},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with operation REMOVE but a value set", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Metrics = []*KubernetesTelemetryMetrics{
					{Overrides: []*KubernetesTelemetryMetricsOverride{
						{TagOverrides: map[string]*KubernetesTelemetryTagOverride{
							"foo": {Operation: ptr("REMOVE"), Value: ptr("x")},
						}},
					}},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with an operation outside the closed set", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Metrics = []*KubernetesTelemetryMetrics{
					{Overrides: []*KubernetesTelemetryMetricsOverride{
						{TagOverrides: map[string]*KubernetesTelemetryTagOverride{
							"foo": {Operation: ptr("DELETE")},
						}},
					}},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When the metrics reporting interval is invalid", func() {
		ginkgo.Context("with a duration below the 1ms minimum", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Metrics = []*KubernetesTelemetryMetrics{{ReportingInterval: ptr("0s")}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a non-duration string", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Metrics = []*KubernetesTelemetryMetrics{{ReportingInterval: ptr("soon")}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When an access logging provider is invalid", func() {
		ginkgo.Context("with an empty provider name", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.AccessLogging = []*KubernetesTelemetryAccessLogging{
					{Providers: []*KubernetesTelemetryProviderRef{{Name: ""}}},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
